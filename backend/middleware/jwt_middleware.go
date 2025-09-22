package middleware

import (
    "net/http"

    "github.com/labstack/echo/v4"
    "github.com/golang-jwt/jwt/v5"
)

func JWTMiddleware(secret []byte) echo.MiddlewareFunc {
    return func(next echo.HandlerFunc) echo.HandlerFunc {
        return func(c echo.Context) error {
            cookie, err := c.Cookie("jwt_token")
            if err != nil {
                return c.Redirect(http.StatusSeeOther, "/login")
            }
            tokenString := cookie.Value

            token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
                if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
                    return nil, echo.NewHTTPError(http.StatusUnauthorized, "Unexpected signing method")
                }
                return secret, nil
            })

            if err != nil || !token.Valid {
                return c.Redirect(http.StatusSeeOther, "/login")
            }
            return next(c)
        }
    }
}


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
                    return c.Redirect(http.StatusSeeOther, "/")
                }
            }
            return next(c)
        }
    }
}