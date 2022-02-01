package proxy

import (
	"crypto/tls"
	"io/ioutil"
	"net"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/ikidev/lightning"
	"github.com/ikidev/lightning/internal/tlstest"
	"github.com/ikidev/lightning/utils"
)

func createProxyTestServer(handler lightning.Handler, t *testing.T) (*lightning.App, string) {
	t.Helper()

	target := lightning.New(lightning.Config{DisableStartupMessage: true})
	target.Get("/", handler)

	ln, err := net.Listen(lightning.NetworkTCP4, "127.0.0.1:0")
	utils.AssertEqual(t, nil, err)

	go func() {
		utils.AssertEqual(t, nil, target.Listener(ln))
	}()

	time.Sleep(2 * time.Second)
	addr := ln.Addr().String()

	return target, addr
}

// go test -run Test_Proxy_Empty_Host
func Test_Proxy_Empty_Upstream_Servers(t *testing.T) {
	t.Parallel()

	defer func() {
		if r := recover(); r != nil {
			utils.AssertEqual(t, "Servers cannot be empty", r)
		}
	}()
	app := lightning.New()
	app.Use(Balancer(Config{Servers: []string{}}))
}

// go test -run Test_Proxy_Next
func Test_Proxy_Next(t *testing.T) {
	t.Parallel()

	app := lightning.New()
	app.Use(Balancer(Config{
		Servers: []string{"127.0.0.1"},
		Next: func(_ *lightning.Request, _ *lightning.Response) bool {
			return true
		},
	}))

	resp, err := app.Test(httptest.NewRequest("GET", "/", nil))
	utils.AssertEqual(t, nil, err)
	utils.AssertEqual(t, lightning.StatusNotFound, resp.StatusCode)
}

// go test -run Test_Proxy
func Test_Proxy(t *testing.T) {
	t.Parallel()

	target, addr := createProxyTestServer(
		func(req *lightning.Request, res *lightning.Response) error {
			return res.Status(lightning.StatusTeapot).Send()
		}, t,
	)

	resp, err := target.Test(httptest.NewRequest("GET", "/", nil), 2000)
	utils.AssertEqual(t, nil, err)
	utils.AssertEqual(t, lightning.StatusTeapot, resp.StatusCode)

	app := lightning.New(lightning.Config{DisableStartupMessage: true})

	app.Use(Balancer(Config{Servers: []string{addr}}))

	req := httptest.NewRequest("GET", "/", nil)
	req.Host = addr
	resp, err = app.Test(req)
	utils.AssertEqual(t, nil, err)
	utils.AssertEqual(t, lightning.StatusTeapot, resp.StatusCode)
}

// go test -run Test_Proxy_Balancer_WithTlsConfig
func Test_Proxy_Balancer_WithTlsConfig(t *testing.T) {
	t.Parallel()

	serverTLSConf, _, err := tlstest.GetTLSConfigs()
	utils.AssertEqual(t, nil, err)

	ln, err := net.Listen(lightning.NetworkTCP4, "127.0.0.1:0")
	utils.AssertEqual(t, nil, err)

	ln = tls.NewListener(ln, serverTLSConf)

	app := lightning.New(lightning.Config{DisableStartupMessage: true})

	app.Get("/tls_balancer", func(req *lightning.Request, res *lightning.Response) error {
		return res.String("tls balancer")
	})

	addr := ln.Addr().String()
	clientTLSConf := &tls.Config{InsecureSkipVerify: true}

	// disable certificate verification in Balancer
	app.Use(Balancer(Config{
		Servers:   []string{addr},
		TlsConfig: clientTLSConf,
	}))

	go func() { utils.AssertEqual(t, nil, app.Listener(ln)) }()

	code, body, errs := lightning.Get("https://" + addr + "/tls_balancer").TLSConfig(clientTLSConf).String()

	utils.AssertEqual(t, 0, len(errs))
	utils.AssertEqual(t, lightning.StatusOK, code)
	utils.AssertEqual(t, "tls balancer", body)
}

// go test -run Test_Proxy_Forward
func Test_Proxy_Forward(t *testing.T) {
	t.Parallel()

	app := lightning.New()

	_, addr := createProxyTestServer(
		func(req *lightning.Request, res *lightning.Response) error { return res.String("forwarded") }, t,
	)

	app.Use(Forward("http://" + addr))

	resp, err := app.Test(httptest.NewRequest("GET", "/", nil))
	utils.AssertEqual(t, nil, err)
	utils.AssertEqual(t, lightning.StatusOK, resp.StatusCode)

	b, err := ioutil.ReadAll(resp.Body)
	utils.AssertEqual(t, nil, err)
	utils.AssertEqual(t, "forwarded", string(b))
}

// go test -run Test_Proxy_Forward_WithTlsConfig
func Test_Proxy_Forward_WithTlsConfig(t *testing.T) {
	t.Parallel()

	serverTLSConf, _, err := tlstest.GetTLSConfigs()
	utils.AssertEqual(t, nil, err)

	ln, err := net.Listen(lightning.NetworkTCP4, "127.0.0.1:0")
	utils.AssertEqual(t, nil, err)

	ln = tls.NewListener(ln, serverTLSConf)

	app := lightning.New(lightning.Config{DisableStartupMessage: true})

	app.Get("/tlsfwd", func(req *lightning.Request, res *lightning.Response) error {
		return res.String("tls forward")
	})

	addr := ln.Addr().String()
	clientTLSConf := &tls.Config{InsecureSkipVerify: true}

	// disable certificate verification
	WithTlsConfig(clientTLSConf)
	app.Use(Forward("https://" + addr + "/tlsfwd"))

	go func() { utils.AssertEqual(t, nil, app.Listener(ln)) }()

	code, body, errs := lightning.Get("https://" + addr).TLSConfig(clientTLSConf).String()

	utils.AssertEqual(t, 0, len(errs))
	utils.AssertEqual(t, lightning.StatusOK, code)
	utils.AssertEqual(t, "tls forward", body)
}

