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

	// 无效 / 过期 / 已吊销的 refresh token 统一映射，避免泄露细节。
	ErrInvalidRefreshToken = &Error{
		Code:       "auth.invalid_refresh_token",
		Message:    "invalid refresh token",
		HTTPStatus: http.StatusUnauthorized,
	}
)
