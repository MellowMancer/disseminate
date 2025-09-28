package handlers

import (
	"backend/services"
	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo-contrib/session"
	"github.com/labstack/echo/v4"
	"net/http"
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

func (h *TwitterHandler) Login(c echo.Context) error {
	authURL, requestToken, requestSecret, err := h.twitterService.GetAuthorizationURL()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to get Twitter authorization URL"})
	}

	sess, _ := session.Get("session", c)
	sess.Values["requestSecret"] = requestSecret
	sess.Values["requestToken"] = requestToken
	sess.Save(c.Request(), c.Response())

	return c.Redirect(http.StatusFound, authURL.String())

}

func (h *TwitterHandler) Callback(c echo.Context) error {
	claims := c.Get("userClaims").(jwt.MapClaims)
	email := claims["email"].(string)

	sess, _ := session.Get("session", c)
	requestSecret, ok := sess.Values["requestSecret"].(string)
	if !ok {
		return c.JSON(http.StatusBadRequest, "requestSecret not found")
	}

	if c.QueryParam("oauth_denied") != "" {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "User denied the authorization"})
	}

	oauthToken := c.QueryParam("oauth_token")
	oauthVerifier := c.QueryParam("oauth_verifier")

	if oauthToken == "" || oauthVerifier == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "missing OAuth token or verifier"})
	}

	accessToken, accessSecret, err := h.twitterService.GetAccessToken(oauthToken, requestSecret, oauthVerifier)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, "failed to get access token")
	}

	err = h.userService.LinkTwitterAccount(email, accessToken, accessSecret)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to link Twitter account"})
	}

	return c.JSON(http.StatusOK, map[string]string{"message": "Twitter account linked successfully"})
}

func (h *TwitterHandler) PostTweet(c echo.Context) error {
	// Post a tweet on behalf of the user
	return nil
}

func (h *TwitterHandler) CheckTwitterToken(c echo.Context) error {
    claims := c.Get("userClaims").(jwt.MapClaims)
    email := claims["email"].(string)

    accessToken, accessSecret, err := h.userService.GetTwitterTokens(email)
    if err != nil || accessToken == "" || accessSecret == "" {
        return c.JSON(http.StatusOK, map[string]interface{}{
            "twitterLinked": false,
            "tokenValid":    false,
        })
    }

    valid, err := h.twitterService.CheckTokensValid(accessToken, accessSecret)
    return c.JSON(http.StatusOK, map[string]interface{}{
        "twitterLinked": true,
        "tokenValid":    valid && err == nil,
    })
}
