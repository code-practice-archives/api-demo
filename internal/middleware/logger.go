package middleware

import (
	"time"

	"github.com/code-practice-archives/api-demo/internal/pkg/logger"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// AccessLog 记录每个 HTTP 请求的访问日志（方法、路径、状态码、耗时等；trace_id 由 WithContext 自动携带）。
func AccessLog(log *logger.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		query := c.Request.URL.RawQuery

		c.Next()

		latency := time.Since(start)
		status := c.Writer.Status()
		reqLog := log.WithContext(c.Request.Context())

		fields := []zap.Field{
			zap.String("method", c.Request.Method),
			zap.String("path", path),
			zap.String("query", query),
			zap.Int("status", status),
			zap.Duration("latency", latency),
			zap.String("client_ip", c.ClientIP()),
		}

		if len(c.Errors) > 0 {
			fields = append(fields, zap.String("errors", c.Errors.ByType(gin.ErrorTypePrivate).String()))
		}

		switch {
		case status >= 500:
			reqLog.Error("access", fields...)
		case status >= 400:
			reqLog.Warn("access", fields...)
		default:
			reqLog.Info("access", fields...)
		}
	}
}
