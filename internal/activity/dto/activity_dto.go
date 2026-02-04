// Package dto 定义活动日志模块的数据传输对象
package dto

import "time"

// ============ 请求 DTO ============

// ListActivitiesRequest 活动列表请求
type ListActivitiesRequest struct {
	Page       int    `form:"page" binding:"omitempty,min=1"`
	PageSize   int    `form:"page_size" binding:"omitempty,min=1,max=100"`
	EntityType string `form:"entity_type" binding:"omitempty,oneof=issue alert project"`
	EntityID   uint64 `form:"entity_id" binding:"omitempty"`
	EntityKey  string `form:"entity_key" binding:"omitempty"`
	UserID     uint64 `form:"user_id" binding:"omitempty"`
}

// GetDefaultPage 获取默认页码
func (r *ListActivitiesRequest) GetDefaultPage() int {
	if r.Page <= 0 {
		return 1
	}
	return r.Page
}

// GetDefaultPageSize 获取默认每页数量
func (r *ListActivitiesRequest) GetDefaultPageSize() int {
	if r.PageSize <= 0 {
		return 20
	}
	return r.PageSize
}

// ============ 响应 DTO ============

// ActivityResponse 活动响应
type ActivityResponse struct {
	ID         uint64    `json:"id"`
	UserID     uint64    `json:"user_id"`
	UserName   string    `json:"user_name"`
	Action     string    `json:"action"`
	EntityType string    `json:"entity_type"`
	EntityID   uint64    `json:"entity_id"`
	EntityKey  string    `json:"entity_key,omitempty"`
	Details    string    `json:"details,omitempty"`
	CreatedAt  time.Time `json:"created_at"`
}
