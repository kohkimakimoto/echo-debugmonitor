package main

import (
	"net/http"
	"time"

	debugmonitor "github.com/kohkimakimoto/echo-debugmonitor"
	"github.com/labstack/echo/v4"
)

func main() {
	e := echo.New()

	dm := debugmonitor.New()
	monitor1 := &debugmonitor.Monitor{
		Name:        "requests",
		DisplayName: "Requests",
		MaxRecords:  100,
		Icon:        debugmonitor.IconExclamationCircle,
		ViewHandler: func(c echo.Context, monitor *debugmonitor.Monitor) error {
			switch c.QueryParam("action") {
			case "renderMainView":
				return c.HTML(http.StatusOK, "<h1>"+c.Request().URL.Path+"</h1>")
			default:
				return echo.NewHTTPError(http.StatusBadRequest)
			}
		},
	}
	dm.AddMonitor(monitor1)
	monitor2 := &debugmonitor.Monitor{
		Name:        "example_monitor2",
		DisplayName: "Example Monitor2",
		MaxRecords:  100,
		Icon:        debugmonitor.IconCircleStack,
	}
	dm.AddMonitor(monitor2)
	e.Any("/monitor", dm.Handler())

	go func() {
		ticker := time.NewTicker(1 * time.Second)
		defer ticker.Stop()
		for range ticker.C {
			currentTime := time.Now().Format("2006-01-02 15:04:05")

			_ = monitor1.Write(map[string]any{
				"time": currentTime,
			})
		}
	}()

	if err := e.Start(":8080"); err != nil {
		e.Logger.Fatal(err)
	}
}

const view = `
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
