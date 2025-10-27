package handlers

import (
	"backend/services"
	"encoding/json"
	_ "github.com/labstack/echo-contrib/session"
	"github.com/labstack/echo/v4"
	_ "io"
	_ "log"
	_ "mime/multipart"
	"net/http"
)

type PlatformHandler struct {
	twitterService services.TwitterService
	userService    services.UserService
}

func NewPlatformHandler(twitterService services.TwitterService, userService services.UserService) *PlatformHandler {
	return &PlatformHandler{
		twitterService: twitterService,
		userService:    userService,
	}
}

func (h *PlatformHandler) PostToPlatform(c echo.Context) error {
	// This endpoint MUST be protected by JWTMiddleware.
	ok, email, err := h.userService.IsLoggedIn(c)
	if err != nil {
		return err
	}
	if !ok {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "User not logged in"})
	}

	platform := c.FormValue("platform")
	if platform == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Platform not specified"})
	}

	platformDataJSON := c.FormValue("platformData")

	// Get the uploaded files
	form, err := c.MultipartForm()
	if err != nil {
		return err
	}
	files := form.File["media"]

	switch platform {
	case "twitter":
		accessToken, accessSecret, err := h.userService.GetTwitterToken(email)
		if err != nil || accessToken == "" || accessSecret == "" {
			return c.JSON(http.StatusUnauthorized, map[string]string{"error": "Twitter account not linked or tokens are missing"})
		}
		// 1. Unmarshal the JSON into the specific Twitter data struct
		var twitterData struct {
			Content string `json:"content"`
			// maybe more fields
		}
		if err := json.Unmarshal([]byte(platformDataJSON), &twitterData); err != nil {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid format for platformData"})
		}
		if twitterData.Content == "" && len(files) == 0 {
		    return c.JSON(http.StatusBadRequest, map[string]string{"error": "A tweet must have either text content or media."})
		}

		err = h.twitterService.PostTweet(accessToken, accessSecret, twitterData.Content, files)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to post tweet: " + err.Error()})
		}
		return c.JSON(http.StatusOK, map[string]string{"message": "Tweet scheduled successfully!"})

	// OTHER PLATFORMS COMING SOON HEHEHEHEE

	default:
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Unsupported platform"})
	}
}
