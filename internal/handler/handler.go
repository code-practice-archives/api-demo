package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func Register(r *gin.Engine) {
	r.GET("/ping", Ping)
}

func Ping(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "pong"})
}
