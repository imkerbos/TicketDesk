// Package service 提供统一的通知服务
package service

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"go.uber.org/zap"

	"github.com/kerbos/ticketdesk/internal/model"
	"github.com/kerbos/ticketdesk/internal/notification/email"
	"github.com/kerbos/ticketdesk/internal/notification/lark"
	"github.com/kerbos/ticketdesk/internal/notification/telegram"
	"github.com/kerbos/ticketdesk/internal/notification/webhook"
	"github.com/kerbos/ticketdesk/pkg/logger"
	"github.com/kerbos/ticketdesk/pkg/redis"
)

// 事件类型常量
const (
	EventIssueCreated      = "issue.created"
	EventIssueUpdated      = "issue.updated"
	EventIssueTransitioned = "issue.transitioned"
	EventIssueAssigned     = "issue.assigned"
	EventIssueCommented    = "issue.commented"
	EventAlertFiring       = "alert.firing"
	EventAlertResolved     = "alert.resolved"
	EventAlertAcked        = "alert.acked"
)

// 通知队列键
const (
	NotificationQueueKey = "notification:queue"
)

// NotificationMessage 通知消息
type NotificationMessage struct {
	Type      string      `json:"type"`       // email, webhook, lark, telegram
	Event     string      `json:"event"`      // 事件类型
	Data      interface{} `json:"data"`       // 数据
	To        []string    `json:"to"`         // 收件人（邮件）
	Subject   string      `json:"subject"`    // 主题（邮件）
	Body      string      `json:"body"`       // 内容（邮件）
	CreatedAt int64       `json:"created_at"` // 创建时间
}

// IssueEventData 工单事件数据
type IssueEventData struct {
	IssueKey    string  `json:"issue_key"`
	IssueTitle  string  `json:"issue_title"`
	ProjectKey  string  `json:"project_key"`
	ProjectName string  `json:"project_name"`
	Status      string  `json:"status"`
	Priority    string  `json:"priority"`
	ReporterID  uint64  `json:"reporter_id"`
	AssigneeID  *uint64 `json:"assignee_id"`
	UpdatedBy   uint64  `json:"updated_by"`
	Comment     string  `json:"comment,omitempty"`
}

// AlertEventData 告警事件数据
type AlertEventData struct {
	AlertID     uint64     `json:"alert_id"`
	Fingerprint string     `json:"fingerprint"`
	AlertName   string     `json:"alert_name"`
	Severity    string     `json:"severity"`
	Status      string     `json:"status"`
	Labels      string     `json:"labels"`
	StartsAt    time.Time  `json:"starts_at"`
	EndsAt      *time.Time `json:"ends_at,omitempty"`
	IssueKey    string     `json:"issue_key,omitempty"`
}

// NotificationService 通知服务接口
type NotificationService interface {
	// 同步发送
	SendEmail(ctx context.Context, to []string, subject, body string) error
	SendHTMLEmail(ctx context.Context, to []string, subject, htmlBody string) error
	SendWebhook(ctx context.Context, event string, data interface{}) error

	// 异步发送（入队）
	QueueEmail(ctx context.Context, to []string, subject, body string) error
	QueueWebhook(ctx context.Context, event string, data interface{}) error

	// 事件通知
	NotifyIssueCreated(ctx context.Context, issue *model.Issue, projectKey, projectName string) error
	NotifyIssueUpdated(ctx context.Context, issue *model.Issue, projectKey, projectName string, updatedBy uint64) error
	NotifyIssueTransitioned(ctx context.Context, issue *model.Issue, projectKey, projectName, oldStatus, newStatus string, updatedBy uint64) error
	NotifyIssueAssigned(ctx context.Context, issue *model.Issue, projectKey, projectName string, assignedBy uint64) error
	NotifyIssueCommented(ctx context.Context, issue *model.Issue, projectKey, projectName, comment string, commentBy uint64) error
	NotifyAlertFiring(ctx context.Context, alert *model.Alert) error
	NotifyAlertResolved(ctx context.Context, alert *model.Alert) error
	NotifyAlertAcked(ctx context.Context, alert *model.Alert, ackedBy uint64) error

	// 队列处理
	ProcessQueue(ctx context.Context) error
}

// notificationService 通知服务实现
type notificationService struct {
	emailService    email.EmailService
	webhookService  webhook.WebhookService
	larkService     lark.LarkService
	telegramService telegram.TelegramService
}

// NewNotificationService 创建通知服务实例
func NewNotificationService(
	emailService email.EmailService,
	webhookService webhook.WebhookService,
	larkService lark.LarkService,
	telegramService telegram.TelegramService,
) NotificationService {
	return &notificationService{
		emailService:    emailService,
		webhookService:  webhookService,
		larkService:     larkService,
		telegramService: telegramService,
	}
}

// ============ 同步发送 ============

// SendEmail 同步发送纯文本邮件
func (s *notificationService) SendEmail(ctx context.Context, to []string, subject, body string) error {
	return s.emailService.SendEmail(ctx, to, subject, body)
}

