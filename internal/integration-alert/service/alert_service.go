// Package service 提供告警业务逻辑层
package service

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"regexp"
	"sort"
	"strings"
	"time"

	issueRepo "github.com/kerbos/ticketdesk/internal/core-issue/repository"
	projectRepo "github.com/kerbos/ticketdesk/internal/core-project/repository"
	"github.com/kerbos/ticketdesk/internal/integration-alert/dto"
	"github.com/kerbos/ticketdesk/internal/integration-alert/repository"
	"github.com/kerbos/ticketdesk/internal/model"
	"github.com/kerbos/ticketdesk/pkg/logger"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// AlertService 告警服务接口
type AlertService interface {
	// Webhook 处理
	HandleWebhook(ctx context.Context, req *dto.AlertWebhookRequest) error
	HandleNightingaleWebhook(ctx context.Context, events []dto.N9eAlertEvent) error
	HandleWebhookWithSource(ctx context.Context, req *dto.AlertWebhookRequest, sourceName string) error
	HandleNightingaleWebhookWithSource(ctx context.Context, events []dto.N9eAlertEvent, sourceName string) error

	// 告警查询
	GetAlert(ctx context.Context, id uint64) (*dto.AlertResponse, error)
	ListAlerts(ctx context.Context, req *dto.AlertListRequest) (*dto.AlertListResponse, error)
	GetAlertStats(ctx context.Context) (*dto.AlertStatsResponse, error)
	GroupAlerts(ctx context.Context, req *dto.AlertGroupRequest) (*dto.AlertGroupResponse, error)

	// 告警操作
	AckAlert(ctx context.Context, id uint64, userID uint64, req *dto.AlertAckRequest) error
	ResolveAlert(ctx context.Context, id uint64, userID uint64, req *dto.AlertResolveRequest) error

	// 告警规则管理
	CreateAlertRule(ctx context.Context, req *dto.CreateAlertRuleRequest) (*dto.AlertRuleResponse, error)
	GetAlertRule(ctx context.Context, id uint64) (*dto.AlertRuleResponse, error)
	UpdateAlertRule(ctx context.Context, id uint64, req *dto.UpdateAlertRuleRequest) (*dto.AlertRuleResponse, error)
	DeleteAlertRule(ctx context.Context, id uint64) error
	ListAlertRules(ctx context.Context, req *dto.AlertRuleListRequest) (*dto.AlertRuleListResponse, error)

	// 告警静默管理
	CreateAlertSilence(ctx context.Context, req *dto.CreateAlertSilenceRequest, userID uint64) (*dto.AlertSilenceResponse, error)
	GetAlertSilence(ctx context.Context, id uint64) (*dto.AlertSilenceResponse, error)
	UpdateAlertSilence(ctx context.Context, id uint64, req *dto.UpdateAlertSilenceRequest) (*dto.AlertSilenceResponse, error)
	DeleteAlertSilence(ctx context.Context, id uint64) error
	CancelAlertSilence(ctx context.Context, id uint64) error
	ListAlertSilences(ctx context.Context, req *dto.AlertSilenceListRequest) (*dto.AlertSilenceListResponse, error)

	// 工单状态同步
	SyncIssueStatus(ctx context.Context, issueID uint64, issueStatus string) error
	// SyncMergedIssueStatus 同步状态到被合并的旧工单
	SyncMergedIssueStatus(ctx context.Context, issueID uint64, issueStatus string) error

	// TryAutoResolveIssue 检查工单下所有告警是否都已恢复，如果是则自动关闭工单
	TryAutoResolveIssue(ctx context.Context, issueID uint64) error
}

// alertService 告警服务实现
type alertService struct {
	alertRepo        repository.AlertRepository
	alertRuleRepo    repository.AlertRuleRepository
	alertSilenceRepo repository.AlertSilenceRepository
	issueRepo        issueRepo.IssueRepository
	commentRepo      issueRepo.CommentRepository
	projectRepo      projectRepo.ProjectRepository
	issueTypeRepo    projectRepo.IssueTypeRepository
	db               *gorm.DB
}

// NewAlertService 创建告警服务
func NewAlertService(
	alertRepo repository.AlertRepository,
	alertRuleRepo repository.AlertRuleRepository,
	alertSilenceRepo repository.AlertSilenceRepository,
	issueRepo issueRepo.IssueRepository,
	commentRepo issueRepo.CommentRepository,
	projectRepo projectRepo.ProjectRepository,
	issueTypeRepo projectRepo.IssueTypeRepository,
	db *gorm.DB,
) AlertService {
	return &alertService{
		alertRepo:        alertRepo,
		alertRuleRepo:    alertRuleRepo,
		alertSilenceRepo: alertSilenceRepo,
		issueRepo:        issueRepo,
		commentRepo:      commentRepo,
		projectRepo:      projectRepo,
		issueTypeRepo:    issueTypeRepo,
		db:               db,
	}
}

