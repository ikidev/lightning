package csrf

import (
	"fmt"
	"net/textproto"
	"strings"
	"time"

	"github.com/ikidev/lightning"
	"github.com/ikidev/lightning/utils"
)

// Config defines the config for middleware.
type Config struct {
	// Next defines a function to skip this middleware when returned true.
	//
	// Optional. Default: nil
	Next func(req *lightning.Request, res *lightning.Response) bool

	// KeyLookup is a string in the form of "<source>:<key>" that is used
	// to extract token from the request.
	// Possible values:
	// - "header:<name>"
	// - "query:<name>"
	// - "param:<name>"
	// - "form:<name>"
	// - "cookie:<name>"
	//
	// Optional. Default: "header:X-CSRF-Token"
	KeyLookup string

	// Name of the session cookie. This cookie will store session key.
	// Optional. Default value "csrf_".
	CookieName string

	// Domain of the CSRF cookie.
	// Optional. Default value "".
	CookieDomain string

	// Path of the CSRF cookie.
	// Optional. Default value "".
	CookiePath string

	// Indicates if CSRF cookie is secure.
	// Optional. Default value false.
	CookieSecure bool

	// Indicates if CSRF cookie is HTTP only.
	// Optional. Default value false.
	CookieHTTPOnly bool

	// Value of SameSite cookie.
	// Optional. Default value "Lax".
	CookieSameSite string

	// Expiration is the duration before csrf token will expire
	//
	// Optional. Default: 1 * time.Hour
	Expiration time.Duration

	// Store is used to store the state of the middleware
	//
	// Optional. Default: memory.New()
	Storage lightning.Storage

	// Context key to store generated CSRF token into context.
	// If left empty, token will not be stored in context.
	//
	// Optional. Default: ""
	ContextKey string

	// KeyGenerator creates a new CSRF token
	//
	// Optional. Default: utils.UUID
	KeyGenerator func() string

	// Deprecated, please use Expiration
	CookieExpires time.Duration

	// Deprecated, please use Cookie* related fields
	Cookie *lightning.Cookie

	// Deprecated, please use KeyLookup
	TokenLookup string

	// ErrorHandler is executed when an error is returned from fiber.Handler.
	//
	// Optional. Default: DefaultErrorHandler
	ErrorHandler lightning.ErrorHandler

	// extractor returns the csrf token from the request based on KeyLookup
	extractor func(req *lightning.Request, res *lightning.Response) (string, error)
}

// ConfigDefault is the default config
var ConfigDefault = Config{
	KeyLookup:      "header:X-Csrf-Token",
	CookieName:     "csrf_",
	CookieSameSite: "Lax",
	Expiration:     1 * time.Hour,
	KeyGenerator:   utils.UUID,
	ErrorHandler:   defaultErrorHandler,
	extractor:      csrfFromHeader("X-Csrf-Token"),
}

// default ErrorHandler that process return error from fiber.Handler
var defaultErrorHandler = func(req *lightning.Request, res *lightning.Response, err error) error {
	return lightning.ErrForbidden
}

// Helper function to set default values
func configDefault(config ...Config) Config {
	// Return default config if nothing provided
	if len(config) < 1 {
		return ConfigDefault
	}

	// Override default config
	cfg := config[0]

	// Set default values
	if cfg.TokenLookup != "" {
		fmt.Println("[CSRF] TokenLookup is deprecated, please use KeyLookup")
		cfg.KeyLookup = cfg.TokenLookup
	}
	if int(cfg.CookieExpires.Seconds()) > 0 {
		fmt.Println("[CSRF] CookieExpires is deprecated, please use Expiration")
		cfg.Expiration = cfg.CookieExpires
	}
	if cfg.Cookie != nil {
		fmt.Println("[CSRF] Cookie is deprecated, please use Cookie* related fields")
		if cfg.Cookie.Name != "" {
			cfg.CookieName = cfg.Cookie.Name
		}
		if cfg.Cookie.Domain != "" {
			cfg.CookieDomain = cfg.Cookie.Domain
		}
		if cfg.Cookie.Path != "" {
			cfg.CookiePath = cfg.Cookie.Path
		}
		cfg.CookieSecure = cfg.Cookie.Secure
		cfg.CookieHTTPOnly = cfg.Cookie.HTTPOnly
		if cfg.Cookie.SameSite != "" {
			cfg.CookieSameSite = cfg.Cookie.SameSite
		}
	}
	if cfg.KeyLookup == "" {
		cfg.KeyLookup = ConfigDefault.KeyLookup
	}
	if int(cfg.Expiration.Seconds()) <= 0 {
		cfg.Expiration = ConfigDefault.Expiration
	}
	if cfg.CookieName == "" {
		cfg.CookieName = ConfigDefault.CookieName
	}
	if cfg.CookieSameSite == "" {
		cfg.CookieSameSite = ConfigDefault.CookieSameSite
	}
	if cfg.KeyGenerator == nil {
		cfg.KeyGenerator = ConfigDefault.KeyGenerator
	}
	if cfg.ErrorHandler == nil {
		cfg.ErrorHandler = ConfigDefault.ErrorHandler
	}

	// Generate the correct extractor to get the token from the correct location
	selectors := strings.Split(cfg.KeyLookup, ":")

	if len(selectors) != 2 {
		panic("[CSRF] KeyLookup must in the form of <source>:<key>")
	}

	// By default we extract from a header
	cfg.extractor = csrfFromHeader(textproto.CanonicalMIMEHeaderKey(selectors[1]))

	switch selectors[0] {
	case "form":
		cfg.extractor = csrfFromForm(selectors[1])
	case "query":
		cfg.extractor = csrfFromQuery(selectors[1])
	case "param":
		cfg.extractor = csrfFromParam(selectors[1])
	case "cookie":
		cfg.extractor = csrfFromCookie(selectors[1])
	}

	return cfg
}
