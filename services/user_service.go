package services

import (
	"backend/models"
	"backend/repositories"
	"fmt"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"
	"golang.org/x/crypto/bcrypt"
)

const api_path = "/rest/v1/"
const profile_path = "profiles"

type UserIDResponse struct {
	ID string `json:"id"`
}

type UserService interface {
	CreateUser(user *models.User) error
	LoginUser(user *models.User) (string, error)
	GetJWTSecret() []byte
	SaveTwitterToken(email string, accessToken string, accessSecret string) error
	GetTwitterToken(email string) (string, string, error)
	IsLoggedIn(c echo.Context) (bool, string, error)
	SaveInstagramToken(email string, accessToken string, expiresIn int) error
	GetInstagramToken(email string) (string, error)
}

type userServiceImpl struct {
	repository *repositories.SupabaseRepository
	jwtSecret  []byte
}

func NewUserService(repo *repositories.SupabaseRepository, jwtSecret []byte) UserService {
	return &userServiceImpl{
		repository: repo,
		jwtSecret:  []byte(jwtSecret),
	}
}

func (s *userServiceImpl) CreateUser(user *models.User) error {
	exists, err := s.repository.ExistsByEmail(user.Email)
	if err != nil {
		return err
	}
	if exists {
		return fmt.Errorf("user already exists")
	}

	return s.repository.Create(user)
}

func (s *userServiceImpl) LoginUser(user *models.User) (string, error) {
	data, err := s.repository.FindByEmail(user.Email)
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
		return false, "", c.JSON(http.StatusUnauthorized, map[string]string{"error": "Invalid token claims"})
	}

	email, ok := claims["sub"].(string)
	if !ok {
		return false, "", c.JSON(http.StatusBadRequest, map[string]string{"error": "Email (sub) claim not found in token"})
	}

	exists, err := s.repository.ExistsByEmail(email)
	if err != nil || !exists {
		return false, "", err
	}
	return true, email, nil
}

func (s *userServiceImpl) SaveTwitterToken(email, accessToken, accessSecret string) error {
	userID, err := s.repository.UserIDByEmail(email)
	if err != nil {
		return err
	}

	return s.repository.SaveTwitterToken(userID, accessToken, accessSecret)
}

func (s *userServiceImpl) GetTwitterToken(email string) (string, string, error) {
	userID, err := s.repository.UserIDByEmail(email)
	if err != nil {
		return "", "", err
	}
	return s.repository.GetTwitterToken(userID)
}

func (s *userServiceImpl) SaveInstagramToken(email string, accessToken string, expiresIn int) error {
	userID, err := s.repository.UserIDByEmail(email)
	if err != nil {
		return err
	}

	expirationTime := time.Now().Add(time.Duration(expiresIn) * time.Second).UTC()

	instagramID, err := s.repository.GetInstagramID(accessToken)
	if err != nil {
		return err
	}

	return s.repository.SaveInstagramToken(userID, instagramID, accessToken, expirationTime)
}

func (s *userServiceImpl) GetInstagramToken(email string) (string, error) {

	userID, err := s.repository.UserIDByEmail(email)
	if err != nil {
		return "", err
	}

	return s.repository.GetInstagramToken(userID)
}
