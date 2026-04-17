// Package dto 定义工单模块的数据传输对象
package dto

import "time"

// ============ 请求 DTO ============

// CreateIssueRequest 创建工单请求
type CreateIssueRequest struct {
	ProjectKey       string             `json:"project_key" binding:"required,max=20"`
	IssueTypeID      uint64             `json:"issue_type_id" binding:"required"`
	Title            string             `json:"title" binding:"required,min=1,max=200"`
	Description      string             `json:"description" binding:"max=10000"`
	Priority         string             `json:"priority" binding:"omitempty,oneof=P0 P1 P2 P3"`
	AssigneeID       *uint64            `json:"assignee_id"`
	ParentID         *uint64            `json:"parent_id"`
	EpicID           *uint64            `json:"epic_id"`
	DueDate          *string            `json:"due_date"`           // 格式: 2006-01-02
	PlannedStartDate *string            `json:"planned_start_date"` // 格式: 2006-01-02
	PlannedEndDate   *string            `json:"planned_end_date"`   // 格式: 2006-01-02
	CustomFields     []CustomFieldValue `json:"custom_fields"`
}

// CustomFieldValue 自定义字段值
type CustomFieldValue struct {
	FieldID uint64 `json:"field_id"`
	Value   any    `json:"value"`
}

// UpdateIssueRequest 更新工单请求
type UpdateIssueRequest struct {
	Title            *string            `json:"title" binding:"omitempty,min=1,max=200"`
	Description      *string            `json:"description" binding:"omitempty"`
	Priority         *string            `json:"priority" binding:"omitempty,oneof=P0 P1 P2 P3"`
	Resolution       *string            `json:"resolution" binding:"omitempty,oneof=fixed wont_fix duplicate cannot_reproduce works_as_designed incomplete done"`
	AssigneeID       *uint64            `json:"assignee_id"`
	EpicID           *uint64            `json:"epic_id"`
	DueDate          *string            `json:"due_date"`
	PlannedStartDate *string            `json:"planned_start_date"`
	PlannedEndDate   *string            `json:"planned_end_date"`
	IssueTypeID      *uint64            `json:"issue_type_id"`
	CustomFields     []CustomFieldValue `json:"custom_fields"`
}

// ListIssuesRequest 工单列表请求
type ListIssuesRequest struct {
	Page        int      `form:"page" binding:"omitempty,min=1"`
	PageSize    int      `form:"page_size" binding:"omitempty,min=1,max=100"`
	ProjectKey  string   `form:"project_key" binding:"omitempty,max=20"`
	Status      string   `form:"status" binding:"omitempty"`
	Priority    string   `form:"priority" binding:"omitempty"`
	AssigneeID  *uint64  `form:"assignee_id"`
	ReporterID  *uint64  `form:"reporter_id"`
	IssueTypeID *uint64  `form:"issue_type_id"`
	EpicID      *uint64  `form:"epic_id"`
	Keyword     string   `form:"keyword" binding:"omitempty,max=100"`
	Category    string   `form:"category" binding:"omitempty,oneof=normal alert"`
	SortBy      string   `form:"sort_by" binding:"omitempty,oneof=id priority status issue_type_id created_at updated_at"`
	Order       string   `form:"order" binding:"omitempty,oneof=asc desc"`
	StartDate   string   `form:"start_date" binding:"omitempty"` // 开始日期 YYYY-MM-DD
	EndDate     string   `form:"end_date" binding:"omitempty"`   // 结束日期 YYYY-MM-DD
	ProjectIDs  []uint64 `form:"-" json:"-"`                     // 内部使用：限定可访问的项目范围
}

// GetDefaultPage 获取默认页码
func (r *ListIssuesRequest) GetDefaultPage() int {
	if r.Page <= 0 {
		return 1
	}
	return r.Page
}

// GetDefaultPageSize 获取默认每页数量
func (r *ListIssuesRequest) GetDefaultPageSize() int {
	if r.PageSize <= 0 {
		return 20
	}
	return r.PageSize
}

// CreateCommentRequest 创建评论请求
type CreateCommentRequest struct {
	Content string `json:"content" binding:"required,min=1,max=5000"`
}

// AddWatcherRequest 添加关注人请求
type AddWatcherRequest struct {
	UserID uint64 `json:"user_id" binding:"required"`
}

// AssignIssueRequest 指派工单请求
type AssignIssueRequest struct {
	AssigneeID uint64 `json:"assignee_id" binding:"required"`
}

// CreateWorklogRequest 创建工作日志请求
type CreateWorklogRequest struct {
	Description string `json:"description" binding:"required,min=1,max=5000"`
	TimeSpent   string `json:"time_spent" binding:"required"` // 格式：2h 30m, 1d 4h, 30m
	WorkedAt    string `json:"worked_at" binding:"required"`  // ISO 8601 格式
	WorkType    string `json:"work_type" binding:"max=30"`
}

// UpdateWorklogRequest 更新工作日志请求
type UpdateWorklogRequest struct {
	Description string `json:"description" binding:"required,min=1,max=5000"`
	TimeSpent   string `json:"time_spent" binding:"required"`
	WorkedAt    string `json:"worked_at" binding:"required"`
	WorkType    string `json:"work_type" binding:"max=30"`
}

// ============ 附件相关 DTO ============

