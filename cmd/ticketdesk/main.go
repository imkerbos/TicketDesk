// Package main 应用入口
package main

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/kerbos/ticketdesk/internal/api/router"
	"github.com/kerbos/ticketdesk/internal/model"
	"github.com/kerbos/ticketdesk/pkg/config"
	"github.com/kerbos/ticketdesk/pkg/database"
	"github.com/kerbos/ticketdesk/pkg/jwt"
	"github.com/kerbos/ticketdesk/pkg/logger"
	"github.com/kerbos/ticketdesk/pkg/redis"
	"go.uber.org/zap"
)

var configPath string

func init() {
	flag.StringVar(&configPath, "config", "configs/config.yaml", "配置文件路径")
}

func main() {
	flag.Parse()

	// 加载配置
	cfg, err := config.Load(configPath)
	if err != nil {
		fmt.Printf("Failed to load config: %v\n", err)
		os.Exit(1)
	}

	// 初始化日志
	if err := logger.Init(&cfg.Log); err != nil {
		fmt.Printf("Failed to init logger: %v\n", err)
		os.Exit(1)
	}
	defer func() {
		if err := logger.Sync(); err != nil {
			fmt.Printf("Failed to sync logger: %v\n", err)
		}
	}()

	logger.Info("starting ticketdesk",
		zap.String("env", cfg.App.Env),
		zap.Int("port", cfg.App.Port),
	)

	// 初始化数据库
	if err := database.Init(&cfg.Database, cfg.App.Debug); err != nil {
		logger.Fatal("failed to init database", zap.Error(err))
	}
	defer func() {
		if err := database.Close(); err != nil {
			logger.Error("failed to close database", zap.Error(err))
		}
	}()

	// 自动迁移数据库
	if err := model.AutoMigrate(database.GetDB()); err != nil {
		logger.Fatal("failed to auto migrate database", zap.Error(err))
	}

	// 初始化种子数据
	if err := model.SeedData(database.GetDB()); err != nil {
		logger.Fatal("failed to seed data", zap.Error(err))
	}

	// 初始化 Redis
	if err := redis.Init(&cfg.Redis); err != nil {
		logger.Fatal("failed to init redis", zap.Error(err))
	}
	defer func() {
		if err := redis.Close(); err != nil {
			logger.Error("failed to close redis", zap.Error(err))
		}
	}()

	// 初始化 JWT 管理器
	jwtManager := jwt.NewManager(&cfg.JWT)

	// 设置路由
	appRouter := router.NewRouter(cfg, jwtManager, database.GetDB())
	r := appRouter.Setup()

	// 创建 HTTP 服务器
	srv := &http.Server{
		Addr:    fmt.Sprintf(":%d", cfg.App.Port),
		Handler: r,
	}

	// 启动服务器
	go func() {
		logger.Info("server started", zap.String("addr", srv.Addr))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("failed to start server", zap.Error(err))
		}
	}()

	// 优雅关闭
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("shutting down server...")

	// 停止所有数据源轮询器
	appRouter.StopPollers()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		logger.Error("server forced to shutdown", zap.Error(err))
	}

	logger.Info("server exited")
}
