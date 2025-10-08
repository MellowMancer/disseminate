package services

import (
	"golang.org/x/oauth2"
	_"golang.org/x/oauth2/facebook"
	"net/http"
)

type InstagramService interface {
	HandleLogin(w http.ResponseWriter, r *http.Request)
}

type instagramServiceImpl struct {
	instagramConfig *oauth2.Config
	metaConfigurationID string
}

func NewInstagramService(config *oauth2.Config, configId string) InstagramService {
	return &instagramServiceImpl{
		instagramConfig: config,
		metaConfigurationID: configId,
	}
}

func (i *instagramServiceImpl) HandleLogin(w http.ResponseWriter, r *http.Request) {
	url := i.instagramConfig.AuthCodeURL("state")
	url += "&config_id=" + i.metaConfigurationID
	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}