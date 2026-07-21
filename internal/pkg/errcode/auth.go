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

	ErrAccountLocked = &Error{
		Code:       "auth.account_locked",
		Message:    "too many failed login attempts, try again later",
		HTTPStatus: http.StatusTooManyRequests,
	}

	ErrUnauthorized = &Error{
		Code:       "auth.unauthorized",
		Message:    "unauthorized",
		HTTPStatus: http.StatusUnauthorized,
	}
)
