package twitter

import (
	"backend/models"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"log"
	"strconv"
	"mime/multipart"
	repo_supabase "backend/repositories/supabase"
)

type v2StatusResponse struct {
	Data struct {
		ProcessingInfo struct {
			State          string `json:"state"`
			Progress       int    `json:"progress_percent"`
			CheckAfterSecs int    `json:"check_after_secs"`
		} `json:"processing_info"`
	} `json:"data"`
	Error struct {
		Message string `json:"message"`
	} `json:"error"`
}

type v2InitResponse struct {
	Data struct {
		ID string `json:"id"`
	} `json:"data"`
}

type TwitterRepository interface {
	SaveToken(userID string, accessToken string, accessSecret string) error
	GetToken(userID string) (string, string, error)
	CheckTokens(client *http.Client) (bool, error)
	InitUpload(httpClient *http.Client, mediaData []byte, mediaType string, mediaCategory string) (string, error)
	AppendUpload(httpClient *http.Client, mediaID string, mediaData []byte, segmentIndex int) (int, error)
	FinalizeUpload(httpClient *http.Client, mediaID string) error
	StatusUpload(httpClient *http.Client, mediaID string) (*v2StatusResponse, error)
	PostTweet(client *http.Client, postURL string, payload map[string]interface{}) error
}

type twitterRepositoryImpl struct {
	repo_supabase *repo_supabase.SupabaseRepository
}

func NewTwitterRepository(supabaseRepository *repo_supabase.SupabaseRepository) TwitterRepository {
	return &twitterRepositoryImpl{
		repo_supabase: supabaseRepository,
	}
}

func (t *twitterRepositoryImpl) SaveToken(userID string, accessToken string, accessSecret string) error {
    payload := map[string]string{
        "user_id":      userID,
        "access_token": accessToken,
        "access_secret": accessSecret,
    }
    payloadBytes, err := json.Marshal(payload)
    if err != nil {
        return err
    }
    url := t.repo_supabase.SupabaseURL + "twitter"
    req, err := http.NewRequest("POST", url, bytes.NewBuffer(payloadBytes))
    if err != nil {
        return err
    }
    req.Header.Set("apikey", t.repo_supabase.SupabaseKey)
    req.Header.Set("Authorization", "Bearer "+t.repo_supabase.SupabaseKey)
    req.Header.Set("Content-Type", "application/json")
    req.Header.Set("Prefer", "return=representation")
    resp, err := http.DefaultClient.Do(req)
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

func (t *twitterRepositoryImpl) GetToken(userID string) (string, string, error) {
	req, _ := http.NewRequest("GET", t.repo_supabase.SupabaseURL + "twitter", nil)
	req.Header.Set("apikey", t.repo_supabase.SupabaseKey)
	req.Header.Set("Authorization", "Bearer "+t.repo_supabase.SupabaseKey)

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

	var twitterModel []models.TwitterModel
	if err := json.NewDecoder(resp.Body).Decode(&twitterModel); err != nil {
		return "", "", fmt.Errorf("failed to decode response: %v", err)
	}

	if len(twitterModel) == 0 || twitterModel[0].AccessToken == "" || twitterModel[0].AccessSecret == "" {
		return "", "", fmt.Errorf("twitter tokens not found")
	}

	return twitterModel[0].AccessToken, twitterModel[0].AccessSecret, nil
}

func (t *twitterRepositoryImpl) CheckTokens(client *http.Client) (bool, error) {
	resp, err := client.Get("https://api.twitter.com/1.1/account/verify_credentials.json")
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case http.StatusOK:
		return true, nil
	case http.StatusUnauthorized:
		return false, fmt.Errorf("tokens invalid or revoked")
	default:
		return false, fmt.Errorf("unexpected response from Twitter API")
	}
}

