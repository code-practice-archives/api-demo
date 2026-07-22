package oauth

import "net/http"

// Error RFC 6749 风格的 OAuth 错误。
type Error struct {
	Code        string `json:"error"`
	Description string `json:"error_description,omitempty"`
	HTTPStatus  int    `json:"-"`
}

func (e *Error) Error() string {
	if e.Description != "" {
		return e.Code + ": " + e.Description
	}
	return e.Code
}

func newError(code, description string, status int) *Error {
	return &Error{Code: code, Description: description, HTTPStatus: status}
}

var (
	ErrInvalidRequest          = newError("invalid_request", "invalid request", http.StatusBadRequest)
	ErrInvalidClient           = newError("invalid_client", "invalid client", http.StatusUnauthorized)
	ErrInvalidGrant            = newError("invalid_grant", "invalid grant", http.StatusBadRequest)
	ErrUnsupportedGrantType    = newError("unsupported_grant_type", "unsupported grant type", http.StatusBadRequest)
	ErrUnsupportedResponseType = newError("unsupported_response_type", "unsupported response type", http.StatusBadRequest)
	ErrServerError             = newError("server_error", "server error", http.StatusInternalServerError)
)

// WithDescription 返回带自定义描述的错误副本。
func (e *Error) WithDescription(desc string) *Error {
	cp := *e
	cp.Description = desc
	return &cp
}
