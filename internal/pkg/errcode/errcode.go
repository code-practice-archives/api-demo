package errcode

type Error struct {
	Code       string
	Message    string
	HTTPStatus int
}

func (e *Error) Error() string {
	return e.Message
}

// Is 按错误码比较，忽略 Message 差异，便于 errors.Is 匹配 WithMessage 副本。
func (e *Error) Is(target error) bool {
	t, ok := target.(*Error)
	if !ok {
		return false
	}
	return e.Code == t.Code
}

// WithMessage 返回带自定义文案的副本，错误码与 HTTP 状态不变。
func (e *Error) WithMessage(msg string) *Error {
	cp := *e
	cp.Message = msg
	return &cp
}
