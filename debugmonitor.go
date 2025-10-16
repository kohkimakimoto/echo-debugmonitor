package debugmonitor

import (
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

func Handler(m *DebugMonitor) echo.HandlerFunc {
	v := viewkit.New()

	v.FS = viewsFS
	v.FSBaseDir = "views"
	v.AnonymousComponentsDirectories = []*pongo2.AnonymousComponentsDirectory{
		{Dir: "components"},
	}
	v.PreProcessors = []pongo2.PreProcessor{
		pongo2.MustNewRegexRemove(`(?i)(?s)<style[^>]*\bdata-extract\b[^>]*>.*?</style>`, `(?i)(?s)<script[^>]*\bdata-extract\b[^>]*>.*?</script>`),
	}

	// デバッグモードの場合は、テンプレートをキャッシュしない
	if isDebug() {
		return func(c echo.Context) error {
			r := v.MustRenderer()
			return viewkit.Render(r, c, http.StatusOK, "monitor", map[string]any{})
		}
	}

	// 本番モードの場合は、レンダラーを一度だけ作成してキャッシュする
	r := v.MustRenderer()
	return func(c echo.Context) error {
		return viewkit.Render(r, c, http.StatusOK, "monitor", map[string]any{})
	}
}
