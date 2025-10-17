package handlers

import (
	"backend/services"
	
	"fmt"
	"github.com/labstack/echo-contrib/session"
	"github.com/labstack/echo/v4"
	"net/http"
	"log"
)

type TwitterHandler struct {
	twitterService services.TwitterService
	userService    services.UserService
}

func NewTwitterHandler(twitterService services.TwitterService, userService services.UserService) *TwitterHandler {
	return &TwitterHandler{
		twitterService: twitterService,
		userService:    userService,
	}
}

func (h *TwitterHandler) BeginTwitterLink(c echo.Context) error {
	// This endpoint MUST be protected by JWTMiddleware.
	ok, email, err := h.userService.IsLoggedIn(c)
	if err != nil {
		return err
	}
	if !ok {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "User not logged in"})
	}

	authURL, requestSecret, err := h.twitterService.GetAuthorizationURL()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to get Twitter authorization URL"})
	}

	sess, err := session.Get("twitter-link-session", c) // Use a specific session name
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to create session"})
	}

	sess.Values["requestSecret"] = requestSecret
	sess.Values["userEmail"] = email
	sess.Save(c.Request(), c.Response())

	// TO-DO: return the URL instead of redirecting from the backend.
	// The frontend can then open it in a popup.
	// return c.JSON(http.StatusOK, map[string]string{"authorization_url": authURL})

	return c.Redirect(http.StatusFound, authURL)
}

func (h *TwitterHandler) Callback(c echo.Context) error {
	const profilePath = "/profile"
	log.Println("[CALLBACK_TRACE] --- Callback handler initiated ---")

	// 1. Check session
	sess, err := session.Get("twitter-link-session", c)
	if err != nil {
		redirectURL := fmt.Sprintf("%s?status=error&provider=twitter&code=session_expired", profilePath)
		return c.Redirect(http.StatusSeeOther, redirectURL)
	}

	// 2. Check for user email in session
	email, ok := sess.Values["userEmail"].(string)
	if !ok {
		redirectURL := fmt.Sprintf("%s?status=error&provider=twitter&code=no_user_in_session", profilePath)
		return c.Redirect(http.StatusSeeOther, redirectURL)
	}

	// 3. Check for request secret in session
	requestSecret, ok := sess.Values["requestSecret"].(string)
	if !ok {
		redirectURL := fmt.Sprintf("%s?status=error&provider=twitter&code=no_secret_in_session", profilePath)
		return c.Redirect(http.StatusSeeOther, redirectURL)
	}

	sess.Options.MaxAge = -1 // Clean up session
	sess.Save(c.Request(), c.Response())

	// 4. Check if user denied the request on Twitter's site
	if c.QueryParam("denied") != "" {
		log.Println("[CALLBACK_TRACE] INFO: Step 4 redirecting. User denied authorization on Twitter.")
		redirectURL := fmt.Sprintf("%s?status=denied&provider=twitter", profilePath)
		return c.Redirect(http.StatusSeeOther, redirectURL)
	}
	log.Println("[CALLBACK_TRACE] Step 4 OK: User did not deny.")

	// 5. Check for OAuth parameters in the callback URL
	oauthToken := c.QueryParam("oauth_token")
	oauthVerifier := c.QueryParam("oauth_verifier")
	if oauthToken == "" || oauthVerifier == "" {
		log.Println("[CALLBACK_TRACE] FATAL: Step 5 failed. oauth_token or oauth_verifier is missing from URL.")
		redirectURL := fmt.Sprintf("%s?status=error&provider=twitter&code=invalid_callback_params", profilePath)
		return c.Redirect(http.StatusSeeOther, redirectURL)
	}
	log.Println("[CALLBACK_TRACE] Step 5 OK: OAuth token and verifier found in URL.")

	// 6. Exchange tokens with the Twitter API
	log.Println("[CALLBACK_TRACE] Attempting Step 6: Exchanging tokens with Twitter API...")
	accessToken, accessSecret, err := h.twitterService.GetAccessToken(oauthToken, requestSecret, oauthVerifier)
	if err != nil {
		// This is a very likely point of failure.
		log.Printf("[CALLBACK_TRACE] FATAL: Step 6 failed. Error from GetAccessToken: %v", err)
		redirectURL := fmt.Sprintf("%s?status=error&provider=twitter&code=token_exchange_failed", profilePath)
		return c.Redirect(http.StatusSeeOther, redirectURL)
	}
	log.Println("[CALLBACK_TRACE] Step 6 OK: Successfully exchanged tokens with Twitter.")

	// 7. Call the user service to update the database
	log.Println("[CALLBACK_TRACE] Attempting Step 7: Calling LinkTwitterAccount in user service...")
	err = h.userService.LinkTwitterAccount(email, accessToken, accessSecret)
	if err != nil {
		log.Printf("[CALLBACK_TRACE] FATAL: Step 7 failed. Error from LinkTwitterAccount: %v", err)
		redirectURL := fmt.Sprintf("%s?status=error&provider=twitter&code=db_link_failed", profilePath)
		return c.Redirect(http.StatusSeeOther, redirectURL)
	}
	log.Println("[CALLBACK_TRACE] Step 7 OK: LinkTwitterAccount returned successfully.")

	// 8. Success!
	log.Println("[CALLBACK_TRACE] --- Callback handler finished successfully. Redirecting to profile. ---")
	successRedirectURL := fmt.Sprintf("%s?status=success&provider=twitter", profilePath)
	return c.Redirect(http.StatusSeeOther, successRedirectURL)
}

func (h *TwitterHandler) CheckTwitterToken(c echo.Context) error {
	ok, email, err := h.userService.IsLoggedIn(c)
	if err != nil {
		return err
	}
	if !ok {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "User not logged in"})
	}
	
	accessToken, accessSecret, err := h.userService.GetTwitterTokens(email)
	if err != nil || accessToken == "" || accessSecret == "" {
		return c.JSON(http.StatusOK, map[string]interface{}{
			"twitterLinked": false,
			"twitterTokenValid":    false,
		})
	}

	valid, err := h.twitterService.CheckTokensValid(accessToken, accessSecret)
	return c.JSON(http.StatusOK, map[string]interface{}{
		"twitterLinked": true,
		"twitterTokenValid":    valid && err == nil,
	})
}
