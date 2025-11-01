package instagram

import (
	"backend/models"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	_ "log"

	// "mime/multipart"
	repo_supabase "backend/repositories/supabase"
	"net/http"
	_ "net/http/httputil"
	"time"
)

const instagram_api_path = "https://graph.instagram.com/v24.0/"
const longTimeTokenURL = "https://graph.instagram.com/access_token"

type InstagramRepository interface {
	SaveToken(accessToken string, userID string, instagramID string, expirationTime string) error
	GetInstagramID(accessToken string) (string, error)
	GetAccessToken(accessToken string, clientSecret string) (string, int, error)
	CheckTokens(accessToken string) error
	GetCredentials(userID string) (string, string, error)
	// CheckPublishLimit(accessToken string, instagramID string) (bool, error)
	// UploadMedia(accessToken string, file multipart.File) (string, error)
	// CreateContainer(accessToken string, instagramID string, caption string, mediaType string, isCarouselItem bool) (string, error)
	// CreateCarouselContainer(accessToken string, instagramID string, caption string, containerIDs []string) (string, error)
	// ContainerStatus(accessToken string, containerID string) (string, error)
}

type instagramRepositoryImpl struct {
	repo_supabase *repo_supabase.SupabaseRepository
}

func NewInstagramRepository(supabaseRepository *repo_supabase.SupabaseRepository) InstagramRepository {
	return &instagramRepositoryImpl{
		repo_supabase: supabaseRepository,
	}
}

func (i *instagramRepositoryImpl) SaveToken(accessToken string, userID string, instagramID string, expirationTime string) error {
	log.Printf("[SAVE_TOKEN] --- Start saving token for userID: %s ---", userID)

	payload := models.InstagramModel{
		UserID:      userID,
		InstagramID: instagramID,
		AccessToken: accessToken,
		ExpiresAt:   expirationTime,
	}

	// Check if a row exists for the userID by fetching credentials
	existingAccessToken, _, err := i.GetCredentials(userID)
	if err != nil && err.Error() != "instagram token not found" {
		log.Printf("[SAVE_TOKEN] --- Error fetching existing credentials: %v", err)
		return err
	}

	log.Printf("[SAVE_TOKEN] --- Existing access token found: %t", existingAccessToken != "")

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		log.Printf("[SAVE_TOKEN] --- Failed to marshal payload: %v", err)
		return err
	}
	log.Printf("[SAVE_TOKEN] --- Payload JSON: %s", string(payloadBytes))

	client := &http.Client{}
	if existingAccessToken == "" {
		// No existing token, create a new one with POST
		url := i.repo_supabase.SupabaseURL + "instagram"
		log.Printf("[SAVE_TOKEN] --- Creating new token with POST to URL: %s", url)

		req, err := i.newRequest("POST", url, bytes.NewBuffer(payloadBytes))
		if err != nil {
			log.Printf("[SAVE_TOKEN] --- Error creating POST request: %v", err)
			return err
		}

		resp, err := client.Do(req)
		if err != nil {
			log.Printf("[SAVE_TOKEN] --- HTTP POST request error: %v", err)
			return err
		}
		defer resp.Body.Close()

		log.Printf("[SAVE_TOKEN] --- HTTP POST response status: %d", resp.StatusCode)
		if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
			body, _ := io.ReadAll(resp.Body)
			log.Printf("[SAVE_TOKEN] --- POST response body on failure: %s", string(body))
			return fmt.Errorf("failed to create instagram tokens, status: %d, response: %s", resp.StatusCode, string(body))
		}
	} else {
		// Update existing token with PATCH
		url := i.repo_supabase.SupabaseURL + "instagram?user_id=eq." + userID
		log.Printf("[SAVE_TOKEN] --- Updating existing token with PATCH to URL: %s", url)

		req, err := i.newRequest("PATCH", url, bytes.NewBuffer(payloadBytes))
		if err != nil {
			log.Printf("[SAVE_TOKEN] --- Error creating PATCH request: %v", err)
			return err
		}

		resp, err := client.Do(req)
		if err != nil {
			log.Printf("[SAVE_TOKEN] --- HTTP PATCH request error: %v", err)
			return err
		}
		defer resp.Body.Close()

		log.Printf("[SAVE_TOKEN] --- HTTP PATCH response status: %d", resp.StatusCode)
		if resp.StatusCode != http.StatusOK {
			body, _ := io.ReadAll(resp.Body)
			log.Printf("[SAVE_TOKEN] --- PATCH response body on failure: %s", string(body))
			return fmt.Errorf("failed to update instagram tokens, status: %d, response: %s", resp.StatusCode, string(body))
		}
	}

	log.Println("[SAVE_TOKEN] --- Token save operation completed successfully ---")
	return nil
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

