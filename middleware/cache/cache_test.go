// Special thanks to @codemicro for moving this to fiber core
// Original middleware: github.com/codemicro/fiber-cache
package cache

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"

	"github.com/ikidev/lightning"
	"github.com/ikidev/lightning/internal/storage/memory"
	"github.com/ikidev/lightning/utils"
	"github.com/valyala/fasthttp"
)

func Test_Cache_CacheControl(t *testing.T) {
	app := lightning.New()

	app.Use(New(Config{
		CacheControl: true,
		Expiration:   10 * time.Second,
	}))

	app.Get("/", func(req *lightning.Request, res *lightning.Response) error {
		return res.String("Hello, World!")
	})

	_, err := app.Test(httptest.NewRequest("GET", "/", nil))
	utils.AssertEqual(t, nil, err)

	resp, err := app.Test(httptest.NewRequest("GET", "/", nil))
	utils.AssertEqual(t, nil, err)
	utils.AssertEqual(t, "public, max-age=10", resp.Header.Get(lightning.HeaderCacheControl))
}

func Test_Cache_Expired(t *testing.T) {
	t.Parallel()

	app := lightning.New()
	app.Use(New(Config{Expiration: 2 * time.Second}))

	app.Get("/", func(req *lightning.Request, res *lightning.Response) error {
		return res.String(fmt.Sprintf("%d", time.Now().UnixNano()))
	})

	resp, err := app.Test(httptest.NewRequest("GET", "/", nil))
	utils.AssertEqual(t, nil, err)
	body, err := ioutil.ReadAll(resp.Body)
	utils.AssertEqual(t, nil, err)

	// Sleep until the cache is expired
	time.Sleep(3 * time.Second)

	respCached, err := app.Test(httptest.NewRequest("GET", "/", nil))
	utils.AssertEqual(t, nil, err)
	bodyCached, err := ioutil.ReadAll(respCached.Body)
	utils.AssertEqual(t, nil, err)

	if bytes.Equal(body, bodyCached) {
		t.Errorf("Cache should have expired: %s, %s", body, bodyCached)
	}

	// Next response should be also cached
	respCachedNextRound, err := app.Test(httptest.NewRequest("GET", "/", nil))
	utils.AssertEqual(t, nil, err)
	bodyCachedNextRound, err := ioutil.ReadAll(respCachedNextRound.Body)
	utils.AssertEqual(t, nil, err)

	if !bytes.Equal(bodyCachedNextRound, bodyCached) {
		t.Errorf("Cache should not have expired: %s, %s", bodyCached, bodyCachedNextRound)
	}
}

func Test_Cache(t *testing.T) {
	app := lightning.New()
	app.Use(New())

	app.Get("/", func(req *lightning.Request, res *lightning.Response) error {
		now := fmt.Sprintf("%d", time.Now().UnixNano())
		return res.String(now)
	})

	req := httptest.NewRequest("GET", "/", nil)
	resp, err := app.Test(req)
	utils.AssertEqual(t, nil, err)

	cachedReq := httptest.NewRequest("GET", "/", nil)
	cachedResp, err := app.Test(cachedReq)
	utils.AssertEqual(t, nil, err)

	body, err := ioutil.ReadAll(resp.Body)
	utils.AssertEqual(t, nil, err)
	cachedBody, err := ioutil.ReadAll(cachedResp.Body)
	utils.AssertEqual(t, nil, err)

	utils.AssertEqual(t, cachedBody, body)
}

func Test_Cache_WithSeveralRequests(t *testing.T) {
	app := lightning.New()

	app.Use(New(Config{
		CacheControl: true,
		Expiration:   10 * time.Second,
	}))

	app.Get("/:id", func(req *lightning.Request, res *lightning.Response) error {
		return res.String(req.UrlParam("id"))
	})

	for runs := 0; runs < 10; runs++ {
		for i := 0; i < 10; i++ {
			func(id int) {
				rsp, err := app.Test(httptest.NewRequest(http.MethodGet, fmt.Sprintf("/%d", id), nil))
				utils.AssertEqual(t, nil, err)

				defer rsp.Body.Close()

				idFromServ, err := ioutil.ReadAll(rsp.Body)
				utils.AssertEqual(t, nil, err)

				a, err := strconv.Atoi(string(idFromServ))
				utils.AssertEqual(t, nil, err)

				// SomeTimes,The id is not equal with a
				utils.AssertEqual(t, id, a)
			}(i)
		}
	}
}

