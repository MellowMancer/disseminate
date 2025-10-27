package services

import (
    "backend/repositories"
	"net/http"
    "context"
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
    repository   *repositories.SupabaseRepository
}

func NewInstagramService(config *oauth2.Config, repo *repositories.SupabaseRepository) InstagramService {
	return &instagramServiceImpl{
		instagramConfig: config,
        repository: repo,
	}
}

func (i *instagramServiceImpl) HandleLogin(w http.ResponseWriter, r *http.Request, state string) {
	url := i.instagramConfig.Endpoint.AuthURL + "&state=" + state
	http.Redirect(w, r, url, http.StatusFound)
}

func (i *instagramServiceImpl) GetAccessToken(code string) (string, int, error) {
    shortTermToken, err := i.instagramConfig.Exchange(context.Background(), code)
    if err != nil {
        return "", 0, err
    }

    return i.repository.GetAccessToken(longTimeTokenURL, shortTermToken.AccessToken, i.instagramConfig.ClientSecret)
}

func (i *instagramServiceImpl) CheckTokensValid(accessToken string) error {
    return i.repository.CheckInstagramTokens(accessToken)
}
