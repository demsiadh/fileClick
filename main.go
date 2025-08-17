package main

import (
	"context"
	"errors"
	"fileClick/api"
	"fileClick/config"
	"fileClick/system"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	// 1. 初始化 Engine
	var err error
	system.RankEngine, err = system.NewEngine()
	if err != nil {
		config.Error("init engine failed: %v", err)
	}

	// 2. 数据恢复
	if err := system.RankEngine.Recover(); err != nil {
		config.Error("recover failed: %v", err)
	}

	// 3.启动后台调度器
	system.RankEngine.StartScheduler()

	// 4.配置HTTP路由
	webSever := api.InitRouter()

	// 5. 优雅退出
	go func() {
		stop := make(chan os.Signal, 1)
		signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
		<-stop
		config.Info("Shutting down HTTP server...")
		// 停止 Engine
		system.RankEngine.Stop()
		config.Info("Engine stopped")

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = webSever.Shutdown(ctx)
		config.Info("Http server stopped")
	}()

	// 6.启动Http服务
	config.Info("HTTP server started at :8080")
	if err := webSever.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		config.Error("http listen failed: %v", err)
	}
}
