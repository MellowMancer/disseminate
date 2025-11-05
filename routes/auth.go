package routes

import (
	"backend/handlers"
	"backend/middlewares"

	"github.com/labstack/echo/v4"
)

func RegisterAuthRoutes(a *echo.Group, h *handlers.Handler, jwtSecret []byte) {
	a.POST("/login", h.Login, middlewares.RedirectIfAuthenticated(jwtSecret))
	a.POST("/signup", h.SignUp, middlewares.RedirectIfAuthenticated(jwtSecret))
	a.POST("/logout", h.Logout, middlewares.RedirectIfAuthenticated(jwtSecret))
	a.GET("/status", h.AuthStatus)
	a.GET("/oauth_status", h.OAuthStatus)
}
