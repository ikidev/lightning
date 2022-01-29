package csrf

import (
	"errors"
	"time"

	"github.com/ikidev/lightning"
)

var (
	errTokenNotFound = errors.New("csrf token not found")
)

// New creates a new middleware handler
func New(config ...Config) lightning.Handler {
	// Set default config
	cfg := configDefault(config...)

	// Create manager to simplify storage operations ( see manager.go )
	manager := newManager(cfg.Storage)

	dummyValue := []byte{'+'}

	// Return new handler
	return func(req *lightning.Request, res *lightning.Response) (err error) {
		// Don't execute middleware if Next returns true
		if cfg.Next != nil && cfg.Next(req, res) {
			return req.Next()
		}

		var token string

		// Action depends on the HTTP method
		switch req.Method() {
		case lightning.MethodGet, lightning.MethodHead, lightning.MethodOptions, lightning.MethodTrace:
			// Declare empty token and try to get existing CSRF from cookie
			token = req.GetCookie(cfg.CookieName)
		default:
			// Assume that anything not defined as 'safe' by RFC7231 needs protection

			// Extract token from client request i.e. header, query, param, form or cookie
			token, err = cfg.extractor(req, res)
			if err != nil {
				return cfg.ErrorHandler(req, res, err)
			}

			// if token does not exist in Storage
			if manager.getRaw(token) == nil {
				// Expire cookie
				res.SetCookie(&lightning.Cookie{
					Name:     cfg.CookieName,
					Domain:   cfg.CookieDomain,
					Path:     cfg.CookiePath,
					Expires:  time.Now().Add(-1 * time.Minute),
					Secure:   cfg.CookieSecure,
					HTTPOnly: cfg.CookieHTTPOnly,
					SameSite: cfg.CookieSameSite,
				})
				return cfg.ErrorHandler(req, res, errTokenNotFound)
			}
		}

		// Generate CSRF token if not exist
		if token == "" {
			// And generate a new token
			token = cfg.KeyGenerator()
		}

		// Add/update token to Storage
		manager.setRaw(token, dummyValue, cfg.Expiration)

		// Create cookie to pass token to client
		cookie := &lightning.Cookie{
			Name:     cfg.CookieName,
			Value:    token,
			Domain:   cfg.CookieDomain,
			Path:     cfg.CookiePath,
			Expires:  time.Now().Add(cfg.Expiration),
			Secure:   cfg.CookieSecure,
			HTTPOnly: cfg.CookieHTTPOnly,
			SameSite: cfg.CookieSameSite,
		}
		// Set cookie to response
		res.SetCookie(cookie)

		// Protect clients from caching the response by telling the browser
		// a new header value is generated
		res.Ctx().Vary(lightning.HeaderCookie)

		// Store token in context if set
		if cfg.ContextKey != "" {
			req.Locals(cfg.ContextKey, token)
		}

		// Continue stack
		return req.Next()
	}
}
