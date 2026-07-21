package middleware

import (
	"context"

	"github.com/code-practice-archives/api-demo/internal/pkg/ctxkey"
	"github.com/gin-contrib/requestid"
	"github.com/gin-gonic/gin"
)

// headerTraceID 为链路追踪 ID 的 HTTP 头名称。
// 客户端可主动传入以便跨服务关联；未传入时由服务端生成。
const headerTraceID = "X-Trace-ID"

// TraceID 注入请求级追踪 ID：写入 Gin Context、Request.Context 与响应头。
// Request.Context 中的值可供 logger.WithContext 自动携带到业务日志。
func TraceID() gin.HandlerFunc {
	return requestid.New(
		requestid.WithCustomHeaderStrKey(headerTraceID),
		requestid.WithHandler(func(c *gin.Context, id string) {
			c.Set(ctxkey.TraceID, id)
			c.Request = c.Request.WithContext(context.WithValue(
				c.Request.Context(),
				ctxkey.TraceID,
				id,
			))
		}),
	)
}
