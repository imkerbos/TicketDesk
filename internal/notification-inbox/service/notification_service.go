// Package service 提供站内通知业务逻辑层
package service

import (
	"context"
	"fmt"

	"github.com/kerbos/ticketdesk/internal/model"
	"github.com/kerbos/ticketdesk/internal/notification-inbox/dto"
	"github.com/kerbos/ticketdesk/internal/notification-inbox/repository"
	"github.com/kerbos/ticketdesk/internal/notification-inbox/websocket"
	"github.com/kerbos/ticketdesk/pkg/logger"
	"go.uber.org/zap"
)

// NotificationService 站内通知服务接口
type NotificationService interface {
	// CreateNotification 创建通知并推送
	CreateNotification(ctx context.Context, req *dto.CreateNotificationRequest) error
	// ListNotifications 获取通知列表
	ListNotifications(ctx context.Context, userID uint64, req *dto.ListNotificationsRequest) ([]*dto.NotificationResponse, int64, error)
	// GetUnreadCount 获取未读数量
	GetUnreadCount(ctx context.Context, userID uint64) (int64, error)
	// MarkAsRead 标记为已读
	MarkAsRead(ctx context.Context, id uint64, userID uint64) error
	// MarkAllAsRead 全部标记为已读
	MarkAllAsRead(ctx context.Context, userID uint64) error
	// DeleteNotification 删除通知
	DeleteNotification(ctx context.Context, id uint64, userID uint64) error
}

// notificationService 站内通知服务实现
type notificationService struct {
	repo      repository.NotificationRepository
	wsManager *websocket.Manager
}

// NewNotificationService 创建通知服务实例
func NewNotificationService(
	repo repository.NotificationRepository,
	wsManager *websocket.Manager,
) NotificationService {
	return &notificationService{
		repo:      repo,
		wsManager: wsManager,
	}
}

// CreateNotification 创建通知并通过 WebSocket 推送
func (s *notificationService) CreateNotification(ctx context.Context, req *dto.CreateNotificationRequest) error {
	notification := &model.Notification{
		UserID:     req.UserID,
		Type:       req.Type,
		Title:      req.Title,
		Content:    req.Content,
		EntityType: req.EntityType,
		EntityID:   req.EntityID,
		EntityKey:  req.EntityKey,
		ActorID:    req.ActorID,
		ActorName:  req.ActorName,
		IsRead:     false,
	}

	if err := s.repo.Create(ctx, notification); err != nil {
		logger.Error("failed to create notification",
			zap.Uint64("user_id", req.UserID),
			zap.String("type", req.Type),
			zap.Error(err),
		)
		return fmt.Errorf("创建通知失败: %w", err)
	}

	// 通过 WebSocket 推送实时通知
	s.wsManager.SendToUser(req.UserID, &websocket.WSMessage{
		Type: "notification",
		Data: toNotificationResponse(notification),
	})

	logger.Info("notification created and pushed",
		zap.Uint64("user_id", req.UserID),
		zap.String("type", req.Type),
		zap.String("entity_key", req.EntityKey),
	)

	return nil
}

// ListNotifications 获取通知列表
func (s *notificationService) ListNotifications(ctx context.Context, userID uint64, req *dto.ListNotificationsRequest) ([]*dto.NotificationResponse, int64, error) {
	page := req.GetDefaultPage()
	pageSize := req.GetDefaultPageSize()
	offset := (page - 1) * pageSize

	notifications, total, err := s.repo.ListByUser(ctx, userID, req.IsRead, req.Type, offset, pageSize)
	if err != nil {
		return nil, 0, fmt.Errorf("查询通知列表失败: %w", err)
	}

	responses := make([]*dto.NotificationResponse, len(notifications))
	for i, n := range notifications {
		responses[i] = toNotificationResponse(n)
	}

	return responses, total, nil
}

// GetUnreadCount 获取未读数量
func (s *notificationService) GetUnreadCount(ctx context.Context, userID uint64) (int64, error) {
	count, err := s.repo.CountUnread(ctx, userID)
	if err != nil {
		return 0, fmt.Errorf("查询未读数量失败: %w", err)
	}
	return count, nil
}

// MarkAsRead 标记为已读
func (s *notificationService) MarkAsRead(ctx context.Context, id uint64, userID uint64) error {
	if err := s.repo.MarkAsRead(ctx, id, userID); err != nil {
		return fmt.Errorf("标记已读失败: %w", err)
	}
	return nil
}

// MarkAllAsRead 全部标记为已读
func (s *notificationService) MarkAllAsRead(ctx context.Context, userID uint64) error {
	if err := s.repo.MarkAllAsRead(ctx, userID); err != nil {
		return fmt.Errorf("全部标记已读失败: %w", err)
	}
	return nil
}

// DeleteNotification 删除通知
func (s *notificationService) DeleteNotification(ctx context.Context, id uint64, userID uint64) error {
	if err := s.repo.Delete(ctx, id, userID); err != nil {
		return fmt.Errorf("删除通知失败: %w", err)
	}
	return nil
}

// toNotificationResponse 将模型转换为响应 DTO
func toNotificationResponse(n *model.Notification) *dto.NotificationResponse {
	return &dto.NotificationResponse{
		ID:         n.ID,
		UserID:     n.UserID,
		Type:       n.Type,
		Title:      n.Title,
		Content:    n.Content,
		EntityType: n.EntityType,
		EntityID:   n.EntityID,
		EntityKey:  n.EntityKey,
		ActorID:    n.ActorID,
		ActorName:  n.ActorName,
		IsRead:     n.IsRead,
		ReadAt:     n.ReadAt,
		CreatedAt:  n.CreatedAt,
	}
}
