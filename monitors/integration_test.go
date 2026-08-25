package monitors

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	debugmonitor "github.com/kohkimakimoto/echo-debugmonitor/v4"
	"github.com/labstack/echo/v4"
)

func TestRequestsMonitorRecordsResponse(t *testing.T) {
	e := echo.New()
	manager := debugmonitor.New()
	monitor, middleware := NewRequestsMonitor(nil)
	manager.AddMonitor(monitor)
	e.Use(middleware)
	e.GET("/created", func(c echo.Context) error {
		return c.String(http.StatusCreated, "created")
	})
	e.GET("/monitor", manager.Handler())

	response := httptest.NewRecorder()
	e.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/created", nil))
	if response.Code != http.StatusCreated {
		t.Fatalf("expected status %d, got %d", http.StatusCreated, response.Code)
	}

	entries := readMonitorEntries[RequestPayload](t, e, "requests")
	if len(entries) != 1 {
		t.Fatalf("expected one request entry, got %d", len(entries))
	}
	if entries[0].Payload.Method != http.MethodGet || entries[0].Payload.URI != "/created" {
		t.Fatalf("unexpected request payload: %#v", entries[0].Payload)
	}
	if entries[0].Payload.Status != http.StatusCreated {
		t.Fatalf("expected recorded status %d, got %d", http.StatusCreated, entries[0].Payload.Status)
	}
}

func TestRequestsMonitorRecordsHTTPError(t *testing.T) {
	e := echo.New()
	manager := debugmonitor.New()
	monitor, middleware := NewRequestsMonitor(nil)
	manager.AddMonitor(monitor)
	e.Use(middleware)
	e.GET("/teapot", func(c echo.Context) error {
		return echo.NewHTTPError(http.StatusTeapot, "short and stout")
	})
	e.GET("/monitor", manager.Handler())

	response := httptest.NewRecorder()
	e.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/teapot", nil))
	if response.Code != http.StatusTeapot {
		t.Fatalf("expected status %d, got %d", http.StatusTeapot, response.Code)
	}

	entries := readMonitorEntries[RequestPayload](t, e, "requests")
	if len(entries) != 1 {
		t.Fatalf("expected one request entry, got %d", len(entries))
	}
	if entries[0].Payload.Status != http.StatusTeapot {
		t.Fatalf("expected recorded status %d, got %d", http.StatusTeapot, entries[0].Payload.Status)
	}
	if entries[0].Payload.Error != "short and stout" {
		t.Fatalf("expected recorded error message, got %q", entries[0].Payload.Error)
	}
}

func TestHTTPErrorHandlerWrapperRecordsAndDelegates(t *testing.T) {
	e := echo.New()
	manager := debugmonitor.New()
	monitor, recorder := NewErrorsMonitor(ErrorsMonitorConfig{})
	manager.AddMonitor(monitor)
	e.GET("/monitor", manager.Handler())

	delegated := false
	handler := HTTPErrorHandlerWrapper(recorder, func(err error, c echo.Context) {
		delegated = true
	})
	wantErr := errors.New("handler failed")
	handler(wantErr, e.NewContext(
		httptest.NewRequest(http.MethodGet, "/failed", nil),
		httptest.NewRecorder(),
	))

	if !delegated {
		t.Fatal("expected wrapped error handler to delegate")
	}
	entries := readMonitorEntries[ErrorPayload](t, e, "errors")
	if len(entries) != 1 {
		t.Fatalf("expected one error entry, got %d", len(entries))
	}
	if entries[0].Payload.Message != wantErr.Error() {
		t.Fatalf("expected error %q, got %q", wantErr, entries[0].Payload.Message)
	}
}

func TestLogsMonitorRecordsMessage(t *testing.T) {
	e := echo.New()
	e.Logger.SetOutput(io.Discard)
	manager := debugmonitor.New()
	monitor, logger := NewLogsMonitor(LogsMonitorConfig{Logger: e.Logger})
	manager.AddMonitor(monitor)
	e.GET("/monitor", manager.Handler())

	logger.Info("server started")

	entries := readMonitorEntries[LogPayload](t, e, "logs")
	if len(entries) != 1 {
		t.Fatalf("expected one log entry, got %d", len(entries))
	}
	if entries[0].Payload.Level != "INFO" || entries[0].Payload.Message != "server started" {
		t.Fatalf("unexpected log payload: %#v", entries[0].Payload)
	}
}

type monitorEntry[T any] struct {
	ID      int64 `json:"id"`
	Payload T     `json:"payload"`
}

func readMonitorEntries[T any](t *testing.T, e *echo.Echo, monitorName string) []monitorEntry[T] {
	t.Helper()
	recorder := httptest.NewRecorder()
	target := "/monitor?monitor=" + monitorName + "&action=data"
	e.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, target, nil))
	if recorder.Code != http.StatusOK {
		t.Fatalf("expected monitor data status %d, got %d: %s", http.StatusOK, recorder.Code, recorder.Body.String())
	}

	var entries []monitorEntry[T]
	if err := json.Unmarshal(recorder.Body.Bytes(), &entries); err != nil {
		t.Fatalf("decode monitor data: %v", err)
	}
	return entries
}
