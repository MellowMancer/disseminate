package handlers

import (
	"log"
	"fmt"
	"net/http"
	"backend/models"
	service_user "backend/services/user"
    "github.com/labstack/echo-contrib/session"
	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"
	"golang.org/x/crypto/bcrypt"
)


type Handler struct {
	UserService service_user.UserService
}

func NewHandler(userService service_user.UserService) *Handler {
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

	tokenString, err := h.UserService.LoginUser(&req)
	if err != nil {
		log.Printf("Login failed for email '%s': %v", req.Email, err)

		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Invalid email or password"})
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
	jwtCookie := &http.Cookie{
		Name:     "jwt_token",
		Value:    "",
		Path:     "/", // Same path as the original cookie
		HttpOnly: true,
		MaxAge:   -1,    // MaxAge < 0 deletes the cookie
		Secure:   false, // Set true in production if using HTTPS
		SameSite: http.SameSiteStrictMode,
	}
	c.SetCookie(jwtCookie)

    twitterSess, err := session.Get("twitter-link-session", c)
	if err == nil { // Only proceed if we can successfully get the session
		twitterSess.Options.MaxAge = -1
		twitterSess.Save(c.Request(), c.Response())
	}

	return c.Redirect(http.StatusSeeOther, "/login")
}

func (h *Handler) AuthStatus(c echo.Context) error {
	log.Println("[AuthStatus] --- AuthStatus handler initiated ---")
	fmt.Println("[AuthStatus] --- AuthStatus handler initiated ---")
    cookie, err := c.Cookie("jwt_token")
    if err != nil {
        // Log if cookie is missing
        log.Println("[AuthStatus] No jwt_token cookie in request:", err)
        return c.JSON(http.StatusOK, map[string]any{"authenticated": false})
    }

    log.Println("[AuthStatus] jwt_token cookie received:", cookie.Value)

    token, err := jwt.Parse(cookie.Value, func(token *jwt.Token) (interface{}, error) {
        if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
            log.Println("[AuthStatus] Unexpected signing method:", token.Header["alg"])
            return nil, echo.ErrUnauthorized
        }
        jwtSecret := h.UserService.GetJWTSecret()
        return jwtSecret, nil
    })
    if err != nil {
        // Log detailed JWT parse error
        log.Println("[AuthStatus] JWT parsing error:", err)
    }
    if token == nil || !token.Valid {
        // Log if JWT is not valid
        log.Println("[AuthStatus] Invalid or nil token. Valid:", token != nil && token.Valid)
        return c.JSON(http.StatusOK, map[string]any{"authenticated": false})
    }

    claims, ok := token.Claims.(jwt.MapClaims)
    if !ok {
        // Log if claims extraction fails
        log.Println("[AuthStatus] Failed to extract jwt.MapClaims")
        return c.JSON(http.StatusOK, map[string]any{"authenticated": false})
    }

    email, ok := claims["sub"].(string)
    if !ok {
        // Log if 'sub' claim is missing or not a string
        log.Println("[AuthStatus] 'sub' claim missing or not string in JWT claims")
        return c.JSON(http.StatusOK, map[string]any{"authenticated": false})
    }

    // Print claims (for debugging; remove in production for privacy/security)
    log.Println("[AuthStatus] JWT claims:", claims)

    return c.JSON(http.StatusOK, map[string]any{
        "authenticated": true,
        "email":         email,
    })
}

