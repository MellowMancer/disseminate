package repo_instagram

import (
	repo_instagram "backend/repositories/instagram"
	"context"
	"fmt"
	"golang.org/x/oauth2"
	"log"
	"mime"
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
	publishMedia(accessToken string, instagramID string, creationID string) (string, error)
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

func (i *instagramServiceImpl) checkPublishLimit(accessToken string, instagramID string) (bool, error) {
	return i.repo_instagram.CheckPublishLimit(accessToken, instagramID)
}

func (i *instagramServiceImpl) uploadMedia(file multipart.File, ext string, mimeType string) (string, error) {

	return i.repo_instagram.UploadMedia(file, ext, mimeType)
}

func (i *instagramServiceImpl) createContainer(accessToken string, instagramID string, caption string, file *multipart.FileHeader, isCarouselItem bool) (string, error) {
	log.Println("[CREATE_CONTAINER] --- Starting container creation ---")
	log.Printf("[CREATE_CONTAINER] --- InstagramID: %s, Caption length: %d, IsCarouselItem: %t ---", instagramID, len(caption), isCarouselItem)
	
	f, err := file.Open()
	if err != nil {
		log.Printf("[CREATE_CONTAINER] --- Failed to open uploaded file: %v", err)
		return "", fmt.Errorf("failed to open uploaded file: %w", err)
	}
	defer func() {
		err := f.Close()
		if err != nil {
			log.Printf("[CREATE_CONTAINER] --- Failed to close file: %v", err)
		}
	}()

	buffer := make([]byte, 512)
	bytesRead, err := f.Read(buffer)
	if err != nil {
		log.Printf("[CREATE_CONTAINER] --- Failed to read media file: %v", err)
		return "", fmt.Errorf("failed to read media file: %w", err)
	}
	mimeType := http.DetectContentType(buffer[:bytesRead])
	ext, err := getFileExtension(mimeType)
	if err != nil {
		// handle error, fallback extension if needed
		ext = ".jpg"
	}
	log.Printf("[CREATE_CONTAINER] --- Using file extension: %s", ext)
	log.Printf("[CREATE_CONTAINER] --- Detected media content type: %s", mimeType)

	mediaType := ""
	switch {
	case strings.HasPrefix(mimeType, "image/"):
		mediaType = "IMAGE"
	case strings.HasPrefix(mimeType, "video/"):
		mediaType = "VIDEO"
	default:
		log.Printf("[CREATE_CONTAINER] --- Unsupported media type: %s", mimeType)
		return "", fmt.Errorf("unsupported media type: %s", mimeType)
	}

	log.Println("[CREATE_CONTAINER] --- Uploading media...")
	mediaURL, err := i.uploadMedia(f, ext, mimeType)
	if err != nil {
		log.Printf("[CREATE_CONTAINER] --- Failed to upload media: %v", err)
		return "", fmt.Errorf("failed to upload media: %w", err)
	}
	log.Println("[CREATE_CONTAINER] --- Media uploaded successfully")

	containerID, err := i.repo_instagram.CreateContainer(accessToken, instagramID, caption, mediaURL, mediaType, isCarouselItem)
	if err != nil {
		log.Printf("[CREATE_CONTAINER] --- Failed to create media container: %v", err)
		return "", fmt.Errorf("failed to create media container: %w", err)
	}

	log.Printf("[CREATE_CONTAINER] --- Media container created successfully with ContainerID: %s", containerID)
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
	return i.repo_instagram.WaitForContainerReady(accessToken, containerID)
}

func (i *instagramServiceImpl) publishMedia(accessToken string, instagramID string, creationID string) (string, error) {
	return i.repo_instagram.PublishMedia(accessToken, instagramID, creationID)
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
	_, err = i.containerStatus(accessToken, containerID)
	if err != nil {
		return "", err
	}

	fmt.Printf("--------CONTAINER STATUS CHECKED-----------")

	// 3. Publish media
	postID, err := i.repo_instagram.PublishMedia(accessToken, instagramID, containerID)
	if err != nil {
		return "", err
	}

	fmt.Printf("--------MEDIA PUBLISHED: %s-----------", postID)

	// 4. Return Post URL

	return "", nil
}

func getFileExtension(mimeType string) (string, error) {
	exts, err := mime.ExtensionsByType(mimeType)
	if err != nil || len(exts) == 0 {
		return "", fmt.Errorf("no extension found for MIME type: %s", mimeType)
	}
	return exts[0], nil
}
