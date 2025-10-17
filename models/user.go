package models

import "time"

type User struct {
	Email              string `json:"email"`
	Password           string `json:"password"` // Store hashed password
	TwitterAccessToken  *string `json:"twitter_access_token"`
	TwitterAccessSecret *string `json:"twitter_access_secret"`
	InstagramAccessToken *string `json:"instagram_access_token"`
	InstagramTokenExpiresAt     *time.Time `json:"instagram_token_expires_at"`
}
