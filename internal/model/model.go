// Package model 定义基础数据模型
package model

import (
	"time"

	"gorm.io/gorm"
)

// BaseModel 基础模型，所有模型都应嵌入此结构
type BaseModel struct {
	ID        uint64         `gorm:"primaryKey;autoIncrement" json:"id"`
	CreatedAt time.Time      `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt time.Time      `gorm:"autoUpdateTime" json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
}

// User 用户模型
type User struct {
	BaseModel
	Username     string `gorm:"size:50;uniqueIndex;not null" json:"username"`
	Email        string `gorm:"size:100;uniqueIndex;not null" json:"email"`
	PasswordHash string `gorm:"size:255;not null" json:"-"`
	DisplayName  string `gorm:"size:100" json:"display_name"`
	AvatarURL    string `gorm:"size:255" json:"avatar_url"`
	Status       int8   `gorm:"default:1;index" json:"status"` // 0-禁用, 1-启用
}

// TableName 指定表名
func (User) TableName() string {
	return "users"
}

// Role 角色模型
type Role struct {
	BaseModel
	Name        string `gorm:"size:50;uniqueIndex;not null" json:"name"`
	DisplayName string `gorm:"size:100" json:"display_name"`
	Description string `gorm:"type:text" json:"description"`
}

// TableName 指定表名
func (Role) TableName() string {
	return "roles"
}

// UserRole 用户角色关联模型
type UserRole struct {
	ID        uint64    `gorm:"primaryKey;autoIncrement" json:"id"`
	UserID    uint64    `gorm:"uniqueIndex:uk_user_role;index;not null" json:"user_id"`
	RoleID    uint64    `gorm:"uniqueIndex:uk_user_role;index;not null" json:"role_id"`
	CreatedAt time.Time `gorm:"autoCreateTime" json:"created_at"`
}

// TableName 指定表名
func (UserRole) TableName() string {
	return "user_roles"
}

// Project 项目模型
type Project struct {
	BaseModel
	ProjectKey  string `gorm:"size:20;uniqueIndex;not null" json:"project_key"`
	Name        string `gorm:"size:100;not null" json:"name"`
	Description string `gorm:"type:text" json:"description"`
	LeadUserID  uint64 `gorm:"index" json:"lead_user_id"`
	Status      int8   `gorm:"default:1;index" json:"status"` // 0-归档, 1-活跃
}

// TableName 指定表名
func (Project) TableName() string {
	return "projects"
}

// ProjectMember 项目成员模型
type ProjectMember struct {
	ID        uint64    `gorm:"primaryKey;autoIncrement" json:"id"`
	ProjectID uint64    `gorm:"uniqueIndex:uk_project_member;index;not null" json:"project_id"`
	UserID    uint64    `gorm:"uniqueIndex:uk_project_member;index;not null" json:"user_id"`
	Role      string    `gorm:"size:20;default:member;index" json:"role"` // owner, admin, member
	CreatedAt time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt time.Time `gorm:"autoUpdateTime" json:"updated_at"`
}

// TableName 指定表名
func (ProjectMember) TableName() string {
	return "project_members"
}

// IssueType 工单类型模型
type IssueType struct {
	BaseModel
	ProjectID   *uint64 `gorm:"index" json:"project_id"` // NULL 表示全局类型
	Name        string  `gorm:"size:50;index;not null" json:"name"`
	DisplayName string  `gorm:"size:100" json:"display_name"`
	Description string  `gorm:"type:text" json:"description"`
	Icon        string  `gorm:"size:50" json:"icon"`
	Color       string  `gorm:"size:20" json:"color"`
}

// TableName 指定表名
func (IssueType) TableName() string {
	return "issue_types"
}

// Issue 工单模型
type Issue struct {
	BaseModel
	IssueKey    string     `gorm:"size:30;uniqueIndex;not null" json:"issue_key"`
	ProjectID   uint64     `gorm:"index;not null" json:"project_id"`
	IssueTypeID uint64     `gorm:"index;not null" json:"issue_type_id"`
	Title       string     `gorm:"size:200;not null" json:"title"`
	Description string     `gorm:"type:text" json:"description"`
	Priority    string     `gorm:"size:10;default:P2;index" json:"priority"`
	Status      string     `gorm:"size:30;default:open;index" json:"status"`
	ReporterID  uint64     `gorm:"index;not null" json:"reporter_id"`
	AssigneeID  *uint64    `gorm:"index" json:"assignee_id"`
	ParentID    *uint64    `gorm:"index" json:"parent_id"`
	DueDate     *time.Time `json:"due_date"`
	ResolvedAt  *time.Time `json:"resolved_at"`
	ClosedAt    *time.Time `json:"closed_at"`
}

// TableName 指定表名
func (Issue) TableName() string {
	return "issues"
}

// IssueComment 工单评论模型
type IssueComment struct {
	BaseModel
	IssueID uint64 `gorm:"index;not null" json:"issue_id"`
	UserID  uint64 `gorm:"index;not null" json:"user_id"`
	Content string `gorm:"type:text;not null" json:"content"`
}

// TableName 指定表名
func (IssueComment) TableName() string {
	return "issue_comments"
}

