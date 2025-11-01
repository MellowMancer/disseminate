package instagram

import (
	"backend/models"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"

	repo_cloudflare "backend/repositories/cloudflare"
	repo_supabase "backend/repositories/supabase"
	"mime/multipart"
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
	CheckPublishLimit(accessToken string, instagramID string) (bool, error)
	UploadMedia(file multipart.File) (string, error)
	CreateContainer(accessToken string, instagramID string, caption string, mediaType string, isCarouselItem bool) (string, error)
	CreateCarouselContainer(accessToken string, instagramID string, caption string, containerIDs []string) (string, error)
	ContainerStatus(accessToken string, containerID string) (string, error)
}

type instagramRepositoryImpl struct {
	repo_supabase   *repo_supabase.SupabaseRepository
	repo_cloudflare *repo_cloudflare.CloudflareRepository
}

func NewInstagramRepository(supabaseRepository *repo_supabase.SupabaseRepository, cloudflareRepository *repo_cloudflare.CloudflareRepository) InstagramRepository {
	return &instagramRepositoryImpl{
		repo_supabase:   supabaseRepository,
		repo_cloudflare: cloudflareRepository,
	}
}

func (i *instagramRepositoryImpl) SaveToken(accessToken string, userID string, instagramID string, expirationTime string) error {
	payload := models.InstagramModel{
		UserID:      userID,
		InstagramID: instagramID,
		AccessToken: accessToken,
		ExpiresAt:   expirationTime,
	}

	// Check if a row exists for the userID by fetching credentials
	existingAccessToken, _, err := i.GetCredentials(userID)
	if err != nil && err.Error() != "instagram token not found" {
		return err
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	client := &http.Client{}
	if existingAccessToken == "" {
		// No existing token, create a new one with POST
		url := i.repo_supabase.SupabaseURL + "instagram"
		req, err := i.newRequest("POST", url, bytes.NewBuffer(payloadBytes))
		if err != nil {
			return err
		}

		resp, err := client.Do(req)
		if err != nil {
			return err
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
			body, _ := io.ReadAll(resp.Body)
			return fmt.Errorf("failed to create instagram tokens, status: %d, response: %s", resp.StatusCode, string(body))
		}
	} else {
		// Update existing token with PATCH
		url := i.repo_supabase.SupabaseURL + "instagram?user_id=eq." + userID

		req, err := i.newRequest("PATCH", url, bytes.NewBuffer(payloadBytes))
		if err != nil {
			return err
		}

		resp, err := client.Do(req)
		if err != nil {
			return err
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			body, _ := io.ReadAll(resp.Body)
			return fmt.Errorf("failed to update instagram tokens, status: %d, response: %s", resp.StatusCode, string(body))
		}
	}

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

	req, err := http.NewRequest("GET", i.repo_supabase.SupabaseURL+"instagram", nil)
	if err != nil {
		return "", "", err
	}
	req.Header.Set("apikey", i.repo_supabase.SupabaseKey)
	req.Header.Set("Authorization", "Bearer "+i.repo_supabase.SupabaseKey)

	q := req.URL.Query()
	q.Add("user_id", "eq."+userID)
	req.URL.RawQuery = q.Encode()

	resp, err := http.DefaultClient.Do(req)
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
		return "", "", err
	}

	if len(instagramModel) == 0 {
		return "", "", nil
	}

	if instagramModel[0].AccessToken == "" {
		return "", "", fmt.Errorf("instagram token not found")
	}

	expireTime, err := time.Parse(time.RFC3339, instagramModel[0].ExpiresAt)
	if err != nil {
		expireTime, err = time.Parse("2006-01-02T15:04:05", instagramModel[0].ExpiresAt)
		if err != nil {
			return "", "", fmt.Errorf("invalid expiration time format")
		}
	}

	if !expireTime.IsZero() && time.Now().UTC().After(expireTime) {
		return "", "", fmt.Errorf("instagram token expired")
	}

	if instagramModel[0].InstagramID == "" {
		return "", "", fmt.Errorf("instagram tokens not found")
	}

	return instagramModel[0].AccessToken, instagramModel[0].InstagramID, nil
}

