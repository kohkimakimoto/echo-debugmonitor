package writermonitor

import (
	"io"
	"net/http"

	debugmonitor "github.com/kohkimakimoto/echo-debugmonitor"
	"github.com/labstack/echo/v4"
)

type Payload struct {
	Data string
}

// TeeWriter writes to both the original writer and the monitor
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

	payload := &Payload{
		Data: string(p),
	}
	// Also send the payload to the monitor
	t.monitor.Write(payload)

	return n, nil
}

func New(w io.Writer) (*debugmonitor.Monitor, io.Writer) {
	m := &debugmonitor.Monitor{
		Name:        "writer",
		DisplayName: "Writer",
		MaxRecords:  1000,
		Icon:        debugmonitor.IconCircleStack,
		ViewHandler: func(ctx *debugmonitor.MonitorViewContext) error {
			switch ctx.EchoContext().QueryParam("action") {
			case "renderMainView":
				return ctx.Render(http.StatusOK, mainView, nil)
			default:
				return echo.NewHTTPError(http.StatusBadRequest)
			}
		},
	}
	return m, &TeeWriter{original: w, monitor: m}
}

const mainView = `
<div class="overflow-x-auto w-full rounded border dark:border-gray-700 border-gray-200">
  <table class="w-full">
    <thead>
      <tr class="border-b dark:bg-gray-700 bg-gray-50 dark:border-b-gray-700 border-b-gray-200 [&>th]:px-4 [&>th]:py-2 [&>th]:text-sm [&>th]:font-semibold [&>th]:table-cell">
        <th>Id</th>
        <th>Time</th>
      </tr>
    </thead>
    <tbody>
    </tbody>
  </table>
</div>
`
