// Package service 提供活动日志业务逻辑层
package service

import (
	"context"
	"fmt"

	"github.com/kerbos/ticketdesk/internal/activity/dto"
	"github.com/kerbos/ticketdesk/internal/activity/repository"
	"github.com/kerbos/ticketdesk/internal/model"
	"github.com/kerbos/ticketdesk/pkg/logger"
	"go.uber.org/zap"
)

// ActivityService 活动日志服务接口
type ActivityService interface {
	// LogActivity 记录活动
	LogActivity(ctx context.Context, userID uint64, userName, action, entityType string, entityID uint64, entityKey, details string) error
	// ListActivities 分页查询活动列表
	ListActivities(ctx context.Context, req *dto.ListActivitiesRequest) ([]*dto.ActivityResponse, int64, error)
	// GetRecentActivities 获取最近的活动
	GetRecentActivities(ctx context.Context, limit int) ([]*dto.ActivityResponse, error)
}

// activityService 活动日志服务实现
type activityService struct {
	activityRepo repository.ActivityRepository
}

// NewActivityService 创建活动日志服务实例
func NewActivityService(activityRepo repository.ActivityRepository) ActivityService {
	return &activityService{
		activityRepo: activityRepo,
	}
}

// LogActivity 记录活动
func (s *activityService) LogActivity(ctx context.Context, userID uint64, userName, action, entityType string, entityID uint64, entityKey, details string) error {
	activity := &model.ActivityLog{
		UserID:     userID,
		UserName:   userName,
		Action:     action,
		EntityType: entityType,
		EntityID:   entityID,
		EntityKey:  entityKey,
		Details:    details,
	}

	if err := s.activityRepo.Create(ctx, activity); err != nil {
		logger.Error("failed to log activity",
			zap.Uint64("user_id", userID),
			zap.String("action", action),
			zap.String("entity_type", entityType),
			zap.Error(err),
		)
		return fmt.Errorf("记录活动失败: %w", err)
	}

	return nil
}

// ListActivities 分页查询活动列表
func (s *activityService) ListActivities(ctx context.Context, req *dto.ListActivitiesRequest) ([]*dto.ActivityResponse, int64, error) {
	page := req.GetDefaultPage()
	pageSize := req.GetDefaultPageSize()
	offset := (page - 1) * pageSize

	activities, total, err := s.activityRepo.List(ctx, offset, pageSize, req.EntityType, req.EntityKey, req.EntityID, req.UserID)
	if err != nil {
		logger.Error("failed to list activities", zap.Error(err))
		return nil, 0, fmt.Errorf("查询活动列表失败: %w", err)
	}

	responses := make([]*dto.ActivityResponse, len(activities))
	for i, activity := range activities {
		responses[i] = s.toActivityResponse(activity)
	}

	return responses, total, nil
}

// GetRecentActivities 获取最近的活动
func (s *activityService) GetRecentActivities(ctx context.Context, limit int) ([]*dto.ActivityResponse, error) {
	if limit <= 0 {
		limit = 10
	}

	activities, err := s.activityRepo.GetRecent(ctx, limit)
	if err != nil {
		logger.Error("failed to get recent activities", zap.Error(err))
		return nil, fmt.Errorf("获取最近活动失败: %w", err)
	}

	responses := make([]*dto.ActivityResponse, len(activities))
	for i, activity := range activities {
		responses[i] = s.toActivityResponse(activity)
	}

	return responses, nil
}

// toActivityResponse 将活动模型转换为响应 DTO
func (s *activityService) toActivityResponse(activity *model.ActivityLog) *dto.ActivityResponse {
	return &dto.ActivityResponse{
		ID:         activity.ID,
		UserID:     activity.UserID,
		UserName:   activity.UserName,
		Action:     activity.Action,
		EntityType: activity.EntityType,
		EntityID:   activity.EntityID,
		EntityKey:  activity.EntityKey,
		Details:    activity.Details,
		CreatedAt:  activity.CreatedAt,
	}
}
