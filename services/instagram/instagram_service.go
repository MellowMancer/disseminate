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

type InstagramService interface {
	HandleLogin(w http.ResponseWriter, r *http.Request, state string)
	GetAccessToken(code string) (string, int, error)
	CheckTokensValid(accessToken string) error
	createContainer(accessToken string, instagramID string, caption string, file *multipart.FileHeader, isCarouselItem bool) (string, error)
	createCarouselContainer(accessToken string, instagramID string, caption string, files []*multipart.FileHeader) (string, error)
	publishMedia()
	containerStatus(accessToken string, containerID string) (string, error)
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

	fmt.Printf("[INSTAGRAM_SERVICE] --- Short-term token obtained ---")

	return i.repo_instagram.GetAccessToken(shortTermToken.AccessToken, i.instagramConfig.ClientSecret)
}

func (i *instagramServiceImpl) CheckTokensValid(accessToken string) error {
	return i.repo_instagram.CheckTokens(accessToken)
}

func (i *instagramServiceImpl) checkPublishLimit(accessToken string, instagramID string, ) (bool, error) {
	return i.repo_instagram.CheckPublishLimit( accessToken, instagramID)
}

func (i *instagramServiceImpl) uploadMedia(file multipart.File) (string, error) {
	return i.repo_instagram.UploadMedia(file)
}

func (i *instagramServiceImpl) createContainer(accessToken string, instagramID string, caption string, file *multipart.FileHeader, isCarouselItem bool) (string, error) {
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
	
	_, err = i.uploadMedia(f)
	if err != nil {
		return "", fmt.Errorf("failed to upload media: %w", err)
	}

	containerID, err := i.repo_instagram.CreateContainer(accessToken, instagramID, caption, mediaType, isCarouselItem)
	if err != nil {
		return "", fmt.Errorf("failed to create media container: %w", err)
	}

	return containerID, nil
}

func (i *instagramServiceImpl) createCarouselContainer(accessToken string, instagramID string, caption string, files []*multipart.FileHeader) (string, error) {
	containerIDs := []string{}
	for _, file := range files {
		containerID, err := i.createContainer(accessToken, instagramID, caption, file, true)
		if err != nil {
			return "", err
		}
		containerIDs = append(containerIDs, containerID)
	}

	// Create a carousel container
	containerID, err := i.repo_instagram.CreateCarouselContainer(accessToken, instagramID, caption, containerIDs)
	if err != nil {
		return "", fmt.Errorf("failed to create carousel container: %w", err)
	}
	return containerID, nil
}

func (i *instagramServiceImpl) containerStatus(accessToken string, containerID string) (string, error) {
	return i.repo_instagram.ContainerStatus(accessToken, containerID)
}

func (i *instagramServiceImpl) publishMedia() {

}

func (i *instagramServiceImpl) PostToInstagram(accessToken string, instagramID string, caption string, files []*multipart.FileHeader) (string, error) {
	var containerID string
	var err error

	fmt.Printf("--------POST TO INSTAGRAM STARTING-----------")

	// 0. Publish Limit Check
	canPublish, err := i.checkPublishLimit(accessToken, instagramID)
	if err != nil {
		return "", err
	}
	if !canPublish {
		return "", fmt.Errorf("Publish limit reached for today")
	}

	fmt.Printf("--------CAN PUBLISH-----------")

	// 1. Upload media / Create containers
	if len(files) == 0 {
		return "", fmt.Errorf("No files attached")
	}
	if len(files) == 1 {
		containerID, err = i.createContainer(accessToken, instagramID, caption, files[0], false)
		if err != nil {
			return "", err
		}
	} else {
		containerID, err = i.createCarouselContainer(accessToken, instagramID, caption, files)
		if err != nil {
			return "", err
		}
	}

	fmt.Printf("--------CONTAINER CREATED: %s-----------", containerID)

	// 2. Check container status
	i.containerStatus(accessToken, containerID)

	fmt.Printf("--------CONTAINER STATUS CHECKED-----------")

	// 3. Publish media

	return "", nil
}
