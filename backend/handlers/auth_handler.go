package handlers

import (
	"log"
	"net/http"

	"backend/models"
	"backend/services"

    "github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"
	"golang.org/x/crypto/bcrypt"
)

type Handler struct {
    UserService services.UserService
}

func NewHandler(userService services.UserService) *Handler {
    return &Handler{UserService: userService}
}

func (h *Handler) SignUp(c echo.Context) error {
    var req models.User
    if err := c.Bind(&req); err != nil {
        log.Printf("SignUp Bind error: %v", err)
        return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid input"})
    }

    hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
    if err != nil {
        log.Printf("SignUp Hash error: %v", err)
        return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Error hashing password"})
    }
    req.Password = string(hashedPassword)

    if err := h.UserService.CreateUser(&req); err != nil {
        log.Printf("SignUp CreateUser error: %v", err)
        return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
    }

    return c.JSON(http.StatusCreated, map[string]string{"message": "User created"})
}

func (h *Handler) Login(c echo.Context) error {
	var req models.User
    if err := c.Bind(&req); err != nil {
        return c.JSON(http.StatusBadRequest, "Invalid input")
    }

	// hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
    // if err != nil {
    //     return c.JSON(http.StatusInternalServerError, "Error hashing password")
    // }
	// req.Password = string(hashedPassword)

    tokenString, err := h.UserService.LoginUser(&req); 
	if err != nil {
        return c.JSON(http.StatusInternalServerError, err)
    }

	cookie := new(http.Cookie)
    cookie.Name = "jwt_token"
    cookie.Value = tokenString
    cookie.Path = "/"
    cookie.HttpOnly = true
    cookie.Secure = false // ONLY FOR NOW, REMEMBER TO TURN THIS ON FOR DEPLOYMENT
    cookie.SameSite = http.SameSiteStrictMode
    cookie.MaxAge = 86400 * 3

    c.SetCookie(cookie)
	

	return c.JSON(http.StatusOK, "Successfully logged in")
}

func (h *Handler) Logout(c echo.Context) error {
    cookie := &http.Cookie{
        Name:     "jwt_token",
        Value:    "",                 
        Path:     "/",                // Same path as the original cookie
        HttpOnly: true,
        MaxAge:   -1,                 // MaxAge < 0 deletes the cookie
        Secure:   false,              // Set true in production if using HTTPS
        SameSite: http.SameSiteStrictMode,
    }
    c.SetCookie(cookie)

    return c.Redirect(http.StatusSeeOther, "/login")
}

func (h *Handler) AuthStatus(c echo.Context) error {
    cookie, err := c.Cookie("jwt_token")
    if err != nil {
        return c.JSON(http.StatusOK, map[string]any{"authenticated": false})
    }
    token, err := jwt.Parse(cookie.Value, func(token *jwt.Token) (interface{}, error) {
        if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
            return nil, echo.ErrUnauthorized
        }
        jwtSecret := h.UserService.GetJWTSecret()
        return jwtSecret, nil 
    })
    if err != nil || !token.Valid {
        return c.JSON(http.StatusOK, map[string]any{"authenticated": false})
    }

    claims := token.Claims.(jwt.MapClaims)
    email := claims["sub"].(string)
    return c.JSON(http.StatusOK, map[string]any{
        "authenticated": true,
        "email": email,
    })
}



