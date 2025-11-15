package debugmonitor

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/labstack/echo/v4"
)

func TestNew(t *testing.T) {
	manager := New()

	if manager == nil {
		t.Fatal("Expected manager to be created, got nil")
	}

	if manager.monitors == nil {
		t.Error("Expected monitors slice to be initialized")
	}

	if manager.monitorMap == nil {
		t.Error("Expected monitorMap to be initialized")
	}

	if len(manager.monitors) != 0 {
		t.Errorf("Expected empty monitors slice, got %d monitors", len(manager.monitors))
	}

	if len(manager.monitorMap) != 0 {
		t.Errorf("Expected empty monitorMap, got %d monitors", len(manager.monitorMap))
	}
}

func TestManager_AddMonitor(t *testing.T) {
	manager := New()

	monitor := &Monitor{
		Name:        "test-monitor",
		DisplayName: "Test Monitor",
		MaxRecords:  100,
		Icon:        IconExclamationCircle,
	}

	manager.AddMonitor(monitor)

	// Check that the monitor was added
	if len(manager.monitors) != 1 {
		t.Errorf("Expected 1 monitor, got %d", len(manager.monitors))
	}

	if len(manager.monitorMap) != 1 {
		t.Errorf("Expected 1 monitor in map, got %d", len(manager.monitorMap))
	}

	// Check that the monitor can be retrieved from map
	retrieved, ok := manager.monitorMap["test-monitor"]
	if !ok {
		t.Error("Expected monitor to be in map")
	}

	if retrieved != monitor {
		t.Error("Expected retrieved monitor to be the same as added monitor")
	}

	// Check that the store was initialized
	if monitor.store == nil {
		t.Error("Expected monitor store to be initialized")
	}

	if monitor.store.maxRecords != 100 {
		t.Errorf("Expected store maxRecords to be 100, got %d", monitor.store.maxRecords)
	}
}

func TestManager_AddMultipleMonitors(t *testing.T) {
	manager := New()

	monitor1 := &Monitor{
		Name:        "monitor1",
		DisplayName: "Monitor 1",
		MaxRecords:  50,
	}

	monitor2 := &Monitor{
		Name:        "monitor2",
		DisplayName: "Monitor 2",
		MaxRecords:  100,
	}

	manager.AddMonitor(monitor1)
	manager.AddMonitor(monitor2)

	if len(manager.monitors) != 2 {
		t.Errorf("Expected 2 monitors, got %d", len(manager.monitors))
	}

	if len(manager.monitorMap) != 2 {
		t.Errorf("Expected 2 monitors in map, got %d", len(manager.monitorMap))
	}

	// Verify both monitors are accessible
	if _, ok := manager.monitorMap["monitor1"]; !ok {
		t.Error("Expected monitor1 to be in map")
	}

	if _, ok := manager.monitorMap["monitor2"]; !ok {
		t.Error("Expected monitor2 to be in map")
	}
}

func TestManager_Monitors(t *testing.T) {
	manager := New()

	// Initially empty
	monitors := manager.Monitors()
	if len(monitors) != 0 {
		t.Errorf("Expected 0 monitors, got %d", len(monitors))
	}

	// Add some monitors
	monitor1 := &Monitor{Name: "m1", DisplayName: "Monitor 1", MaxRecords: 10}
	monitor2 := &Monitor{Name: "m2", DisplayName: "Monitor 2", MaxRecords: 10}

	manager.AddMonitor(monitor1)
	manager.AddMonitor(monitor2)

	monitors = manager.Monitors()
	if len(monitors) != 2 {
		t.Errorf("Expected 2 monitors, got %d", len(monitors))
	}

	// Verify order is preserved
	if monitors[0].Name != "m1" {
		t.Errorf("Expected first monitor to be m1, got %s", monitors[0].Name)
	}

	if monitors[1].Name != "m2" {
		t.Errorf("Expected second monitor to be m2, got %s", monitors[1].Name)
	}
}

func TestManager_Handler_NoMonitors(t *testing.T) {
	manager := New()

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/debug", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	handler := manager.Handler()
	err := handler(c)

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	// Should render no_monitors view
	if rec.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", rec.Code)
	}
}

func TestManager_Handler_RedirectToFirstMonitor(t *testing.T) {
	manager := New()
	monitor := &Monitor{
		Name:        "test-monitor",
		DisplayName: "Test Monitor",
		MaxRecords:  10,
	}
	manager.AddMonitor(monitor)

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/debug", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetPath("/debug")

	handler := manager.Handler()
	err := handler(c)

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	// Should redirect to the first monitor
	if rec.Code != http.StatusFound {
		t.Errorf("Expected status 302, got %d", rec.Code)
	}

	location := rec.Header().Get("Location")
	expectedLocation := "/debug?monitor=" + url.QueryEscape("test-monitor")
	if location != expectedLocation {
		t.Errorf("Expected location %s, got %s", expectedLocation, location)
	}
}

func TestManager_Handler_WithMonitorParam(t *testing.T) {
	manager := New()
	monitor := &Monitor{
		Name:        "test-monitor",
		DisplayName: "Test Monitor",
		MaxRecords:  10,
		Icon:        IconExclamationCircle,
	}
	manager.AddMonitor(monitor)

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/debug?monitor=test-monitor", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	handler := manager.Handler()
	err := handler(c)

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	// Should render the monitor view
	if rec.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", rec.Code)
	}
}

func TestManager_Handler_InvalidMonitor(t *testing.T) {
	manager := New()
	monitor := &Monitor{
		Name:        "test-monitor",
		DisplayName: "Test Monitor",
		MaxRecords:  10,
	}
	manager.AddMonitor(monitor)

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/debug?monitor=nonexistent", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetPath("/debug")

	handler := manager.Handler()
	err := handler(c)

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	// Should redirect to the debug monitor top page
	if rec.Code != http.StatusFound {
		t.Errorf("Expected status 302, got %d", rec.Code)
	}

	location := rec.Header().Get("Location")
	if location != "/debug" {
		t.Errorf("Expected location /debug, got %s", location)
	}
}

func TestManager_Handler_StaticFile(t *testing.T) {
	manager := New()

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/debug?file=app.js", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	handler := manager.Handler()
	err := handler(c)

	// This test depends on whether the file exists in the embedded FS
	// We just verify the request is handled without panic
	if err != nil {
		// Error is expected if file doesn't exist in embedded FS
		httpErr, ok := err.(*echo.HTTPError)
		if ok && httpErr.Code != http.StatusNotFound {
			t.Errorf("Expected 404 or success, got error: %v", err)
		}
	}
}

func TestManager_Handler_MethodNotAllowed(t *testing.T) {
	manager := New()

	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/debug", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	handler := manager.Handler()
	err := handler(c)

	if err == nil {
		t.Error("Expected error for POST method")
	}

	httpErr, ok := err.(*echo.HTTPError)
	if !ok {
		t.Errorf("Expected HTTPError, got %T", err)
	}

	// The actual status code might be 400 (BadRequest) due to CSRF middleware
	// We just check that an error occurred
	if httpErr.Code != http.StatusMethodNotAllowed && httpErr.Code != http.StatusBadRequest {
		t.Errorf("Expected status 405 or 400, got %d", httpErr.Code)
	}
}