// HandleWebhook 处理 Webhook 告警
func (s *alertService) HandleWebhook(ctx context.Context, req *dto.AlertWebhookRequest) error {
	logger.Info("received alert webhook",
		zap.String("status", req.Status),
		zap.Int("alert_count", len(req.Alerts)),
	)

	for _, alertItem := range req.Alerts {
		if err := s.processAlertWithSource(ctx, &alertItem, "prometheus"); err != nil {
			logger.Error("failed to process alert",
				zap.String("fingerprint", alertItem.Fingerprint),
				zap.Error(err),
			)
			// 继续处理其他告警，不中断
			continue
		}
	}

	return nil
}

// HandleNightingaleWebhook 处理夜莺 Webhook 告警
func (s *alertService) HandleNightingaleWebhook(ctx context.Context, events []dto.N9eAlertEvent) error {
	logger.Info("received nightingale webhook",
		zap.Int("event_count", len(events)),
	)

	for _, event := range events {
		alertItem := event.ToAlertWebhookItem()
		if err := s.processAlertWithSource(ctx, alertItem, "nightingale"); err != nil {
			logger.Error("failed to process nightingale alert",
				zap.String("hash", event.Hash),
				zap.String("rule_name", event.RuleName),
				zap.Error(err),
			)
			continue
		}
	}

	return nil
}

// HandleWebhookWithSource 处理 Webhook 告警（带数据源名称）
func (s *alertService) HandleWebhookWithSource(ctx context.Context, req *dto.AlertWebhookRequest, sourceName string) error {
	logger.Info("received alert webhook",
		zap.String("source", sourceName),
		zap.String("status", req.Status),
		zap.Int("alert_count", len(req.Alerts)),
	)

	for _, alertItem := range req.Alerts {
		if err := s.processAlertWithSource(ctx, &alertItem, sourceName); err != nil {
			logger.Error("failed to process alert",
				zap.String("source", sourceName),
				zap.String("fingerprint", alertItem.Fingerprint),
				zap.Error(err),
			)
			continue
		}
	}

	return nil
}

// HandleNightingaleWebhookWithSource 处理夜莺 Webhook 告警（带数据源名称）
func (s *alertService) HandleNightingaleWebhookWithSource(ctx context.Context, events []dto.N9eAlertEvent, sourceName string) error {
	logger.Info("received nightingale webhook",
		zap.String("source", sourceName),
		zap.Int("event_count", len(events)),
	)

	for _, event := range events {
		alertItem := event.ToAlertWebhookItem()
		if err := s.processAlertWithSource(ctx, alertItem, sourceName); err != nil {
			logger.Error("failed to process nightingale alert",
				zap.String("source", sourceName),
				zap.String("hash", event.Hash),
				zap.String("rule_name", event.RuleName),
				zap.Error(err),
			)
			continue
		}
	}

	return nil
}

// processAlertWithSource 处理单个告警（带来源参数）
func (s *alertService) processAlertWithSource(ctx context.Context, alertItem *dto.AlertWebhookAlertItem, source string) error {
	// 1. 检查告警是否被静默
	if silenced, err := s.isAlertSilenced(ctx, alertItem.Labels); err != nil {
		logger.Error("failed to check alert silence", zap.Error(err))
	} else if silenced {
		logger.Info("alert is silenced, skipping",
			zap.String("alert_name", alertItem.Labels["alertname"]),
		)
		return nil
	}

	// 2. 计算告警指纹（如果 Webhook 没有提供）
	fingerprint := alertItem.Fingerprint
	if fingerprint == "" {
		fingerprint = s.calculateFingerprint(alertItem.Labels)
	}

	// 3. 查询是否已存在该告警
	existingAlert, err := s.alertRepo.GetByFingerprint(ctx, fingerprint)
	if err != nil && err != gorm.ErrRecordNotFound {
		return fmt.Errorf("failed to get alert by fingerprint: %w", err)
	}

	// 4. 如果告警已存在，更新状态
	if existingAlert != nil {
		return s.updateExistingAlert(ctx, existingAlert, alertItem)
	}

	// 5. 创建新告警
	return s.createNewAlert(ctx, fingerprint, alertItem, source)
}

