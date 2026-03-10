package utils

import (
	"fmt"
	"github.com/golang-jwt/jwt/v5"
)

// ValidateJWT validates a JWT token string and returns the claims
func ValidateJWT(tokenString string, secret []byte) (jwt.MapClaims, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// Validate signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return secret, nil
	})

	if err != nil {
		return nil, err
	}

	if !token.Valid {
		return nil, ErrInvalidToken
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, ErrInvalidToken
	}

	return claims, nil
}

// ExtractEmailFromClaims extracts the email (sub claim) from JWT claims
func ExtractEmailFromClaims(claims jwt.MapClaims) (string, error) {
	email, ok := claims["sub"].(string)
	if !ok || email == "" {
		return "", ErrInvalidToken
	}
	return email, nil
}
