// Package dto 定义告警模块的数据传输对象
package dto

import "time"

// ============ Webhook 相关 DTO ============

// AlertWebhookRequest Prometheus/Alertmanager Webhook 请求
type AlertWebhookRequest struct {
	Version           string                  `json:"version"`
	GroupKey          string                  `json:"groupKey"`
	TruncatedAlerts   int                     `json:"truncatedAlerts"`
	Status            string                  `json:"status"` // firing, resolved
	Receiver          string                  `json:"receiver"`
	GroupLabels       map[string]string       `json:"groupLabels"`
	CommonLabels      map[string]string       `json:"commonLabels"`
	CommonAnnotations map[string]string       `json:"commonAnnotations"`
	ExternalURL       string                  `json:"externalURL"`
	Alerts            []AlertWebhookAlertItem `json:"alerts"`
}

// AlertWebhookAlertItem Webhook 中的单个告警
type AlertWebhookAlertItem struct {
	Status       string            `json:"status"` // firing, resolved
	Labels       map[string]string `json:"labels"`
	Annotations  map[string]string `json:"annotations"`
	StartsAt     time.Time         `json:"startsAt"`
	EndsAt       *time.Time        `json:"endsAt,omitempty"`
	GeneratorURL string            `json:"generatorURL"`
	Fingerprint  string            `json:"fingerprint"`
}

// ============ 告警查询相关 DTO ============

// AlertListRequest 告警列表查询请求
type AlertListRequest struct {
	Page         int    `form:"page" binding:"omitempty,min=1"`
	PageSize     int    `form:"page_size" binding:"omitempty,min=1,max=100"`
	Status       string `form:"status" binding:"omitempty,oneof=firing resolved"`
	Severity     string `form:"severity" binding:"omitempty,oneof=critical warning info"`
	Source       string `form:"source"`
	AlertName    string `form:"alert_name"`
	IssueID      uint64 `form:"issue_id"`
	LabelFilters string `form:"label_filters"` // 标签筛选，格式: key==value,key!=value 多个用逗号分隔
}

// AlertResponse 告警响应
type AlertResponse struct {
	ID             uint64            `json:"id"`
	Fingerprint    string            `json:"fingerprint"`
	Source         string            `json:"source"`
	AlertName      string            `json:"alert_name"`
	Severity       string            `json:"severity"`
	Status         string            `json:"status"`
	Labels         map[string]string `json:"labels"`
	Annotations    map[string]string `json:"annotations"`
	StartsAt       time.Time         `json:"starts_at"`
	EndsAt         *time.Time        `json:"ends_at,omitempty"`
	IssueID        *uint64           `json:"issue_id,omitempty"`
	IssueKey       *string           `json:"issue_key,omitempty"`
	AckAt          *time.Time        `json:"ack_at,omitempty"`
	AckBy          *uint64           `json:"ack_by,omitempty"`
	AckByName      *string           `json:"ack_by_name,omitempty"`
	ResolvedAt     *time.Time        `json:"resolved_at,omitempty"`
	ResolvedBy     *uint64           `json:"resolved_by,omitempty"`
	ResolvedByName *string           `json:"resolved_by_name,omitempty"`
	CreatedAt      time.Time         `json:"created_at"`
	UpdatedAt      time.Time         `json:"updated_at"`
}

// AlertListResponse 告警列表响应
type AlertListResponse struct {
	Items    []AlertResponse `json:"items"`
	Total    int64           `json:"total"`
	Page     int             `json:"page"`
	PageSize int             `json:"page_size"`
}

// ============ 告警操作相关 DTO ============

// AlertAckRequest 告警确认请求
type AlertAckRequest struct {
	Comment string `json:"comment" binding:"max=500"`
}

// AlertResolveRequest 告警解决请求
type AlertResolveRequest struct {
	Comment string `json:"comment" binding:"max=500"`
}

// ============ 告警规则相关 DTO ============

// CreateAlertRuleRequest 创建告警规则请求
type CreateAlertRuleRequest struct {
	Name          string         `json:"name" binding:"required,min=1,max=100"`
	Description   string         `json:"description" binding:"max=500"`
	ProjectID     uint64         `json:"project_id" binding:"required"`
	IssueTypeID   uint64         `json:"issue_type_id" binding:"required"`
	LabelMatchers []LabelMatcher `json:"label_matchers" binding:"required,min=1"`
	Priority      string         `json:"priority" binding:"required,oneof=P0 P1 P2 P3"`
	AssigneeID    *uint64        `json:"assignee_id"`
	AutoResolve   bool           `json:"auto_resolve"`
	MergeWindow   int            `json:"merge_window" binding:"omitempty,min=0"` // 合并窗口（秒），0表示不合并
}

// UpdateAlertRuleRequest 更新告警规则请求
type UpdateAlertRuleRequest struct {
	Name          *string        `json:"name" binding:"omitempty,min=1,max=100"`
	Description   *string        `json:"description" binding:"omitempty,max=500"`
	LabelMatchers []LabelMatcher `json:"label_matchers" binding:"omitempty,min=1"`
	Priority      *string        `json:"priority" binding:"omitempty,oneof=P0 P1 P2 P3"`
	AssigneeID    *uint64        `json:"assignee_id"`
	AutoResolve   *bool          `json:"auto_resolve"`
	MergeWindow   *int           `json:"merge_window" binding:"omitempty,min=0"`
	Status        *int8          `json:"status" binding:"omitempty,oneof=0 1"`
}

// LabelMatcher 标签匹配器
type LabelMatcher struct {
	Key      string `json:"key" binding:"required"`
	Operator string `json:"operator" binding:"required,oneof== != =~ !~"` // ==, !=, =~(regex), !~(not regex)
	Value    string `json:"value" binding:"required"`
}

