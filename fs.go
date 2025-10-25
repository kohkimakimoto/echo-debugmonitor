package debugmonitor

import (
	"embed"

	"github.com/labstack/echo/v4"
)

var (
	//go:embed resources/build/assets
	_assetsFS embed.FS
	assetsFS  = echo.MustSubFS(_assetsFS, "resources/build/assets")

	//go:embed resources/views
	eViewsFS embed.FS
	viewsFS  = echo.MustSubFS(eViewsFS, "resources/views")
)
