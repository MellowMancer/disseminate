package main

import (
	_ "database/sql"
	_ "fmt"
	_ "net/http"
	"os"
	_ "time"

	"backend/handlers"
	"backend/routes"
	"backend/services"
	_ "github.com/dgrijalva/jwt-go"
	_ "github.com/jinzhu/gorm/dialects/postgres"
	"github.com/labstack/echo/v4"
	_ "github.com/labstack/echo/v4/middleware"
)

func main() {
	supabaseURL := os.Getenv("SUPABASE_URL")
	supabaseKey := os.Getenv("SUPABASE_SECRET_KEY")
	jwtSecret := os.Getenv("SUPABASE_JWT_SECRET")

	userService := services.NewUserService(supabaseURL, supabaseKey, jwtSecret)
	userHandler := handlers.NewHandler(userService)

	e := echo.New()

	e.Static("/assets", "../frontend/dist/assets")
	e.Static("/favicon.ico", "../frontend/dist/favicon.ico")
	e.Static("/vite.svg", "../frontend/dist/vite.svg")

	// Register auth routes on /auth prefix
	authGroup := e.Group("/auth")
	routes.RegisterAuthRoutes(authGroup, userService)
	routes.RegisterAPIRoutes(e, userHandler)

	e.File("/", "../frontend/dist/index.html")
	e.GET("/*", func(c echo.Context) error {
		return c.File("../frontend/dist/index.html")
	})

	e.Logger.Fatal(e.Start(":8080"))
}
