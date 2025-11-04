package middlewares

import (
	"log"
	"net/http"
	"strings"
	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"
	"fmt"
	"errors"
)

const login_path = "/login"

// JWTMiddleware creates an Echo middleware to validate JWT tokens.
func JWTMiddleware(secret []byte, excludedPaths []string) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// 1. Skip JWT check if the path is excluded.
			if isPathExcluded(c.Request().URL.Path, excludedPaths) {
				return next(c)
			}

			// 2. Get token from the cookie.
			cookie, err := c.Cookie("jwt_token")
			if err != nil {
				log.Println("JWTMiddleware: no cookie found, redirecting to /login")
				return c.Redirect(http.StatusSeeOther, login_path)
			}

			// 3. Validate the token and extract claims.
			claims, err := validateToken(cookie.Value, secret)
			if err != nil {
				log.Printf("JWTMiddleware: invalid token (%v), redirecting to /login", err)
				return c.Redirect(http.StatusSeeOther, login_path)
			}

			// 4. Store claims in context and proceed.
			c.Set("userClaims", claims)
			return next(c)
		}
	}
}

func isPathExcluded(path string, excluded []string) bool {
	for _, p := range excluded {
		if strings.HasPrefix(path, p) {
			return true
		}
	}
	return false
}

func validateToken(tokenString string, secret []byte) (jwt.MapClaims, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// Check the signing method.
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return secret, nil
	})

	// This single check handles parsing errors (e.g., malformed token, signature mismatch).
	if err != nil {
		return nil, err
	}

	// Check if the token is valid and if the claims can be asserted to MapClaims.
	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, errors.New("invalid token or claims type")
}

const default_redirect_path = "/"

// RedirectIfAuthenticated redirects authenticated users away from login/signup pages.
func RedirectIfAuthenticated(secret []byte) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// Try to validate the token. If it's valid, the user is authenticated.
			_, err := isAuthenticated(c, secret)

			if err == nil {
				// User is authenticated, redirect them away from the login/signup page.
				log.Println("RedirectIfAuthenticated: user authenticated, redirecting to", default_redirect_path)
				return c.Redirect(http.StatusSeeOther, default_redirect_path)
			}

			// User is not authenticated or token is invalid/missing, proceed to the next handler.
			return next(c)
		}
	}
}

// isAuthenticated checks if a user is authenticated by validating their JWT token.
// It returns true if authenticated, false otherwise
func isAuthenticated(c echo.Context, secret []byte) (jwt.MapClaims, error) {
	cookie, err := c.Cookie("jwt_token")
	if err != nil {
		// No JWT cookie found, user is not authenticated.
		return nil, errors.New("jwt_token cookie not found")
	}

	tokenString := cookie.Value
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return secret, nil
	})

	if err != nil || !token.Valid {
		// Token parsing failed or token is invalid.
		return nil, fmt.Errorf("token validation failed: %w", err)
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		// Claims could not be asserted to MapClaims.
		return nil, errors.New("could not assert token claims to MapClaims")
	}

	return claims, nil
}
