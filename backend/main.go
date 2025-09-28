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
	"github.com/dghubble/oauth1"
	"github.com/labstack/echo-contrib/session"
    "github.com/gorilla/sessions"
)

func main() {
	frontend_path := "../frontend/dist/index.html"
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	sessionSecret := os.Getenv("SESSION_SECRET")

	twitterConsumerKey := os.Getenv("TWITTER_CONSUMER_KEY")
	twitterConsumerSecret := os.Getenv("TWITTER_CONSUMER_SECRET")
	twitterCallbackURL := os.Getenv("TWITTER_CALLBACK_URL")

	AuthorizeEndpoint := oauth1.Endpoint{
		RequestTokenURL: "https://api.twitter.com/oauth/request_token",
		AuthorizeURL:    "https://api.twitter.com/oauth/authorize",
		AccessTokenURL:  "https://api.twitter.com/oauth/access_token",
	}

	twitterConfig := oauth1.Config{
		ConsumerKey:    twitterConsumerKey,
		ConsumerSecret: twitterConsumerSecret,
		CallbackURL:    twitterCallbackURL,
		Endpoint:       AuthorizeEndpoint,
	}

	supabaseURL := os.Getenv("SUPABASE_URL")
	supabaseKey := os.Getenv("SUPABASE_KEY")
	jwtSecret := []byte(os.Getenv("JWT_SECRET"))

	userService := services.NewUserService(supabaseURL, supabaseKey, jwtSecret)
	userHandler := handlers.NewHandler(userService)

	twitterService := services.NewTwitterService(&twitterConfig)
	twitterHandler := handlers.NewTwitterHandler(twitterService, userService)

	e := echo.New()

	e.Use(session.Middleware(sessions.NewCookieStore([]byte(sessionSecret))))

	authGroup := e.Group("/auth")
	authGroup.Use(middlewares.RedirectIfAuthenticated(jwtSecret))

	e.GET("/login", func(c echo.Context) error {
		return c.File(frontend_path)
	}, middlewares.RedirectIfAuthenticated(jwtSecret))

	e.GET("/signup", func(c echo.Context) error {
		return c.File(frontend_path)
	}, middlewares.RedirectIfAuthenticated(jwtSecret))

	protectedGroup := e.Group("")
	protectedGroup.Use(middlewares.JWTMiddleware(jwtSecret))

	routes.RegisterAuthRoutes(e, userHandler)
	routes.RegisterPageRoutes(protectedGroup)
	routes.RegisterTwitterRoutes(e, twitterHandler)

	e.Static("/assets", "../frontend/dist/assets")
	e.Static("/favicon.ico", "../frontend/dist/favicon.ico")
	e.Static("/vite.svg", "../frontend/dist/vite.svg")

	e.File("/", frontend_path)

	e.Logger.Fatal(e.Start(":8080"))
}
