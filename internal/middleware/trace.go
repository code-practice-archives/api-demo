package middleware

import (
	"github.com/code-practice-archives/api-demo/internal/pkg/response"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// headerTraceID 为链路追踪 ID 的 HTTP 头名称。
// 客户端可主动传入以便跨服务关联；未传入时由服务端生成。
const headerTraceID = "X-Trace-ID"

// TraceID 注入请求级追踪 ID，写入 Gin Context 并回写响应头，
// 供后续 handler / 统一响应体读取，便于日志与排错关联。
func TraceID() gin.HandlerFunc {
	return func(c *gin.Context) {
		traceID := c.GetHeader(headerTraceID)
		if traceID == "" {
			// 无上游透传时本地生成，保证每个请求都有唯一标识
			traceID = uuid.NewString()
		}

		c.Set(response.TraceIDKey, traceID)
		c.Writer.Header().Set(headerTraceID, traceID)
		c.Next()
	}
}