// SendHTMLEmail 同步发送 HTML 邮件
func (s *notificationService) SendHTMLEmail(ctx context.Context, to []string, subject, htmlBody string) error {
	return s.emailService.SendHTMLEmail(ctx, to, subject, htmlBody)
}

// SendWebhook 同步发送 Webhook
func (s *notificationService) SendWebhook(ctx context.Context, event string, data interface{}) error {
	return s.webhookService.SendEvent(ctx, event, data)
}

// ============ 异步发送 ============

// QueueEmail 将邮件加入队列
func (s *notificationService) QueueEmail(ctx context.Context, to []string, subject, body string) error {
	msg := NotificationMessage{
		Type:      "email",
		To:        to,
		Subject:   subject,
		Body:      body,
		CreatedAt: time.Now().Unix(),
	}
	return s.enqueue(ctx, msg)
}

// QueueWebhook 将 Webhook 加入队列
func (s *notificationService) QueueWebhook(ctx context.Context, event string, data interface{}) error {
	msg := NotificationMessage{
		Type:      "webhook",
		Event:     event,
		Data:      data,
		CreatedAt: time.Now().Unix(),
	}
	return s.enqueue(ctx, msg)
}

// QueueLark 将飞书通知加入队列
func (s *notificationService) QueueLark(ctx context.Context, event string, data interface{}) error {
	msg := NotificationMessage{
		Type:      "lark",
		Event:     event,
		Data:      data,
		CreatedAt: time.Now().Unix(),
	}
	return s.enqueue(ctx, msg)
}

// QueueTelegram 将 Telegram 通知加入队列
func (s *notificationService) QueueTelegram(ctx context.Context, event string, data interface{}) error {
	msg := NotificationMessage{
		Type:      "telegram",
		Event:     event,
		Data:      data,
		CreatedAt: time.Now().Unix(),
	}
	return s.enqueue(ctx, msg)
}

// enqueue 加入队列
func (s *notificationService) enqueue(ctx context.Context, msg NotificationMessage) error {
	if redis.Client == nil {
		// Redis 不可用，直接同步发送
		return s.processMessage(ctx, msg)
	}

	data, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("序列化消息失败: %w", err)
	}

	if err := redis.Client.RPush(ctx, NotificationQueueKey, data).Err(); err != nil {
		logger.Warn("failed to enqueue notification, sending synchronously", zap.Error(err))
		return s.processMessage(ctx, msg)
	}

	return nil
}

// ProcessQueue 处理队列中的消息
func (s *notificationService) ProcessQueue(ctx context.Context) error {
	if redis.Client == nil {
		return nil
	}

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			// 从队列取出消息（阻塞 1 秒）
			result, err := redis.Client.BLPop(ctx, time.Second, NotificationQueueKey).Result()
			if err != nil {
				if err.Error() == "redis: nil" {
					continue
				}
				logger.Error("failed to pop from queue", zap.Error(err))
				continue
			}

			if len(result) < 2 {
				continue
			}

			var msg NotificationMessage
			if err := json.Unmarshal([]byte(result[1]), &msg); err != nil {
				logger.Error("failed to unmarshal notification message", zap.Error(err))
				continue
			}

			if err := s.processMessage(ctx, msg); err != nil {
				logger.Error("failed to process notification",
					zap.String("type", msg.Type),
					zap.Error(err),
				)
			}
		}
	}
}

// processMessage 处理单条消息
func (s *notificationService) processMessage(ctx context.Context, msg NotificationMessage) error {
	switch msg.Type {
	case "email":
		return s.emailService.SendEmail(ctx, msg.To, msg.Subject, msg.Body)
	case "webhook":
		return s.webhookService.SendEvent(ctx, msg.Event, msg.Data)
	case "lark":
		return s.larkService.SendNotification(ctx, msg.Event, msg.Data)
	case "telegram":
		return s.telegramService.SendNotification(ctx, msg.Event, msg.Data)
	default:
		return fmt.Errorf("unknown notification type: %s", msg.Type)
	}
}

// ============ 事件通知 ============

// notifyAllChannels 向所有渠道发送事件通知（Webhook + 飞书 + Telegram）
func (s *notificationService) notifyAllChannels(ctx context.Context, event string, data interface{}) {
	// Webhook
	if err := s.QueueWebhook(ctx, event, data); err != nil {
		logger.Warn("failed to queue webhook notification", zap.String("event", event), zap.Error(err))
	}
	// 飞书
	if err := s.QueueLark(ctx, event, data); err != nil {
		logger.Warn("failed to queue lark notification", zap.String("event", event), zap.Error(err))
	}
	// Telegram
	if err := s.QueueTelegram(ctx, event, data); err != nil {
		logger.Warn("failed to queue telegram notification", zap.String("event", event), zap.Error(err))
	}
}