func (i *instagramRepositoryImpl) GetAccessToken(accessToken string, clientSecret string) (string, int, error) {
	req, err := http.NewRequest("GET", longTimeTokenURL, nil)
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

	fmt.Printf("[INSTAGRAM_REPOSITORY] --- Exchanging for long-term token ---")

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
	log.Println("[GET_CREDENTIALS] --- Start fetching credentials for userID:", userID)

	req, err := http.NewRequest("GET", i.repo_supabase.SupabaseURL+"instagram", nil)
	if err != nil {
		log.Printf("[GET_CREDENTIALS] --- Failed to create request: %v", err)
		return "", "", err
	}
	req.Header.Set("apikey", i.repo_supabase.SupabaseKey)
	req.Header.Set("Authorization", "Bearer "+i.repo_supabase.SupabaseKey)

	q := req.URL.Query()
	q.Add("user_id", "eq."+userID)
	req.URL.RawQuery = q.Encode()
	log.Printf("[GET_CREDENTIALS] --- Request URL: %s", req.URL.String())

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Printf("[GET_CREDENTIALS] --- Request error: %v", err)
		return "", "", err
	}
	defer resp.Body.Close()

	log.Printf("[GET_CREDENTIALS] --- Response status code: %d", resp.StatusCode)
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		log.Printf("[GET_CREDENTIALS] --- Response body on error: %s", string(body))
		return "", "", fmt.Errorf("failed to fetch user, status: %d, response: %s", resp.StatusCode, string(body))
	}

	var instagramModel []models.InstagramModel
	if err := json.NewDecoder(resp.Body).Decode(&instagramModel); err != nil {
		log.Printf("[GET_CREDENTIALS] --- JSON decode error: %v", err)
		return "", "", err
	}

	log.Printf("[GET_CREDENTIALS] --- Number of records fetched: %d", len(instagramModel))
	if len(instagramModel) == 0 {
		log.Println("[GET_CREDENTIALS] --- No records found for userID")
		return "", "", nil
	}

	if instagramModel[0].AccessToken == "" {
		log.Println("[GET_CREDENTIALS] --- Access token missing in fetched record")
		return "", "", fmt.Errorf("instagram token not found")
	}

	expireTime, err := time.Parse(time.RFC3339, instagramModel[0].ExpiresAt)
	if err != nil {
		expireTime, err = time.Parse("2006-01-02T15:04:05", instagramModel[0].ExpiresAt)
		if err != nil {
			log.Printf("[GET_CREDENTIALS] --- Invalid expiration time format: %v", err)
			return "", "", fmt.Errorf("invalid expiration time format")
		}
	}

	log.Printf("[GET_CREDENTIALS] --- Parsed expiration time: %s", expireTime.String())

	if !expireTime.IsZero() && time.Now().UTC().After(expireTime) {
		log.Println("[GET_CREDENTIALS] --- Instagram token expired")
		return "", "", fmt.Errorf("instagram token expired")
	}

	if instagramModel[0].InstagramID == "" {
		log.Println("[GET_CREDENTIALS] --- Instagram ID missing in fetched record")
		return "", "", fmt.Errorf("instagram tokens not found")
	}

	log.Println("[GET_CREDENTIALS] --- Successfully retrieved Instagram credentials")
	return instagramModel[0].AccessToken, instagramModel[0].InstagramID, nil
}

