package handlers

import (
	service_instagram "backend/services/instagram"
	service_user "backend/services/user"
	"backend/utils"
	"fmt"
	"log"
	"net/http"

	"github.com/labstack/echo-contrib/session"
	"github.com/labstack/echo/v4"
)

type InstagramHandler struct {
	instagramService service_instagram.InstagramService
	userService      service_user.UserService
	frontendURL      string
}

func NewInstagramHandler(instagramService service_instagram.InstagramService, userService service_user.UserService, frontendURL string) *InstagramHandler {
	return &InstagramHandler{
		instagramService: instagramService,
		userService:      userService,
		frontendURL:      frontendURL,
	}
}


// BeginInstagramLink initiates the Instagram OAuth linking process.
func (h *InstagramHandler) BeginInstagramLink(c echo.Context) error {
	email, err := h.userService.IsLoggedIn(c)
	if err != nil {
		return utils.UnauthorizedResponse(c, "Authentication required")
	}

	sess, err := session.Get("instagram-link-session", c)
	if err != nil {
		return utils.InternalErrorResponse(c, err, "session creation")
	}

	state, err := utils.GenerateOAuthState()
	if err != nil {
		return utils.InternalErrorResponse(c, err, "state generation")
	}

	sess.Values["state"] = state
	sess.Values["userEmail"] = email
	if err := sess.Save(c.Request(), c.Response()); err != nil {
		return utils.InternalErrorResponse(c, err, "session save")
	}

	responseWriter := c.Response().Writer
	request := c.Request()
	h.instagramService.HandleLogin(responseWriter, request, state)
	return nil
}


func (h *InstagramHandler) Callback(c echo.Context) error {
	profilePath := h.frontendURL + "/profile"

	log.Println("[CALLBACK_TRACE] --- Callback handler initiated ---")

	// 1. Check session
	sess, err := session.Get("instagram-link-session", c)
	if err != nil {
		log.Printf("[ERROR] Failed to get Instagram session: %v", err)
		redirectURL := fmt.Sprintf("%s?status=error&provider=instagram&code=session_expired", profilePath)
		return c.Redirect(http.StatusSeeOther, redirectURL)
	}

	log.Println("[CALLBACK_TRACE] --- Session retrieved successfully ---")

	// 2. Check for user email in session (with safe type assertion)
	email, ok := sess.Values["userEmail"].(string)
	if !ok || email == "" {
		log.Printf("[ERROR] User email not found in Instagram session")
		redirectURL := fmt.Sprintf("%s?status=error&provider=instagram&code=no_user_in_session", profilePath)
		return c.Redirect(http.StatusSeeOther, redirectURL)
	}

	log.Println("[CALLBACK_TRACE] --- User email found in session: ", email, " ---")

	// 3. Validate OAuth callback parameters
	r := c.Request()
	code := r.URL.Query().Get("code")
	if code == "" {
		log.Printf("[ERROR] OAuth code not found in Instagram callback")
		redirectURL := fmt.Sprintf("%s?status=error&provider=instagram&code=missing_code", profilePath)
		return c.Redirect(http.StatusSeeOther, redirectURL)
	}

	// 4. Validate state (with safe type assertion)
	receivedState := r.URL.Query().Get("state")
	sessionState, ok := sess.Values["state"].(string)
	if !ok || sessionState == "" || receivedState != sessionState {
		log.Printf("[SECURITY] Instagram OAuth state mismatch - potential CSRF attack")
		redirectURL := fmt.Sprintf("%s?status=error&provider=instagram&code=invalid_state", profilePath)
		return c.Redirect(http.StatusSeeOther, redirectURL)
	}

	log.Println("[CALLBACK_TRACE] --- State validated successfully ---")

	// 5. Exchange code for token BEFORE cleaning up session
	token, expiresIn, err := h.instagramService.GetAccessToken(code)
	if err != nil {
		log.Printf("[ERROR] Failed to exchange Instagram token: %v", err)
		redirectURL := fmt.Sprintf("%s?status=error&provider=instagram&code=token_exchange_failed", profilePath)
		return c.Redirect(http.StatusSeeOther, redirectURL)
	}

	log.Println("[CALLBACK_TRACE] --- Access token obtained successfully ---")

	// 6. Save token to user account
	err = h.userService.SaveInstagramToken(email, token, expiresIn)
	if err != nil {
		log.Printf("[ERROR] Failed to save Instagram token for user %s: %v", email, err)
		redirectURL := fmt.Sprintf("%s?status=error&provider=instagram&code=save_failed", profilePath)
		return c.Redirect(http.StatusSeeOther, redirectURL)
	}

	log.Println("[CALLBACK_TRACE] --- Instagram linked to user account successfully ---")

	// 7. Clean up session AFTER successful completion
	sess.Options.MaxAge = -1
	sess.Save(c.Request(), c.Response())

	log.Println("[CALLBACK_TRACE] --- Session cleaned up ---")

	successRedirectURL := fmt.Sprintf("%s?status=success&provider=instagram", profilePath)
	return c.Redirect(http.StatusSeeOther, successRedirectURL)
}
