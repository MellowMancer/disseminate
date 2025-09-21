package handlers

import (
    "net/http"

	"github.com/labstack/echo/v4"
	"golang.org/x/crypto/bcrypt"
	"backend/models"
	"backend/services"
)

type Handler struct {
    UserService services.UserService
}

func NewHandler(userService services.UserService) *Handler {
    return &Handler{UserService: userService}
}

func (h *Handler) SignUp(c echo.Context) error {
    var req models.User
    if err := c.Bind(&req); err != nil {
        return c.JSON(http.StatusBadRequest, "Invalid input")
    }

    hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
    if err != nil {
        return c.JSON(http.StatusInternalServerError, "Error hashing password")
    }
    req.Password = string(hashedPassword)

    if err := h.UserService.CreateUser(&req); err != nil {
        return c.JSON(http.StatusInternalServerError, "Failed to create user")
    }
    return c.JSON(http.StatusCreated, "User created")
}

func (h *Handler) Login(c echo.Context) error {
	return nil
}

func (h *Handler) Logout(c echo.Context) error {
	return nil
}

