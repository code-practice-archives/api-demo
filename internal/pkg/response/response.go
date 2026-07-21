package response

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/code-practice-archives/api-demo/internal/pkg/errcode"
	"github.com/gin-gonic/gin"
)

const TraceIDKey = "trace_id"

type Response struct {
	Code    string `json:"code,omitempty"`
	Message string `json:"message,omitempty"`
	Data    any    `json:"data,omitempty"`
	TraceID string `json:"trace_id,omitempty"`
}

func Success(c *gin.Context, data any) {
	writeJSON(c.Writer, http.StatusOK, Response{
		Data:    data,
		TraceID: traceID(c),
	})
}

// Error 将 error 写入统一错误响应；能解析为 *errcode.Error 则按其码与状态返回，否则回落为内部错误。
func Error(c *gin.Context, err error) {
	var appErr *errcode.Error
	if !errors.As(err, &appErr) {
		appErr = errcode.ErrInternal
	}

	writeJSON(c.Writer, appErr.HTTPStatus, Response{
		Code:    appErr.Code,
		Message: appErr.Message,
		TraceID: traceID(c),
	})
}

func writeJSON(w http.ResponseWriter, status int, resp Response) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(resp)
}

func traceID(c *gin.Context) string {
	if v, ok := c.Get(TraceIDKey); ok {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}
