package instagram

import (
	"backend/models"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	_ "log"
	"mime/multipart"
	"net/http"
	_ "net/http/httputil"
	"time"
	repo_supabase "backend/repositories/supabase"
)

const instagram_api_path = "https://graph.instagram.com/v24.0/"

type InstagramRepository interface {
	SaveToken(userID string, instagramID string, accessToken string, expirationTime time.Time) error
	GetToken(userID string) (string, error)
	GetInstagramID(accessToken string) (string, error)
	GetAccessToken(URL string, accessToken string, clientSecret string) (string, int, error)
	CheckTokens(accessToken string) error
	GetCredentials(userID string) (string, string, error)
	CheckPublishLimit(instagramID string, accessToken string) (bool, error)
	UploadMedia(accessToken string, file multipart.File) (string, error)
	CreateContainer(accessToken string, instagramID string, caption string, isCarouselItem bool) (string, error)
}

type instagramRepositoryImpl struct {
	repo_supabase *repo_supabase.SupabaseRepository
}

func NewInstagramRepository(supabaseRepository *repo_supabase.SupabaseRepository) InstagramRepository {
	return &instagramRepositoryImpl{
		repo_supabase: supabaseRepository,
	}
}

func (i *instagramRepositoryImpl) SaveToken(userID string, instagramID string, accessToken string, expirationTime time.Time) error {
	payload := models.InstagramModel{
		UserID:      userID,
		InstagramID: instagramID,
		AccessToken: accessToken,
		ExpiresAt:   expirationTime,
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	url := i.repo_supabase.SupabaseURL + "instagram"

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(payloadBytes))
	if err != nil {
		return err
	}

	req.Header.Set("apikey", i.repo_supabase.SupabaseKey)
	req.Header.Set("Authorization", "Bearer "+i.repo_supabase.SupabaseKey)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Prefer", "return=representation")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed to update instagram tokens, status: %d, response: %s", resp.StatusCode, string(body))
	}

	return nil
}

func (i *instagramRepositoryImpl) GetToken(userID string) (string, error) {
	req, _ := http.NewRequest("GET", i.repo_supabase.SupabaseURL + "instagram", nil)
	req.Header.Set("apikey", i.repo_supabase.SupabaseKey)
	req.Header.Set("Authorization", "Bearer "+i.repo_supabase.SupabaseKey)

	q := req.URL.Query()
	q.Add("user_id", "eq."+userID)
	req.URL.RawQuery = q.Encode()

	resp, err := http.DefaultClient.Do(req)
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

func (i *instagramRepositoryImpl) GetInstagramID(accessToken string) (string, error) {
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

func (i *instagramRepositoryImpl) GetAccessToken(URL string, accessToken string, clientSecret string) (string, int, error) {
	req, err := http.NewRequest("GET", URL, nil)
    if err != nil {
        return "", 0, err
    }

    q := req.URL.Query()
    q.Add("grant_type", "ig_exchange_token")
    q.Add("client_secret", clientSecret)
    q.Add("access_token", accessToken)
    req.URL.RawQuery = q.Encode()

    resp, err := http.DefaultClient.Do(req)
    if err != nil {
        return "", 0, err
    }
    defer resp.Body.Close()

    if resp.StatusCode != http.StatusOK {
        body, _ := io.ReadAll(resp.Body)
        return "", 0, fmt.Errorf("instagram token exchange failed: status %d: %s", resp.StatusCode, string(body))
    }

    var longToken struct {
        AccessToken string `json:"access_token"`
        TokenType   string `json:"token_type"`
        ExpiresIn   int    `json:"expires_in"`
    }

    if err = json.NewDecoder(resp.Body).Decode(&longToken); err != nil {
        return "", 0, err
    }

    return longToken.AccessToken, longToken.ExpiresIn, nil
}

func (i *instagramRepositoryImpl) CheckTokens(accessToken string) error {
	resp, err := http.Get("https://graph.instagram.com/me?access_token=" + accessToken)
    if err != nil || resp.StatusCode != 200 {
        return fmt.Errorf("tokens have been revoked, please connect your account again")
    }
	return nil
}

func (i *instagramRepositoryImpl) GetCredentials(userID string) (string, string, error) {
	req, _ := http.NewRequest("GET", i.repo_supabase.SupabaseURL+"instagram", nil)
	req.Header.Set("apikey", i.repo_supabase.SupabaseKey)
	req.Header.Set("Authorization", "Bearer "+i.repo_supabase.SupabaseKey)

	q := req.URL.Query()
	q.Add("user_id", "eq."+userID)
	req.URL.RawQuery = q.Encode()

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", "", fmt.Errorf("failed to fetch user, status: %d, response: %s", resp.StatusCode, string(body))
	}

	var instagramModel []models.InstagramModel
	if err := json.NewDecoder(resp.Body).Decode(&instagramModel); err != nil {
		return "", "", fmt.Errorf("failed to decode response: %v", err)
	}

	if len(instagramModel) == 0 || instagramModel[0].AccessToken == "" || instagramModel[0].InstagramID == "" {
		return "", "", fmt.Errorf("twitter tokens not found")
	}

	return instagramModel[0].AccessToken, instagramModel[0].InstagramID, nil
}

func (i *instagramRepositoryImpl) CheckPublishLimit(instagramID string, accessToken string) (bool, error) {
	req, err := http.NewRequest("GET", instagram_api_path + instagramID + "/content_publishing_limit", nil)
	if err != nil {
		return false, err
	}

	q, err := req.URL.Query(), error(nil)
	q.Add("access_token", accessToken)
	q.Add("fields", "quota_usage,rate_limit_settings")
	req.URL.RawQuery = q.Encode()
	
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		return false, fmt.Errorf("failed to fetch publishing limit, status %d: %s", resp.StatusCode, string(body))
	}
	var result struct {
		QuotaUsage        int `json:"quota_usage"`
		Config struct {
			QuotaTotal int `json:"quota_total"`
		} `json:"config"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return false, err
	}

	if result.QuotaUsage >= result.Config.QuotaTotal {
		return false, nil
	}

	return true, nil
}

func (i *instagramRepositoryImpl) UploadMedia(accessToken string, file multipart.File) (string, error) {
	
}

func (i *instagramRepositoryImpl) CreateContainer(accessToken string, instagramID string, caption string, isCarouselItem bool) (string, error) {
	url := instagram_api_path + instagramID + "media"
	req, err := http.NewRequest("POST", url, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+accessToken)

	q := req.URL.Query()

	if(isCarouselItem) {
		q.Add("is_carousel_item", "true")
	} else {
		q.Add("caption", caption)
	}
	req.URL.RawQuery = q.Encode()
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("failed to upload media, status %d: %s", resp.StatusCode, string(body))
	}

	var result struct {
		ID string `json:"id"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}

	return result.ID, nil
}