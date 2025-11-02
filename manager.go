package debugmonitor

import (
	"io"
	"net/http"
	"net/url"
	"sync"

	"github.com/kohkimakimoto/echo-debugmonitor/pkg/htmx"
	viewkit "github.com/kohkimakimoto/echo-viewkit"
	"github.com/kohkimakimoto/echo-viewkit/pongo2"
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

func (m *Manager) AddMonitor(mo *Monitor) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	// Initialize the channel for this monitor
	bufferSize := mo.ChannelBufferSize
	if bufferSize <= 0 {
		bufferSize = 100 // Default buffer size
	}
	mo.dataChan = make(chan Data, bufferSize)

	m.monitorMap[mo.Name] = mo
	m.monitors = append(m.monitors, mo)

	// Start a goroutine to receive data from the monitor
	go m.receiveData(mo)
}

func (m *Manager) Monitors() []*Monitor {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	return m.monitors
}

// receiveData is a goroutine that receives data from a monitor's channel
func (m *Manager) receiveData(monitor *Monitor) {
	for data := range monitor.dataChan {
		// TODO: Store data in memory buffer
		// For now, just receive it to prevent channel blocking
		_ = data
	}
}

func (m *Manager) Handler() echo.HandlerFunc {
	v := viewkit.New()

	v.FS = viewsFS
	v.AnonymousComponentsDirectories = []*pongo2.AnonymousComponentsDirectory{
		{Dir: "components"},
	}
	v.SharedContextProviders = map[string]viewkit.SharedContextProviderFunc{
		"manager": func(c echo.Context) (any, error) {
			return m, nil
		},
		"monitors": func(c echo.Context) (any, error) {
			return newViewMonitorSlice(m.Monitors()), nil
		},
	}

	r := v.MustRenderer()

	return func(c echo.Context) error {
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
				encoded := url.QueryEscape(monitor.Name)
				return c.Redirect(http.StatusFound, c.Path()+"?monitor="+encoded)
			} else {
				return viewkit.Render(r, c, http.StatusOK, "no_monitors", nil)
			}
		}

		monitor, ok := m.monitorMap[monitorName]
		if !ok {
			// Monitor not found. Redirect to the Echo Debug Monitor top page.
			return htmx.ReloadRedirect(c, http.StatusFound, c.Path())
		}
		return viewkit.Render(r, c, http.StatusOK, "monitor", map[string]any{
			"monitor": newViewMonitor(monitor),
		})
	}
}

// serveStaticFile serves static files (app.js or app.css) from assetsFS
func serveStaticFile(c echo.Context, filename string) error {
	switch filename {
	case "app.js":
		return serveAsset(c, "app.js", "application/javascript")
	case "app.css":
		return serveAsset(c, "app.css", "text/css")
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