// IssueWatcher 工单关注人模型
type IssueWatcher struct {
	ID        uint64    `gorm:"primaryKey;autoIncrement" json:"id"`
	IssueID   uint64    `gorm:"uniqueIndex:uk_issue_watcher;index;not null" json:"issue_id"`
	UserID    uint64    `gorm:"uniqueIndex:uk_issue_watcher;index;not null" json:"user_id"`
	CreatedAt time.Time `gorm:"autoCreateTime" json:"created_at"`
}

// TableName 指定表名
func (IssueWatcher) TableName() string {
	return "issue_watchers"
}

// Workflow 工作流定义模型
type Workflow struct {
	BaseModel
	ProjectID   *uint64 `gorm:"index" json:"project_id"`
	Name        string  `gorm:"size:100;not null" json:"name"`
	Description string  `gorm:"type:text" json:"description"`
	Status      int8    `gorm:"default:1;index" json:"status"` // 0-禁用, 1-启用
}

// TableName 指定表名
func (Workflow) TableName() string {
	return "workflows"
}

// WorkflowNode 工作流节点模型
type WorkflowNode struct {
	BaseModel
	WorkflowID uint64 `gorm:"index;not null" json:"workflow_id"`
	Name       string `gorm:"size:100;not null" json:"name"`
	NodeType   string `gorm:"size:20;index;not null" json:"node_type"` // start, end, approval, work, system
	Config     string `gorm:"type:json" json:"config"`
	PositionX  int    `gorm:"default:0" json:"position_x"`
	PositionY  int    `gorm:"default:0" json:"position_y"`
}

// TableName 指定表名
func (WorkflowNode) TableName() string {
	return "workflow_nodes"
}

// WorkflowEdge 工作流边模型
type WorkflowEdge struct {
	BaseModel
	WorkflowID    uint64 `gorm:"index;not null" json:"workflow_id"`
	SourceNodeID  uint64 `gorm:"index;not null" json:"source_node_id"`
	TargetNodeID  uint64 `gorm:"index;not null" json:"target_node_id"`
	ConditionExpr string `gorm:"size:500" json:"condition_expr"`
}

// TableName 指定表名
func (WorkflowEdge) TableName() string {
	return "workflow_edges"
}

// Alert 告警模型
type Alert struct {
	BaseModel
	Fingerprint string     `gorm:"size:64;index;not null" json:"fingerprint"`
	Source      string     `gorm:"size:50;index;not null" json:"source"`
	AlertName   string     `gorm:"size:200;not null" json:"alert_name"`
	Severity    string     `gorm:"size:20;default:warning;index" json:"severity"`
	Status      string     `gorm:"size:20;default:firing;index" json:"status"`
	Labels      string     `gorm:"type:json" json:"labels"`
	Annotations string     `gorm:"type:json" json:"annotations"`
	StartsAt    time.Time  `gorm:"index;not null" json:"starts_at"`
	EndsAt      *time.Time `json:"ends_at"`
	IssueID     *uint64    `gorm:"index" json:"issue_id"`
	AckAt       *time.Time `json:"ack_at"`       // 确认时间
	AckBy       *uint64    `gorm:"index" json:"ack_by"` // 确认人
	ResolvedAt  *time.Time `json:"resolved_at"`  // 解决时间
	ResolvedBy  *uint64    `gorm:"index" json:"resolved_by"` // 解决人
}

// TableName 指定表名
func (Alert) TableName() string {
	return "alerts"
}

// AlertRule 告警规则模型（自动建单规则）
type AlertRule struct {
	BaseModel
	Name          string  `gorm:"size:100;not null" json:"name"`
	Description   string  `gorm:"type:text" json:"description"`
	ProjectID     uint64  `gorm:"index;not null" json:"project_id"`
	IssueTypeID   uint64  `gorm:"index;not null" json:"issue_type_id"`
	LabelMatchers string  `gorm:"type:json;not null" json:"label_matchers"` // 标签匹配规则
	Priority      string  `gorm:"size:10;default:P2" json:"priority"`
	AssigneeID    *uint64 `gorm:"index" json:"assignee_id"` // 默认指派人
	AutoResolve   bool    `gorm:"default:true" json:"auto_resolve"` // 告警恢复时自动解决工单
	MergeWindow   int     `gorm:"default:3600" json:"merge_window"` // 告警合并时间窗口（秒），0表示不合并
	Status        int8    `gorm:"default:1;index" json:"status"` // 0-禁用, 1-启用
}

// TableName 指定表名
func (AlertRule) TableName() string {
	return "alert_rules"
}

// AlertSilence 告警静默模型
type AlertSilence struct {
	BaseModel
	Name          string     `gorm:"size:100;not null" json:"name"`
	Description   string     `gorm:"type:text" json:"description"`
	LabelMatchers string     `gorm:"type:json;not null" json:"label_matchers"` // 标签匹配规则
	StartsAt      time.Time  `gorm:"index;not null" json:"starts_at"` // 静默开始时间
	EndsAt        time.Time  `gorm:"index;not null" json:"ends_at"`   // 静默结束时间
	CreatedBy     uint64     `gorm:"index;not null" json:"created_by"` // 创建人
	Comment       string     `gorm:"type:text" json:"comment"` // 静默原因
	Status        int8       `gorm:"default:1;index" json:"status"` // 0-已取消, 1-生效中, 2-已过期
}

// TableName 指定表名
func (AlertSilence) TableName() string {
	return "alert_silences"
}
