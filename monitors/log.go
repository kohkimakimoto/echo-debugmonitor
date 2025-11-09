package monitors

import (
	debugmonitor "github.com/kohkimakimoto/echo-debugmonitor"
	"github.com/labstack/echo/v4"
)

type LogMonitorConfig struct {
	Logger echo.Logger
}

func LogMonitor(config LogMonitorConfig) (*debugmonitor.Monitor, echo.Logger) {
	return nil, nil
}
