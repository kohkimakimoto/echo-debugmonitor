package debugmonitor

import (
	"io"
	"net/http"

	viewkit "github.com/kohkimakimoto/echo-viewkit"
	"github.com/kohkimakimoto/echo-viewkit/pongo2"
	"github.com/labstack/echo/v4"
)

type DebugMonitor struct {
	watchers []*Watcher
}

func New() *DebugMonitor {
	return &DebugMonitor{}
}

func (m *DebugMonitor) AddWatcher(w *Watcher) {
	m.watchers = append(m.watchers, w)
}

func (m *DebugMonitor) Handler() echo.HandlerFunc {
	v := viewkit.New()

	v.FS = viewsFS
	v.Debug = isDev()
	v.AnonymousComponentsDirectories = []*pongo2.AnonymousComponentsDirectory{
		{Dir: "components"},
	}
	r := v.MustRenderer()

	return func(c echo.Context) error {
		// Check if a file query parameter is present
		file := c.QueryParam("file")
		if file != "" {
			// Serve the requested file from publicFS
			return serveStaticFile(c, file)
		}

		return viewkit.Render(r, c, http.StatusOK, "monitor", map[string]any{})
	}
}

// serveStaticFile serves static files (app.js or app.css) from publicFS
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

	// Open the file from publicFS
	f, err := publicFS.Open(filename)
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
