package middlewares

import (
	"log"
	"net/http"
	"strings"
	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"
)

const login_path = "/login"

// JWTMiddleware verifies JWT from cookie and redirects unauthenticated users to login
func JWTMiddleware(secret []byte, excludedPaths []string) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			path := c.Request().URL.Path
            for _, exclude := range excludedPaths {
                if strings.HasPrefix(path, exclude) {
                    return next(c) // Skip JWT check for these paths
                }
            }
			cookie, err := c.Cookie("jwt_token")
			if err != nil {
				log.Println("JWTMiddleware: no cookie found, redirecting to /login")
				return c.Redirect(http.StatusSeeOther, login_path)
			}

			tokenString := cookie.Value
			token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
				if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
					return nil, echo.NewHTTPError(http.StatusUnauthorized, "Unexpected signing method")
				}
				return secret, nil
			})

			if err != nil {
				log.Printf("JWTMiddleware: token parse error: %v", err)
				return c.Redirect(http.StatusSeeOther, login_path)
			}
			if !token.Valid {
				log.Println("JWTMiddleware: invalid token, redirecting to /login")
				return c.Redirect(http.StatusSeeOther, login_path)
			}

			// Store claims in context for handlers to access
			if claims, ok := token.Claims.(jwt.MapClaims); ok {
				c.Set("userClaims", claims)
			}

			return next(c)
		}
	}
}

// RedirectIfAuthenticated redirects authenticated users away from login/signup pages
func RedirectIfAuthenticated(secret []byte) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			cookie, err := c.Cookie("jwt_token")
			if err == nil {
				tokenString := cookie.Value
				token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
					if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
						return nil, echo.NewHTTPError(http.StatusUnauthorized, "Unexpected signing method")
					}
					return secret, nil
				})
				if err == nil && token.Valid {
					log.Println("RedirectIfAuthenticated: user authenticated, redirecting to /")
					return c.Redirect(http.StatusSeeOther, "/")
				}
			}
			return next(c)
		}
	}
}
