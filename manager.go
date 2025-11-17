package debugmonitor

import (
	"bytes"
	"html/template"
	"io"
	"net/http"
	"net/url"
	"sync"

	"github.com/labstack/echo/v4"
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
	t := template.Must(template.New("T").ParseFS(viewsFS, "*.html"))

	return func(c echo.Context) error {
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
					return renderView(t, c, http.StatusOK, "no_monitors.html", nil)
				}
			}

			monitor, ok := m.monitorMap[monitorName]
			if !ok {
				// monitor not found. Redirect to the Echo Debug monitor top page.
				return c.Redirect(http.StatusFound, c.Path())
			}

			action := c.QueryParam("action")
			if action != "" {
				if monitor.ActionHandler == nil {
					return c.JSON(http.StatusInternalServerError, map[string]any{
						"error": "Monitor " + monitor.Name + " does not have a ActionHandler implementation.",
					})
				}
				// handle monitor action
				return monitor.ActionHandler(c, monitor.store, action)
			}

			return renderView(t, c, http.StatusOK, "monitor.html", map[string]any{
				"Manager": m,
				"Monitor": monitor,
				"Title":   monitor.DisplayName + " - Echo Debug Monitor",
			})
		}

		return echo.NewHTTPError(http.StatusMethodNotAllowed)
	}
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

func renderView(t *template.Template, c echo.Context, code int, viewName string, data map[string]any) error {
	buf := new(bytes.Buffer)
	if err := t.ExecuteTemplate(buf, viewName, data); err != nil {
		return err
	}
	return c.HTMLBlob(code, buf.Bytes())
}
