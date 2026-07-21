# errcode

业务错误码定义。业务层直接 `return errcode.ErrXxx`，由 `response.Error` 通过 `errors.As` 解析后写入统一响应。

## Code 命名

格式：

```text
{module}.{error}[.{detail}]
```

- 全小写，段与段之间用 `.`，段内用 `_`
- **默认两段**；同一类错误需要再细分时再加第三段
- `module` 与文件划分一致（如 `auth`、`common`）
- `error` / `detail` 描述「发生了什么」，不要写笼统的失败结果（如 `register_failed`）
- 不要把 HTTP 状态写进 Code（状态用 `HTTPStatus`）

示例：

| Code | 说明 |
|------|------|
| `auth.username_taken` | 用户名已被占用 |
| `auth.invalid_credentials` | 账号或密码错误 |
| `auth.invalid_refresh_token` | refresh token 无效 / 过期 / 已吊销 |
| `common.invalid_argument` | 参数不合法 |
| `common.internal` | 未识别的内部错误 |

## 文件划分

按业务模块拆分文件，同一包名 `errcode`：

- `errcode.go`：`Error` 类型与方法
- `common.go`：通用错误
- `auth.go`：认证相关错误
- 新增模块时新增对应文件，例如 `user.go`

## 定义方式

预定义指针对象，抛错处直接返回：

```go
var (
	ErrUsernameTaken = &Error{
		Code:       "auth.username_taken",
		Message:    "username already taken",
		HTTPStatus: http.StatusConflict,
	}
)

return nil, errcode.ErrUsernameTaken
```

需要覆盖文案时用 `WithMessage`（Code / HTTPStatus 不变）：

```go
return errcode.ErrInvalidArgument.WithMessage("username and password are required")
```

## 约定

- **Code**：给机器与客户端做稳定契约，一经对外尽量不改
- **Message**：给人看的说明，可用 `WithMessage` 按场景调整
- **HTTPStatus**：对应响应状态码
- `errors.Is` 按 Code 比较，因此 `WithMessage` 后的副本仍可匹配原错误
- 无法解析为 `*errcode.Error` 的 error，统一回落为 `common.internal`
