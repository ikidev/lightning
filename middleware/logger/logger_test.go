package logger

import (
	"errors"
	"fmt"
	"github.com/ikidev/lightning"
	"github.com/ikidev/lightning/internal/bytebufferpool"
	"github.com/ikidev/lightning/middleware/requestid"
	"github.com/ikidev/lightning/utils"
	"github.com/valyala/fasthttp"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"sync"
	"testing"
)

// go test -run Test_Logger
func Test_Logger(t *testing.T) {
	app := lightning.New()

	buf := bytebufferpool.Get()
	defer bytebufferpool.Put(buf)

	app.Use(New(Config{
		Format: "${error}",
		Output: buf,
	}))

	app.Get("/", func(_ *lightning.Request, _ *lightning.Response) error {
		return errors.New("some random error")
	})

	resp, err := app.Test(httptest.NewRequest("GET", "/", nil))
	utils.AssertEqual(t, nil, err)
	utils.AssertEqual(t, lightning.StatusInternalServerError, resp.StatusCode)
	utils.AssertEqual(t, "some random error", buf.String())
}

// go test -run Test_Logger_locals
func Test_Logger_locals(t *testing.T) {
	app := lightning.New()

	buf := bytebufferpool.Get()
	defer bytebufferpool.Put(buf)

	app.Use(New(Config{
		Format: "${locals:demo}",
		Output: buf,
	}))

	app.Get("/", func(req *lightning.Request, res *lightning.Response) error {
		req.Locals("demo", "johndoe")
		return res.Status(lightning.StatusOK).Send()
	})

	app.Get("/int", func(req *lightning.Request, res *lightning.Response) error {
		req.Locals("demo", 55)
		return res.Status(lightning.StatusOK).Send()
	})

	app.Get("/empty", func(req *lightning.Request, res *lightning.Response) error {
		return res.Status(lightning.StatusOK).Send()
	})

	resp, err := app.Test(httptest.NewRequest("GET", "/", nil))
	utils.AssertEqual(t, nil, err)
	utils.AssertEqual(t, lightning.StatusOK, resp.StatusCode)
	utils.AssertEqual(t, "johndoe", buf.String())

	buf.Reset()

	resp, err = app.Test(httptest.NewRequest("GET", "/int", nil))
	utils.AssertEqual(t, nil, err)
	utils.AssertEqual(t, lightning.StatusOK, resp.StatusCode)
	utils.AssertEqual(t, "55", buf.String())

	buf.Reset()

	resp, err = app.Test(httptest.NewRequest("GET", "/empty", nil))
	utils.AssertEqual(t, nil, err)
	utils.AssertEqual(t, lightning.StatusOK, resp.StatusCode)
	utils.AssertEqual(t, "", buf.String())
}

