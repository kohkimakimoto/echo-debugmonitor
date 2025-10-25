package debugmonitor

import (
	"embed"

	"github.com/labstack/echo/v4"
)

var (
	//go:embed resources/public
	_publicFS embed.FS
	publicFS  = echo.MustSubFS(_publicFS, "resources/public")

	//go:embed resources/views
	eViewsFS embed.FS
	viewsFS  = echo.MustSubFS(eViewsFS, "resources/views")
)
