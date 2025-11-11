package main

import (
	debugmonitor "github.com/kohkimakimoto/echo-debugmonitor"
	"github.com/kohkimakimoto/echo-debugmonitor/monitors"
	"github.com/labstack/echo/v4"
)

func main() {
	e := echo.New()

	m := debugmonitor.New()

	// logger writer monitor
	m.AddMonitor(monitors.NewLoggerWriterMonitor(e.Logger))

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
