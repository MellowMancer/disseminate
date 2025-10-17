package services

import (
	"context"
	"net/http"
	"fmt"
	"io"
	"encoding/json"
	"golang.org/x/oauth2"
	_ "golang.org/x/oauth2/facebook"
)

const longTimeTokenURL = "https://graph.instagram.com/access_token"

type InstagramService interface {
	HandleLogin(w http.ResponseWriter, r *http.Request, state string)
	GetAccessToken(code string) (string, int, error)
    CheckTokensValid(accessToken string) error
}

type instagramServiceImpl struct {
	instagramConfig *oauth2.Config
}

func NewInstagramService(config *oauth2.Config) InstagramService {
	return &instagramServiceImpl{
		instagramConfig: config,
	}
}

func (i *instagramServiceImpl) HandleLogin(w http.ResponseWriter, r *http.Request, state string) {
	url := i.instagramConfig.Endpoint.AuthURL + "&state=" + state
	http.Redirect(w, r, url, http.StatusFound)
}

func (i *instagramServiceImpl) GetAccessToken(code string) (string, int, error) {
    fmt.Println("[GetAccessToken] Starting token exchange with code:", code)

    shortTermToken, err := i.instagramConfig.Exchange(context.Background(), code)
    if err != nil {
        fmt.Printf("[GetAccessToken] Error exchanging code for short-lived token: %v\n", err)
        return "", 0, err
    }
    fmt.Println("[GetAccessToken] Obtained short-lived token")

    req, err := http.NewRequest("GET", longTimeTokenURL, nil)
    if err != nil {
        fmt.Printf("[GetAccessToken] Failed to create request for long-lived token: %v\n", err)
        return "", 0, err
    }

    q := req.URL.Query()
    q.Add("grant_type", "ig_exchange_token")
    q.Add("client_secret", i.instagramConfig.ClientSecret)
    q.Add("access_token", shortTermToken.AccessToken)
    req.URL.RawQuery = q.Encode()

    fmt.Printf("[GetAccessToken] Making request to exchange for long-lived token: %s\n", req.URL.String())

    resp, err := http.DefaultClient.Do(req)
    if err != nil {
        fmt.Printf("[GetAccessToken] Error making HTTP request for long-lived token: %v\n", err)
        return "", 0, err
    }
    defer resp.Body.Close()

    if resp.StatusCode != http.StatusOK {
        body, _ := io.ReadAll(resp.Body)
        fmt.Printf("[GetAccessToken] Token exchange failed with status %d: %s\n", resp.StatusCode, string(body))
        return "", 0, fmt.Errorf("instagram token exchange failed: status %d: %s", resp.StatusCode, string(body))
    }

    var longToken struct {
        AccessToken string `json:"access_token"`
        TokenType   string `json:"token_type"`
        ExpiresIn   int    `json:"expires_in"`
    }

    if err = json.NewDecoder(resp.Body).Decode(&longToken); err != nil {
        fmt.Printf("[GetAccessToken] Error decoding long-lived token response: %v\n", err)
        return "", 0, err
    }

    fmt.Printf("[GetAccessToken] Long-lived token obtained, expires in %d seconds\n", longToken.ExpiresIn)

    return longToken.AccessToken, longToken.ExpiresIn, nil
}

func (i *instagramServiceImpl) CheckTokensValid(accessToken string) error {
    resp, err := http.Get("https://graph.instagram.com/me?access_token=" + accessToken)
    if err != nil || resp.StatusCode != 200 {
        return fmt.Errorf("tokens have been revoked, please connect your account again")
    }
    return nil
}
