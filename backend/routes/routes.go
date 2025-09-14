package routes

import (
	"backend/handlers"

	"github.com/labstack/echo/v4"
)

func RegisterRoutes(e *echo.Echo) {
	e.GET("/", handlers.Hello)
}
