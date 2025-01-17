package requestid

import (
	"github.com/ikidev/lightning"
	"github.com/ikidev/lightning/utils"
)

// Config defines the config for middleware.
type Config struct {
	// Next defines a function to skip this middleware when returned true.
	//
	// Optional. Default: nil
	Next func(req *lightning.Request, res *lightning.Response) bool

	// Header is the header key where to get/set the unique request ID
	//
	// Optional. Default: "X-Request-ID"
	Header string

	// Generator defines a function to generate the unique identifier.
	//
	// Optional. Default: utils.UUID
	Generator func() string

	// ContextKey defines the key used when storing the request ID in
	// the locals for a specific request.
	//
	// Optional. Default: requestid
	ContextKey string
}

// ConfigDefault is the default config
var ConfigDefault = Config{
	Next:       nil,
	Header:     lightning.HeaderXRequestID,
	Generator:  utils.UUID,
	ContextKey: "requestid",
}

// Helper function to set default values
func configDefault(config ...Config) Config {
	// Return default config if nothing provided
	if len(config) < 1 {
		return ConfigDefault
	}

	// Override default config
	cfg := config[0]

	// Set default values
	if cfg.Header == "" {
		cfg.Header = ConfigDefault.Header
	}
	if cfg.Generator == nil {
		cfg.Generator = ConfigDefault.Generator
	}
	if cfg.ContextKey == "" {
		cfg.ContextKey = ConfigDefault.ContextKey
	}
	return cfg
}
