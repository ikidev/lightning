package lightning

import (
	"bytes"
	"crypto/tls"
	"encoding/base64"
	stdjson "encoding/json"
	"errors"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/ikidev/lightning/internal/tlstest"
	"github.com/ikidev/lightning/internal/uuid"
	"github.com/ikidev/lightning/utils"
	"github.com/valyala/fasthttp/fasthttputil"
)

func Test_Client_Invalid_URL(t *testing.T) {
	t.Parallel()

	ln := fasthttputil.NewInmemoryListener()

	app := New(Config{DisableStartupMessage: true})

	app.Get("/", func(req *Request, res *Response) error {
		return res.String(req.Hostname())
	})

	go func() { utils.AssertEqual(t, nil, app.Listener(ln)) }()

	a := Get("http://example.com\r\n\r\nGET /\r\n\r\n")

	a.HostClient.Dial = func(addr string) (net.Conn, error) { return ln.Dial() }

	_, body, errs := a.String()

	utils.AssertEqual(t, "", body)
	utils.AssertEqual(t, 1, len(errs))
	utils.AssertEqual(t, "missing required Host header in request", errs[0].Error())
}

func Test_Client_Unsupported_Protocol(t *testing.T) {
	t.Parallel()

	a := Get("ftp://example.com")

	_, body, errs := a.String()

	utils.AssertEqual(t, "", body)
	utils.AssertEqual(t, 1, len(errs))
	utils.AssertEqual(t, `unsupported protocol "ftp". http and https are supported`,
		errs[0].Error())
}

func Test_Client_Get(t *testing.T) {
	t.Parallel()

	ln := fasthttputil.NewInmemoryListener()

	app := New(Config{DisableStartupMessage: true})

	app.Get("/", func(req *Request, res *Response) error {
		return res.String(req.Hostname())
	})

	go func() { utils.AssertEqual(t, nil, app.Listener(ln)) }()

	for i := 0; i < 5; i++ {
		a := Get("http://example.com")

		a.HostClient.Dial = func(addr string) (net.Conn, error) { return ln.Dial() }

		code, body, errs := a.String()

		utils.AssertEqual(t, StatusOK, code)
		utils.AssertEqual(t, "example.com", body)
		utils.AssertEqual(t, 0, len(errs))
	}
}

func Test_Client_Head(t *testing.T) {
	t.Parallel()

	ln := fasthttputil.NewInmemoryListener()

	app := New(Config{DisableStartupMessage: true})

	app.Get("/", func(req *Request, res *Response) error {
		return res.String(req.Hostname())
	})

	go func() { utils.AssertEqual(t, nil, app.Listener(ln)) }()

	for i := 0; i < 5; i++ {
		a := Head("http://example.com")

		a.HostClient.Dial = func(addr string) (net.Conn, error) { return ln.Dial() }

		code, body, errs := a.String()

		utils.AssertEqual(t, StatusOK, code)
		utils.AssertEqual(t, "", body)
		utils.AssertEqual(t, 0, len(errs))
	}
}

func Test_Client_Post(t *testing.T) {
	t.Parallel()

	ln := fasthttputil.NewInmemoryListener()

	app := New(Config{DisableStartupMessage: true})

	app.Post("/", func(req *Request, res *Response) error {
		return res.Status(StatusCreated).String(req.FormValue("foo"))
	})

	go func() { utils.AssertEqual(t, nil, app.Listener(ln)) }()

	for i := 0; i < 5; i++ {
		args := AcquireArgs()

		args.Set("foo", "bar")

		a := Post("http://example.com").
			Form(args)

		a.HostClient.Dial = func(addr string) (net.Conn, error) { return ln.Dial() }

		code, body, errs := a.String()

		utils.AssertEqual(t, StatusCreated, code)
		utils.AssertEqual(t, "bar", body)
		utils.AssertEqual(t, 0, len(errs))

		ReleaseArgs(args)
	}
}

func Test_Client_Put(t *testing.T) {
	t.Parallel()

	ln := fasthttputil.NewInmemoryListener()

	app := New(Config{DisableStartupMessage: true})

	app.Put("/", func(req *Request, res *Response) error {
		return res.String(req.FormValue("foo"))
	})

	go func() { utils.AssertEqual(t, nil, app.Listener(ln)) }()

	for i := 0; i < 5; i++ {
		args := AcquireArgs()

		args.Set("foo", "bar")

		a := Put("http://example.com").
			Form(args)

		a.HostClient.Dial = func(addr string) (net.Conn, error) { return ln.Dial() }

		code, body, errs := a.String()

		utils.AssertEqual(t, StatusOK, code)
		utils.AssertEqual(t, "bar", body)
		utils.AssertEqual(t, 0, len(errs))

		ReleaseArgs(args)
	}
}

