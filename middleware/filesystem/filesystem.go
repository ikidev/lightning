package filesystem

import (
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"

	"github.com/ikidev/lightning"
	"github.com/ikidev/lightning/utils"
)

// Config defines the config for middleware.
type Config struct {
	// Next defines a function to skip this middleware when returned true.
	//
	// Optional. Default: nil
	Next func(req *lightning.Request, res *lightning.Response) bool

	// Root is a FileSystem that provides access
	// to a collection of files and directories.
	//
	// Required. Default: nil
	Root http.FileSystem `json:"-"`

	// PathPrefix defines a prefix to be added to a filepath when
	// reading a file from the FileSystem.
	//
	// Use when using Go 1.16 embed.FS
	//
	// Optional. Default ""
	PathPrefix string `json:"path_prefix"`

	// Enable directory browsing.
	//
	// Optional. Default: false
	Browse bool `json:"browse"`

	// Index file for serving a directory.
	//
	// Optional. Default: "index.html"
	Index string `json:"index"`

	// The value for the Cache-Control HTTP-header
	// that is set on the file response. MaxAge is defined in seconds.
	//
	// Optional. Default value 0.
	MaxAge int `json:"max_age"`

	// File to return if path is not found. Useful for SPA's.
	//
	// Optional. Default: ""
	NotFoundFile string `json:"not_found_file"`
}

// ConfigDefault is the default config
var ConfigDefault = Config{
	Next:       nil,
	Root:       nil,
	PathPrefix: "",
	Browse:     false,
	Index:      "/index.html",
	MaxAge:     0,
}

// New creates a new middleware handler
func New(config ...Config) lightning.Handler {
	// Set default config
	cfg := ConfigDefault

	// Override config if provided
	if len(config) > 0 {
		cfg = config[0]

		// Set default values
		if cfg.Index == "" {
			cfg.Index = ConfigDefault.Index
		}
		if !strings.HasPrefix(cfg.Index, "/") {
			cfg.Index = "/" + cfg.Index
		}
		if cfg.NotFoundFile != "" && !strings.HasPrefix(cfg.NotFoundFile, "/") {
			cfg.NotFoundFile = "/" + cfg.NotFoundFile
		}
	}

	if cfg.Root == nil {
		panic("filesystem: Root cannot be nil")
	}

	if cfg.PathPrefix != "" && !strings.HasPrefix(cfg.PathPrefix, "/") {
		cfg.PathPrefix = "/" + cfg.PathPrefix
	}

	var once sync.Once
	var prefix string
	cacheControlStr := "public, max-age=" + strconv.Itoa(cfg.MaxAge)

	// Return new handler
	return func(req *lightning.Request, res *lightning.Response) (err error) {
		// Don't execute middleware if Next returns true
		if cfg.Next != nil && cfg.Next(req, res) {
			return req.Next()
		}

		method := req.Method()

		// We only serve static assets on GET or HEAD methods
		if method != lightning.MethodGet && method != lightning.MethodHead {
			return req.Next()
		}

		// Set prefix once
		once.Do(func() {
			prefix = req.Route().Path
		})

		// Strip prefix
		path := strings.TrimPrefix(req.Path(), prefix)
		if !strings.HasPrefix(path, "/") {
			path = "/" + path
		}
		// Add PathPrefix
		if cfg.PathPrefix != "" {
			// PathPrefix already has a "/" prefix
			path = cfg.PathPrefix + path
		}

		var (
			file http.File
			stat os.FileInfo
		)

		if len(path) > 1 {
			path = utils.TrimRight(path, '/')
		}
		file, err = cfg.Root.Open(path)
		if err != nil && os.IsNotExist(err) && cfg.NotFoundFile != "" {
			file, err = cfg.Root.Open(cfg.NotFoundFile)
		}

		if err != nil {
			if os.IsNotExist(err) {
				return res.Status(lightning.StatusNotFound).Send()
			}
			return
		}

		if stat, err = file.Stat(); err != nil {
			return
		}

		// Serve index if path is directory
		if stat.IsDir() {
			indexPath := utils.TrimRight(path, '/') + cfg.Index
			index, err := cfg.Root.Open(indexPath)
			if err == nil {
				indexStat, err := index.Stat()
				if err == nil {
					file = index
					stat = indexStat
				}
			}
		}

		// Browse directory if no index found and browsing is enabled
		if stat.IsDir() {
			if cfg.Browse {
				return dirList(req.Ctx(), file)
			}
			return lightning.ErrForbidden
		}

		modTime := stat.ModTime()
		contentLength := int(stat.Size())

		// Set Content Type header
		res.Type(getFileExtension(stat.Name()))

		// Set Last Modified header
		if !modTime.IsZero() {
			res.Header.Set(lightning.HeaderLastModified, modTime.UTC().Format(http.TimeFormat))
		}

		if method == lightning.MethodGet {
			if cfg.MaxAge > 0 {
				res.Header.Set(lightning.HeaderCacheControl, cacheControlStr)
			}
			res.Ctx().Response().SetBodyStream(file, contentLength)
			return nil
		}
		if method == lightning.MethodHead {
			res.Ctx().Request().ResetBody()
			// Fasthttp should skipbody by default if HEAD?
			res.Ctx().Response().SkipBody = true
			res.Ctx().Response().Header.SetContentLength(contentLength)
			if err := file.Close(); err != nil {
				return err
			}
			return nil
		}

		return req.Next()
	}
}

// SendFile ...
func SendFile(req *lightning.Request, res *lightning.Response, fs http.FileSystem, path string) (err error) {
	var (
		file http.File
		stat os.FileInfo
	)

	file, err = fs.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return lightning.ErrNotFound
		}
		return err
	}

	if stat, err = file.Stat(); err != nil {
		return err
	}

	// Serve index if path is directory
	if stat.IsDir() {
		indexPath := utils.TrimRight(path, '/') + ConfigDefault.Index
		index, err := fs.Open(indexPath)
		if err == nil {
			indexStat, err := index.Stat()
			if err == nil {
				file = index
				stat = indexStat
			}
		}
	}

	// Return forbidden if no index found
	if stat.IsDir() {
		return lightning.ErrForbidden
	}

	modTime := stat.ModTime()
	contentLength := int(stat.Size())

	// Set Content Type header
	res.Type(getFileExtension(stat.Name()))

	// Set Last Modified header
	if !modTime.IsZero() {
		res.Header.Set(lightning.HeaderLastModified, modTime.UTC().Format(http.TimeFormat))
	}

	method := req.Method()
	if method == lightning.MethodGet {
		res.Ctx().Response().SetBodyStream(file, contentLength)
		return nil
	}
	if method == lightning.MethodHead {
		res.Ctx().Request().ResetBody()
		// Fasthttp should skipbody by default if HEAD?
		res.Ctx().Response().SkipBody = true
		res.Ctx().Response().Header.SetContentLength(contentLength)
		if err := file.Close(); err != nil {
			return err
		}
		return nil
	}

	return nil
}
