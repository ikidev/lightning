package expvar

import (
	"strings"

	"github.com/ikidev/lightning"
	"github.com/valyala/fasthttp/expvarhandler"
)

// New creates a new middleware handler
func New() lightning.Handler {
	// Return new handler
	return func(req *lightning.Request, res *lightning.Response) error {
		path := req.Path()
		// We are only interested in /debug/vars routes
		if len(path) < 11 || !strings.HasPrefix(path, "/debug/vars") {
			return req.Next()
		}
		if path == "/debug/vars" {
			expvarhandler.ExpvarHandler(req.Ctx().Context())
			return nil
		}

		return req.Redirect("/debug/vars", 302)
	}
}
