package requestid

import (
	"github.com/ikidev/lightning"
)

// New creates a new middleware handler
func New(config ...Config) lightning.Handler {
	// Set default config
	cfg := configDefault(config...)

	// Return new handler
	return func(req *lightning.Request, res *lightning.Response) error {
		// Don't execute middleware if Next returns true
		if cfg.Next != nil && cfg.Next(req, res) {
			return req.Next()
		}
		// Get id from request, else we generate one
		rid := res.Header.Get(cfg.Header, cfg.Generator())

		// Set new id to response header
		res.Header.Set(cfg.Header, rid)

		// Add the request ID to locals
		req.Locals(cfg.ContextKey, rid)

		// Continue stack
		return req.Next()
	}
}
