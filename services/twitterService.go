package services

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/dghubble/oauth1"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	_ "net/http/httputil"
	"strconv"
	_ "strings"
	"sync"
	"time"
)

type TwitterService interface {
	GetAuthorizationURL() (string, string, error)
	GetAccessToken(oauthToken, requestSecret, oauthVerifier string) (string, string, error)
	CheckTokensValid(accessToken, accessSecret string) (bool, error)
	UploadMultipleMedia(httpClient *http.Client, files []*multipart.FileHeader) ([]string, error)
	PostTweet(accessToken, accessSecret, content string, files []*multipart.FileHeader) error
}

type twitterServiceImpl struct {
	twitterConfig *oauth1.Config
}

func NewTwitterService(config *oauth1.Config) TwitterService {
	return &twitterServiceImpl{
		twitterConfig: config,
	}
}
type v2InitResponse struct {
	Data struct {
		ID string `json:"id"`
	} `json:"data"`
}

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

func (s *twitterServiceImpl) GetAuthorizationURL() (string, string, error) {
	requestToken, requestSecret, err := s.twitterConfig.RequestToken()
	if err != nil {
		return "", "", err
	}

	_ = requestSecret // Store this secret for later use in the callback

	authURL, err := s.twitterConfig.AuthorizationURL(requestToken)
	if err != nil {
		return "", "", err
	}

	return authURL.String(), requestSecret, nil
}

func (s *twitterServiceImpl) GetAccessToken(oauthToken, requestSecret, oauthVerifier string) (string, string, error) {
	accessToken, accessSecret, err := s.twitterConfig.AccessToken(oauthToken, requestSecret, oauthVerifier)
	if err != nil {
		return "", "", err
	}
	return accessToken, accessSecret, nil
}

func (s *twitterServiceImpl) CheckTokensValid(accessToken, accessSecret string) (bool, error) {
	token := oauth1.NewToken(accessToken, accessSecret)
	httpClient := s.twitterConfig.Client(oauth1.NoContext, token)

	resp, err := httpClient.Get("https://api.twitter.com/1.1/account/verify_credentials.json")
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case http.StatusOK:
		return true, nil
	case http.StatusUnauthorized:
		return false, errors.New("tokens invalid or revoked")
	default:
		return false, errors.New("unexpected response from Twitter API")
	}
}

func (s *twitterServiceImpl) UploadMultipleMedia(httpClient *http.Client, files []*multipart.FileHeader) ([]string, error) {
	log.Println("--- twitterService.UploadMultipleMedia: START ---")
	// We will use channels and a WaitGroup to handle concurrent uploads safely.
	var wg sync.WaitGroup
	resultChan := make(chan string, len(files))
	errChan := make(chan error, len(files))
	log.Println("Starting upload of", len(files), "files...")
	for _, fileHeader := range files {
		wg.Add(1) // Increment the WaitGroup counter

		// Launch a new goroutine for each file upload.
		go func(fh *multipart.FileHeader) {
			defer wg.Done() // Decrement the counter when the goroutine finishes

			file, err := fh.Open()
			if err != nil {
				log.Println("--- twitterService.UploadMultipleMedia: ERROR opening file ---")
				errChan <- fmt.Errorf("failed to open file %s: %w", fh.Filename, err)
				return
			}
			defer file.Close()

			mediaData, err := io.ReadAll(file)
			if err != nil {
				log.Println("--- twitterService.UploadMultipleMedia: ERROR reading file ---")
				errChan <- fmt.Errorf("failed to read file %s: %w", fh.Filename, err)
				return
			}
			mediaType := http.DetectContentType(mediaData)
			var mediaID string
			mediaID, err = s.uploadSingleChunked(httpClient, mediaData, mediaType)
			if err != nil {
				log.Println("--- twitterService.UploadMultipleMedia: ERROR uploading file ---")
				errChan <- fmt.Errorf("failed to upload %s: %w", fh.Filename, err)
				return
			}

			resultChan <- mediaID
		}(fileHeader)
	}

	// Wait for all goroutines to finish, then close the channels.
	go func() {
		wg.Wait()
		close(resultChan)
		close(errChan)
	}()

	// Collect the results.
	var mediaIDs []string
	// First, check for any errors. If we find one, we fail fast.
	if err := <-errChan; err != nil {
		log.Println("--- twitterService.UploadMultipleMedia: ERROR detected ---")
		return nil, err
	}
	// If no errors, collect all the successful media IDs.
	for id := range resultChan {
		mediaIDs = append(mediaIDs, id)
	}

	if len(mediaIDs) != len(files) {
		log.Println("--- twitterService.UploadMultipleMedia: ERROR - not all media uploaded ---")
		return nil, errors.New("an unknown error occurred: not all media files were uploaded")
	}

	return mediaIDs, nil
}

