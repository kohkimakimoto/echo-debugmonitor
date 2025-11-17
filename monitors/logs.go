package monitors

import (
	_ "embed"
	"fmt"
	"io"
	"net/http"
	"time"

	debugmonitor "github.com/kohkimakimoto/echo-debugmonitor"
	"github.com/labstack/echo/v4"
	"github.com/labstack/gommon/log"
)

// LogPayload represents the data structure for log monitoring
type LogPayload struct {
	Level     string    `json:"level"`
	Message   string    `json:"message"`
	Timestamp time.Time `json:"timestamp"`
}

//go:embed logs.html
var logsView string

// LoggerWrapper wraps an echo.Logger and intercepts all logging calls
type LoggerWrapper struct {
	original echo.Logger
	monitor  *debugmonitor.Monitor
}

// NewLogsMonitor creates a new monitor for logging and returns
// the monitor along with a wrapped logger
func NewLogsMonitor(logger echo.Logger) (*debugmonitor.Monitor, echo.Logger) {
	m := &debugmonitor.Monitor{
		Name:        "logs",
		DisplayName: "Logs",
		MaxRecords:  1000,
		Icon:        debugmonitor.IconDocumentText,
		ActionHandler: func(c echo.Context, store *debugmonitor.Store, action string) error {
			switch action {
			case "render":
				return c.HTML(http.StatusOK, logsView)
			case "stream":
				// SSE endpoint for real-time updates
				return debugmonitor.HandleSSEStream(c, store)
			default:
				return echo.NewHTTPError(http.StatusBadRequest)
			}
		},
	}

	wrapper := &LoggerWrapper{
		original: logger,
		monitor:  m,
	}

	return m, wrapper
}

// addLog is a helper function to add log entries to the monitor
func (l *LoggerWrapper) addLog(level string, message string) {
	l.monitor.Add(&LogPayload{
		Level:     level,
		Message:   message,
		Timestamp: time.Now(),
	})
}

// Output returns the output writer
func (l *LoggerWrapper) Output() io.Writer {
	return l.original.Output()
}

// SetOutput sets the output writer
func (l *LoggerWrapper) SetOutput(w io.Writer) {
	l.original.SetOutput(w)
}

// Prefix returns the prefix
func (l *LoggerWrapper) Prefix() string {
	return l.original.Prefix()
}

// SetPrefix sets the prefix
func (l *LoggerWrapper) SetPrefix(p string) {
	l.original.SetPrefix(p)
}

// Level returns the log level
func (l *LoggerWrapper) Level() log.Lvl {
	return l.original.Level()
}

// SetLevel sets the log level
func (l *LoggerWrapper) SetLevel(v log.Lvl) {
	l.original.SetLevel(v)
}

// SetHeader sets the log header
func (l *LoggerWrapper) SetHeader(h string) {
	l.original.SetHeader(h)
}

// Print logs a message at print level
func (l *LoggerWrapper) Print(i ...interface{}) {
	l.original.Print(i...)
	l.addLog("PRINT", fmt.Sprint(i...))
}

// Printf logs a formatted message at print level
func (l *LoggerWrapper) Printf(format string, args ...interface{}) {
	l.original.Printf(format, args...)
	l.addLog("PRINT", fmt.Sprintf(format, args...))
}

// Printj logs a JSON message at print level
func (l *LoggerWrapper) Printj(j log.JSON) {
	l.original.Printj(j)
	l.addLog("PRINT", fmt.Sprintf("%v", j))
}

// Debug logs a message at debug level
func (l *LoggerWrapper) Debug(i ...interface{}) {
	l.original.Debug(i...)
	l.addLog("DEBUG", fmt.Sprint(i...))
}

// Debugf logs a formatted message at debug level
func (l *LoggerWrapper) Debugf(format string, args ...interface{}) {
	l.original.Debugf(format, args...)
	l.addLog("DEBUG", fmt.Sprintf(format, args...))
}

// Debugj logs a JSON message at debug level
func (l *LoggerWrapper) Debugj(j log.JSON) {
	l.original.Debugj(j)
	l.addLog("DEBUG", fmt.Sprintf("%v", j))
}

// Info logs a message at info level
func (l *LoggerWrapper) Info(i ...interface{}) {
	l.original.Info(i...)
	l.addLog("INFO", fmt.Sprint(i...))
}

// Infof logs a formatted message at info level
func (l *LoggerWrapper) Infof(format string, args ...interface{}) {
	l.original.Infof(format, args...)
	l.addLog("INFO", fmt.Sprintf(format, args...))
}

// Infoj logs a JSON message at info level
func (l *LoggerWrapper) Infoj(j log.JSON) {
	l.original.Infoj(j)
	l.addLog("INFO", fmt.Sprintf("%v", j))
}

// Warn logs a message at warn level
func (l *LoggerWrapper) Warn(i ...interface{}) {
	l.original.Warn(i...)
	l.addLog("WARN", fmt.Sprint(i...))
}

// Warnf logs a formatted message at warn level
func (l *LoggerWrapper) Warnf(format string, args ...interface{}) {
	l.original.Warnf(format, args...)
	l.addLog("WARN", fmt.Sprintf(format, args...))
}

// Warnj logs a JSON message at warn level
func (l *LoggerWrapper) Warnj(j log.JSON) {
	l.original.Warnj(j)
	l.addLog("WARN", fmt.Sprintf("%v", j))
}

// Error logs a message at error level
func (l *LoggerWrapper) Error(i ...interface{}) {
	l.original.Error(i...)
	l.addLog("ERROR", fmt.Sprint(i...))
}

// Errorf logs a formatted message at error level
func (l *LoggerWrapper) Errorf(format string, args ...interface{}) {
	l.original.Errorf(format, args...)
	l.addLog("ERROR", fmt.Sprintf(format, args...))
}

// Errorj logs a JSON message at error level
func (l *LoggerWrapper) Errorj(j log.JSON) {
	l.original.Errorj(j)
	l.addLog("ERROR", fmt.Sprintf("%v", j))
}

// Fatal logs a message at fatal level
func (l *LoggerWrapper) Fatal(i ...interface{}) {
	l.addLog("FATAL", fmt.Sprint(i...))
	l.original.Fatal(i...)
}

// Fatalf logs a formatted message at fatal level
func (l *LoggerWrapper) Fatalf(format string, args ...interface{}) {
	l.addLog("FATAL", fmt.Sprintf(format, args...))
	l.original.Fatalf(format, args...)
}

// Fatalj logs a JSON message at fatal level
func (l *LoggerWrapper) Fatalj(j log.JSON) {
	l.addLog("FATAL", fmt.Sprintf("%v", j))
	l.original.Fatalj(j)
}

// Panic logs a message at panic level
func (l *LoggerWrapper) Panic(i ...interface{}) {
	l.addLog("PANIC", fmt.Sprint(i...))
	l.original.Panic(i...)
}

// Panicf logs a formatted message at panic level
func (l *LoggerWrapper) Panicf(format string, args ...interface{}) {
	l.addLog("PANIC", fmt.Sprintf(format, args...))
	l.original.Panicf(format, args...)
}

// Panicj logs a JSON message at panic level
func (l *LoggerWrapper) Panicj(j log.JSON) {
	l.addLog("PANIC", fmt.Sprintf("%v", j))
	l.original.Panicj(j)
}