func Test_Client_Patch(t *testing.T) {
	t.Parallel()

	ln := fasthttputil.NewInmemoryListener()

	app := New(Config{DisableStartupMessage: true})

	app.Patch("/", func(req *Request, res *Response) error {
		return res.String(req.FormValue("foo"))
	})

	go func() { utils.AssertEqual(t, nil, app.Listener(ln)) }()

	for i := 0; i < 5; i++ {
		args := AcquireArgs()

		args.Set("foo", "bar")

		a := Patch("http://example.com").
			Form(args)

		a.HostClient.Dial = func(addr string) (net.Conn, error) { return ln.Dial() }

		code, body, errs := a.String()

		utils.AssertEqual(t, StatusOK, code)
		utils.AssertEqual(t, "bar", body)
		utils.AssertEqual(t, 0, len(errs))

		ReleaseArgs(args)
	}
}

func Test_Client_Delete(t *testing.T) {
	t.Parallel()

	ln := fasthttputil.NewInmemoryListener()

	app := New(Config{DisableStartupMessage: true})

	app.Delete("/", func(req *Request, res *Response) error {
		return res.Status(StatusNoContent).
			String("deleted")
	})

	go func() { utils.AssertEqual(t, nil, app.Listener(ln)) }()

	for i := 0; i < 5; i++ {
		args := AcquireArgs()

		a := Delete("http://example.com")

		a.HostClient.Dial = func(addr string) (net.Conn, error) { return ln.Dial() }

		code, body, errs := a.String()

		utils.AssertEqual(t, StatusNoContent, code)
		utils.AssertEqual(t, "", body)
		utils.AssertEqual(t, 0, len(errs))

		ReleaseArgs(args)
	}
}

func Test_Client_UserAgent(t *testing.T) {
	t.Parallel()

	ln := fasthttputil.NewInmemoryListener()

	app := New(Config{DisableStartupMessage: true})

	app.Get("/", func(req *Request, res *Response) error {
		return res.String(req.UserAgent())
	})

	go func() { utils.AssertEqual(t, nil, app.Listener(ln)) }()

	t.Run("default", func(t *testing.T) {
		for i := 0; i < 5; i++ {
			a := Get("http://example.com")

			a.HostClient.Dial = func(addr string) (net.Conn, error) { return ln.Dial() }

			code, body, errs := a.String()

			utils.AssertEqual(t, StatusOK, code)
			utils.AssertEqual(t, defaultUserAgent, body)
			utils.AssertEqual(t, 0, len(errs))
		}
	})

	t.Run("custom", func(t *testing.T) {
		for i := 0; i < 5; i++ {
			c := AcquireClient()
			c.UserAgent = "ua"

			a := c.Get("http://example.com")

			a.HostClient.Dial = func(addr string) (net.Conn, error) { return ln.Dial() }

			code, body, errs := a.String()

			utils.AssertEqual(t, StatusOK, code)
			utils.AssertEqual(t, "ua", body)
			utils.AssertEqual(t, 0, len(errs))
			ReleaseClient(c)
		}
	})
}

func Test_Client_Agent_Set_Or_Add_Headers(t *testing.T) {
	handler := func(req *Request, res *Response) error {
		req.ctx.fasthttp.Request.Header.VisitAll(func(key, value []byte) {
			if k := key; string(k) == "K1" || string(k) == "K2" {
				_, _ = res.Append(k)
				_, _ = res.Append(value)
			}
		})
		return nil
	}

	wrapAgent := func(a *Agent) {
		a.Set("k1", "v1").
			SetBytesK([]byte("k1"), "v1").
			SetBytesV("k1", []byte("v1")).
			AddBytesK([]byte("k1"), "v11").
			AddBytesV("k1", []byte("v22")).
			AddBytesKV([]byte("k1"), []byte("v33")).
			SetBytesKV([]byte("k2"), []byte("v2")).
			Add("k2", "v22")
	}

	testAgent(t, handler, wrapAgent, "K1v1K1v11K1v22K1v33K2v2K2v22")
}

