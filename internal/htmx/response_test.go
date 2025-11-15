package htmx

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
)

func TestRedirect_HtmxRequest(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set(HeaderHXRequest, "true")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	redirectURL := "/new-page"
	err := Redirect(c, http.StatusFound, redirectURL)

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	// For HTMX requests, should return 200 OK with HX-Location header
	if rec.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", rec.Code)
	}

	location := rec.Header().Get(HeaderHXLocation)
	if location != redirectURL {
		t.Errorf("Expected HX-Location header %q, got %q", redirectURL, location)
	}

	// Should not have standard Location header
	standardLocation := rec.Header().Get("Location")
	if standardLocation != "" {
		t.Errorf("Expected no standard Location header, got %q", standardLocation)
	}
}

func TestRedirect_NonHtmxRequest(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	redirectURL := "/new-page"
	err := Redirect(c, http.StatusFound, redirectURL)

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	// For non-HTMX requests, should use standard redirect with 302
	if rec.Code != http.StatusFound {
		t.Errorf("Expected status 302, got %d", rec.Code)
	}

	location := rec.Header().Get("Location")
	if location != redirectURL {
		t.Errorf("Expected Location header %q, got %q", redirectURL, location)
	}

	// Should not have HX-Location header
	hxLocation := rec.Header().Get(HeaderHXLocation)
	if hxLocation != "" {
		t.Errorf("Expected no HX-Location header, got %q", hxLocation)
	}
}

func TestRedirect_DifferentStatusCodes(t *testing.T) {
	testCases := []struct {
		name       string
		statusCode int
	}{
		{"Moved Permanently", http.StatusMovedPermanently},
		{"Found", http.StatusFound},
		{"See Other", http.StatusSeeOther},
		{"Temporary Redirect", http.StatusTemporaryRedirect},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			e := echo.New()
			req := httptest.NewRequest(http.MethodGet, "/", nil)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			err := Redirect(c, tc.statusCode, "/redirect")

			if err != nil {
				t.Errorf("Expected no error, got %v", err)
			}

			if rec.Code != tc.statusCode {
				t.Errorf("Expected status %d, got %d", tc.statusCode, rec.Code)
			}
		})
	}
}

func TestReloadRedirect_HtmxRequest(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set(HeaderHXRequest, "true")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	redirectURL := "/reload-page"
	err := ReloadRedirect(c, http.StatusFound, redirectURL)

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	// For HTMX requests, should return 200 OK with HX-Redirect header
	if rec.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", rec.Code)
	}

	redirect := rec.Header().Get(HeaderHXRedirect)
	if redirect != redirectURL {
		t.Errorf("Expected HX-Redirect header %q, got %q", redirectURL, redirect)
	}

	// Should not have standard Location header
	standardLocation := rec.Header().Get("Location")
	if standardLocation != "" {
		t.Errorf("Expected no standard Location header, got %q", standardLocation)
	}

	// Should not have HX-Location header (ReloadRedirect uses HX-Redirect)
	hxLocation := rec.Header().Get(HeaderHXLocation)
	if hxLocation != "" {
		t.Errorf("Expected no HX-Location header, got %q", hxLocation)
	}
}

func TestReloadRedirect_NonHtmxRequest(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	redirectURL := "/reload-page"
	err := ReloadRedirect(c, http.StatusFound, redirectURL)

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	// For non-HTMX requests, should use standard redirect with 302
	if rec.Code != http.StatusFound {
		t.Errorf("Expected status 302, got %d", rec.Code)
	}

	location := rec.Header().Get("Location")
	if location != redirectURL {
		t.Errorf("Expected Location header %q, got %q", redirectURL, location)
	}

	// Should not have HX-Redirect header
	hxRedirect := rec.Header().Get(HeaderHXRedirect)
	if hxRedirect != "" {
		t.Errorf("Expected no HX-Redirect header, got %q", hxRedirect)
	}
}

func TestRedirect_Vs_ReloadRedirect(t *testing.T) {
	// Test that Redirect and ReloadRedirect use different headers for HTMX requests
	e := echo.New()

	// Test Redirect
	req1 := httptest.NewRequest(http.MethodGet, "/", nil)
	req1.Header.Set(HeaderHXRequest, "true")
	rec1 := httptest.NewRecorder()
	c1 := e.NewContext(req1, rec1)

	err1 := Redirect(c1, http.StatusFound, "/page1")
	if err1 != nil {
		t.Errorf("Redirect: Expected no error, got %v", err1)
	}

	// Test ReloadRedirect
	req2 := httptest.NewRequest(http.MethodGet, "/", nil)
	req2.Header.Set(HeaderHXRequest, "true")
	rec2 := httptest.NewRecorder()
	c2 := e.NewContext(req2, rec2)

	err2 := ReloadRedirect(c2, http.StatusFound, "/page2")
	if err2 != nil {
		t.Errorf("ReloadRedirect: Expected no error, got %v", err2)
	}

	// Verify different headers
	hxLocation := rec1.Header().Get(HeaderHXLocation)
	hxRedirect := rec2.Header().Get(HeaderHXRedirect)

	if hxLocation == "" {
		t.Error("Redirect should set HX-Location header")
	}

	if hxRedirect == "" {
		t.Error("ReloadRedirect should set HX-Redirect header")
	}

	if rec1.Header().Get(HeaderHXRedirect) != "" {
		t.Error("Redirect should not set HX-Redirect header")
	}

	if rec2.Header().Get(HeaderHXLocation) != "" {
		t.Error("ReloadRedirect should not set HX-Location header")
	}
}

func TestRedirect_ComplexURLs(t *testing.T) {
	testURLs := []string{
		"/simple",
		"/path/with/multiple/segments",
		"/path?query=value&other=param",
		"/path#fragment",
		"https://example.com/absolute",
		"https://example.com/path?query=value#fragment",
	}

	for _, testURL := range testURLs {
		t.Run(testURL, func(t *testing.T) {
			e := echo.New()
			req := httptest.NewRequest(http.MethodGet, "/", nil)
			req.Header.Set(HeaderHXRequest, "true")
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			err := Redirect(c, http.StatusFound, testURL)

			if err != nil {
				t.Errorf("Expected no error, got %v", err)
			}

			location := rec.Header().Get(HeaderHXLocation)
			if location != testURL {
				t.Errorf("Expected HX-Location %q, got %q", testURL, location)
			}
		})
	}
}
