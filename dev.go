package debugmonitor

// dev is a development mode flag that can be set via ldflags at build time
// Can be set with -ldflags="-X github.com/kohkimakimoto/echo-debugmonitor.dev=true"
var dev string

// devPublicDir is the path to the public directory in development mode
// Can be set with -ldflags="-X github.com/kohkimakimoto/echo-debugmonitor.devPublicDir=../public"
var devPublicDir string

// devViewsDir is the path to the views directory in development mode
// Can be set with -ldflags="-X github.com/kohkimakimoto/echo-debugmonitor.devViewsDir=../views"
var devViewsDir string

// isDev returns whether development mode is enabled
func isDev() bool {
	return dev == "true"
}

// getDevPublicDir returns the public directory path for development mode
// Returns the default path if not set
func getDevPublicDir() string {
	if devPublicDir == "" {
		return "public"
	}
	return devPublicDir
}

// getDevViewsDir returns the views directory path for development mode
// Returns the default path if not set
func getDevViewsDir() string {
	if devViewsDir == "" {
		return "views"
	}
	return devViewsDir
}
