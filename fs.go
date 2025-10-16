package debugmonitor

import (
	"embed"
	"io/fs"
	"os"

	"github.com/labstack/echo/v4"
)

var (
	//go:embed resources/public
	ePublicFS embed.FS

	//go:embed resources/views
	eViewsFS embed.FS

	publicFS fs.FS
	viewsFS  fs.FS
)

func init() {
	if isDev() {
		// In development mode, read directly from the file system
		publicFS = os.DirFS(getDevPublicDir())
		viewsFS = os.DirFS(getDevViewsDir())
	} else {
		// In production mode, use the embedded file system
		publicFS = echo.MustSubFS(ePublicFS, "public")
		viewsFS = echo.MustSubFS(eViewsFS, "views")
	}
}
