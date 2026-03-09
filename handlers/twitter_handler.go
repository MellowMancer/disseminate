package handlers

import (
	service_twitter "backend/services/twitter"
	service_user "backend/services/user"
	"backend/utils"
	"fmt"
	"log"
	"net/http"

	"github.com/labstack/echo-contrib/session"
	"github.com/labstack/echo/v4"
)

type TwitterHandler struct {
	twitterService service_twitter.TwitterService
	userService    service_user.UserService
	frontendURL    string
}

func NewTwitterHandler(twitterService service_twitter.TwitterService, userService service_user.UserService, frontendURL string) *TwitterHandler {
	return &TwitterHandler{
		twitterService: twitterService,
		userService:    userService,
		frontendURL:    frontendURL,
	}
}

func (h *TwitterHandler) BeginTwitterLink(c echo.Context) error {
	// This endpoint MUST be protected by JWTMiddleware.
	email, err := h.userService.IsLoggedIn(c)
	if err != nil {
		return utils.UnauthorizedResponse(c, "Authentication required")
	}

	authURL, requestSecret, err := h.twitterService.GetAuthorizationURL()
	if err != nil {
		return utils.InternalErrorResponse(c, err, "Twitter authorization URL generation")
	}

	sess, err := session.Get("twitter-link-session", c)
	if err != nil {
		return utils.InternalErrorResponse(c, err, "session creation")
	}

	sess.Values["requestSecret"] = requestSecret
	sess.Values["userEmail"] = email
	if err := sess.Save(c.Request(), c.Response()); err != nil {
		return utils.InternalErrorResponse(c, err, "session save")
	}

	return c.Redirect(http.StatusFound, authURL)
}

func (h *TwitterHandler) Callback(c echo.Context) error {
	profilePath := h.frontendURL + "/profile"
	log.Println("[CALLBACK_TRACE] --- Callback handler initiated ---")

	// 1. Check session
	sess, err := session.Get("twitter-link-session", c)
	if err != nil {
		log.Printf("[ERROR] Failed to get Twitter session: %v", err)
		redirectURL := fmt.Sprintf("%s?status=error&provider=twitter&code=session_expired", profilePath)
		return c.Redirect(http.StatusSeeOther, redirectURL)
	}

	// 2. Check for user email in session (with safe type assertion)
	email, ok := sess.Values["userEmail"].(string)
	if !ok || email == "" {
		log.Printf("[ERROR] User email not found in Twitter session")
		redirectURL := fmt.Sprintf("%s?status=error&provider=twitter&code=no_user_in_session", profilePath)
		return c.Redirect(http.StatusSeeOther, redirectURL)
	}

	// 3. Check for request secret in session (with safe type assertion)
	requestSecret, ok := sess.Values["requestSecret"].(string)
	if !ok || requestSecret == "" {
		log.Printf("[ERROR] Request secret not found in Twitter session")
		redirectURL := fmt.Sprintf("%s?status=error&provider=twitter&code=no_secret_in_session", profilePath)
		return c.Redirect(http.StatusSeeOther, redirectURL)
	}

	// 4. Check if user denied the request on Twitter's site
	if c.QueryParam("denied") != "" {
		log.Println("[CALLBACK_TRACE] INFO: User denied authorization on Twitter")
		// Clean up session on denial
		sess.Options.MaxAge = -1
		sess.Save(c.Request(), c.Response())
		redirectURL := fmt.Sprintf("%s?status=denied&provider=twitter", profilePath)
		return c.Redirect(http.StatusSeeOther, redirectURL)
	}
	log.Println("[CALLBACK_TRACE] Step 4 OK: User did not deny")

	// 5. Check for OAuth parameters in the callback URL
	oauthToken := c.QueryParam("oauth_token")
	oauthVerifier := c.QueryParam("oauth_verifier")
	if oauthToken == "" || oauthVerifier == "" {
		log.Println("[ERROR] oauth_token or oauth_verifier is missing from callback URL")
		redirectURL := fmt.Sprintf("%s?status=error&provider=twitter&code=invalid_callback_params", profilePath)
		return c.Redirect(http.StatusSeeOther, redirectURL)
	}
	log.Println("[CALLBACK_TRACE] Step 5 OK: OAuth token and verifier found")

	// 6. Exchange tokens with the Twitter API
	log.Println("[CALLBACK_TRACE] Attempting Step 6: Exchanging tokens with Twitter API...")
	accessToken, accessSecret, err := h.twitterService.GetAccessToken(oauthToken, requestSecret, oauthVerifier)
	if err != nil {
		log.Printf("[ERROR] Failed to exchange Twitter token: %v", err)
		redirectURL := fmt.Sprintf("%s?status=error&provider=twitter&code=token_exchange_failed", profilePath)
		return c.Redirect(http.StatusSeeOther, redirectURL)
	}
	log.Println("[CALLBACK_TRACE] Step 6 OK: Successfully exchanged tokens with Twitter")

	// 7. Call the user service to update the database
	log.Println("[CALLBACK_TRACE] Attempting Step 7: Saving Twitter tokens...")
	err = h.userService.SaveTwitterToken(email, accessToken, accessSecret)
	if err != nil {
		log.Printf("[ERROR] Failed to save Twitter token for user %s: %v", email, err)
		redirectURL := fmt.Sprintf("%s?status=error&provider=twitter&code=db_link_failed", profilePath)
		return c.Redirect(http.StatusSeeOther, redirectURL)
	}
	log.Println("[CALLBACK_TRACE] Step 7 OK: Twitter tokens saved successfully")

	// 8. Clean up session AFTER successful completion
	sess.Options.MaxAge = -1
	sess.Save(c.Request(), c.Response())
	log.Println("[CALLBACK_TRACE] Session cleaned up")

	// 9. Success!
	log.Println("[CALLBACK_TRACE] --- Callback handler finished successfully. Redirecting to profile. ---")
	successRedirectURL := fmt.Sprintf("%s?status=success&provider=twitter", profilePath)
	return c.Redirect(http.StatusSeeOther, successRedirectURL)
}