func Test_Client_Agent_Connection_Close(t *testing.T) {
	handler := func(req *Request, res *Response) error {
		if req.IsConnectionClose() {
			return res.String("close")
		}
		return res.String("not close")
	}

	wrapAgent := func(a *Agent) {
		a.ConnectionClose()
	}

	testAgent(t, handler, wrapAgent, "close")
}

func Test_Client_Agent_UserAgent(t *testing.T) {
	handler := func(req *Request, res *Response) error {
		return res.String(req.UserAgent())
	}

	wrapAgent := func(a *Agent) {
		a.UserAgent("ua").
			UserAgentBytes([]byte("ua"))
	}

	testAgent(t, handler, wrapAgent, "ua")
}

func Test_Client_Agent_Cookie(t *testing.T) {
	handler := func(req *Request, res *Response) error {
		return res.String(req.GetCookie("k1") + req.GetCookie("k2") + req.GetCookie("k3") + req.GetCookie("k4"))
	}

	wrapAgent := func(a *Agent) {
		a.Cookie("k1", "v1").
			CookieBytesK([]byte("k2"), "v2").
			CookieBytesKV([]byte("k2"), []byte("v2")).
			Cookies("k3", "v3", "k4", "v4").
			CookiesBytesKV([]byte("k3"), []byte("v3"), []byte("k4"), []byte("v4"))
	}

	testAgent(t, handler, wrapAgent, "v1v2v3v4")
}

func Test_Client_Agent_Referer(t *testing.T) {
	handler := func(req *Request, res *Response) error {
		return res.String(req.Referer())
	}

	wrapAgent := func(a *Agent) {
		a.Referer("http://referer.com").
			RefererBytes([]byte("http://referer.com"))
	}

	testAgent(t, handler, wrapAgent, "http://referer.com")
}

func Test_Client_Agent_ContentType(t *testing.T) {
	handler := func(req *Request, res *Response) error {
		return res.String(req.Header.ContentType())
	}

	wrapAgent := func(a *Agent) {
		a.ContentType("custom-type").
			ContentTypeBytes([]byte("custom-type"))
	}

	testAgent(t, handler, wrapAgent, "custom-type")
}

func Test_Client_Agent_Host(t *testing.T) {
	t.Parallel()

	ln := fasthttputil.NewInmemoryListener()

	app := New(Config{DisableStartupMessage: true})

	app.Get("/", func(req *Request, res *Response) error {
		return res.String(req.Hostname())
	})

	go func() { utils.AssertEqual(t, nil, app.Listener(ln)) }()

	a := Get("http://1.1.1.1:8080").
		Host("example.com").
		HostBytes([]byte("example.com"))

	utils.AssertEqual(t, "1.1.1.1:8080", a.HostClient.Addr)

	a.HostClient.Dial = func(addr string) (net.Conn, error) { return ln.Dial() }

	code, body, errs := a.String()

	utils.AssertEqual(t, StatusOK, code)
	utils.AssertEqual(t, "example.com", body)
	utils.AssertEqual(t, 0, len(errs))
}

func Test_Client_Agent_QueryString(t *testing.T) {
	handler := func(req *Request, res *Response) error {
		return res.String(req.QueryString())
	}

	wrapAgent := func(a *Agent) {
		a.QueryString("foo=bar&bar=baz").
			QueryStringBytes([]byte("foo=bar&bar=baz"))
	}

	testAgent(t, handler, wrapAgent, "foo=bar&bar=baz")
}

func Test_Client_Agent_BasicAuth(t *testing.T) {
	handler := func(req *Request, res *Response) error {
		// Get authorization header
		auth := req.Header.Get(HeaderAuthorization)
		// Decode the header contents
		raw, err := base64.StdEncoding.DecodeString(auth[6:])
		utils.AssertEqual(t, nil, err)

		return res.Bytes(raw)
	}

	wrapAgent := func(a *Agent) {
		a.BasicAuth("foo", "bar").
			BasicAuthBytes([]byte("foo"), []byte("bar"))
	}

	testAgent(t, handler, wrapAgent, "foo:bar")
}

func Test_Client_Agent_BodyString(t *testing.T) {
	handler := func(req *Request, res *Response) error {
		return res.Bytes(req.Body())
	}

	wrapAgent := func(a *Agent) {
		a.BodyString("foo=bar&bar=baz")
	}

	testAgent(t, handler, wrapAgent, "foo=bar&bar=baz")
}

func Test_Client_Agent_Body(t *testing.T) {
	handler := func(req *Request, res *Response) error {
		return res.Bytes(req.Body())
	}

	wrapAgent := func(a *Agent) {
		a.Body([]byte("foo=bar&bar=baz"))
	}

	testAgent(t, handler, wrapAgent, "foo=bar&bar=baz")
}

