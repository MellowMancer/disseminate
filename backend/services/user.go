package services

import ("encoding/json"
    "net/http"
	"fmt"
    "bytes"
    "backend/models"
	)

type UserService interface {
    CreateUser(user *models.User) error
}

type userServiceImpl struct {
    supabaseURL   string
    supabaseKey   string
    httpClient    *http.Client
}

func NewUserService(url, key string) UserService {
    return &userServiceImpl{
        supabaseURL: url,
        supabaseKey: key,
        httpClient:  &http.Client{},
    }
}

func (s *userServiceImpl) CreateUser(user *models.User) error {
	exists, err := s.UserExists(user.Email)
    if err != nil {
        return err
    }
    if exists {
        return fmt.Errorf("user already exists")
    }


    payloadBytes, _ := json.Marshal(user)

    req, _ := http.NewRequest("POST", s.supabaseURL+"/users", bytes.NewBuffer(payloadBytes))
    req.Header.Set("apikey", s.supabaseKey)
    req.Header.Set("Authorization", "Bearer "+s.supabaseKey)
    req.Header.Set("Content-Type", "application/json")
    req.Header.Set("Prefer", "return=representation")

    resp, err := s.httpClient.Do(req)
    if err != nil || resp.StatusCode != 201 {
        return err
    }
    defer resp.Body.Close()
    return nil
}


func (s *userServiceImpl) UserExists(email string) (bool, error) {
    req, _ := http.NewRequest("GET", s.supabaseURL+"/users?email=eq."+email, nil)
    req.Header.Set("apikey", s.supabaseKey)
    req.Header.Set("Authorization", "Bearer "+s.supabaseKey)

    resp, err := s.httpClient.Do(req)
    if err != nil {
        return false, err
    }
    defer resp.Body.Close()

    var users []models.User
    json.NewDecoder(resp.Body).Decode(&users)

    return len(users) > 0, nil
}