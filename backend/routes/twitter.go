package routes

import (
	"backend/handlers"
	"github.com/labstack/echo/v4"
)

func RegisterTwitterRoutes(e *echo.Echo, twitterHandler *handlers.TwitterHandler) {
	twitterGroup := e.Group("/twitter")
	twitterGroup.GET("/login", twitterHandler.Login)
    twitterGroup.GET("/callback", twitterHandler.Callback)
    twitterGroup.POST("/post", twitterHandler.PostTweet)
}