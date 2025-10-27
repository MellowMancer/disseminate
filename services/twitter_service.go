package services

import (
	"backend/repositories"
	"bytes"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	_ "net/http/httputil"
	
	_ "strings"
	"sync"
	"time"

	"github.com/dghubble/oauth1"
)

type TwitterService interface {
	GetAuthorizationURL() (string, string, error)
	GetAccessToken(oauthToken, requestSecret, oauthVerifier string) (string, string, error)
	CheckTokensValid(accessToken, accessSecret string) (bool, error)
	PostTweet(accessToken, accessSecret, content string, files []*multipart.FileHeader) error
}

type twitterServiceImpl struct {
	twitterConfig *oauth1.Config
	repository   *repositories.SupabaseRepository
}

func NewTwitterService(repo *repositories.SupabaseRepository, config *oauth1.Config) TwitterService {
	return &twitterServiceImpl{
		twitterConfig: config,
		repository: repo,
	}
}

func (s *twitterServiceImpl) GetAuthorizationURL() (string, string, error) {
	requestToken, requestSecret, err := s.twitterConfig.RequestToken()
	if err != nil {
		return "", "", err
	}

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

	return s.repository.CheckTwitterTokens(httpClient)
}

func (s *twitterServiceImpl) uploadMultipleMedia(httpClient *http.Client, files []*multipart.FileHeader) ([]string, error) {
	// Channels and a WaitGroup to handle concurrent uploads safely.
	var wg sync.WaitGroup
	resultChan := make(chan string, len(files))
	errChan := make(chan error, len(files))

	for _, fileHeader := range files {
		wg.Add(1) // Increment the WaitGroup counter

		// Launch a new goroutine for each file upload.
		go func(fh *multipart.FileHeader) {
			defer wg.Done() // Decrement the counter when the goroutine finishes

			file, err := fh.Open()
			if err != nil {
				errChan <- fmt.Errorf("failed to open file %s: %w", fh.Filename, err)
				return
			}
			defer file.Close()

			mediaData, err := io.ReadAll(file)
			if err != nil {
				errChan <- fmt.Errorf("failed to read file %s: %w", fh.Filename, err)
				return
			}
			mediaType := http.DetectContentType(mediaData)
			var mediaID string
			mediaID, err = s.uploadSingleChunked(httpClient, mediaData, mediaType)
			if err != nil {
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

	var mediaIDs []string
	// First, check for any errors. If we find one, we fail fast.
	if err := <-errChan; err != nil {
		return nil, err
	}

	// If no errors, collect all the successful media IDs.
	for id := range resultChan {
		mediaIDs = append(mediaIDs, id)
	}

	if len(mediaIDs) != len(files) {
		return nil, fmt.Errorf("an unknown error occurred: not all media files were uploaded")
	}

	return mediaIDs, nil
}

func (s *twitterServiceImpl) uploadSingleChunked(httpClient *http.Client, mediaData []byte, mediaType string) (string, error) {
	mediaCategory := ""
	switch mediaType {
	case "image/gif":
		mediaCategory = "tweet_gif"
	case "video/mp4":
		mediaCategory = "tweet_video"
	case "image/jpeg", "image/png":
		mediaCategory = "tweet_image"
	default:
		return "", fmt.Errorf("unsupported media type: %s", mediaType)
	}

	// 1. INIT
	mediaID, err := s.repository.InitUpload(httpClient, mediaData, mediaType, mediaCategory)
	if err != nil {
		return "", fmt.Errorf("chunked upload INIT failed: %w", err)
	}

	// 2. APPEND
	err = s.appendUploads(httpClient, mediaID, mediaData)
	if err != nil {
		return "", fmt.Errorf("chunked upload APPEND failed: %w", err)
	}

	// 3. FINALIZE
	err = s.repository.FinalizeUpload(httpClient, mediaID)
	if err != nil {
		return "", fmt.Errorf("chunked upload FINALIZE failed: %w", err)
	}

	// 4. STATUS CHECK
	err = s.statusHandlingLoop(httpClient, mediaCategory, mediaID)
	if err != nil {
		return "", fmt.Errorf("media processing failed: %w", err)
	}

	return mediaID, nil
}

func (s *twitterServiceImpl) appendUploads(httpClient *http.Client, mediaID string, mediaData []byte) error {
	const maxChunkSize = 4 * 1024 * 1024
	var segmentIndex int

	reader := bytes.NewReader(mediaData)
	chunk := make([]byte, maxChunkSize)

	for {
		bytesRead, err := reader.Read(chunk)
		if err != nil && err != io.EOF {
			return err
		}
		if bytesRead == 0 {
			break
		}

		statusCode, err := s.repository.AppendUpload(httpClient, mediaID, chunk[:bytesRead], segmentIndex)
		if err != nil {
			return fmt.Errorf("failed to append chunk %d: %w", segmentIndex, err)
		}
		if statusCode != http.StatusNoContent && statusCode != http.StatusOK {
			return fmt.Errorf("append chunk %d returned bad status: %d", segmentIndex, statusCode)
		}
		segmentIndex++
	}
	return nil
}

func (s *twitterServiceImpl) statusHandlingLoop(httpClient *http.Client, mediaCategory string, mediaID string) error {
	if mediaCategory != "tweet_video" && mediaCategory != "tweet_gif" {
		return nil // No processing needed for images
	}

	timeout := time.After(5 * time.Minute)
	checkAfter := 5 * time.Second

	for {
		select {
		case <-timeout:
			return fmt.Errorf("timed out while waiting for media processing")

		case <-time.After(checkAfter):
			statusResp, err := s.repository.StatusUpload(httpClient, mediaID)
			if err != nil {
				return fmt.Errorf("error during STATUS check: %w", err)
			}

			state := statusResp.Data.ProcessingInfo.State
			fmt.Printf("DEBUG: Media processing state is '%s', progress %d%%", state, statusResp.Data.ProcessingInfo.Progress)

			if state == "succeeded" {
				return nil // Success
			}

			if state == "failed" {
				return fmt.Errorf("media processing failed: %s", statusResp.Error.Message)
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


func (s *twitterServiceImpl) PostTweet(accessToken string, accessSecret string, content string, files []*multipart.FileHeader) error {
	var mediaIDs []string
	var err error

	token := oauth1.NewToken(accessToken, accessSecret)
	httpClient := s.twitterConfig.Client(oauth1.NoContext, token)

	postURL := "https://api.x.com/2/tweets"
	payload := map[string]interface{}{
		"text": content,
	}

	if len(files) > 0 {
		mediaIDs, err = s.uploadMultipleMedia(httpClient, files)

		if err != nil {
			return fmt.Errorf("media upload failed: %w", err)
		}
		payload["media"] = map[string]interface{}{
			"media_ids": mediaIDs,
		}
	}

	return s.repository.PostTweet(httpClient, postURL, payload)
}
