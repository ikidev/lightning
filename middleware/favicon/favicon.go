package favicon

import (
	"io/ioutil"
	"net/http"
	"strconv"

	"github.com/ikidev/lightning"
)

// Config defines the config for middleware.
type Config struct {
	// Next defines a function to skip this middleware when returned true.
	//
	// Optional. Default: nil
	Next func(req *lightning.Request, res *lightning.Response) bool

	// File holds the path to an actual favicon that will be cached
	//
	// Optional. Default: ""
	File string `json:"file"`

	// FileSystem is an optional alternate filesystem to search for the favicon in.
	// An example of this could be an embedded or network filesystem
	//
	// Optional. Default: nil
	FileSystem http.FileSystem `json:"-"`

	// CacheControl defines how the Cache-Control header in the response should be set
	//
	// Optional. Default: "public, max-age=31536000"
	CacheControl string `json:"cache_control"`
}

// ConfigDefault is the default config
var ConfigDefault = Config{
	Next:         nil,
	File:         "",
	CacheControl: "public, max-age=31536000",
}

const (
	hType  = "image/x-icon"
	hAllow = "GET, HEAD, OPTIONS"
	hZero  = "0"
)

// New creates a new middleware handler
func New(config ...Config) lightning.Handler {
	// Set default config
	cfg := ConfigDefault

	// Override config if provided
	if len(config) > 0 {
		cfg = config[0]

		// Set default values
		if cfg.Next == nil {
			cfg.Next = ConfigDefault.Next
		}
		if cfg.File == "" {
			cfg.File = ConfigDefault.File
		}
		if cfg.CacheControl == "" {
			cfg.CacheControl = ConfigDefault.CacheControl
		}
	}

	// Load icon if provided
	var (
		err     error
		icon    []byte
		iconLen string
	)
	if cfg.File != "" {
		// read from configured filesystem if present
		if cfg.FileSystem != nil {
			f, err := cfg.FileSystem.Open(cfg.File)
			if err != nil {
				panic(err)
			}
			if icon, err = ioutil.ReadAll(f); err != nil {
				panic(err)
			}
		} else if icon, err = ioutil.ReadFile(cfg.File); err != nil {
			panic(err)
		}

		iconLen = strconv.Itoa(len(icon))
	}

	// Return new handler
	return func(req *lightning.Request, res *lightning.Response) error {
		// Don't execute middleware if Next returns true
		if cfg.Next != nil && cfg.Next(req, res) {
			return req.Next()
		}

		// Only respond to favicon requests
		if len(req.Path()) != 12 || req.Path() != "/favicon.ico" {
			return req.Next()
		}

		// Only allow GET, HEAD and OPTIONS requests
		if req.Method() != lightning.MethodGet && req.Method() != lightning.MethodHead {
			if req.Method() != lightning.MethodOptions {
				res.Status(lightning.StatusMethodNotAllowed)
			} else {
				res.Status(lightning.StatusOK)
			}
			res.Header.Set(lightning.HeaderAllow, hAllow)
			res.Header.Set(lightning.HeaderContentLength, hZero)
			return nil
		}

		// Serve cached favicon
		if len(icon) > 0 {
			res.Header.Set(lightning.HeaderContentLength, iconLen)
			res.Header.Set(lightning.HeaderContentType, hType)
			res.Header.Set(lightning.HeaderCacheControl, cfg.CacheControl)
			return res.Status(lightning.StatusOK).Bytes(icon)
		}

		return res.Status(lightning.StatusNoContent).Send()
	}
}
