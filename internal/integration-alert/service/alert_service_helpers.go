// Package service 提供告警业务逻辑层 - 辅助方法
package service

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/kerbos/ticketdesk/internal/integration-alert/dto"
	"github.com/kerbos/ticketdesk/internal/model"
	"github.com/kerbos/ticketdesk/pkg/logger"
	"go.uber.org/zap"
)

// isAlertSilenced 检查告警是否被静默
func (s *alertService) isAlertSilenced(ctx context.Context, labels map[string]string) (bool, error) {
	// 获取所有生效中的静默规则
	silences, err := s.alertSilenceRepo.ListActive(ctx)
	if err != nil {
		return false, err
	}

	// 检查是否匹配任何静默规则
	for _, silence := range silences {
		var matchers []dto.LabelMatcher
		if err := json.Unmarshal([]byte(silence.LabelMatchers), &matchers); err != nil {
			logger.Error("failed to unmarshal silence matchers", zap.Error(err))
			continue
		}

		if s.matchLabels(matchers, labels) {
			logger.Info("alert matched silence rule",
				zap.Uint64("silence_id", silence.ID),
				zap.String("silence_name", silence.Name),
			)
			return true, nil
		}
	}

	return false, nil
}

// matchLabels 检查标签是否匹配
func (s *alertService) matchLabels(matchers []dto.LabelMatcher, labels map[string]string) bool {
	for _, matcher := range matchers {
		labelValue, exists := labels[matcher.Key]

		switch matcher.Operator {
		case "==":
			if !exists || labelValue != matcher.Value {
				return false
			}
		case "!=":
			if exists && labelValue == matcher.Value {
				return false
			}
		case "=~":
			if !exists {
				return false
			}
			matched, err := regexp.MatchString(matcher.Value, labelValue)
			if err != nil || !matched {
				return false
			}
		case "!~":
			if !exists {
				continue
			}
			matched, err := regexp.MatchString(matcher.Value, labelValue)
			if err != nil || matched {
				return false
			}
		}
	}

	return true
}

// findMergeableIssue 查找可合并的工单（在合并窗口内的同名活跃工单）
func (s *alertService) findMergeableIssue(ctx context.Context, rule *model.AlertRule, alert *model.Alert) (uint64, error) {
	// 计算时间窗口
	windowStart := time.Now().Add(-time.Duration(rule.MergeWindow) * time.Second)

	// 查询该项目下、该工单类型、同告警名称、在时间窗口内创建的、未关闭的工单
	titlePattern := fmt.Sprintf("[告警] %s%%", alert.AlertName)
	var issue model.Issue
	err := s.db.WithContext(ctx).
		Where("project_id = ?", rule.ProjectID).
		Where("issue_type_id = ?", rule.IssueTypeID).
		Where("title LIKE ?", titlePattern).
		Where("status NOT IN (?)", []string{"resolved", "closed", "merged"}).
		Where("created_at >= ?", windowStart).
		Order("created_at DESC").
		First(&issue).Error

	if err != nil {
		return 0, nil // 没有找到可合并的工单，返回 0
	}

	return issue.ID, nil
}

// mergeOldIssues 将同名旧工单标记为 merged 状态，双向关联新旧工单
func (s *alertService) mergeOldIssues(ctx context.Context, rule *model.AlertRule, alert *model.Alert, newIssueID uint64, newIssueKey string) {
	titlePattern := fmt.Sprintf("[告警] %s%%", alert.AlertName)

	// 查找所有同名的、未关闭的旧工单（排除刚创建的新工单）
	var oldIssues []model.Issue
	err := s.db.WithContext(ctx).
		Where("project_id = ?", rule.ProjectID).
		Where("issue_type_id = ?", rule.IssueTypeID).
		Where("title LIKE ?", titlePattern).
		Where("id != ?", newIssueID).
		Where("status NOT IN (?)", []string{"resolved", "closed", "merged"}).
		Order("created_at DESC").
		Find(&oldIssues).Error

	if err != nil || len(oldIssues) == 0 {
		return
	}

	// 收集旧工单编号
	var mergedKeys []string
	now := time.Now()
	for _, oldIssue := range oldIssues {
		mergedKeys = append(mergedKeys, oldIssue.IssueKey)

		// 旧工单追加关联信息：指向新工单
		oldIssue.Status = "merged"
		oldIssue.Resolution = "merged"
		oldIssue.ClosedAt = &now
		oldIssue.Description += fmt.Sprintf("\n\n---\n> 🔗 **已合并**: 此工单已合并到新工单 **%s**，后续告警将在新工单中跟进\n", newIssueKey)

		if err := s.issueRepo.Update(ctx, &oldIssue); err != nil {
			logger.Error("failed to mark old issue as merged",
				zap.String("issue_key", oldIssue.IssueKey),
				zap.Error(err),
			)
			continue
		}

		logger.Info("old issue marked as merged",
			zap.String("old_issue_key", oldIssue.IssueKey),
			zap.String("new_issue_key", newIssueKey),
			zap.String("alert_name", alert.AlertName),
		)
	}

	// 新工单追加关联信息：来自哪些旧工单
	if len(mergedKeys) > 0 {
		newIssue, err := s.issueRepo.GetByID(ctx, newIssueID)
		if err != nil {
			logger.Error("failed to get new issue for merge link", zap.Error(err))
			return
		}

		newIssue.Description += fmt.Sprintf("\n\n---\n> 📎 **合并来源**: 此工单由 **%s** 合并而来，旧工单已标记为「已合并」状态\n", strings.Join(mergedKeys, ", "))

		if err := s.issueRepo.Update(ctx, newIssue); err != nil {
			logger.Error("failed to update new issue with merge info", zap.Error(err))
		}
	}
}

