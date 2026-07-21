package server

import (
	"github.com/gin-gonic/gin"

	"github.com/code-practice-archives/api-demo/internal/handler"
)

func New() *gin.Engine {
	r := gin.Default()
	handler.Register(r)
	return r
}
