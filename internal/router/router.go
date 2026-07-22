package router

import (
	"github.com/code-practice-archives/api-demo/internal/handler"
	"github.com/code-practice-archives/api-demo/internal/middleware"
	"github.com/code-practice-archives/api-demo/internal/pkg/ipallowlist"
	"github.com/code-practice-archives/api-demo/internal/pkg/jwtx"
	"github.com/code-practice-archives/api-demo/internal/pkg/logger"
	"github.com/code-practice-archives/api-demo/internal/pkg/ratelimit"
	"github.com/gin-gonic/gin"
)

func New(h handler.Handlers, jwt *jwtx.Manager, log *logger.Logger, limiter ratelimit.Limiter, allowlist ipallowlist.Config) *gin.Engine {
	r := gin.New()
	r.Use(
		gin.Recovery(),
		middleware.TraceID(),
		middleware.AccessLog(log),
		middleware.RateLimit(limiter),
	)

	api := r.Group("/api/v1")
	{
		auth := api.Group("/auth")
		auth.POST("/register", h.Auth.Register)
		auth.POST("/login", h.Auth.Login)
		auth.POST("/refresh", h.Auth.Refresh)
		auth.POST("/logout", h.Auth.Logout)

		api.GET("/me", middleware.Auth(jwt), h.Auth.Me)

		// 私有接口：仅白名单 IP 可访问
		private := api.Group("/private", middleware.IPAllowlist(allowlist))
		private.GET("/ping", h.Private.Ping)
	}

	// OAuth 2.0 授权服务器（Authorization Code + PKCE）
	oauthGroup := r.Group("/oauth")
	{
		oauthGroup.POST("/authorize", middleware.Auth(jwt), h.OAuth.Authorize)
		oauthGroup.POST("/token", h.OAuth.Token)
	}

	return r
}
