package routes

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

func RegisterPageRoutes(e *echo.Group) {
	e.GET("/*", func(c echo.Context) error {
		path := c.Request().URL.Path

		validFrontendRoutes := map[string]bool{
			"/":         true,
			"/login":    true,
			"/signup":   true,
			"/schedule": true,
			"/profile":  true,
		}

		if _, ok := validFrontendRoutes[path]; ok {
			return c.File("../frontend/dist/index.html")
		}

		return c.Redirect(http.StatusSeeOther, "/login")
	})
}
