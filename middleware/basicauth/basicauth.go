package basicauth

import (
	"encoding/base64"
	"strings"

	"github.com/ikidev/lightning"
	"github.com/ikidev/lightning/utils"
)

// New creates a new middleware handler
func New(config Config) lightning.Handler {
	// Set default config
	cfg := configDefault(config)

	// Return new handler
	return func(req *lightning.Request, res *lightning.Response) error {
		// Don't execute middleware if Next returns true
		if cfg.Next != nil && cfg.Next(req, res) {
			return req.Next()
		}

		// Get authorization header
		auth := req.Header.Get(lightning.HeaderAuthorization)

		// Check if the header contains content besides "basic".
		if len(auth) <= 6 || strings.ToLower(auth[:5]) != "basic" {
			return cfg.Unauthorized(req, res)
		}

		// Decode the header contents
		raw, err := base64.StdEncoding.DecodeString(auth[6:])
		if err != nil {
			return cfg.Unauthorized(req, res)
		}

		// Get the credentials
		creds := utils.UnsafeString(raw)

		// Check if the credentials are in the correct form
		// which is "username:password".
		index := strings.Index(creds, ":")
		if index == -1 {
			return cfg.Unauthorized(req, res)
		}

		// Get the username and password
		username := creds[:index]
		password := creds[index+1:]

		if cfg.Authorizer(username, password) {
			req.Locals(cfg.ContextUsername, username)
			req.Locals(cfg.ContextPassword, password)
			return req.Next()
		}

		// Authentication failed
		return cfg.Unauthorized(req, res)
	}
}