func (i *instagramRepositoryImpl) CheckPublishLimit(accessToken string, instagramID string) (bool, error) {
	log.Println("[CHECK_PUBLISH_LIMIT] --- Starting check for InstagramID:", instagramID)

	req, err := http.NewRequest("GET", instagram_api_path+instagramID+"/content_publishing_limit", nil)
	if err != nil {
		log.Printf("[CHECK_PUBLISH_LIMIT] --- Error creating request: %v", err)
		return false, err
	}

	q := req.URL.Query()
	q.Add("access_token", accessToken)
	q.Add("fields", "quota_usage,config")
	req.URL.RawQuery = q.Encode()
	log.Printf("[CHECK_PUBLISH_LIMIT] --- Request URL: %s", req.URL.String())

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("[CHECK_PUBLISH_LIMIT] --- HTTP request error: %v", err)
		return false, err
	}
	defer resp.Body.Close()

	log.Printf("[CHECK_PUBLISH_LIMIT] --- Response status code: %d", resp.StatusCode)
	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		log.Printf("[CHECK_PUBLISH_LIMIT] --- Response body on error: %s", string(body))
		return false, fmt.Errorf("failed to fetch publishing limit, status %d: %s", resp.StatusCode, string(body))
	}

	body, _ := io.ReadAll(resp.Body)
	log.Printf("[CHECK_PUBLISH_LIMIT] --- Response body: %s", string(body))

	var responseData struct {
		Data []struct {
			QuotaUsage int `json:"quota_usage"`
			Config     struct {
				QuotaTotal    int `json:"quota_total"`
				QuotaDuration int `json:"quota_duration"`
			} `json:"config"`
		} `json:"data"`
	}
	if err := json.Unmarshal(body, &responseData); err != nil {
		log.Printf("[CHECK_PUBLISH_LIMIT] --- JSON unmarshal error: %v", err)
		return false, err
	}

	if len(responseData.Data) == 0 {
		log.Println("[CHECK_PUBLISH_LIMIT] --- No data found in response")
		return false, fmt.Errorf("no data found in publishing limit response")
	}

	result := responseData.Data[0]

	log.Printf("[CHECK_PUBLISH_LIMIT] --- QuotaUsage: %d, QuotaTotal: %d", result.QuotaUsage, result.Config.QuotaTotal)

	if result.QuotaUsage >= result.Config.QuotaTotal {
		log.Println("[CHECK_PUBLISH_LIMIT] --- Publish limit reached or exceeded")
		return false, nil
	}

	log.Println("[CHECK_PUBLISH_LIMIT] --- Publish limit not exceeded - can publish")
	return true, nil
}

func (i *instagramRepositoryImpl) UploadMedia(file multipart.File) (string, error) {
	// Upload media to cloudflare
	mediaURL, err := i.repo_cloudflare.UploadFile(file, fmt.Sprintf("instagram_%d", time.Now().UnixNano()))
	if err != nil {
		return "", fmt.Errorf("failed to upload media to Cloudflare: %w", err)
	}

	return mediaURL, nil
}

func (i *instagramRepositoryImpl) ContainerStatus(accessToken string, containerID string) (string, error) {
	url := instagram_api_path + containerID
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+accessToken)

	q := req.URL.Query()
	q.Add("fields", "status_code")
	req.URL.RawQuery = q.Encode()
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("failed to fetch container status, status %d: %s", resp.StatusCode, string(body))
	}

	var result struct {
		StatusCode string `json:"status_code"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}

	return result.StatusCode, nil
}

func (i *instagramRepositoryImpl) CreateContainer(accessToken string, instagramID string, caption string, mediaType string, isCarouselItem bool) (string, error) {
	url := instagram_api_path + instagramID + "media"
	req, err := http.NewRequest("POST", url, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+accessToken)

	q := req.URL.Query()

	if isCarouselItem {
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

func (i *instagramRepositoryImpl) CreateCarouselContainer(accessToken string, instagramID string, caption string, containerIDs []string) (string, error) {
	url := instagram_api_path + instagramID + "media"
	req, err := http.NewRequest("POST", url, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+accessToken)

	q := req.URL.Query()
	q.Add("caption", caption)
	q.Add("media_type", "CAROUSEL")
	for idx, containerID := range containerIDs {
		q.Add(fmt.Sprintf("children[%d]", idx), containerID)
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
		return "", fmt.Errorf("failed to create carousel container, status %d: %s", resp.StatusCode, string(body))
	}

	var result struct {
		ID string `json:"id"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}

	return result.ID, nil
}

func (i *instagramRepositoryImpl) newRequest(method, url string, body io.Reader) (*http.Request, error) {
	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("apikey", i.repo_supabase.SupabaseKey)
	req.Header.Set("Authorization", "Bearer "+i.repo_supabase.SupabaseKey)

	if method != "GET" {
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Prefer", "return=representation")
	}
	return req, nil
}
