package utils

import (
	// "crypto/rand"
	// "encoding/base64"
	"errors"
	"fmt"
)

// Error types
var (
	// Authentication errors
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrUnauthorized       = errors.New("unauthorized")
	ErrTokenExpired       = errors.New("token expired")
	ErrInvalidToken       = errors.New("invalid token")

	// Validation errors
	ErrInvalidInput       = errors.New("invalid input")
	ErrInvalidEmail       = errors.New("invalid email address")
	ErrWeakPassword       = errors.New("password does not meet strength requirements")
	ErrInvalidPlatform    = errors.New("invalid platform specified")

	// User errors
	ErrUserExists         = errors.New("user already exists")
	ErrUserNotFound       = errors.New("user not found")
	ErrAccountNotLinked   = errors.New("account not linked")

	// OAuth errors
	ErrOAuthStateMismatch = errors.New("oauth state validation failed")
	ErrOAuthFailed        = errors.New("oauth authentication failed")
	ErrTokenExchange      = errors.New("failed to exchange oauth token")

	// Platform errors
	ErrPostFailed         = errors.New("failed to post to platform")
	ErrMediaUploadFailed  = errors.New("failed to upload media")
	ErrInvalidMediaType   = errors.New("invalid media type")
	ErrFileTooLarge       = errors.New("file size exceeds limit")

	// Database errors
	ErrDatabaseOperation  = errors.New("database operation failed")

	// Internal errors
	ErrInternal           = errors.New("internal server error")
)

// Represents application error in structured format
type AppError struct {
	Err        error
	Message    string
	StatusCode int    // HTTP status code
	Internal   string // Internal details (not exposed to client)
}

func (e *AppError) Error() string {
	if e.Internal != "" {
		return fmt.Sprintf("%s: %s (internal: %s)", e.Message, e.Err.Error(), e.Internal)
	}
	return fmt.Sprintf("%s: %s", e.Message, e.Err.Error())
}

func (e *AppError) Unwrap() error {
	return e.Err
}

// creates a new application error
func NewAppError(err error, message string, statusCode int, internal string) *AppError {
	return &AppError{
		Err:        err,
		Message:    message,
		StatusCode: statusCode,
		Internal:   internal,
	}
}

// wraps an error with a message
func WrapError(err error, message string, statusCode int) *AppError {
	return &AppError{
		Err:        err,
		Message:    message,
		StatusCode: statusCode,
	}
}
