package utils

import (
	"crypto/rand"
	"encoding/base64"
)

// GenerateOAuthState generates a cryptographically secure random state string for OAuth flows
func GenerateOAuthState() (string, error) {
	b := make([]byte, 32) // 32 bytes = 256 bits of entropy
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}
	state := base64.URLEncoding.EncodeToString(b)
	return state, nil
}
