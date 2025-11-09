package debugmonitor

import (
	"io"
	"net/http"
	"net/url"
	"sync"

	"github.com/kohkimakimoto/echo-debugmonitor/internal/htmx"
	viewkit "github.com/kohkimakimoto/echo-viewkit"
	"github.com/kohkimakimoto/echo-viewkit/pongo2"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

type Manager struct {
	monitors   []*Monitor
	monitorMap map[string]*Monitor
	mutex      sync.RWMutex
}

// New creates a new Echo Debug Monitor manager instance.
func New() *Manager {
	return &Manager{
		monitors:   []*Monitor{},
		monitorMap: make(map[string]*Monitor),
	}
}

func (m *Manager) AddMonitor(monitor *Monitor) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	// Initialize the store for this monitor
	// The store will manage ID generation internally
	monitor.store = NewStore(monitor.MaxRecords)

	m.monitorMap[monitor.Name] = monitor
	m.monitors = append(m.monitors, monitor)
}

func (m *Manager) Monitors() []*Monitor {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	return m.monitors
}

func (m *Manager) Handler() echo.HandlerFunc {
	v := viewkit.New()

	v.FS = viewsFS
	v.AnonymousComponentsDirectories = []*pongo2.AnonymousComponentsDirectory{
		{Dir: "components"},
	}
	v.SharedContextProviders = map[string]viewkit.SharedContextProviderFunc{
		"csrf_token": func(c echo.Context) (any, error) {
			return func() string {
				return getCSRFToken(c)
			}, nil
		},
		"manager": func(c echo.Context) (any, error) {
			return m, nil
		},
	}

	r := v.MustRenderer()

	h := func(c echo.Context) error {
		if c.Request().Method == http.MethodGet {
			// Check if a file query parameter is present
			file := c.QueryParam("file")
			if file != "" {
				// Serve the requested file from assetsFS
				return serveStaticFile(c, file)
			}

			monitorName := c.QueryParam("monitor")
			if monitorName == "" {
				if len(m.monitors) > 0 {
					monitor := m.monitors[0]
					return c.Redirect(http.StatusFound, c.Path()+"?monitor="+url.QueryEscape(monitor.Name))
				} else {
					return viewkit.Render(r, c, http.StatusOK, "no_monitors", nil)
				}
			}

			monitor, ok := m.monitorMap[monitorName]
			if !ok {
				// monitor not found. Redirect to the Echo Debug monitor top page.
				return c.Redirect(http.StatusFound, c.Path())
			}

			if htmx.IsHxRequest(c) && !htmx.IsHxBoosted(c) {
				if monitor.ViewHandler == nil {
					return c.JSON(http.StatusInternalServerError, map[string]any{
						"error": "Monitor " + monitor.Name + " does not have a ViewHandler defined.",
					})
				}

				// sub request for monitor content
				return monitor.ViewHandler(&MonitorViewContext{
					ctx:      c,
					monitor:  monitor,
					renderer: r,
				})
			}

			return viewkit.Render(r, c, http.StatusOK, "monitor", map[string]any{
				"monitor": monitor,
				"title":   monitor.DisplayName + " - Echo Debug Monitor",
			})
		}

		return echo.NewHTTPError(http.StatusMethodNotAllowed)
	}

	// add CSRF middleware
	h = middleware.CSRFWithConfig(middleware.CSRFConfig{
		TokenLookup: "form:_csrf_token,header:X-CSRF-Token",
		CookieName:  "echo-debugmonitor.csrf",
		ContextKey:  "echo-debugmonitor.csrf",
	})(h)

	return h
}

// serveStaticFile serves static files (app.js or app.css) from assetsFS
func serveStaticFile(c echo.Context, filename string) error {
	switch filename {
	case "app.js":
		return serveAsset(c, "app.js", "application/javascript")
	case "tailwindcss.js":
		return serveAsset(c, "tailwindcss.js", "application/javascript")
	default:
		return echo.NewHTTPError(http.StatusNotFound)
	}
}

// serveAsset is a helper function that serves a file with the specified content type
func serveAsset(c echo.Context, filename string, contentType string) error {
	// Open the file from assetsFS
	f, err := assetsFS.Open(filename)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound)
	}
	defer f.Close()

	// Set the content type header
	c.Response().Header().Set("Content-Type", contentType)

	// Copy the file contents to the response
	_, err = io.Copy(c.Response().Writer, f)
	return err
}

func getCSRFToken(c echo.Context) string {
	token, ok := c.Get("echo-debugmonitor.csrf").(string)
	if !ok {
		return ""
	}
	return token
}
