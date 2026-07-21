package handler

import (
	"github.com/code-practice-archives/api-demo/internal/pkg/errcode"
	"github.com/code-practice-archives/api-demo/internal/pkg/response"
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
		response.Error(c, errcode.ErrInvalidArgument.WithMessage("username and password are required"))
		return
	}

	result, err := h.auth.Register(c.Request.Context(), service.RegisterInput{
		Username: req.Username,
		Password: req.Password,
	})
	if err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, result)
}

func (h *AuthHandler) Login(c *gin.Context) {
	var req authRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, errcode.ErrInvalidArgument.WithMessage("username and password are required"))
		return
	}

	result, err := h.auth.Login(c.Request.Context(), service.LoginInput{
		Username: req.Username,
		Password: req.Password,
	})
	if err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, result)
}
