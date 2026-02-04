// Package repository 提供活动日志数据访问层
package repository

import (
	"context"

	"github.com/kerbos/ticketdesk/internal/model"
	"gorm.io/gorm"
)

// ActivityRepository 活动日志仓储接口
type ActivityRepository interface {
	// Create 创建活动日志
	Create(ctx context.Context, activity *model.ActivityLog) error
	// List 分页查询活动日志
	List(ctx context.Context, offset, limit int, entityType, entityKey string, entityID, userID uint64) ([]*model.ActivityLog, int64, error)
	// GetRecent 获取最近的活动日志
	GetRecent(ctx context.Context, limit int) ([]*model.ActivityLog, error)
}

// activityRepository 活动日志仓储实现
type activityRepository struct {
	db *gorm.DB
}

// NewActivityRepository 创建活动日志仓储实例
func NewActivityRepository(db *gorm.DB) ActivityRepository {
	return &activityRepository{db: db}
}

// Create 创建活动日志
func (r *activityRepository) Create(ctx context.Context, activity *model.ActivityLog) error {
	return r.db.WithContext(ctx).Create(activity).Error
}

// List 分页查询活动日志
func (r *activityRepository) List(ctx context.Context, offset, limit int, entityType, entityKey string, entityID, userID uint64) ([]*model.ActivityLog, int64, error) {
	var activities []*model.ActivityLog
	var total int64

	query := r.db.WithContext(ctx).Model(&model.ActivityLog{})

	// 筛选条件
	if entityType != "" {
		query = query.Where("entity_type = ?", entityType)
	}
	if entityKey != "" {
		query = query.Where("entity_key = ?", entityKey)
	}
	if entityID > 0 {
		query = query.Where("entity_id = ?", entityID)
	}
	if userID > 0 {
		query = query.Where("user_id = ?", userID)
	}

	// 统计总数
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 分页查询
	if err := query.Order("created_at DESC").Offset(offset).Limit(limit).Find(&activities).Error; err != nil {
		return nil, 0, err
	}

	return activities, total, nil
}

// GetRecent 获取最近的活动日志
func (r *activityRepository) GetRecent(ctx context.Context, limit int) ([]*model.ActivityLog, error) {
	var activities []*model.ActivityLog

	if err := r.db.WithContext(ctx).
		Order("created_at DESC").
		Limit(limit).
		Find(&activities).Error; err != nil {
		return nil, err
	}

	return activities, nil
}
