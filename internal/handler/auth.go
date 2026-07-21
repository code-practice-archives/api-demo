package handler

import (
	"errors"
	"net/http"

	"github.com/code-practice-archives/api-demo/internal/service"
	"github.com/gin-gonic/gin"
)

type AuthHandler struct {
	auth *service.AuthService
}

func NewAuthHandler(auth *service.AuthService) *AuthHandler {
	return &AuthHandler{auth: auth}
}

type authRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

func (h *AuthHandler) Register(c *gin.Context) {
	var req authRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "username and password are required"})
		return
	}

	result, err := h.auth.Register(service.RegisterInput{
		Username: req.Username,
		Password: req.Password,
	})
	if err != nil {
		writeAuthError(c, err)
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"token": result.Token,
		"user":  result.User,
	})
}

func (h *AuthHandler) Login(c *gin.Context) {
	var req authRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "username and password are required"})
		return
	}

	result, err := h.auth.Login(service.LoginInput{
		Username: req.Username,
		Password: req.Password,
	})
	if err != nil {
		writeAuthError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"token": result.Token,
		"user":  result.User,
	})
}

func writeAuthError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, service.ErrInvalidInput):
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "username must be 3-64 chars, password must be 6-72 chars",
		})
	case errors.Is(err, service.ErrUsernameTaken):
		c.JSON(http.StatusConflict, gin.H{"error": "username already taken"})
	case errors.Is(err, service.ErrInvalidCredentials):
		c.JSON(http.StatusUnauthorized, gin.H{"error": "username or password is incorrect"})
	default:
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
	}
}
