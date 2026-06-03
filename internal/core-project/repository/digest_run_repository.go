// Package repository 提供每日日报执行记录数据访问
package repository

import (
	"context"
	"fmt"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/kerbos/ticketdesk/internal/model"
)

// DigestRunRepository 每日日报执行记录数据访问接口
// 用于多副本部署下的去重：唯一键 (project_id, run_date) + INSERT IGNORE 抢占
// 故意不提供 Release：抢到即发，发送失败不释放也不补发，避免渠道半成功后重发
type DigestRunRepository interface {
	// TryClaim 原子占位：返回 true 表示当前实例抢到了发送权，false 表示已被其他实例抢占
	TryClaim(ctx context.Context, projectID uint64, runDate string) (bool, error)
	// UpdateIssueCount 抢占后回写工单数（审计字段，幂等）
	UpdateIssueCount(ctx context.Context, projectID uint64, runDate string, count int) error
}

// digestRunRepository 实现
type digestRunRepository struct {
	db *gorm.DB
}

// NewDigestRunRepository 创建实例
func NewDigestRunRepository(db *gorm.DB) DigestRunRepository {
	return &digestRunRepository{db: db}
}

// TryClaim 通过 INSERT ... ON CONFLICT DO NOTHING 实现原子抢占
// RowsAffected == 1：抢到；== 0：已存在（其他实例已抢占）
func (r *digestRunRepository) TryClaim(ctx context.Context, projectID uint64, runDate string) (bool, error) {
	row := &model.DailyDigestRun{
		ProjectID: projectID,
		RunDate:   runDate,
	}
	res := r.db.WithContext(ctx).
		Clauses(clause.OnConflict{DoNothing: true}).
		Create(row)
	if res.Error != nil {
		return false, fmt.Errorf("抢占日报执行记录失败: %w", res.Error)
	}
	return res.RowsAffected == 1, nil
}

// UpdateIssueCount 回写发送的工单数；不存在记录时静默
func (r *digestRunRepository) UpdateIssueCount(ctx context.Context, projectID uint64, runDate string, count int) error {
	if err := r.db.WithContext(ctx).
		Model(&model.DailyDigestRun{}).
		Where("project_id = ? AND run_date = ?", projectID, runDate).
		Update("issue_count", count).Error; err != nil {
		return fmt.Errorf("更新日报工单数失败: %w", err)
	}
	return nil
}
