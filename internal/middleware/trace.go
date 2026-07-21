package middleware

import (
	"github.com/code-practice-archives/api-demo/internal/pkg/logger"
	"github.com/code-practice-archives/api-demo/internal/pkg/response"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// headerTraceID 为链路追踪 ID 的 HTTP 头名称。
// 客户端可主动传入以便跨服务关联；未传入时由服务端生成。
const headerTraceID = "X-Trace-ID"

// TraceID 注入请求级追踪 ID：写入 Gin Context、Request.Context 与响应头。
// Request.Context 中的值可供 logger.FromContext 自动携带到业务日志。
func TraceID() gin.HandlerFunc {
	return func(c *gin.Context) {
		traceID := c.GetHeader(headerTraceID)
		if traceID == "" {
			// 无上游透传时本地生成，保证每个请求都有唯一标识
			traceID = uuid.NewString()
		}

		c.Set(response.TraceIDKey, traceID)
		c.Request = c.Request.WithContext(logger.WithTraceID(c.Request.Context(), traceID))
		c.Writer.Header().Set(headerTraceID, traceID)
		c.Next()
	}
}
