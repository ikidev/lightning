package timeout

import (
	"github.com/ikidev/lightning"
	"github.com/ikidev/lightning/middleware/recovery"
	"github.com/ikidev/lightning/utils"
	"io/ioutil"
	"net/http/httptest"
	"testing"
	"time"
)

//go test -run Test_Middleware_Timeout
func Test_Middleware_Timeout(t *testing.T) {
	app := lightning.New(lightning.Config{DisableStartupMessage: true})

	h := New(func(req *lightning.Request, res *lightning.Response) error {
		sleepTime, _ := time.ParseDuration(req.UrlParam("sleepTime") + "ms")
		time.Sleep(sleepTime)
		return res.String("After " + req.UrlParam("sleepTime") + "ms sleeping")
	}, 5*time.Millisecond)
	app.Get("/test/:sleepTime", h)

	testTimeout := func(timeoutStr string) {
		resp, err := app.Test(httptest.NewRequest("GET", "/test/"+timeoutStr, nil))
		utils.AssertEqual(t, nil, err, "app.Test(req)")
		utils.AssertEqual(t, lightning.StatusRequestTimeout, resp.StatusCode, "Status code")

		body, err := ioutil.ReadAll(resp.Body)
		utils.AssertEqual(t, nil, err)
		utils.AssertEqual(t, "Request Timeout", string(body))
	}
	testSuccess := func(timeoutStr string) {
		resp, err := app.Test(httptest.NewRequest("GET", "/test/"+timeoutStr, nil))
		utils.AssertEqual(t, nil, err, "app.Test(req)")
		utils.AssertEqual(t, lightning.StatusOK, resp.StatusCode, "Status code")

		body, err := ioutil.ReadAll(resp.Body)
		utils.AssertEqual(t, nil, err)
		utils.AssertEqual(t, "After "+timeoutStr+"ms sleeping", string(body))
	}

	testTimeout("15")
	testSuccess("2")
	testTimeout("30")
	testSuccess("3")
}

//go test -run -v Test_Timeout_Panic
func Test_Timeout_Panic(t *testing.T) {
	app := lightning.New(lightning.Config{DisableStartupMessage: true})

	app.Get("/panic", recovery.New(), New(func(req *lightning.Request, res *lightning.Response) error {
		req.Header.Set("dummy", "this should not be here")
		panic("panic in timeout handler")
	}, 5*time.Millisecond))

	resp, err := app.Test(httptest.NewRequest("GET", "/panic", nil))
	utils.AssertEqual(t, nil, err, "app.Test(req)")
	utils.AssertEqual(t, lightning.StatusRequestTimeout, resp.StatusCode, "Status code")

	body, err := ioutil.ReadAll(resp.Body)
	utils.AssertEqual(t, nil, err)
	utils.AssertEqual(t, "Request Timeout", string(body))
}
