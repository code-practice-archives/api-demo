package logger

import (
	"context"
	"fmt"
	"os"

	"github.com/code-practice-archives/api-demo/internal/pkg/ctxkey"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

// Logger 对 zap 的薄封装：业务日志通过 WithContext 获取，自动携带 trace_id。
type Logger struct {
	base *zap.Logger
}

// New 基于配置创建 Logger：同时写入控制台与按大小/天数轮转的日志文件。
func New(cfg Config) (*Logger, error) {
	level, err := parseLevel(cfg.Level)
	if err != nil {
		return nil, err
	}

	encoderCfg := zap.NewProductionEncoderConfig()
	encoderCfg.EncodeTime = zapcore.ISO8601TimeEncoder
	encoderCfg.EncodeLevel = zapcore.CapitalLevelEncoder

	fileWriter := &lumberjack.Logger{
		Filename:   cfg.Filename,
		MaxSize:    cfg.MaxSize,
		MaxBackups: cfg.MaxBackups,
		MaxAge:     cfg.MaxAge,
		Compress:   cfg.Compress,
	}

	core := zapcore.NewTee(
		zapcore.NewCore(zapcore.NewConsoleEncoder(encoderCfg), zapcore.AddSync(os.Stdout), level),
		zapcore.NewCore(zapcore.NewJSONEncoder(encoderCfg), zapcore.AddSync(fileWriter), level),
	)

	return &Logger{base: zap.New(core, zap.AddCaller())}, nil
}

// Nop 返回丢弃所有日志的 Logger，用于测试。
func Nop() *Logger {
	return &Logger{base: zap.NewNop()}
}

// Named 创建带模块名的子 Logger。
func (l *Logger) Named(name string) *Logger {
	return &Logger{base: l.base.Named(name)}
}

// Zap 返回底层 *zap.Logger，供 gin-contrib/zap 等中间件使用。
func (l *Logger) Zap() *zap.Logger {
	if l == nil || l.base == nil {
		return zap.NewNop()
	}
	return l.base
}

// WithContext 绑定 ctx 中的 trace_id。业务层打日志应始终先调用此方法。
func (l *Logger) WithContext(ctx context.Context) *zap.Logger {
	if l == nil || l.base == nil {
		return zap.NewNop()
	}
	if ctx == nil {
		return l.base
	}
	if tid, ok := ctx.Value(ctxkey.TraceID).(string); ok && tid != "" {
		return l.base.With(zap.String(ctxkey.TraceID, tid))
	}
	return l.base
}

// Sync 刷新缓冲。
func (l *Logger) Sync() error {
	if l == nil || l.base == nil {
		return nil
	}
	return l.base.Sync()
}

func parseLevel(level string) (zapcore.Level, error) {
	var l zapcore.Level
	if err := l.UnmarshalText([]byte(level)); err != nil {
		return zapcore.InfoLevel, fmt.Errorf("invalid log.level %q: %w", level, err)
	}
	return l, nil
}
