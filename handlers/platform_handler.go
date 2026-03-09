package handlers

import (
	service_instagram "backend/services/instagram"
	service_twitter "backend/services/twitter"
	service_user "backend/services/user"
	"backend/utils"
	"encoding/json"
	"mime/multipart"
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"
)

const (
	maxFileSize      = 50 * 1024 * 1024 // 50MB per file
	maxFiles         = 10                // Maximum 10 files per post
	maxPlatformData  = 1 * 1024 * 1024   // 1MB for platformData JSON
)

var allowedPlatforms = map[string]bool{
	"twitter":    true,
	"instagram":  true,
	"bluesky":    false, // Coming soon
	"mastodon":   false, // Coming soon
	"artstation": false, // Coming soon
	"youtube":    false, // Coming soon
}

var allowedMimeTypes = map[string]bool{
	"image/jpeg": true,
	"image/png":  true,
	"image/gif":  true,
	"image/webp": true,
	"video/mp4":  true,
	"video/mov":  true,
}

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
	email, err := h.userService.IsLoggedIn(c)
	if err != nil {
		return utils.UnauthorizedResponse(c, "Authentication required")
	}

	// Validate platform
	platform := strings.ToLower(strings.TrimSpace(c.FormValue("platform")))
	if platform == "" {
		return utils.BadRequestResponse(c, "Platform not specified")
	}

	// Check if platform is in whitelist
	enabled, exists := allowedPlatforms[platform]
	if !exists {
		return utils.ErrorResponse(c, utils.WrapError(
			utils.ErrInvalidPlatform,
			"Invalid platform specified",
			http.StatusBadRequest,
		))
	}
	if !enabled {
		return utils.BadRequestResponse(c, "Platform not yet supported")
	}

	// Validate platformData size
	platformDataJSON := c.FormValue("platformData")
	if len(platformDataJSON) > maxPlatformData {
		return utils.ErrorResponse(c, utils.WrapError(
			utils.ErrInvalidInput,
			"Platform data exceeds size limit",
			http.StatusBadRequest,
		))
	}

	// Get and validate uploaded files
	form, err := c.MultipartForm()
	if err != nil {
		return utils.BadRequestResponse(c, "Invalid form data")
	}
	files := form.File["media"]

	// Validate file count
	if len(files) > maxFiles {
		return utils.ErrorResponse(c, utils.WrapError(
			utils.ErrInvalidInput,
			"Too many files uploaded. Maximum is 10 files",
			http.StatusBadRequest,
		))
	}

	// Validate file sizes and types
	for _, fileHeader := range files {
		if fileHeader.Size > maxFileSize {
			return utils.ErrorResponse(c, utils.WrapError(
				utils.ErrFileTooLarge,
				"File size exceeds 50MB limit",
				http.StatusBadRequest,
			))
		}

		// Check MIME type
		contentType := fileHeader.Header.Get("Content-Type")
		if !allowedMimeTypes[contentType] {
			return utils.ErrorResponse(c, utils.WrapError(
				utils.ErrInvalidMediaType,
				"Invalid file type. Allowed: JPEG, PNG, GIF, WebP, MP4, MOV",
				http.StatusBadRequest,
			))
		}
	}

	switch platform {
	case "twitter":
		return h.postToTwitter(c, email, platformDataJSON, files)

	case "instagram":
		return h.postToInstagram(c, email, platformDataJSON, files)

	default:
		return utils.BadRequestResponse(c, "Platform not supported")
	}
}

func (h *PlatformHandler) postToTwitter(c echo.Context, email string, platformData string, files []*multipart.FileHeader) error {
	accessToken, accessSecret, err := h.userService.GetTwitterToken(email)
	if err != nil || accessToken == "" || accessSecret == "" {
		return utils.ErrorResponse(c, utils.WrapError(
			utils.ErrAccountNotLinked,
			"Twitter account not linked. Please link your account first",
			http.StatusUnauthorized,
		))
	}

	var twitterData struct {
		Content string `json:"content"`
	}
	if err := json.Unmarshal([]byte(platformData), &twitterData); err != nil {
		return utils.ErrorResponse(c, utils.NewAppError(
			utils.ErrInvalidInput,
			"Invalid platform data format",
			http.StatusBadRequest,
			err.Error(),
		))
	}

	if twitterData.Content == "" && len(files) == 0 {
		return utils.BadRequestResponse(c, "A tweet must have either text content or media")
	}

	err = h.twitterService.PostTweet(accessToken, accessSecret, twitterData.Content, files)
	if err != nil {
		return utils.ErrorResponse(c, utils.NewAppError(
			utils.ErrPostFailed,
			"Failed to post tweet. Please try again",
			http.StatusInternalServerError,
			err.Error(),
		))
	}

	return utils.SuccessResponse(c, http.StatusOK, "Tweet posted successfully", nil)
}

func (h *PlatformHandler) postToInstagram(c echo.Context, email string, platformData string, files []*multipart.FileHeader) error {
	accessToken, instagramID, err := h.userService.GetInstagramCredentials(email)
	if err != nil || accessToken == "" || instagramID == "" {
		return utils.ErrorResponse(c, utils.WrapError(
			utils.ErrAccountNotLinked,
			"Instagram account not linked. Please link your account first",
			http.StatusUnauthorized,
		))
	}

	var instagramData struct {
		Caption string `json:"caption"`
	}
	if err := json.Unmarshal([]byte(platformData), &instagramData); err != nil {
		return utils.ErrorResponse(c, utils.NewAppError(
			utils.ErrInvalidInput,
			"Invalid platform data format",
			http.StatusBadRequest,
			err.Error(),
		))
	}

	if instagramData.Caption == "" && len(files) == 0 {
		return utils.BadRequestResponse(c, "An Instagram post must have either a caption or media")
	}

	mediaURL, err := h.instagramService.PostToInstagram(accessToken, instagramID, instagramData.Caption, files)
	if err != nil {
		return utils.ErrorResponse(c, utils.NewAppError(
			utils.ErrPostFailed,
			"Failed to post to Instagram. Please try again",
			http.StatusInternalServerError,
			err.Error(),
		))
	}

	return utils.SuccessResponse(c, http.StatusOK, "Instagram post created successfully", map[string]string{
		"mediaURL": mediaURL,
	})
}
