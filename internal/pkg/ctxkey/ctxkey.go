package ctxkey

// 请求上下文键：Gin Context 与 context.Context 共用同一套字符串，避免各包各自定义。
const (
	// 链路追踪
	TraceID = "trace_id"
	// 用户身份
	UserID = "user_id"
	// 用户名
	Username = "username"
)
