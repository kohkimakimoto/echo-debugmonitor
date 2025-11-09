package main

import (
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
		Renderer: func(c echo.Context, monitor *debugmonitor.Monitor) (string, error) {
			return debugmonitor.ExecuteMonitoTemplateString(`aaaa`, nil)
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
