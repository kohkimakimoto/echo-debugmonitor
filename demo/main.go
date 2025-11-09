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
		TableHeader: `<thead><tr class="border-b dark:border-b-gray-700 border-b-gray-200 [&>th]:px-4 [&>th]:py-2 [&>th]:text-sm [&>th]:font-semibold [&>th]:table-cell"><th>Id</th><th>Time</th></tr></thead>`,
		TableRowRenderer: func(c echo.Context, dataEntry *debugmonitor.DataEntry) string {
			payloadMap, ok := dataEntry.Payload.(map[string]any)
			if !ok {
				return ""
			}
			timeValue, _ := payloadMap["time"].(string)
			return `<tr class="border-b dark:border-b-gray-700 border-b-gray-200 [&>td]:px-4 [&>td]:py-2 [&>td]:text-sm [&>td]:table-cell"><td>` + dataEntry.Id + `</td><td>` + timeValue + `</td></tr>`
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
