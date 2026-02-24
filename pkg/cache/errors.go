package cache

import "errors"

// ErrRedisUnavailable Redis 客户端不可用
var ErrRedisUnavailable = errors.New("redis client unavailable")

// ErrLockNotAcquired 分布式锁获取失败（已被其他实例持有）
var ErrLockNotAcquired = errors.New("distributed lock not acquired")
