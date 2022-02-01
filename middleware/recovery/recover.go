package recovery

import (
	"fmt"
	"os"
	"runtime"

	"github.com/ikidev/lightning"
)

func defaultStackTraceHandler(e interface{}) {
	buf := make([]byte, defaultStackTraceBufLen)
	buf = buf[:runtime.Stack(buf, false)]
	_, _ = os.Stderr.WriteString(fmt.Sprintf("panic: %v\n%s\n", e, buf))
}

// New creates a new middleware handler
func New(config ...Config) lightning.Handler {
	// Set default config
	cfg := configDefault(config...)

	// Return new handler
	return func(req *lightning.Request, res *lightning.Response) (err error) {
		// Don't execute middleware if Next returns true
		if cfg.Next != nil && cfg.Next(req, res) {
			return req.Next()
		}

		// Catch panics
		defer func() {
			if r := recover(); r != nil {
				if cfg.EnableStackTrace {
					cfg.StackTraceHandler(r)
				}

				var ok bool
				if err, ok = r.(error); !ok {
					// Set error that will call the global error handler
					err = fmt.Errorf("%v", r)
				}
			}
		}()

		// Return err if exist, else move to next handler
		return req.Next()
	}
}
