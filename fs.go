package debugmonitor

import (
	"embed"

	"github.com/labstack/echo/v4"
)

var (
	//go:embed resources/build
	_buildFS embed.FS
	assetsFS = echo.MustSubFS(_buildFS, "resources/build/assets")

	//go:embed resources/views
	_viewsFS embed.FS
	viewsFS  = echo.MustSubFS(_viewsFS, "resources/views")
)
