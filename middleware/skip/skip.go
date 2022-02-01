package skip

import "github.com/ikidev/lightning"

// New creates a middleware handler which skips the wrapped handler
// if the exclude predicate returns true.
func New(handler lightning.Handler, exclude func(req *lightning.Request, res *lightning.Response) bool) lightning.Handler {
	if exclude == nil {
		return handler
	}

	return func(req *lightning.Request, res *lightning.Response) error {
		if exclude(req, res) {
			return req.Next()
		}

		return handler(req, res)
	}
}
