package routes

import (
	"backend/handlers"
	"github.com/labstack/echo/v4"
)

func RegisterTwitterRoutes(api *echo.Group, h *handlers.TwitterHandler) {
	twitter := api.Group("/twitter")
	
	twitter.GET("/link/begin", h.BeginTwitterLink) // GET /api/twitter/link/begin
	
	twitter.GET("/check", h.CheckTwitterToken) // GET /api/twitter/check
}