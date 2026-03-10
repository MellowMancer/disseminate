package routes

import (
	"backend/handlers"
	"backend/middlewares"
	"github.com/labstack/echo/v4"
)

func RegisterInstagramRoutes(api *echo.Group, h *handlers.InstagramHandler) {
	instagram := api.Group("/instagram")

	instagram.GET("/link/begin", h.BeginInstagramLink, middlewares.StrictAuthRateLimiter()) // GET /api/instagram/link/begin


}