func Test_Client_Agent_BodyStream(t *testing.T) {
	handler := func(req *Request, res *Response) error {
		return res.Bytes(req.Body())
	}

	wrapAgent := func(a *Agent) {
		a.BodyStream(strings.NewReader("body stream"), -1)
	}

	testAgent(t, handler, wrapAgent, "body stream")
}

func Test_Client_Agent_Custom_Response(t *testing.T) {
	t.Parallel()

	ln := fasthttputil.NewInmemoryListener()

	app := New(Config{DisableStartupMessage: true})

	app.Get("/", func(req *Request, res *Response) error {
		return res.String("custom")
	})

	go func() { utils.AssertEqual(t, nil, app.Listener(ln)) }()

	for i := 0; i < 5; i++ {
		a := AcquireAgent()
		resp := AcquireResponse()

		req := a.Request()
		req.Header.SetMethod(MethodGet)
		req.SetRequestURI("http://example.com")

		utils.AssertEqual(t, nil, a.Parse())

		a.HostClient.Dial = func(addr string) (net.Conn, error) { return ln.Dial() }

		code, body, errs := a.SetResponse(resp).
			String()

		utils.AssertEqual(t, StatusOK, code)
		utils.AssertEqual(t, "custom", body)
		utils.AssertEqual(t, "custom", string(resp.Body()))
		utils.AssertEqual(t, 0, len(errs))

		ReleaseResponse(resp)
	}
}

func Test_Client_Agent_Dest(t *testing.T) {
	t.Parallel()

	ln := fasthttputil.NewInmemoryListener()

	app := New(Config{DisableStartupMessage: true})

	app.Get("/", func(req *Request, res *Response) error {
		return res.String("dest")
	})

	go func() { utils.AssertEqual(t, nil, app.Listener(ln)) }()

	t.Run("small dest", func(t *testing.T) {
		dest := []byte("de")

		a := Get("http://example.com")

		a.HostClient.Dial = func(addr string) (net.Conn, error) { return ln.Dial() }

		code, body, errs := a.Dest(dest[:0]).String()

		utils.AssertEqual(t, StatusOK, code)
		utils.AssertEqual(t, "dest", body)
		utils.AssertEqual(t, "de", string(dest))
		utils.AssertEqual(t, 0, len(errs))
	})

	t.Run("enough dest", func(t *testing.T) {
		dest := []byte("foobar")

		a := Get("http://example.com")

		a.HostClient.Dial = func(addr string) (net.Conn, error) { return ln.Dial() }

		code, body, errs := a.Dest(dest[:0]).String()

		utils.AssertEqual(t, StatusOK, code)
		utils.AssertEqual(t, "dest", body)
		utils.AssertEqual(t, "destar", string(dest))
		utils.AssertEqual(t, 0, len(errs))
	})
}

func Test_Client_Stdjson_Gojson(t *testing.T) {
	type User struct {
		Account  *string `json:"account"`
		Password *string `json:"password"`
		Nickname *string `json:"nickname"`
		Address  *string `json:"address,omitempty"`
		Friends  []*User `json:"friends,omitempty"`
	}
	user1Account, user1Password, user1Nickname := "abcdef", "123456", "user1"
	user1 := &User{
		Account:  &user1Account,
		Password: &user1Password,
		Nickname: &user1Nickname,
		Address:  nil,
	}
	user2Account, user2Password, user2Nickname := "ghijkl", "123456", "user2"
	user2 := &User{
		Account:  &user2Account,
		Password: &user2Password,
		Nickname: &user2Nickname,
		Address:  nil,
	}
	user1.Friends = []*User{user2}
	expected, err := stdjson.Marshal(user1)
	utils.AssertEqual(t, nil, err)

	got, err := stdjson.Marshal(user1)
	utils.AssertEqual(t, nil, err)

	utils.AssertEqual(t, expected, got)

	type config struct {
		// debug enable a debug logging.
		debug bool
		// log used for logging on debug mode.
		log func(...interface{})
	}

	type res struct {
		config `json:"-"`
		// ID of the ent.
		ID uuid.UUID `json:"id,omitempty"`
	}

	u := uuid.New()
	test := res{
		ID: u,
	}

	expected, err = stdjson.Marshal(test)
	utils.AssertEqual(t, nil, err)

	got, err = stdjson.Marshal(test)
	utils.AssertEqual(t, nil, err)

	utils.AssertEqual(t, expected, got)
}

