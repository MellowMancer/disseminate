package main

import (
	"log"
	"os"

	"backend/handlers"
	"backend/middlewares"
	"backend/routes"
	"backend/services"

	"github.com/joho/godotenv"
	"github.com/labstack/echo/v4"
	"github.com/dghubble/oauth1"
	"github.com/labstack/echo-contrib/session"
    "github.com/gorilla/sessions"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	sessionSecret := os.Getenv("SESSION_SECRET")
	twitterConsumerKey := os.Getenv("TWITTER_CONSUMER_KEY")
	twitterConsumerSecret := os.Getenv("TWITTER_CONSUMER_SECRET")
	twitterCallbackURL := os.Getenv("TWITTER_CALLBACK_URL")
	supabaseURL := os.Getenv("SUPABASE_URL")
	supabaseKey := os.Getenv("SUPABASE_KEY")
	jwtSecret := []byte(os.Getenv("JWT_SECRET"))

	AuthorizeEndpoint := oauth1.Endpoint{
		RequestTokenURL: "https://api.twitter.com/oauth/request_token",
		AuthorizeURL:    "https://api.twitter.com/oauth/authorize",
		AccessTokenURL:  "https://api.twitter.com/oauth/access_token",
	}

	twitterConfig := &oauth1.Config{
		ConsumerKey:    twitterConsumerKey,
		ConsumerSecret: twitterConsumerSecret,
		CallbackURL:    twitterCallbackURL,
		Endpoint:       AuthorizeEndpoint,
	}


	userService := services.NewUserService(supabaseURL, supabaseKey, jwtSecret)
	userHandler := handlers.NewHandler(userService)
	twitterService := services.NewTwitterService(twitterConfig)
	twitterHandler := handlers.NewTwitterHandler(twitterService, userService)

	e := echo.New()
	e.Use(session.Middleware(sessions.NewCookieStore([]byte(sessionSecret))))

	auth := e.Group("/auth")
	routes.RegisterAuthRoutes(auth, userHandler)

	api := e.Group("/api")
	api.Use(middlewares.JWTMiddleware(jwtSecret, []string{}))
	routes.RegisterTwitterRoutes(api, twitterHandler)

	// The Twitter callback is a public page route.
	// It's initiated by a protected action, but the callback itself doesn't have a JWT.
	e.GET("/twitter/link/callback", twitterHandler.Callback)

	// e.GET("/login", serveApp, middlewares.RedirectIfAuthenticated(jwtSecret))
	// e.GET("/signup", serveApp, middlewares.RedirectIfAuthenticated(jwtSecret))

	protectedPages := e.Group("")
	excludedPaths := []string{"/login", "/signup", "/auth", "/api", "/assets", "/twitter/link/callback", "/favicon.ico", "/vite.svg"}
	protectedPages.Use(middlewares.JWTMiddleware(jwtSecret, excludedPaths))
	routes.RegisterPageRoutes(protectedPages)

	e.Static("/assets", "../frontend/dist/assets")
	e.Static("/favicon.ico", "../frontend/dist/favicon.ico")
	e.Static("/vite.svg", "../frontend/dist/vite.svg")

	e.File("/", "../frontend/dist/index.html")

	e.Logger.Fatal(e.Start(":8080"))
}

func serveApp(c echo.Context) error {
	return c.File("../frontend/dist/index.html")
}