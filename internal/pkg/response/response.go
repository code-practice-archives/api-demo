package response

import (
	"errors"
	"net/http"

	"github.com/code-practice-archives/api-demo/internal/pkg/errcode"
	"github.com/gin-contrib/requestid"
	"github.com/gin-gonic/gin"
)

type Response struct {
	Code    string `json:"code,omitempty"`
	Message string `json:"message,omitempty"`
	Data    any    `json:"data,omitempty"`
	TraceID string `json:"trace_id,omitempty"`
}

func Success(c *gin.Context, data any) {
	c.JSON(http.StatusOK, Response{
		Data:    data,
		TraceID: requestid.Get(c),
	})
}

// Error 将 error 写入统一错误响应；能解析为 *errcode.Error 则按其码与状态返回，否则回落为内部错误。
func Error(c *gin.Context, err error) {
	var appErr *errcode.Error
	if !errors.As(err, &appErr) {
		appErr = errcode.ErrInternal
	}

	c.JSON(appErr.HTTPStatus, Response{
		Code:    appErr.Code,
		Message: appErr.Message,
		TraceID: requestid.Get(c),
	})
}
