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

const api_path = "/rest/v1/"
const profile_path = "profiles"

type UserIDResponse struct {
	ID string `json:"id"`
}

type UserService interface {
	CreateUser(user *models.User) error
	LoginUser(user *models.User) (string, error)
	GetJWTSecret() []byte
	LinkTwitterAccount(email string, accessToken string, accessSecret string) error
	GetTwitterTokens(email string) (string, string, error)
	IsLoggedIn(c echo.Context) (bool, string, error)
	LinkInstagramAccount(email string, accessToken string, expiresIn int) error
	GetInstagramTokens(email string) (string, error)
	GetUserID(email string) (string, error)
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

	url := s.supabaseURL + api_path + profile_path
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
	url := s.supabaseURL + api_path + profile_path + "?email=eq." + url.QueryEscape(user.Email)
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
	req, _ := http.NewRequest("GET", s.supabaseURL+api_path+profile_path+"?email=eq."+email, nil)
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
	userID, err := s.GetUserID(email)
	if err != nil {
		return err
	}

	payload := map[string]string{
		"user_id":       userID,
		"access_token":  accessToken,
		"access_secret": accessSecret,
	}
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	url := s.supabaseURL + api_path + "twitter"

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(payloadBytes))
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
		return fmt.Errorf("failed to save twitter tokens, status: %d, response: %s", resp.StatusCode, string(body))
	}

	return nil
}

func (s *userServiceImpl) GetTwitterTokens(email string) (string, string, error) {
	fmt.Printf("[GetTwitterTokens] Started for email: %s\n", email)

	userID, err := s.GetUserID(email)
	if err != nil {
		fmt.Printf("[GetTwitterTokens] Failed to get user ID for email %s: %v\n", email, err)
		return "", "", err
	}
	fmt.Printf("[GetTwitterTokens] Retrieved user ID: %s\n", userID)

	req, _ := http.NewRequest("GET", s.supabaseURL+api_path+"twitter", nil)
	req.Header.Set("apikey", s.supabaseKey)
	req.Header.Set("Authorization", "Bearer "+s.supabaseKey)

	q := req.URL.Query()
	q.Add("user_id", "eq."+userID)
	req.URL.RawQuery = q.Encode()

	fmt.Printf("[GetTwitterTokens] Request URL with query: %s\n", req.URL.String())

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("[GetTwitterTokens] HTTP request failed: %v\n", err)
		return "", "", err
	}
	defer resp.Body.Close()

	fmt.Printf("[GetTwitterTokens] Received response with status: %d\n", resp.StatusCode)

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		fmt.Printf("[GetTwitterTokens] Error response body: %s\n", string(body))
		return "", "", fmt.Errorf("failed to fetch user, status: %d, response: %s", resp.StatusCode, string(body))
	}

	var twitterModel []models.TwitterModel
	if err := json.NewDecoder(resp.Body).Decode(&twitterModel); err != nil {
		fmt.Printf("[GetTwitterTokens] Failed to decode response: %v\n", err)
		return "", "", fmt.Errorf("failed to decode response: %v", err)
	}
	fmt.Printf("[GetTwitterTokens] Decoded twitter model count: %d\n", len(twitterModel))

	if len(twitterModel) == 0 || twitterModel[0].AccessToken == "" || twitterModel[0].AccessSecret == "" {
		fmt.Printf("[GetTwitterTokens] Twitter tokens not found in response\n")
		return "", "", fmt.Errorf("twitter tokens not found")
	}

	fmt.Printf("[GetTwitterTokens] Successfully retrieved twitter tokens for user ID: %s\n", userID)
	return twitterModel[0].AccessToken, twitterModel[0].AccessSecret, nil
}

