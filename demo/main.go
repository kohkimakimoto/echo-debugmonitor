package main

import (
	debugmonitor "github.com/kohkimakimoto/echo-debugmonitor"
	writermonitor "github.com/kohkimakimoto/echo-debugmonitor/monitors/writermonitor"
	"github.com/labstack/echo/v4"
)

func main() {
	e := echo.New()

	m := debugmonitor.New()

	// write monitor
	wMonitor, out := writermonitor.New(e.Logger.Output())
	wMonitor.Name = "logger_output"
	wMonitor.DisplayName = "Logger Output"

	e.Logger.SetOutput(out)

	m.AddMonitor(wMonitor)

	m.AddMonitor(&debugmonitor.Monitor{
		Name:        "dummy",
		DisplayName: "Dummy Monitor",
	})

	e.Any("/monitor", m.Handler())

	// Add a test endpoint that logs messages
	e.GET("/test", func(c echo.Context) error {
		e.Logger.Errorf("Test endpoint called")
		return c.String(200, "Test endpoint - check the monitor!")
	})

	if err := e.Start(":8080"); err != nil {
		e.Logger.Fatal(err)
	}
}
