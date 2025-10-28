package user

import (
	"backend/models"
	repo_user "backend/repositories/user"
	repo_instagram "backend/repositories/instagram"
	repo_twitter "backend/repositories/twitter"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"
	"golang.org/x/crypto/bcrypt"
)

type UserService interface {
	CreateUser(user *models.User) error
	LoginUser(user *models.User) (string, error)
	GetJWTSecret() []byte
	IsLoggedIn(c echo.Context) (bool, string, error)
	SaveTwitterToken(email string, accessToken string, accessSecret string) error
	GetTwitterToken(email string) (string, string, error)
	SaveInstagramToken(email string, accessToken string, expiresIn int) error
	GetInstagramCredentials(email string) (string, string, error)
}

type userServiceImpl struct {
	repo_user      repo_user.UserRepository
	repo_instagram repo_instagram.InstagramRepository
	repo_twitter   repo_twitter.TwitterRepository
	jwtSecret      []byte
}

func NewUserService(repoUser repo_user.UserRepository, repoInstagram repo_instagram.InstagramRepository, repoTwitter repo_twitter.TwitterRepository, jwtSecret []byte) UserService {
	return &userServiceImpl{
		repo_user: repoUser,
		repo_instagram: repoInstagram,
		repo_twitter: repoTwitter,
		jwtSecret: []byte(jwtSecret),
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

func (s *userServiceImpl) IsLoggedIn(c echo.Context) (bool, string, error) {
	claims, ok := c.Get("userClaims").(jwt.MapClaims)
	if !ok {
		return false, "", fmt.Errorf("Invalid token claims")
	}

	email, ok := claims["sub"].(string)
	if !ok {
		return false, "", fmt.Errorf("Email (sub) claim not found in token")
	}

	exists, err := s.repo_user.ExistsByEmail(email)
	if err != nil || !exists {
		return false, "", err
	}
	return true, email, nil
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
	return s.repo_twitter.GetToken(userID)
}

func (s *userServiceImpl) SaveInstagramToken(email string, accessToken string, expiresIn int) error {
	userID, err := s.repo_user.UserIDByEmail(email)
	if err != nil {
		return err
	}

	expirationTime := time.Now().Add(time.Duration(expiresIn) * time.Second).UTC()

	instagramID, err := s.repo_instagram.GetInstagramID(accessToken)
	if err != nil {
		return err
	}

	return s.repo_instagram.SaveToken(userID, instagramID, accessToken, expirationTime)
}

func (s *userServiceImpl) GetInstagramCredentials(email string) (string, string, error) {
	userID, err := s.repo_user.UserIDByEmail(email)
	if err != nil {
		return "", "", err
	}

	return s.repo_instagram.GetCredentials(userID)
}
