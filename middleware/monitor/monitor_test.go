package monitor

import (
	"bytes"
	"io/ioutil"
	"net/http/httptest"
	"testing"

	"github.com/ikidev/lightning"
	"github.com/ikidev/lightning/utils"
	"github.com/valyala/fasthttp"
)

func Test_Monitor_405(t *testing.T) {
	t.Parallel()

	app := lightning.New()

	app.Use("/", New())

	resp, err := app.Test(httptest.NewRequest(lightning.MethodPost, "/", nil))
	utils.AssertEqual(t, nil, err)
	utils.AssertEqual(t, 405, resp.StatusCode)
}

func Test_Monitor_Html(t *testing.T) {
	t.Parallel()

	app := lightning.New()

	app.Get("/", New())

	resp, err := app.Test(httptest.NewRequest(lightning.MethodGet, "/", nil))
	utils.AssertEqual(t, nil, err)
	utils.AssertEqual(t, 200, resp.StatusCode)
	utils.AssertEqual(t, lightning.MIMETextHTMLCharsetUTF8, resp.Header.Get(lightning.HeaderContentType))

	b, err := ioutil.ReadAll(resp.Body)
	utils.AssertEqual(t, nil, err)
	utils.AssertEqual(t, true, bytes.Contains(b, []byte("<title>Fiber Monitor</title>")))
}

// go test -run Test_Monitor_JSON -race
func Test_Monitor_JSON(t *testing.T) {
	t.Parallel()

	app := lightning.New()

	app.Get("/", New())

	req := httptest.NewRequest(lightning.MethodGet, "/", nil)
	req.Header.Set(lightning.HeaderAccept, lightning.MIMEApplicationJSON)
	resp, err := app.Test(req)
	utils.AssertEqual(t, nil, err)
	utils.AssertEqual(t, 200, resp.StatusCode)
	utils.AssertEqual(t, lightning.MIMEApplicationJSON, resp.Header.Get(lightning.HeaderContentType))

	b, err := ioutil.ReadAll(resp.Body)
	utils.AssertEqual(t, nil, err)
	utils.AssertEqual(t, true, bytes.Contains(b, []byte("pid")))
	utils.AssertEqual(t, true, bytes.Contains(b, []byte("os")))
}

// go test -v -run=^$ -bench=Benchmark_Monitor -benchmem -count=4
func Benchmark_Monitor(b *testing.B) {
	app := lightning.New()

	app.Get("/", New())

	h := app.Handler()

	fctx := &fasthttp.RequestCtx{}
	fctx.Request.Header.SetMethod("GET")
	fctx.Request.SetRequestURI("/")
	fctx.Request.Header.Set(lightning.HeaderAccept, lightning.MIMEApplicationJSON)

	b.ReportAllocs()
	b.ResetTimer()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			h(fctx)
		}
	})

	utils.AssertEqual(b, 200, fctx.Response.Header.StatusCode())
	utils.AssertEqual(b,
		lightning.MIMEApplicationJSON,
		string(fctx.Response.Header.Peek(lightning.HeaderContentType)))
}

// go test -run Test_Monitor_Next
func Test_Monitor_Next(t *testing.T) {
	t.Parallel()

	app := lightning.New()

	app.Use("/", New(Config{
		Next: func(_ *lightning.Request, _ *lightning.Response) bool {
			return true
		},
	}))

	resp, err := app.Test(httptest.NewRequest(lightning.MethodPost, "/", nil))
	utils.AssertEqual(t, nil, err)
	utils.AssertEqual(t, 404, resp.StatusCode)
}

// go test -run Test_Monitor_APIOnly -race
func Test_Monitor_APIOnly(t *testing.T) {
	//t.Parallel()

	app := lightning.New()

	app.Get("/", New(Config{
		APIOnly: true,
	}))

	req := httptest.NewRequest(lightning.MethodGet, "/", nil)
	req.Header.Set(lightning.HeaderAccept, lightning.MIMEApplicationJSON)
	resp, err := app.Test(req)
	utils.AssertEqual(t, nil, err)
	utils.AssertEqual(t, 200, resp.StatusCode)
	utils.AssertEqual(t, lightning.MIMEApplicationJSON, resp.Header.Get(lightning.HeaderContentType))

	b, err := ioutil.ReadAll(resp.Body)
	utils.AssertEqual(t, nil, err)
	utils.AssertEqual(t, true, bytes.Contains(b, []byte("pid")))
	utils.AssertEqual(t, true, bytes.Contains(b, []byte("os")))
}