func Test_Client_Agent_Json(t *testing.T) {
	handler := func(req *Request, res *Response) error {
		utils.AssertEqual(t, MIMEApplicationJSON, req.Header.ContentType())

		return res.Bytes(req.Body())
	}

	wrapAgent := func(a *Agent) {
		a.JSON(data{Success: true})
	}

	testAgent(t, handler, wrapAgent, `{"success":true}`)
}

func Test_Client_Agent_Json_Error(t *testing.T) {
	a := Get("http://example.com").
		JSONEncoder(stdjson.Marshal).
		JSON(complex(1, 1))

	_, body, errs := a.String()

	utils.AssertEqual(t, "", body)
	utils.AssertEqual(t, 1, len(errs))
	utils.AssertEqual(t, "json: unsupported type: complex128", errs[0].Error())
}

func Test_Client_Agent_XML(t *testing.T) {
	handler := func(req *Request, res *Response) error {
		utils.AssertEqual(t, MIMEApplicationXML, req.Header.ContentType())

		return res.Bytes(req.Body())
	}

	wrapAgent := func(a *Agent) {
		a.XML(data{Success: true})
	}

	testAgent(t, handler, wrapAgent, "<data><success>true</success></data>")
}

func Test_Client_Agent_XML_Error(t *testing.T) {
	a := Get("http://example.com").
		XML(complex(1, 1))

	_, body, errs := a.String()

	utils.AssertEqual(t, "", body)
	utils.AssertEqual(t, 1, len(errs))
	utils.AssertEqual(t, "xml: unsupported type: complex128", errs[0].Error())
}

func Test_Client_Agent_Form(t *testing.T) {
	handler := func(req *Request, res *Response) error {
		utils.AssertEqual(t, MIMEApplicationForm, req.Header.ContentType())

		return res.Bytes(req.Body())
	}

	args := AcquireArgs()

	args.Set("foo", "bar")

	wrapAgent := func(a *Agent) {
		a.Form(args)
	}

	testAgent(t, handler, wrapAgent, "foo=bar")

	ReleaseArgs(args)
}

func Test_Client_Agent_MultipartForm(t *testing.T) {
	t.Parallel()

	ln := fasthttputil.NewInmemoryListener()

	app := New(Config{DisableStartupMessage: true})

	app.Post("/", func(req *Request, res *Response) error {
		utils.AssertEqual(t, "multipart/form-data; boundary=myBoundary", req.Header.Get(HeaderContentType))

		mf, err := req.MultipartForm()
		utils.AssertEqual(t, nil, err)
		utils.AssertEqual(t, "bar", mf.Value["foo"][0])

		return res.Bytes(req.Body())
	})

	go func() { utils.AssertEqual(t, nil, app.Listener(ln)) }()

	args := AcquireArgs()

	args.Set("foo", "bar")

	a := Post("http://example.com").
		Boundary("myBoundary").
		MultipartForm(args)

	a.HostClient.Dial = func(addr string) (net.Conn, error) { return ln.Dial() }

	code, body, errs := a.String()

	utils.AssertEqual(t, StatusOK, code)
	utils.AssertEqual(t, "--myBoundary\r\nContent-Disposition: form-data; name=\"foo\"\r\n\r\nbar\r\n--myBoundary--\r\n", body)
	utils.AssertEqual(t, 0, len(errs))
	ReleaseArgs(args)
}

func Test_Client_Agent_MultipartForm_Errors(t *testing.T) {
	t.Parallel()

	a := AcquireAgent()
	a.mw = &errorMultipartWriter{}

	args := AcquireArgs()
	args.Set("foo", "bar")

	ff1 := &FormFile{"", "name1", []byte("content"), false}
	ff2 := &FormFile{"", "name2", []byte("content"), false}
	a.FileData(ff1, ff2).
		MultipartForm(args)

	utils.AssertEqual(t, 4, len(a.errs))
	ReleaseArgs(args)
}