func (s *userServiceImpl) LinkInstagramAccount(email string, accessToken string, expiresIn int) error {
	fmt.Printf("[LinkInstagramAccount] Starting linking process for email: %s\n", email)

	userID, err := s.GetUserID(email)
	if err != nil {
		fmt.Printf("[LinkInstagramAccount] Error getting user ID for email %s: %v\n", email, err)
		return err
	}
	fmt.Printf("[LinkInstagramAccount] Retrieved user ID: %s\n", userID)

	expirationTime := time.Now().Add(time.Duration(expiresIn) * time.Second).UTC()
	fmt.Printf("[LinkInstagramAccount] Calculated token expiration time: %s\n", expirationTime.Format(time.RFC3339))

	instagramID, err := GetInstagramUserProfile(accessToken)
	if err != nil {
		fmt.Printf("[LinkInstagramAccount] Failed to get Instagram profile for token: %v\n", err)
		return err
	}
	fmt.Printf("[LinkInstagramAccount] Retrieved Instagram ID: %s\n", instagramID)

	payload := models.InstagramModel{
		UserID:      userID,
		InstagramID: instagramID,
		AccessToken: accessToken,
		ExpiresAt:   expirationTime,
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		fmt.Printf("[LinkInstagramAccount] Failed to marshal payload: %v\n", err)
		return err
	}
	fmt.Printf("[LinkInstagramAccount] Payload marshaled successfully")

	url := s.supabaseURL + api_path + "instagram"
	fmt.Printf("[LinkInstagramAccount] Sending POST request to URL: %s\n", url)

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(payloadBytes))
	if err != nil {
		fmt.Printf("[LinkInstagramAccount] Failed to create HTTP request: %v\n", err)
		return err
	}

	req.Header.Set("apikey", s.supabaseKey)
	req.Header.Set("Authorization", "Bearer "+s.supabaseKey)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Prefer", "return=representation")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		fmt.Printf("[LinkInstagramAccount] HTTP request failed: %v\n", err)
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		fmt.Printf("[LinkInstagramAccount] Request failed with status %d: %s\n", resp.StatusCode, string(body))
		return fmt.Errorf("failed to update instagram tokens, status: %d, response: %s", resp.StatusCode, string(body))
	}

	fmt.Println("[LinkInstagramAccount] Instagram tokens linked successfully")

	return nil
}

func GetInstagramUserProfile(accessToken string) (string, error) {
	url := "https://graph.instagram.com/me"
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", err
	}

	q := req.URL.Query()
	q.Add("fields", "id")
	q.Add("access_token", accessToken)

	req.URL.RawQuery = q.Encode()

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("failed to fetch user profile, status %d: %s", resp.StatusCode, string(body))
	}

	var result struct {
		ID string `json:"id"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}

	return result.ID, nil
}

func (s *userServiceImpl) GetInstagramTokens(email string) (string, error) {

	userID, err := s.GetUserID(email)
	if err != nil {
		return "", err
	}

	req, _ := http.NewRequest("GET", s.supabaseURL+api_path+"instagram", nil)
	req.Header.Set("apikey", s.supabaseKey)
	req.Header.Set("Authorization", "Bearer "+s.supabaseKey)

	q := req.URL.Query()
	q.Add("user_id", "eq."+userID)
	req.URL.RawQuery = q.Encode()

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("failed to fetch user, status: %d, response: %s", resp.StatusCode, string(body))
	}
	var instagramModel []models.InstagramModel
	json.NewDecoder(resp.Body).Decode(&instagramModel)
	if len(instagramModel) == 0 || instagramModel[0].AccessToken == "" {
		return "", fmt.Errorf("instagram token not found")
	}

	if !instagramModel[0].ExpiresAt.IsZero() {
		expireTime := instagramModel[0].ExpiresAt
		if time.Now().UTC().After(expireTime) {
			return "", fmt.Errorf("instagram token expired")
		}
	}

	return instagramModel[0].AccessToken, nil
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

func (s *userServiceImpl) GetUserID(email string) (string, error) {
	url := s.supabaseURL + api_path + profile_path + "?email=eq." + url.QueryEscape(email) + "&select=id"

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", err
	}

	req.Header.Set("apikey", s.supabaseKey)
	req.Header.Set("Authorization", "Bearer "+s.supabaseKey)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Prefer", "return=representation")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("failed to update get user_id, status: %d, response: %s", resp.StatusCode, string(body))
	}

	var users []UserIDResponse
	if err := json.NewDecoder(resp.Body).Decode(&users); err != nil {
		return "", fmt.Errorf("failed to decode user id response: %w", err)
	}
	if len(users) == 0 {
		return "", fmt.Errorf("no user found for email")
	}

	userID := users[0].ID

	return userID, nil
}
