// Package scheduler 提供基于 cron 的定时任务调度
package scheduler

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/robfig/cron/v3"
	"go.uber.org/zap"

	"github.com/kerbos/ticketdesk/pkg/logger"
)

// JobFunc 调度任务执行函数
type JobFunc func(ctx context.Context)

// Scheduler 简单的 cron 调度器，按 jobID 注册/注销
type Scheduler struct {
	mu      sync.Mutex
	cron    *cron.Cron
	entries map[string]cron.EntryID
}

// NewScheduler 创建调度器实例（cron 表达式支持 5 段：分 时 日 月 周）
func NewScheduler() *Scheduler {
	return &Scheduler{
		cron:    cron.New(cron.WithLogger(cron.PrintfLogger(noopPrinter{}))),
		entries: make(map[string]cron.EntryID),
	}
}

// Start 启动调度
func (s *Scheduler) Start() {
	s.cron.Start()
	logger.Info("scheduler started")
}

// Stop 停止调度，等待已运行任务完成
func (s *Scheduler) Stop() {
	ctx := s.cron.Stop()
	<-ctx.Done()
	logger.Info("scheduler stopped")
}

// Register 注册或替换一个任务；timezone 为空则用 UTC
// spec 例如 "0 9 * * *" 代表每天 9 点；timezone 例如 "Asia/Shanghai"
func (s *Scheduler) Register(jobID, spec, timezone string, fn JobFunc) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// 已有则先移除
	if old, ok := s.entries[jobID]; ok {
		s.cron.Remove(old)
		delete(s.entries, jobID)
	}

	loc := time.UTC
	if timezone != "" {
		l, err := time.LoadLocation(timezone)
		if err != nil {
			return fmt.Errorf("加载时区 %s 失败: %w", timezone, err)
		}
		loc = l
	}

	// 用 TZ= 前缀让 robfig/cron 按指定时区解析
	tzSpec := fmt.Sprintf("CRON_TZ=%s %s", loc.String(), spec)
	id, err := s.cron.AddFunc(tzSpec, func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
		defer cancel()
		defer func() {
			if r := recover(); r != nil {
				logger.Error("scheduler job panic",
					zap.String("job_id", jobID),
					zap.Any("recover", r),
				)
			}
		}()
		fn(ctx)
	})
	if err != nil {
		return fmt.Errorf("注册 cron 任务失败: %w", err)
	}

	s.entries[jobID] = id
	logger.Info("scheduler job registered",
		zap.String("job_id", jobID),
		zap.String("spec", spec),
		zap.String("timezone", loc.String()),
	)
	return nil
}

// Unregister 注销指定任务；不存在时静默返回
func (s *Scheduler) Unregister(jobID string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if id, ok := s.entries[jobID]; ok {
		s.cron.Remove(id)
		delete(s.entries, jobID)
		logger.Info("scheduler job unregistered", zap.String("job_id", jobID))
	}
}

// noopPrinter 实现 cron.Printer，吞掉默认日志（用项目自身日志代替）
type noopPrinter struct{}

func (noopPrinter) Printf(string, ...any) {}
