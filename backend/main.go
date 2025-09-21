package main

import (
	_ "database/sql"
	_ "fmt"
	_ "net/http"
	_ "time"
	"os"

	"backend/routes"
	"backend/models"
	"backend/services"
	"github.com/labstack/echo/v4"
    _ "github.com/labstack/echo/v4/middleware"
    _ "github.com/dgrijalva/jwt-go"
    "github.com/jinzhu/gorm"
    _ "github.com/jinzhu/gorm/dialects/postgres"
)

var (
    db *gorm.DB
)

func init() {
    var err error
    dsn := "postgresql://user:password@localhost/database_name?sslmode=disable" // Update with your database credentials
    db, err = gorm.Open("postgres", dsn)
    if err != nil {
        panic("failed to connect to database")
    }

    // Auto migrate the User model
    db.AutoMigrate(&models.User{})
}

func main() {
	supabaseURL := os.Getenv("SUPABASE_URL")
    supabaseKey := os.Getenv("SUPABASE_KEY")

	userService := services.NewUserService(supabaseURL, supabaseKey)


	e := echo.New()

	e.Static("/assets", ".../frontend/dist/assets")
	e.Static("/favicon.ico", ".../frontend/dist/favicon.ico")
	e.Static("/vite.svg", ".../frontend/dist/vite.svg")

	e.File("/", ".../frontend/dist/index.html")

	// Register auth routes on /auth prefix
	authGroup := e.Group("/auth")
	routes.RegisterAuthRoutes(authGroup, userService)
	routes.RegisterAPIRoutes(e)

	e.Logger.Fatal(e.Start(":8080"))
}