// go test -run Test_Logger_Next
func Test_Logger_Next(t *testing.T) {
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

// go test -run Test_Logger_ErrorTimeZone
func Test_Logger_ErrorTimeZone(t *testing.T) {
	app := lightning.New()
	app.Use(New(Config{
		TimeZone: "invalid",
	}))

	resp, err := app.Test(httptest.NewRequest("GET", "/", nil))
	utils.AssertEqual(t, nil, err)
	utils.AssertEqual(t, lightning.StatusNotFound, resp.StatusCode)
}

type fakeOutput int

func (o *fakeOutput) Write([]byte) (int, error) {
	*o++
	return 0, errors.New("fake output")
}

// go test -run Test_Logger_ErrorOutput
func Test_Logger_ErrorOutput(t *testing.T) {
	o := new(fakeOutput)
	app := lightning.New()
	app.Use(New(Config{
		Output: o,
	}))

	resp, err := app.Test(httptest.NewRequest("GET", "/", nil))
	utils.AssertEqual(t, nil, err)
	utils.AssertEqual(t, lightning.StatusNotFound, resp.StatusCode)

	utils.AssertEqual(t, 2, int(*o))
}

// go test -run Test_Logger_All
func Test_Logger_All(t *testing.T) {
	buf := bytebufferpool.Get()
	defer bytebufferpool.Put(buf)

	app := lightning.New()
	app.Use(New(Config{
		Format: "${pid}${reqHeaders}${referer}${protocol}${ip}${ips}${host}${url}${ua}${body}${route}${black}${red}${green}${yellow}${blue}${magenta}${cyan}${white}${reset}${error}${header:test}${query:test}${form:test}${cookie:test}${non}",
		Output: buf,
	}))

	resp, err := app.Test(httptest.NewRequest("GET", "/?foo=bar", nil))
	utils.AssertEqual(t, nil, err)
	utils.AssertEqual(t, lightning.StatusNotFound, resp.StatusCode)

	expected := fmt.Sprintf("%dHost=example.comhttp0.0.0.0example.com/?foo=bar/%s%s%s%s%s%s%s%s%s-", os.Getpid(), cBlack, cRed, cGreen, cYellow, cBlue, cMagenta, cCyan, cWhite, cReset)
	utils.AssertEqual(t, expected, buf.String())
}

// go test -run Test_Query_Params
func Test_Query_Params(t *testing.T) {
	buf := bytebufferpool.Get()
	defer bytebufferpool.Put(buf)

	app := lightning.New()
	app.Use(New(Config{
		Format: "${queryParams}",
		Output: buf,
	}))

	resp, err := app.Test(httptest.NewRequest("GET", "/?foo=bar&baz=moz", nil))
	utils.AssertEqual(t, nil, err)
	utils.AssertEqual(t, lightning.StatusNotFound, resp.StatusCode)

	expected := "foo=bar&baz=moz"
	utils.AssertEqual(t, expected, buf.String())
}

// go test -run Test_Response_Body
func Test_Response_Body(t *testing.T) {
	buf := bytebufferpool.Get()
	defer bytebufferpool.Put(buf)

	app := lightning.New()
	app.Use(New(Config{
		Format: "${resBody}",
		Output: buf,
	}))

	app.Get("/", func(req *lightning.Request, res *lightning.Response) error {
		return res.String("Sample response body")
	})

	app.Post("/test", func(req *lightning.Request, res *lightning.Response) error {
		return res.Bytes([]byte("Post in test"))
	})

	_, err := app.Test(httptest.NewRequest("GET", "/", nil))
	utils.AssertEqual(t, nil, err)

	expectedGetResponse := "Sample response body"
	utils.AssertEqual(t, expectedGetResponse, buf.String())

	buf.Reset() // Reset buffer to test POST

	_, err = app.Test(httptest.NewRequest("POST", "/test", nil))
	utils.AssertEqual(t, nil, err)

	expectedPostResponse := "Post in test"
	utils.AssertEqual(t, expectedPostResponse, buf.String())
}

// go test -run Test_Logger_AppendUint
func Test_Logger_AppendUint(t *testing.T) {
	app := lightning.New()

	buf := bytebufferpool.Get()
	defer bytebufferpool.Put(buf)

	app.Use(New(Config{
		Format: "${bytesReceived} ${bytesSent} ${status}",
		Output: buf,
	}))

	app.Get("/", func(req *lightning.Request, res *lightning.Response) error {
		return res.String("hello")
	})

	resp, err := app.Test(httptest.NewRequest("GET", "/", nil))
	utils.AssertEqual(t, nil, err)
	utils.AssertEqual(t, lightning.StatusOK, resp.StatusCode)
	utils.AssertEqual(t, "0 5 200", buf.String())
}

// go test -run Test_Logger_Data_Race -race
func Test_Logger_Data_Race(t *testing.T) {
	app := lightning.New()

	buf := bytebufferpool.Get()
	defer bytebufferpool.Put(buf)

	app.Use(New(ConfigDefault))

	app.Get("/", func(req *lightning.Request, res *lightning.Response) error {
		return res.String("hello")
	})

	var (
		resp1, resp2 *http.Response
		err1, err2   error
	)
	wg := &sync.WaitGroup{}
	wg.Add(1)
	go func() {
		resp1, err1 = app.Test(httptest.NewRequest("GET", "/", nil))
		wg.Done()
	}()
	resp2, err2 = app.Test(httptest.NewRequest("GET", "/", nil))
	wg.Wait()

	utils.AssertEqual(t, nil, err1)
	utils.AssertEqual(t, lightning.StatusOK, resp1.StatusCode)
	utils.AssertEqual(t, nil, err2)
	utils.AssertEqual(t, lightning.StatusOK, resp2.StatusCode)
}

// go test -v -run=^$ -bench=Benchmark_Logger -benchmem -count=4
func Benchmark_Logger(b *testing.B) {
	app := lightning.New()

	app.Use(New(Config{
		Format: "${bytesReceived} ${bytesSent} ${status}",
		Output: ioutil.Discard,
	}))
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
}

// go test -run Test_Response_Header
func Test_Response_Header(t *testing.T) {
	buf := bytebufferpool.Get()
	defer bytebufferpool.Put(buf)

	app := lightning.New()
	app.Use(requestid.New(requestid.Config{
		Next:       nil,
		Header:     lightning.HeaderXRequestID,
		Generator:  func() string { return "Hello fiber!" },
		ContextKey: "requestid",
	}))
	app.Use(New(Config{
		Format: fmt.Sprintf("${respHeader:%s}", lightning.HeaderXRequestID),
		Output: buf,
	}))
	app.Get("/", func(req *lightning.Request, res *lightning.Response) error {
		return res.String("Hello fiber!")
	})

	resp, err := app.Test(httptest.NewRequest("GET", "/", nil))

	utils.AssertEqual(t, nil, err)
	utils.AssertEqual(t, lightning.StatusOK, resp.StatusCode)
	utils.AssertEqual(t, "Hello fiber!", buf.String())
}

// go test -run Test_Req_Header
func Test_Req_Header(t *testing.T) {
	buf := bytebufferpool.Get()
	defer bytebufferpool.Put(buf)

	app := lightning.New()
	app.Use(New(Config{
		Format: "${reqHeader:test}",
		Output: buf,
	}))

	app.Get("/", func(req *lightning.Request, res *lightning.Response) error {
		return res.String(req.Header.Get("test"))
	})
	headerReq := httptest.NewRequest("GET", "/", nil)
	headerReq.Header.Add("test", "Hello fiber!")
	resp, err := app.Test(headerReq)

	utils.AssertEqual(t, nil, err)
	utils.AssertEqual(t, lightning.StatusOK, resp.StatusCode)
	utils.AssertEqual(t, "Hello fiber!", buf.String())
}

// go test -run Test_ReqHeader_Header
func Test_ReqHeader_Header(t *testing.T) {
	buf := bytebufferpool.Get()
	defer bytebufferpool.Put(buf)

	app := lightning.New()
	app.Use(New(Config{
		Format: "${reqHeader:test}",
		Output: buf,
	}))
	app.Get("/", func(req *lightning.Request, res *lightning.Response) error {
		return res.String("Hello fiber!")
	})
	reqHeaderReq := httptest.NewRequest("GET", "/", nil)
	reqHeaderReq.Header.Add("test", "Hello fiber!")
	resp, err := app.Test(reqHeaderReq)

	utils.AssertEqual(t, nil, err)
	utils.AssertEqual(t, lightning.StatusOK, resp.StatusCode)
	utils.AssertEqual(t, "Hello fiber!", buf.String())
}