// go test -run Test_Proxy_Modify_Response
func Test_Proxy_Modify_Response(t *testing.T) {
	t.Parallel()

	_, addr := createProxyTestServer(func(req *lightning.Request, res *lightning.Response) error {
		return res.Status(500).String("not modified")
	}, t)

	app := lightning.New()
	app.Use(Balancer(Config{
		Servers: []string{addr},
		ModifyResponse: func(req *lightning.Request, res *lightning.Response) error {
			res.Status(lightning.StatusOK)
			return res.String("modified response")
		},
	}))

	resp, err := app.Test(httptest.NewRequest("GET", "/", nil))
	utils.AssertEqual(t, nil, err)
	utils.AssertEqual(t, lightning.StatusOK, resp.StatusCode)

	b, err := ioutil.ReadAll(resp.Body)
	utils.AssertEqual(t, nil, err)
	utils.AssertEqual(t, "modified response", string(b))
}

// go test -run Test_Proxy_Modify_Request
func Test_Proxy_Modify_Request(t *testing.T) {
	t.Parallel()

	_, addr := createProxyTestServer(func(req *lightning.Request, res *lightning.Response) error {
		b := req.FastHTTPRequest().Body()
		return res.String(string(b))
	}, t)

	app := lightning.New()
	app.Use(Balancer(Config{
		Servers: []string{addr},
		ModifyRequest: func(req *lightning.Request, res *lightning.Response) error {
			req.FastHTTPRequest().SetBody([]byte("modified request"))
			return nil
		},
	}))

	resp, err := app.Test(httptest.NewRequest("GET", "/", nil))
	utils.AssertEqual(t, nil, err)
	utils.AssertEqual(t, lightning.StatusOK, resp.StatusCode)

	b, err := ioutil.ReadAll(resp.Body)
	utils.AssertEqual(t, nil, err)
	utils.AssertEqual(t, "modified request", string(b))
}

// go test -run Test_Proxy_Timeout_Slow_Server
func Test_Proxy_Timeout_Slow_Server(t *testing.T) {
	t.Parallel()

	_, addr := createProxyTestServer(func(req *lightning.Request, res *lightning.Response) error {
		time.Sleep(2 * time.Second)
		return res.String("fiber is awesome")
	}, t)

	app := lightning.New()
	app.Use(Balancer(Config{
		Servers: []string{addr},
		Timeout: 3 * time.Second,
	}))

	resp, err := app.Test(httptest.NewRequest("GET", "/", nil), 5000)
	utils.AssertEqual(t, nil, err)
	utils.AssertEqual(t, lightning.StatusOK, resp.StatusCode)

	b, err := ioutil.ReadAll(resp.Body)
	utils.AssertEqual(t, nil, err)
	utils.AssertEqual(t, "fiber is awesome", string(b))
}

// go test -run Test_Proxy_With_Timeout
func Test_Proxy_With_Timeout(t *testing.T) {
	t.Parallel()

	_, addr := createProxyTestServer(func(req *lightning.Request, res *lightning.Response) error {
		time.Sleep(1 * time.Second)
		return res.String("fiber is awesome")
	}, t)

	app := lightning.New()
	app.Use(Balancer(Config{
		Servers: []string{addr},
		Timeout: 100 * time.Millisecond,
	}))

	resp, err := app.Test(httptest.NewRequest("GET", "/", nil), 2000)
	utils.AssertEqual(t, nil, err)
	utils.AssertEqual(t, lightning.StatusInternalServerError, resp.StatusCode)

	b, err := ioutil.ReadAll(resp.Body)
	utils.AssertEqual(t, nil, err)
	utils.AssertEqual(t, "timeout", string(b))
}

// go test -run Test_Proxy_Buffer_Size_Response
func Test_Proxy_Buffer_Size_Response(t *testing.T) {
	t.Parallel()

	_, addr := createProxyTestServer(func(req *lightning.Request, res *lightning.Response) error {
		long := strings.Join(make([]string, 5000), "-")
		res.Header.Set("Very-Long-Header", long)
		return res.String("ok")
	}, t)

	app := lightning.New()
	app.Use(Balancer(Config{Servers: []string{addr}}))

	resp, err := app.Test(httptest.NewRequest("GET", "/", nil))
	utils.AssertEqual(t, nil, err)
	utils.AssertEqual(t, lightning.StatusInternalServerError, resp.StatusCode)

	app = lightning.New()
	app.Use(Balancer(Config{
		Servers:        []string{addr},
		ReadBufferSize: 1024 * 8,
	}))

	resp, err = app.Test(httptest.NewRequest("GET", "/", nil))
	utils.AssertEqual(t, nil, err)
	utils.AssertEqual(t, lightning.StatusOK, resp.StatusCode)
}
