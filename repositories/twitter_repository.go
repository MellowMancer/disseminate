package repositories

import (
	"backend/models"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type TwitterRepository interface {
	SaveTwitterToken(userID string, accessToken string, accessSecret string) error
	GetTwitterToken(userID string) (string, string, error)
}

func (r *SupabaseRepository) SaveTwitterToken(userID string, accessToken string, accessSecret string) error {
    payload := map[string]string{
        "user_id":      userID,
        "access_token": accessToken,
        "access_secret": accessSecret,
    }
    payloadBytes, err := json.Marshal(payload)
    if err != nil {
        return err
    }
    url := r.supabaseURL + api_path + "twitter"
    req, err := http.NewRequest("POST", url, bytes.NewBuffer(payloadBytes))
    if err != nil {
        return err
    }
    req.Header.Set("apikey", r.supabaseKey)
    req.Header.Set("Authorization", "Bearer "+r.supabaseKey)
    req.Header.Set("Content-Type", "application/json")
    req.Header.Set("Prefer", "return=representation")
    resp, err := r.httpClient.Do(req)
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

func (r *SupabaseRepository) GetTwitterToken(userID string) (string, string, error) {
	req, _ := http.NewRequest("GET", r.supabaseURL+api_path+"twitter", nil)
	req.Header.Set("apikey", r.supabaseKey)
	req.Header.Set("Authorization", "Bearer "+r.supabaseKey)

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
