package etag

import (
	"bytes"
	"io/ioutil"
	"net/http/httptest"
	"testing"

	"github.com/ikidev/lightning"
	"github.com/ikidev/lightning/utils"
	"github.com/valyala/fasthttp"
)

// go test -run Test_ETag_Next
func Test_ETag_Next(t *testing.T) {
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

// go test -run Test_ETag_SkipError
func Test_ETag_SkipError(t *testing.T) {
	app := lightning.New()

	app.Use(New())

	app.Get("/", func(_ *lightning.Request, _ *lightning.Response) error {
		return lightning.ErrForbidden
	})

	resp, err := app.Test(httptest.NewRequest("GET", "/", nil))
	utils.AssertEqual(t, nil, err)
	utils.AssertEqual(t, lightning.StatusForbidden, resp.StatusCode)
}

// go test -run Test_ETag_NotStatusOK
func Test_ETag_NotStatusOK(t *testing.T) {
	app := lightning.New()

	app.Use(New())

	app.Get("/", func(req *lightning.Request, res *lightning.Response) error {
		return res.Status(lightning.StatusCreated).Send()
	})

	resp, err := app.Test(httptest.NewRequest("GET", "/", nil))
	utils.AssertEqual(t, nil, err)
	utils.AssertEqual(t, lightning.StatusCreated, resp.StatusCode)
}

// go test -run Test_ETag_NoBody
func Test_ETag_NoBody(t *testing.T) {
	app := lightning.New()

	app.Use(New())

	app.Get("/", func(_ *lightning.Request, _ *lightning.Response) error {
		return nil
	})

	resp, err := app.Test(httptest.NewRequest("GET", "/", nil))
	utils.AssertEqual(t, nil, err)
	utils.AssertEqual(t, lightning.StatusOK, resp.StatusCode)
}

// go test -run Test_ETag_NewEtag
func Test_ETag_NewEtag(t *testing.T) {
	t.Run("without HeaderIfNoneMatch", func(t *testing.T) {
		testETagNewEtag(t, false, false)
	})
	t.Run("with HeaderIfNoneMatch and not matched", func(t *testing.T) {
		testETagNewEtag(t, true, false)
	})
	t.Run("with HeaderIfNoneMatch and matched", func(t *testing.T) {
		testETagNewEtag(t, true, true)
	})
}

func testETagNewEtag(t *testing.T, headerIfNoneMatch, matched bool) {
	t.Helper()

	app := lightning.New()

	app.Use(New())

	app.Get("/", func(req *lightning.Request, res *lightning.Response) error {
		return res.String("Hello, World!")
	})

	req := httptest.NewRequest("GET", "/", nil)
	if headerIfNoneMatch {
		etag := `"non-match"`
		if matched {
			etag = `"13-1831710635"`
		}
		req.Header.Set(lightning.HeaderIfNoneMatch, etag)
	}

	resp, err := app.Test(req)
	utils.AssertEqual(t, nil, err)

	if !headerIfNoneMatch || !matched {
		utils.AssertEqual(t, lightning.StatusOK, resp.StatusCode)
		utils.AssertEqual(t, `"13-1831710635"`, resp.Header.Get(lightning.HeaderETag))
		return
	}

	if matched {
		utils.AssertEqual(t, lightning.StatusNotModified, resp.StatusCode)
		b, err := ioutil.ReadAll(resp.Body)
		utils.AssertEqual(t, nil, err)
		utils.AssertEqual(t, 0, len(b))
	}
}

// go test -run Test_ETag_WeakEtag
func Test_ETag_WeakEtag(t *testing.T) {
	t.Run("without HeaderIfNoneMatch", func(t *testing.T) {
		testETagWeakEtag(t, false, false)
	})
	t.Run("with HeaderIfNoneMatch and not matched", func(t *testing.T) {
		testETagWeakEtag(t, true, false)
	})
	t.Run("with HeaderIfNoneMatch and matched", func(t *testing.T) {
		testETagWeakEtag(t, true, true)
	})
}

func testETagWeakEtag(t *testing.T, headerIfNoneMatch, matched bool) {
	t.Helper()

	app := lightning.New()

	app.Use(New(Config{Weak: true}))

	app.Get("/", func(req *lightning.Request, res *lightning.Response) error {
		return res.String("Hello, World!")
	})

	req := httptest.NewRequest("GET", "/", nil)
	if headerIfNoneMatch {
		etag := `W/"non-match"`
		if matched {
			etag = `W/"13-1831710635"`
		}
		req.Header.Set(lightning.HeaderIfNoneMatch, etag)
	}

	resp, err := app.Test(req)
	utils.AssertEqual(t, nil, err)

	if !headerIfNoneMatch || !matched {
		utils.AssertEqual(t, lightning.StatusOK, resp.StatusCode)
		utils.AssertEqual(t, `W/"13-1831710635"`, resp.Header.Get(lightning.HeaderETag))
		return
	}

	if matched {
		utils.AssertEqual(t, lightning.StatusNotModified, resp.StatusCode)
		b, err := ioutil.ReadAll(resp.Body)
		utils.AssertEqual(t, nil, err)
		utils.AssertEqual(t, 0, len(b))
	}
}

// go test -run Test_ETag_CustomEtag
func Test_ETag_CustomEtag(t *testing.T) {
	t.Run("without HeaderIfNoneMatch", func(t *testing.T) {
		testETagCustomEtag(t, false, false)
	})
	t.Run("with HeaderIfNoneMatch and not matched", func(t *testing.T) {
		testETagCustomEtag(t, true, false)
	})
	t.Run("with HeaderIfNoneMatch and matched", func(t *testing.T) {
		testETagCustomEtag(t, true, true)
	})
}

func testETagCustomEtag(t *testing.T, headerIfNoneMatch, matched bool) {
	t.Helper()

	app := lightning.New()

	app.Use(New())

	app.Get("/", func(req *lightning.Request, res *lightning.Response) error {
		res.Header.Set(lightning.HeaderETag, `"custom"`)
		if bytes.Equal(req.Ctx().Request().Header.Peek(lightning.HeaderIfNoneMatch), []byte(`"custom"`)) {
			return res.Status(lightning.StatusNotModified).Send()
		}
		return res.String("Hello, World!")
	})

	req := httptest.NewRequest("GET", "/", nil)
	if headerIfNoneMatch {
		etag := `"non-match"`
		if matched {
			etag = `"custom"`
		}
		req.Header.Set(lightning.HeaderIfNoneMatch, etag)
	}

	resp, err := app.Test(req)
	utils.AssertEqual(t, nil, err)

	if !headerIfNoneMatch || !matched {
		utils.AssertEqual(t, lightning.StatusOK, resp.StatusCode)
		utils.AssertEqual(t, `"custom"`, resp.Header.Get(lightning.HeaderETag))
		return
	}

	if matched {
		utils.AssertEqual(t, lightning.StatusNotModified, resp.StatusCode)
		b, err := ioutil.ReadAll(resp.Body)
		utils.AssertEqual(t, nil, err)
		utils.AssertEqual(t, 0, len(b))
	}
}

// go test -run Test_ETag_CustomEtagPut
func Test_ETag_CustomEtagPut(t *testing.T) {
	app := lightning.New()

	app.Use(New())

	app.Put("/", func(req *lightning.Request, res *lightning.Response) error {
		req.Header.Set(lightning.HeaderETag, `"custom"`)
		if !bytes.Equal(req.Ctx().Request().Header.Peek(lightning.HeaderIfMatch), []byte(`"custom"`)) {
			return res.Status(lightning.StatusPreconditionFailed).Send()
		}
		return res.String("Hello, World!")
	})

	req := httptest.NewRequest("PUT", "/", nil)
	req.Header.Set(lightning.HeaderIfMatch, `"non-match"`)
	resp, err := app.Test(req)
	utils.AssertEqual(t, nil, err)
	utils.AssertEqual(t, lightning.StatusPreconditionFailed, resp.StatusCode)
}

// go test -v -run=^$ -bench=Benchmark_Etag -benchmem -count=4
func Benchmark_Etag(b *testing.B) {
	app := lightning.New()

	app.Use(New())

	app.Get("/", func(req *lightning.Request, res *lightning.Response) error {
		return res.String("Hello, World!")
	})

	h := app.Handler()

	fCtx := &fasthttp.RequestCtx{}
	fCtx.Request.Header.SetMethod("GET")
	fCtx.Request.SetRequestURI("/")

	b.ReportAllocs()
	b.ResetTimer()

	for n := 0; n < b.N; n++ {
		h(fCtx)
	}

	utils.AssertEqual(b, 200, fCtx.Response.Header.StatusCode())
	utils.AssertEqual(b, `"13-1831710635"`, string(fCtx.Response.Header.Peek(lightning.HeaderETag)))
}
