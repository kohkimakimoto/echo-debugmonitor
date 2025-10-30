package main

import (
	debugmonitor "github.com/kohkimakimoto/echo-debugmonitor"
	"github.com/labstack/echo/v4"
)

func main() {
	e := echo.New()

	dm := debugmonitor.New()
	monitor1 := &debugmonitor.Monitor{
		Name:        "example_monitor1",
		DisplayName: "Example Monitor1",
		MaxRecords:  100,
	}
	dm.AddMonitor(monitor1)
	monitor2 := &debugmonitor.Monitor{
		Name:        "example_monitor2",
		DisplayName: "Example Monitor2",
		MaxRecords:  100,
	}
	dm.AddMonitor(monitor2)
	e.Any("/monitor", dm.Handler())

	if err := e.Start(":8080"); err != nil {
		e.Logger.Fatal(err)
	}
}
