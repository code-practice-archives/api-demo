package errcode

import "net/http"

// 通用错误码。
var (
	ErrInvalidArgument = &Error{
		Code:       "common.invalid_argument",
		Message:    "invalid argument",
		HTTPStatus: http.StatusBadRequest,
	}

	ErrInternal = &Error{
		Code:       "common.internal",
		Message:    "internal server error",
		HTTPStatus: http.StatusInternalServerError,
	}
)
