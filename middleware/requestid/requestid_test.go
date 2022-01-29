package requestid

import (
	"net/http/httptest"
	"testing"

	"github.com/ikidev/lightning"
	"github.com/ikidev/lightning/utils"
)

// go test -run Test_RequestID
func Test_RequestID(t *testing.T) {
	app := lightning.New()

	app.Use(New())

	app.Get("/", func(req *lightning.Request, res *lightning.Response) error {
		return res.String("Hello, World 👋!")
	})

	resp, err := app.Test(httptest.NewRequest("GET", "/", nil))
	utils.AssertEqual(t, nil, err)
	utils.AssertEqual(t, lightning.StatusOK, resp.StatusCode)

	reqid := resp.Header.Get(lightning.HeaderXRequestID)
	utils.AssertEqual(t, 36, len(reqid))

	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Add(lightning.HeaderXRequestID, reqid)

	resp, err = app.Test(req)
	utils.AssertEqual(t, nil, err)
	utils.AssertEqual(t, lightning.StatusOK, resp.StatusCode)
	utils.AssertEqual(t, reqid, resp.Header.Get(lightning.HeaderXRequestID))
}

// go test -run Test_RequestID_Next
func Test_RequestID_Next(t *testing.T) {
	app := lightning.New()
	app.Use(New(Config{
		Next: func(_ *lightning.Request, _ *lightning.Response) bool {
			return true
		},
	}))

	resp, err := app.Test(httptest.NewRequest("GET", "/", nil))
	utils.AssertEqual(t, nil, err)
	utils.AssertEqual(t, resp.Header.Get(lightning.HeaderXRequestID), "")
	utils.AssertEqual(t, lightning.StatusNotFound, resp.StatusCode)
}

// go test -run Test_RequestID_Locals
func Test_RequestID_Locals(t *testing.T) {
	reqId := "ThisIsARequestId"
	ctxKey := "ThisIsAContextKey"

	app := lightning.New()
	app.Use(New(Config{
		Generator: func() string {
			return reqId
		},
		ContextKey: ctxKey,
	}))

	var ctxVal string

	app.Use(func(req *lightning.Request, res *lightning.Response) error {
		ctxVal = req.Locals(ctxKey).(string)
		return req.Next()
	})

	_, err := app.Test(httptest.NewRequest("GET", "/", nil))
	utils.AssertEqual(t, nil, err)
	utils.AssertEqual(t, reqId, ctxVal)
}
