package errcode

import "net/http"

// 认证相关错误码。
var (
	ErrUsernameTaken = &Error{
		Code:       "auth.username_taken",
		Message:    "username already taken",
		HTTPStatus: http.StatusConflict,
	}

	ErrInvalidCredentials = &Error{
		Code:       "auth.invalid_credentials",
		Message:    "username or password is incorrect",
		HTTPStatus: http.StatusUnauthorized,
	}
)