// calculateFingerprint 计算告警指纹
func (s *alertService) calculateFingerprint(labels map[string]string) string {
	// 排序标签键
	keys := make([]string, 0, len(labels))
	for k := range labels {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	// 构建指纹字符串
	var parts []string
	for _, k := range keys {
		parts = append(parts, fmt.Sprintf("%s=%s", k, labels[k]))
	}
	fingerprintStr := strings.Join(parts, ";")

	// 计算 SHA256
	hash := sha256.Sum256([]byte(fingerprintStr))
	return fmt.Sprintf("%x", hash)
}

// updateExistingAlert 更新已存在的告警
func (s *alertService) updateExistingAlert(ctx context.Context, alert *model.Alert, alertItem *dto.AlertWebhookAlertItem) error {
	// 更新告警状态
	oldStatus := alert.Status
	alert.Status = alertItem.Status
	alert.EndsAt = alertItem.EndsAt

	// 如果告警恢复，检查是否所有关联告警都已恢复，若是则自动解决工单
	if alertItem.Status == "resolved" && oldStatus != "resolved" && alert.IssueID != nil {
		logger.Info("alert resolved, checking if all alerts resolved for issue",
			zap.Uint64("alert_id", alert.ID),
			zap.Uint64("issue_id", *alert.IssueID),
		)

		// 先保存当前告警状态（因为 TryAutoResolveIssue 需要读到最新状态）
		if err := s.alertRepo.Update(ctx, alert); err != nil {
			return fmt.Errorf("failed to update alert: %w", err)
		}

		if err := s.TryAutoResolveIssue(ctx, *alert.IssueID); err != nil {
			logger.Error("failed to try auto resolve issue",
				zap.Uint64("issue_id", *alert.IssueID),
				zap.Error(err),
			)
		}

		return nil // 已在上面保存过，提前返回
	}

	// 如果告警仍在触发且没有关联工单，尝试自动建单
	if alert.Status == "firing" && alert.IssueID == nil {
		if err := s.tryAutoCreateIssue(ctx, alert); err != nil {
			logger.Error("failed to auto create issue for existing alert",
				zap.Uint64("alert_id", alert.ID),
				zap.Error(err),
			)
		}
	}

	return s.alertRepo.Update(ctx, alert)
}

// createNewAlert 创建新告警
func (s *alertService) createNewAlert(ctx context.Context, fingerprint string, alertItem *dto.AlertWebhookAlertItem, source string) error {
	// 提取告警名称和严重程度
	alertName := alertItem.Labels["alertname"]
	severity := alertItem.Labels["severity"]
	if severity == "" {
		severity = "warning"
	}

	// 序列化标签和注解
	labelsJSON := repository.LabelsToJSON(alertItem.Labels)
	annotationsJSON := repository.LabelsToJSON(alertItem.Annotations)

	// 创建告警记录
	alert := &model.Alert{
		Fingerprint: fingerprint,
		Source:      source,
		AlertName:   alertName,
		Severity:    severity,
		Status:      alertItem.Status,
		Labels:      labelsJSON,
		Annotations: annotationsJSON,
		StartsAt:    alertItem.StartsAt,
		EndsAt:      alertItem.EndsAt,
	}

	if err := s.alertRepo.Create(ctx, alert); err != nil {
		return fmt.Errorf("failed to create alert: %w", err)
	}

	// 尝试自动建单
	if err := s.tryAutoCreateIssue(ctx, alert); err != nil {
		logger.Error("failed to auto create issue",
			zap.Uint64("alert_id", alert.ID),
			zap.Error(err),
		)
		// 不中断告警创建流程
	}

	return nil
}

// tryAutoCreateIssue 尝试自动建单
func (s *alertService) tryAutoCreateIssue(ctx context.Context, alert *model.Alert) error {
	// 获取所有启用的告警规则
	rules, err := s.alertRuleRepo.ListEnabled(ctx)
	if err != nil {
		return fmt.Errorf("failed to list enabled rules: %w", err)
	}

	// 解析告警标签
	labels := repository.JSONToLabels(alert.Labels)
	annotations := repository.JSONToLabels(alert.Annotations)

	// 匹配规则
	for _, rule := range rules {
		if s.matchRule(rule, labels) {
			logger.Info("matched alert rule",
				zap.Uint64("alert_id", alert.ID),
				zap.Uint64("rule_id", rule.ID),
			)

			// 检查是否需要合并到现有工单
			if rule.MergeWindow > 0 {
				existingIssueID, err := s.findMergeableIssue(ctx, rule, alert)
				if err != nil {
					logger.Error("failed to find mergeable issue", zap.Error(err))
				} else if existingIssueID > 0 {
					// 合并到现有工单
					logger.Info("merging alert to existing issue",
						zap.Uint64("alert_id", alert.ID),
						zap.Uint64("issue_id", existingIssueID),
					)
					alert.IssueID = &existingIssueID
					if err := s.alertRepo.Update(ctx, alert); err != nil {
						return fmt.Errorf("failed to update alert with issue_id: %w", err)
					}
					// 追加实例信息到工单
					if err := s.appendAlertToIssue(ctx, existingIssueID, alert, labels, annotations); err != nil {
						logger.Error("failed to append alert info to issue",
							zap.Uint64("issue_id", existingIssueID),
							zap.Error(err),
						)
					}
					return nil
				}
			}

			// 创建新工单
			issueID, newIssueKey, err := s.createIssueFromAlert(ctx, alert, rule, labels, annotations)
			if err != nil {
				return fmt.Errorf("failed to create issue: %w", err)
			}

			// 将同名旧工单标记为 merged，并双向关联
			if rule.MergeWindow > 0 {
				s.mergeOldIssues(ctx, rule, alert, issueID, newIssueKey)
			}

			// 更新告警关联的工单 ID
			alert.IssueID = &issueID
			if err := s.alertRepo.Update(ctx, alert); err != nil {
				return fmt.Errorf("failed to update alert with issue_id: %w", err)
			}

			logger.Info("issue created from alert",
				zap.Uint64("alert_id", alert.ID),
				zap.Uint64("issue_id", issueID),
			)

			break
		}
	}

	return nil
}

// createIssueFromAlert 从告警创建工单
func (s *alertService) createIssueFromAlert(
	ctx context.Context,
	alert *model.Alert,
	rule *model.AlertRule,
	labels map[string]string,
	annotations map[string]string,
) (uint64, string, error) {
	// 获取项目信息
	project, err := s.projectRepo.GetByID(ctx, rule.ProjectID)
	if err != nil {
		return 0, "", fmt.Errorf("failed to get project: %w", err)
	}

	// 获取工单类型（用于验证）
	_, err = s.issueTypeRepo.GetByID(ctx, rule.IssueTypeID)
	if err != nil {
		return 0, "", fmt.Errorf("failed to get issue type: %w", err)
	}

	// 构建工单标题和描述
	title := fmt.Sprintf("[告警] %s (1 个实例)", alert.AlertName)

	description := s.buildIssueDescription(alert, labels, annotations)

	// 生成工单 Key
	issueKey, err := s.generateIssueKey(ctx, project.ProjectKey)
	if err != nil {
		return 0, "", fmt.Errorf("failed to generate issue key: %w", err)
	}

	// 确定指派人：规则配置 > 项目负责人（lead_user_id）
	assigneeID := rule.AssigneeID
	if assigneeID == nil && project.LeadUserID > 0 {
		assigneeID = &project.LeadUserID
	}

	// 获取告警机器人用户 ID 作为创建人
	var reporterID uint64 = 1
	var botUser model.User
	if err := s.db.Where("username = ?", "alert-bot").First(&botUser).Error; err == nil {
		reporterID = botUser.ID
	}

	// 创建工单
	issue := &model.Issue{
		IssueKey:    issueKey,
		ProjectID:   rule.ProjectID,
		IssueTypeID: rule.IssueTypeID,
		Title:       title,
		Description: description,
		Priority:    rule.Priority,
		Status:      "open",
		ReporterID:  reporterID,
		AssigneeID:  assigneeID,
	}

	if err := s.issueRepo.Create(ctx, issue); err != nil {
		return 0, "", fmt.Errorf("failed to create issue: %w", err)
	}

	return issue.ID, issueKey, nil
}

// buildIssueDescription 构建工单描述
func (s *alertService) buildIssueDescription(
	alert *model.Alert,
	labels map[string]string,
	annotations map[string]string,
) string {
	var desc strings.Builder

	// 告警基本信息
	desc.WriteString("## 告警信息\n\n")
	desc.WriteString(fmt.Sprintf("- **告警名称**: %s\n", alert.AlertName))
	desc.WriteString(fmt.Sprintf("- **严重程度**: %s\n", alert.Severity))
	desc.WriteString(fmt.Sprintf("- **告警状态**: %s\n", alert.Status))
	desc.WriteString(fmt.Sprintf("- **开始时间**: %s\n", alert.StartsAt.Format("2006-01-02 15:04:05")))
	desc.WriteString(fmt.Sprintf("- **告警指纹**: %s\n", alert.Fingerprint))

	// 告警描述
	if description, ok := annotations["description"]; ok {
		desc.WriteString(fmt.Sprintf("\n**描述**: %s\n", description))
	}

	// 告警标签
	if len(labels) > 0 {
		desc.WriteString("\n## 标签\n\n")
		for k, v := range labels {
			desc.WriteString(fmt.Sprintf("- **%s**: %s\n", k, v))
		}
	}

	// 其他注解
	if len(annotations) > 0 {
		desc.WriteString("\n## 详细信息\n\n")
		for k, v := range annotations {
			if k != "summary" && k != "description" {
				desc.WriteString(fmt.Sprintf("- **%s**: %s\n", k, v))
			}
		}
	}

	desc.WriteString("\n---\n")
	desc.WriteString("*此工单由告警系统自动创建*\n")

	return desc.String()
}

// generateIssueKey 生成工单 Key
func (s *alertService) generateIssueKey(ctx context.Context, projectKey string) (string, error) {
	// 查询该项目下最新的工单序号（包含软删除的记录，避免 key 冲突）
	var maxSeq int
	err := s.db.WithContext(ctx).
		Unscoped().
		Model(&model.Issue{}).
		Where("issue_key LIKE ?", projectKey+"-%").
		Select("COALESCE(MAX(CAST(SUBSTRING(issue_key, LENGTH(?) + 2) AS UNSIGNED)), 0)", projectKey).
		Scan(&maxSeq).Error

	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%s-%d", projectKey, maxSeq+1), nil
}

// appendAlertToIssue 合并告警时追加实例信息到工单
func (s *alertService) appendAlertToIssue(
	ctx context.Context,
	issueID uint64,
	alert *model.Alert,
	labels map[string]string,
	annotations map[string]string,
) error {
	issue, err := s.issueRepo.GetByID(ctx, issueID)
	if err != nil {
		return fmt.Errorf("failed to get issue: %w", err)
	}

	// 统计该工单关联的告警数量
	linkedAlerts, err := s.alertRepo.ListByIssueID(ctx, issueID)
	if err != nil {
		return fmt.Errorf("failed to count linked alerts: %w", err)
	}
	alertCount := len(linkedAlerts)

	// 更新标题：[告警] AlertName (N 个实例)
	issue.Title = fmt.Sprintf("[告警] %s (%d 个实例)", alert.AlertName, alertCount)

	// 追加实例信息到描述
	instance := labels["instance"]
	if instance == "" {
		instance = labels["target_ident"]
	}
	nodename := labels["nodename"]
	description := ""
	if d, ok := annotations["description"]; ok {
		description = d
	}

	appendInfo := fmt.Sprintf("\n\n---\n### 合并告警 #%d (%s)\n", alertCount, alert.StartsAt.Format("2006-01-02 15:04:05"))
	appendInfo += fmt.Sprintf("- **实例**: %s\n", instance)
	if nodename != "" {
		appendInfo += fmt.Sprintf("- **主机名**: %s\n", nodename)
	}
	appendInfo += fmt.Sprintf("- **指纹**: %s\n", alert.Fingerprint)
	if description != "" {
		appendInfo += fmt.Sprintf("- **描述**: %s\n", description)
	}

	issue.Description += appendInfo

	if err := s.issueRepo.Update(ctx, issue); err != nil {
		return fmt.Errorf("failed to update issue: %w", err)
	}

	logger.Info("appended alert info to issue",
		zap.Uint64("issue_id", issueID),
		zap.String("instance", instance),
		zap.Int("alert_count", alertCount),
	)

	return nil
}

// addSystemComment 添加系统评论（UserID=0 表示系统）
func (s *alertService) addSystemComment(ctx context.Context, issueID uint64, content string) error {
	comment := &model.IssueComment{
		IssueID: issueID,
		UserID:  0,
		Content: content,
	}
	if err := s.commentRepo.Create(ctx, comment); err != nil {
		return fmt.Errorf("failed to create system comment: %w", err)
	}
	return nil
}

// autoUpdateIssueOnRecovery 告警恢复后更新工单状态
// autoResolve=true → resolved（自动关闭）；autoResolve=false → pending_review（待确认）
func (s *alertService) autoUpdateIssueOnRecovery(ctx context.Context, issueID uint64, autoResolve bool) error {
	issue, err := s.issueRepo.GetByID(ctx, issueID)
	if err != nil {
		return fmt.Errorf("failed to get issue: %w", err)
	}

	// 如果工单已经是终态，不需要处理
	if issue.Status == "resolved" || issue.Status == "closed" || issue.Status == "merged" || issue.Status == "pending_review" {
		logger.Info("autoUpdateIssueOnRecovery: issue already in terminal status, skipping",
			zap.Uint64("issue_id", issueID),
			zap.String("current_status", issue.Status),
			zap.Bool("auto_resolve", autoResolve),
		)
		return nil
	}

	now := time.Now()

	if autoResolve {
		issue.Status = "resolved"
		issue.ResolvedAt = &now
		if err := s.issueRepo.Update(ctx, issue); err != nil {
			return fmt.Errorf("failed to update issue status: %w", err)
		}

		// 同步完成工作流实例，避免工单状态与工作流状态不一致
		s.forceCompleteWorkflow(ctx, issueID)

		commentContent := "告警已自动恢复，工单已自动关闭"
		if err := s.addSystemComment(ctx, issueID, commentContent); err != nil {
			logger.Error("failed to add system comment for auto-resolve",
				zap.Uint64("issue_id", issueID),
				zap.Error(err),
			)
		}

		logger.Info("issue auto-resolved by alert recovery",
			zap.Uint64("issue_id", issueID),
		)
	} else {
		issue.Status = "pending_review"
		if err := s.issueRepo.Update(ctx, issue); err != nil {
			return fmt.Errorf("failed to update issue status: %w", err)
		}

		commentContent := fmt.Sprintf("告警已自动恢复于 %s，请确认是否可以关闭工单", now.Format("15:04"))
		if err := s.addSystemComment(ctx, issueID, commentContent); err != nil {
			logger.Error("failed to add system comment for pending_review",
				zap.Uint64("issue_id", issueID),
				zap.Error(err),
			)
		}

		logger.Info("issue set to pending_review by alert recovery",
			zap.Uint64("issue_id", issueID),
		)
	}

	return nil
}

// forceCompleteWorkflow 强制完成工单关联的工作流实例
// 当告警恢复自动关闭工单时，需要同步完成工作流，避免工单"已完成"但工作流仍"进行中"的不一致
func (s *alertService) forceCompleteWorkflow(ctx context.Context, issueID uint64) {
	now := time.Now()
	result := s.db.WithContext(ctx).
		Model(&model.WorkflowInstance{}).
		Where("issue_id = ? AND status = ?", issueID, "active").
		Updates(map[string]interface{}{
			"status":       "completed",
			"completed_at": now,
			"updated_at":   now,
		})
	if result.Error != nil {
		logger.Warn("failed to force-complete workflow instance on alert recovery",
			zap.Uint64("issue_id", issueID),
			zap.Error(result.Error),
		)
	} else if result.RowsAffected > 0 {
		logger.Info("workflow instance force-completed by alert recovery",
			zap.Uint64("issue_id", issueID),
		)
	}
}

// TryAutoResolveIssue 检查工单关联的所有告警是否都已恢复，如果是则根据规则设置更新工单状态
func (s *alertService) TryAutoResolveIssue(ctx context.Context, issueID uint64) error {
	// 获取该工单下所有告警
	alerts, err := s.alertRepo.ListByIssueID(ctx, issueID)
	if err != nil {
		return fmt.Errorf("failed to list alerts by issue_id: %w", err)
	}

	if len(alerts) == 0 {
		return nil
	}

	// 检查是否所有告警都已恢复
	for _, a := range alerts {
		if a.Status == "firing" {
			return nil // 还有活跃告警，不关闭工单
		}
	}

	// 遍历所有告警尝试匹配规则（不同告警可能匹配不同规则）
	autoResolve := false
	matchedAny := false
	for _, a := range alerts {
		rule, err := s.getMatchedRule(ctx, a)
		if err != nil {
			logger.Error("TryAutoResolveIssue: failed to get matched rule",
				zap.Uint64("alert_id", a.ID),
				zap.Error(err),
			)
			continue
		}
		if rule != nil {
			matchedAny = true
			logger.Info("TryAutoResolveIssue: matched rule",
				zap.Uint64("issue_id", issueID),
				zap.Uint64("alert_id", a.ID),
				zap.Uint64("rule_id", rule.ID),
				zap.String("rule_name", rule.Name),
				zap.Bool("rule_auto_resolve", rule.AutoResolve),
			)
			if rule.AutoResolve {
				autoResolve = true
				break // 只要有一条规则是 AutoResolve=true，就自动关闭
			}
		}
	}

	if !matchedAny {
		// 无匹配规则，默认设为待确认
		logger.Info("TryAutoResolveIssue: no matching rule found, setting to pending_review",
			zap.Uint64("issue_id", issueID),
		)
	}

	// 所有告警都已恢复，根据 autoResolve 设置更新工单
	if err := s.autoUpdateIssueOnRecovery(ctx, issueID, autoResolve); err != nil {
		return fmt.Errorf("failed to auto update issue on recovery: %w", err)
	}

	logger.Info("TryAutoResolveIssue: all alerts resolved, issue updated",
		zap.Uint64("issue_id", issueID),
		zap.Bool("auto_resolve", autoResolve),
		zap.Int("alert_count", len(alerts)),
	)
	return nil
}

// getMatchedRule 获取匹配的告警规则
func (s *alertService) getMatchedRule(ctx context.Context, alert *model.Alert) (*model.AlertRule, error) {
	rules, err := s.alertRuleRepo.ListEnabled(ctx)
	if err != nil {
		return nil, err
	}

	labels := repository.JSONToLabels(alert.Labels)
	for _, rule := range rules {
		if s.matchRule(rule, labels) {
			return rule, nil
		}
	}

	return nil, nil
}


// matchRule 匹配告警规则
func (s *alertService) matchRule(rule *model.AlertRule, labels map[string]string) bool {
	// 解析规则的标签匹配器
	var matchers []dto.LabelMatcher
	if err := json.Unmarshal([]byte(rule.LabelMatchers), &matchers); err != nil {
		logger.Error("failed to unmarshal label matchers", zap.Error(err))
		return false
	}

	// 检查所有匹配器
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

// GetAlert 获取告警详情
func (s *alertService) GetAlert(ctx context.Context, id uint64) (*dto.AlertResponse, error) {
	alert, err := s.alertRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	return s.toAlertResponse(alert), nil
}

// ListAlerts 获取告警列表
func (s *alertService) ListAlerts(ctx context.Context, req *dto.AlertListRequest) (*dto.AlertListResponse, error) {
	alerts, total, err := s.alertRepo.List(ctx, req)
	if err != nil {
		return nil, err
	}

	// 批量获取关联工单的 issue_key，避免 N+1 查询
	issueKeyMap := s.batchGetIssueKeys(ctx, alerts)

	items := make([]dto.AlertResponse, 0, len(alerts))
	for _, alert := range alerts {
		resp := s.toAlertResponse(alert)
		if alert.IssueID != nil {
			if key, ok := issueKeyMap[*alert.IssueID]; ok {
				resp.IssueKey = &key
			}
		}
		items = append(items, *resp)
	}

	page := req.Page
	if page < 1 {
		page = 1
	}
	pageSize := req.PageSize
	if pageSize < 1 {
		pageSize = 20
	}

	return &dto.AlertListResponse{
		Items:    items,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	}, nil
}

// batchGetIssueKeys 批量获取告警关联工单的 issue_key
func (s *alertService) batchGetIssueKeys(ctx context.Context, alerts []*model.Alert) map[uint64]string {
	issueIDs := make([]uint64, 0)
	for _, a := range alerts {
		if a.IssueID != nil {
			issueIDs = append(issueIDs, *a.IssueID)
		}
	}
	if len(issueIDs) == 0 {
		return nil
	}

	var issues []model.Issue
	s.db.WithContext(ctx).Select("id, issue_key").Where("id IN ?", issueIDs).Find(&issues)

	keyMap := make(map[uint64]string, len(issues))
	for _, iss := range issues {
		keyMap[iss.ID] = iss.IssueKey
	}
	return keyMap
}

// GetAlertStats 获取告警统计数据
func (s *alertService) GetAlertStats(ctx context.Context) (*dto.AlertStatsResponse, error) {
	return s.alertRepo.Stats(ctx)
}

// AckAlert 确认告警
func (s *alertService) AckAlert(ctx context.Context, id uint64, userID uint64, req *dto.AlertAckRequest) error {
	// 检查告警是否存在
	alert, err := s.alertRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	// 检查是否已确认
	if alert.AckAt != nil {
		return fmt.Errorf("alert already acknowledged")
	}

	// 确认告警
	now := time.Now()
	return s.alertRepo.Ack(ctx, id, userID, now)
}

// ResolveAlert 解决告警
func (s *alertService) ResolveAlert(ctx context.Context, id uint64, userID uint64, req *dto.AlertResolveRequest) error {
	// 检查告警是否存在
	alert, err := s.alertRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	// 检查是否已解决
	if alert.Status == "resolved" {
		return fmt.Errorf("alert already resolved")
	}

	// 解决告警
	now := time.Now()
	if err := s.alertRepo.Resolve(ctx, id, userID, now); err != nil {
		return err
	}

	// 手动解决告警后，检查关联工单是否可以自动处理
	if alert.IssueID != nil {
		if err := s.TryAutoResolveIssue(ctx, *alert.IssueID); err != nil {
			logger.Error("failed to try auto resolve issue after manual alert resolve",
				zap.Uint64("alert_id", id),
				zap.Uint64("issue_id", *alert.IssueID),
				zap.Error(err),
			)
		}
	}

	return nil
}

// toAlertResponse 转换为告警响应
func (s *alertService) toAlertResponse(alert *model.Alert) *dto.AlertResponse {
	resp := &dto.AlertResponse{
		ID:          alert.ID,
		Fingerprint: alert.Fingerprint,
		Source:      alert.Source,
		AlertName:   alert.AlertName,
		Severity:    alert.Severity,
		Status:      alert.Status,
		Labels:      repository.JSONToLabels(alert.Labels),
		Annotations: repository.JSONToLabels(alert.Annotations),
		StartsAt:    alert.StartsAt,
		EndsAt:      alert.EndsAt,
		IssueID:     alert.IssueID,
		AckAt:       alert.AckAt,
		AckBy:       alert.AckBy,
		ResolvedAt:  alert.ResolvedAt,
		ResolvedBy:  alert.ResolvedBy,
		CreatedAt:   alert.CreatedAt,
		UpdatedAt:   alert.UpdatedAt,
	}

	// 填充关联工单 Key
	if alert.IssueID != nil {
		if issue, err := s.issueRepo.GetByID(context.Background(), *alert.IssueID); err == nil && issue != nil {
			resp.IssueKey = &issue.IssueKey
		}
	}

	return resp
}

// ============ 告警规则管理 ============

// CreateAlertRule 创建告警规则
func (s *alertService) CreateAlertRule(ctx context.Context, req *dto.CreateAlertRuleRequest) (*dto.AlertRuleResponse, error) {
	// 序列化标签匹配器
	matchersJSON, err := json.Marshal(req.LabelMatchers)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal label matchers: %w", err)
	}

	mergeWindow := req.MergeWindow
	if mergeWindow == 0 {
		mergeWindow = 3600 // 默认 1 小时
	}

	rule := &model.AlertRule{
		Name:          req.Name,
		Description:   req.Description,
		ProjectID:     req.ProjectID,
		IssueTypeID:   req.IssueTypeID,
		LabelMatchers: string(matchersJSON),
		Priority:      req.Priority,
		AssigneeID:    req.AssigneeID,
		AutoResolve:   req.AutoResolve,
		MergeWindow:   mergeWindow,
		Status:        1, // 默认启用
	}

	if err := s.alertRuleRepo.Create(ctx, rule); err != nil {
		return nil, err
	}

	return s.GetAlertRule(ctx, rule.ID)
}

// GetAlertRule 获取告警规则详情
func (s *alertService) GetAlertRule(ctx context.Context, id uint64) (*dto.AlertRuleResponse, error) {
	rule, err := s.alertRuleRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	return s.toAlertRuleResponse(rule), nil
}

// UpdateAlertRule 更新告警规则
func (s *alertService) UpdateAlertRule(ctx context.Context, id uint64, req *dto.UpdateAlertRuleRequest) (*dto.AlertRuleResponse, error) {
	rule, err := s.alertRuleRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// 更新字段
	if req.Name != nil {
		rule.Name = *req.Name
	}
	if req.Description != nil {
		rule.Description = *req.Description
	}
	if req.LabelMatchers != nil {
		matchersJSON, err := json.Marshal(req.LabelMatchers)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal label matchers: %w", err)
		}
		rule.LabelMatchers = string(matchersJSON)
	}
	if req.Priority != nil {
		rule.Priority = *req.Priority
	}
	if req.AssigneeID != nil {
		rule.AssigneeID = req.AssigneeID
	}
	if req.AutoResolve != nil {
		rule.AutoResolve = *req.AutoResolve
	}
	if req.MergeWindow != nil {
		rule.MergeWindow = *req.MergeWindow
	}
	if req.Status != nil {
		rule.Status = *req.Status
	}

	if err := s.alertRuleRepo.Update(ctx, rule); err != nil {
		return nil, err
	}

	return s.GetAlertRule(ctx, id)
}