func Test_Cache_Invalid_Expiration(t *testing.T) {
	app := lightning.New()
	cache := New(Config{Expiration: 0 * time.Second})
	app.Use(cache)

	app.Get("/", func(req *lightning.Request, res *lightning.Response) error {
		now := fmt.Sprintf("%d", time.Now().UnixNano())
		return res.String(now)
	})

	req := httptest.NewRequest("GET", "/", nil)
	resp, err := app.Test(req)
	utils.AssertEqual(t, nil, err)

	cachedReq := httptest.NewRequest("GET", "/", nil)
	cachedResp, err := app.Test(cachedReq)
	utils.AssertEqual(t, nil, err)

	body, err := ioutil.ReadAll(resp.Body)
	utils.AssertEqual(t, nil, err)
	cachedBody, err := ioutil.ReadAll(cachedResp.Body)
	utils.AssertEqual(t, nil, err)

	utils.AssertEqual(t, cachedBody, body)
}

func Test_Cache_Invalid_Method(t *testing.T) {
	app := lightning.New()

	app.Use(New())

	app.Post("/", func(req *lightning.Request, res *lightning.Response) error {
		return res.String(req.Query("cache"))
	})

	app.Get("/get", func(req *lightning.Request, res *lightning.Response) error {
		return res.String(req.Query("cache"))
	})

	resp, err := app.Test(httptest.NewRequest("POST", "/?cache=123", nil))
	utils.AssertEqual(t, nil, err)
	body, err := ioutil.ReadAll(resp.Body)
	utils.AssertEqual(t, nil, err)
	utils.AssertEqual(t, "123", string(body))

	resp, err = app.Test(httptest.NewRequest("POST", "/?cache=12345", nil))
	utils.AssertEqual(t, nil, err)
	body, err = ioutil.ReadAll(resp.Body)
	utils.AssertEqual(t, nil, err)
	utils.AssertEqual(t, "12345", string(body))

	resp, err = app.Test(httptest.NewRequest("GET", "/get?cache=123", nil))
	utils.AssertEqual(t, nil, err)
	body, err = ioutil.ReadAll(resp.Body)
	utils.AssertEqual(t, nil, err)
	utils.AssertEqual(t, "123", string(body))

	resp, err = app.Test(httptest.NewRequest("GET", "/get?cache=12345", nil))
	utils.AssertEqual(t, nil, err)
	body, err = ioutil.ReadAll(resp.Body)
	utils.AssertEqual(t, nil, err)
	utils.AssertEqual(t, "123", string(body))
}

func Test_Cache_NothingToCache(t *testing.T) {
	app := lightning.New()

	app.Use(New(Config{Expiration: -(time.Second * 1)}))

	app.Get("/", func(req *lightning.Request, res *lightning.Response) error {
		return res.String(time.Now().String())
	})

	resp, err := app.Test(httptest.NewRequest("GET", "/", nil))
	utils.AssertEqual(t, nil, err)
	body, err := ioutil.ReadAll(resp.Body)
	utils.AssertEqual(t, nil, err)

	time.Sleep(500 * time.Millisecond)

	respCached, err := app.Test(httptest.NewRequest("GET", "/", nil))
	utils.AssertEqual(t, nil, err)
	bodyCached, err := ioutil.ReadAll(respCached.Body)
	utils.AssertEqual(t, nil, err)

	if bytes.Equal(body, bodyCached) {
		t.Errorf("Cache should have expired: %s, %s", body, bodyCached)
	}
}

func Test_Cache_CustomNext(t *testing.T) {
	app := lightning.New()

	app.Use(New(Config{
		Next: func(req *lightning.Request, res *lightning.Response) bool {
			return res.Ctx().Response().StatusCode() != lightning.StatusOK
		},
		CacheControl: true,
	}))

	app.Get("/", func(req *lightning.Request, res *lightning.Response) error {
		return res.String(time.Now().String())
	})

	app.Get("/error", func(req *lightning.Request, res *lightning.Response) error {
		return res.Status(lightning.StatusInternalServerError).String(time.Now().String())
	})

	resp, err := app.Test(httptest.NewRequest("GET", "/", nil))
	utils.AssertEqual(t, nil, err)
	body, err := ioutil.ReadAll(resp.Body)
	utils.AssertEqual(t, nil, err)

	respCached, err := app.Test(httptest.NewRequest("GET", "/", nil))
	utils.AssertEqual(t, nil, err)
	bodyCached, err := ioutil.ReadAll(respCached.Body)
	utils.AssertEqual(t, nil, err)
	utils.AssertEqual(t, true, bytes.Equal(body, bodyCached))
	utils.AssertEqual(t, true, respCached.Header.Get(lightning.HeaderCacheControl) != "")

	_, err = app.Test(httptest.NewRequest("GET", "/error", nil))
	utils.AssertEqual(t, nil, err)

	errRespCached, err := app.Test(httptest.NewRequest("GET", "/error", nil))
	utils.AssertEqual(t, nil, err)
	utils.AssertEqual(t, true, errRespCached.Header.Get(lightning.HeaderCacheControl) == "")
}

