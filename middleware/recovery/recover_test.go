package recovery

import (
	"net/http/httptest"
	"testing"

	"github.com/ikidev/lightning"
	"github.com/ikidev/lightning/utils"
)

// go test -run Test_Recover
func Test_Recover(t *testing.T) {
	app := lightning.New(lightning.Config{
		ErrorHandler: func(req *lightning.Request, res *lightning.Response, err error) error {
			utils.AssertEqual(t, "Hi, I'm an error!", err.Error())
			return res.Status(lightning.StatusTeapot).Send()
		},
	})

	app.Use(New())

	app.Get("/panic", func(_ *lightning.Request, _ *lightning.Response) error {
		panic("Hi, I'm an error!")
	})

	resp, err := app.Test(httptest.NewRequest("GET", "/panic", nil))
	utils.AssertEqual(t, nil, err)
	utils.AssertEqual(t, lightning.StatusTeapot, resp.StatusCode)
}

// go test -run Test_Recover_Next
func Test_Recover_Next(t *testing.T) {
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

func Test_Recover_EnableStackTrace(t *testing.T) {
	app := lightning.New()
	app.Use(New(Config{
		EnableStackTrace: true,
	}))

	app.Get("/panic", func(_ *lightning.Request, _ *lightning.Response) error {
		panic("Hi, I'm an error!")
	})

	resp, err := app.Test(httptest.NewRequest("GET", "/panic", nil))
	utils.AssertEqual(t, nil, err)
	utils.AssertEqual(t, lightning.StatusInternalServerError, resp.StatusCode)
}
