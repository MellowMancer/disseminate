package instagram

import (
	"backend/models"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"

	repo "backend/repositories"
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
	UploadMedia(file multipart.File, fileName string, mimeType string) (string, error)
	CreateContainer(accessToken, instagramID, caption, mediaURL, mediaType string, isCarouselItem bool) (string, error)
	CreateCarouselContainer(accessToken string, instagramID string, caption string, containerIDs []string) (string, error)
	WaitForContainerReady(accessToken string, containerID string) (string, error)
	PublishMedia(accessToken string, instagramID string, creationID string) (string, error)
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

	existingAccessToken, _, err := i.GetCredentials(userID)
	if err != nil && err.Error() != "instagram token not found" {
		return err
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	if existingAccessToken == "" {
		return i.createToken(payloadBytes)
	}
	return i.updateToken(userID, payloadBytes)
}

func (i *instagramRepositoryImpl) createToken(payloadBytes []byte) error {
	url := i.repo_supabase.SupabaseURL + "instagram"
	req, err := repo.NewRequest(i.repo_supabase, "POST", url, bytes.NewBuffer(payloadBytes))
	if err != nil {
		return err
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed to create instagram tokens, status: %d, response: %s", resp.StatusCode, string(body))
	}
	return nil
}

func (i *instagramRepositoryImpl) updateToken(userID string, payloadBytes []byte) error {
	url := i.repo_supabase.SupabaseURL + "instagram?user_id=eq." + userID
	req, err := repo.NewRequest(i.repo_supabase, "PATCH", url, bytes.NewBuffer(payloadBytes))
	if err != nil {
		return err
	}

	client := &http.Client{}
	resp, err := client.Do(req)
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

func (i *instagramRepositoryImpl) UploadMedia(file multipart.File, fileExtension string, mimeType string) (string, error) {
	// Upload media to cloudflare
	mediaURL, err := i.repo_cloudflare.UploadFile(file, fmt.Sprintf("instagram_%d%s", time.Now().UnixNano(), fileExtension), mimeType)
	if err != nil {
		return "", fmt.Errorf("failed to upload media to Cloudflare: %w", err)
	}

	return mediaURL, nil
}

func (i *instagramRepositoryImpl) WaitForContainerReady(accessToken string, containerID string) (string, error) {
	const maxWait = 2 * time.Minute
	const initialBackoff = 2 * time.Second
	backoff := initialBackoff

	start := time.Now()

	for {
		status, err := i.containerStatus(accessToken, containerID)
		if err != nil {
			return "", fmt.Errorf("error getting container status: %w", err)
		}
		log.Printf("[WAIT_CONTAINER] --- Current container status: %s", status)

		if status == "FINISHED" {
			log.Println("[WAIT_CONTAINER] --- Container is ready")
			return status, nil
		}

		if time.Since(start) > maxWait {
			return "", fmt.Errorf("timeout waiting for container to be ready, last status: %s", status)
		}

		log.Printf("[WAIT_CONTAINER] --- Container not ready, retrying after %v...", backoff)
		time.Sleep(backoff)

		// Gradually increase backoff but cap it
		if backoff < 10*time.Second {
			backoff *= 2
		}
	}
}

func (i *instagramRepositoryImpl) containerStatus(accessToken string, containerID string) (string, error) {
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

func (i *instagramRepositoryImpl) CreateContainer(accessToken string, instagramID string, caption string, mediaURL string, mediaType string, isCarouselItem bool) (string, error) {
	log.Println("[CREATE_CONTAINER] --- Starting CreateContainer operation")
	log.Printf("[CREATE_CONTAINER] --- instagramID: %s, mediaType: %s, isCarouselItem: %v", instagramID, mediaType, isCarouselItem)
	log.Printf("[CREATE_CONTAINER] --- Caption: %.40s...", caption) // shows the first 40 characters if long
	log.Printf("[CREATE_CONTAINER] --- MediaURL: %s", mediaURL)

	url := instagram_api_path + instagramID + "/media"
	log.Printf("[CREATE_CONTAINER] --- Request URL: %s", url)

	req, err := http.NewRequest("POST", url, nil)
	if err != nil {
		log.Printf("[CREATE_CONTAINER] --- Error creating new request: %v", err)
		return "", err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+accessToken)

	q := req.URL.Query()

	if isCarouselItem {
		q.Add("is_carousel_item", "true")
		log.Println("[CREATE_CONTAINER] --- Added is_carousel_item true to query params")
	} else {
		q.Add("caption", caption)
		log.Printf("[CREATE_CONTAINER] --- Added caption to query params")
	}

	switch mediaType {
	case "IMAGE":
		q.Add("image_url", mediaURL)
		log.Println("[CREATE_CONTAINER] --- Added image_url to query params")
	case "VIDEO":
		if !isCarouselItem {
			q.Add("media_type", "REELS")
			log.Printf("[CREATE_CONTAINER] --- Changed media_type to REELS for video post")
		} else {
			q.Add("media_type", mediaType)
		}
		q.Add("video_url", mediaURL)

		log.Println("[CREATE_CONTAINER] --- Added video_url to query params")
	default:
		log.Printf("[CREATE_CONTAINER] --- Unsupported mediaType: %s", mediaType)
		return "", fmt.Errorf("unsupported media type: %s", mediaType)
	}

	req.URL.RawQuery = q.Encode()
	log.Printf("[CREATE_CONTAINER] --- Final request URL with query params: %s", req.URL.String())

	client := &http.Client{}
	log.Println("[CREATE_CONTAINER] --- Sending HTTP request to Instagram API...")
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("[CREATE_CONTAINER] --- HTTP request failed: %v", err)
		return "", err
	}
	defer resp.Body.Close()
	log.Printf("[CREATE_CONTAINER] --- HTTP response status: %d", resp.StatusCode)

	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		log.Printf("[CREATE_CONTAINER] --- ERROR response: %s", string(body))
		return "", fmt.Errorf("failed to create container, status %d: %s", resp.StatusCode, string(body))
	}

	var result struct {
		ID string `json:"id"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		log.Printf("[CREATE_CONTAINER] --- Failed to decode response body: %v", err)
		return "", err
	}

	log.Printf("[CREATE_CONTAINER] --- Successfully created container with ID: %s", result.ID)
	return result.ID, nil
}

func (i *instagramRepositoryImpl) CreateCarouselContainer(accessToken string, instagramID string, caption string, containerIDs []string) (string, error) {
	log.Println("[CREATE_CAROUSEL_CONTAINER] --- Starting carousel container creation")
	log.Printf("[CREATE_CAROUSEL_CONTAINER] --- instagramID: %s, caption length: %d, number of children: %d", instagramID, len(caption), len(containerIDs))

	url := instagram_api_path + instagramID + "/media"
	log.Printf("[CREATE_CAROUSEL_CONTAINER] --- Request URL: %s", url)

	req, err := http.NewRequest("POST", url, nil)
	if err != nil {
		log.Printf("[CREATE_CAROUSEL_CONTAINER] --- Error creating request: %v", err)
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+accessToken)

	q := req.URL.Query()
	q.Add("caption", caption)
	q.Add("media_type", "CAROUSEL")
	log.Println("[CREATE_CAROUSEL_CONTAINER] --- Added caption and media_type=CAROUSEL to query params")

	for idx, containerID := range containerIDs {
		q.Add(fmt.Sprintf("children[%d]", idx), containerID)
		log.Printf("[CREATE_CAROUSEL_CONTAINER] --- Added child container ID at index %d: %s", idx, containerID)
	}

	req.URL.RawQuery = q.Encode()
	log.Printf("[CREATE_CAROUSEL_CONTAINER] --- Final request URL with query params: %s", req.URL.String())

	// Try sending the request and handle potential transient errors
	resp, err := tryContainerCreation(req)
	if err != nil {
		return "", err
	}

	var result struct {
		ID string `json:"id"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		log.Printf("[CREATE_CAROUSEL_CONTAINER] --- JSON decode error: %v", err)
		return "", err
	}

	log.Printf("[CREATE_CAROUSEL_CONTAINER] --- Successfully created carousel container with ID: %s", result.ID)
	return result.ID, nil
}

func tryContainerCreation(req *http.Request) (*http.Response, error) {
	const maxRetries = 5
	backoff := 2 * time.Second
	client := &http.Client{}

	for attempt := 1; attempt <= maxRetries; attempt++ {
		log.Printf("[CREATE_CAROUSEL_CONTAINER] --- Attempt %d/%d: Sending HTTP request to Instagram API...", attempt, maxRetries)
		resp, err := client.Do(req)
		if err != nil {
			log.Printf("[CREATE_CAROUSEL_CONTAINER] --- HTTP request error: %v", err)
			return nil, err
		}

		if resp.StatusCode == 200 {
			log.Println("[CREATE_CAROUSEL_CONTAINER] --- Request succeeded with status 200")
			return resp, nil
		}

		body, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		log.Printf("[CREATE_CAROUSEL_CONTAINER] --- ERROR response body: %s", string(body))

		// Check if error is transient by inspecting the body
		var apiErr struct {
			Error struct {
				IsTransient bool   `json:"is_transient"`
				Message     string `json:"message"`
			} `json:"error"`
		}
		if err := json.Unmarshal(body, &apiErr); err == nil && apiErr.Error.IsTransient {
			log.Printf("[CREATE_CAROUSEL_CONTAINER] --- Transient error detected: %s", apiErr.Error.Message)
			if attempt < maxRetries {
				log.Printf("[CREATE_CAROUSEL_CONTAINER] --- Retrying after %v...", backoff)
				time.Sleep(backoff)
				backoff *= 2
				continue
			}
			return nil, fmt.Errorf("transient error after %d attempts: %s", attempt, apiErr.Error.Message)
		}

		return nil, fmt.Errorf("failed to create carousel container, status %d: %s", resp.StatusCode, string(body))
	}

	return nil, fmt.Errorf("max retries (%d) exceeded in tryContainerCreation", maxRetries)
}

func (i *instagramRepositoryImpl) PublishMedia(accessToken string, instagramID string, creationID string) (string, error) {
	log.Println("[PUBLISH_MEDIA] --- Starting media publish")
	log.Printf("[PUBLISH_MEDIA] --- instagramID: %s, creationID: %s", instagramID, creationID)

	url := instagram_api_path + instagramID + "/media_publish"
	log.Printf("[PUBLISH_MEDIA] --- Request URL: %s", url)

	req, err := http.NewRequest("POST", url, nil)
	if err != nil {
		log.Printf("[PUBLISH_MEDIA] --- Error creating request: %v", err)
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+accessToken)

	q := req.URL.Query()
	q.Add("creation_id", creationID)
	req.URL.RawQuery = q.Encode()
	log.Printf("[PUBLISH_MEDIA] --- Final request URL with query: %s", req.URL.String())

	client := &http.Client{}
	log.Println("[PUBLISH_MEDIA] --- Sending HTTP request to Instagram API...")
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("[PUBLISH_MEDIA] --- HTTP request error: %v", err)
		return "", err
	}
	defer resp.Body.Close()
	log.Printf("[PUBLISH_MEDIA] --- HTTP response status: %d", resp.StatusCode)

	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		log.Printf("[PUBLISH_MEDIA] --- ERROR response body: %s", string(body))
		return "", fmt.Errorf("failed to publish media, status %d: %s", resp.StatusCode, string(body))
	}

	var result struct {
		ID string `json:"id"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		log.Printf("[PUBLISH_MEDIA] --- JSON decode error: %v", err)
		return "", err
	}

	log.Printf("[PUBLISH_MEDIA] --- Media published successfully with ID: %s", result.ID)
	return result.ID, nil
}
