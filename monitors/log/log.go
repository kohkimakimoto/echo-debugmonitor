package log

import (
	"io"
	"net/http"

	debugmonitor "github.com/kohkimakimoto/echo-debugmonitor"
	"github.com/labstack/echo/v4"
	"github.com/labstack/gommon/log"
)

type LoggerAdapter struct {
	monitor *debugmonitor.Monitor
	logger  echo.Logger
}

func (l *LoggerAdapter) Output() io.Writer {
	return l.logger.Output()
}

func (l *LoggerAdapter) SetOutput(w io.Writer) {
	l.logger.SetOutput(w)
}

func (l *LoggerAdapter) Prefix() string {
	return l.logger.Prefix()
}

func (l *LoggerAdapter) SetPrefix(p string) {
	l.logger.SetPrefix(p)
}

func (l *LoggerAdapter) Level() log.Lvl {
	return l.logger.Level()
}

func (l *LoggerAdapter) SetLevel(v log.Lvl) {
	l.logger.SetLevel(v)
}

func (l *LoggerAdapter) SetHeader(h string) {
	l.logger.SetHeader(h)
}

func (l *LoggerAdapter) Print(i ...interface{}) {
	l.logger.Print(i...)
}

func (l *LoggerAdapter) Printf(format string, args ...interface{}) {
	l.logger.Printf(format, args...)
}

func (l *LoggerAdapter) Printj(j log.JSON) {
	l.logger.Printj(j)
}

func (l *LoggerAdapter) Debug(i ...interface{}) {
	l.logger.Debug(i...)
}

func (l *LoggerAdapter) Debugf(format string, args ...interface{}) {
	l.logger.Debugf(format, args...)
}

func (l *LoggerAdapter) Debugj(j log.JSON) {
	l.logger.Debugj(j)
}

func (l *LoggerAdapter) Info(i ...interface{}) {
	l.logger.Info(i...)
}

func (l *LoggerAdapter) Infof(format string, args ...interface{}) {
	l.logger.Infof(format, args...)
}

func (l *LoggerAdapter) Infoj(j log.JSON) {
	l.logger.Infoj(j)
}

func (l *LoggerAdapter) Warn(i ...interface{}) {
	l.logger.Warn(i...)
}

func (l *LoggerAdapter) Warnf(format string, args ...interface{}) {
	l.logger.Warnf(format, args...)
}

func (l *LoggerAdapter) Warnj(j log.JSON) {
	l.logger.Warnj(j)
}

func (l *LoggerAdapter) Error(i ...interface{}) {
	l.logger.Error(i...)
}

func (l *LoggerAdapter) Errorf(format string, args ...interface{}) {
	l.logger.Errorf(format, args...)
}

func (l *LoggerAdapter) Errorj(j log.JSON) {
	l.logger.Errorj(j)
}

func (l *LoggerAdapter) Fatal(i ...interface{}) {
	l.logger.Fatal(i...)
}

func (l *LoggerAdapter) Fatalj(j log.JSON) {
	l.logger.Fatalj(j)
}

func (l *LoggerAdapter) Fatalf(format string, args ...interface{}) {
	l.logger.Fatalf(format, args...)
}

func (l *LoggerAdapter) Panic(i ...interface{}) {
	l.logger.Panic(i...)
}

func (l *LoggerAdapter) Panicj(j log.JSON) {
	l.logger.Panicj(j)
}

func (l *LoggerAdapter) Panicf(format string, args ...interface{}) {
	l.logger.Panicf(format, args...)
}

// TeeWriter writes to both the original writer and the monitor
type TeeWriter struct {
	original io.Writer
	monitor  *debugmonitor.Monitor
}

//func (t *TeeWriter) Write(p []byte) (n int, err error) {
//	// Write to the original writer
//	n, err = t.original.Write(p)
//	if err != nil {
//		return n, err
//	}
//
//	// Also send the log data to the monitor
//	t.monitor.Write(string(p))
//
//	return n, nil
//}

type Config struct {
	Logger echo.Logger
}

func New(config *Config) (*debugmonitor.Monitor, echo.Logger) {
	m := &debugmonitor.Monitor{
		Name:        "log",
		DisplayName: "Log",
		MaxRecords:  1000,
		Icon:        debugmonitor.IconCircleStack,
		ViewHandler: func(ctx *debugmonitor.MonitorViewContext) error {
			switch ctx.EchoContext().QueryParam("action") {
			case "renderMainView":
				return ctx.Render(http.StatusOK, logMonitorMainView, nil)
			default:
				return echo.NewHTTPError(http.StatusBadRequest)
			}
		},
	}

	logger := config.Logger
	originalOutput := logger.Output()

	// Create a TeeWriter that sends output to both the original writer and the monitor
	teeWriter := &TeeWriter{
		original: originalOutput,
		monitor:  m,
	}

	// Set the logger to use the TeeWriter
	logger.SetOutput(teeWriter)

	return m, &LoggerAdapter{monitor: m, logger: logger}
}

const logMonitorMainView = `
<div class="overflow-x-auto w-full rounded border dark:border-gray-700 border-gray-200">
  <table class="w-full">
    <thead>
      <tr class="border-b dark:bg-gray-700 bg-gray-50 dark:border-b-gray-700 border-b-gray-200 [&>th]:px-4 [&>th]:py-2 [&>th]:text-sm [&>th]:font-semibold [&>th]:table-cell">
        <th>Id</th>
        <th>Time</th>
      </tr>
    </thead>
    <tbody>
    </tbody>
  </table>
</div>
`
