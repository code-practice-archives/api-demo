package handler

import (
	"errors"
	"net/http"

	"github.com/code-practice-archives/api-demo/internal/pkg/ctxkey"
	"github.com/code-practice-archives/api-demo/internal/pkg/oauth"
	"github.com/code-practice-archives/api-demo/internal/service"
	"github.com/gin-gonic/gin"
)

type OAuthHandler struct {
	svc *service.Services
}

func NewOAuthHandler(svc *service.Services) *OAuthHandler {
	return &OAuthHandler{svc: svc}
}

type authorizeRequest struct {
	ClientID            string `json:"client_id" binding:"required"`
	RedirectURI         string `json:"redirect_uri" binding:"required"`
	ResponseType        string `json:"response_type" binding:"required"`
	State               string `json:"state"`
	Scope               string `json:"scope"`
	CodeChallenge       string `json:"code_challenge" binding:"required"`
	CodeChallengeMethod string `json:"code_challenge_method" binding:"required"`
}

type authorizeResponse struct {
	Code        string `json:"code"`
	State       string `json:"state,omitempty"`
	RedirectURI string `json:"redirect_uri"`
}

type tokenResponse struct {
	AccessToken  string `json:"access_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int64  `json:"expires_in"`
	RefreshToken string `json:"refresh_token,omitempty"`
	Scope        string `json:"scope,omitempty"`
}

// Authorize POST /oauth/authorize — 纯 API 授权码发放（需已登录 Bearer）。
func (h *OAuthHandler) Authorize(c *gin.Context) {
	userID, ok := c.Get(ctxkey.UserID)
	if !ok {
		oauthError(c, oauth.ErrInvalidRequest.WithDescription("authentication required"))
		return
	}
	id, ok := userID.(int64)
	if !ok {
		oauthError(c, oauth.ErrInvalidRequest.WithDescription("authentication required"))
		return
	}

	var req authorizeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		oauthError(c, oauth.ErrInvalidRequest.WithDescription("invalid request body"))
		return
	}

	result, err := h.svc.OAuth.Authorize(c, service.AuthorizeInput{
		UserID:              id,
		ClientID:            req.ClientID,
		RedirectURI:         req.RedirectURI,
		ResponseType:        req.ResponseType,
		State:               req.State,
		Scope:               req.Scope,
		CodeChallenge:       req.CodeChallenge,
		CodeChallengeMethod: req.CodeChallengeMethod,
	})
	if err != nil {
		oauthError(c, err)
		return
	}

	c.JSON(http.StatusOK, authorizeResponse{
		Code:        result.Code,
		State:       result.State,
		RedirectURI: result.RedirectURI,
	})
}

// Token POST /oauth/token — application/x-www-form-urlencoded。
func (h *OAuthHandler) Token(c *gin.Context) {
	if err := c.Request.ParseForm(); err != nil {
		oauthError(c, oauth.ErrInvalidRequest.WithDescription("invalid form body"))
		return
	}

	result, err := h.svc.OAuth.Token(c, service.TokenInput{
		GrantType:    c.PostForm("grant_type"),
		Code:         c.PostForm("code"),
		RedirectURI:  c.PostForm("redirect_uri"),
		ClientID:     c.PostForm("client_id"),
		ClientSecret: c.PostForm("client_secret"),
		CodeVerifier: c.PostForm("code_verifier"),
		RefreshToken: c.PostForm("refresh_token"),
	})
	if err != nil {
		oauthError(c, err)
		return
	}

	c.JSON(http.StatusOK, tokenResponse{
		AccessToken:  result.AccessToken,
		TokenType:    result.TokenType,
		ExpiresIn:    result.ExpiresIn,
		RefreshToken: result.RefreshToken,
		Scope:        result.Scope,
	})
}

func oauthError(c *gin.Context, err error) {
	var oe *oauth.Error
	if !errors.As(err, &oe) {
		oe = oauth.ErrServerError
	}
	c.JSON(oe.HTTPStatus, gin.H{
		"error":             oe.Code,
		"error_description": oe.Description,
	})
}
