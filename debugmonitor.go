package debugmonitor

import (
	viewkit "github.com/kohkimakimoto/echo-viewkit"
	"github.com/labstack/echo/v4"
	"net/http"
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

func Handler(m *DebugMonitor) echo.HandlerFunc {
	v := viewkit.New()
	r := v.MustRenderer()

	return func(c echo.Context) error {
		return viewkit.Render(r, c, http.StatusOK, "monitor", map[string]any{})
	}
}
