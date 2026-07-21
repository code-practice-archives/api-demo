// Package validator 提供项目统一的结构体校验入口。
//
// 分层约定：
//   - Handler DTO：只用 json + binding 做传输层校验（能否解析、必填）
//   - Service Input：用 validate tag，方法入口先调用 Struct
//   - 跨字段或需查库的规则：留在 service 方法内，不塞进本包
//   - Config：继续用各自的 validate，不走 HTTP errcode
package validator

import (
	"errors"
	"fmt"
	"strings"

	"github.com/code-practice-archives/api-demo/internal/pkg/errcode"
	playground "github.com/go-playground/validator/v10"
)

var v = playground.New()

// Struct 校验带 validate tag 的结构体。
// 失败时返回 ErrInvalidArgument，消息取首个字段错误。
func Struct(s any) error {
	if err := v.Struct(s); err != nil {
		return errcode.ErrInvalidArgument.WithMessage(format(err))
	}
	return nil
}

func format(err error) string {
	var errs playground.ValidationErrors
	if !errors.As(err, &errs) || len(errs) == 0 {
		return "invalid argument"
	}

	fe := errs[0]
	field := strings.ToLower(fe.Field())
	switch fe.Tag() {
	case "required":
		return fmt.Sprintf("%s is required", field)
	case "min":
		return fmt.Sprintf("%s must be at least %s characters", field, fe.Param())
	case "max":
		return fmt.Sprintf("%s must be at most %s characters", field, fe.Param())
	default:
		return fmt.Sprintf("%s is invalid", field)
	}
}