func Test_Client_Agent_MultipartForm_SendFiles(t *testing.T) {
	t.Parallel()

	ln := fasthttputil.NewInmemoryListener()

	app := New(Config{DisableStartupMessage: true})

	app.Post("/", func(req *Request, res *Response) error {
		utils.AssertEqual(t, "multipart/form-data; boundary=myBoundary", req.Header.Get(HeaderContentType))

		fh1, err := req.FormFile("field1")
		utils.AssertEqual(t, nil, err)
		utils.AssertEqual(t, fh1.Filename, "name")
		buf := make([]byte, fh1.Size)
		f, err := fh1.Open()
		utils.AssertEqual(t, nil, err)
		defer func() { _ = f.Close() }()
		_, err = f.Read(buf)
		utils.AssertEqual(t, nil, err)
		utils.AssertEqual(t, "form file", string(buf))

		fh2, err := req.FormFile("index")
		utils.AssertEqual(t, nil, err)
		checkFormFile(t, fh2, ".github/testdata/index.html")

		fh3, err := req.FormFile("file3")
		utils.AssertEqual(t, nil, err)
		checkFormFile(t, fh3, ".github/testdata/index.tmpl")

		return res.String("multipart form files")
	})

	go func() { utils.AssertEqual(t, nil, app.Listener(ln)) }()

	for i := 0; i < 5; i++ {
		ff := AcquireFormFile()
		ff.Fieldname = "field1"
		ff.Name = "name"
		ff.Content = []byte("form file")

		a := Post("http://example.com").
			Boundary("myBoundary").
			FileData(ff).
			SendFiles(".github/testdata/index.html", "index", ".github/testdata/index.tmpl").
			MultipartForm(nil)

		a.HostClient.Dial = func(addr string) (net.Conn, error) { return ln.Dial() }

		code, body, errs := a.String()

		utils.AssertEqual(t, StatusOK, code)
		utils.AssertEqual(t, "multipart form files", body)
		utils.AssertEqual(t, 0, len(errs))

		ReleaseFormFile(ff)
	}
}

func checkFormFile(t *testing.T, fh *multipart.FileHeader, filename string) {
	t.Helper()

	basename := filepath.Base(filename)
	utils.AssertEqual(t, fh.Filename, basename)

	b1, err := ioutil.ReadFile(filename)
	utils.AssertEqual(t, nil, err)

	b2 := make([]byte, fh.Size)
	f, err := fh.Open()
	utils.AssertEqual(t, nil, err)
	defer func() { _ = f.Close() }()
	_, err = f.Read(b2)
	utils.AssertEqual(t, nil, err)
	utils.AssertEqual(t, b1, b2)
}

func Test_Client_Agent_Multipart_Random_Boundary(t *testing.T) {
	t.Parallel()

	a := Post("http://example.com").
		MultipartForm(nil)

	reg := regexp.MustCompile(`multipart/form-data; boundary=\w{30}`)

	utils.AssertEqual(t, true, reg.Match(a.req.Header.Peek(HeaderContentType)))
}

func Test_Client_Agent_Multipart_Invalid_Boundary(t *testing.T) {
	t.Parallel()

	a := Post("http://example.com").
		Boundary("*").
		MultipartForm(nil)

	utils.AssertEqual(t, 1, len(a.errs))
	utils.AssertEqual(t, "mime: invalid boundary character", a.errs[0].Error())
}

func Test_Client_Agent_SendFile_Error(t *testing.T) {
	t.Parallel()

	a := Post("http://example.com").
		SendFile("non-exist-file!", "")

	utils.AssertEqual(t, 1, len(a.errs))
	utils.AssertEqual(t, true, strings.Contains(a.errs[0].Error(), "open non-exist-file!"))
}

func Test_Client_Debug(t *testing.T) {
	handler := func(req *Request, res *Response) error {
		return res.String("debug")
	}

	var output bytes.Buffer

	wrapAgent := func(a *Agent) {
		a.Debug(&output)
	}

	testAgent(t, handler, wrapAgent, "debug", 1)

	str := output.String()

	utils.AssertEqual(t, true, strings.Contains(str, "Connected to example.com(pipe)"))
	utils.AssertEqual(t, true, strings.Contains(str, "GET / HTTP/1.1"))
	utils.AssertEqual(t, true, strings.Contains(str, "User-Agent: lighting"))
	utils.AssertEqual(t, true, strings.Contains(str, "Host: example.com\r\n\r\n"))
	utils.AssertEqual(t, true, strings.Contains(str, "HTTP/1.1 200 OK"))
	utils.AssertEqual(t, true, strings.Contains(str, "Content-Type: text/plain; charset=utf-8\r\nContent-Length: 5\r\n\r\ndebug"))
}

