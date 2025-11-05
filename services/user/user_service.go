package user

import (
	"backend/models"
	repo_instagram "backend/repositories/instagram"
	repo_twitter "backend/repositories/twitter"
	repo_user "backend/repositories/user"
	"fmt"
	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"
	"golang.org/x/crypto/bcrypt"
	"time"
)

type UserService interface {
	CreateUser(user *models.User) error
	LoginUser(user *models.User) (string, error)
	GetJWTSecret() []byte
	IsLoggedIn(c echo.Context) (string, error)
	SaveTwitterToken(email string, accessToken string, accessSecret string) error
	GetTwitterToken(email string) (string, string, error)
	SaveInstagramToken(email string, accessToken string, expiresIn int) error
	GetInstagramCredentials(email string) (string, string, error)
	GetOAuthLinkStatus(jwtToken string) (OAuthStatus, error)
}

type userServiceImpl struct {
	repo_user      repo_user.UserRepository
	repo_instagram repo_instagram.InstagramRepository
	repo_twitter   repo_twitter.TwitterRepository
	jwtSecret      []byte
}

type OAuthStatus struct {
	Twitter    bool
	Instagram  bool
	Bluesky    bool
	Mastodon   bool
	Artstation bool
	Youtube    bool
}

func NewUserService(repoUser repo_user.UserRepository, repoInstagram repo_instagram.InstagramRepository, repoTwitter repo_twitter.TwitterRepository, jwtSecret []byte) UserService {
	return &userServiceImpl{
		repo_user:      repoUser,
		repo_instagram: repoInstagram,
		repo_twitter:   repoTwitter,
		jwtSecret:      []byte(jwtSecret),
	}
}

func (s *userServiceImpl) CreateUser(user *models.User) error {
	exists, err := s.repo_user.ExistsByEmail(user.Email)
	if err != nil {
		return err
	}
	if exists {
		return fmt.Errorf("user already exists")
	}

	return s.repo_user.Create(user)
}

func (s *userServiceImpl) LoginUser(user *models.User) (string, error) {
	data, err := s.repo_user.FindByEmail(user.Email)
	if err != nil {
		return "", err
	}

	err = bcrypt.CompareHashAndPassword([]byte(data.Password), []byte(user.Password))
	if err != nil {
		return "", fmt.Errorf("invalid password")
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub": data.Email,
		"exp": time.Now().Add(72 * time.Hour).Unix(),
	})

	tokenString, err := token.SignedString([]byte(s.jwtSecret))
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

func (s *userServiceImpl) GetJWTSecret() []byte {
	return s.jwtSecret
}

func (s *userServiceImpl) IsLoggedIn(c echo.Context) (string, error) {
	cookie, err := c.Cookie("jwt_token")
	if err != nil {
		return "", err
	}

	token, err := jwt.Parse(cookie.Value, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, echo.ErrUnauthorized
		}
		jwtSecret := s.GetJWTSecret()
		return jwtSecret, nil
	})
	if token == nil || !token.Valid {
		return "", err
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return "", fmt.Errorf("Could not find claims")
	}

	email, ok := claims["sub"].(string)
	if !ok {
		return "", fmt.Errorf("Could not find email")
	}
	return email, nil
}

func (s *userServiceImpl) SaveTwitterToken(email, accessToken, accessSecret string) error {
	userID, err := s.repo_user.UserIDByEmail(email)
	if err != nil {
		return err
	}

	return s.repo_twitter.SaveToken(userID, accessToken, accessSecret)
}

func (s *userServiceImpl) GetTwitterToken(email string) (string, string, error) {
	userID, err := s.repo_user.UserIDByEmail(email)
	if err != nil {
		return "", "", err
	}
	return s.repo_twitter.GetCredentials(userID)
}

func (s *userServiceImpl) SaveInstagramToken(email string, accessToken string, expiresIn int) error {
	userID, err := s.repo_user.UserIDByEmail(email)
	if err != nil {
		return err
	}

	expirationTime := time.Now().Add(time.Duration(expiresIn) * time.Second).UTC()
	//convert to string
	expirationTimeStr := expirationTime.Format(time.RFC3339)

	instagramID, err := s.repo_instagram.GetInstagramID(accessToken)
	if err != nil {
		return err
	}

	return s.repo_instagram.SaveToken(accessToken, userID, instagramID, expirationTimeStr)
}

func (s *userServiceImpl) GetInstagramCredentials(email string) (string, string, error) {
	userID, err := s.repo_user.UserIDByEmail(email)
	if err != nil {
		return "", "", err
	}

	return s.repo_instagram.GetCredentials(userID)
}

func (s *userServiceImpl) GetOAuthLinkStatus(email string) (OAuthStatus, error) {
	var status OAuthStatus

	userID, err := s.repo_user.UserIDByEmail(email)
	if err != nil {
		return status, err
	}

	status.Twitter = s.checkTwitter(userID)
	status.Instagram = s.checkInstagram(userID)

	// Placeholders for future platforms
	status.Bluesky = false
	status.Mastodon = false
	status.Artstation = false
	status.Youtube = false

	return status, nil
}

func (s *userServiceImpl) checkTwitter(userID string) bool {
	twitterAT, twitterAS, err := s.repo_twitter.GetCredentials(userID)
	if err != nil {
		return false
	}
	return s.repo_twitter.CheckTokens(twitterAT, twitterAS) == nil
}

func (s *userServiceImpl) checkInstagram(userID string) bool {
	instagramAT, _, err := s.repo_instagram.GetCredentials(userID)
	if err != nil {
		return false
	}
	return s.repo_instagram.CheckTokens(instagramAT) == nil
}
