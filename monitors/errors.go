package monitors

import (
	_ "embed"
	"fmt"
	"html/template"
	"net/http"
	"time"

	debugmonitor "github.com/kohkimakimoto/echo-debugmonitor"
	"github.com/labstack/echo/v4"
)

// ErrorPayload represents the data structure for error monitoring
type ErrorPayload struct {
	Error      string    `json:"error"`
	Type       string    `json:"type"`
	Message    string    `json:"message"`
	StackTrace string    `json:"stackTrace"`
	Timestamp  time.Time `json:"timestamp"`
}

//go:embed errors.html
var errorsView string

// errorsViewTemplate is the parsed template for the errors view
var errorsViewTemplate = template.Must(template.New("errorsView").Parse(errorsView))

// ErrorRecorder is a function type for recording errors
type ErrorRecorder func(err error)

// ErrorsMonitorConfig defines the config for Errors monitor.
type ErrorsMonitorConfig struct {
	// UsePolling enables polling mode instead of SSE for real-time updates.
	UsePolling bool
}

// NewErrorsMonitor creates a new monitor for errors and returns
// the monitor along with an error recording function
func NewErrorsMonitor(config ErrorsMonitorConfig) (*debugmonitor.Monitor, ErrorRecorder) {
	m := &debugmonitor.Monitor{
		Name:        "errors",
		DisplayName: "Errors",
		MaxRecords:  1000,
		Icon:        debugmonitor.IconExclamationCircle,
		ActionHandler: func(c echo.Context, store *debugmonitor.Store, action string) error {
			switch action {
			case "render":
				return debugmonitor.RenderTemplate(c, errorsViewTemplate, map[string]any{
					"UsePolling": config.UsePolling,
				})
			case "stream":
				// SSE endpoint for real-time updates
				return debugmonitor.HandleSSEStream(c, store)
			case "data":
				// JSON endpoint for polling mode
				return debugmonitor.HandleDataJSON(c, store)
			default:
				return echo.NewHTTPError(http.StatusBadRequest)
			}
		},
	}

	// Create error recorder function
	recorder := func(err error) {
		if err == nil {
			return
		}

		// Get error type
		errorType := fmt.Sprintf("%T", err)

		// Get error message
		errorMessage := err.Error()

		// Extract stack trace from the error
		stackTrace := extractStackTrace(err)

		// Add error to monitor
		m.Add(&ErrorPayload{
			Error:      errorMessage,
			Type:       errorType,
			Message:    errorMessage,
			StackTrace: stackTrace,
			Timestamp:  time.Now(),
		})
	}

	return m, recorder
}

// HTTPErrorHandlerWrapper returns an echo.HTTPErrorHandler that records errors
// and then delegates to the provided handler
func HTTPErrorHandlerWrapper(recorder ErrorRecorder, handler echo.HTTPErrorHandler) echo.HTTPErrorHandler {
	return func(err error, c echo.Context) {
		// Record the error
		recorder(err)
		// Delegate to the original handler
		handler(err, c)
	}
}

// extractStackTrace attempts to extract stack trace information from an error
// It supports:
// 1. Errors formatted with %+v that include stack traces (e.g., errors wrapped with pkg/errors)
// 2. Errors that implement a StackTrace() method
// Returns empty string if no stack trace information is found
func extractStackTrace(err error) string {
	if err == nil {
		return ""
	}

	// Try to format the error with %+v to see if it includes stack trace information
	// This works with errors wrapped using github.com/pkg/errors or similar packages
	detailedError := fmt.Sprintf("%+v", err)

	// Check if the detailed format contains stack trace information
	// Stack traces typically contain file paths and line numbers
	if containsStackTrace(detailedError) {
		return detailedError
	}

	// Check if the error implements a StackTrace() method
	// This is a common interface for errors with stack traces
	type stackTracer interface {
		StackTrace() string
	}
	if st, ok := err.(stackTracer); ok {
		return st.StackTrace()
	}

	// No stack trace information found
	return ""
}

// containsStackTrace checks if a string appears to contain stack trace information
func containsStackTrace(s string) bool {
	// Simple heuristic: stack traces usually contain file paths with line numbers
	// Look for patterns like ".go:" which are common in Go stack traces
	return len(s) > 100 && (
	// Common patterns in stack traces
	containsPattern(s, ".go:") ||
		containsPattern(s, "goroutine ") ||
		containsPattern(s, "\tat ") ||
		containsPattern(s, "\n\t"))
}

// containsPattern checks if a string contains a specific pattern
func containsPattern(s, pattern string) bool {
	for i := 0; i <= len(s)-len(pattern); i++ {
		if s[i:i+len(pattern)] == pattern {
			return true
		}
	}
	return false
}
