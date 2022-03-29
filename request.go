package lightning

import (
	"github.com/ikidev/lightning/utils"
	"github.com/valyala/fasthttp"
	"mime/multipart"
	"strconv"
	"strings"
)

// Request is a struct holding all the request information.
type Request struct {
	// Ctx represents the Context which hold the HTTP request and response.
	ctx    *Ctx
	Header *HeaderMap
}

// Accepts checks if the specified extensions or content types are acceptable.
func (req *Request) Accepts(offers ...string) string {
	if len(offers) == 0 {
		return ""
	}
	header := req.Header.Get(HeaderAccept)
	if header == "" {
		return offers[0]
	}

	spec, commaPos := "", 0
	for len(header) > 0 && commaPos != -1 {
		commaPos = strings.IndexByte(header, ',')
		if commaPos != -1 {
			spec = utils.Trim(header[:commaPos], ' ')
		} else {
			spec = utils.TrimLeft(header, ' ')
		}
		if factorSign := strings.IndexByte(spec, ';'); factorSign != -1 {
			spec = spec[:factorSign]
		}

		var mimetype string
		for _, offer := range offers {
			if len(offer) == 0 {
				continue
				// Accept: */*
			} else if spec == "*/*" {
				return offer
			}

			if strings.IndexByte(offer, '/') != -1 {
				mimetype = offer // MIME type
			} else {
				mimetype = utils.GetMIME(offer) // extension
			}

			if spec == mimetype {
				// Accept: <MIME_type>/<MIME_subtype>
				return offer
			}

			s := strings.IndexByte(mimetype, '/')
			// Accept: <MIME_type>/*
			if strings.HasPrefix(spec, mimetype[:s]) && (spec[s:] == "/*" || mimetype[s:] == "/*") {
				return offer
			}
		}
		if commaPos != -1 {
			header = header[commaPos+1:]
		}
	}

	return ""
}

// AcceptsCharsets checks if the specified charset is acceptable.
func (req *Request) AcceptsCharsets(offers ...string) string {
	return getOffer(req.Header.Get(HeaderAcceptCharset), offers...)
}

// AcceptsEncodings checks if the specified encoding is acceptable.
func (req *Request) AcceptsEncodings(offers ...string) string {
	return getOffer(req.Header.Get(HeaderAcceptEncoding), offers...)
}

// AcceptsLanguages checks if the specified language is acceptable.
func (req *Request) AcceptsLanguages(offers ...string) string {
	return getOffer(req.Header.Get(HeaderAcceptLanguage), offers...)
}

func (req *Request) Path(override ...string) string {
	return req.ctx.Path(override...)
}

func (req *Request) Route() *Route {
	return req.ctx.Route()
}

func (req *Request) FastHTTPRequest() *fasthttp.Request {
	return req.ctx.Request()
}

func (req *Request) IP() string {
	return req.ctx.IP()
}
func (req *Request) Port() string {
	return req.ctx.Port()
}

func (req *Request) Protocol() string {
	return req.ctx.Protocol()
}

func (req *Request) IPs() []string {
	return req.ctx.IPs()
}

func (req *Request) Ctx() *Ctx {
	return req.ctx
}

func (req *Request) FastHTTPContext() *fasthttp.RequestCtx {
	return req.ctx.Context()
}

func (req *Request) Hostname() string {
	return req.ctx.Hostname()
}

func (req *Request) OriginalURL() string {
	return utils.UnsafeString(req.ctx.fasthttp.Request.Header.RequestURI())
}

func (req *Request) Method() string {
	return req.ctx.Method()
}

func (req *Request) SetCookie(cookie *Cookie) {
	req.ctx.Cookie(cookie)
}

func (req *Request) GetCookie(key string, defaultValue ...string) string {
	return defaultString(utils.UnsafeString(req.ctx.fasthttp.Request.Header.Cookie(key)), defaultValue)
}

func (req *Request) ClearCookie(key string) {
	req.ctx.ClearCookie(key)
}

func (req *Request) UserAgent() string {
	return utils.UnsafeString(req.ctx.Request().Header.UserAgent())
}

func (req *Request) Referer() string {
	return utils.UnsafeString(req.ctx.Request().Header.Referer())
}

func (req *Request) IsConnectionClose() bool {
	return req.ctx.Request().Header.ConnectionClose()
}

func (req *Request) IsXHR() bool {
	return utils.EqualFoldBytes(utils.UnsafeBytes(req.Header.Get(HeaderXRequestedWith)), []byte("xmlhttprequest"))
}

func (req *Request) IsBodyStream() bool {
	return req.ctx.Request().IsBodyStream()
}

func (req *Request) MultipartForm() (*multipart.Form, error) {
	return req.ctx.Request().MultipartForm()
}

func (req *Request) Query(key string, defaultValue ...string) string {
	return req.ctx.Query(key, defaultValue...)
}

// Locals makes it possible to pass interface{} values under string keys scoped to the request
// and therefore available to all following routes that match the request.
func (req *Request) Locals(key string, value ...interface{}) (val interface{}) {
	if len(value) == 0 {
		return req.ctx.fasthttp.UserValue(key)
	}
	req.ctx.fasthttp.SetUserValue(key, value[0])
	return value[0]
}

func (req *Request) FormValue(key string) string {
	return req.ctx.FormValue(key)
}

func (req *Request) FormFile(key string) (*multipart.FileHeader, error) {
	return req.ctx.FormFile(key)
}

func (req *Request) SaveFile(file *multipart.FileHeader, path string) error {
	return fasthttp.SaveMultipartFile(file, path)
}

func (req *Request) QueryString() string {
	return utils.UnsafeString(req.ctx.Request().URI().QueryString())
}

func (req *Request) QueryArgs() *Args {
	return req.ctx.Request().URI().QueryArgs()
}

func (req *Request) URI() *fasthttp.URI {
	return req.ctx.Request().URI()
}

func (req *Request) Body() []byte {
	return req.ctx.Request().Body()
}

func (req *Request) Next() error {
	return req.ctx.Next()
}

func (req *Request) Redirect(location string, status ...int) error {
	return req.ctx.Redirect(location, status...)
}

func (req *Request) Param(key string, defaultValue ...string) string {
	return req.ctx.Params(key, defaultValue...)
}

func (req *Request) IntParam(key string, defaultValue ...int) int {
	// Use Atoi to convert the param to an int or return zero and an error
	value, err := strconv.Atoi(req.Param(key))
	if err != nil {
		if len(defaultValue) > 0 {
			return defaultValue[0]
		}
	}
	return value
}
