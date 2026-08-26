package monitors

import (
	_ "embed"
	"html/template"
	"io"
	"net/http"

	debugmonitor "github.com/kohkimakimoto/echo-debugmonitor/v5"
	"github.com/labstack/echo/v5"
)

type WriterPayload struct {
	Data string `json:"data"`
}

type TeeWriter struct {
	original io.Writer
	monitor  *debugmonitor.Monitor
}

func (t *TeeWriter) Write(p []byte) (n int, err error) {
	n, err = t.original.Write(p)
	if err != nil {
		return n, err
	}

	t.monitor.Add(&WriterPayload{
		Data: string(p),
	})

	return n, nil
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
		ActionHandler: func(c *echo.Context, store *debugmonitor.Store, action string) error {
			switch action {
			case "render":
				return debugmonitor.RenderTemplate(c, writerViewTemplate, map[string]any{
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
	return m, &TeeWriter{original: config.Writer, monitor: m}
}