func (t *twitterRepositoryImpl) InitUpload(httpClient *http.Client, mediaData []byte, mediaType string, mediaCategory string) (string, error) {
	log.Println("--- twitterService.initUpload: START ---")

	const initializeURL = "https://api.x.com/2/media/upload/initialize"

	payload := struct {
		TotalBytes    int    `json:"total_bytes"`
		MediaType     string `json:"media_type"`
		MediaCategory string `json:"media_category"`
	}{
		TotalBytes:    len(mediaData),
		MediaType:     mediaType,
		MediaCategory: mediaCategory,
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("failed to create INIT payload: %w", err)
	}

	req, err := http.NewRequest("POST", initializeURL, bytes.NewBuffer(payloadBytes))
	if err != nil {
		return "", fmt.Errorf("failed to create INIT request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("http client failed during INIT: %w", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusOK {
		// This log is still useful if the status is not 200
		return "", fmt.Errorf("bad status on INIT: %s, response: %s", resp.Status, string(body))
	}

	var initResp v2InitResponse
	if err := json.Unmarshal(body, &initResp); err != nil {
		return "", fmt.Errorf("failed to parse INIT response: %w", err)
	}

	// Check the correct field for the media ID.
	if initResp.Data.ID == "" {
		return "", fmt.Errorf("INIT response did not contain a media id")
	}

	log.Println("--- twitterService.initUpload: SUCCESS ---")
	// Return the correct field.
	return initResp.Data.ID, nil
}

func (t *twitterRepositoryImpl) AppendUpload(httpClient *http.Client, mediaID string, mediaData []byte, segmentIndex int) (int, error) {
	appendURL := fmt.Sprintf("https://api.x.com/2/media/upload/%s/append", mediaID)

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	// Part 1: The media chunk itself. The field name MUST be "media".
	mediaPart, err := writer.CreateFormFile("media", "media.bin") // filename is arbitrary
	if err != nil {
		return 0, fmt.Errorf("failed to create append request: %w", err)
	}
	_, err = mediaPart.Write(mediaData)
	if err != nil {
		return 0, fmt.Errorf("failed to write media data to form: %w", err)
	}

	// Part 2: The segment_index. This is a regular form field.
	// The field name MUST be "segment_index".
	err = writer.WriteField("segment_index", strconv.Itoa(segmentIndex))
	if err != nil {
		return 0, fmt.Errorf("failed to write segment_index to form: %w", err)
	}

	// This finalizes the multipart body.
	writer.Close()

	req, err := http.NewRequest("POST", appendURL, body)
	if err != nil {
		return 0, fmt.Errorf("failed to create append request: %w", err)
	}

	// CRITICAL FIX: The Content-Type must be set from the multipart writer,
	// as it includes the unique boundary string.
	req.Header.Set("Content-Type", writer.FormDataContentType())

	// --- The rest of your function is correct ---
	resp, err := httpClient.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	// A successful APPEND to the v2 endpoint returns a 204 No Content status.
	if resp.StatusCode != http.StatusNoContent && resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		log.Printf("--> ERROR from Twitter API (APPEND): status=%d, body=%s", resp.StatusCode, string(respBody))
		return 0, fmt.Errorf("bad status on APPEND: %s, response: %s", resp.Status, string(respBody))
	}

	// On success, return the status code we received.
	return resp.StatusCode, nil
}

func (t *twitterRepositoryImpl) FinalizeUpload(httpClient *http.Client, mediaID string) error {
	log.Println("--- twitterService.finalizeUpload: START ---")
	finalizeURL := fmt.Sprintf("https://api.x.com/2/media/upload/%s/finalize", mediaID)

	// The v2 finalize request has an empty body.
	req, err := http.NewRequest("POST", finalizeURL, nil)
	if err != nil {
		return fmt.Errorf("failed to create finalize request: %w", err)
	}

	resp, err := httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// The documented success code is 204 No Content.
	if resp.StatusCode != http.StatusNoContent && resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		log.Printf("--> ERROR from Twitter API (FINALIZE): status=%d, body=%s", resp.StatusCode, string(respBody))
		return fmt.Errorf("bad status on FINALIZE: %s, response: %s", resp.Status, string(respBody))
	}

	log.Println("--- twitterService.finalizeUpload: SUCCESS ---")
	return nil
}

func (t *twitterRepositoryImpl) StatusUpload(httpClient *http.Client, mediaID string) (*v2StatusResponse, error) {
	log.Println("--- twitterService.statusUpload: START ---")
	statusURL := "https://api.x.com/2/media/upload"

	// The status check is a GET request.
	req, err := http.NewRequest("GET", statusURL, nil)
	if err != nil {
		log.Println("--- twitterService.statusUpload: ERROR creating request ---")
		return nil, err
	}

	q := req.URL.Query()
	q.Add("media_id", mediaID)
	q.Add("command", "STATUS")
	req.URL.RawQuery = q.Encode()

	resp, err := httpClient.Do(req)
	if err != nil {
		log.Println("--- twitterService.statusUpload: ERROR during httpClient.Do ---")
		return nil, err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		log.Printf("--> ERROR from Twitter API (STATUS): status=%d, body=%s", resp.StatusCode, string(body))
		return nil, fmt.Errorf("bad status on STATUS check: %s, response: %s", resp.Status, string(body))
	}

	var statusResp v2StatusResponse
	if err := json.Unmarshal(body, &statusResp); err != nil {
		log.Println("--- twitterService.statusUpload: ERROR unmarshalling response ---")
		return nil, fmt.Errorf("failed to parse STATUS response: %w", err)
	}

	return &statusResp, nil
}

func (t *twitterRepositoryImpl) PostTweet(client *http.Client, postURL string, payload map[string]interface{}) error {
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal tweet payload: %w", err)
	}

	req, err := http.NewRequest("POST", postURL, bytes.NewBuffer(payloadBytes))
	if err != nil {
		return fmt.Errorf("failed to create tweet request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("tweet request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed to post tweet, status: %d, response: %s", resp.StatusCode, string(body))
	}

	if resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed to post tweet, status: %d, response: %s", resp.StatusCode, string(body))
	}
	return nil
}