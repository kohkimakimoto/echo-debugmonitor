package debugmonitor

import (
	"io"
	"net/http"

	viewkit "github.com/kohkimakimoto/echo-viewkit"
	"github.com/kohkimakimoto/echo-viewkit/pongo2"
	"github.com/labstack/echo/v4"
)

type Manager struct {
	monitors []*Monitor
}

// New creates a new Echo Debug Monitor manager instance.
func New() *Manager {
	return &Manager{}
}

func (m *Manager) AddMonitor(w *Monitor) {
	m.monitors = append(m.monitors, w)
}

func (m *Manager) Handler() echo.HandlerFunc {
	v := viewkit.New()

	v.FS = viewsFS
	v.AnonymousComponentsDirectories = []*pongo2.AnonymousComponentsDirectory{
		{Dir: "components"},
	}
	r := v.MustRenderer()

	return func(c echo.Context) error {
		// Check if a file query parameter is present
		file := c.QueryParam("file")
		if file != "" {
			// Serve the requested file from assetsFS
			return serveStaticFile(c, file)
		}

		return viewkit.Render(r, c, http.StatusOK, "home", map[string]any{
			"manager": m,
		})
	}
}

// serveStaticFile serves static files (app.js or app.css) from assetsFS
func serveStaticFile(c echo.Context, filename string) error {
	var contentType string

	switch filename {
	case "app.js":
		contentType = "application/javascript"
	case "app.css":
		contentType = "text/css"
	default:
		return echo.NewHTTPError(http.StatusNotFound)
	}

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
