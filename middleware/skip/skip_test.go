package skip_test

import (
	"net/http/httptest"
	"testing"

	"github.com/ikidev/lightning"
	"github.com/ikidev/lightning/middleware/skip"
	"github.com/ikidev/lightning/utils"
)

// go test -run Test_Skip
func Test_Skip(t *testing.T) {
	app := lightning.New()

	app.Use(skip.New(errTeapotHandler, func(*lightning.Ctx) bool { return true }))
	app.Get("/", helloWorldHandler)

	resp, err := app.Test(httptest.NewRequest("GET", "/", nil))
	utils.AssertEqual(t, nil, err)
	utils.AssertEqual(t, lightning.StatusOK, resp.StatusCode)
}

// go test -run Test_SkipFalse
func Test_SkipFalse(t *testing.T) {
	app := lightning.New()

	app.Use(skip.New(errTeapotHandler, func(*lightning.Ctx) bool { return false }))
	app.Get("/", helloWorldHandler)

	resp, err := app.Test(httptest.NewRequest("GET", "/", nil))
	utils.AssertEqual(t, nil, err)
	utils.AssertEqual(t, lightning.StatusTeapot, resp.StatusCode)
}

// go test -run Test_SkipNilFunc
func Test_SkipNilFunc(t *testing.T) {
	app := lightning.New()

	app.Use(skip.New(errTeapotHandler, nil))
	app.Get("/", helloWorldHandler)

	resp, err := app.Test(httptest.NewRequest("GET", "/", nil))
	utils.AssertEqual(t, nil, err)
	utils.AssertEqual(t, lightning.StatusTeapot, resp.StatusCode)
}

func helloWorldHandler(c *lightning.Ctx) error {
	return c.SendString("Hello, World 👋!")
}

func errTeapotHandler(*lightning.Ctx) error {
	return lightning.ErrTeapot
}
