package dto

import (
	"time"

	"github.com/kerbos/ticketdesk/internal/model"
)

// CreateRequirementPoolRequest 创建需求池请求
type CreateRequirementPoolRequest struct {
	Name        string                    `json:"name" binding:"required,min=1,max=200"`
	Description string                    `json:"description" binding:"max=5000"`
	Type        model.RequirementPoolType `json:"type" binding:"required,oneof=global project"`
	ProjectID   *uint64                   `json:"project_id"`
	OwnerID     uint64                    `json:"owner_id" binding:"required"`
}

// UpdateRequirementPoolRequest 更新需求池请求
type UpdateRequirementPoolRequest struct {
	Name        *string                      `json:"name" binding:"omitempty,min=1,max=200"`
	Description *string                      `json:"description" binding:"omitempty,max=5000"`
	OwnerID     *uint64                      `json:"owner_id"`
	Status      *model.RequirementPoolStatus `json:"status" binding:"omitempty,oneof=active archived"`
}

// RequirementPoolResponse 需求池响应
type RequirementPoolResponse struct {
	ID               uint64                      `json:"id"`
	Name             string                      `json:"name"`
	Description      string                      `json:"description"`
	Type             model.RequirementPoolType   `json:"type"`
	ProjectID        *uint64                     `json:"project_id,omitempty"`
	ProjectName      string                      `json:"project_name,omitempty"`
	OwnerID          uint64                      `json:"owner_id"`
	OwnerName        string                      `json:"owner_name"`
	Status           model.RequirementPoolStatus `json:"status"`
	RequirementCount int64                       `json:"requirement_count"` // 需求数量
	CreatedAt        time.Time                   `json:"created_at"`
	UpdatedAt        time.Time                   `json:"updated_at"`
}

// RequirementPoolListRequest 需求池列表请求
type RequirementPoolListRequest struct {
	Type      *model.RequirementPoolType   `form:"type" binding:"omitempty,oneof=global project"`
	Status    *model.RequirementPoolStatus `form:"status" binding:"omitempty,oneof=active archived"`
	ProjectID *uint64                      `form:"project_id"`
	OwnerID   *uint64                      `form:"owner_id"`
	Page      int                          `form:"page" binding:"omitempty,min=1"`
	PageSize  int                          `form:"page_size" binding:"omitempty,min=1,max=100"`
}