// NotifyIssueCreated 通知工单创建
func (s *notificationService) NotifyIssueCreated(ctx context.Context, issue *model.Issue, projectKey, projectName string) error {
	data := IssueEventData{
		IssueKey:    issue.IssueKey,
		IssueTitle:  issue.Title,
		ProjectKey:  projectKey,
		ProjectName: projectName,
		Status:      issue.Status,
		Priority:    issue.Priority,
		ReporterID:  issue.ReporterID,
		AssigneeID:  issue.AssigneeID,
		UpdatedBy:   issue.ReporterID,
	}
	s.notifyAllChannels(ctx, EventIssueCreated, data)
	return nil
}

// NotifyIssueUpdated 通知工单更新
func (s *notificationService) NotifyIssueUpdated(ctx context.Context, issue *model.Issue, projectKey, projectName string, updatedBy uint64) error {
	data := IssueEventData{
		IssueKey:    issue.IssueKey,
		IssueTitle:  issue.Title,
		ProjectKey:  projectKey,
		ProjectName: projectName,
		Status:      issue.Status,
		Priority:    issue.Priority,
		ReporterID:  issue.ReporterID,
		AssigneeID:  issue.AssigneeID,
		UpdatedBy:   updatedBy,
	}
	s.notifyAllChannels(ctx, EventIssueUpdated, data)
	return nil
}

// NotifyIssueTransitioned 通知工单状态变更
func (s *notificationService) NotifyIssueTransitioned(ctx context.Context, issue *model.Issue, projectKey, projectName, oldStatus, newStatus string, updatedBy uint64) error {
	data := map[string]interface{}{
		"issue_key":    issue.IssueKey,
		"issue_title":  issue.Title,
		"project_key":  projectKey,
		"project_name": projectName,
		"old_status":   oldStatus,
		"new_status":   newStatus,
		"updated_by":   updatedBy,
	}
	s.notifyAllChannels(ctx, EventIssueTransitioned, data)
	return nil
}

// NotifyIssueAssigned 通知工单指派
func (s *notificationService) NotifyIssueAssigned(ctx context.Context, issue *model.Issue, projectKey, projectName string, assignedBy uint64) error {
	data := IssueEventData{
		IssueKey:    issue.IssueKey,
		IssueTitle:  issue.Title,
		ProjectKey:  projectKey,
		ProjectName: projectName,
		Status:      issue.Status,
		Priority:    issue.Priority,
		ReporterID:  issue.ReporterID,
		AssigneeID:  issue.AssigneeID,
		UpdatedBy:   assignedBy,
	}
	s.notifyAllChannels(ctx, EventIssueAssigned, data)
	return nil
}

// NotifyIssueCommented 通知工单评论
func (s *notificationService) NotifyIssueCommented(ctx context.Context, issue *model.Issue, projectKey, projectName, comment string, commentBy uint64) error {
	data := IssueEventData{
		IssueKey:    issue.IssueKey,
		IssueTitle:  issue.Title,
		ProjectKey:  projectKey,
		ProjectName: projectName,
		Status:      issue.Status,
		Priority:    issue.Priority,
		ReporterID:  issue.ReporterID,
		AssigneeID:  issue.AssigneeID,
		UpdatedBy:   commentBy,
		Comment:     comment,
	}
	s.notifyAllChannels(ctx, EventIssueCommented, data)
	return nil
}

// NotifyAlertFiring 通知告警触发
func (s *notificationService) NotifyAlertFiring(ctx context.Context, alert *model.Alert) error {
	data := AlertEventData{
		AlertID:     alert.ID,
		Fingerprint: alert.Fingerprint,
		AlertName:   alert.AlertName,
		Severity:    alert.Severity,
		Status:      alert.Status,
		Labels:      alert.Labels,
		StartsAt:    alert.StartsAt,
	}
	s.notifyAllChannels(ctx, EventAlertFiring, data)
	return nil
}

// NotifyAlertResolved 通知告警恢复
func (s *notificationService) NotifyAlertResolved(ctx context.Context, alert *model.Alert) error {
	data := AlertEventData{
		AlertID:     alert.ID,
		Fingerprint: alert.Fingerprint,
		AlertName:   alert.AlertName,
		Severity:    alert.Severity,
		Status:      alert.Status,
		Labels:      alert.Labels,
		StartsAt:    alert.StartsAt,
		EndsAt:      alert.EndsAt,
	}
	s.notifyAllChannels(ctx, EventAlertResolved, data)
	return nil
}

// NotifyAlertAcked 通知告警确认
func (s *notificationService) NotifyAlertAcked(ctx context.Context, alert *model.Alert, ackedBy uint64) error {
	data := map[string]interface{}{
		"alert_id":    alert.ID,
		"fingerprint": alert.Fingerprint,
		"alert_name":  alert.AlertName,
		"severity":    alert.Severity,
		"status":      alert.Status,
		"acked_by":    ackedBy,
		"acked_at":    alert.AckAt,
	}
	s.notifyAllChannels(ctx, EventAlertAcked, data)
	return nil
}