func Test_Client_Agent_Timeout(t *testing.T) {
	t.Parallel()

	ln := fasthttputil.NewInmemoryListener()

	app := New(Config{DisableStartupMessage: true})

	app.Get("/", func(req *Request, res *Response) error {
		time.Sleep(time.Millisecond * 200)
		return res.String("timeout")
	})

	go func() { utils.AssertEqual(t, nil, app.Listener(ln)) }()

	a := Get("http://example.com").
		Timeout(time.Millisecond * 100)

	a.HostClient.Dial = func(addr string) (net.Conn, error) { return ln.Dial() }

	_, body, errs := a.String()

	utils.AssertEqual(t, "", body)
	utils.AssertEqual(t, 1, len(errs))
	utils.AssertEqual(t, "timeout", errs[0].Error())
}

func Test_Client_Agent_Reuse(t *testing.T) {
	t.Parallel()

	ln := fasthttputil.NewInmemoryListener()

	app := New(Config{DisableStartupMessage: true})

	app.Get("/", func(req *Request, res *Response) error {
		return res.String("reuse")
	})

	go func() { utils.AssertEqual(t, nil, app.Listener(ln)) }()

	a := Get("http://example.com").
		Reuse()

	a.HostClient.Dial = func(addr string) (net.Conn, error) { return ln.Dial() }

	code, body, errs := a.String()

	utils.AssertEqual(t, StatusOK, code)
	utils.AssertEqual(t, "reuse", body)
	utils.AssertEqual(t, 0, len(errs))

	code, body, errs = a.String()

	utils.AssertEqual(t, StatusOK, code)
	utils.AssertEqual(t, "reuse", body)
	utils.AssertEqual(t, 0, len(errs))
}

func Test_Client_Agent_InsecureSkipVerify(t *testing.T) {
	t.Parallel()

	cer, err := tls.LoadX509KeyPair("./.github/testdata/ssl.pem", "./.github/testdata/ssl.key")
	utils.AssertEqual(t, nil, err)

	serverTLSConf := &tls.Config{
		Certificates: []tls.Certificate{cer},
	}

	ln, err := net.Listen(NetworkTCP4, "127.0.0.1:0")
	utils.AssertEqual(t, nil, err)

	ln = tls.NewListener(ln, serverTLSConf)

	app := New(Config{DisableStartupMessage: true})

	app.Get("/", func(req *Request, res *Response) error {
		return res.String("ignore tls")
	})

	go func() { utils.AssertEqual(t, nil, app.Listener(ln)) }()

	code, body, errs := Get("https://" + ln.Addr().String()).
		InsecureSkipVerify().
		InsecureSkipVerify().
		String()

	utils.AssertEqual(t, 0, len(errs))
	utils.AssertEqual(t, StatusOK, code)
	utils.AssertEqual(t, "ignore tls", body)
}

func Test_Client_Agent_TLS(t *testing.T) {
	t.Parallel()

	serverTLSConf, clientTLSConf, err := tlstest.GetTLSConfigs()
	utils.AssertEqual(t, nil, err)

	ln, err := net.Listen(NetworkTCP4, "127.0.0.1:0")
	utils.AssertEqual(t, nil, err)

	ln = tls.NewListener(ln, serverTLSConf)

	app := New(Config{DisableStartupMessage: true})

	app.Get("/", func(req *Request, res *Response) error {
		return res.String("tls")
	})

	go func() { utils.AssertEqual(t, nil, app.Listener(ln)) }()

	code, body, errs := Get("https://" + ln.Addr().String()).
		TLSConfig(clientTLSConf).
		String()

	utils.AssertEqual(t, 0, len(errs))
	utils.AssertEqual(t, StatusOK, code)
	utils.AssertEqual(t, "tls", body)
}

func Test_Client_Agent_MaxRedirectsCount(t *testing.T) {
	t.Parallel()

	ln := fasthttputil.NewInmemoryListener()

	app := New(Config{DisableStartupMessage: true})

	app.Get("/", func(req *Request, res *Response) error {
		if req.QueryArgs().Has("foo") {
			return req.Redirect("/foo")
		}
		return req.Redirect("/")
	})
	app.Get("/foo", func(req *Request, res *Response) error {
		return res.String("redirect")
	})

	go func() { utils.AssertEqual(t, nil, app.Listener(ln)) }()

	t.Run("success", func(t *testing.T) {
		a := Get("http://example.com?foo").
			MaxRedirectsCount(1)

		a.HostClient.Dial = func(addr string) (net.Conn, error) { return ln.Dial() }

		code, body, errs := a.String()

		utils.AssertEqual(t, 200, code)
		utils.AssertEqual(t, "redirect", body)
		utils.AssertEqual(t, 0, len(errs))
	})

	t.Run("error", func(t *testing.T) {
		a := Get("http://example.com").
			MaxRedirectsCount(1)

		a.HostClient.Dial = func(addr string) (net.Conn, error) { return ln.Dial() }

		_, body, errs := a.String()

		utils.AssertEqual(t, "", body)
		utils.AssertEqual(t, 1, len(errs))
		utils.AssertEqual(t, "too many redirects detected when doing the request", errs[0].Error())
	})
}

