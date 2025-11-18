package monitors

import (
	_ "embed"
	"html/template"
	"io"
	"net/http"

	debugmonitor "github.com/kohkimakimoto/echo-debugmonitor"
	"github.com/labstack/echo/v4"
)

type WriterPayload struct {
	Data string `json:"data"`
}

type TeeWriter struct {
	original io.Writer
	monitor  *debugmonitor.Monitor
}

func (t *TeeWriter) Write(p []byte) (n int, err error) {
	// Add to the original writer
	n, err = t.original.Write(p)
	if err != nil {
		return n, err
	}

	// Also send the payload to the monitor
	t.monitor.Add(&WriterPayload{
		Data: string(p),
	})

	return n, nil
}

// LoggerWriterMonitorConfig is the configuration for the logger writer monitor.
type LoggerWriterMonitorConfig struct {
	// Logger is the echo.Logger to wrap with monitoring.
	Logger echo.Logger
	// UsePolling enables polling mode instead of SSE for real-time updates.
	UsePolling bool
}

// NewLoggerWriterMonitor creates a logger writer monitor with the given configuration.
func NewLoggerWriterMonitor(config LoggerWriterMonitorConfig) *debugmonitor.Monitor {
	o := config.Logger.Output()
	m, w := NewWriterMonitor(WriterMonitorConfig{
		UsePolling: config.UsePolling,
		Writer:     o,
	})
	m.Name = "logger_writer"
	m.DisplayName = "Logger Writer"
	config.Logger.SetOutput(w)
	return m
}

//go:embed writer.html
var writerView string

// writerViewTemplate is the parsed template for the writer view
var writerViewTemplate = template.Must(template.New("writerView").Parse(writerView))

// WriterMonitorConfig is the configuration for the writer monitor.
type WriterMonitorConfig struct {
	// Writer is the original io.Writer to write to.
	Writer io.Writer
	// UsePolling enables polling mode instead of SSE for real-time updates.
	UsePolling bool
}

// NewWriterMonitor creates a new writer monitor with the given configuration.
// It returns the monitor and a new io.Writer that writes to both the original writer
// and the monitor's store.
func NewWriterMonitor(config WriterMonitorConfig) (*debugmonitor.Monitor, io.Writer) {
	m := &debugmonitor.Monitor{
		Name:        "writer",
		DisplayName: "Writer",
		MaxRecords:  1000,
		Icon:        debugmonitor.IconPencilSquare,
		ActionHandler: func(c echo.Context, store *debugmonitor.Store, action string) error {
			switch action {
			case "render":
				return debugmonitor.RenderTemplate(c, writerViewTemplate, map[string]any{
					"UsePolling": config.UsePolling,
				})
			case "stream":
				// SSE endpoint for real-time updates
				return debugmonitor.HandleSSEStream(c, store)
			case "data":
				// JSON endpoint for polling mode
				return debugmonitor.HandleDataJSON(c, store)
			default:
				return echo.NewHTTPError(http.StatusBadRequest)
			}
		},
	}
	return m, &TeeWriter{original: config.Writer, monitor: m}
}
