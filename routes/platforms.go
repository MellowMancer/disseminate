package routes

import (
	"backend/handlers"
	"github.com/labstack/echo/v4"
)

func RegisterPlatformRoute(api *echo.Group, p *handlers.PlatformHandler) {
	api.POST("/create", p.PostToPlatform)
}