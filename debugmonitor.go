package debugmonitor

import (
	"net/http"

	viewkit "github.com/kohkimakimoto/echo-viewkit"
	"github.com/kohkimakimoto/echo-viewkit/pongo2"
	"github.com/labstack/echo/v4"
)

// debug is a debug flag that can be set via ldflags at build time
// Can be set with -ldflags="-X github.com/kohkimakimoto/echo-debugmonitor.debug=true"
var debug string

// isDebug returns whether debug mode is enabled
func isDebug() bool {
	return debug == "true"
}

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

	v.FS = viewsFS
	v.Debug = isDebug()
	v.AnonymousComponentsDirectories = []*pongo2.AnonymousComponentsDirectory{
		{Dir: "components"},
	}
	v.PreProcessors = []pongo2.PreProcessor{
		pongo2.MustNewRegexRemove(`(?i)(?s)<style[^>]*\bdata-extract\b[^>]*>.*?</style>`, `(?i)(?s)<script[^>]*\bdata-extract\b[^>]*>.*?</script>`),
	}

	// In debug mode, don't cache templates
	if isDebug() {
		return func(c echo.Context) error {
			r := v.MustRenderer()
			return viewkit.Render(r, c, http.StatusOK, "monitor", map[string]any{})
		}
	}

	// In production mode, create and cache the renderer only once
	r := v.MustRenderer()
	return func(c echo.Context) error {
		return viewkit.Render(r, c, http.StatusOK, "monitor", map[string]any{})
	}
}
