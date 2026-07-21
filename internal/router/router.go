package router

import (
	"github.com/code-practice-archives/api-demo/internal/handler"
	"github.com/code-practice-archives/api-demo/internal/middleware"
	"github.com/code-practice-archives/api-demo/internal/pkg/jwtx"
	"github.com/code-practice-archives/api-demo/internal/pkg/logger"
	"github.com/gin-gonic/gin"
)

func New(h handler.Handlers, jwt *jwtx.Manager, log *logger.Logger) *gin.Engine {
	r := gin.New()
	r.Use(
		gin.Recovery(),
		middleware.TraceID(),
		middleware.AccessLog(log),
	)

	api := r.Group("/api/v1")
	{
		auth := api.Group("/auth")
		auth.POST("/register", h.Auth.Register)
		auth.POST("/login", h.Auth.Login)

		api.GET("/me", middleware.Auth(jwt), h.Auth.Me)
	}

	return r
}
