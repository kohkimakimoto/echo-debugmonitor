package debugmonitor

import (
	"github.com/labstack/echo/v4"
)

type DebugMonitor struct {
}

func New() *DebugMonitor {
	return &DebugMonitor{}
}

func Handler(m *DebugMonitor) echo.HandlerFunc {
	return func(c echo.Context) error {
		return nil
	}
}
