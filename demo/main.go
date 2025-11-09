package main

import (
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
		TableHeader: `<thead><tr class="border-b dark:border-b-gray-700 border-b-gray-200 [&>th]:px-4 [&>th]:py-2 [&>th]:text-sm [&>th]:font-semibold [&>th]:table-cell"><th>Id</th><th>Time</th></tr></thead>`,
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

	if err := e.Start(":8080"); err != nil {
		e.Logger.Fatal(err)
	}
}