// DeleteAlertRule 删除告警规则
func (s *alertService) DeleteAlertRule(ctx context.Context, id uint64) error {
	return s.alertRuleRepo.Delete(ctx, id)
}

// ListAlertRules 获取告警规则列表
func (s *alertService) ListAlertRules(ctx context.Context, req *dto.AlertRuleListRequest) (*dto.AlertRuleListResponse, error) {
	rules, total, err := s.alertRuleRepo.List(ctx, req)
	if err != nil {
		return nil, err
	}

	items := make([]dto.AlertRuleResponse, 0, len(rules))
	for _, rule := range rules {
		items = append(items, *s.toAlertRuleResponse(rule))
	}

	page := req.Page
	if page < 1 {
		page = 1
	}
	pageSize := req.PageSize
	if pageSize < 1 {
		pageSize = 20
	}

	return &dto.AlertRuleListResponse{
		Items:    items,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	}, nil
}

// toAlertRuleResponse 转换为告警规则响应
func (s *alertService) toAlertRuleResponse(rule *model.AlertRule) *dto.AlertRuleResponse {
	var matchers []dto.LabelMatcher
	json.Unmarshal([]byte(rule.LabelMatchers), &matchers)

	resp := &dto.AlertRuleResponse{
		ID:            rule.ID,
		Name:          rule.Name,
		Description:   rule.Description,
		ProjectID:     rule.ProjectID,
		IssueTypeID:   rule.IssueTypeID,
		LabelMatchers: matchers,
		Priority:      rule.Priority,
		AssigneeID:    rule.AssigneeID,
		AutoResolve:   rule.AutoResolve,
		MergeWindow:   rule.MergeWindow,
		Status:        rule.Status,
		CreatedAt:     rule.CreatedAt,
		UpdatedAt:     rule.UpdatedAt,
	}

	// 填充项目名称
	if project, err := s.projectRepo.GetByID(context.Background(), rule.ProjectID); err == nil && project != nil {
		resp.ProjectName = project.Name
	}

	// 填充工单类型名称
	if issueType, err := s.issueTypeRepo.GetByID(context.Background(), rule.IssueTypeID); err == nil && issueType != nil {
		if issueType.DisplayName != "" {
			resp.IssueTypeName = issueType.DisplayName
		} else {
			resp.IssueTypeName = issueType.Name
		}
	}

	return resp
}

