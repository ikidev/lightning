package logger

import (
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/ikidev/lightning"
	"github.com/ikidev/lightning/internal/bytebufferpool"
	"github.com/ikidev/lightning/internal/colorable"
	"github.com/ikidev/lightning/internal/fasttemplate"
	"github.com/ikidev/lightning/internal/isatty"
	"github.com/valyala/fasthttp"
)

// Logger variables
const (
	TagPid               = "pid"
	TagTime              = "time"
	TagReferer           = "referer"
	TagProtocol          = "protocol"
	TagPort              = "port"
	TagIP                = "ip"
	TagIPs               = "ips"
	TagHost              = "host"
	TagMethod            = "method"
	TagPath              = "path"
	TagURL               = "url"
	TagUA                = "ua"
	TagLatency           = "latency"
	TagStatus            = "status"
	TagResBody           = "resBody"
	TagReqHeaders        = "reqHeaders"
	TagQueryStringParams = "queryParams"
	TagBody              = "body"
	TagBytesSent         = "bytesSent"
	TagBytesReceived     = "bytesReceived"
	TagRoute             = "route"
	TagError             = "error"
	TagReqHeader         = "reqHeader:"
	TagRespHeader        = "respHeader:"
	TagLocals            = "locals:"
	TagQuery             = "query:"
	TagForm              = "form:"
	TagCookie            = "cookie:"
	TagBlack             = "black"
	TagRed               = "red"
	TagGreen             = "green"
	TagYellow            = "yellow"
	TagBlue              = "blue"
	TagMagenta           = "magenta"
	TagCyan              = "cyan"
	TagWhite             = "white"
	TagReset             = "reset"
)

// Color values
const (
	cBlack   = "\u001b[90m"
	cRed     = "\u001b[91m"
	cGreen   = "\u001b[92m"
	cYellow  = "\u001b[93m"
	cBlue    = "\u001b[94m"
	cMagenta = "\u001b[95m"
	cCyan    = "\u001b[96m"
	cWhite   = "\u001b[97m"
	cReset   = "\u001b[0m"
)

