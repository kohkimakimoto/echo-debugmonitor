package htmx

import "github.com/labstack/echo/v4"

func IsHxRequest(c echo.Context) bool {
	return c.Request().Header.Get(HeaderHXRequest) == "true"
}

func IsHxBoosted(c echo.Context) bool {
	return c.Request().Header.Get(HeaderHXBoosted) == "true"
}

func CurrentURL(c echo.Context) string {
	return c.Request().Header.Get(HeaderHXCurrentURL)
}

func IsHxHistoryRestoreRequest(c echo.Context) bool {
	return c.Request().Header.Get(HeaderHXHistoryRestoreRequest) == "true"
}
