package cors

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/ikidev/lightning"
)

// Config defines the config for middleware.
type Config struct {
	// Next defines a function to skip this middleware when returned true.
	//
	// Optional. Default: nil
	Next func(req *lightning.Request, res *lightning.Response) bool

	// AllowOrigin defines a list of origins that may access the resource.
	//
	// Optional. Default value "*"
	AllowOrigins string

	// AllowMethods defines a list methods allowed when accessing the resource.
	// This is used in response to a preflight request.
	//
	// Optional. Default value "GET,POST,HEAD,PUT,DELETE,PATCH"
	AllowMethods string

	// AllowHeaders defines a list of request headers that can be used when
	// making the actual request. This is in response to a preflight request.
	//
	// Optional. Default value "".
	AllowHeaders string

	// AllowCredentials indicates whether or not the response to the request
	// can be exposed when the credentials flag is true. When used as part of
	// a response to a preflight request, this indicates whether or not the
	// actual request can be made using credentials.
	//
	// Optional. Default value false.
	AllowCredentials bool

	// ExposeHeaders defines a whitelist headers that clients are allowed to
	// access.
	//
	// Optional. Default value "".
	ExposeHeaders string

	// MaxAge indicates how long (in seconds) the results of a preflight request
	// can be cached.
	//
	// Optional. Default value 0.
	MaxAge int
}

// ConfigDefault is the default config
var ConfigDefault = Config{
	Next:         nil,
	AllowOrigins: "*",
	AllowMethods: strings.Join([]string{
		lightning.MethodGet,
		lightning.MethodPost,
		lightning.MethodHead,
		lightning.MethodPut,
		lightning.MethodDelete,
		lightning.MethodPatch,
	}, ","),
	AllowHeaders:     "",
	AllowCredentials: false,
	ExposeHeaders:    "",
	MaxAge:           0,
}

// New creates a new middleware handler
func New(config ...Config) lightning.Handler {
	// Set default config
	cfg := ConfigDefault

	// Override config if provided
	if len(config) > 0 {
		cfg = config[0]

		// Set default values
		if cfg.AllowMethods == "" {
			cfg.AllowMethods = ConfigDefault.AllowMethods
		}
		if cfg.AllowOrigins == "" {
			cfg.AllowOrigins = ConfigDefault.AllowOrigins
		}
	}

	// Convert string to slice
	allowOrigins := strings.Split(strings.ReplaceAll(cfg.AllowOrigins, " ", ""), ",")

	// Strip white spaces
	allowMethods := strings.ReplaceAll(cfg.AllowMethods, " ", "")
	allowHeaders := strings.ReplaceAll(cfg.AllowHeaders, " ", "")
	exposeHeaders := strings.ReplaceAll(cfg.ExposeHeaders, " ", "")

	// Convert int to string
	maxAge := strconv.Itoa(cfg.MaxAge)

	// Return new handler
	return func(req *lightning.Request, res *lightning.Response) error {
		// Don't execute middleware if Next returns true
		if cfg.Next != nil && cfg.Next(req, res) {
			return req.Next()
		}

		// Get origin header
		origin := req.Header.Get(lightning.HeaderOrigin)
		allowOrigin := ""

		// Check allowed origins
		for _, o := range allowOrigins {
			if o == "*" && cfg.AllowCredentials {
				allowOrigin = origin
				break
			}
			if o == "*" || o == origin {
				allowOrigin = o
				break
			}
			if matchSubdomain(origin, o) {
				allowOrigin = origin
				break
			}
		}

		// Simple request
		if req.Method() != http.MethodOptions {
			res.Ctx().Vary(lightning.HeaderOrigin)
			res.Header.Set(lightning.HeaderAccessControlAllowOrigin, allowOrigin)

			if cfg.AllowCredentials {
				res.Header.Set(lightning.HeaderAccessControlAllowCredentials, "true")
			}
			if exposeHeaders != "" {
				res.Header.Set(lightning.HeaderAccessControlExposeHeaders, exposeHeaders)
			}
			return req.Next()
		}

		// Preflight request
		res.Ctx().Vary(lightning.HeaderOrigin)
		res.Ctx().Vary(lightning.HeaderAccessControlRequestMethod)
		res.Ctx().Vary(lightning.HeaderAccessControlRequestHeaders)
		res.Header.Set(lightning.HeaderAccessControlAllowOrigin, allowOrigin)
		res.Header.Set(lightning.HeaderAccessControlAllowMethods, allowMethods)

		// Set Allow-Credentials if set to true
		if cfg.AllowCredentials {
			res.Header.Set(lightning.HeaderAccessControlAllowCredentials, "true")
		}

		// Set Allow-Headers if not empty
		if allowHeaders != "" {
			res.Header.Set(lightning.HeaderAccessControlAllowHeaders, allowHeaders)
		} else {
			h := req.Header.Get(lightning.HeaderAccessControlRequestHeaders)
			if h != "" {
				res.Header.Set(lightning.HeaderAccessControlAllowHeaders, h)
			}
		}

		// Set MaxAge is set
		if cfg.MaxAge > 0 {
			res.Header.Set(lightning.HeaderAccessControlMaxAge, maxAge)
		}

		// Send 204 No Content
		return res.Status(lightning.StatusNoContent).Send()
	}
}
