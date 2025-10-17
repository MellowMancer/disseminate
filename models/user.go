package models

import (
	"time"
)

type User struct {
	Email                   string     `json:"email"`
	Password                string     `json:"password"` // Store hashed password
	
}

type TwitterModel struct {
	UserID              string `json:"user_id"`
	AccessToken  string   `json:"access_token"`
	AccessSecret string   `json:"access_secret"`
}

type InstagramModel struct {
	UserID              string `json:"user_id"`
	InstagramID             string `json:"instagram_id"`
	AccessToken    string    `json:"access_token"`
	ExpiresAt time.Time `json:"expires_at"`
}