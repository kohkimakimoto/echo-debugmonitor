package monitors

import (
	_ "embed"
	"html/template"
	"net/http"
	"time"

	debugmonitor "github.com/kohkimakimoto/echo-debugmonitor/v5"
	"github.com/labstack/echo/v5"
	"github.com/labstack/echo/v5/middleware"
)

// RequestPayload represents the data structure for HTTP request monitoring
type RequestPayload struct {
	Method     string            `json:"method"`
	URI        string            `json:"uri"`
	Status     int               `json:"status"`
	Latency    int64             `json:"latency"` // in milliseconds
	RemoteAddr string            `json:"remoteAddr"`
	UserAgent  string            `json:"userAgent"`
	Error      string            `json:"error,omitempty"`
	Headers    map[string]string `json:"headers,omitempty"`
	Timestamp  time.Time         `json:"timestamp"`
}

// RequestMonitorConfig defines the config for Request monitor.
type RequestMonitorConfig struct {
	// Skipper defines a function to skip middleware.
	// Optional. Default: DefaultSkipper
	Skipper middleware.Skipper
	// UsePolling enables polling mode instead of SSE for real-time updates.
	UsePolling bool
}

//go:embed request.html
var requestView string

// requestViewTemplate is the parsed template for the request view
var requestViewTemplate = template.Must(template.New("requestView").Parse(requestView))

// NewRequestMonitor creates a new monitor for HTTP requests and returns
// the monitor along with an Echo middleware function that captures request information
func NewRequestMonitor(config *RequestMonitorConfig) (*debugmonitor.Monitor, echo.MiddlewareFunc) {
	// Defaults
	if config == nil {
		config = &RequestMonitorConfig{}
	}
	if config.Skipper == nil {
		config.Skipper = middleware.DefaultSkipper
	}

	m := &debugmonitor.Monitor{
		Name:        "request",
		DisplayName: "Requests",
		MaxRecords:  1000,
		Icon:        debugmonitor.IconGlobeAlt,
		ActionHandler: func(c *echo.Context, store *debugmonitor.Store, action string) error {
			switch action {
			case "render":
				return debugmonitor.RenderTemplate(c, requestViewTemplate, map[string]any{
					"UsePolling": config.UsePolling,
				})
			case "stream":
				// SSE endpoint for real-time updates
				return debugmonitor.HandleSSEStream(c, store)
			case "data":
				// JSON endpoint for polling mode
				return debugmonitor.HandleDataJSON(c, store)
			default:
				return echo.NewHTTPError(http.StatusBadRequest, "unknown action")
			}
		},
	}

	// Create middleware that captures request information
	mw := func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c *echo.Context) error {
			if config.Skipper(c) {
				return next(c)
			}

			start := time.Now()
			err := next(c)
			latency := time.Since(start)

			_, status := echo.ResolveResponseStatus(c.Response(), err)

			payload := &RequestPayload{
				Method:     c.Request().Method,
				URI:        c.Request().RequestURI,
				Status:     status,
				Latency:    latency.Milliseconds(),
				RemoteAddr: c.RealIP(),
				UserAgent:  c.Request().UserAgent(),
				Timestamp:  start,
			}

			payload.Headers = make(map[string]string)
			for key, values := range c.Request().Header {
				if len(values) > 0 {
					payload.Headers[key] = values[0]
				}
			}

			if err != nil {
				if he, ok := err.(*echo.HTTPError); ok {
					payload.Status = he.Code
					payload.Error = he.Message
				} else {
					payload.Error = err.Error()
				}
			}

			m.Add(payload)

			return err
		}
	}

	return m, mw
}
