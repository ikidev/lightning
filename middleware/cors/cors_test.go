package cors

import (
	"net/http/httptest"
	"testing"

	"github.com/ikidev/lightning"
	"github.com/ikidev/lightning/utils"
	"github.com/valyala/fasthttp"
)

func Test_CORS_Defaults(t *testing.T) {
	app := lightning.New()
	app.Use(New())

	testDefaultOrEmptyConfig(t, app)
}

func Test_CORS_Empty_Config(t *testing.T) {
	app := lightning.New()
	app.Use(New(Config{}))

	testDefaultOrEmptyConfig(t, app)
}

func testDefaultOrEmptyConfig(t *testing.T, app *lightning.App) {
	t.Helper()

	h := app.Handler()

	// Test default GET response headers
	ctx := &fasthttp.RequestCtx{}
	ctx.Request.Header.SetMethod(lightning.MethodGet)
	h(ctx)

	utils.AssertEqual(t, "*", string(ctx.Response.Header.Peek(lightning.HeaderAccessControlAllowOrigin)))
	utils.AssertEqual(t, "", string(ctx.Response.Header.Peek(lightning.HeaderAccessControlAllowCredentials)))
	utils.AssertEqual(t, "", string(ctx.Response.Header.Peek(lightning.HeaderAccessControlExposeHeaders)))

	// Test default OPTIONS (preflight) response headers
	ctx = &fasthttp.RequestCtx{}
	ctx.Request.Header.SetMethod(lightning.MethodOptions)
	h(ctx)

	utils.AssertEqual(t, "GET,POST,HEAD,PUT,DELETE,PATCH", string(ctx.Response.Header.Peek(lightning.HeaderAccessControlAllowMethods)))
	utils.AssertEqual(t, "", string(ctx.Response.Header.Peek(lightning.HeaderAccessControlAllowHeaders)))
	utils.AssertEqual(t, "", string(ctx.Response.Header.Peek(lightning.HeaderAccessControlMaxAge)))
}

// go test -run -v Test_CORS_Wildcard
func Test_CORS_Wildcard(t *testing.T) {
	// New fiber instance
	app := lightning.New()
	// OPTIONS (preflight) response headers when AllowOrigins is *
	app.Use(New(Config{
		AllowOrigins:     "*",
		AllowCredentials: true,
		MaxAge:           3600,
		ExposeHeaders:    "X-FHRequest-ID",
		AllowHeaders:     "Authentication",
	}))
	// Get handler pointer
	handler := app.Handler()

	// Make request
	ctx := &fasthttp.RequestCtx{}
	ctx.Request.SetRequestURI("/")
	ctx.Request.Header.Set(lightning.HeaderOrigin, "localhost")
	ctx.Request.Header.SetMethod(lightning.MethodOptions)

	// Perform request
	handler(ctx)

	// Check result
	utils.AssertEqual(t, "localhost", string(ctx.Response.Header.Peek(lightning.HeaderAccessControlAllowOrigin)))
	utils.AssertEqual(t, "true", string(ctx.Response.Header.Peek(lightning.HeaderAccessControlAllowCredentials)))
	utils.AssertEqual(t, "3600", string(ctx.Response.Header.Peek(lightning.HeaderAccessControlMaxAge)))
	utils.AssertEqual(t, "Authentication", string(ctx.Response.Header.Peek(lightning.HeaderAccessControlAllowHeaders)))

	// Test non OPTIONS (preflight) response headers
	ctx = &fasthttp.RequestCtx{}
	ctx.Request.Header.SetMethod(lightning.MethodGet)
	handler(ctx)

	utils.AssertEqual(t, "true", string(ctx.Response.Header.Peek(lightning.HeaderAccessControlAllowCredentials)))
	utils.AssertEqual(t, "X-FHRequest-ID", string(ctx.Response.Header.Peek(lightning.HeaderAccessControlExposeHeaders)))
}

// go test -run -v Test_CORS_Subdomain
func Test_CORS_Subdomain(t *testing.T) {
	// New fiber instance
	app := lightning.New()
	// OPTIONS (preflight) response headers when AllowOrigins is set to a subdomain
	app.Use("/", New(Config{AllowOrigins: "http://*.example.com"}))

	// Get handler pointer
	handler := app.Handler()

	// Make request with disallowed origin
	ctx := &fasthttp.RequestCtx{}
	ctx.Request.SetRequestURI("/")
	ctx.Request.Header.SetMethod(lightning.MethodOptions)
	ctx.Request.Header.Set(lightning.HeaderOrigin, "http://google.com")

	// Perform request
	handler(ctx)

	// Allow-Origin header should be "" because http://google.com does not satisfy http://*.example.com
	utils.AssertEqual(t, "", string(ctx.Response.Header.Peek(lightning.HeaderAccessControlAllowOrigin)))

	ctx.Request.Reset()
	ctx.Response.Reset()

	// Make request with allowed origin
	ctx.Request.SetRequestURI("/")
	ctx.Request.Header.SetMethod(lightning.MethodOptions)
	ctx.Request.Header.Set(lightning.HeaderOrigin, "http://test.example.com")

	handler(ctx)

	utils.AssertEqual(t, "http://test.example.com", string(ctx.Response.Header.Peek(lightning.HeaderAccessControlAllowOrigin)))
}

