package skip

import "github.com/ikidev/lightning"

// New creates a middleware handler which skips the wrapped handler
// if the exclude predicate returns true.
func New(handler lightning.Handler, exclude func(c *lightning.Ctx) bool) lightning.Handler {
	if exclude == nil {
		return handler
	}

	return func(c *lightning.Ctx) error {
		if exclude(c) {
			return c.Next()
		}

		return handler(c)
	}
}
