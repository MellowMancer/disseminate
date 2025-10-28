package user

import (
	"backend/models"
	repo_supabase "backend/repositories/supabase"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	_ "log"
	"net/http"
	_ "net/http/httputil"
	"net/url"
)

const profile_path = "profiles"

type UserIDResponse struct {
	ID string `json:"id"`
}

type UserRepository interface {
	Create(user *models.User) error
	FindByEmail(email string) (*models.User, error)
	ExistsByEmail(email string) (bool, error)
	UserIDByEmail(email string) (string, error)
}

type userRepositoryImpl struct {
	repo_supabase *repo_supabase.SupabaseRepository
}

func NewUserRepository(supabaseRepository *repo_supabase.SupabaseRepository) UserRepository {
	return &userRepositoryImpl{
		repo_supabase: supabaseRepository,
	}
}

func (u *userRepositoryImpl) Create(user *models.User) error {
	supabaseUserData := models.User{
		Email:    user.Email,
		Password: user.Password,
	}
	data, err := json.Marshal(supabaseUserData)
	if err != nil {
		return err
	}
	
	url := u.repo_supabase.SupabaseURL + profile_path
	req, err := u.newRequest("POST", url, bytes.NewBuffer(data))

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

func (u *userRepositoryImpl) FindByEmail(email string) (*models.User, error) {
	url := u.repo_supabase.SupabaseURL + profile_path + "?email=eq." + url.QueryEscape(email)
	req, err := u.newRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("login failed, status: %d, response: %s", resp.StatusCode, string(body))
	}

	var users []models.User
	if err := json.NewDecoder(resp.Body).Decode(&users); err != nil {
		return nil, err
	}
	if len(users) == 0 {
		return nil, fmt.Errorf("user not found")
	}

	return &users[0], nil
}

func (u *userRepositoryImpl) ExistsByEmail(email string) (bool, error) {
	req, err := u.newRequest("GET", u.repo_supabase.SupabaseURL+profile_path+"?email=eq."+email, nil)
	if err != nil {
		return false, err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()

	var users []models.User
	json.NewDecoder(resp.Body).Decode(&users)

	return len(users) > 0, nil
}

func (u *userRepositoryImpl) UserIDByEmail(email string) (string, error) {
	url := u.repo_supabase.SupabaseURL + profile_path + "?email=eq." + url.QueryEscape(email) + "&select=id"

	req, err := u.newRequest("GET", url, nil)
	if err != nil {
		return "", err
	}

	resp, err := http.DefaultClient.Do(req)
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




func (u *userRepositoryImpl) newRequest(method, url string, body io.Reader) (*http.Request, error) {
    req, err := http.NewRequest(method, url, body)
    if err != nil {
        return nil, err
    }
    req.Header.Set("apikey", u.repo_supabase.SupabaseKey)
    req.Header.Set("Authorization", "Bearer "+u.repo_supabase.SupabaseKey)

	if(method != "GET") {
    req.Header.Set("Content-Type", "application/json")
    req.Header.Set("Prefer", "return=representation")
	}
    return req, nil
}
