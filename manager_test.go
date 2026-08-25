package debugmonitor

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/labstack/echo/v4"
)

func TestManagerHandlerWithoutMonitors(t *testing.T) {
	e := echo.New()
	e.GET("/monitor", New().Handler())

	recorder := httptest.NewRecorder()
	e.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/monitor", nil))

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, recorder.Code)
	}
	if !strings.Contains(recorder.Body.String(), "no monitors have been installed") {
		t.Fatalf("expected no-monitors view, got %q", recorder.Body.String())
	}
}

func TestManagerHandlerRedirectsToFirstMonitor(t *testing.T) {
	e := echo.New()
	manager := New()
	manager.AddMonitor(&Monitor{Name: "requests", DisplayName: "Requests"})
	e.GET("/monitor", manager.Handler())

	recorder := httptest.NewRecorder()
	e.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/monitor", nil))

	if recorder.Code != http.StatusFound {
		t.Fatalf("expected status %d, got %d", http.StatusFound, recorder.Code)
	}
	if location := recorder.Header().Get("Location"); location != "/monitor?monitor=requests" {
		t.Fatalf("expected monitor redirect, got %q", location)
	}
}

func TestManagerHandlerServesAsset(t *testing.T) {
	e := echo.New()
	e.GET("/monitor", New().Handler())

	recorder := httptest.NewRecorder()
	e.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/monitor?file=app.js", nil))

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, recorder.Code)
	}
	if contentType := recorder.Header().Get("Content-Type"); contentType != "application/javascript" {
		t.Fatalf("expected JavaScript content type, got %q", contentType)
	}
	if recorder.Body.Len() == 0 {
		t.Fatal("expected a non-empty asset response")
	}
}

func TestManagerHandlerRejectsUnsupportedMethod(t *testing.T) {
	e := echo.New()
	e.Any("/monitor", New().Handler())

	recorder := httptest.NewRecorder()
	e.ServeHTTP(recorder, httptest.NewRequest(http.MethodPost, "/monitor", nil))

	if recorder.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected status %d, got %d", http.StatusMethodNotAllowed, recorder.Code)
	}
}
