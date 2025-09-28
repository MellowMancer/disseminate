package routes

import (
	"backend/handlers"
	"github.com/labstack/echo/v4"
)

func RegisterAuthRoutes(a *echo.Group, h *handlers.Handler) {
	a.POST("/login", h.Login)
	a.POST("/signup", h.SignUp)
	a.POST("/logout", h.Logout)
    a.GET("/status", h.AuthStatus)
}
