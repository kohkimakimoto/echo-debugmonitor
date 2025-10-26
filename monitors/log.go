package monitors

import (
	debugmonitor "github.com/kohkimakimoto/echo-debugmonitor"
	"github.com/labstack/echo/v4"
)

func LogWatcher() *debugmonitor.Monitor {
	return nil
}

type LoggerAdapter struct {
	w      *debugmonitor.Monitor
	logger echo.Logger
}

func NewLoggerAdapter(w *debugmonitor.Monitor, logger echo.Logger) *LoggerAdapter {
	return &LoggerAdapter{
		w:      w,
		logger: logger,
	}
}
