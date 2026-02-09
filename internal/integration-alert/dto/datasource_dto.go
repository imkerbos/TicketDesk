// Package dto 定义告警数据源模块的数据传输对象
package dto

import "time"

// NightingaleDatasourceConfig 夜莺数据源配置
type NightingaleDatasourceConfig struct {
	BaseURL string `json:"base_url"`
	Token   string `json:"token"`
}

// PrometheusDatasourceConfig Prometheus 数据源配置
type PrometheusDatasourceConfig struct {
	BaseURL string `json:"base_url"`
}

// CreateDatasourceRequest 创建数据源请求
type CreateDatasourceRequest struct {
	Name         string                 `json:"name" binding:"required,min=1,max=100"`
	Type         string                 `json:"type" binding:"required,oneof=prometheus nightingale"`
	Description  string                 `json:"description" binding:"max=500"`
	Config       map[string]interface{} `json:"config" binding:"required"`
	PushMode     bool                   `json:"push_mode"`
	PollInterval int                    `json:"poll_interval" binding:"omitempty,min=5,max=3600"`
}

// UpdateDatasourceRequest 更新数据源请求
type UpdateDatasourceRequest struct {
	Name         *string                `json:"name" binding:"omitempty,min=1,max=100"`
	Description  *string                `json:"description" binding:"omitempty,max=500"`
	Config       map[string]interface{} `json:"config"`
	PushMode     *bool                  `json:"push_mode"`
	PollInterval *int                   `json:"poll_interval" binding:"omitempty,min=5,max=3600"`
	Status       *int8                  `json:"status" binding:"omitempty,oneof=0 1"`
}

// DatasourceResponse 数据源响应
type DatasourceResponse struct {
	ID           uint64                 `json:"id"`
	Name         string                 `json:"name"`
	Type         string                 `json:"type"`
	Description  string                 `json:"description"`
	Config       map[string]interface{} `json:"config"`
	PushMode     bool                   `json:"push_mode"`
	PollInterval int                    `json:"poll_interval"`
	Status       int8                   `json:"status"`
	WebhookURL   string                 `json:"webhook_url"`
	LastCheckAt  *time.Time             `json:"last_check_at,omitempty"`
	LastCheckOk  *bool                  `json:"last_check_ok,omitempty"`
	LastCheckMsg string                 `json:"last_check_msg,omitempty"`
	CreatedBy    uint64                 `json:"created_by"`
	CreatedAt    time.Time              `json:"created_at"`
	UpdatedAt    time.Time              `json:"updated_at"`
}

// DatasourceListRequest 数据源列表请求
type DatasourceListRequest struct {
	Page     int    `form:"page" binding:"omitempty,min=1"`
	PageSize int    `form:"page_size" binding:"omitempty,min=1,max=100"`
	Type     string `form:"type" binding:"omitempty,oneof=prometheus nightingale"`
	Status   *int8  `form:"status" binding:"omitempty,oneof=0 1"`
}

// DatasourceListResponse 数据源列表响应
type DatasourceListResponse struct {
	Items    []DatasourceResponse `json:"items"`
	Total    int64                `json:"total"`
	Page     int                  `json:"page"`
	PageSize int                  `json:"page_size"`
}

// TestDatasourceRequest 测试数据源连接请求
type TestDatasourceRequest struct {
	Type   string                 `json:"type" binding:"required,oneof=prometheus nightingale"`
	Config map[string]interface{} `json:"config" binding:"required"`
}

// TestDatasourceResponse 测试数据源连接响应
type TestDatasourceResponse struct {
	Success   bool   `json:"success"`
	Message   string `json:"message"`
	LatencyMs int64  `json:"latency_ms"`
}
