package handlers

import (
	service_instagram "backend/services/instagram"
	service_user "backend/services/user"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"github.com/labstack/echo-contrib/session"
	"github.com/labstack/echo/v4"
	"log"
	"net/http"
)

type InstagramHandler struct {
	instagramService service_instagram.InstagramService
	userService      service_user.UserService
}

func NewInstagramHandler(instagramService service_instagram.InstagramService, userService service_user.UserService) *InstagramHandler {
	return &InstagramHandler{
		instagramService: instagramService,
		userService:      userService,
	}
}

func generateState() (string, error) {
	b := make([]byte, 16) // 16 bytes
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}
	state := base64.URLEncoding.EncodeToString(b)
	return state, nil
}

// BeginInstagramLink initiates the Instagram OAuth linking process.
func (h *InstagramHandler) BeginInstagramLink(c echo.Context) error {
	ok, email, err := h.userService.IsLoggedIn(c)
	if err != nil {
		return err
	}
	if !ok {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "User not logged in"})
	}

	sess, err := session.Get("instagram-link-session", c)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to create session"})
	}

	state, err := generateState()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to generate state"})
	}

	sess.Values["state"] = state
	sess.Values["userEmail"] = email
	sess.Save(c.Request(), c.Response())

	responseWriter := c.Response().Writer
	request := c.Request()
	h.instagramService.HandleLogin(responseWriter, request, state)
	return nil
}

// CheckInstagramToken checks the validity of the Instagram token.
func (h *InstagramHandler) CheckInstagramToken(c echo.Context) error {
	ok, email, err := h.userService.IsLoggedIn(c)
	if err != nil {
		return err
	}
	if !ok {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "User not logged in"})
	}
	
	accessToken, _, err := h.userService.GetInstagramCredentials(email)
	if err != nil || accessToken == "" {
		return c.JSON(http.StatusOK, map[string]interface{}{
			"instagramLinked": false,
			"instagramTokenValid":    false,
		})
	}

	err = h.instagramService.CheckTokensValid(accessToken)
	return c.JSON(http.StatusOK, map[string]interface{}{
		"instagramLinked": true,
		"instagramTokenValid":  err == nil,
	})
}

func (h *InstagramHandler) Callback(c echo.Context) error {
	const profilePath = "/profile"

	log.Println("[CALLBACK_TRACE] --- Callback handler initiated ---")

	// 1. Check session
	sess, err := session.Get("instagram-link-session", c)
	if err != nil {
		redirectURL := fmt.Sprintf("%s?status=error&provider=instagram&code=session_expired", profilePath)
		return c.Redirect(http.StatusSeeOther, redirectURL)
	}

	// 2. Check for user email in session
	email, ok := sess.Values["userEmail"].(string)
	if !ok {
		redirectURL := fmt.Sprintf("%s?status=error&provider=twitter&code=no_user_in_session", profilePath)
		return c.Redirect(http.StatusSeeOther, redirectURL)
	}

	// 3. Handle Instagram callback
	r := c.Request()
	code := r.URL.Query().Get("code")
	if code == "" {
		return fmt.Errorf("code not found in request")
	}
	state := r.URL.Query().Get("state")
	if state != sess.Values["state"].(string) {
		return fmt.Errorf("state mismatch")
	}

	sess.Options.MaxAge = -1 // Clean up session
	sess.Save(c.Request(), c.Response())

	token, expiresIn, err := h.instagramService.GetAccessToken(code)
	if err != nil {
		return fmt.Errorf("failed to exchange token: %v", err)
	}
	log.Println(email)

	err = h.userService.SaveInstagramToken(email, token, expiresIn)
	if err != nil {
		return fmt.Errorf("failed to link instagram to your account")
	}

	successRedirectURL := fmt.Sprintf("%s?status=success&provider=instagram", profilePath)
	return c.Redirect(http.StatusSeeOther, successRedirectURL)
}
