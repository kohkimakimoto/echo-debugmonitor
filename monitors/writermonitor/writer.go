package writermonitor

import (
	"io"
	"net/http"

	debugmonitor "github.com/kohkimakimoto/echo-debugmonitor"
	"github.com/kohkimakimoto/echo-viewkit/pongo2"
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
				return renderMainView(ctx)
			default:
				return echo.NewHTTPError(http.StatusBadRequest)
			}
		},
	}
	return m, &TeeWriter{original: w, monitor: m}
}

func renderMainView(ctx *debugmonitor.MonitorViewContext) error {
	// Get the latest records from the store
	entries := ctx.Store().GetLatest(100)

	// Prepare data for template
	records := make([]map[string]any, 0, len(entries))
	for _, entry := range entries {
		payload, ok := entry.Payload.(*Payload)
		if !ok {
			continue
		}
		records = append(records, map[string]any{
			"id":   entry.Id,
			"data": payload.Data,
		})
	}

	data := pongo2.Context{
		"records": records,
	}

	return ctx.Render(http.StatusOK, mainView, data)
}

const mainView = `
<div class="overflow-x-auto w-full rounded border dark:border-gray-700 border-gray-200">
  <table class="w-full">
    <thead>
      <tr class="border-b dark:bg-gray-700 bg-gray-50 dark:border-b-gray-700 border-b-gray-200">
        <th class="py-2 px-4 text-xs text-left">Output</th>
      </tr>
    </thead>
    <tbody class="bg-white dark:bg-gray-800">
      {% for record in records %}
        <tr class="border-b dark:border-b-gray-700 border-b-gray-200 last:border-0">
          <td class="py-2 px-4 font-mono text-xs text-left align-top whitespace-pre-wrap break-all">{{ record.data }}</td>
        </tr>
      {% empty %}
        <tr>
          <td colspan="2" class="px-4 py-8 text-center text-gray-500 dark:text-gray-400 text-sm">
            No data written yet
          </td>
        </tr>
      {% endfor %}
    </tbody>
  </table>
</div>
`
