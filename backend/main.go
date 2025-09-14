package main

import (
    "backend/routes"
    "github.com/labstack/echo/v4"
)

func main() {
    e := echo.New()

    routes.RegisterRoutes(e)

    e.Logger.Fatal(e.Start(":8080"))
}
