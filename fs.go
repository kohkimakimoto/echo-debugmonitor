package debugmonitor

import (
	"embed"
	"io/fs"
	"os"

	"github.com/labstack/echo/v4"
)

var (
	//go:embed public
	ePublicFS embed.FS

	//go:embed views
	eViewsFS embed.FS

	publicFS fs.FS
	viewsFS  fs.FS
)

func init() {
	if isDebug() {
		// In debug mode, read directly from the file system
		publicFS = os.DirFS("../public")
		viewsFS = os.DirFS("../views")
	} else {
		// In production mode, use the embedded file system
		publicFS = echo.MustSubFS(ePublicFS, "public")
		viewsFS = echo.MustSubFS(eViewsFS, "views")
	}
}
