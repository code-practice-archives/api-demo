package main

import (
	"context"
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/code-practice-archives/api-demo/internal/config"
	"github.com/code-practice-archives/api-demo/internal/server"
	"go.uber.org/zap"
)

// main 启动 HTTP 服务，并在收到退出信号后执行优雅关闭。
func main() {
	configPath := flag.String("config", config.DefaultConfigFile, "path to config file")
	flag.Parse()

	cfg, err := config.Load(*configPath)
	if err != nil {
		log.Fatalf("load config failed: %v", err)
	}

	srv, err := server.New(cfg)
	if err != nil {
		log.Fatalf("server init failed: %v", err)
	}
	zlog := srv.Log()

	// 1. 在独立 goroutine 中启动服务，避免 ListenAndServe 阻塞主流程，
	//    使主 goroutine 能继续监听系统信号。
	go func() {
		if err := srv.Start(); err != nil {
			zlog.WithContext(context.Background()).Fatal("server start failed", zap.Error(err))
		}
	}()

	// 2. 阻塞等待 SIGINT/SIGTERM（如 Ctrl+C 或容器终止），再进入关闭流程。
	//    channel 容量为 1，避免信号在尚未读取时被丢弃。
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	zlog.WithContext(context.Background()).Info("shutdown signal received")

	// 3. 限时优雅关闭：给进行中的请求最多 10 秒完成，超时则强制退出。
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Stop(ctx); err != nil {
		zlog.WithContext(ctx).Fatal("server shutdown failed", zap.Error(err))
	}

	zlog.WithContext(ctx).Info("server exited")
}
