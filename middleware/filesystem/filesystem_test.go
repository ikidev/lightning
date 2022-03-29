package filesystem

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ikidev/lightning"
	"github.com/ikidev/lightning/utils"
)

// go test -run Test_FileSystem
func Test_FileSystem(t *testing.T) {
	app := lightning.New()

	app.Use("/test", New(Config{
		Root: http.Dir("../../.github/testdata/fs"),
	}))

	app.Use("/dir", New(Config{
		Root:   http.Dir("../../.github/testdata/fs"),
		Browse: true,
	}))

	app.Get("/", func(req *lightning.Request, res *lightning.Response) error {
		return res.String("Hello, World!")
	})

	app.Use("/spatest", New(Config{
		Root:         http.Dir("../../.github/testdata/fs"),
		Index:        "index.html",
		NotFoundFile: "index.html",
	}))

	app.Use("/prefix", New(Config{
		Root:       http.Dir("../../.github/testdata/fs"),
		PathPrefix: "img",
	}))

	tests := []struct {
		name         string
		url          string
		statusCode   int
		contentType  string
		modifiedTime string
	}{
		{
			name:        "Should be returns status 200 with suitable content-type",
			url:         "/test/index.html",
			statusCode:  200,
			contentType: "text/html",
		},
		{
			name:        "Should be returns status 200 with suitable content-type",
			url:         "/test",
			statusCode:  200,
			contentType: "text/html",
		},
		{
			name:        "Should be returns status 200 with suitable content-type",
			url:         "/test/css/style.css",
			statusCode:  200,
			contentType: "text/css",
		},
		{
			name:       "Should be returns status 404",
			url:        "/test/nofile.js",
			statusCode: 404,
		},
		{
			name:       "Should be returns status 404",
			url:        "/test/nofile",
			statusCode: 404,
		},
		{
			name:        "Should be returns status 200",
			url:         "/",
			statusCode:  200,
			contentType: "text/plain; charset=utf-8",
		},
		{
			name:       "Should be returns status 403",
			url:        "/test/img",
			statusCode: 403,
		},
		{
			name:        "Should list the directory contents",
			url:         "/dir/img",
			statusCode:  200,
			contentType: "text/html",
		},
		{
			name:        "Should list the directory contents",
			url:         "/dir/img/",
			statusCode:  200,
			contentType: "text/html",
		},
		{
			name:        "Should be returns status 200",
			url:         "/dir/img/fiber.png",
			statusCode:  200,
			contentType: "image/png",
		},
		{
			name:        "Should be return status 200",
			url:         "/spatest/doesnotexist",
			statusCode:  200,
			contentType: "text/html",
		},
		{
			name:        "PathPrefix should be applied",
			url:         "/prefix/fiber.png",
			statusCode:  200,
			contentType: "image/png",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := app.Test(httptest.NewRequest("GET", tt.url, nil))
			utils.AssertEqual(t, nil, err)
			utils.AssertEqual(t, tt.statusCode, resp.StatusCode)

			if tt.contentType != "" {
				ct := resp.Header.Get("Content-Type")
				utils.AssertEqual(t, tt.contentType, ct)
			}
		})
	}
}

// go test -run Test_FileSystem_Next
func Test_FileSystem_Next(t *testing.T) {
	app := lightning.New()
	app.Use(New(Config{
		Root: http.Dir("../../.github/testdata/fs"),
		Next: func(_ *lightning.Request, _ *lightning.Response) bool {
			return true
		},
	}))

	resp, err := app.Test(httptest.NewRequest("GET", "/", nil))
	utils.AssertEqual(t, nil, err)
	utils.AssertEqual(t, lightning.StatusNotFound, resp.StatusCode)
}

func Test_FileSystem_NonGetAndHead(t *testing.T) {
	app := lightning.New()

	app.Use("/test", New(Config{
		Root: http.Dir("../../.github/testdata/fs"),
	}))

	resp, err := app.Test(httptest.NewRequest(lightning.MethodPost, "/test", nil))
	utils.AssertEqual(t, nil, err)
	utils.AssertEqual(t, 404, resp.StatusCode)
}

func Test_FileSystem_Head(t *testing.T) {
	app := lightning.New()

	app.Use("/test", New(Config{
		Root: http.Dir("../../.github/testdata/fs"),
	}))

	req, _ := http.NewRequest(lightning.MethodHead, "/test", nil)
	resp, err := app.Test(req)
	utils.AssertEqual(t, nil, err)
	utils.AssertEqual(t, 200, resp.StatusCode)
}

func Test_FileSystem_NoRoot(t *testing.T) {
	defer func() {
		utils.AssertEqual(t, "filesystem: Root cannot be nil", recover())
	}()

	app := lightning.New()
	app.Use(New())
	_, _ = app.Test(httptest.NewRequest(lightning.MethodGet, "/", nil))
}

func Test_FileSystem_UsingParam(t *testing.T) {
	app := lightning.New()

	app.Use("/:path", func(req *lightning.Request, res *lightning.Response) error {
		return SendFile(req, res, http.Dir("../../.github/testdata/fs"), req.Param("path")+".html")
	})

	req, _ := http.NewRequest(lightning.MethodHead, "/index", nil)
	resp, err := app.Test(req)
	utils.AssertEqual(t, nil, err)
	utils.AssertEqual(t, 200, resp.StatusCode)
}

func Test_FileSystem_UsingParam_NonFile(t *testing.T) {
	app := lightning.New()

	app.Use("/:path", func(req *lightning.Request, res *lightning.Response) error {
		return SendFile(req, res, http.Dir("../../.github/testdata/fs"), req.Param("path")+".html")
	})

	req, _ := http.NewRequest(lightning.MethodHead, "/template", nil)
	resp, err := app.Test(req)
	utils.AssertEqual(t, nil, err)
	utils.AssertEqual(t, 404, resp.StatusCode)
}
