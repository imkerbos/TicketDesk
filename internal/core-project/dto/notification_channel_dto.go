// Package dto 提供项目通知渠道数据传输对象
package dto

import "time"

// ============ 通知渠道配置结构 ============

// LarkChannelConfig 飞书渠道配置
type LarkChannelConfig struct {
	WebhookURL string `json:"webhook_url"`
	Secret     string `json:"secret,omitempty"`
}

// TelegramChannelConfig Telegram 渠道配置
type TelegramChannelConfig struct {
	BotToken string `json:"bot_token,omitempty"`
	ChatID   string `json:"chat_id"`
}

// ============ 请求/响应 DTO ============

// NotificationChannelResponse 通知渠道响应
type NotificationChannelResponse struct {
	ID          uint64    `json:"id"`
	ProjectID   uint64    `json:"project_id"`
	ChannelType string    `json:"channel_type"`
	Name        string    `json:"name"`
	Config      any       `json:"config"`
	EventTypes  []string  `json:"event_types"`
	MentionAll  bool      `json:"mention_all"`
	Enabled     bool      `json:"enabled"`
	CreatedBy   uint64    `json:"created_by"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// CreateNotificationChannelRequest 创建通知渠道请求
type CreateNotificationChannelRequest struct {
	ChannelType string   `json:"channel_type" binding:"required,oneof=lark telegram"`
	Name        string   `json:"name" binding:"required,max=100"`
	Config      any      `json:"config" binding:"required"`
	EventTypes  []string `json:"event_types" binding:"required,min=1,dive,oneof=issue.created issue.transitioned issue.assigned alert.merged"`
	MentionAll  bool     `json:"mention_all"`
	Enabled     bool     `json:"enabled"`
}

// UpdateNotificationChannelRequest 更新通知渠道请求
type UpdateNotificationChannelRequest struct {
	Name       *string   `json:"name" binding:"omitempty,max=100"`
	Config     any       `json:"config"`
	EventTypes *[]string `json:"event_types" binding:"omitempty,min=1,dive,oneof=issue.created issue.transitioned issue.assigned alert.merged"`
	MentionAll *bool     `json:"mention_all"`
	Enabled    *bool     `json:"enabled"`
}
