package repositories

import (
	"backend/models"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	_ "log"
	"net/http"
	_ "net/http/httputil"
	"time"
)

func (r *SupabaseRepository) SaveInstagramToken(userID string, instagramID string, accessToken string, expirationTime time.Time) error {
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

	url := r.supabaseURL + api_path + "instagram"

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

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed to update instagram tokens, status: %d, response: %s", resp.StatusCode, string(body))
	}

	return nil
}

func (r *SupabaseRepository) GetInstagramToken(userID string) (string, error) {
	req, _ := http.NewRequest("GET", r.supabaseURL+api_path+"instagram", nil)
	req.Header.Set("apikey", r.supabaseKey)
	req.Header.Set("Authorization", "Bearer "+r.supabaseKey)

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

func (r *SupabaseRepository) GetInstagramID(accessToken string) (string, error) {
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

func (r *SupabaseRepository) GetAccessToken(URL string, accessToken string, clientSecret string) (string, int, error) {
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

func (r *SupabaseRepository) CheckInstagramTokens(accessToken string) error {
	resp, err := http.Get("https://graph.instagram.com/me?access_token=" + accessToken)
    if err != nil || resp.StatusCode != 200 {
        return fmt.Errorf("tokens have been revoked, please connect your account again")
    }
	return nil
}