package main

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

func main() {
	e := echo.New()

	e.Static("/assets", "../frontend/dist/assets")
	e.Static("/favicon.ico", "../frontend/dist/favicon.ico")   
	e.Static("/vite.svg", "../frontend/dist/vite.svg") 

	e.File("/", "../frontend/dist/index.html")
	e.GET("/*", func(c echo.Context) error {
		return c.File("../frontend/dist/index.html")
	})

	e.GET("/test", func(c echo.Context) error {
		return c.JSON(http.StatusOK, map[string]string{"message": "API works"})
	})

	e.Logger.Fatal(e.Start(":8080"))
}
