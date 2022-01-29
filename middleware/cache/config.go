package cache

import (
	"time"

	"github.com/ikidev/lightning"
	"github.com/ikidev/lightning/utils"
)

// Config defines the config for middleware.
type Config struct {
	// Next defines a function to skip this middleware when returned true.
	//
	// Optional. Default: nil
	Next func(req *lightning.Request, res *lightning.Response) bool

	// Expiration is the time that an cached response will live
	//
	// Optional. Default: 1 * time.Minute
	Expiration time.Duration

	// CacheHeader header on response header, indicate cache status, with the following possible return value
	//
	// hit, miss, unreachable
	//
	// Optional. Default: X-Cache
	CacheHeader string

	// CacheControl enables client side caching if set to true
	//
	// Optional. Default: false
	CacheControl bool

	// Key allows you to generate custom keys, by default c.Path() is used
	//
	// Default: func(c *fiber.Ctx) string {
	//   return utils.CopyString(c.Path())
	// }
	KeyGenerator func(req *lightning.Request, res *lightning.Response) string

	// allows you to generate custom Expiration Key By Key, default is Expiration (Optional)
	//
	// Default: nil
	ExpirationGenerator func(*lightning.Request, *lightning.Response, *Config) time.Duration

	// Store is used to store the state of the middleware
	//
	// Default: an in memory store for this process only
	Storage lightning.Storage
}

// ConfigDefault is the default config
var ConfigDefault = Config{
	Next:         nil,
	Expiration:   1 * time.Minute,
	CacheHeader:  "X-Cache",
	CacheControl: false,
	KeyGenerator: func(req *lightning.Request, res *lightning.Response) string {
		return utils.CopyString(req.Path())
	},
	ExpirationGenerator: nil,
	Storage:             nil,
}

// Helper function to set default values
func configDefault(config ...Config) Config {
	// Return default config if nothing provided
	if len(config) < 1 {
		return ConfigDefault
	}

	// Override default config
	cfg := config[0]

	if cfg.Next == nil {
		cfg.Next = ConfigDefault.Next
	}
	if int(cfg.Expiration.Seconds()) == 0 {
		cfg.Expiration = ConfigDefault.Expiration
	}
	if cfg.CacheHeader == "" {
		cfg.CacheHeader = ConfigDefault.CacheHeader
	}
	if cfg.KeyGenerator == nil {
		cfg.KeyGenerator = ConfigDefault.KeyGenerator
	}
	return cfg
}
