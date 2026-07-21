# ratelimit

基于 Redis 的固定窗口请求限流。业务层只关心「这个 key 还能不能过」，不直接拼 Redis 命令。

与 `loginjail` 互补：本包按 **IP（或任意 key）** 限制请求频率；`loginjail` 按 **用户名** 在登录失败后锁定账号。

## 设计

目标是收成一个可注入的 `Limiter`，与配置解耦、与 HTTP 框架解耦：

| 职责 | 归属 |
|------|------|
| 开关、阈值、窗口 | `Config`（从 YAML 加载） |
| 计数 / 判断是否放行 | `Limiter`（Redis 实现） |
| 取 `ClientIP`、返回 429 | `middleware.RateLimit`（调用方） |

刻意保持精简：

- **算法固定窗口**：窗口内 `INCR`，首次写入时 `EXPIRE`；实现简单，多实例共享同一 Redis 即可
- **存储固定 Redis**：生产与单测都走 Redis（单测用 miniredis）；关闭限流时用 `Noop`
- **key 由调用方决定**：本包不绑定 IP 或路由；中间件传入 `c.ClientIP()`
- **超限只返回 bool**：不抛业务错误码；HTTP 层映射为 `errcode.ErrTooManyRequests`

## 运作方式

```text
配置 rate_limit.enabled / limit / window_seconds
        │
        ▼
  NewRedis(rdb, limit, window)   或   Noop{}（enabled=false）
        │
        ▼
     Allow(ctx, key)
        │
        ├─ INCR ratelimit:{key}
        ├─ 若 count == 1 → EXPIRE 窗口时长
        └─ count <= limit → 放行，否则拒绝
```

**`Allow`**

1. Redis key 为 `ratelimit:{key}`（`key` 通常是客户端 IP）
2. `INCR` 得到当前窗口内计数
3. 第一次出现时设置 TTL 为 `window`
4. `count <= limit` 返回 `true`，否则 `false`

在本项目中的调用链：

```text
全局中间件 middleware.RateLimit
  → Limiter.Allow(ctx, ClientIP)
  → 超限 → 429 / common.too_many_requests
  → Redis 出错 → fail-open（放行，避免抖动拖垮服务）
```

## 配置

对应 `configs/*.yaml` 中的 `rate_limit` 段：

```yaml
rate_limit:
  enabled: true          # 默认关闭；需显式打开（bool 不用 default 标签，避免与零值冲突）
  limit: 120             # 窗口内同一 key 最多请求数；≤0 时回落为 120
  window_seconds: 60     # 固定窗口秒数；≤0 时回落为 60
```

| 字段 | 说明 |
|------|------|
| `enabled` | 为 `false` 时注入 `Noop`，始终放行 |
| `limit` | 窗口内允许的最大次数；经 `NewRedis` 校验后生效 |
| `window_seconds` | 窗口时长；经 `Window()` 转为 `time.Duration` |

Wire 注入示例（见 `cmd/server/providers.go`）：

```go
func provideRateLimiter(cfg ratelimit.Config, rdb *redis.Client) ratelimit.Limiter {
	if !cfg.Enabled {
		return ratelimit.Noop{}
	}
	return ratelimit.NewRedis(rdb, cfg.Limit, cfg.Window())
}
```

## 使用

### 中间件（本项目用法）

挂在全局 `r.Use` 上，按 IP 限流所有请求：

```go
r.Use(middleware.RateLimit(limiter))
```

### 直接调用

```go
allowed, err := limiter.Allow(ctx, clientIP)
if err != nil {
	// 调用方自行决定 fail-open / fail-closed
}
if !allowed {
	return errcode.ErrTooManyRequests
}
```

### 单测

用 miniredis，不依赖真实 Redis：

```go
mr := miniredis.RunT(t)
rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
limiter := ratelimit.NewRedis(rdb, 3, time.Minute)
```

## 文件

| 文件 | 内容 |
|------|------|
| `config.go` | `Config`、`Window()` |
| `limiter.go` | `Limiter` 接口、`Noop` |
| `redis.go` | Redis 固定窗口实现 |

## 约定

- 依赖已建立的 `*redis.Client`；本包不负责连接与关闭
- 固定窗口在边界处可能出现短时「双倍」突发（窗口切换瞬间），可接受则无需改算法
- `INCR` 成功但 `EXPIRE` 失败时 key 可能无 TTL；属极端情况，运维侧可定期扫描孤儿 key
- 与 `loginjail` 同时开启时：限流挡刷量，jail 挡撞库；不要用本包替代账号锁定
