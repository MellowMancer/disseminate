package middlewares

import (
	"net/http"
	service_twitter "backend/services/twitter"
    service_user "backend/services/user"
	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"
)

func TwitterValidationMiddleware(twitterService service_twitter.TwitterService, userService service_user.UserService) echo.MiddlewareFunc {
    return func(next echo.HandlerFunc) echo.HandlerFunc {
        return func(c echo.Context) error {
            claims := c.Get("userClaims").(jwt.MapClaims)
            email := claims["email"].(string)

            accessToken, accessSecret, err := userService.GetTwitterToken(email)
            if err != nil {
                return c.JSON(http.StatusUnauthorized, map[string]string{"error": "Twitter tokens missing"})
            }

            valid, err := twitterService.CheckTokensValid(accessToken, accessSecret)
            if err != nil || !valid {
                return c.JSON(http.StatusUnauthorized, map[string]string{"error": "Invalid or expired Twitter tokens"})
            }

            return next(c)
        }
    }
}
