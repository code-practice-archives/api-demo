package middleware

import (
	"github.com/code-practice-archives/api-demo/internal/pkg/errcode"
	"github.com/code-practice-archives/api-demo/internal/pkg/ratelimit"
	"github.com/code-practice-archives/api-demo/internal/pkg/response"
	"github.com/gin-gonic/gin"
)

// RateLimit 按 ClientIP 限流；超限返回 429。
// Limiter 故障时 fail-open，避免 Redis 抖动拖垮服务。
func RateLimit(limiter ratelimit.Limiter) gin.HandlerFunc {
	return func(c *gin.Context) {
		if limiter == nil {
			c.Next()
			return
		}

		allowed, err := limiter.Allow(c.Request.Context(), c.ClientIP())
		if err != nil {
			c.Next()
			return
		}
		if !allowed {
			response.Error(c, errcode.ErrTooManyRequests)
			c.Abort()
			return
		}
		c.Next()
	}
}
