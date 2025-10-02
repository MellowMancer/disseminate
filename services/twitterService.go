package services

import (
	"errors"
	"github.com/dghubble/oauth1"
	"net/http"
)

type TwitterService interface {
	GetAuthorizationURL() (string, string, error)
	GetAccessToken(oauthToken, requestSecret, oauthVerifier string) (string, string, error)
	CheckTokensValid(accessToken, accessSecret string) (bool, error)
}

type twitterServiceImpl struct {
	twitterConfig *oauth1.Config
}

func NewTwitterService(config *oauth1.Config) TwitterService {
	return &twitterServiceImpl{
		twitterConfig: config,
	}
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
