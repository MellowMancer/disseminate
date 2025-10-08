package routes

import (
	"backend/handlers"
	"github.com/labstack/echo/v4"
)

func RegisterInstagramRoutes(api *echo.Group, h *handlers.InstagramHandler) {
	twitter := api.Group("/instagram")
	
	twitter.GET("/link/begin", h.BeginInstagramLink) // GET /api/instagram/link/begin
	
	twitter.GET("/check", h.CheckInstagramToken) // GET /api/instagram/check

	
}