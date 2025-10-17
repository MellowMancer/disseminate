package services

import (
	"backend/models"
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"
	"golang.org/x/crypto/bcrypt"
	"io"
	_ "log"
	"net/http"
	_ "net/http/httputil"
	"net/url"
	"time"
)

const api_path = "/rest/v1/profiles?email=eq."

type UserService interface {
	CreateUser(user *models.User) error
	LoginUser(user *models.User) (string, error)
	GetJWTSecret() []byte
	LinkTwitterAccount(email string, accessToken string, accessSecret string) error
	GetTwitterTokens(email string) (string, string, error)
	IsLoggedIn(c echo.Context) (bool, string, error)
	LinkInstagramAccount(email string, accessToken string, expiresIn int) error
    GetInstagramTokens(email string) (string, error)
}

type userServiceImpl struct {
	supabaseURL string
	supabaseKey string
	jwtSecret   []byte
	httpClient  *http.Client
}

func NewUserService(url, key string, jwtSecret []byte) UserService {
	return &userServiceImpl{
		supabaseURL: url,
		supabaseKey: key,
		jwtSecret:   []byte(jwtSecret),
		httpClient:  &http.Client{},
	}
}

func (s *userServiceImpl) CreateUser(user *models.User) error {
	exists, err := s.UserExists(user.Email)
	if err != nil {
		return err
	}
	if exists {
		return fmt.Errorf("user already exists")
	}

	payloadBytes, err := json.Marshal(user)
	if err != nil {
		return fmt.Errorf("error marshalling user: %w", err)
	}

	url := s.supabaseURL + api_path
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(payloadBytes))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("apikey", s.supabaseKey)
	req.Header.Set("Authorization", "Bearer "+s.supabaseKey)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Prefer", "return=representation")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("request error: %w", err)
	}
	defer resp.Body.Close()

	bodyBytes, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusCreated {
		return fmt.Errorf("signup failed. Status: %d, Response: %s", resp.StatusCode, string(bodyBytes))
	}
	return nil
}

func (s *userServiceImpl) LoginUser(user *models.User) (string, error) {
	url := s.supabaseURL + api_path + url.QueryEscape(user.Email)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", err
	}

	req.Header.Set("apikey", s.supabaseKey)
	req.Header.Set("Authorization", "Bearer "+s.supabaseKey)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("login failed, status: %d, response: %s", resp.StatusCode, string(body))
	}

	var users []models.User
	if err := json.NewDecoder(resp.Body).Decode(&users); err != nil {
		return "", err
	}
	if len(users) == 0 {
		return "", fmt.Errorf("user not found")
	}

	err = bcrypt.CompareHashAndPassword([]byte(users[0].Password), []byte(user.Password))
	if err != nil {
		return "", fmt.Errorf("invalid password")
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub": users[0].Email,
		"exp": time.Now().Add(72 * time.Hour).Unix(),
	})

	tokenString, err := token.SignedString([]byte(s.jwtSecret))
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

func (s *userServiceImpl) UserExists(email string) (bool, error) {
	req, _ := http.NewRequest("GET", s.supabaseURL+api_path+email, nil)
	req.Header.Set("apikey", s.supabaseKey)
	req.Header.Set("Authorization", "Bearer "+s.supabaseKey)

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()

	var users []models.User
	json.NewDecoder(resp.Body).Decode(&users)

	return len(users) > 0, nil
}

func (s *userServiceImpl) GetJWTSecret() []byte {
	return s.jwtSecret
}

func (s *userServiceImpl) LinkTwitterAccount(email, accessToken, accessSecret string) error {
	payload := map[string]string{
		"twitter_access_token":  accessToken,
		"twitter_access_secret": accessSecret,
	}
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	url := s.supabaseURL + api_path + url.QueryEscape(email)

	req, err := http.NewRequest("PATCH", url, bytes.NewBuffer(payloadBytes))
	if err != nil {
		return err
	}

	req.Header.Set("apikey", s.supabaseKey)
	req.Header.Set("Authorization", "Bearer "+s.supabaseKey)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Prefer", "return=representation")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed to update twitter tokens, status: %d, response: %s", resp.StatusCode, string(body))
	}

	return nil
}

func (s *userServiceImpl) GetTwitterTokens(email string) (string, string, error) {
	req, _ := http.NewRequest("GET", s.supabaseURL+api_path+email, nil)
	req.Header.Set("apikey", s.supabaseKey)
	req.Header.Set("Authorization", "Bearer "+s.supabaseKey)
	resp, err := s.httpClient.Do(req)
	if err != nil {
		return "", "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", "", fmt.Errorf("failed to fetch user, status: %d, response: %s", resp.StatusCode, string(body))
	}
	var users []models.User
	json.NewDecoder(resp.Body).Decode(&users)
	if len(users) == 0 || users[0].TwitterAccessToken == nil || users[0].TwitterAccessSecret == nil {
		return "", "", fmt.Errorf("twitter tokens not found")
	}
	return *users[0].TwitterAccessToken, *users[0].TwitterAccessSecret, nil
}

func (s *userServiceImpl) LinkInstagramAccount(email string, instagramAccessToken string, expiresIn int) error {
	expirationTime := time.Now().Add(time.Duration(expiresIn) * time.Second).UTC()

	expirationTimeStr := expirationTime.Format(time.RFC3339)

	payload := map[string]string{
		"instagram_access_token":     instagramAccessToken,
		"instagram_token_expires_at": expirationTimeStr,
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	url := s.supabaseURL + api_path + url.QueryEscape(email)

	req, err := http.NewRequest("PATCH", url, bytes.NewBuffer(payloadBytes))
	if err != nil {
		return err
	}

	req.Header.Set("apikey", s.supabaseKey)
	req.Header.Set("Authorization", "Bearer "+s.supabaseKey)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Prefer", "return=representation")
	resp, err := s.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed to update instagram tokens, status: %d, response: %s", resp.StatusCode, string(body))
	}

	return nil
}

func (s *userServiceImpl) GetInstagramTokens(email string) (string, error) {
	req, _ := http.NewRequest("GET", s.supabaseURL+api_path+email, nil)
	req.Header.Set("apikey", s.supabaseKey)
	req.Header.Set("Authorization", "Bearer "+s.supabaseKey)
	resp, err := s.httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("failed to fetch user, status: %d, response: %s", resp.StatusCode, string(body))
	}
	var users []models.User
	json.NewDecoder(resp.Body).Decode(&users)
	if len(users) == 0 || users[0].InstagramAccessToken == nil {
		return "", fmt.Errorf("instagram token not found")
	}

	if users[0].InstagramTokenExpiresAt != nil {
		expired := false
		expireTime := *users[0].InstagramTokenExpiresAt
		if time.Now().UTC().After(expireTime) {
			expired = true
		}
		if expired {
			return "", fmt.Errorf("instagram token expired")
		}
	}

	return *users[0].InstagramAccessToken, nil
}

func (s *userServiceImpl) IsLoggedIn(c echo.Context) (bool, string, error) {
	claims, ok := c.Get("userClaims").(jwt.MapClaims)
	if !ok {
		return false, "", c.JSON(http.StatusUnauthorized, map[string]string{"error": "Invalid token claims"})
	}

	email, ok := claims["sub"].(string)
	if !ok {
		return false, "", c.JSON(http.StatusBadRequest, map[string]string{"error": "Email (sub) claim not found in token"})
	}

	exists, err := s.UserExists(email)
	if err != nil || !exists {
		return false, "", err
	}
	return true, email, nil
}
