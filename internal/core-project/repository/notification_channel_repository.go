// Package repository 提供项目通知渠道数据访问层
package repository

import (
	"context"
	"fmt"

	"github.com/kerbos/ticketdesk/internal/model"
	"gorm.io/gorm"
)

// NotificationChannelRepository 项目通知渠道数据访问接口
type NotificationChannelRepository interface {
	// Create 创建通知渠道
	Create(ctx context.Context, channel *model.ProjectNotificationChannel) error
	// GetByID 根据 ID 获取通知渠道
	GetByID(ctx context.Context, id uint64) (*model.ProjectNotificationChannel, error)
	// Update 更新通知渠道
	Update(ctx context.Context, channel *model.ProjectNotificationChannel) error
	// Delete 删除通知渠道
	Delete(ctx context.Context, id uint64) error
	// ListByProject 获取项目的所有通知渠道
	ListByProject(ctx context.Context, projectID uint64) ([]*model.ProjectNotificationChannel, error)
	// ListEnabledByProject 获取项目已启用的通知渠道
	ListEnabledByProject(ctx context.Context, projectID uint64) ([]*model.ProjectNotificationChannel, error)
}

// notificationChannelRepository 项目通知渠道数据访问实现
type notificationChannelRepository struct {
	db *gorm.DB
}

// NewNotificationChannelRepository 创建项目通知渠道数据访问实例
func NewNotificationChannelRepository(db *gorm.DB) NotificationChannelRepository {
	return &notificationChannelRepository{db: db}
}

// Create 创建通知渠道
func (r *notificationChannelRepository) Create(ctx context.Context, channel *model.ProjectNotificationChannel) error {
	if err := r.db.WithContext(ctx).Create(channel).Error; err != nil {
		return fmt.Errorf("创建通知渠道失败: %w", err)
	}
	return nil
}

// GetByID 根据 ID 获取通知渠道
func (r *notificationChannelRepository) GetByID(ctx context.Context, id uint64) (*model.ProjectNotificationChannel, error) {
	var channel model.ProjectNotificationChannel
	if err := r.db.WithContext(ctx).First(&channel, id).Error; err != nil {
		return nil, err
	}
	return &channel, nil
}

// Update 更新通知渠道
func (r *notificationChannelRepository) Update(ctx context.Context, channel *model.ProjectNotificationChannel) error {
	if err := r.db.WithContext(ctx).Save(channel).Error; err != nil {
		return fmt.Errorf("更新通知渠道失败: %w", err)
	}
	return nil
}

// Delete 删除通知渠道
func (r *notificationChannelRepository) Delete(ctx context.Context, id uint64) error {
	if err := r.db.WithContext(ctx).Delete(&model.ProjectNotificationChannel{}, id).Error; err != nil {
		return fmt.Errorf("删除通知渠道失败: %w", err)
	}
	return nil
}

// ListByProject 获取项目的所有通知渠道
func (r *notificationChannelRepository) ListByProject(ctx context.Context, projectID uint64) ([]*model.ProjectNotificationChannel, error) {
	var channels []*model.ProjectNotificationChannel
	if err := r.db.WithContext(ctx).Where("project_id = ?", projectID).Order("created_at DESC").Find(&channels).Error; err != nil {
		return nil, fmt.Errorf("查询通知渠道失败: %w", err)
	}
	return channels, nil
}

// ListEnabledByProject 获取项目已启用的通知渠道
func (r *notificationChannelRepository) ListEnabledByProject(ctx context.Context, projectID uint64) ([]*model.ProjectNotificationChannel, error) {
	var channels []*model.ProjectNotificationChannel
	if err := r.db.WithContext(ctx).Where("project_id = ? AND enabled = ?", projectID, true).Find(&channels).Error; err != nil {
		return nil, fmt.Errorf("查询已启用通知渠道失败: %w", err)
	}
	return channels, nil
}