// AlertRuleResponse 告警规则响应
type AlertRuleResponse struct {
	ID            uint64         `json:"id"`
	Name          string         `json:"name"`
	Description   string         `json:"description"`
	ProjectID     uint64         `json:"project_id"`
	ProjectKey    string         `json:"project_key"`
	ProjectName   string         `json:"project_name"`
	IssueTypeID   uint64         `json:"issue_type_id"`
	IssueTypeName string         `json:"issue_type_name"`
	LabelMatchers []LabelMatcher `json:"label_matchers"`
	Priority      string         `json:"priority"`
	AssigneeID    *uint64        `json:"assignee_id,omitempty"`
	AssigneeName  *string        `json:"assignee_name,omitempty"`
	AutoResolve   bool           `json:"auto_resolve"`
	MergeWindow   int            `json:"merge_window"`
	Status        int8           `json:"status"`
	CreatedAt     time.Time      `json:"created_at"`
	UpdatedAt     time.Time      `json:"updated_at"`
}

// AlertRuleListRequest 告警规则列表查询请求
type AlertRuleListRequest struct {
	Page      int    `form:"page" binding:"omitempty,min=1"`
	PageSize  int    `form:"page_size" binding:"omitempty,min=1,max=100"`
	ProjectID uint64 `form:"project_id"`
	Status    *int8  `form:"status" binding:"omitempty,oneof=0 1"`
}

// AlertRuleListResponse 告警规则列表响应
type AlertRuleListResponse struct {
	Items    []AlertRuleResponse `json:"items"`
	Total    int64               `json:"total"`
	Page     int                 `json:"page"`
	PageSize int                 `json:"page_size"`
}

// ============ 告警静默相关 DTO ============

// CreateAlertSilenceRequest 创建告警静默请求
type CreateAlertSilenceRequest struct {
	Name          string         `json:"name" binding:"required,min=1,max=100"`
	Description   string         `json:"description" binding:"max=500"`
	LabelMatchers []LabelMatcher `json:"label_matchers" binding:"required,min=1"`
	StartsAt      time.Time      `json:"starts_at" binding:"required"`
	EndsAt        time.Time      `json:"ends_at" binding:"required,gtfield=StartsAt"`
	Comment       string         `json:"comment" binding:"max=500"`
}

// UpdateAlertSilenceRequest 更新告警静默请求
type UpdateAlertSilenceRequest struct {
	Name          *string        `json:"name" binding:"omitempty,min=1,max=100"`
	Description   *string        `json:"description" binding:"omitempty,max=500"`
	LabelMatchers []LabelMatcher `json:"label_matchers" binding:"omitempty,min=1"`
	StartsAt      *time.Time     `json:"starts_at"`
	EndsAt        *time.Time     `json:"ends_at"`
	Comment       *string        `json:"comment" binding:"omitempty,max=500"`
	Status        *int8          `json:"status" binding:"omitempty,oneof=0 1"`
}

// AlertSilenceResponse 告警静默响应
type AlertSilenceResponse struct {
	ID            uint64         `json:"id"`
	Name          string         `json:"name"`
	Description   string         `json:"description"`
	LabelMatchers []LabelMatcher `json:"label_matchers"`
	StartsAt      time.Time      `json:"starts_at"`
	EndsAt        time.Time      `json:"ends_at"`
	CreatedBy     uint64         `json:"created_by"`
	CreatedByName string         `json:"created_by_name"`
	Comment       string         `json:"comment"`
	Status        int8           `json:"status"`
	CreatedAt     time.Time      `json:"created_at"`
	UpdatedAt     time.Time      `json:"updated_at"`
}

// AlertSilenceListRequest 告警静默列表查询请求
type AlertSilenceListRequest struct {
	Page     int   `form:"page" binding:"omitempty,min=1"`
	PageSize int   `form:"page_size" binding:"omitempty,min=1,max=100"`
	Status   *int8 `form:"status" binding:"omitempty,oneof=0 1 2"`
}

// AlertSilenceListResponse 告警静默列表响应
type AlertSilenceListResponse struct {
	Items    []AlertSilenceResponse `json:"items"`
	Total    int64                  `json:"total"`
	Page     int                    `json:"page"`
	PageSize int                    `json:"page_size"`
}

// ============ 告警分组相关 DTO ============

// AlertGroupRequest 告警分组查询请求
type AlertGroupRequest struct {
	GroupBy  string `form:"group_by" binding:"required"` // 分组字段（标签键）
	Status   string `form:"status" binding:"omitempty,oneof=firing resolved"`
	Severity string `form:"severity" binding:"omitempty,oneof=critical warning info"`
}

// AlertGroupItem 告警分组项
type AlertGroupItem struct {
	GroupValue string           `json:"group_value"` // 分组值
	Count      int64            `json:"count"`       // 告警数量
	Severity   map[string]int64 `json:"severity"`    // 按严重程度统计
	Status     map[string]int64 `json:"status"`      // 按状态统计
}

// AlertGroupResponse 告警分组响应
type AlertGroupResponse struct {
	GroupBy string           `json:"group_by"`
	Items   []AlertGroupItem `json:"items"`
	Total   int64            `json:"total"`
}

// AlertStatsResponse 告警统计响应
type AlertStatsResponse struct {
	Total    int64 `json:"total"`
	Firing   int64 `json:"firing"`
	Resolved int64 `json:"resolved"`
	Critical int64 `json:"critical"`
	Warning  int64 `json:"warning"`
	Info     int64 `json:"info"`
}
