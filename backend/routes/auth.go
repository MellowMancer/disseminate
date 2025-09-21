package routes

import (
    "github.com/labstack/echo/v4"
    "backend/handlers"
	"backend/services"
)

func RegisterAuthRoutes(e *echo.Group, userService services.UserService) {
	h := handlers.NewHandler(userService)

    e.POST("/login", h.Login)
    e.POST("/signup", h.SignUp)
	e.POST("/logout", h.Logout)
}
