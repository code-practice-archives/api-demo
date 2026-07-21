package handler

import (
	"github.com/code-practice-archives/api-demo/internal/pkg/errcode"
	"github.com/code-practice-archives/api-demo/internal/pkg/response"
	"github.com/code-practice-archives/api-demo/internal/service"
	"github.com/gin-gonic/gin"
)

type AuthHandler struct {
	svc *service.Services
}

func NewAuthHandler(svc *service.Services) *AuthHandler {
	return &AuthHandler{svc: svc}
}

type authRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type authResponse struct {
	Token string           `json:"token"`
	User  authUserResponse `json:"user"`
}

type authUserResponse struct {
	ID        int64  `json:"id"`
	Username  string `json:"username"`
	CreatedAt int64  `json:"created_at"`
	UpdatedAt int64  `json:"updated_at"`
}

func (h *AuthHandler) Register(c *gin.Context) {
	var req authRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, errcode.ErrInvalidArgument)
		return
	}

	result, err := h.svc.Auth.Register(c, service.RegisterInput{
		Username: req.Username,
		Password: req.Password,
	})
	if err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, toAuthResponse(result))
}

func (h *AuthHandler) Login(c *gin.Context) {
	var req authRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, errcode.ErrInvalidArgument)
		return
	}

	result, err := h.svc.Auth.Login(c, service.LoginInput{
		Username: req.Username,
		Password: req.Password,
	})
	if err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, toAuthResponse(result))
}

func toAuthResponse(result *service.AuthResult) authResponse {
	return authResponse{
		Token: result.Token,
		User: authUserResponse{
			ID:        result.User.Id,
			Username:  result.User.Username,
			CreatedAt: result.User.CreatedAt,
			UpdatedAt: result.User.UpdatedAt,
		},
	}
}