func Test_CORS_AllowOriginScheme(t *testing.T) {
	tests := []struct {
		reqOrigin, pattern string
		shouldAllowOrigin  bool
	}{
		{
			pattern:           "http://example.com",
			reqOrigin:         "http://example.com",
			shouldAllowOrigin: true,
		},
		{
			pattern:           "https://example.com",
			reqOrigin:         "https://example.com",
			shouldAllowOrigin: true,
		},
		{
			pattern:           "http://example.com",
			reqOrigin:         "https://example.com",
			shouldAllowOrigin: false,
		},
		{
			pattern:           "http://*.example.com",
			reqOrigin:         "http://aaa.example.com",
			shouldAllowOrigin: true,
		},
		{
			pattern:           "http://*.example.com",
			reqOrigin:         "http://bbb.aaa.example.com",
			shouldAllowOrigin: true,
		},
		{
			pattern:           "http://*.aaa.example.com",
			reqOrigin:         "http://bbb.aaa.example.com",
			shouldAllowOrigin: true,
		},
		{
			pattern:           "http://*.example.com:8080",
			reqOrigin:         "http://aaa.example.com:8080",
			shouldAllowOrigin: true,
		},
		{
			pattern:           "http://example.com",
			reqOrigin:         "http://gofiber.com",
			shouldAllowOrigin: false,
		},
		{
			pattern:           "http://*.aaa.example.com",
			reqOrigin:         "http://ccc.bbb.example.com",
			shouldAllowOrigin: false,
		},
		{
			pattern: "http://*.example.com",
			reqOrigin: `http://1234567890.1234567890.1234567890.1234567890.1234567890.1234567890.1234567890.1234567890.1234567890.1234567890\
		  .1234567890.1234567890.1234567890.1234567890.1234567890.1234567890.1234567890.1234567890.1234567890.1234567890\
		  .1234567890.1234567890.1234567890.1234567890.1234567890.1234567890.1234567890.1234567890.1234567890.1234567890\
			.1234567890.1234567890.1234567890.1234567890.1234567890.1234567890.1234567890.1234567890.1234567890.1234567890.example.com`,
			shouldAllowOrigin: false,
		},
		{
			pattern:           "http://example.com",
			reqOrigin:         "http://ccc.bbb.example.com",
			shouldAllowOrigin: false,
		},
		{
			pattern:           "https://*--aaa.bbb.com",
			reqOrigin:         "https://prod-preview--aaa.bbb.com",
			shouldAllowOrigin: false,
		},
		{
			pattern:           "http://*.example.com",
			reqOrigin:         "http://ccc.bbb.example.com",
			shouldAllowOrigin: true,
		},
		{
			pattern:           "http://foo.[a-z]*.example.com",
			reqOrigin:         "http://ccc.bbb.example.com",
			shouldAllowOrigin: false,
		},
	}

	for _, tt := range tests {
		app := lightning.New()
		app.Use("/", New(Config{AllowOrigins: tt.pattern}))

		handler := app.Handler()

		ctx := &fasthttp.RequestCtx{}
		ctx.Request.SetRequestURI("/")
		ctx.Request.Header.SetMethod(lightning.MethodOptions)
		ctx.Request.Header.Set(lightning.HeaderOrigin, tt.reqOrigin)

		handler(ctx)

		if tt.shouldAllowOrigin {
			utils.AssertEqual(t, tt.reqOrigin, string(ctx.Response.Header.Peek(lightning.HeaderAccessControlAllowOrigin)))
		} else {
			utils.AssertEqual(t, "", string(ctx.Response.Header.Peek(lightning.HeaderAccessControlAllowOrigin)))
		}
	}
}

// go test -run Test_CORS_Next
func Test_CORS_Next(t *testing.T) {
	app := lightning.New()
	app.Use(New(Config{
		Next: func(_ *lightning.Request, _ *lightning.Response) bool {
			return true
		},
	}))

	resp, err := app.Test(httptest.NewRequest("GET", "/", nil))
	utils.AssertEqual(t, nil, err)
	utils.AssertEqual(t, lightning.StatusNotFound, resp.StatusCode)
}