func (s *twitterServiceImpl) uploadSingleChunked(httpClient *http.Client, mediaData []byte, mediaType string) (string, error) {
	log.Println("--- twitterService.uploadSingleChunked: START ---")
	mediaCategory := ""
	switch mediaType {
	case "image/gif":
		mediaCategory = "tweet_gif"
	case "video/mp4":
		mediaCategory = "tweet_video"
	case "image/jpeg", "image/png":
		mediaCategory = "tweet_image"
	default:
		log.Println("--- twitterService.uploadSingleChunked: ERROR - unsupported media type ---")
		log.Printf("Media type: %s", mediaType)
		return "", fmt.Errorf("unsupported media type: %s", mediaType)
	}

	// 1. INIT
	mediaID, err := s.initUpload(httpClient, mediaData, mediaType, mediaCategory)
	if err != nil {
		log.Println("--- twitterService.uploadSingleChunked: ERROR during INIT ---")
		return "", fmt.Errorf("chunked upload INIT failed: %w", err)
	}

	// 2. APPEND
	const maxChunkSize = 4 * 1024 * 1024
	var segmentIndex int

	reader := bytes.NewReader(mediaData)
	chunk := make([]byte, maxChunkSize)

	for {
		bytesRead, err := reader.Read(chunk)
		if err != nil && err != io.EOF {
			return "", err
		}
		if bytesRead == 0 {
			break
		}

		statusCode, err := s.appendUpload(httpClient, mediaID, chunk[:bytesRead], segmentIndex)
		if err != nil {
			return "", fmt.Errorf("failed to append chunk %d: %w", segmentIndex, err)
		}
		if statusCode != http.StatusNoContent && statusCode != http.StatusOK {
			return "", fmt.Errorf("append chunk %d returned bad status: %d", segmentIndex, statusCode)
		}

		segmentIndex++
	}

	// --- 3. FINALIZE (Now conditional based on append's response) ---
	err = s.finalizeUpload(httpClient, mediaID)
	if err != nil {
		return "", fmt.Errorf("chunked upload FINALIZE failed: %w", err)
	}

	// --- 4. V2 STATUS CHECK  ---
	// This check is only necessary for media that requires server-side processing.
	if mediaCategory == "tweet_video" || mediaCategory == "tweet_gif" {
		log.Println("Media requires processing, beginning status checks...")

		timeout := time.After(5 * time.Minute)
		checkAfter := 5 * time.Second

		for {
			select {
			case <-timeout:
				return "", errors.New("timed out while waiting for media processing")

			case <-time.After(checkAfter):
				statusResp, err := s.statusUpload(httpClient, mediaID)
				if err != nil {
					return "", fmt.Errorf("error during STATUS check: %w", err)
				}

				state := statusResp.Data.ProcessingInfo.State
				log.Printf("DEBUG: Media processing state is '%s', progress %d%%", state, statusResp.Data.ProcessingInfo.Progress)

				if state == "succeeded" {
					log.Println("--- twitterService.uploadSingleChunked: SUCCESS (Video Processed) ---")
					return mediaID, nil // Success!
				}

				if state == "failed" {
					return "", fmt.Errorf("media processing failed: %s", statusResp.Error.Message)
				}

				// Update the wait time for the next loop from the API's suggestion.
				if statusResp.Data.ProcessingInfo.CheckAfterSecs > 0 {
					checkAfter = time.Duration(statusResp.Data.ProcessingInfo.CheckAfterSecs) * time.Second
				} else {
					checkAfter = 5 * time.Second // Fallback delay
				}
			}
		}
	}
	return mediaID, nil
}