func Test_CustomKey(t *testing.T) {
	app := lightning.New()
	var called bool
	app.Use(New(Config{KeyGenerator: func(req *lightning.Request, res *lightning.Response) string {
		called = true
		return req.Path()
	}}))

	app.Get("/", func(req *lightning.Request, res *lightning.Response) error {
		return res.String("hi")
	})

	req := httptest.NewRequest("GET", "/", nil)
	_, err := app.Test(req)
	utils.AssertEqual(t, nil, err)
	utils.AssertEqual(t, true, called)
}

func Test_CustomExpiration(t *testing.T) {
	app := lightning.New()
	var called bool
	var newCacheTime int
	app.Use(New(Config{ExpirationGenerator: func(req *lightning.Request, res *lightning.Response, cfg *Config) time.Duration {
		called = true
		newCacheTime, _ = strconv.Atoi(res.Ctx().GetRespHeader("Cache-Time", "600"))
		return time.Second * time.Duration(newCacheTime)
	}}))

	app.Get("/", func(req *lightning.Request, res *lightning.Response) error {
		res.Ctx().Response().Header.Add("Cache-Time", "6000")
		return res.String("hi")
	})

	req := httptest.NewRequest("GET", "/", nil)
	_, err := app.Test(req)
	utils.AssertEqual(t, nil, err)
	utils.AssertEqual(t, true, called)
	utils.AssertEqual(t, 6000, newCacheTime)
}

func Test_CacheHeader(t *testing.T) {
	app := lightning.New()

	app.Use(New(Config{
		Expiration: 10 * time.Second,
		Next: func(req *lightning.Request, res *lightning.Response) bool {
			return res.Ctx().Response().StatusCode() != lightning.StatusOK
		},
	}))

	app.Get("/", func(req *lightning.Request, res *lightning.Response) error {
		return res.String("Hello, World!")
	})

	app.Post("/", func(req *lightning.Request, res *lightning.Response) error {
		return res.String(req.Query("cache"))
	})

	app.Get("/error", func(req *lightning.Request, res *lightning.Response) error {
		return res.Status(lightning.StatusInternalServerError).String(time.Now().String())
	})

	resp, err := app.Test(httptest.NewRequest("GET", "/", nil))
	utils.AssertEqual(t, nil, err)
	utils.AssertEqual(t, cacheMiss, resp.Header.Get("X-Cache"))

	resp, err = app.Test(httptest.NewRequest("GET", "/", nil))
	utils.AssertEqual(t, nil, err)
	utils.AssertEqual(t, cacheHit, resp.Header.Get("X-Cache"))

	resp, err = app.Test(httptest.NewRequest("POST", "/?cache=12345", nil))
	utils.AssertEqual(t, nil, err)
	utils.AssertEqual(t, cacheUnreachable, resp.Header.Get("X-Cache"))

	errRespCached, err := app.Test(httptest.NewRequest("GET", "/error", nil))
	utils.AssertEqual(t, nil, err)
	utils.AssertEqual(t, cacheUnreachable, errRespCached.Header.Get("X-Cache"))
}

func Test_Cache_WithHead(t *testing.T) {
	app := lightning.New()
	app.Use(New())

	app.Get("/", func(req *lightning.Request, res *lightning.Response) error {
		now := fmt.Sprintf("%d", time.Now().UnixNano())
		return res.String(now)
	})

	req := httptest.NewRequest("HEAD", "/", nil)
	resp, err := app.Test(req)
	utils.AssertEqual(t, cacheMiss, resp.Header.Get("X-Cache"))

	cachedReq := httptest.NewRequest("HEAD", "/", nil)
	cachedResp, err := app.Test(cachedReq)
	utils.AssertEqual(t, cacheHit, cachedResp.Header.Get("X-Cache"))

	body, err := ioutil.ReadAll(resp.Body)
	utils.AssertEqual(t, nil, err)
	cachedBody, err := ioutil.ReadAll(cachedResp.Body)
	utils.AssertEqual(t, nil, err)

	utils.AssertEqual(t, cachedBody, body)
}

