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
	"github.com/kerbos/ticketdesk/internal/integration-alert/repository"
	"github.com/kerbos/ticketdesk/internal/model"
	"github.com/kerbos/ticketdesk/pkg/cache"
	"github.com/kerbos/ticketdesk/pkg/logger"
	"go.uber.org/zap"
)

// 静默规则缓存常量
const (
	cacheKeyActiveSilences = "alert:silences:active"
	cacheTTLSilences       = 1 * time.Minute
)

// getCachedActiveSilences 获取生效中的静默规则（优先从缓存读取）
func (s *alertService) getCachedActiveSilences(ctx context.Context) ([]*model.AlertSilence, error) {
	var silences []*model.AlertSilence
	if cache.GetJSON(ctx, cacheKeyActiveSilences, &silences) {
		return silences, nil
	}

	silences, err := s.alertSilenceRepo.ListActive(ctx)
	if err != nil {
		return nil, err
	}

	cache.SetJSON(ctx, cacheKeyActiveSilences, silences, cacheTTLSilences)
	return silences, nil
}

// invalidateSilencesCache 失效静默规则缓存
func (s *alertService) invalidateSilencesCache(ctx context.Context) {
	cache.Del(ctx, cacheKeyActiveSilences)
}

// isAlertSilenced 检查告警是否被静默
func (s *alertService) isAlertSilenced(ctx context.Context, labels map[string]string) (bool, error) {
	// 获取所有生效中的静默规则（优先从缓存获取）
	silences, err := s.getCachedActiveSilences(ctx)
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
		// 如果没找到活跃工单，再尝试查找 pending_review 状态的工单（告警恢复后待确认的工单仍可合并）
		err = s.db.WithContext(ctx).
			Where("project_id = ?", rule.ProjectID).
			Where("issue_type_id = ?", rule.IssueTypeID).
			Where("title LIKE ?", titlePattern).
			Where("status = ?", "pending_review").
			Where("created_at >= ?", windowStart).
			Order("created_at DESC").
			First(&issue).Error

		if err != nil {
			return 0, nil // 没有找到可合并的工单，返回 0
		}
	}

	return issue.ID, nil
}