func (i *instagramRepositoryImpl) CheckPublishLimit(accessToken string, instagramID string) (bool, error) {
	req, err := http.NewRequest("GET", instagram_api_path+instagramID+"/content_publishing_limit", nil)
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
		QuotaUsage int `json:"quota_usage"`
		Config     struct {
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

// func (i *instagramRepositoryImpl) UploadMedia(accessToken string, file multipart.File) (string, error) {
// 	// Upload media to cloudflare
// }

// func (i *instagramRepositoryImpl) ContainerStatus(accessToken string, containerID string) (string, error) {
// 	url := instagram_api_path + containerID
// 	req, err := http.NewRequest("GET", url, nil)
// 	if err != nil {
// 		return "", err
// 	}
// 	req.Header.Set("Content-Type", "application/json")
// 	req.Header.Set("Authorization", "Bearer "+accessToken)

// 	q := req.URL.Query()
// 	q.Add("fields", "status_code")
// 	req.URL.RawQuery = q.Encode()
// 	client := &http.Client{}
// 	resp, err := client.Do(req)
// 	if err != nil {
// 		return "", err
// 	}
// 	defer resp.Body.Close()

// 	if resp.StatusCode != 200 {
// 		body, _ := io.ReadAll(resp.Body)
// 		return "", fmt.Errorf("failed to fetch container status, status %d: %s", resp.StatusCode, string(body))
// 	}

// 	var result struct {
// 		StatusCode string `json:"status_code"`
// 	}
// 	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
// 		return "", err
// 	}

// 	return result.StatusCode, nil
// }

// func (i *instagramRepositoryImpl) CreateContainer(accessToken string, instagramID string, caption string, mediaType string, isCarouselItem bool) (string, error) {
// 	url := instagram_api_path + instagramID + "media"
// 	req, err := http.NewRequest("POST", url, nil)
// 	if err != nil {
// 		return "", err
// 	}
// 	req.Header.Set("Content-Type", "application/json")
// 	req.Header.Set("Authorization", "Bearer "+accessToken)

// 	q := req.URL.Query()

// 	if(isCarouselItem) {
// 		q.Add("is_carousel_item", "true")
// 	} else {
// 		q.Add("caption", caption)
// 	}
// 	req.URL.RawQuery = q.Encode()
// 	client := &http.Client{}
// 	resp, err := client.Do(req)
// 	if err != nil {
// 		return "", err
// 	}
// 	defer resp.Body.Close()

// 	if resp.StatusCode != 200 {
// 		body, _ := io.ReadAll(resp.Body)
// 		return "", fmt.Errorf("failed to upload media, status %d: %s", resp.StatusCode, string(body))
// 	}

// 	var result struct {
// 		ID string `json:"id"`
// 	}
// 	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
// 		return "", err
// 	}

// 	return result.ID, nil
// }

// func (i *instagramRepositoryImpl) CreateCarouselContainer(accessToken string, instagramID string, caption string, containerIDs []string) (string, error) {
// 	url := instagram_api_path + instagramID + "media"
// 	req, err := http.NewRequest("POST", url, nil)
// 	if err != nil {
// 		return "", err
// 	}
// 	req.Header.Set("Content-Type", "application/json")
// 	req.Header.Set("Authorization", "Bearer "+accessToken)

// 	q := req.URL.Query()
// 	q.Add("caption", caption)
// 	q.Add("media_type", "CAROUSEL")
// 	for idx, containerID := range containerIDs {
// 		q.Add(fmt.Sprintf("children[%d]", idx), containerID)
// 	}
// 	req.URL.RawQuery = q.Encode()
// 	client := &http.Client{}
// 	resp, err := client.Do(req)
// 	if err != nil {
// 		return "", err
// 	}
// 	defer resp.Body.Close()

// 	if resp.StatusCode != 200 {
// 		body, _ := io.ReadAll(resp.Body)
// 		return "", fmt.Errorf("failed to create carousel container, status %d: %s", resp.StatusCode, string(body))
// 	}

// 	var result struct {
// 		ID string `json:"id"`
// 	}
// 	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
// 		return "", err
// 	}

// 	return result.ID, nil
// }


func (i *instagramRepositoryImpl) newRequest(method, url string, body io.Reader) (*http.Request, error) {
    req, err := http.NewRequest(method, url, body)
    if err != nil {
        return nil, err
    }
    req.Header.Set("apikey", i.repo_supabase.SupabaseKey)
    req.Header.Set("Authorization", "Bearer "+i.repo_supabase.SupabaseKey)

	if(method != "GET") {
    req.Header.Set("Content-Type", "application/json")
    req.Header.Set("Prefer", "return=representation")
	}
    return req, nil
}