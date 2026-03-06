// Package cache 提供 Redis 缓存操作的薄封装
// 统一 nil 守卫、JSON 序列化、错误日志；Redis 不可用时静默降级
package cache

import (
	"context"
	"encoding/json"
	"time"

	"go.uber.org/zap"

	"github.com/kerbos/ticketdesk/pkg/logger"
	"github.com/kerbos/ticketdesk/pkg/redis"
)

// Get 获取字符串值，返回 (值, 是否命中)
// Redis 不可用或 key 不存在时返回 ("", false)
func Get(ctx context.Context, key string) (string, bool) {
	client := redis.GetClient()
	if client == nil {
		return "", false
	}

	val, err := client.Get(ctx, key).Result()
	if err != nil {
		// key 不存在也是正常 miss，不打日志
		return "", false
	}
	return val, true
}

// Set 设置字符串值
// Redis 不可用时静默跳过
func Set(ctx context.Context, key, value string, ttl time.Duration) {
	client := redis.GetClient()
	if client == nil {
		return
	}

	if err := client.Set(ctx, key, value, ttl).Err(); err != nil {
		logger.Error("cache Set failed", zap.String("key", key), zap.Error(err))
	}
}

// Del 删除一个或多个 key
// Redis 不可用时静默跳过
func Del(ctx context.Context, keys ...string) {
	client := redis.GetClient()
	if client == nil {
		return
	}

	if err := client.Del(ctx, keys...).Err(); err != nil {
		logger.Error("cache Del failed", zap.Strings("keys", keys), zap.Error(err))
	}
}

// Incr 原子递增，返回递增后的值
// Redis 不可用时返回 error 供调用方降级到 DB
func Incr(ctx context.Context, key string) (int64, error) {
	client := redis.GetClient()
	if client == nil {
		return 0, ErrRedisUnavailable
	}

	return client.Incr(ctx, key).Result()
}

// SetNX 当 key 不存在时设置值，返回是否设置成功
// Redis 不可用时返回 false
func SetNX(ctx context.Context, key string, value interface{}, ttl time.Duration) bool {
	client := redis.GetClient()
	if client == nil {
		return false
	}

	ok, err := client.SetNX(ctx, key, value, ttl).Result()
	if err != nil {
		logger.Error("cache SetNX failed", zap.String("key", key), zap.Error(err))
		return false
	}
	return ok
}

// GetJSON 从缓存获取 JSON 并反序列化到 dest，返回是否命中
// Redis 不可用或反序列化失败时返回 false
func GetJSON[T any](ctx context.Context, key string, dest *T) bool {
	val, ok := Get(ctx, key)
	if !ok {
		return false
	}

	if err := json.Unmarshal([]byte(val), dest); err != nil {
		logger.Error("cache GetJSON unmarshal failed", zap.String("key", key), zap.Error(err))
		return false
	}
	return true
}

// SetJSON 将值序列化为 JSON 并写入缓存
// Redis 不可用或序列化失败时静默跳过
func SetJSON[T any](ctx context.Context, key string, value T, ttl time.Duration) {
	data, err := json.Marshal(value)
	if err != nil {
		logger.Error("cache SetJSON marshal failed", zap.String("key", key), zap.Error(err))
		return
	}

	Set(ctx, key, string(data), ttl)
}
