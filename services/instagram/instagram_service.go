package repo_instagram

import (
	repo_instagram "backend/repositories/instagram"
	"context"
	"fmt"
	"golang.org/x/oauth2"
	"mime/multipart"
	"net/http"
	"strings"
)

const longTimeTokenURL = "https://graph.repo_instagram.com/access_token"

type InstagramService interface {
	HandleLogin(w http.ResponseWriter, r *http.Request, state string)
	GetAccessToken(code string) (string, int, error)
	CheckTokensValid(accessToken string) error
	createContainer(accessToken string) (string, error)
	uploadMultipleMedia(accessToken string, files []*multipart.FileHeader) ([]string, error)
	publishMedia()
	containerStatus()
	checkPublishLimit(instagramID string, accessToken string) (bool, error)
	PostToInstagram(accessToken string, instagramID string, caption string, files []*multipart.FileHeader) (string, error)
}

type instagramServiceImpl struct {
	instagramConfig *oauth2.Config
	repo_instagram  repo_instagram.InstagramRepository
}

func NewInstagramService(config *oauth2.Config, repo repo_instagram.InstagramRepository) InstagramService {
	return &instagramServiceImpl{
		instagramConfig: config,
		repo_instagram:  repo,
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

	return i.repo_instagram.GetInstagramAccessToken(longTimeTokenURL, shortTermToken.AccessToken, i.instagramConfig.ClientSecret)
}

func (i *instagramServiceImpl) CheckTokensValid(accessToken string) error {
	return i.repo_instagram.CheckInstagramTokens(accessToken)
}

func (i *instagramServiceImpl) checkPublishLimit(instagramID string, accessToken string) (bool, error) {
	return i.repo_instagram.CheckInstagramPublishLimit(instagramID, accessToken)
}

func (i *instagramServiceImpl) uploadMedia(accessToken string, file multipart.File) (string, error) {
	return i.repo_instagram.UploadInstagramMedia(accessToken, file)
}

func (i *instagramServiceImpl) createContainer(accessToken string, instagramID string, caption string, file *multipart.FileHeader, isCarouselItem bool) (string, error) {
	mediaURL := ""
	f, err := file.Open()
	if err != nil {
		return "", fmt.Errorf("failed to open uploaded file: %w", err)
	}
	defer f.Close()
	buffer := make([]byte, 512)
	bytesRead, err := f.Read(buffer)
	if err != nil {
		return "", fmt.Errorf("failed to read media file: %w", err)
	}
	mediaType := http.DetectContentType(buffer[:bytesRead])
	switch {
	case strings.HasPrefix(mediaType, "image/"):
		mediaType = "IMAGE"
	case strings.HasPrefix(mediaType, "video/"):
		mediaType = "VIDEO"
	default:
		return "", fmt.Errorf("unsupported media type: %s", mediaType)
	}
	i.repo_instagram.CreateInstagramContainer(accessToken, instagramID, caption, mediaType, isCarouselItem)
	i.uploadMedia(accessToken, f)

	return
}

func (i *instagramServiceImpl) uploadMultipleMedia(accessToken string, files []*multipart.FileHeader) ([]string, error) {

}

func (i *instagramServiceImpl) containerStatus() {

}

func (i *instagramServiceImpl) publishMedia() {

}

func (i *instagramServiceImpl) PostToInstagram(accessToken string, instagramID string, caption string, files []*multipart.FileHeader) error {
	var containerID string
	var err error

	// 0. Publish Limit Check
	canPublish, err := i.checkPublishLimit(instagramID, accessToken)
	if err != nil {
		return err
	}
	if !canPublish {
		return fmt.Errorf("Publish limit reached for today")
	}

	fmt.Printf("--------CAN PUBLISH-----------")

	// 1. Upload media / Create containers
	if len(files) == 0 {
		return fmt.Errorf("No files attached")
	}
	if len(files) == 1 {
		containerID, err = i.createContainer(accessToken, instagramID, caption, files[0], false)
		if err != nil {
			return err
		}
	} else {
		containerID, err = i.uploadMultipleMedia(accessToken, instagramID, caption, files)
		if err != nil {
			return err
		}
	}

	// 2. Check publish limits and container status
	i.publishLimitCheck(containerID)

	// 3. Publish media
}
