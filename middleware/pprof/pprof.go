package pprof

import (
	"net/http/pprof"
	"strings"

	"github.com/ikidev/lightning"
	"github.com/valyala/fasthttp/fasthttpadaptor"
)

// Set pprof adaptors
var (
	pprofIndex        = fasthttpadaptor.NewFastHTTPHandlerFunc(pprof.Index)
	pprofCmdline      = fasthttpadaptor.NewFastHTTPHandlerFunc(pprof.Cmdline)
	pprofProfile      = fasthttpadaptor.NewFastHTTPHandlerFunc(pprof.Profile)
	pprofSymbol       = fasthttpadaptor.NewFastHTTPHandlerFunc(pprof.Symbol)
	pprofTrace        = fasthttpadaptor.NewFastHTTPHandlerFunc(pprof.Trace)
	pprofAllocs       = fasthttpadaptor.NewFastHTTPHandlerFunc(pprof.Handler("allocs").ServeHTTP)
	pprofBlock        = fasthttpadaptor.NewFastHTTPHandlerFunc(pprof.Handler("block").ServeHTTP)
	pprofGoroutine    = fasthttpadaptor.NewFastHTTPHandlerFunc(pprof.Handler("goroutine").ServeHTTP)
	pprofHeap         = fasthttpadaptor.NewFastHTTPHandlerFunc(pprof.Handler("heap").ServeHTTP)
	pprofMutex        = fasthttpadaptor.NewFastHTTPHandlerFunc(pprof.Handler("mutex").ServeHTTP)
	pprofThreadcreate = fasthttpadaptor.NewFastHTTPHandlerFunc(pprof.Handler("threadcreate").ServeHTTP)
)

// New creates a new middleware handler
func New() lightning.Handler {
	// Return new handler
	return func(req *lightning.Request, res *lightning.Response) error {
		path := req.Path()
		// We are only interested in /debug/pprof routes
		if len(path) < 12 || !strings.HasPrefix(path, "/debug/pprof") {
			return req.Next()
		}
		// Switch to original path without stripped slashes
		switch path {
		case "/debug/pprof/":
			pprofIndex(req.FastHTTPContext())
		case "/debug/pprof/cmdline":
			pprofCmdline(req.FastHTTPContext())
		case "/debug/pprof/profile":
			pprofProfile(req.FastHTTPContext())
		case "/debug/pprof/symbol":
			pprofSymbol(req.FastHTTPContext())
		case "/debug/pprof/trace":
			pprofTrace(req.FastHTTPContext())
		case "/debug/pprof/allocs":
			pprofAllocs(req.FastHTTPContext())
		case "/debug/pprof/block":
			pprofBlock(req.FastHTTPContext())
		case "/debug/pprof/goroutine":
			pprofGoroutine(req.FastHTTPContext())
		case "/debug/pprof/heap":
			pprofHeap(req.FastHTTPContext())
		case "/debug/pprof/mutex":
			pprofMutex(req.FastHTTPContext())
		case "/debug/pprof/threadcreate":
			pprofThreadcreate(req.FastHTTPContext())
		default:
			// pprof index only works with trailing slash
			if strings.HasSuffix(path, "/") {
				path = strings.TrimRight(path, "/")
			} else {
				path = "/debug/pprof/"
			}

			return req.Redirect(path, lightning.StatusFound)
		}
		return nil
	}
}
