package routes

import (
	"backend/handlers"
	"backend/middlewares"
	"github.com/labstack/echo/v4"
)

func RegisterTwitterRoutes(api *echo.Group, h *handlers.TwitterHandler) {
	twitter := api.Group("/twitter")

	twitter.GET("/link/begin", h.BeginTwitterLink, middlewares.StrictAuthRateLimiter()) // GET /api/twitter/link/begin
}