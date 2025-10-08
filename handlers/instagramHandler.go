package handlers

import (
	"backend/services"

	"github.com/labstack/echo/v4"
	_"net/http"
)

type InstagramHandler struct {
	instagramService services.InstagramService
	userService    services.UserService
}

func NewInstagramHandler(instagramService services.InstagramService, userService services.UserService) *InstagramHandler {
	return &InstagramHandler{
		instagramService: instagramService,
		userService:    userService,
	}
}

// BeginInstagramLink initiates the Instagram OAuth linking process.
func (h *InstagramHandler) BeginInstagramLink(c echo.Context) error {
	responseWriter := c.Response().Writer
	request := c.Request()
	h.instagramService.HandleLogin(responseWriter, request)
	return nil
}

// CheckInstagramToken checks the validity of the Instagram token.
func (h *InstagramHandler) CheckInstagramToken(c echo.Context) error {
	// Implementation for checking Instagram token
	return nil
}