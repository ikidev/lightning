package lightning

import (
	"encoding/json"
	"github.com/ikidev/lightning/utils"
	"github.com/valyala/fasthttp"
)

// Response is a struct holding all the response information. Allows you to set request properties
type Response struct {
	// Ctx represents the Context which hold the HTTP request and response.
	ctx        *Ctx
	Header     *HeaderMap
	body       interface{}
	rType      string
	StatusCode int
}

func (res *Response) Ctx() *Ctx {
	return res.ctx
}

func (res *Response) Status(status int) *Response {
	res.ctx.fasthttp.Response.SetStatusCode(status)
	res.rType = "StatusCode"
	res.StatusCode = status
	return res
}

func (res *Response) FastHTTPResponse() *fasthttp.Response {
	return res.ctx.Response()
}

// Append appends p into response body.
func (res *Response) Append(p []byte) (int, error) {
	res.ctx.fasthttp.Response.AppendBody(p)
	return len(p), nil
}

func (res *Response) Type(dataType string, extension ...string) *Response {
	res.ctx.Type(dataType, extension...)
	return res
}

func (res *Response) String(data string) error {
	res.ctx.fasthttp.Response.SetBodyString(data)
	res.rType = "string"

	return nil
}

func (res *Response) Bytes(data []byte) error {
	res.ctx.fasthttp.Response.SetBodyRaw(data)
	res.rType = "bytes"
	return nil
}
func (res *Response) Unauthorized(data []byte) error {
	res.ctx.fasthttp.Response.SetBodyRaw(data)
	res.rType = "bytes"
	return nil
}

func (res *Response) JSON(data interface{}) error {
	raw, err := res.ctx.app.config.JSONEncoder(data)
	if err != nil {
		return err
	}
	res.ctx.fasthttp.Response.SetBodyRaw(raw)
	res.ctx.fasthttp.Response.Header.SetContentType(MIMEApplicationJSON)
	res.rType = "json"
	return nil
}

// JSONP sends a JSON response with JSONP support.
// This method is identical to JSON, except that it opts-in to JSONP callback support.
// By default, the callback name is simply callback.
func (res *Response) JSONP(data interface{}, callback ...string) error {
	raw, err := json.Marshal(data)
	if err != nil {
		return err
	}

	var result, cb string

	if len(callback) > 0 {
		cb = callback[0]
	} else {
		cb = "callback"
	}

	result = cb + "(" + utils.UnsafeString(raw) + ");"

	res.ctx.setCanonical(HeaderXContentTypeOptions, "nosniff")
	res.ctx.fasthttp.Response.Header.SetContentType(MIMEApplicationJavaScriptCharsetUTF8)
	return res.String(result)
}

func (res *Response) SetCookie(cookie *Cookie) *Response {
	fCookie := fasthttp.AcquireCookie()
	fCookie.SetKey(cookie.Name)
	fCookie.SetValue(cookie.Value)
	fCookie.SetPath(cookie.Path)
	fCookie.SetDomain(cookie.Domain)
	fCookie.SetMaxAge(cookie.MaxAge)
	fCookie.SetExpire(cookie.Expires)
	fCookie.SetSecure(cookie.Secure)
	fCookie.SetHTTPOnly(cookie.HTTPOnly)

	switch utils.ToLower(cookie.SameSite) {
	case CookieSameSiteStrictMode:
		fCookie.SetSameSite(fasthttp.CookieSameSiteStrictMode)
	case CookieSameSiteNoneMode:
		fCookie.SetSameSite(fasthttp.CookieSameSiteNoneMode)
	case CookieSameSiteDisabled:
		fCookie.SetSameSite(fasthttp.CookieSameSiteDisabled)
	default:
		fCookie.SetSameSite(fasthttp.CookieSameSiteLaxMode)
	}

	res.ctx.fasthttp.Response.Header.SetCookie(fCookie)
	fasthttp.ReleaseCookie(fCookie)
	return res
}

func (res *Response) File(file string, compress ...bool) error {
	res.rType = "file"
	return res.ctx.SendFile(file, compress...)
}

func (res *Response) Send() error {

	if res.StatusCode == 0 {
		res.StatusCode = StatusOK
	}

	if res.rType == "StatusCode" {
		// Only set StatusCode body when there is no response body
		if len(res.ctx.fasthttp.Response.Body()) == 0 {
			return res.String(utils.StatusMessage(res.StatusCode))
		}
	}

	return nil
}
