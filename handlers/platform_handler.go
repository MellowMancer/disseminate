package handlers

import (
	service_twitter "backend/services/twitter"
	service_instagram "backend/services/instagram"
	service_user "backend/services/user"
	"encoding/json"
	_ "github.com/labstack/echo-contrib/session"
	"github.com/labstack/echo/v4"
	_ "io"
	_ "log"
	"mime/multipart"
	"net/http"
)

type PlatformHandler struct {
	twitterService   service_twitter.TwitterService
	instagramService service_instagram.InstagramService
	userService      service_user.UserService
}

func NewPlatformHandler(twitterService service_twitter.TwitterService, instagramService service_instagram.InstagramService, userService service_user.UserService) *PlatformHandler {
	return &PlatformHandler{
		twitterService:   twitterService,
		instagramService: instagramService,
		userService:      userService,
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
		return h.postToTwitter(c, email, platformDataJSON, files)

	case "instagram":
		return h.postToInstagram(c, email, platformDataJSON, files)
		// OTHER PLATFORMS COMING SOON HEHEHEHEE

	default:
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Unsupported platform"})
	}
}

func (h *PlatformHandler) postToTwitter(c echo.Context, email string, platformData string, files []*multipart.FileHeader) error {
	accessToken, accessSecret, err := h.userService.GetTwitterToken(email)
	if err != nil || accessToken == "" || accessSecret == "" {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "Twitter account not linked or tokens are missing"})
	}
	var twitterData struct {
		Content string `json:"content"`
	}
	if err := json.Unmarshal([]byte(platformData), &twitterData); err != nil {
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
}

func (h *PlatformHandler) postToInstagram(c echo.Context, email string, platformData string, files []*multipart.FileHeader) error {
	accessToken, instagramID, err := h.userService.GetInstagramCredentials(email)
	if err != nil || accessToken == "" || instagramID == "" {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "Instagram account not linked or tokens are missing"})
	}
	var instagramData struct {
		Caption string `json:"caption"`
		// THINK OF MORE PARAMS LATER BECAUSE I AM SLEEP DEPRIVED RIGHT NOW
	}
	if err := json.Unmarshal([]byte(platformData), &instagramData); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid format for platformData"})
	}
	if instagramData.Caption == "" && len(files) == 0 {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "An Instagram post must have either a caption or media."})
	}

	mediaURL, err := h.instagramService.PostToInstagram(accessToken, instagramID, instagramData.Caption, files)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to post to Instagram: " + err.Error()})
	}
	return c.JSON(http.StatusOK, map[string]string{"message": "Instagram post scheduled successfully!", "mediaURL": mediaURL})

}
