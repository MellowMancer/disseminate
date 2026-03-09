package handlers

import (
	"backend/models"
	service_user "backend/services/user"
	"backend/utils"
	"log"
	"net/http"
	"unicode"

	emailverifier "github.com/AfterShip/email-verifier"
	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo-contrib/session"
	"github.com/labstack/echo/v4"
	"golang.org/x/crypto/bcrypt"
)

type Handler struct {
	UserService      service_user.UserService
	IsProduction     bool
}

func NewHandler(userService service_user.UserService, isProduction bool) *Handler {
	return &Handler{
		UserService:  userService,
		IsProduction: isProduction,
	}
}

// validatePassword checks password strength requirements
func validatePassword(password string) error {
	if len(password) < 8 {
		return utils.ErrWeakPassword
	}

	var (
		hasUpper   bool
		hasLower   bool
		hasNumber  bool
		hasSpecial bool
	)

	for _, char := range password {
		switch {
		case unicode.IsUpper(char):
			hasUpper = true
		case unicode.IsLower(char):
			hasLower = true
		case unicode.IsNumber(char):
			hasNumber = true
		case unicode.IsPunct(char) || unicode.IsSymbol(char):
			hasSpecial = true
		}
	}

	if !hasUpper || !hasLower || !hasNumber || !hasSpecial {
		return utils.ErrWeakPassword
	}

	return nil
}

func (h *Handler) SignUp(c echo.Context) error {
	var req models.User
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid input"})
	}

	verifier := emailverifier.NewVerifier()
	ret, err := verifier.Verify(req.Email)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid email address"})
	}
	if !ret.Syntax.Valid {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid email address"})
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Error hashing password"})
	}
	req.Password = string(hashedPassword)

	if err := h.UserService.CreateUser(&req); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusCreated, map[string]string{"message": "User created"})
}

func (h *Handler) Login(c echo.Context) error {
	var req models.User
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, "Invalid input")
	}

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
	cookie.Secure = h.IsProduction // Secure in production, allow HTTP in development
	cookie.SameSite = http.SameSiteStrictMode
	cookie.MaxAge = 86400 // Reduced from 3 days to 1 day

	c.SetCookie(cookie)

	return c.JSON(http.StatusOK, "Successfully logged in")
}

func (h *Handler) Logout(c echo.Context) error {
	jwtCookie := &http.Cookie{
		Name:     "jwt_token",
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		MaxAge:   -1, // MaxAge < 0 deletes the cookie
		Secure:   h.IsProduction,
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
	if token == nil || !token.Valid {
		return c.JSON(http.StatusOK, map[string]any{"authenticated": false})
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return c.JSON(http.StatusOK, map[string]any{"authenticated": false})
	}

	email, ok := claims["sub"].(string)
	if !ok {
		return c.JSON(http.StatusOK, map[string]any{"authenticated": false})
	}

	return c.JSON(http.StatusOK, map[string]any{
		"authenticated": true,
		"email":         email,
	})
}

func (h *Handler) OAuthStatus(c echo.Context) error {
    email, err := h.UserService.IsLoggedIn(c)
    if err != nil {
        return c.JSON(http.StatusUnauthorized, map[string]any{"error": "Invalid token"})
    }

    status, err := h.UserService.GetOAuthLinkStatus(email)
    if err != nil {
        return c.JSON(http.StatusInternalServerError, map[string]any{"error": "Failed to get OAuth status"})
    }

    return c.JSON(http.StatusOK, map[string]any{
        "twitter_linked":    status.Twitter,
        "instagram_linked":  status.Instagram,
        "bluesky_linked":    status.Bluesky,
        "mastodon_linked":   status.Mastodon,
        "artstation_linked": status.Artstation,
        "youtube_linked":    status.Youtube,
    })
}

	
