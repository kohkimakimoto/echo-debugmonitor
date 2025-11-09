package monitors

import (
	debugmonitor "github.com/kohkimakimoto/echo-debugmonitor"
	"github.com/labstack/echo/v4"
)

type LogMonitorConfig struct {
	Logger echo.Logger
}

type LoggerAdapter struct {
}

func NewLogMonitor(config *LogMonitorConfig) (*debugmonitor.Monitor, echo.Logger) {
	return nil, nil
}