// mergeOldIssues 将同名旧工单标记为 merged 状态，扁平化合并：
// - 所有旧工单的 merged_into_issue_id 都直接指向最新工单（不嵌套）
// - 之前已合并到旧工单的更旧工单也重新指向最新工单
// - 所有告警迁移到最新工单
// - 最新工单的描述汇总全部历史实例信息
func (s *alertService) mergeOldIssues(ctx context.Context, rule *model.AlertRule, alert *model.Alert, newIssueID uint64, newIssueKey string) {
	titlePattern := fmt.Sprintf("[告警] %s%%", alert.AlertName)

	// 查找所有同名的、未关闭的旧工单（排除刚创建的新工单）
	var directOldIssues []model.Issue
	err := s.db.WithContext(ctx).
		Where("project_id = ?", rule.ProjectID).
		Where("issue_type_id = ?", rule.IssueTypeID).
		Where("title LIKE ?", titlePattern).
		Where("id != ?", newIssueID).
		Where("status NOT IN (?)", []string{"resolved", "closed", "merged"}).
		Order("created_at ASC").
		Find(&directOldIssues).Error

	if err != nil || len(directOldIssues) == 0 {
		return
	}

	// 收集所有需要合并的工单 ID（包括直接旧工单 + 它们之前已吸收的更旧工单）
	oldIssueIDs := make([]uint64, 0)
	for _, oi := range directOldIssues {
		oldIssueIDs = append(oldIssueIDs, oi.ID)
	}

	// 扁平化：查找之前已经合并到这些旧工单的更旧工单，让它们也指向最新工单
	var deeperMergedIssues []model.Issue
	if err := s.db.WithContext(ctx).
		Where("merged_into_issue_id IN (?)", oldIssueIDs).
		Where("deleted_at IS NULL").
		Find(&deeperMergedIssues).Error; err == nil && len(deeperMergedIssues) > 0 {
		// 更新它们的 merged_into_issue_id 指向最新工单
		s.db.WithContext(ctx).
			Model(&model.Issue{}).
			Where("merged_into_issue_id IN (?)", oldIssueIDs).
			Update("merged_into_issue_id", newIssueID)

		logger.Info("flattened deeper merged issues to new target",
			zap.Int("count", len(deeperMergedIssues)),
			zap.String("new_issue_key", newIssueKey),
		)
	}

	// 收集所有最终指向新工单的旧工单（用于生成完整描述）
	// = 本次直接合并的 + 之前深层合并的
	allMergedIssueIDs := make([]uint64, 0)
	allMergedIssueIDs = append(allMergedIssueIDs, oldIssueIDs...)
	for _, dmi := range deeperMergedIssues {
		allMergedIssueIDs = append(allMergedIssueIDs, dmi.ID)
	}

	// 迁移所有旧工单关联的告警到新工单
	var migratedAlertCount int
	var historyInfo strings.Builder
	for _, oldIssueID := range allMergedIssueIDs {
		oldAlerts, err := s.alertRepo.ListByIssueID(ctx, oldIssueID)
		if err != nil || len(oldAlerts) == 0 {
			continue
		}

		// 查旧工单 key 用于描述
		oldIssue, _ := s.issueRepo.GetByID(ctx, oldIssueID)
		oldIssueKey := ""
		if oldIssue != nil {
			oldIssueKey = oldIssue.IssueKey
		}

		for _, oldAlert := range oldAlerts {
			// 迁移告警到新工单
			oldAlert.IssueID = &newIssueID
			if err := s.alertRepo.Update(ctx, oldAlert); err != nil {
				logger.Error("failed to migrate alert to new issue",
					zap.Uint64("alert_id", oldAlert.ID),
					zap.Error(err),
				)
				continue
			}
			migratedAlertCount++

			// 收集实例信息
			oldLabels := repository.JSONToLabels(oldAlert.Labels)
			oldAnnotations := repository.JSONToLabels(oldAlert.Annotations)
			instance := oldLabels["instance"]
			if instance == "" {
				instance = oldLabels["target_ident"]
			}
			nodename := oldLabels["nodename"]
			desc := ""
			if d, ok := oldAnnotations["description"]; ok {
				desc = d
			}

			fmt.Fprintf(&historyInfo, "\n\n---\n### 历史告警 - 来自 %s (%s)\n",
				oldIssueKey, oldAlert.StartsAt.Format("2006-01-02 15:04:05"))
			fmt.Fprintf(&historyInfo, "- **实例**: %s\n", instance)
			if nodename != "" {
				fmt.Fprintf(&historyInfo, "- **主机名**: %s\n", nodename)
			}
			fmt.Fprintf(&historyInfo, "- **指纹**: %s\n", oldAlert.Fingerprint)
			if desc != "" {
				fmt.Fprintf(&historyInfo, "- **描述**: %s\n", desc)
			}
		}
	}

	// 标记直接旧工单为 merged，设置 merged_into_issue_id
	var mergedKeys []string
	now := time.Now()
	for _, oldIssue := range directOldIssues {
		mergedKeys = append(mergedKeys, oldIssue.IssueKey)

		oldIssue.Status = "merged"
		oldIssue.Resolution = "merged"
		oldIssue.ClosedAt = &now
		oldIssue.MergedIntoIssueID = &newIssueID
		oldIssue.Description += fmt.Sprintf("\n\n---\n> 🔗 **已合并**: 此工单已合并到新工单 **%s**，后续告警将在新工单中跟进\n", newIssueKey)

		if err := s.issueRepo.Update(ctx, &oldIssue); err != nil {
			logger.Error("failed to mark old issue as merged",
				zap.String("issue_key", oldIssue.IssueKey),
				zap.Error(err),
			)
		}
	}

	// 收集深层工单 key（用于合并来源显示）
	for _, dmi := range deeperMergedIssues {
		mergedKeys = append(mergedKeys, dmi.IssueKey)
	}

	// 更新新工单：汇总全部历史实例 + 合并来源 + 标题实例计数
	newIssue, err := s.issueRepo.GetByID(ctx, newIssueID)
	if err != nil {
		logger.Error("failed to get new issue for merge link", zap.Error(err))
		return
	}

	if historyInfo.Len() > 0 {
		newIssue.Description += historyInfo.String()
	}

	newIssue.Description += fmt.Sprintf("\n\n---\n> 📎 **合并来源**: 此工单汇总了 **%s** 的全部告警信息，旧工单已标记为「已合并」状态\n",
		strings.Join(mergedKeys, ", "))

	// 统计新工单关联的全部告警数更新标题
	totalAlerts, err := s.alertRepo.ListByIssueID(ctx, newIssueID)
	if err == nil && len(totalAlerts) > 0 {
		newIssue.Title = fmt.Sprintf("[告警] %s (%d 个实例)", alert.AlertName, len(totalAlerts))
	}

	if err := s.issueRepo.Update(ctx, newIssue); err != nil {
		logger.Error("failed to update new issue with merge info", zap.Error(err))
	}

	logger.Info("merge completed: alerts migrated, history flattened",
		zap.String("new_issue_key", newIssueKey),
		zap.Int("merged_issues", len(mergedKeys)),
		zap.Int("migrated_alerts", migratedAlertCount),
	)
}

