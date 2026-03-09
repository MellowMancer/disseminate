package utils

import (
	"log"
	"net/http"

	"github.com/labstack/echo/v4"
)

// represents a fixed response structure
type StandardResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
}

// sends a successful JSON response
func SuccessResponse(c echo.Context, statusCode int, message string, data interface{}) error {
	return c.JSON(statusCode, StandardResponse{
		Success: true,
		Message: message,
		Data:    data,
	})
}

// sends an error JSON response
// logs the internal error details, exposes safe messages to clients
func ErrorResponse(c echo.Context, err error) error {
	// Default to internal server error
	statusCode := http.StatusInternalServerError
	userMessage := "An unexpected error occurred"

	if appErr, ok := err.(*AppError); ok {
		statusCode = appErr.StatusCode
		userMessage = appErr.Message

		// Log the full error for debugging
		if appErr.Internal != "" {
			log.Printf("[ERROR] %s | Internal: %s | Error: %v",
				userMessage, appErr.Internal, appErr.Err)
		} else {
			log.Printf("[ERROR] %s | Error: %v", userMessage, appErr.Err)
		}
	} else {
		// For non-AppError errors, log them but don't expose details
		log.Printf("[ERROR] Unexpected error: %v", err)
	}

	return c.JSON(statusCode, StandardResponse{
		Success: false,
		Error:   userMessage,
	})
}

// 400 Bad Request
func BadRequestResponse(c echo.Context, message string) error {
	return c.JSON(http.StatusBadRequest, StandardResponse{
		Success: false,
		Error:   message,
	})
}

// 401 Unauthorized response
func UnauthorizedResponse(c echo.Context, message string) error {
	return c.JSON(http.StatusUnauthorized, StandardResponse{
		Success: false,
		Error:   message,
	})
}

// 404 Not Found response
func NotFoundResponse(c echo.Context, message string) error {
	return c.JSON(http.StatusNotFound, StandardResponse{
		Success: false,
		Error:   message,
	})
}

// 500 Internal Server Error response
func InternalErrorResponse(c echo.Context, err error, context string) error {
	log.Printf("[ERROR] Internal error in %s: %v", context, err)
	return c.JSON(http.StatusInternalServerError, StandardResponse{
		Success: false,
		Error:   "An internal error occurred. Please try again later.",
	})
}
