// Special thanks to @codemicro for moving this to fiber core
// Original middleware: github.com/codemicro/fiber-cache
package cache

import (
	"strconv"
	"sync"
	"sync/atomic"
	"time"

	"github.com/ikidev/lightning"
	"github.com/ikidev/lightning/utils"
)

// timestampUpdatePeriod is the period which is used to check the cache expiration.
// It should not be too long to provide more or less acceptable expiration error, and in the same
// time it should not be too short to avoid overwhelming of the system
const timestampUpdatePeriod = 300 * time.Millisecond

// cache status
// unreachable: when cache is bypass, or invalid
// hit: cache is served
// miss: do not have cache record
const (
	cacheUnreachable = "unreachable"
	cacheHit         = "hit"
	cacheMiss        = "miss"
)

// New creates a new middleware handler
func New(config ...Config) lightning.Handler {
	// Set default config
	cfg := configDefault(config...)

	// Nothing to cache
	if int(cfg.Expiration.Seconds()) < 0 {
		return func(req *lightning.Request, res *lightning.Response) error {
			return req.Next()
		}
	}

	var (
		// Cache settings
		mux       = &sync.RWMutex{}
		timestamp = uint64(time.Now().Unix())
	)
	// Create manager to simplify storage operations ( see manager.go )
	manager := newManager(cfg.Storage)

	// Update timestamp in the configured interval
	go func() {
		for {
			atomic.StoreUint64(&timestamp, uint64(time.Now().Unix()))
			time.Sleep(timestampUpdatePeriod)
		}
	}()

	// Return new handler
	return func(req *lightning.Request, res *lightning.Response) error {
		// Only cache GET and HEAD methods
		if req.Method() != lightning.MethodGet && req.Method() != lightning.MethodHead {
			res.Header.Set(cfg.CacheHeader, cacheUnreachable)
			return req.Next()
		}

		// Get key from request
		// TODO(allocation optimization): try to minimize the allocation from 2 to 1
		key := cfg.KeyGenerator(req, res) + "_" + req.Method()

		// Get entry from pool
		e := manager.get(key)

		// Lock entry and unlock when finished
		mux.Lock()
		defer mux.Unlock()

		// Get timestamp
		ts := atomic.LoadUint64(&timestamp)

		if e.exp != 0 && ts >= e.exp {
			// Check if entry is expired
			manager.delete(key)
			// External storage saves body data with different key
			if cfg.Storage != nil {
				manager.delete(key + "_body")
			}
		} else if e.exp != 0 {
			// Separate body value to avoid msgp serialization
			// We can store raw bytes with Storage 👍
			if cfg.Storage != nil {
				e.body = manager.getRaw(key + "_body")
			}
			// Set response headers from cache
			res.Ctx().Response().SetBodyRaw(e.body)
			res.Ctx().Response().SetStatusCode(e.status)
			res.Ctx().Response().Header.SetContentTypeBytes(e.ctype)
			if len(e.cencoding) > 0 {
				res.Ctx().Response().Header.SetBytesV(lightning.HeaderContentEncoding, e.cencoding)
			}
			// Set Cache-Control header if enabled
			if cfg.CacheControl {
				maxAge := strconv.FormatUint(e.exp-ts, 10)
				res.Ctx().Set(lightning.HeaderCacheControl, "public, max-age="+maxAge)
			}

			res.Header.Set(cfg.CacheHeader, cacheHit)

			// Return response
			return nil
		}

		// Continue stack, return err to Fiber if exist
		if err := req.Next(); err != nil {
			return err
		}

		// Don't cache response if Next returns true
		if cfg.Next != nil && cfg.Next(req, res) {
			res.Header.Set(cfg.CacheHeader, cacheUnreachable)
			return nil
		}

		// Cache response
		e.body = utils.CopyBytes(res.Ctx().Response().Body())
		e.status = res.Ctx().Response().StatusCode()
		e.ctype = utils.CopyBytes(res.Ctx().Response().Header.ContentType())
		e.cencoding = utils.CopyBytes(res.Ctx().Response().Header.Peek(lightning.HeaderContentEncoding))

		// default cache expiration
		expiration := uint64(cfg.Expiration.Seconds())
		// Calculate expiration by response header or other setting
		if cfg.ExpirationGenerator != nil {
			expiration = uint64(cfg.ExpirationGenerator(req, res, &cfg).Seconds())
		}
		e.exp = ts + expiration

		// For external Storage we store raw body separated
		if cfg.Storage != nil {
			manager.setRaw(key+"_body", e.body, cfg.Expiration)
			// avoid body msgp encoding
			e.body = nil
			manager.set(key, e, cfg.Expiration)
			manager.release(e)
		} else {
			// Store entry in memory
			manager.set(key, e, cfg.Expiration)
		}

		res.Header.Set(cfg.CacheHeader, cacheMiss)

		// Finish response
		return nil
	}
}
