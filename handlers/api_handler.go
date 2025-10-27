package handlers

import (
    "net/http"
    "github.com/labstack/echo/v4"
	"github.com/golang-jwt/jwt/v5"
)

func ProtectedHandler(c echo.Context) error {
    user := c.Get("user").(jwt.MapClaims)
    email := user["email"].(string)
    return c.JSON(http.StatusOK, map[string]string{"message": "Hello " + email})
}