func Test_Cache_WithHeadThenGet(t *testing.T) {
	app := lightning.New()
	app.Use(New())
	app.Get("/get", func(req *lightning.Request, res *lightning.Response) error {
		return res.String(req.Query("cache"))
	})

	headResp, err := app.Test(httptest.NewRequest("HEAD", "/head?cache=123", nil))
	utils.AssertEqual(t, nil, err)
	headBody, err := ioutil.ReadAll(headResp.Body)
	utils.AssertEqual(t, nil, err)
	utils.AssertEqual(t, "", string(headBody))
	utils.AssertEqual(t, cacheMiss, headResp.Header.Get("X-Cache"))

	headResp, err = app.Test(httptest.NewRequest("HEAD", "/head?cache=123", nil))
	utils.AssertEqual(t, nil, err)
	headBody, err = ioutil.ReadAll(headResp.Body)
	utils.AssertEqual(t, nil, err)
	utils.AssertEqual(t, "", string(headBody))
	utils.AssertEqual(t, cacheHit, headResp.Header.Get("X-Cache"))

	getResp, err := app.Test(httptest.NewRequest("GET", "/get?cache=123", nil))
	utils.AssertEqual(t, nil, err)
	getBody, err := ioutil.ReadAll(getResp.Body)
	utils.AssertEqual(t, nil, err)
	utils.AssertEqual(t, "123", string(getBody))
	utils.AssertEqual(t, cacheMiss, getResp.Header.Get("X-Cache"))

	getResp, err = app.Test(httptest.NewRequest("GET", "/get?cache=123", nil))
	utils.AssertEqual(t, nil, err)
	getBody, err = ioutil.ReadAll(getResp.Body)
	utils.AssertEqual(t, nil, err)
	utils.AssertEqual(t, "123", string(getBody))
	utils.AssertEqual(t, cacheHit, getResp.Header.Get("X-Cache"))
}

func Test_CustomCacheHeader(t *testing.T) {
	app := lightning.New()

	app.Use(New(Config{
		CacheHeader: "Cache-Status",
	}))

	app.Get("/", func(req *lightning.Request, res *lightning.Response) error {
		return res.String("Hello, World!")
	})

	resp, err := app.Test(httptest.NewRequest("GET", "/", nil))
	utils.AssertEqual(t, nil, err)
	utils.AssertEqual(t, cacheMiss, resp.Header.Get("Cache-Status"))
}

// go test -v -run=^$ -bench=Benchmark_Cache -benchmem -count=4
func Benchmark_Cache(b *testing.B) {
	app := lightning.New()

	app.Use(New())

	app.Get("/demo", func(req *lightning.Request, res *lightning.Response) error {
		data, _ := ioutil.ReadFile("../../.github/README.md")
		return res.Status(lightning.StatusTeapot).Bytes(data)
	})

	h := app.Handler()

	fCtx := &fasthttp.RequestCtx{}
	fCtx.Request.Header.SetMethod("GET")
	fCtx.Request.SetRequestURI("/demo")

	b.ReportAllocs()
	b.ResetTimer()

	for n := 0; n < b.N; n++ {
		h(fCtx)
	}

	utils.AssertEqual(b, lightning.StatusTeapot, fCtx.Response.Header.StatusCode())
	utils.AssertEqual(b, true, len(fCtx.Response.Body()) > 30000)
}

// go test -v -run=^$ -bench=Benchmark_Cache_Storage -benchmem -count=4
func Benchmark_Cache_Storage(b *testing.B) {
	app := lightning.New()

	app.Use(New(Config{
		Storage: memory.New(),
	}))

	app.Get("/demo", func(req *lightning.Request, res *lightning.Response) error {
		data, _ := ioutil.ReadFile("../../.github/README.md")
		return res.Status(lightning.StatusTeapot).Bytes(data)
	})

	h := app.Handler()

	fctx := &fasthttp.RequestCtx{}
	fctx.Request.Header.SetMethod("GET")
	fctx.Request.SetRequestURI("/demo")

	b.ReportAllocs()
	b.ResetTimer()

	for n := 0; n < b.N; n++ {
		h(fctx)
	}

	utils.AssertEqual(b, lightning.StatusTeapot, fctx.Response.Header.StatusCode())
	utils.AssertEqual(b, true, len(fctx.Response.Body()) > 30000)
}
