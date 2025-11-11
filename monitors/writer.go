package monitors

import (
	"io"
	"net/http"

	debugmonitor "github.com/kohkimakimoto/echo-debugmonitor"
	"github.com/labstack/echo/v4"
)

type WriterPayload struct {
	Data string
}

type TeeWriter struct {
	original io.Writer
	monitor  *debugmonitor.Monitor
}

func (t *TeeWriter) Write(p []byte) (n int, err error) {
	// Write to the original writer
	n, err = t.original.Write(p)
	if err != nil {
		return n, err
	}

	// Also send the payload to the monitor
	t.monitor.Write(&WriterPayload{
		Data: string(p),
	})

	return n, nil
}

func NewWriterMonitor(w io.Writer) (*debugmonitor.Monitor, io.Writer) {
	m := &debugmonitor.Monitor{
		Name:        "writer",
		DisplayName: "Writer",
		MaxRecords:  1000,
		Icon:        debugmonitor.IconCircleStack,
		ActionHandler: func(c echo.Context, monitor *debugmonitor.Monitor, action string) error {
			switch action {
			case "renderMainView":
				return echo.NewHTTPError(http.StatusBadRequest)
			default:
				return echo.NewHTTPError(http.StatusBadRequest)
			}

		},
	}
	return m, &TeeWriter{original: w, monitor: m}
}

func NewLoggerWriterMonitor(logger echo.Logger) *debugmonitor.Monitor {
	o := logger.Output()
	m, w := NewWriterMonitor(o)
	m.Name = "logger_writer"
	m.DisplayName = "Logger Writer"
	logger.SetOutput(w)
	return m
}
