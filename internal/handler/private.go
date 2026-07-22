package handler

import (
	"github.com/code-practice-archives/api-demo/internal/pkg/response"
	"github.com/gin-gonic/gin"
)

// PrivateHandler 提供受 IP 白名单保护的假私有接口，仅用于演示。
type PrivateHandler struct{}

func NewPrivateHandler() *PrivateHandler {
	return &PrivateHandler{}
}

// Ping 返回固定 mock 数据，便于验证白名单是否生效。
func (h *PrivateHandler) Ping(c *gin.Context) {
	response.Success(c, gin.H{
		"message":   "private api ok",
		"client_ip": c.ClientIP(),
	})
}
