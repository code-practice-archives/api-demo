package middleware

import (
	"github.com/code-practice-archives/api-demo/internal/pkg/errcode"
	"github.com/code-practice-archives/api-demo/internal/pkg/ipallowlist"
	"github.com/code-practice-archives/api-demo/internal/pkg/response"
	"github.com/gin-gonic/gin"
)

// IPAllowlist 仅放行白名单内的 ClientIP；未启用时直接放行。
// 启用以空名单视为拒绝全部，避免误配导致私有接口对外暴露。
func IPAllowlist(cfg ipallowlist.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		if !cfg.Enabled {
			c.Next()
			return
		}

		if !ipallowlist.Allowed(c.ClientIP(), cfg.IPs) {
			response.Error(c, errcode.ErrForbidden)
			c.Abort()
			return
		}
		c.Next()
	}
}
