// Package dto 定义字段模块的数据传输对象
package dto

import "time"

// ============ 请求 DTO ============

// CreateFieldRequest 创建字段定义请求
type CreateFieldRequest struct {
	FieldKey     string `json:"field_key" binding:"required,min=2,max=50"`
	FieldName    string `json:"field_name" binding:"required,min=1,max=100"`
	FieldType    string `json:"field_type" binding:"required,oneof=text textarea number date datetime select multiselect user multiuser version component label epic_link time_estimate url checkbox"`
	Description  string `json:"description" binding:"max=500"`
	Options      string `json:"options"`    // JSON格式的选项配置
	Validation   string `json:"validation"` // JSON格式的校验规则
	DefaultValue string `json:"default_value" binding:"max=500"`
}

// UpdateFieldRequest 更新字段定义请求
type UpdateFieldRequest struct {
	FieldName    *string `json:"field_name" binding:"omitempty,min=1,max=100"`
	Description  *string `json:"description" binding:"omitempty,max=500"`
	Options      *string `json:"options"`
	Validation   *string `json:"validation"`
	DefaultValue *string `json:"default_value" binding:"omitempty,max=500"`
	IsActive     *bool   `json:"is_active"`
	SortOrder    *int    `json:"sort_order"`
}

// UpdateFieldSchemeRequest 更新字段方案请求
type UpdateFieldSchemeRequest struct {
	Items []FieldSchemeItem `json:"items" binding:"required,dive"`
}

// FieldSchemeItem 字段方案项
type FieldSchemeItem struct {
	FieldID         uint64 `json:"field_id" binding:"required"`
	IsRequired      bool   `json:"is_required"`
	IsVisibleCreate bool   `json:"is_visible_create"`
	IsVisibleEdit   bool   `json:"is_visible_edit"`
	IsVisibleDetail bool   `json:"is_visible_detail"`
	SortOrder       int    `json:"sort_order"`
	DefaultValue    string `json:"default_value"`
}

// CreateVersionRequest 创建项目版本请求
type CreateVersionRequest struct {
	Name        string  `json:"name" binding:"required,min=1,max=50"`
	Description string  `json:"description" binding:"max=500"`
	ReleaseDate *string `json:"release_date"` // 格式: 2006-01-02
}

// UpdateVersionRequest 更新项目版本请求
type UpdateVersionRequest struct {
	Name        *string `json:"name" binding:"omitempty,min=1,max=50"`
	Description *string `json:"description" binding:"omitempty,max=500"`
	ReleaseDate *string `json:"release_date"`
	Status      *string `json:"status" binding:"omitempty,oneof=unreleased released archived"`
}

// CreateComponentRequest 创建项目组件请求
type CreateComponentRequest struct {
	Name        string  `json:"name" binding:"required,min=1,max=100"`
	Description string  `json:"description" binding:"max=500"`
	LeadUserID  *uint64 `json:"lead_user_id"`
}

// UpdateComponentRequest 更新项目组件请求
type UpdateComponentRequest struct {
	Name        *string `json:"name" binding:"omitempty,min=1,max=100"`
	Description *string `json:"description" binding:"omitempty,max=500"`
	LeadUserID  *uint64 `json:"lead_user_id"`
}

// CreateLabelRequest 创建标签请求
type CreateLabelRequest struct {
	Name        string `json:"name" binding:"required,min=1,max=50"`
	Color       string `json:"color" binding:"max=20"`
	Description string `json:"description" binding:"max=200"`
}

// UpdateLabelRequest 更新标签请求
type UpdateLabelRequest struct {
	Name        *string `json:"name" binding:"omitempty,min=1,max=50"`
	Color       *string `json:"color" binding:"omitempty,max=20"`
	Description *string `json:"description" binding:"omitempty,max=200"`
}

// ============ 响应 DTO ============

