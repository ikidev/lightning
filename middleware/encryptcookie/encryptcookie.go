package encryptcookie

import (
	"github.com/ikidev/lightning"
	"github.com/valyala/fasthttp"
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

		// Decrypt request cookies
		req.Ctx().Request().Header.VisitAllCookie(func(key, value []byte) {
			keyString := string(key)
			if !isDisabled(keyString, cfg.Except) {
				decryptedValue, err := cfg.Decryptor(string(value), cfg.Key)
				if err != nil {
					req.Ctx().Request().Header.SetCookieBytesKV(key, nil)
				} else {
					req.Ctx().Request().Header.SetCookie(string(key), decryptedValue)
				}
			}
		})

		// Continue stack
		err := req.Next()

		// Encrypt response cookies
		req.Ctx().Response().Header.VisitAllCookie(func(key, value []byte) {
			keyString := string(key)
			if !isDisabled(keyString, cfg.Except) {
				cookieValue := fasthttp.Cookie{}
				cookieValue.SetKeyBytes(key)
				if req.Ctx().Response().Header.Cookie(&cookieValue) {
					encryptedValue, err := cfg.Encryptor(string(cookieValue.Value()), cfg.Key)
					if err == nil {
						cookieValue.SetValue(encryptedValue)
						req.Ctx().Response().Header.SetCookie(&cookieValue)
					} else {
						panic(err)
					}
				}
			}
		})

		return err
	}
}
