package main

import (
	"log"
	"os"

	"backend/handlers"
	"backend/middleware"
	"backend/routes"
	"backend/services"
	"github.com/joho/godotenv"
	"github.com/labstack/echo/v4"
)

func main() {
	frontend_path := "../frontend/dist/index.html"
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	supabaseURL := os.Getenv("SUPABASE_URL")
	supabaseKey := os.Getenv("SUPABASE_KEY")
	jwtSecret := []byte(os.Getenv("JWT_SECRET"))

	userService := services.NewUserService(supabaseURL, supabaseKey, jwtSecret)
	userHandler := handlers.NewHandler(userService)

	e := echo.New()

	authGroup := e.Group("/auth")
	authGroup.Use(middleware.RedirectIfAuthenticated(jwtSecret))

	e.GET("/login", func(c echo.Context) error {
		return c.File(frontend_path)
	}, middleware.RedirectIfAuthenticated(jwtSecret))

	e.GET("/signup", func(c echo.Context) error {
		return c.File(frontend_path)
	}, middleware.RedirectIfAuthenticated(jwtSecret))

	protectedGroup := e.Group("")
	protectedGroup.Use(middleware.JWTMiddleware(jwtSecret))

	routes.RegisterAuthRoutes(e, userService)
	routes.RegisterPageRoutes(protectedGroup, userHandler)

	e.Static("/assets", "../frontend/dist/assets")
	e.Static("/favicon.ico", "../frontend/dist/favicon.ico")
	e.Static("/vite.svg", "../frontend/dist/vite.svg")

	e.File("/", frontend_path)

	e.Logger.Fatal(e.Start(":8080"))
}