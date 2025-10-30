package main

import (
	debugmonitor "github.com/kohkimakimoto/echo-debugmonitor"
	"github.com/labstack/echo/v4"
)

func main() {
	e := echo.New()

	dm := debugmonitor.New()
	monitor1 := &debugmonitor.Monitor{
		Name:        "example_monitor",
		DisplayName: "Example Monitor",
		MaxRecords:  100,
	}
	dm.AddMonitor(monitor1)
	e.Any("/monitor", dm.Handler())

	if err := e.Start(":8080"); err != nil {
		e.Logger.Fatal(err)
	}
}