// ============ 告警静默管理 ============

// CreateAlertSilence 创建告警静默
func (s *alertService) CreateAlertSilence(ctx context.Context, req *dto.CreateAlertSilenceRequest, userID uint64) (*dto.AlertSilenceResponse, error) {
	// 序列化标签匹配器
	matchersJSON, err := json.Marshal(req.LabelMatchers)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal label matchers: %w", err)
	}

	silence := &model.AlertSilence{
		Name:          req.Name,
		Description:   req.Description,
		LabelMatchers: string(matchersJSON),
		StartsAt:      req.StartsAt,
		EndsAt:        req.EndsAt,
		CreatedBy:     userID,
		Comment:       req.Comment,
		Status:        1, // 默认生效中
	}

	if err := s.alertSilenceRepo.Create(ctx, silence); err != nil {
		return nil, err
	}

	return s.GetAlertSilence(ctx, silence.ID)
}

// GetAlertSilence 获取告警静默详情
func (s *alertService) GetAlertSilence(ctx context.Context, id uint64) (*dto.AlertSilenceResponse, error) {
	silence, err := s.alertSilenceRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	return s.toAlertSilenceResponse(silence), nil
}

// UpdateAlertSilence 更新告警静默
func (s *alertService) UpdateAlertSilence(ctx context.Context, id uint64, req *dto.UpdateAlertSilenceRequest) (*dto.AlertSilenceResponse, error) {
	silence, err := s.alertSilenceRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// 更新字段
	if req.Name != nil {
		silence.Name = *req.Name
	}
	if req.Description != nil {
		silence.Description = *req.Description
	}
	if req.LabelMatchers != nil {
		matchersJSON, err := json.Marshal(req.LabelMatchers)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal label matchers: %w", err)
		}
		silence.LabelMatchers = string(matchersJSON)
	}
	if req.StartsAt != nil {
		silence.StartsAt = *req.StartsAt
	}
	if req.EndsAt != nil {
		silence.EndsAt = *req.EndsAt
	}
	if req.Comment != nil {
		silence.Comment = *req.Comment
	}
	if req.Status != nil {
		silence.Status = *req.Status
	}

	if err := s.alertSilenceRepo.Update(ctx, silence); err != nil {
		return nil, err
	}

	return s.GetAlertSilence(ctx, id)
}

// DeleteAlertSilence 删除告警静默
func (s *alertService) DeleteAlertSilence(ctx context.Context, id uint64) error {
	return s.alertSilenceRepo.Delete(ctx, id)
}

// CancelAlertSilence 取消告警静默
func (s *alertService) CancelAlertSilence(ctx context.Context, id uint64) error {
	return s.alertSilenceRepo.Cancel(ctx, id)
}

// ListAlertSilences 获取告警静默列表
func (s *alertService) ListAlertSilences(ctx context.Context, req *dto.AlertSilenceListRequest) (*dto.AlertSilenceListResponse, error) {
	silences, total, err := s.alertSilenceRepo.List(ctx, req)
	if err != nil {
		return nil, err
	}

	items := make([]dto.AlertSilenceResponse, 0, len(silences))
	for _, silence := range silences {
		items = append(items, *s.toAlertSilenceResponse(silence))
	}

	page := req.Page
	if page < 1 {
		page = 1
	}
	pageSize := req.PageSize
	if pageSize < 1 {
		pageSize = 20
	}

	return &dto.AlertSilenceListResponse{
		Items:    items,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	}, nil
}

// toAlertSilenceResponse 转换为告警静默响应
func (s *alertService) toAlertSilenceResponse(silence *model.AlertSilence) *dto.AlertSilenceResponse {
	var matchers []dto.LabelMatcher
	json.Unmarshal([]byte(silence.LabelMatchers), &matchers)

	return &dto.AlertSilenceResponse{
		ID:            silence.ID,
		Name:          silence.Name,
		Description:   silence.Description,
		LabelMatchers: matchers,
		StartsAt:      silence.StartsAt,
		EndsAt:        silence.EndsAt,
		CreatedBy:     silence.CreatedBy,
		CreatedByName: "", // TODO: 从用户服务获取用户名
		Comment:       silence.Comment,
		Status:        silence.Status,
		CreatedAt:     silence.CreatedAt,
		UpdatedAt:     silence.UpdatedAt,
	}
}

// ============ 告警分组查询 ============

// GroupAlerts 按标签分组统计告警
func (s *alertService) GroupAlerts(ctx context.Context, req *dto.AlertGroupRequest) (*dto.AlertGroupResponse, error) {
	items, err := s.alertRepo.GroupBy(ctx, req.GroupBy, req)
	if err != nil {
		return nil, err
	}

	var total int64
	for _, item := range items {
		total += item.Count
	}

	return &dto.AlertGroupResponse{
		GroupBy: req.GroupBy,
		Items:   items,
		Total:   total,
	}, nil
}

// ============ 工单状态同步 ============

// SyncIssueStatus 同步工单状态到告警
func (s *alertService) SyncIssueStatus(ctx context.Context, issueID uint64, issueStatus string) error {
	// 根据工单状态映射告警状态
	var alertStatus string
	switch issueStatus {
	case "resolved", "closed":
		alertStatus = "resolved"
	case "reopened":
		alertStatus = "firing"
	case "merged":
		// 已合并的工单不同步告警状态，告警会关联到新工单
		return nil
	default:
		// 其他状态不同步
		return nil
	}

	logger.Info("syncing issue status to alerts",
		zap.Uint64("issue_id", issueID),
		zap.String("issue_status", issueStatus),
		zap.String("alert_status", alertStatus),
	)

	return s.alertRepo.UpdateStatusByIssueID(ctx, issueID, alertStatus)
}
