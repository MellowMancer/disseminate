package models

import ()

type User struct {
	Email              string `json:"email"`
	Password           string `json:"password"` // Store hashed password
	TwitterAccessToken  *string `json:"twitter_access_token"`
	TwitterAccessSecret *string `json:"twitter_access_secret"`
}
