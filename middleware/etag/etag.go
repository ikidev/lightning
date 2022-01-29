package etag

import (
	"bytes"
	"hash/crc32"

	"github.com/ikidev/lightning"
	"github.com/ikidev/lightning/internal/bytebufferpool"
)

var (
	normalizedHeaderETag = []byte("Etag")
	weakPrefix           = []byte("W/")
)

// New creates a new middleware handler
func New(config ...Config) lightning.Handler {
	// Set default config
	cfg := configDefault(config...)

	crc32q := crc32.MakeTable(0xD5828281)

	// Return new handler
	return func(req *lightning.Request, res *lightning.Response) (err error) {
		// Don't execute middleware if Next returns true
		if cfg.Next != nil && cfg.Next(req, res) {
			return req.Next()
		}

		// Return err if next handler returns one
		if err = req.Next(); err != nil {
			return
		}

		// Don't generate ETags for invalid responses
		if res.Ctx().Response().StatusCode() != lightning.StatusOK {
			return
		}
		body := res.Ctx().Response().Body()
		// Skips ETag if no response body is present
		if len(body) == 0 {
			return
		}
		// Skip ETag if header is already present
		if res.Ctx().Response().Header.PeekBytes(normalizedHeaderETag) != nil {
			return
		}

		// Generate ETag for response
		bb := bytebufferpool.Get()
		defer bytebufferpool.Put(bb)

		// Enable weak tag
		if cfg.Weak {
			_, _ = bb.Write(weakPrefix)
		}

		_ = bb.WriteByte('"')
		bb.B = appendUint(bb.Bytes(), uint32(len(body)))
		_ = bb.WriteByte('-')
		bb.B = appendUint(bb.Bytes(), crc32.Checksum(body, crc32q))
		_ = bb.WriteByte('"')

		etag := bb.Bytes()

		// Get ETag header from request
		clientEtag := res.Ctx().Request().Header.Peek(lightning.HeaderIfNoneMatch)

		// Check if client's ETag is weak
		if bytes.HasPrefix(clientEtag, weakPrefix) {
			// Check if server's ETag is weak
			if bytes.Equal(clientEtag[2:], etag) || bytes.Equal(clientEtag[2:], etag[2:]) {
				// W/1 == 1 || W/1 == W/1
				res.Ctx().Context().ResetBody()

				return res.Status(lightning.StatusNotModified).Send()
			}
			// W/1 != W/2 || W/1 != 2
			res.Ctx().Response().Header.SetCanonical(normalizedHeaderETag, etag)

			return
		}

		if bytes.Contains(clientEtag, etag) {
			// 1 == 1
			res.Ctx().Context().ResetBody()

			return res.Status(lightning.StatusNotModified).Send()
		}
		// 1 != 2
		res.Ctx().Response().Header.SetCanonical(normalizedHeaderETag, etag)

		return
	}
}

// appendUint appends n to dst and returns the extended dst.
func appendUint(dst []byte, n uint32) []byte {
	var b [20]byte
	buf := b[:]
	i := len(buf)
	var q uint32
	for n >= 10 {
		i--
		q = n / 10
		buf[i] = '0' + byte(n-q*10)
		n = q
	}
	i--
	buf[i] = '0' + byte(n)

	dst = append(dst, buf[i:]...)
	return dst
}
