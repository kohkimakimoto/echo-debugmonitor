package monitors

import (
	_ "embed"
	"fmt"
	"net/http"
	"time"

	debugmonitor "github.com/kohkimakimoto/echo-debugmonitor"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
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

// RequestsMonitorConfig defines the config for Requests monitor.
type RequestsMonitorConfig struct {
	// Skipper defines a function to skip middleware.
	// Optional. Default: DefaultSkipper
	Skipper middleware.Skipper
}

//go:embed requests.html
var requestsView string

// NewRequestsMonitor creates a new monitor for HTTP requests and returns
// the monitor along with an Echo middleware function that captures request information
func NewRequestsMonitor(config *RequestsMonitorConfig) (*debugmonitor.Monitor, echo.MiddlewareFunc) {
	// Defaults
	if config == nil {
		config = &RequestsMonitorConfig{}
	}
	if config.Skipper == nil {
		config.Skipper = middleware.DefaultSkipper
	}

	m := &debugmonitor.Monitor{
		Name:        "requests",
		DisplayName: "Requests",
		MaxRecords:  1000,
		Icon:        debugmonitor.IconGlobeAlt,
		ActionHandler: func(c echo.Context, store *debugmonitor.Store, action string) error {
			switch action {
			case "render":
				return c.HTML(http.StatusOK, requestsView)
			case "stream":
				// SSE endpoint for real-time updates
				return debugmonitor.HandleSSEStream(c, store)
			default:
				return echo.NewHTTPError(http.StatusBadRequest)
			}
		},
	}

	// Create middleware that captures request information
	mw := func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// Check if request should be skipped
			if config.Skipper(c) {
				return next(c)
			}

			start := time.Now()

			// Process the request
			err := next(c)

			// Calculate latency
			latency := time.Since(start)

			// Get response status
			status := c.Response().Status
			if status == 0 {
				status = http.StatusOK
			}

			// Create payload
			payload := &RequestPayload{
				Method:     c.Request().Method,
				URI:        c.Request().RequestURI,
				Status:     status,
				Latency:    latency.Milliseconds(),
				RemoteAddr: c.RealIP(),
				UserAgent:  c.Request().UserAgent(),
				Timestamp:  start,
			}

			// Include headers if configured
			payload.Headers = make(map[string]string)
			for key, values := range c.Request().Header {
				if len(values) > 0 {
					payload.Headers[key] = values[0]
				}
			}

			// Include error if any
			if err != nil {
				if he, ok := err.(*echo.HTTPError); ok {
					payload.Status = he.Code
					payload.Error = fmt.Sprintf("%v", he.Message)
				} else {
					payload.Error = err.Error()
				}
			}

			// Add to monitor
			m.Add(payload)

			return err
		}
	}

	return m, mw
}
