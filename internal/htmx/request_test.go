package htmx

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
)

func TestIsHxRequest_True(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set(HeaderHXRequest, "true")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	if !IsHxRequest(c) {
		t.Error("Expected IsHxRequest to return true")
	}
}

func TestIsHxRequest_False(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	if IsHxRequest(c) {
		t.Error("Expected IsHxRequest to return false")
	}
}

func TestIsHxRequest_OtherValue(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set(HeaderHXRequest, "false")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	if IsHxRequest(c) {
		t.Error("Expected IsHxRequest to return false when header is not 'true'")
	}
}

func TestIsHxBoosted_True(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set(HeaderHXBoosted, "true")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	if !IsHxBoosted(c) {
		t.Error("Expected IsHxBoosted to return true")
	}
}

func TestIsHxBoosted_False(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	if IsHxBoosted(c) {
		t.Error("Expected IsHxBoosted to return false")
	}
}

func TestCurrentURL(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	expectedURL := "https://example.com/page"
	req.Header.Set(HeaderHXCurrentURL, expectedURL)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	currentURL := CurrentURL(c)

	if currentURL != expectedURL {
		t.Errorf("Expected CurrentURL to return %q, got %q", expectedURL, currentURL)
	}
}

func TestCurrentURL_Empty(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	currentURL := CurrentURL(c)

	if currentURL != "" {
		t.Errorf("Expected CurrentURL to return empty string, got %q", currentURL)
	}
}

func TestIsHxHistoryRestoreRequest_True(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set(HeaderHXHistoryRestoreRequest, "true")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	if !IsHxHistoryRestoreRequest(c) {
		t.Error("Expected IsHxHistoryRestoreRequest to return true")
	}
}

func TestIsHxHistoryRestoreRequest_False(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	if IsHxHistoryRestoreRequest(c) {
		t.Error("Expected IsHxHistoryRestoreRequest to return false")
	}
}

func TestMultipleHtmxHeaders(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set(HeaderHXRequest, "true")
	req.Header.Set(HeaderHXBoosted, "true")
	req.Header.Set(HeaderHXCurrentURL, "https://example.com/test")
	req.Header.Set(HeaderHXHistoryRestoreRequest, "true")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	if !IsHxRequest(c) {
		t.Error("Expected IsHxRequest to return true")
	}

	if !IsHxBoosted(c) {
		t.Error("Expected IsHxBoosted to return true")
	}

	if CurrentURL(c) != "https://example.com/test" {
		t.Errorf("Expected CurrentURL to return 'https://example.com/test', got %q", CurrentURL(c))
	}

	if !IsHxHistoryRestoreRequest(c) {
		t.Error("Expected IsHxHistoryRestoreRequest to return true")
	}
}