// New creates a new middleware handler
func New(config ...Config) lightning.Handler {
	// Set default config
	cfg := configDefault(config...)

	// Get timezone location
	tz, err := time.LoadLocation(cfg.TimeZone)
	if err != nil || tz == nil {
		cfg.timeZoneLocation = time.Local
	} else {
		cfg.timeZoneLocation = tz
	}

	// Check if format contains latency
	cfg.enableLatency = strings.Contains(cfg.Format, "${latency}")

	// Create template parser
	tmpl := fasttemplate.New(cfg.Format, "${", "}")

	// Create correct timeformat
	var timestamp atomic.Value
	timestamp.Store(time.Now().In(cfg.timeZoneLocation).Format(cfg.TimeFormat))

	// Update date/time every 750 milliseconds in a separate go routine
	if strings.Contains(cfg.Format, "${time}") {
		go func() {
			for {
				time.Sleep(cfg.TimeInterval)
				timestamp.Store(time.Now().In(cfg.timeZoneLocation).Format(cfg.TimeFormat))
			}
		}()
	}

	// Set PID once
	pid := strconv.Itoa(os.Getpid())

	// Set variables
	var (
		once       sync.Once
		mu         sync.Mutex
		errHandler lightning.ErrorHandler
	)

	// If colors are enabled, check terminal compatibility
	if cfg.enableColors {
		cfg.Output = colorable.NewColorableStdout()
		if os.Getenv("TERM") == "dumb" || os.Getenv("NO_COLOR") == "1" || (!isatty.IsTerminal(os.Stdout.Fd()) && !isatty.IsCygwinTerminal(os.Stdout.Fd())) {
			cfg.Output = colorable.NewNonColorable(os.Stdout)
		}
	}
	errPadding := 15
	errPaddingStr := strconv.Itoa(errPadding)
	// Return new handler
	return func(req *lightning.Request, res *lightning.Response) (err error) {
		// Don't execute middleware if Next returns true
		if cfg.Next != nil && cfg.Next(req, res) {
			return req.Next()
		}

		// Set error handler once
		once.Do(func() {
			// get longested possible path
			stack := req.Ctx().App().Stack()
			for m := range stack {
				for r := range stack[m] {
					if len(stack[m][r].Path) > errPadding {
						errPadding = len(stack[m][r].Path)
						errPaddingStr = strconv.Itoa(errPadding)
					}
				}
			}
			// override error handler
			errHandler = req.Ctx().App().ErrorHandler
		})

		var start, stop time.Time

		// Set latency start time
		if cfg.enableLatency {
			start = time.Now()
		}

		// Handle request, store err for logging
		chainErr := req.Next()

		// Manually call error handler
		if chainErr != nil {
			if err := errHandler(req, res, chainErr); err != nil {
				_ = res.Status(lightning.StatusInternalServerError).Send()
			}
		}

		// Set latency stop time
		if cfg.enableLatency {
			stop = time.Now()
		}

		// Get new buffer
		buf := bytebufferpool.Get()

		// Default output when no custom Format or io.Writer is given
		if cfg.enableColors && cfg.Format == ConfigDefault.Format {
			// Format error if exist
			formatErr := ""
			if chainErr != nil {
				formatErr = cRed + " | " + chainErr.Error() + cReset
			}

			// Format log to buffer
			_, _ = buf.WriteString(fmt.Sprintf("%s |%s %3d %s| %7v | %15s |%s %-7s %s| %-"+errPaddingStr+"s %s\n",
				timestamp.Load().(string),
				statusColor(res.Ctx().Response().StatusCode()), res.Ctx().Response().StatusCode(), cReset,
				stop.Sub(start).Round(time.Millisecond),
				req.IP(),
				methodColor(req.Method()), req.Method(), cReset,
				req.Path(),
				formatErr,
			))

			// Write buffer to output
			_, _ = cfg.Output.Write(buf.Bytes())

			// Put buffer back to pool
			bytebufferpool.Put(buf)

			// End chain
			return nil
		}

		// Loop over template tags to replace it with the correct value
		_, err = tmpl.ExecuteFunc(buf, func(w io.Writer, tag string) (int, error) {
			switch tag {
			case TagTime:
				return buf.WriteString(timestamp.Load().(string))
			case TagReferer:
				return buf.WriteString(req.Header.Get(lightning.HeaderReferer))
			case TagProtocol:
				return buf.WriteString(req.Protocol())
			case TagPid:
				return buf.WriteString(pid)
			case TagPort:
				return buf.WriteString(req.Port())
			case TagIP:
				return buf.WriteString(req.IP())
			case TagIPs:
				return buf.WriteString(req.Header.Get(lightning.HeaderXForwardedFor))
			case TagHost:
				return buf.WriteString(req.Hostname())
			case TagPath:
				return buf.WriteString(req.Path())
			case TagURL:
				return buf.WriteString(req.OriginalURL())
			case TagUA:
				return buf.WriteString(req.Header.Get(lightning.HeaderUserAgent))
			case TagLatency:
				return buf.WriteString(stop.Sub(start).String())
			case TagBody:
				return buf.Write(req.Body())
			case TagBytesReceived:
				return appendInt(buf, len(req.Ctx().Request().Body()))
			case TagBytesSent:
				return appendInt(buf, len(req.Ctx().Response().Body()))
			case TagRoute:
				return buf.WriteString(req.Ctx().Route().Path)
			case TagStatus:
				if cfg.enableColors {
					return buf.WriteString(fmt.Sprintf("%s %3d %s", statusColor(res.Ctx().Response().StatusCode()), res.Ctx().Response().StatusCode(), cReset))
				}
				return appendInt(buf, res.Ctx().Response().StatusCode())
			case TagResBody:
				return buf.Write(res.Ctx().Response().Body())
			case TagReqHeaders:
				reqHeaders := make([]string, 0)
				for k, v := range req.Header.All() {
					reqHeaders = append(reqHeaders, k+"="+v)
				}
				return buf.Write([]byte(strings.Join(reqHeaders, "&")))
			case TagQueryStringParams:
				return buf.WriteString(req.QueryArgs().String())
			case TagMethod:
				if cfg.enableColors {
					return buf.WriteString(fmt.Sprintf("%s %-7s %s", methodColor(req.Method()), req.Method(), cReset))
				}
				return buf.WriteString(req.Method())
			case TagBlack:
				return buf.WriteString(cBlack)
			case TagRed:
				return buf.WriteString(cRed)
			case TagGreen:
				return buf.WriteString(cGreen)
			case TagYellow:
				return buf.WriteString(cYellow)
			case TagBlue:
				return buf.WriteString(cBlue)
			case TagMagenta:
				return buf.WriteString(cMagenta)
			case TagCyan:
				return buf.WriteString(cCyan)
			case TagWhite:
				return buf.WriteString(cWhite)
			case TagReset:
				return buf.WriteString(cReset)
			case TagError:
				if chainErr != nil {
					return buf.WriteString(chainErr.Error())
				}
				return buf.WriteString("-")
			default:
				// Check if we have a value tag i.e.: "reqHeader:x-key"
				switch {
				case strings.HasPrefix(tag, TagReqHeader):
					return buf.WriteString(req.Header.Get(tag[10:]))
				case strings.HasPrefix(tag, TagRespHeader):
					return buf.WriteString(res.Header.Get(tag[11:]))
				case strings.HasPrefix(tag, TagQuery):
					return buf.WriteString(req.Query(tag[6:]))
				case strings.HasPrefix(tag, TagForm):
					return buf.WriteString(req.FormValue(tag[5:]))
				case strings.HasPrefix(tag, TagCookie):
					return buf.WriteString(req.GetCookie(tag[7:]))
				case strings.HasPrefix(tag, TagLocals):
					switch v := req.Locals(tag[7:]).(type) {
					case []byte:
						return buf.Write(v)
					case string:
						return buf.WriteString(v)
					case nil:
						return 0, nil
					default:
						return buf.WriteString(fmt.Sprintf("%v", v))
					}
				}
			}
			return 0, nil
		})
		// Also write errors to the buffer
		if err != nil {
			_, _ = buf.WriteString(err.Error())
		}
		mu.Lock()

		// Write buffer to output
		if _, err := cfg.Output.Write(buf.Bytes()); err != nil {
			// Write error to output
			if _, err := cfg.Output.Write([]byte(err.Error())); err != nil {
				// There is something wrong with the given io.Writer
				fmt.Fprintf(os.Stderr, "Failed to write to log, %v\n", err)
			}
		}
		mu.Unlock()
		// Put buffer back to pool
		bytebufferpool.Put(buf)

		return nil
	}
}

func appendInt(buf *bytebufferpool.ByteBuffer, v int) (int, error) {
	old := len(buf.B)
	buf.B = fasthttp.AppendUint(buf.B, v)
	return len(buf.B) - old, nil
}
