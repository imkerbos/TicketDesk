package cache

import "errors"

// ErrRedisUnavailable Redis 客户端不可用
var ErrRedisUnavailable = errors.New("redis client unavailable")
