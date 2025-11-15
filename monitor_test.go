package debugmonitor

import (
	"net/http"
	"net/http/httptest"
	"testing"

	viewkit "github.com/kohkimakimoto/echo-viewkit"
	"github.com/kohkimakimoto/echo-viewkit/pongo2"
	"github.com/labstack/echo/v4"
)

func TestMonitor_Add_WithStore(t *testing.T) {
	monitor := &Monitor{
		Name:        "test",
		DisplayName: "Test Monitor",
		MaxRecords:  10,
	}

	// Initialize store (simulating AddMonitor)
	monitor.store = NewStore(monitor.MaxRecords)

	// Add some data
	monitor.Add(map[string]any{"message": "test1"})
	monitor.Add(map[string]any{"message": "test2"})

	if monitor.store.Len() != 2 {
		t.Errorf("Expected 2 records in store, got %d", monitor.store.Len())
	}

	// Verify data is actually stored
	data := monitor.store.GetLatest(2)
	if len(data) != 2 {
		t.Errorf("Expected 2 records, got %d", len(data))
	}
}

func TestMonitor_Add_WithoutStore(t *testing.T) {
	monitor := &Monitor{
		Name:        "test",
		DisplayName: "Test Monitor",
		MaxRecords:  10,
	}

	// Store is not initialized (monitor not connected to Manager)
	// This should not panic
	monitor.Add(map[string]any{"message": "test"})

	// No error expected, it's a noop
}

func TestMonitor_Add_MaxRecordsLimit(t *testing.T) {
	monitor := &Monitor{
		Name:        "test",
		DisplayName: "Test Monitor",
		MaxRecords:  3,
	}

	monitor.store = NewStore(monitor.MaxRecords)

	// Add more records than the limit
	for i := 0; i < 5; i++ {
		monitor.Add(map[string]any{"index": i})
	}

	// Should only have 3 records (the limit)
	if monitor.store.Len() != 3 {
		t.Errorf("Expected 3 records in store, got %d", monitor.store.Len())
	}

	// Verify the last 3 records remain
	data := monitor.store.GetLatest(3)
	expectedIndexes := []int{4, 3, 2} // reverse order
	for i, entry := range data {
		payload := entry.Payload.(map[string]any)
		if payload["index"] != expectedIndexes[i] {
			t.Errorf("Expected index %d at position %d, got %v",
				expectedIndexes[i], i, payload["index"])
		}
	}
}

func TestMonitorViewContext_EchoContext(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	monitor := &Monitor{Name: "test", DisplayName: "Test", MaxRecords: 10}
	manager := New()
	manager.AddMonitor(monitor)

	// Create a simple viewkit with minimal config
	v := viewkit.New()
	v.BaseDir = "/tmp" // Set a base directory to avoid panic
	renderer := v.MustRenderer()

	viewContext := &MonitorViewContext{
		ctx:      c,
		monitor:  monitor,
		renderer: renderer,
	}

	if viewContext.EchoContext() != c {
		t.Error("Expected EchoContext to return the echo.Context")
	}
}

func TestMonitorViewContext_Monitor(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	monitor := &Monitor{Name: "test", DisplayName: "Test", MaxRecords: 10}
	manager := New()
	manager.AddMonitor(monitor)

	v := viewkit.New()
	v.BaseDir = "/tmp"
	renderer := v.MustRenderer()

	viewContext := &MonitorViewContext{
		ctx:      c,
		monitor:  monitor,
		renderer: renderer,
	}

	if viewContext.Monitor() != monitor {
		t.Error("Expected Monitor to return the monitor")
	}
}

func TestMonitorViewContext_Store(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	monitor := &Monitor{Name: "test", DisplayName: "Test", MaxRecords: 10}
	manager := New()
	manager.AddMonitor(monitor)

	v := viewkit.New()
	v.BaseDir = "/tmp"
	renderer := v.MustRenderer()

	viewContext := &MonitorViewContext{
		ctx:      c,
		monitor:  monitor,
		renderer: renderer,
	}

	if viewContext.Store() == nil {
		t.Error("Expected Store to return the monitor's store")
	}
}

func TestMonitorViewContext_Render(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	monitor := &Monitor{Name: "test", DisplayName: "Test", MaxRecords: 10}
	manager := New()
	manager.AddMonitor(monitor)

	v := viewkit.New()
	v.BaseDir = "/tmp"
	renderer := v.MustRenderer()

	viewContext := &MonitorViewContext{
		ctx:      c,
		monitor:  monitor,
		renderer: renderer,
	}

	// Test rendering a simple template
	err := viewContext.Render(http.StatusOK, "<p>Hello {{ name }}</p>", pongo2.Context{
		"name": "World",
	})

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if rec.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", rec.Code)
	}

	body := rec.Body.String()
	expected := "<p>Hello World</p>"
	if body != expected {
		t.Errorf("Expected body %q, got %q", expected, body)
	}
}

func TestMonitorViewContext_Render_InvalidTemplate(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	monitor := &Monitor{Name: "test", DisplayName: "Test", MaxRecords: 10}
	manager := New()
	manager.AddMonitor(monitor)

	v := viewkit.New()
	v.BaseDir = "/tmp"
	renderer := v.MustRenderer()

	viewContext := &MonitorViewContext{
		ctx:      c,
		monitor:  monitor,
		renderer: renderer,
	}

	// Test rendering an invalid template
	err := viewContext.Render(http.StatusOK, "{{ invalid syntax", pongo2.Context{})

	if err == nil {
		t.Error("Expected error for invalid template syntax")
	}
}

func TestMonitorViewContext_renderTemplateString(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	monitor := &Monitor{Name: "test", DisplayName: "Test", MaxRecords: 10}
	manager := New()
	manager.AddMonitor(monitor)

	v := viewkit.New()
	v.BaseDir = "/tmp"
	renderer := v.MustRenderer()

	viewContext := &MonitorViewContext{
		ctx:      c,
		monitor:  monitor,
		renderer: renderer,
	}

	// Test rendering template string
	result, err := viewContext.renderTemplateString("<h1>{{ title }}</h1>", pongo2.Context{
		"title": "Test Title",
	})

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	expected := "<h1>Test Title</h1>"
	if result != expected {
		t.Errorf("Expected result %q, got %q", expected, result)
	}
}

func TestMonitorViewContext_renderTemplateString_WithVariables(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	monitor := &Monitor{Name: "test", DisplayName: "Test", MaxRecords: 10}
	manager := New()
	manager.AddMonitor(monitor)

	v := viewkit.New()
	v.BaseDir = "/tmp"
	renderer := v.MustRenderer()

	viewContext := &MonitorViewContext{
		ctx:      c,
		monitor:  monitor,
		renderer: renderer,
	}

	// Test rendering with multiple variables
	result, err := viewContext.renderTemplateString(
		"<div>{{ name }} - {{ count }}</div>",
		pongo2.Context{
			"name":  "Item",
			"count": 42,
		},
	)

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	expected := "<div>Item - 42</div>"
	if result != expected {
		t.Errorf("Expected result %q, got %q", expected, result)
	}
}
