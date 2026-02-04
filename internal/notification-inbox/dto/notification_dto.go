// Package dto 定义站内通知数据传输对象
package dto

import "time"

// CreateNotificationRequest 创建通知请求
type CreateNotificationRequest struct {
	UserID     uint64 `json:"user_id" binding:"required"`
	Type       string `json:"type" binding:"required"`
	Title      string `json:"title" binding:"required"`
	Content    string `json:"content"`
	EntityType string `json:"entity_type" binding:"required"`
	EntityID   uint64 `json:"entity_id" binding:"required"`
	EntityKey  string `json:"entity_key"`
	ActorID    uint64 `json:"actor_id"`
	ActorName  string `json:"actor_name"`
}

// NotificationResponse 通知响应
type NotificationResponse struct {
	ID         uint64     `json:"id"`
	UserID     uint64     `json:"user_id"`
	Type       string     `json:"type"`
	Title      string     `json:"title"`
	Content    string     `json:"content"`
	EntityType string     `json:"entity_type"`
	EntityID   uint64     `json:"entity_id"`
	EntityKey  string     `json:"entity_key"`
	ActorID    uint64     `json:"actor_id"`
	ActorName  string     `json:"actor_name"`
	IsRead     bool       `json:"is_read"`
	ReadAt     *time.Time `json:"read_at,omitempty"`
	CreatedAt  time.Time  `json:"created_at"`
}

// ListNotificationsRequest 通知列表请求
type ListNotificationsRequest struct {
	Page     int    `form:"page"`
	PageSize int    `form:"page_size"`
	IsRead   *bool  `form:"is_read"`
	Type     string `form:"type"`
}

// GetDefaultPage 获取默认页码
func (r *ListNotificationsRequest) GetDefaultPage() int {
	if r.Page <= 0 {
		return 1
	}
	return r.Page
}

// GetDefaultPageSize 获取默认页大小
func (r *ListNotificationsRequest) GetDefaultPageSize() int {
	if r.PageSize <= 0 {
		return 20
	}
	if r.PageSize > 100 {
		return 100
	}
	return r.PageSize
}

// UnreadCountResponse 未读数量响应
type UnreadCountResponse struct {
	Count int64 `json:"count"`
}
