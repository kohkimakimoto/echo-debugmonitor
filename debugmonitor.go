package debugmonitor

import (
	"embed"
	viewkit "github.com/kohkimakimoto/echo-viewkit"
	"github.com/labstack/echo/v4"
	"net/http"
)

type DebugMonitor struct {
	watchers []*Watcher
	DBPath   string
}

func New() *DebugMonitor {
	return &DebugMonitor{}
}

func (m *DebugMonitor) AddWatcher(w *Watcher) {
	m.watchers = append(m.watchers, w)
}

//go:embed views
var viewsFS embed.FS

func Handler(m *DebugMonitor) echo.HandlerFunc {
	v := viewkit.New()
	v.FS = viewsFS
	v.FSBaseDir = "views"

	r := v.MustRenderer()

	return func(c echo.Context) error {
		return viewkit.Render(r, c, http.StatusOK, "monitor", map[string]any{})
	}
}
