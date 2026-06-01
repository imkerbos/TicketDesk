// Package service 提供每日日报的 cron 调度封装
package service

import (
	"context"
	"fmt"

	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/kerbos/ticketdesk/internal/model"
	"github.com/kerbos/ticketdesk/internal/scheduler"
	"github.com/kerbos/ticketdesk/pkg/logger"
)

// 默认 cron 与时区，新建项目或字段空时使用
const (
	defaultDigestCron = "0 9 * * *"
	defaultDigestTZ   = "Asia/Shanghai"
)

// DailyDigestScheduler 把项目级 cron 注册到 scheduler.Scheduler
type DailyDigestScheduler struct {
	sched     *scheduler.Scheduler
	db        *gorm.DB
	digestSvc DigestService
}

// NewDailyDigestScheduler 创建日报调度器
func NewDailyDigestScheduler(sched *scheduler.Scheduler, db *gorm.DB, digestSvc DigestService) *DailyDigestScheduler {
	return &DailyDigestScheduler{
		sched:     sched,
		db:        db,
		digestSvc: digestSvc,
	}
}

// LoadAll 启动时扫所有启用日报的项目并注册 cron
func (d *DailyDigestScheduler) LoadAll(ctx context.Context) error {
	var projects []model.Project
	if err := d.db.WithContext(ctx).Where("daily_digest_enabled = ?", true).Find(&projects).Error; err != nil {
		return fmt.Errorf("加载已启用日报项目失败: %w", err)
	}
	for i := range projects {
		d.Reload(&projects[i])
	}
	logger.Info("daily digest jobs loaded", zap.Int("count", len(projects)))
	return nil
}

// Reload 根据项目最新配置注册或注销 cron
// 实现 projectService.DigestScheduler 接口
func (d *DailyDigestScheduler) Reload(project *model.Project) {
	jobID := fmt.Sprintf("digest:%d", project.ID)
	if !project.DailyDigestEnabled {
		d.sched.Unregister(jobID)
		return
	}

	spec := project.DailyDigestCron
	if spec == "" {
		spec = defaultDigestCron
	}
	tz := project.DailyDigestTZ
	if tz == "" {
		tz = defaultDigestTZ
	}

	projectID := project.ID
	if err := d.sched.Register(jobID, spec, tz, func(ctx context.Context) {
		if err := d.digestSvc.RunForProjectScheduled(ctx, projectID); err != nil {
			logger.Warn("daily digest execution failed",
				zap.Uint64("project_id", projectID),
				zap.Error(err),
			)
		}
	}); err != nil {
		logger.Error("failed to register daily digest job",
			zap.Uint64("project_id", projectID),
			zap.String("spec", spec),
			zap.String("tz", tz),
			zap.Error(err),
		)
	}
}
