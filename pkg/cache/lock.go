// Package cache 提供分布式锁工具
package cache

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"time"

	"github.com/kerbos/ticketdesk/pkg/logger"
	"github.com/kerbos/ticketdesk/pkg/redis"
	"go.uber.org/zap"
)

// generateLockValue 生成随机锁值，用于所有权验证
func generateLockValue() string {
	b := make([]byte, 16)
	rand.Read(b)
	return hex.EncodeToString(b)
}

// TryLock 尝试获取分布式锁，成功返回锁值（非空字符串）
// 调用方必须在完成后调用 UnlockWithValue 释放锁
func TryLock(ctx context.Context, key string, ttl time.Duration) string {
	client := redis.GetClient()
	if client == nil {
		return ""
	}

	value := generateLockValue()
	ok, err := client.SetNX(ctx, key, value, ttl).Result()
	if err != nil {
		logger.Error("cache TryLock SetNX failed", zap.String("key", key), zap.Error(err))
		return ""
	}
	if !ok {
		return ""
	}
	return value
}

// UnlockWithValue 释放分布式锁（仅当锁值匹配时才释放，防止误删他人的锁）
func UnlockWithValue(ctx context.Context, key string, value string) {
	client := redis.GetClient()
	if client == nil {
		return
	}

	// Lua 脚本保证 GET + DEL 的原子性
	script := `
		if redis.call("GET", KEYS[1]) == ARGV[1] then
			return redis.call("DEL", KEYS[1])
		else
			return 0
		end
	`
	client.Eval(ctx, script, []string{key}, value)
}

// Unlock 释放分布式锁（简单版，不验证所有权，仅用于向后兼容）
func Unlock(ctx context.Context, key string) {
	Del(ctx, key)
}

// WithLock 在分布式锁保护下执行 fn，获取失败返回 ErrLockNotAcquired
// 锁自动在 fn 执行完毕后释放
func WithLock(ctx context.Context, key string, ttl time.Duration, fn func() error) error {
	value := TryLock(ctx, key, ttl)
	if value == "" {
		return ErrLockNotAcquired
	}
	defer UnlockWithValue(ctx, key, value)
	return fn()
}

// TryLockWithRetry 带重试的分布式锁获取
// maxRetries: 最大重试次数，retryInterval: 重试间隔
// 返回锁值（非空表示成功），调用方必须用 UnlockWithValue 释放
func TryLockWithRetry(ctx context.Context, key string, ttl time.Duration, maxRetries int, retryInterval time.Duration) string {
	// Redis 不可用时直接降级放行，避免浪费重试时间
	client := redis.GetClient()
	if client == nil {
		logger.Warn("distributed lock degraded: redis unavailable, allowing through",
			zap.String("key", key))
		return "degraded"
	}

	if value := TryLock(ctx, key, ttl); value != "" {
		return value
	}

	for i := 0; i < maxRetries; i++ {
		select {
		case <-ctx.Done():
			return ""
		case <-time.After(retryInterval):
			if value := TryLock(ctx, key, ttl); value != "" {
				return value
			}
		}
	}

	return ""
}
