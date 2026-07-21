package middleware

import (
	"strings"

	"github.com/code-practice-archives/api-demo/internal/pkg/ctxkey"
	"github.com/code-practice-archives/api-demo/internal/pkg/errcode"
	"github.com/code-practice-archives/api-demo/internal/pkg/jwtx"
	"github.com/code-practice-archives/api-demo/internal/pkg/response"
	"github.com/gin-gonic/gin"
)

// Auth 校验 Authorization: Bearer <token>，通过后将用户身份写入 gin.Context。
// 下游 handler 用 ctxkey.UserID / Username 读取；任意失败统一返回 Unauthorized，避免泄露令牌细节。
func Auth(jwt *jwtx.Manager) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 1. 解析 Bearer Token（不接受 Basic 等其它方案）
		header := c.GetHeader("Authorization")
		if header == "" {
			response.Error(c, errcode.ErrUnauthorized)
			c.Abort()
			return
		}

		const prefix = "Bearer "
		if !strings.HasPrefix(header, prefix) {
			response.Error(c, errcode.ErrUnauthorized)
			c.Abort()
			return
		}

		tokenStr := strings.TrimSpace(header[len(prefix):])
		if tokenStr == "" {
			response.Error(c, errcode.ErrUnauthorized)
			c.Abort()
			return
		}

		// 2. 验签并解析 claims（过期、篡改、算法不符均视为未授权）
		claims, err := jwt.Parse(tokenStr)
		if err != nil {
			response.Error(c, errcode.ErrUnauthorized)
			c.Abort()
			return
		}

		// 3. 注入身份后放行；Abort 未调用时才会进入后续 handler
		c.Set(ctxkey.UserID, claims.UserID)
		c.Set(ctxkey.Username, claims.Username)
		c.Next()
	}
}
