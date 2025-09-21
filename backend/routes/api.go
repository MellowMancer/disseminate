package routes

import (
    "github.com/labstack/echo/v4"
    "backend/handlers"
    "backend/middleware"
)

func RegisterAPIRoutes(e *echo.Echo) {
    apiGroup := e.Group("/api", middleware.AuthMiddleware)

    apiGroup.GET("/protected", handlers.Protected)
}
