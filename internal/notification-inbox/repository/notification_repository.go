// Package repository 提供站内通知数据访问层
package repository

import (
	"context"

	"gorm.io/gorm"

	"github.com/kerbos/ticketdesk/internal/model"
)

// NotificationRepository 通知数据访问接口
type NotificationRepository interface {
	Create(ctx context.Context, notification *model.Notification) error
	GetByID(ctx context.Context, id uint64) (*model.Notification, error)
	ListByUser(ctx context.Context, userID uint64, isRead *bool, notifType string, offset, limit int) ([]*model.Notification, int64, error)
	CountUnread(ctx context.Context, userID uint64) (int64, error)
	MarkAsRead(ctx context.Context, id uint64, userID uint64) error
	MarkAllAsRead(ctx context.Context, userID uint64) error
	Delete(ctx context.Context, id uint64, userID uint64) error
}

// notificationRepository 通知数据访问实现
type notificationRepository struct {
	db *gorm.DB
}

// NewNotificationRepository 创建通知数据访问实例
func NewNotificationRepository(db *gorm.DB) NotificationRepository {
	return &notificationRepository{db: db}
}

// Create 创建通知
func (r *notificationRepository) Create(ctx context.Context, notification *model.Notification) error {
	return r.db.WithContext(ctx).Create(notification).Error
}

// GetByID 根据ID获取通知
func (r *notificationRepository) GetByID(ctx context.Context, id uint64) (*model.Notification, error) {
	var notification model.Notification
	err := r.db.WithContext(ctx).First(&notification, id).Error
	return &notification, err
}

// ListByUser 获取用户通知列表
func (r *notificationRepository) ListByUser(ctx context.Context, userID uint64, isRead *bool, notifType string, offset, limit int) ([]*model.Notification, int64, error) {
	query := r.db.WithContext(ctx).Where("user_id = ?", userID)

	if isRead != nil {
		query = query.Where("is_read = ?", *isRead)
	}
	if notifType != "" {
		query = query.Where("type = ?", notifType)
	}

	var total int64
	if err := query.Model(&model.Notification{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var notifications []*model.Notification
	err := query.Order("created_at DESC").Offset(offset).Limit(limit).Find(&notifications).Error
	return notifications, total, err
}

// CountUnread 获取未读通知数量
func (r *notificationRepository) CountUnread(ctx context.Context, userID uint64) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&model.Notification{}).
		Where("user_id = ? AND is_read = ?", userID, false).
		Count(&count).Error
	return count, err
}

// MarkAsRead 标记通知为已读
func (r *notificationRepository) MarkAsRead(ctx context.Context, id, userID uint64) error {
	return r.db.WithContext(ctx).Model(&model.Notification{}).
		Where("id = ? AND user_id = ?", id, userID).
		Updates(map[string]any{
			"is_read": true,
			"read_at": gorm.Expr("NOW()"),
		}).Error
}

// MarkAllAsRead 标记用户所有通知为已读
func (r *notificationRepository) MarkAllAsRead(ctx context.Context, userID uint64) error {
	return r.db.WithContext(ctx).Model(&model.Notification{}).
		Where("user_id = ? AND is_read = ?", userID, false).
		Updates(map[string]any{
			"is_read": true,
			"read_at": gorm.Expr("NOW()"),
		}).Error
}

// Delete 硬删除通知
func (r *notificationRepository) Delete(ctx context.Context, id, userID uint64) error {
	return r.db.WithContext(ctx).Unscoped().
		Where("id = ? AND user_id = ?", id, userID).
		Delete(&model.Notification{}).Error
}