// AttachmentResponse 附件响应
type AttachmentResponse struct {
	ID         uint64     `json:"id"`
	IssueID    uint64     `json:"issue_id"`
	FileName   string     `json:"file_name"`
	FilePath   string     `json:"file_path"`
	FileSize   int64      `json:"file_size"`
	FileType   string     `json:"file_type"`
	IsImage    bool       `json:"is_image"`
	UploadedBy uint64     `json:"uploaded_by"`
	Uploader   *UserBrief `json:"uploader,omitempty"`
	CreatedAt  time.Time  `json:"created_at"`
}

// ============ 响应 DTO ============

// IssueResponse 工单响应
type IssueResponse struct {
	ID                  uint64                 `json:"id"`
	IssueKey            string                 `json:"issue_key"`
	ProjectID           uint64                 `json:"project_id"`
	ProjectKey          string                 `json:"project_key"`
	IssueTypeID         uint64                 `json:"issue_type_id"`
	IssueType           *IssueTypeBrief        `json:"issue_type,omitempty"`
	Title               string                 `json:"title"`
	Description         string                 `json:"description"`
	Priority            string                 `json:"priority"`
	Status              string                 `json:"status"`
	Resolution          string                 `json:"resolution"`
	ReporterID          uint64                 `json:"reporter_id"`
	Reporter            *UserBrief             `json:"reporter,omitempty"`
	AssigneeID          *uint64                `json:"assignee_id"`
	Assignee            *UserBrief             `json:"assignee,omitempty"`
	ParentID            *uint64                `json:"parent_id"`
	ParentKey           string                 `json:"parent_key,omitempty"`
	EpicID              *uint64                `json:"epic_id"`
	EpicKey             string                 `json:"epic_key,omitempty"`
	EpicTitle           string                 `json:"epic_title,omitempty"`
	MergedIntoIssueID   *uint64                `json:"merged_into_issue_id,omitempty"`
	MergedIntoIssueKey  string                 `json:"merged_into_issue_key,omitempty"`
	MergedFromIssueKeys []string               `json:"merged_from_issue_keys,omitempty"` // 合并来源工单 Key 列表
	DueDate             *time.Time             `json:"due_date"`
	PlannedStartDate    *time.Time             `json:"planned_start_date"`
	PlannedEndDate      *time.Time             `json:"planned_end_date"`
	ActualStartDate     *time.Time             `json:"actual_start_date"`
	ActualEndDate       *time.Time             `json:"actual_end_date"`
	ResolvedAt          *time.Time             `json:"resolved_at"`
	ClosedAt            *time.Time             `json:"closed_at"`
	CreatedAt           time.Time              `json:"created_at"`
	UpdatedAt           time.Time              `json:"updated_at"`
	Comments            []*CommentResponse     `json:"comments,omitempty"`
	Watchers            []*UserBrief           `json:"watchers,omitempty"`
	CustomFields        []*CustomFieldResponse `json:"custom_fields,omitempty"`
}

// CustomFieldResponse 自定义字段值响应
type CustomFieldResponse struct {
	FieldID      uint64 `json:"field_id"`
	FieldKey     string `json:"field_key"`
	FieldName    string `json:"field_name"`
	FieldType    string `json:"field_type"`
	Value        any    `json:"value"`
	DisplayValue string `json:"display_value"`
}

// IssueBrief 工单简要信息
type IssueBrief struct {
	ID       uint64 `json:"id"`
	IssueKey string `json:"issue_key"`
	Title    string `json:"title"`
	Status   string `json:"status"`
	Priority string `json:"priority"`
}

// IssueTypeBrief 工单类型简要信息
type IssueTypeBrief struct {
	ID          uint64 `json:"id"`
	Name        string `json:"name"`
	DisplayName string `json:"display_name"`
	Icon        string `json:"icon"`
	Color       string `json:"color"`
}

// UserBrief 用户简要信息
type UserBrief struct {
	ID          uint64 `json:"id"`
	Username    string `json:"username"`
	DisplayName string `json:"display_name"`
	AvatarURL   string `json:"avatar_url"`
}

// CommentResponse 评论响应
type CommentResponse struct {
	ID        uint64     `json:"id"`
	IssueID   uint64     `json:"issue_id"`
	UserID    uint64     `json:"user_id"`
	User      *UserBrief `json:"user,omitempty"`
	Content   string     `json:"content"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
}

// WatcherResponse 关注人响应
type WatcherResponse struct {
	ID        uint64     `json:"id"`
	IssueID   uint64     `json:"issue_id"`
	UserID    uint64     `json:"user_id"`
	User      *UserBrief `json:"user,omitempty"`
	CreatedAt time.Time  `json:"created_at"`
}

// WorklogResponse 工作日志响应
type WorklogResponse struct {
	ID           uint64     `json:"id"`
	IssueID      uint64     `json:"issue_id"`
	UserID       uint64     `json:"user_id"`
	User         *UserBrief `json:"user,omitempty"`
	Description  string     `json:"description"`
	TimeSpent    string     `json:"time_spent"`
	TimeSpentSec int        `json:"time_spent_sec"`
	WorkedAt     time.Time  `json:"worked_at"`
	WorkType     string     `json:"work_type"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
}

// IssueListStatsResponse 工单列表统计响应
type IssueListStatsResponse struct {
	Total           int64   `json:"total"`
	Resolved        int64   `json:"resolved"`
	CompletionRate  float64 `json:"completion_rate"`
	AvgResolveHours float64 `json:"avg_resolve_hours"`
}
