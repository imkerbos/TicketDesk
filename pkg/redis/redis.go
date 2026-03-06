// Package redis 提供 Redis 客户端管理
package redis

import (
	"context"
	"fmt"

	"github.com/redis/go-redis/v9"

	"github.com/kerbos/ticketdesk/pkg/config"
)

// Client 全局 Redis 客户端
var Client *redis.Client

// Init 初始化 Redis 连接
func Init(cfg *config.RedisConfig) error {
	Client = redis.NewClient(&redis.Options{
		Addr:     cfg.Addr(),
		Password: cfg.Password,
		DB:       cfg.DB,
		PoolSize: cfg.PoolSize,
	})

	// 测试连接
	ctx := context.Background()
	if err := Client.Ping(ctx).Err(); err != nil {
		return fmt.Errorf("failed to connect redis: %w", err)
	}

	return nil
}

// GetClient 获取 Redis 客户端
func GetClient() *redis.Client {
	return Client
}

// Close 关闭 Redis 连接
func Close() error {
	if Client != nil {
		return Client.Close()
	}
	return nil
}
