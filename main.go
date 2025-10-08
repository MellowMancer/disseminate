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
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/facebook"
)

//go:embed all:frontend/dist
var embeddedFrontend embed.FS

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using environment variables")
	}

	sessionSecret := os.Getenv("SESSION_SECRET")
	jwtSecret := []byte(os.Getenv("JWT_SECRET"))
	appEnv := os.Getenv("APP_ENV")

	twitterConsumerKey := os.Getenv("TWITTER_CONSUMER_KEY")
	twitterConsumerSecret := os.Getenv("TWITTER_CONSUMER_SECRET")
	twitterCallbackURL := os.Getenv("TWITTER_CALLBACK_URL")

	metaAppID := os.Getenv("META_APP_ID")
	metaAppSecret := os.Getenv("META_APP_SECRET")
	metaRedirectURL := os.Getenv("META_REDIRECT_URL")
	metaConfigurationID := os.Getenv("META_CONFIGURATION_ID")

	supabaseURL := os.Getenv("SUPABASE_URL")
	supabaseKey := os.Getenv("SUPABASE_KEY")

	twitterEndpoint := oauth1.Endpoint{
		RequestTokenURL: "https://api.twitter.com/oauth/request_token",
		AuthorizeURL:    "https://api.twitter.com/oauth/authorize",
		AccessTokenURL:  "https://api.twitter.com/oauth/access_token",
	}

	metaConfig := &oauth2.Config{
  		ClientID:     metaAppID,
  		ClientSecret: metaAppSecret,
  		RedirectURL:  metaRedirectURL,
  		Scopes:       []string{"email"},
		Endpoint:     facebook.Endpoint,
	}




	twitterConfig := &oauth1.Config{
		ConsumerKey:    twitterConsumerKey,
		ConsumerSecret: twitterConsumerSecret,
		CallbackURL:    twitterCallbackURL,
		Endpoint: twitterEndpoint,
	}

	const TWITTERCALLBACKPATH = "/twitter/link/callback"
	const INSTAGRAMCALLBACKPATH = "/instagram/link/callback"

	// --- Services and Handlers ---
	userService := services.NewUserService(supabaseURL, supabaseKey, jwtSecret)
	userHandler := handlers.NewHandler(userService)
	twitterService := services.NewTwitterService(twitterConfig)
	twitterHandler := handlers.NewTwitterHandler(twitterService, userService)
	platformHandler := handlers.NewPlatformHandler(twitterService, userService)
	instagramService := services.NewInstagramService(metaConfig, metaConfigurationID)
	instagramHandler := handlers.NewInstagramHandler(instagramService, userService)

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
	routes.RegisterPlatformRoute(apiGroup, platformHandler)
	routes.RegisterTwitterRoutes(apiGroup, twitterHandler)
	routes.RegisterInstagramRoutes(apiGroup, instagramHandler)

	e.GET(TWITTERCALLBACKPATH, twitterHandler.Callback)

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
				return strings.HasPrefix(path, "/api") || strings.HasPrefix(path, "/auth") || strings.HasPrefix(path, TWITTERCALLBACKPATH)
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
				return strings.HasPrefix(path, "/api") || strings.HasPrefix(path, "/auth") || strings.HasPrefix(path, TWITTERCALLBACKPATH)
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