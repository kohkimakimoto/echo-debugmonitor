package monitors

import (
	debugmonitor "github.com/kohkimakimoto/echo-debugmonitor"
	"github.com/labstack/echo/v4"
)

type RequestsMonitorConfig struct {
}

func NewRequestsMonitor(config *RequestsMonitorConfig) (*debugmonitor.Monitor, echo.MiddlewareFunc) {
	return nil, nil
}
