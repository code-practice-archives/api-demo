package middleware

import (
	"time"

	"github.com/code-practice-archives/api-demo/internal/pkg/ctxkey"
	"github.com/code-practice-archives/api-demo/internal/pkg/logger"
	"github.com/gin-contrib/requestid"
	ginzap "github.com/gin-contrib/zap"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// AccessLog 记录每个 HTTP 请求的访问日志（方法、路径、状态码、耗时等；含 trace_id，鉴权后含 user_id）。
func AccessLog(log *logger.Logger) gin.HandlerFunc {
	return ginzap.GinzapWithConfig(log.Zap(), &ginzap.Config{
		TimeFormat: time.RFC3339,
		UTC:        true,
		Context: func(c *gin.Context) []zapcore.Field {
			var fields []zapcore.Field
			if tid := requestid.Get(c); tid != "" {
				fields = append(fields, zap.String(ctxkey.TraceID, tid))
			}
			if uid, ok := c.Get(ctxkey.UserID); ok {
				if id, ok := uid.(int64); ok {
					fields = append(fields, zap.Int64(ctxkey.UserID, id))
				}
			}
			return fields
		},
	})
}