func Test_Client_Agent_Struct(t *testing.T) {
	t.Parallel()

	ln := fasthttputil.NewInmemoryListener()

	app := New(Config{DisableStartupMessage: true})

	app.Get("/", func(req *Request, res *Response) error {
		return res.JSON(data{true})
	})

	app.Get("/error", func(req *Request, res *Response) error {
		return res.String(`{"success"`)
	})

	go func() { utils.AssertEqual(t, nil, app.Listener(ln)) }()

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		a := Get("http://example.com")

		a.HostClient.Dial = func(addr string) (net.Conn, error) { return ln.Dial() }

		var d data

		code, body, errs := a.Struct(&d)

		utils.AssertEqual(t, StatusOK, code)
		utils.AssertEqual(t, `{"success":true}`, string(body))
		utils.AssertEqual(t, 0, len(errs))
		utils.AssertEqual(t, true, d.Success)
	})

	t.Run("pre error", func(t *testing.T) {
		t.Parallel()
		a := Get("http://example.com")

		a.HostClient.Dial = func(addr string) (net.Conn, error) { return ln.Dial() }
		a.errs = append(a.errs, errors.New("pre errors"))

		var d data
		_, body, errs := a.Struct(&d)

		utils.AssertEqual(t, "", string(body))
		utils.AssertEqual(t, 1, len(errs))
		utils.AssertEqual(t, "pre errors", errs[0].Error())
		utils.AssertEqual(t, false, d.Success)
	})

	t.Run("error", func(t *testing.T) {
		a := Get("http://example.com/error")

		a.HostClient.Dial = func(addr string) (net.Conn, error) { return ln.Dial() }

		var d data

		code, body, errs := a.JSONDecoder(stdjson.Unmarshal).Struct(&d)

		utils.AssertEqual(t, StatusOK, code)
		utils.AssertEqual(t, `{"success"`, string(body))
		utils.AssertEqual(t, 1, len(errs))
		utils.AssertEqual(t, "unexpected end of JSON input", errs[0].Error())
	})
}

func Test_Client_Agent_Parse(t *testing.T) {
	t.Parallel()

	a := Get("https://example.com:10443")

	utils.AssertEqual(t, nil, a.Parse())
}

func Test_AddMissingPort_TLS(t *testing.T) {
	addr := addMissingPort("example.com", true)
	utils.AssertEqual(t, "example.com:443", addr)
}

func testAgent(t *testing.T, handler Handler, wrapAgent func(agent *Agent), excepted string, count ...int) {
	t.Parallel()

	ln := fasthttputil.NewInmemoryListener()

	app := New(Config{DisableStartupMessage: true})

	app.Get("/", handler)

	go func() { utils.AssertEqual(t, nil, app.Listener(ln)) }()

	c := 1
	if len(count) > 0 {
		c = count[0]
	}

	for i := 0; i < c; i++ {
		a := Get("http://example.com")

		wrapAgent(a)

		a.HostClient.Dial = func(addr string) (net.Conn, error) { return ln.Dial() }

		code, body, errs := a.String()

		utils.AssertEqual(t, StatusOK, code)
		utils.AssertEqual(t, excepted, body)
		utils.AssertEqual(t, 0, len(errs))
	}
}

type data struct {
	Success bool `json:"success" xml:"success"`
}

type errorMultipartWriter struct {
	count int
}

func (e *errorMultipartWriter) Boundary() string           { return "myBoundary" }
func (e *errorMultipartWriter) SetBoundary(_ string) error { return nil }
func (e *errorMultipartWriter) CreateFormFile(_, _ string) (io.Writer, error) {
	if e.count == 0 {
		e.count++
		return nil, errors.New("CreateFormFile error")
	}
	return errorWriter{}, nil
}
func (e *errorMultipartWriter) WriteField(_, _ string) error { return errors.New("WriteField error") }
func (e *errorMultipartWriter) Close() error                 { return errors.New("Close error") }

type errorWriter struct{}

func (errorWriter) Write(_ []byte) (int, error) { return 0, errors.New("Write error") }
