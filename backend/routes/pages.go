package routes

import (
    "github.com/labstack/echo/v4"
    "backend/handlers"
)

func RegisterPageRoutes(e *echo.Group, userHandler *handlers.Handler) {
    e.GET("/*", func(c echo.Context) error {
        return c.File("../frontend/dist/index.html")
    })
}
