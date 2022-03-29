package csrf

import (
	"errors"

	"github.com/ikidev/lightning"
)

var (
	errMissingHeader = errors.New("missing csrf token in header")
	errMissingQuery  = errors.New("missing csrf token in query")
	errMissingParam  = errors.New("missing csrf token in param")
	errMissingForm   = errors.New("missing csrf token in form")
	errMissingCookie = errors.New("missing csrf token in cookie")
)

// csrfFromParam returns a function that extracts token from the url param string.
func csrfFromParam(param string) func(req *lightning.Request, res *lightning.Response) (string, error) {
	return func(req *lightning.Request, res *lightning.Response) (string, error) {
		token := req.Param(param)
		if token == "" {
			return "", errMissingParam
		}
		return token, nil
	}
}

// csrfFromForm returns a function that extracts a token from a multipart-form.
func csrfFromForm(param string) func(req *lightning.Request, res *lightning.Response) (string, error) {
	return func(req *lightning.Request, res *lightning.Response) (string, error) {
		token := req.FormValue(param)
		if token == "" {
			return "", errMissingForm
		}
		return token, nil
	}
}

// csrfFromCookie returns a function that extracts token from the cookie header.
func csrfFromCookie(param string) func(req *lightning.Request, res *lightning.Response) (string, error) {
	return func(req *lightning.Request, res *lightning.Response) (string, error) {
		token := req.GetCookie(param)
		if token == "" {
			return "", errMissingCookie
		}
		return token, nil
	}
}

// csrfFromHeader returns a function that extracts token from the request header.
func csrfFromHeader(param string) func(req *lightning.Request, res *lightning.Response) (string, error) {
	return func(req *lightning.Request, res *lightning.Response) (string, error) {
		token := req.Header.Get(param)
		if token == "" {
			return "", errMissingHeader
		}
		return token, nil
	}
}

// csrfFromQuery returns a function that extracts token from the query string.
func csrfFromQuery(param string) func(req *lightning.Request, res *lightning.Response) (string, error) {
	return func(req *lightning.Request, res *lightning.Response) (string, error) {
		token := req.Query(param)
		if token == "" {
			return "", errMissingQuery
		}
		return token, nil
	}
}