// SyncMergedIssueStatus 同步状态到被合并的旧工单（扁平化查询，无需递归）
// 因为 merged_into_issue_id 是扁平化的（所有旧工单直接指向最终目标），
// 所以一次查询 WHERE merged_into_issue_id = ? 就能找到全部关联工单
func (s *alertService) SyncMergedIssueStatus(ctx context.Context, issueID uint64, issueStatus string) error {
	// 仅当状态变为 resolved/closed 时才级联
	if issueStatus != "resolved" && issueStatus != "closed" {
		return nil
	}

	// 扁平化查询：所有 merged_into_issue_id 指向当前工单的旧工单
	var mergedIssues []model.Issue
	err := s.db.WithContext(ctx).
		Where("merged_into_issue_id = ?", issueID).
		Where("deleted_at IS NULL").
		Find(&mergedIssues).Error
	if err != nil {
		return fmt.Errorf("failed to find merged issues: %w", err)
	}

	if len(mergedIssues) == 0 {
		return nil
	}

	now := time.Now()
	for _, mi := range mergedIssues {
		updates := map[string]interface{}{
			"status":     issueStatus,
			"updated_at": now,
		}
		if issueStatus == "resolved" {
			updates["resolved_at"] = &now
			updates["resolution"] = "done"
		}
		if issueStatus == "closed" {
			updates["closed_at"] = &now
		}

		if err := s.db.WithContext(ctx).Model(&model.Issue{}).Where("id = ?", mi.ID).Updates(updates).Error; err != nil {
			logger.Error("failed to cascade status to merged issue",
				zap.Uint64("merged_issue_id", mi.ID),
				zap.String("merged_issue_key", mi.IssueKey),
				zap.String("new_status", issueStatus),
				zap.Error(err),
			)
			continue
		}

		logger.Info("merged issue status cascaded",
			zap.String("merged_issue_key", mi.IssueKey),
			zap.Uint64("parent_issue_id", issueID),
			zap.String("new_status", issueStatus),
		)
	}

	logger.Info("merged issues status cascade completed",
		zap.Uint64("parent_issue_id", issueID),
		zap.String("new_status", issueStatus),
		zap.Int("merged_count", len(mergedIssues)),
	)

	return nil
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

	s.invalidateSilencesCache(ctx)
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

	s.invalidateSilencesCache(ctx)
	return s.GetAlertSilence(ctx, id)
}

// DeleteAlertSilence 删除告警静默
func (s *alertService) DeleteAlertSilence(ctx context.Context, id uint64) error {
	if err := s.alertSilenceRepo.Delete(ctx, id); err != nil {
		return err
	}
	s.invalidateSilencesCache(ctx)
	return nil
}

// CancelAlertSilence 取消告警静默
func (s *alertService) CancelAlertSilence(ctx context.Context, id uint64) error {
	if err := s.alertSilenceRepo.Cancel(ctx, id); err != nil {
		return err
	}
	s.invalidateSilencesCache(ctx)
	return nil
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
