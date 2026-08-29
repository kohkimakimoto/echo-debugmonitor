package monitors

import (
	"context"
	_ "embed"
	"html/template"
	"log/slog"
	"net/http"
	"time"

	debugmonitor "github.com/kohkimakimoto/echo-debugmonitor/v5"
	"github.com/labstack/echo/v5"
)

// LogPayload represents the data structure for log monitoring
type LogPayload struct {
	Level     string    `json:"level"`
	Message   string    `json:"message"`
	Timestamp time.Time `json:"timestamp"`
}

//go:embed logs.html
var logsView string

// logsViewTemplate is the parsed template for the logs view
var logsViewTemplate = template.Must(template.New("logsView").Parse(logsView))

// LogsSkipper defines a function to skip recording a log entry.
// Return true to skip collecting the log into the monitor.
// The underlying logger still receives the log.
type LogsSkipper func(r slog.Record) bool

// monitorHandler wraps a slog.Handler and records each log to the monitor.
type monitorHandler struct {
	next    slog.Handler
	monitor *debugmonitor.Monitor
	skipper LogsSkipper
}

func (h *monitorHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return h.next.Enabled(ctx, level)
}

func (h *monitorHandler) Handle(ctx context.Context, r slog.Record) error {
	if h.skipper == nil || !h.skipper(r) {
		h.monitor.Add(&LogPayload{
			Level:     r.Level.String(),
			Message:   r.Message,
			Timestamp: r.Time,
		})
	}
	return h.next.Handle(ctx, r)
}

func (h *monitorHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &monitorHandler{next: h.next.WithAttrs(attrs), monitor: h.monitor, skipper: h.skipper}
}

func (h *monitorHandler) WithGroup(name string) slog.Handler {
	return &monitorHandler{next: h.next.WithGroup(name), monitor: h.monitor, skipper: h.skipper}
}

// LogsMonitorConfig defines the config for Logs monitor.
type LogsMonitorConfig struct {
	// Logger is the original slog.Logger whose handler will be wrapped.
	// If nil, slog.Default() is used.
	Logger *slog.Logger
	// Skipper defines a function to skip recording a log entry.
	// Optional. Default: never skip.
	// Useful for excluding middleware.RequestLogger entries
	// (messages "REQUEST" / "REQUEST_ERROR"), which belong in RequestsMonitor.
	Skipper LogsSkipper
	// UsePolling enables polling mode instead of SSE for real-time updates.
	UsePolling bool
}

// NewLogsMonitor creates a new monitor for logging and returns
// the monitor along with a wrapped slog.Logger.
func NewLogsMonitor(config LogsMonitorConfig) (*debugmonitor.Monitor, *slog.Logger) {
	m := &debugmonitor.Monitor{
		Name:        "logs",
		DisplayName: "Logs",
		MaxRecords:  1000,
		Icon:        debugmonitor.IconDocumentText,
		ActionHandler: func(c *echo.Context, store *debugmonitor.Store, action string) error {
			switch action {
			case "render":
				return debugmonitor.RenderTemplate(c, logsViewTemplate, map[string]any{
					"UsePolling": config.UsePolling,
				})
			case "stream":
				return debugmonitor.HandleSSEStream(c, store)
			case "data":
				return debugmonitor.HandleDataJSON(c, store)
			default:
				return echo.NewHTTPError(http.StatusBadRequest, "unknown action")
			}
		},
	}

	base := config.Logger
	if base == nil {
		base = slog.Default()
	}

	return m, slog.New(&monitorHandler{
		next:    base.Handler(),
		monitor: m,
		skipper: config.Skipper,
	})
}
