package main

import (
	_ "database/sql"
	_ "fmt"
	"log"
	_ "net/http"
	"os"
	_ "time"

	"backend/handlers"
	"backend/middleware"
	"backend/routes"
	"backend/services"
	_ "github.com/dgrijalva/jwt-go"
	_ "github.com/jinzhu/gorm/dialects/postgres"
	"github.com/joho/godotenv"
	"github.com/labstack/echo/v4"
	_ "github.com/labstack/echo/v4/middleware"
)

func main() {
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

	authGroup := e.Group("/")
	authGroup.Use(middleware.RedirectIfAuthenticated(jwtSecret))

	e.GET("/login", func(c echo.Context) error {
		return c.File("../frontend/dist/index.html")
	}, middleware.RedirectIfAuthenticated(jwtSecret))

	e.GET("/signup", func(c echo.Context) error {
		return c.File("../frontend/dist/index.html")
	}, middleware.RedirectIfAuthenticated(jwtSecret))

	protectedGroup := e.Group("")
	protectedGroup.Use(middleware.JWTMiddleware(jwtSecret))

	routes.RegisterAuthRoutes(e, userService)
	routes.RegisterPageRoutes(protectedGroup, userHandler)

	e.Static("/assets", "../frontend/dist/assets")
	e.Static("/favicon.ico", "../frontend/dist/favicon.ico")
	e.Static("/vite.svg", "../frontend/dist/vite.svg")

	e.File("/", "../frontend/dist/index.html")

	e.Logger.Fatal(e.Start(":8080"))
}

// package main

// import (
// 	"fmt"
// 	"io"
// 	"log"
// 	"net/http"
// 	"os"
// 	"github.com/joho/godotenv"
// )

// func main() {
// 	err := godotenv.Load()
//     if err != nil {
//         log.Fatal("Error loading .env file")
//     }

// 	supabaseURL := os.Getenv("SUPABASE_URL")   // e.g. https://xyzcompany.supabase.co/
// 	supabaseKey := os.Getenv("SUPABASE_KEY")   // service role secret key

// 	if supabaseURL == "" || supabaseKey == "" {
// 		log.Fatal("SUPABASE_URL and SUPABASE_KEY environment variables must be set")
// 	}

// 	url := supabaseURL + "/rest/v1/profiles?email=eq.wyatharth@gmail.com"
// 	req, err := http.NewRequest("GET", url, nil)
// 	if err != nil {
// 		log.Fatalf("Failed to create request: %v", err)
// 	}

// 	req.Header.Set("apikey", supabaseKey)
// 	req.Header.Set("Authorization", "Bearer "+supabaseKey)

// 	client := &http.Client{}
// 	resp, err := client.Do(req)
// 	if err != nil {
// 		log.Fatalf("Request error: %v", err)
// 	}
// 	defer resp.Body.Close()

// 	bodyBytes, _ := io.ReadAll(resp.Body)

// 	if resp.StatusCode != http.StatusOK {
// 		log.Fatalf("Ping failed. Status: %d, Response: %s", resp.StatusCode, string(bodyBytes))
// 	}

// 	fmt.Println("Ping successful. Response:", string(bodyBytes))
// }
