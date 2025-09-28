package routes

import (
	"backend/handlers"
	"github.com/labstack/echo/v4"
)

func RegisterAuthRoutes(e *echo.Echo, h *handlers.Handler) {
	authGroup := e.Group("/auth")

	authGroup.POST("/login", h.Login)
	authGroup.POST("/signup", h.SignUp)
	authGroup.POST("/logout", h.Logout)
    authGroup.GET("/status", h.AuthStatus)
}
