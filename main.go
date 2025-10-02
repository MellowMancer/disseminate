package main

import (
	"embed"
	"io/fs"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"

	"backend/handlers"
	"backend/middlewares"
	"backend/routes"
	"backend/services"

	"github.com/dghubble/oauth1"
	"github.com/gorilla/sessions"
	"github.com/joho/godotenv"
	"github.com/labstack/echo-contrib/session"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

//go:embed all:frontend/dist
var embeddedFrontend embed.FS

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using environment variables")
	}

	// --- Configuration ---
	sessionSecret := os.Getenv("SESSION_SECRET")
	jwtSecret := []byte(os.Getenv("JWT_SECRET"))
	appEnv := os.Getenv("APP_ENV")

	// ... (rest of your configuration like twitterConfig, supabase, etc.)
	twitterConsumerKey := os.Getenv("TWITTER_CONSUMER_KEY")
	twitterConsumerSecret := os.Getenv("TWITTER_CONSUMER_SECRET")
	twitterCallbackURL := os.Getenv("TWITTER_CALLBACK_URL")
	supabaseURL := os.Getenv("SUPABASE_URL")
	supabaseKey := os.Getenv("SUPABASE_KEY")

	twitterConfig := &oauth1.Config{
		ConsumerKey:    twitterConsumerKey,
		ConsumerSecret: twitterConsumerSecret,
		CallbackURL:    twitterCallbackURL,
		Endpoint: oauth1.Endpoint{ /* ... endpoint URLs ... */ },
	}

	// --- Services and Handlers ---
	userService := services.NewUserService(supabaseURL, supabaseKey, jwtSecret)
	userHandler := handlers.NewHandler(userService)
	twitterService := services.NewTwitterService(twitterConfig)
	twitterHandler := handlers.NewTwitterHandler(twitterService, userService)

	e := echo.New()

	// --- Global Middlewares ---
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Use(session.Middleware(sessions.NewCookieStore([]byte(sessionSecret))))

	// --- API Routes (These are handled directly by Go in both dev and prod) ---
	authGroup := e.Group("/auth")
	routes.RegisterAuthRoutes(authGroup, userHandler)

	apiGroup := e.Group("/api")
	apiGroup.Use(middlewares.JWTMiddleware(jwtSecret, []string{}))
	routes.RegisterTwitterRoutes(apiGroup, twitterHandler)

	e.GET("/twitter/link/callback", twitterHandler.Callback)

	// --- Frontend Serving Logic (The Core Change) ---
	if appEnv == "production" {
		// --- PRODUCTION MODE ---
		// Serve the frontend from the embedded filesystem.
		log.Println("Running in PRODUCTION mode")
		staticFilesFS, err := fs.Sub(embeddedFrontend, "frontend/dist")
		if err != nil {
			log.Fatal("Failed to create sub-filesystem for embedded assets:", err)
		}
		e.Use(middleware.StaticWithConfig(middleware.StaticConfig{
			Skipper: func(c echo.Context) bool {
				path := c.Request().URL.Path
				// Skip static file serving for API routes
				return strings.HasPrefix(path, "/api") || strings.HasPrefix(path, "/auth") || strings.HasPrefix(path, "/twitter/link/callback")
			},
			Filesystem: http.FS(staticFilesFS),
			HTML5:      true, // Crucial for SPAs
		}))
	} else {
		// --- DEVELOPMENT MODE ---
		// Reverse proxy all non-API requests to the Vite dev server.
		log.Println("Running in DEVELOPMENT mode")
		viteServerURL, err := url.Parse("http://localhost:5173")
		if err != nil {
			log.Fatal("Invalid Vite server URL:", err)
		}
		e.Use(middleware.ProxyWithConfig(middleware.ProxyConfig{
			Skipper: func(c echo.Context) bool {
				path := c.Request().URL.Path
				// Skip proxying for API routes, let them be handled by Echo
				return strings.HasPrefix(path, "/api") || strings.HasPrefix(path, "/auth") || strings.HasPrefix(path, "/twitter/link/callback")
			},
			Balancer: middleware.NewRoundRobinBalancer([]*middleware.ProxyTarget{
				{
					URL: viteServerURL,
				},
			}),
		}))
	}

	log.Println("Starting server on :8080")
	e.Logger.Fatal(e.Start(":8080"))
}