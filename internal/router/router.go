package router

import (
	"github.com/code-practice-archives/api-demo/internal/handler"
	"github.com/code-practice-archives/api-demo/internal/middleware"
	"github.com/gin-gonic/gin"
)

func New(h handler.Handlers) *gin.Engine {
	r := gin.Default()
	r.Use(middleware.TraceID())

	api := r.Group("/api/v1")
	{
		auth := api.Group("/auth")
		auth.POST("/register", h.Auth.Register)
		auth.POST("/login", h.Auth.Login)
	}

	return r
}
