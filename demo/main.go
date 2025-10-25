package main

import (
	debugmonitor "github.com/kohkimakimoto/echo-debugmonitor"
	"github.com/labstack/echo/v4"
)

func main() {
	e := echo.New()

	m := debugmonitor.New()
	e.Any("/monitor", m.Handler())

	if err := e.Start(":8080"); err != nil {
		e.Logger.Fatal(err)
	}
}
