package main

import (
	debugmonitor "github.com/kohkimakimoto/echo-debugmonitor"
	"github.com/kohkimakimoto/echo-debugmonitor/monitors"
	"github.com/labstack/echo/v4"
)

func main() {
	e := echo.New()

	m := debugmonitor.New()
	logMonitor, logger := monitors.NewLogMonitor(&monitors.LogMonitorConfig{
		Logger: e.Logger,
	})
	e.Logger = logger
	m.AddMonitor(logMonitor)
	e.Any("/monitor", m.Handler())

	// Add a test endpoint that logs messages
	e.GET("/test", func(c echo.Context) error {
		e.Logger.Info("Test endpoint called")
		return c.String(200, "Test endpoint - check the monitor!")
	})

	//
	//monitor1 := &debugmonitor.Monitor{
	//	Name:        "requests",
	//	DisplayName: "Requests",
	//	MaxRecords:  100,
	//	Icon:        debugmonitor.IconExclamationCircle,
	//	ViewHandler: func(ctx *debugmonitor.MonitorViewContext) error {
	//		switch ctx.EchoContext().QueryParam("action") {
	//		case "renderMainView":
	//			return ctx.Render(http.StatusOK, mainView, nil)
	//		default:
	//			return echo.NewHTTPError(http.StatusBadRequest)
	//		}
	//	},
	//}
	//dm.AddMonitor(monitor1)
	//monitor2 := &debugmonitor.Monitor{
	//	Name:        "example_monitor2",
	//	DisplayName: "Example Monitor2",
	//	MaxRecords:  100,
	//	Icon:        debugmonitor.IconCircleStack,
	//}
	//dm.AddMonitor(monitor2)

	//go func() {
	//	ticker := time.NewTicker(1 * time.Second)
	//	defer ticker.Stop()
	//	for range ticker.C {
	//		currentTime := time.Now().Format("2006-01-02 15:04:05")
	//
	//		monitor1.Write(map[string]any{
	//			"time": currentTime,
	//		})
	//	}
	//}()

	if err := e.Start(":8080"); err != nil {
		e.Logger.Fatal(err)
	}
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
