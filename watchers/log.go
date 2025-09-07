package watchers

import (
	debugmonitor "github.com/kohkimakimoto/echo-debugmonitor"
	"github.com/labstack/echo/v4"
)

func LogWatcher() *debugmonitor.Watcher {
	return nil
}

type LoggerAdapter struct {
	w      *debugmonitor.Watcher
	logger echo.Logger
}

func NewLoggerAdapter(w *debugmonitor.Watcher, logger echo.Logger) *LoggerAdapter {
	return &LoggerAdapter{
		w:      w,
		logger: logger,
	}
}
