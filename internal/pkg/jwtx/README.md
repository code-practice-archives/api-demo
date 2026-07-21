# jwtx

JWT 签发与校验封装。业务层只关心「给谁签发」和「令牌是否有效」，不直接依赖 `golang-jwt` 的细节。

## 设计

目标是把认证令牌收成一个可注入的 `Manager`，与配置解耦、与 HTTP 框架解耦：

| 职责 | 归属 |
|------|------|
| 密钥、过期时间 | `Config`（从 YAML 加载） |
| 签发 / 验签 | `Manager` |
| 从 Header 取 Token、注入用户身份 | `middleware.Auth`（调用方） |

刻意保持精简：

- **算法固定 HS256**：对称密钥，部署简单；`Parse` 会拒绝其它 `alg`，降低算法混淆风险
- **Claims 只带业务身份**：`user_id` + `username`，再加标准 `exp` / `iat`
- **错误统一**：校验失败一律返回 `ErrInvalidToken`，避免把「过期 / 篡改 / 算法不符」等细节泄露给客户端
- **无状态**：不存会话、不支持主动吊销；过期即失效，需要更短有效期靠配置调 `expire_hours`

## 运作方式

```text
配置 jwt.secret / jwt.expire_hours
        │
        ▼
  NewManager(secret, expire)
        │
   ┌────┴────┐
   ▼         ▼
 Sign()    Parse()
   │         │
   │         ├─ 校验签名算法必须为 HS256
   │         ├─ 用同一 secret 验签
   │         └─ 检查 Valid（含过期）
   ▼
 HS256 签名的 JWT 字符串
```

**签发（`Sign`）**

1. 写入自定义字段 `UserID`、`Username`
2. 设置 `IssuedAt` 为当前时间，`ExpiresAt` 为 `now + expire`
3. 用 HS256 + secret 签名，返回 token 字符串

**校验（`Parse`）**

1. 按 `Claims` 解析，并强制算法为 HS256
2. 任一环节失败（签名错误、过期、claims 类型不对、`Valid == false`）→ `ErrInvalidToken`
3. 成功则返回 `*Claims`，供中间件写入 `gin.Context`

在本项目中的调用链：

```text
登录/注册成功 → AuthService 调 Manager.Sign
请求受保护接口 → middleware.Auth 调 Manager.Parse → 写入 ctxkey.UserID / Username
```

## 配置

对应 `configs/*.yaml` 中的 `jwt` 段：

```yaml
jwt:
  secret: "dev-secret-change-me"  # 必填；生产环境务必更换为足够长的随机串
  expire_hours: 24                # 可选；≤0 时回落为 24 小时
```

| 字段 | 说明 |
|------|------|
| `secret` | HMAC 密钥；`Validate()` 要求非空 |
| `expire_hours` | Token 有效小时数；经 `Expire()` 转为 `time.Duration` |

Wire 注入示例（见 `cmd/server/providers.go`）：

```go
func provideJWTManager(cfg jwtx.Config) *jwtx.Manager {
	return jwtx.NewManager(cfg.Secret, cfg.Expire())
}
```

## 使用

### 签发

登录或注册成功后：

```go
token, err := jwtMgr.Sign(user.Id, user.Username)
if err != nil {
	return nil, err
}
```

### 校验

中间件或其它需要验签的地方：

```go
claims, err := jwtMgr.Parse(tokenStr)
if err != nil {
	// errors.Is(err, jwtx.ErrInvalidToken)
	return // 按未授权处理
}
_ = claims.UserID
_ = claims.Username
```

### 单测

不依赖配置文件，直接构造短过期 Manager：

```go
jwtMgr := jwtx.NewManager("test-secret", time.Hour)
```

## 文件

| 文件 | 内容 |
|------|------|
| `config.go` | `Config`、`Expire()`、`Validate()` |
| `jwt.go` | `Claims`、`Manager`、`Sign` / `Parse`、`ErrInvalidToken` |

## 约定

- 同一进程内签发与校验必须使用**同一** `secret`；轮换密钥会使旧 Token 全部失效
- 调用方不要把 `Parse` 的底层错误原样返回给客户端，统一映射为未授权即可
- 本包不负责从 `Authorization` Header 拆 `Bearer`；那是 middleware 的职责
