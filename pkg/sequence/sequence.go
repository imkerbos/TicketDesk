// Package sequence 提供基于 Redis 的原子序号生成器
// 首次使用时通过 Lua 脚本从 DB 种子值初始化，保证并发安全
package sequence

import (
	"context"
	"fmt"

	"github.com/kerbos/ticketdesk/pkg/cache"
	"github.com/kerbos/ticketdesk/pkg/logger"
	pkgRedis "github.com/kerbos/ticketdesk/pkg/redis"
	"go.uber.org/zap"
)

// key 格式: issue:seq:{project_key}
func seqKey(projectKey string) string {
	return fmt.Sprintf("issue:seq:%s", projectKey)
}

// DBSeedFunc 从数据库获取当前最大序号的回调函数
type DBSeedFunc func(ctx context.Context, projectKey string) (int64, error)

// NextIssueNumber 获取下一个工单序号
// 优先使用 Redis INCR；首次使用（INCR 返回 1）时通过 Lua 脚本从 DB 种子值初始化
// Redis 不可用时降级到 dbFallback
func NextIssueNumber(ctx context.Context, projectKey string, dbFallback DBSeedFunc) (int64, error) {
	key := seqKey(projectKey)

	// 尝试 Redis INCR
	val, err := cache.Incr(ctx, key)
	if err != nil {
		// Redis 不可用，降级到 DB
		logger.Warn("sequence Redis unavailable, falling back to DB",
			zap.String("project_key", projectKey),
			zap.Error(err),
		)
		return dbFallback(ctx, projectKey)
	}

	// val == 1 说明 key 刚创建（之前不存在），需要从 DB 初始化种子值
	if val == 1 {
		seedVal, seedErr := initFromDB(ctx, key, projectKey, dbFallback)
		if seedErr != nil {
			// 初始化失败，删除 key 让下次重试，降级到 DB
			cache.Del(ctx, key)
			return dbFallback(ctx, projectKey)
		}
		return seedVal, nil
	}

	return val, nil
}

// initFromDB 使用 Lua 脚本原子地从 DB 种子值初始化 Redis 计数器
// Lua 脚本逻辑：如果当前值 == 1（说明是首次），则 SET 为 seed+1 并返回 seed+1
// 如果当前值 > 1（说明被并发请求抢先初始化了），则不操作，返回当前值
var initScript = `
local key = KEYS[1]
local seed = tonumber(ARGV[1])
local current = tonumber(redis.call('GET', key))
if current == 1 then
    local newVal = seed + 1
    redis.call('SET', key, newVal)
    return newVal
end
return current
`

func initFromDB(ctx context.Context, key, projectKey string, dbFallback DBSeedFunc) (int64, error) {
	client := pkgRedis.GetClient()
	if client == nil {
		return 0, cache.ErrRedisUnavailable
	}

	// 从 DB 获取当前最大序号
	maxSeq, err := dbFallback(ctx, projectKey)
	if err != nil {
		return 0, fmt.Errorf("failed to get DB seed for %s: %w", projectKey, err)
	}

	// dbFallback 返回的是"下一个序号"，所以种子值 = maxSeq - 1（即当前最大值）
	// 但如果 DB 没有记录，dbFallback 返回 1，此时种子值 = 0
	seed := maxSeq - 1
	if seed < 0 {
		seed = 0
	}

	result, err := client.Eval(ctx, initScript, []string{key}, seed).Int64()
	if err != nil {
		return 0, fmt.Errorf("failed to run init script for %s: %w", projectKey, err)
	}

	logger.Info("sequence initialized from DB seed",
		zap.String("project_key", projectKey),
		zap.Int64("seed", seed),
		zap.Int64("next_number", result),
	)

	return result, nil
}
