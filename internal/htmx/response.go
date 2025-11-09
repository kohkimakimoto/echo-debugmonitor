package htmx

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

// Redirect performs a redirect to the specified URL.
// For htmx requests, it triggers a client-side redirect without a full page reload.
func Redirect(c echo.Context, code int, url string) error {
	if IsHxRequest(c) {
		// For htmx requests, client-side redirection is performed by adding HX-Location to the response header.
		// see https://htmx.org/headers/hx-location/
		c.Response().Header().Set(HeaderHXLocation, url)

		// Return a 200 OK status. Because HTMX cannot handle 3xx status codes with HX-Location header.
		// https://htmx.org/docs/#response-headers
		return c.NoContent(http.StatusOK)
	}

	// For non-htmx requests, perform a standard server-side redirect.
	return c.Redirect(code, url)
}

// ReloadRedirect performs a redirect to the specified URL and reloads the entire page on the client side.
func ReloadRedirect(c echo.Context, code int, url string) error {
	if IsHxRequest(c) {
		// For htmx requests, client-side redirection is performed by adding HX-Redirect to the response header.
		// https://htmx.org/headers/hx-redirect/
		c.Response().Header().Set(HeaderHXRedirect, url)

		// Return a 200 OK status. Because HTMX cannot handle 3xx status codes with HX-Redirect header.
		// https://htmx.org/docs/#response-headers
		return c.NoContent(http.StatusOK)
	}

	// For non-htmx requests, perform a standard server-side redirect.
	return c.Redirect(code, url)
}
