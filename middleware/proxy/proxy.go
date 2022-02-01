package proxy

import (
	"crypto/tls"
	"fmt"
	"net/url"
	"strings"

	"github.com/ikidev/lightning"
	"github.com/ikidev/lightning/utils"
	"github.com/valyala/fasthttp"
)

// New is deprecated
func New(config Config) lightning.Handler {
	fmt.Println("proxy.New is deprecated, please use proxy.Balancer instead")
	return Balancer(config)
}

// Balancer creates a load balancer among multiple upstream servers
func Balancer(config Config) lightning.Handler {
	// Set default config
	cfg := configDefault(config)

	// Load balanced client
	var lbc fasthttp.LBClient
	// Set timeout
	lbc.Timeout = cfg.Timeout

	// Scheme must be provided, falls back to http/https
	for _, server := range cfg.Servers {
		if !strings.HasPrefix(server, "http") {
			server = "http://" + server
			if cfg.SupportsHTTPS {
				server = "https://" + server
			}

		}

		u, err := url.Parse(server)
		if err != nil {
			panic(err)
		}

		client := &fasthttp.HostClient{
			NoDefaultUserAgentHeader: true,
			DisablePathNormalizing:   true,
			Addr:                     u.Host,

			ReadBufferSize:  config.ReadBufferSize,
			WriteBufferSize: config.WriteBufferSize,

			TLSConfig: config.TlsConfig,
		}

		lbc.Clients = append(lbc.Clients, client)
	}

	// Return new handler
	return func(req *lightning.Request, res *lightning.Response) (err error) {
		// Don't execute middleware if Next returns true
		if cfg.Next != nil && cfg.Next(req, res) {
			return req.Next()
		}

		// Set request and response (FastHTTP request/response)
		fReq := req.FastHTTPRequest()
		fRes := res.FastHTTPResponse()

		// Don't proxy "Connection" header
		fReq.Header.Del(lightning.HeaderConnection)

		// Modify request
		if cfg.ModifyRequest != nil {
			if err = cfg.ModifyRequest(req, res); err != nil {
				return err
			}
		}

		fReq.SetRequestURI(utils.UnsafeString(fReq.RequestURI()))

		// Forward request
		if err = lbc.Do(fReq, fRes); err != nil {
			return err
		}

		// Don't proxy "Connection" header
		fRes.Header.Del(lightning.HeaderConnection)

		// Modify response
		if cfg.ModifyResponse != nil {
			if err = cfg.ModifyResponse(req, res); err != nil {
				return err
			}
		}

		// Return nil to end proxying if no error
		return nil
	}
}

var client = fasthttp.Client{
	NoDefaultUserAgentHeader: true,
	DisablePathNormalizing:   true,
}

// WithTlsConfig update http client with a user specified tls.config
// This function should be called before Do and Forward.
func WithTlsConfig(tlsConfig *tls.Config) {
	client.TLSConfig = tlsConfig
}

// Forward performs the given http request and fills the given http response.
// This method will return an fiber.Handler
func Forward(addr string) lightning.Handler {
	return func(req *lightning.Request, res *lightning.Response) error {
		return Do(req, res, addr)
	}
}

// Do performs the given http request and fills the given http response.
// This method can be used within a fiber.Handler
func Do(req *lightning.Request, res *lightning.Response, addr string) error {
	req.FastHTTPRequest().SetRequestURI(addr)
	req.FastHTTPRequest().Header.Del(lightning.HeaderConnection)
	if err := client.Do(req.FastHTTPRequest(), res.FastHTTPResponse()); err != nil {
		return err
	}
	res.FastHTTPResponse().Header.Del(lightning.HeaderConnection)
	return nil
}
