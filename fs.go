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
		// デバッグモードの場合は、ファイルシステムから直接読み込む
		publicFS = os.DirFS("public")
		viewsFS = os.DirFS("views")
	} else {
		// 本番モードの場合は、埋め込まれたファイルシステムを使用
		publicFS = echo.MustSubFS(ePublicFS, "public")
		viewsFS = echo.MustSubFS(eViewsFS, "views")
	}
}