// FieldDefinitionResponse 字段定义响应
type FieldDefinitionResponse struct {
	ID           uint64    `json:"id"`
	ProjectID    *uint64   `json:"project_id"`
	FieldKey     string    `json:"field_key"`
	FieldName    string    `json:"field_name"`
	FieldType    string    `json:"field_type"`
	Description  string    `json:"description"`
	IsSystem     bool      `json:"is_system"`
	IsActive     bool      `json:"is_active"`
	Options      string    `json:"options"`
	Validation   string    `json:"validation"`
	DefaultValue string    `json:"default_value"`
	SortOrder    int       `json:"sort_order"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// FieldSchemeResponse 字段方案响应
type FieldSchemeResponse struct {
	ID              uint64                   `json:"id"`
	ProjectID       uint64                   `json:"project_id"`
	IssueTypeID     uint64                   `json:"issue_type_id"`
	FieldID         uint64                   `json:"field_id"`
	Field           *FieldDefinitionResponse `json:"field,omitempty"`
	IsRequired      bool                     `json:"is_required"`
	IsVisibleCreate bool                     `json:"is_visible_create"`
	IsVisibleEdit   bool                     `json:"is_visible_edit"`
	IsVisibleDetail bool                     `json:"is_visible_detail"`
	SortOrder       int                      `json:"sort_order"`
	DefaultValue    string                   `json:"default_value"`
	CreatedAt       time.Time                `json:"created_at"`
	UpdatedAt       time.Time                `json:"updated_at"`
}

// FieldValueResponse 字段值响应
type FieldValueResponse struct {
	FieldID      uint64      `json:"field_id"`
	FieldKey     string      `json:"field_key"`
	FieldName    string      `json:"field_name"`
	FieldType    string      `json:"field_type"`
	Value        interface{} `json:"value"`
	DisplayValue string      `json:"display_value"`
}

// VersionResponse 项目版本响应
type VersionResponse struct {
	ID          uint64     `json:"id"`
	ProjectID   uint64     `json:"project_id"`
	Name        string     `json:"name"`
	Description string     `json:"description"`
	ReleaseDate *time.Time `json:"release_date"`
	Status      string     `json:"status"`
	SortOrder   int        `json:"sort_order"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

// ComponentResponse 项目组件响应
type ComponentResponse struct {
	ID          uint64     `json:"id"`
	ProjectID   uint64     `json:"project_id"`
	Name        string     `json:"name"`
	Description string     `json:"description"`
	LeadUserID  *uint64    `json:"lead_user_id"`
	LeadUser    *UserBrief `json:"lead_user,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

// LabelResponse 标签响应
type LabelResponse struct {
	ID          uint64    `json:"id"`
	ProjectID   uint64    `json:"project_id"`
	Name        string    `json:"name"`
	Color       string    `json:"color"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// UserBrief 用户简要信息
type UserBrief struct {
	ID          uint64 `json:"id"`
	Username    string `json:"username"`
	DisplayName string `json:"display_name"`
	AvatarURL   string `json:"avatar_url"`
}

// ============ 方案模板 DTO ============

// CreateTemplateRequest 创建方案模板请求
type CreateTemplateRequest struct {
	Name        string `json:"name" binding:"required,min=1,max=100"`
	Description string `json:"description" binding:"max=500"`
}

// UpdateTemplateRequest 更新方案模板请求
type UpdateTemplateRequest struct {
	Name        *string `json:"name" binding:"omitempty,min=1,max=100"`
	Description *string `json:"description" binding:"omitempty,max=500"`
	IsActive    *bool   `json:"is_active"`
}

// UpdateTemplateItemsRequest 更新模板字段项请求
type UpdateTemplateItemsRequest struct {
	Items []TemplateItemInput `json:"items" binding:"required,dive"`
}

// TemplateItemInput 模板字段项输入
type TemplateItemInput struct {
	FieldID         uint64 `json:"field_id" binding:"required"`
	IsRequired      bool   `json:"is_required"`
	IsVisibleCreate bool   `json:"is_visible_create"`
	IsVisibleEdit   bool   `json:"is_visible_edit"`
	IsVisibleDetail bool   `json:"is_visible_detail"`
	SortOrder       int    `json:"sort_order"`
	DefaultValue    string `json:"default_value"`
}

// ApplyTemplateRequest 套用模板请求
type ApplyTemplateRequest struct {
	TemplateID uint64 `json:"template_id" binding:"required"`
	Mode       string `json:"mode" binding:"required,oneof=replace merge"`
}

// TemplateResponse 模板响应
type TemplateResponse struct {
	ID          uint64    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	CreatedBy   uint64    `json:"created_by"`
	IsActive    bool      `json:"is_active"`
	ItemCount   int       `json:"item_count"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// TemplateDetailResponse 模板详情响应（含字段项）
type TemplateDetailResponse struct {
	ID          uint64                  `json:"id"`
	Name        string                  `json:"name"`
	Description string                  `json:"description"`
	CreatedBy   uint64                  `json:"created_by"`
	IsActive    bool                    `json:"is_active"`
	Items       []*TemplateItemResponse `json:"items"`
	CreatedAt   time.Time               `json:"created_at"`
	UpdatedAt   time.Time               `json:"updated_at"`
}

// TemplateItemResponse 模板字段项响应
type TemplateItemResponse struct {
	ID              uint64                   `json:"id"`
	TemplateID      uint64                   `json:"template_id"`
	FieldID         uint64                   `json:"field_id"`
	Field           *FieldDefinitionResponse `json:"field,omitempty"`
	IsRequired      bool                     `json:"is_required"`
	IsVisibleCreate bool                     `json:"is_visible_create"`
	IsVisibleEdit   bool                     `json:"is_visible_edit"`
	IsVisibleDetail bool                     `json:"is_visible_detail"`
	SortOrder       int                      `json:"sort_order"`
	DefaultValue    string                   `json:"default_value"`
}

// FieldUsageResponse 字段使用情况响应
type FieldUsageResponse struct {
	SchemeCount   int64 `json:"scheme_count"`
	ValueCount    int64 `json:"value_count"`
	TemplateCount int64 `json:"template_count"`
}

// ============ 自定义字段值 ============

// CustomFieldValue 自定义字段值（用于工单创建/更新）
type CustomFieldValue struct {
	FieldID uint64      `json:"field_id"`
	Value   interface{} `json:"value"`
}