func (s *twitterServiceImpl) initUpload(httpClient *http.Client, mediaData []byte, mediaType string, mediaCategory string) (string, error) {
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
		log.Printf("--> ERROR from httpClient.Do in initUpload: %v", err)
		return "", fmt.Errorf("http client failed during INIT: %w", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	log.Printf("DEBUG: Twitter INIT response status=%d, body=%s", resp.StatusCode, string(body))

	if resp.StatusCode != http.StatusOK {
		// This log is still useful if the status is not 200
		log.Printf("--> ERROR from Twitter API (INIT): status=%d, body=%s", resp.StatusCode, string(body))
		return "", fmt.Errorf("bad status on INIT: %s, response: %s", resp.Status, string(body))
	}

	var initResp v2InitResponse
	if err := json.Unmarshal(body, &initResp); err != nil {
		log.Printf("--> ERROR during json.Unmarshal in initUpload: %v", err)
		return "", fmt.Errorf("failed to parse INIT response: %w", err)
	}

	// CHANGE: Check the correct field for the media ID.
	if initResp.Data.ID == "" {
		log.Printf("--> ERROR: Media ID is empty in INIT response's data object")
		return "", errors.New("INIT response did not contain a media id")
	}

	log.Println("--- twitterService.initUpload: SUCCESS ---")
	// CHANGE: Return the correct field.
	return initResp.Data.ID, nil
}

func (s *twitterServiceImpl) appendUpload(httpClient *http.Client, mediaID string, mediaData []byte, segmentIndex int) (int, error) {
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

func (s *twitterServiceImpl) finalizeUpload(httpClient *http.Client, mediaID string) error {
	log.Println("--- twitterService.finalizeUpload: START ---")
	finalizeURL := fmt.Sprintf("https://api.x.com/2/media/upload/%s/finalize", mediaID)

	// The v2 finalize request has an empty body.
	req, err := http.NewRequest("POST", finalizeURL, nil)
	if err != nil {
		log.Println("--- twitterService.finalizeUpload: ERROR creating request ---")
		return fmt.Errorf("failed to create finalize request: %w", err)
	}

	resp, err := httpClient.Do(req)
	if err != nil {
		log.Println("--- twitterService.finalizeUpload: ERROR during httpClient.Do ---")
		return err
	}
	defer resp.Body.Close()
	log.Println("DEBUG: Finalize request sent, awaiting response...")

	// The response body is typically empty for a successful finalize.
	// We log it anyway for debugging purposes.
	body, _ := io.ReadAll(resp.Body)
	log.Printf("DEBUG: Twitter FINALIZE response status=%d, body=%s", resp.StatusCode, string(body))

	// The documented success code is 204 No Content.
	if resp.StatusCode != http.StatusNoContent && resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		log.Printf("--> ERROR from Twitter API (FINALIZE): status=%d, body=%s", resp.StatusCode, string(respBody))
		return fmt.Errorf("bad status on FINALIZE: %s, response: %s", resp.Status, string(respBody))
	}

	log.Println("--- twitterService.finalizeUpload: SUCCESS ---")
	return nil
}

func (s *twitterServiceImpl) statusUpload(httpClient *http.Client, mediaID string) (*v2StatusResponse, error) {
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

func (s *twitterServiceImpl) PostTweet(accessToken string, accessSecret string, content string, files []*multipart.FileHeader) error {
	log.Println("--- twitterService.PostTweet: START ---")

	var mediaIDs []string
	var err error

	token := oauth1.NewToken(accessToken, accessSecret)
	httpClient := s.twitterConfig.Client(oauth1.NoContext, token)

	postURL := "https://api.x.com/2/tweets"
	payload := map[string]interface{}{
		"text": content,
	}

	if len(files) > 0 {
		log.Println("Files detected, calling UploadMultipleMedia...")
		mediaIDs, err = s.UploadMultipleMedia(httpClient, files)

		if err != nil {
			return fmt.Errorf("media upload failed: %w", err)
		}
		payload["media"] = map[string]interface{}{
			"media_ids": mediaIDs,
		}
		log.Printf("UploadMultipleMedia successful. Media IDs: %v", mediaIDs)
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal tweet payload: %w", err)
	}

	req, err := http.NewRequest("POST", postURL, bytes.NewBuffer(payloadBytes))
	if err != nil {
		return fmt.Errorf("failed to create tweet request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	log.Println("Sending final request to POST /2/tweets")
	resp, err := httpClient.Do(req)
	if err != nil {
		log.Printf("--> ERROR from httpClient.Do: %v", err)
		return fmt.Errorf("tweet request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		log.Printf("--> ERROR from Twitter API: status=%d, body=%s", resp.StatusCode, string(body))
		return fmt.Errorf("failed to post tweet, status: %d, response: %s", resp.StatusCode, string(body))
	}

	if resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed to post tweet, status: %d, response: %s", resp.StatusCode, string(body))
	}
	log.Println("--- twitterService.PostTweet: SUCCESS ---")
	return nil
}
