package expvar

import (
	"bytes"
	"io/ioutil"
	"net/http/httptest"
	"testing"

	"github.com/ikidev/lightning"
	"github.com/ikidev/lightning/utils"
)

func Test_Non_Expvar_Path(t *testing.T) {
	app := lightning.New()

	app.Use(New())

	app.Get("/", func(req *lightning.Request, res *lightning.Response) error {
		return res.String("escaped")
	})

	resp, err := app.Test(httptest.NewRequest(lightning.MethodGet, "/", nil))
	utils.AssertEqual(t, nil, err)
	utils.AssertEqual(t, 200, resp.StatusCode)

	b, err := ioutil.ReadAll(resp.Body)
	utils.AssertEqual(t, nil, err)
	utils.AssertEqual(t, "escaped", string(b))
}

func Test_Expvar_Index(t *testing.T) {
	app := lightning.New()

	app.Use(New())

	app.Get("/", func(req *lightning.Request, res *lightning.Response) error {
		return res.String("escaped")
	})

	resp, err := app.Test(httptest.NewRequest(lightning.MethodGet, "/debug/vars", nil))
	utils.AssertEqual(t, nil, err)
	utils.AssertEqual(t, 200, resp.StatusCode)
	utils.AssertEqual(t, lightning.MIMEApplicationJSONCharsetUTF8, resp.Header.Get(lightning.HeaderContentType))

	b, err := ioutil.ReadAll(resp.Body)
	utils.AssertEqual(t, nil, err)
	utils.AssertEqual(t, true, bytes.Contains(b, []byte("cmdline")))
	utils.AssertEqual(t, true, bytes.Contains(b, []byte("memstat")))
}

func Test_Expvar_Filter(t *testing.T) {
	app := lightning.New()

	app.Use(New())

	app.Get("/", func(req *lightning.Request, res *lightning.Response) error {
		return res.String("escaped")
	})

	resp, err := app.Test(httptest.NewRequest(lightning.MethodGet, "/debug/vars?r=cmd", nil))
	utils.AssertEqual(t, nil, err)
	utils.AssertEqual(t, 200, resp.StatusCode)
	utils.AssertEqual(t, lightning.MIMEApplicationJSONCharsetUTF8, resp.Header.Get(lightning.HeaderContentType))

	b, err := ioutil.ReadAll(resp.Body)
	utils.AssertEqual(t, nil, err)
	utils.AssertEqual(t, true, bytes.Contains(b, []byte("cmdline")))
	utils.AssertEqual(t, false, bytes.Contains(b, []byte("memstat")))
}

func Test_Expvar_Other_Path(t *testing.T) {
	app := lightning.New()

	app.Use(New())

	app.Get("/", func(req *lightning.Request, res *lightning.Response) error {
		return res.String("escaped")
	})

	resp, err := app.Test(httptest.NewRequest(lightning.MethodGet, "/debug/vars/302", nil))
	utils.AssertEqual(t, nil, err)
	utils.AssertEqual(t, 302, resp.StatusCode)
}
