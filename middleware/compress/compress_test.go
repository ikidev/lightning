package compress

import (
	"errors"
	"fmt"
	"io/ioutil"
	"net/http/httptest"
	"testing"

	"github.com/ikidev/lightning"
	"github.com/ikidev/lightning/utils"
)

var filedata []byte

func init() {
	dat, err := ioutil.ReadFile("../../.github/README.md")
	if err != nil {
		panic(err)
	}
	filedata = dat
}

// go test -run Test_Compress_Gzip
func Test_Compress_Gzip(t *testing.T) {
	app := lightning.New()

	app.Use(New())

	app.Get("/", func(req *lightning.Request, res *lightning.Response) error {
		res.Header.Set(lightning.HeaderContentType, lightning.MIMETextPlainCharsetUTF8)
		return res.Bytes(filedata)
	})

	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("Accept-Encoding", "gzip")

	resp, err := app.Test(req)
	utils.AssertEqual(t, nil, err, "app.Test(req)")
	utils.AssertEqual(t, 200, resp.StatusCode, "Status code")
	utils.AssertEqual(t, "gzip", resp.Header.Get(lightning.HeaderContentEncoding))

	// Validate that the file size has shrunk
	body, err := ioutil.ReadAll(resp.Body)
	utils.AssertEqual(t, nil, err)
	utils.AssertEqual(t, true, len(body) < len(filedata))
}

// go test -run Test_Compress_Different_Level
func Test_Compress_Different_Level(t *testing.T) {
	levels := []Level{LevelBestSpeed, LevelBestCompression}
	for _, level := range levels {
		t.Run(fmt.Sprintf("level %d", level), func(t *testing.T) {
			app := lightning.New()

			app.Use(New(Config{Level: level}))

			app.Get("/", func(req *lightning.Request, res *lightning.Response) error {
				res.Header.Set(lightning.HeaderContentType, lightning.MIMETextPlainCharsetUTF8)
				return res.Bytes(filedata)
			})

			req := httptest.NewRequest("GET", "/", nil)
			req.Header.Set("Accept-Encoding", "gzip")

			resp, err := app.Test(req)
			utils.AssertEqual(t, nil, err, "app.Test(req)")
			utils.AssertEqual(t, 200, resp.StatusCode, "Status code")
			utils.AssertEqual(t, "gzip", resp.Header.Get(lightning.HeaderContentEncoding))

			// Validate that the file size has shrunk
			body, err := ioutil.ReadAll(resp.Body)
			utils.AssertEqual(t, nil, err)
			utils.AssertEqual(t, true, len(body) < len(filedata))
		})
	}
}

func Test_Compress_Deflate(t *testing.T) {
	app := lightning.New()

	app.Use(New())

	app.Get("/", func(req *lightning.Request, res *lightning.Response) error {
		return res.Bytes(filedata)
	})

	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("Accept-Encoding", "deflate")

	resp, err := app.Test(req)
	utils.AssertEqual(t, nil, err, "app.Test(req)")
	utils.AssertEqual(t, 200, resp.StatusCode, "Status code")
	utils.AssertEqual(t, "deflate", resp.Header.Get(lightning.HeaderContentEncoding))

	// Validate that the file size has shrunk
	body, err := ioutil.ReadAll(resp.Body)
	utils.AssertEqual(t, nil, err)
	utils.AssertEqual(t, true, len(body) < len(filedata))
}

func Test_Compress_Brotli(t *testing.T) {
	app := lightning.New()

	app.Use(New())

	app.Get("/", func(req *lightning.Request, res *lightning.Response) error {
		return res.Bytes(filedata)
	})

	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("Accept-Encoding", "br")

	resp, err := app.Test(req, 10000)
	utils.AssertEqual(t, nil, err, "app.Test(req)")
	utils.AssertEqual(t, 200, resp.StatusCode, "Status code")
	utils.AssertEqual(t, "br", resp.Header.Get(lightning.HeaderContentEncoding))

	// Validate that the file size has shrunk
	body, err := ioutil.ReadAll(resp.Body)
	utils.AssertEqual(t, nil, err)
	utils.AssertEqual(t, true, len(body) < len(filedata))
}

func Test_Compress_Disabled(t *testing.T) {
	app := lightning.New()

	app.Use(New(Config{Level: LevelDisabled}))

	app.Get("/", func(req *lightning.Request, res *lightning.Response) error {
		return res.Bytes(filedata)
	})

	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("Accept-Encoding", "br")

	resp, err := app.Test(req)
	utils.AssertEqual(t, nil, err, "app.Test(req)")
	utils.AssertEqual(t, 200, resp.StatusCode, "Status code")
	utils.AssertEqual(t, "", resp.Header.Get(lightning.HeaderContentEncoding))

	// Validate the file size is not shrunk
	body, err := ioutil.ReadAll(resp.Body)
	utils.AssertEqual(t, nil, err)
	utils.AssertEqual(t, true, len(body) == len(filedata))
}

func Test_Compress_Next_Error(t *testing.T) {
	app := lightning.New()

	app.Use(New())

	app.Get("/", func(req *lightning.Request, res *lightning.Response) error {
		return errors.New("next error")
	})

	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("Accept-Encoding", "gzip")

	resp, err := app.Test(req)
	utils.AssertEqual(t, nil, err, "app.Test(req)")
	utils.AssertEqual(t, 500, resp.StatusCode, "Status code")
	utils.AssertEqual(t, "", resp.Header.Get(lightning.HeaderContentEncoding))

	body, err := ioutil.ReadAll(resp.Body)
	utils.AssertEqual(t, nil, err)
	utils.AssertEqual(t, "next error", string(body))
}

// go test -run Test_Compress_Next
func Test_Compress_Next(t *testing.T) {
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
