package timeout

import (
	"fmt"
	"sync"
	"time"

	"github.com/ikidev/lightning"
)

var once sync.Once

// New wraps a handler and aborts the process of the handler if the timeout is reached
func New(handler lightning.Handler, timeout time.Duration) lightning.Handler {
	once.Do(func() {
		fmt.Println("[Warning] timeout contains data race issues, not ready for production!")
	})

	if timeout <= 0 {
		return handler
	}

	// logic is from fasthttp.TimeoutWithCodeHandler https://github.com/valyala/fasthttp/blob/master/server.go#L418
	return func(ctx *lightning.Ctx) error {
		ch := make(chan struct{}, 1)

		go func() {
			defer func() {
				_ = recover()
			}()
			_ = handler(ctx)
			ch <- struct{}{}
		}()

		select {
		case <-ch:
		case <-time.After(timeout):
			return lightning.ErrRequestTimeout
		}

		return nil
	}
}
