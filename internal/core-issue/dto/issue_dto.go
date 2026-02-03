// Package dto 定义工单模块的数据传输对象
package dto

import "time"

// ============ 请求 DTO ============

// CreateIssueRequest 创建工单请求
type CreateIssueRequest struct {
	ProjectKey  string  `json:"project_key" binding:"required,max=20"`
	IssueTypeID uint64  `json:"issue_type_id" binding:"required"`
	Title       string  `json:"title" binding:"required,min=1,max=200"`
	Description string  `json:"description" binding:"max=10000"`
	Priority    string  `json:"priority" binding:"omitempty,oneof=P0 P1 P2 P3"`
	AssigneeID  *uint64 `json:"assignee_id"`
	ParentID    *uint64 `json:"parent_id"`
	DueDate     *string `json:"due_date"` // 格式: 2006-01-02
}

// UpdateIssueRequest 更新工单请求
type UpdateIssueRequest struct {
	Title       *string `json:"title" binding:"omitempty,min=1,max=200"`
	Description *string `json:"description" binding:"omitempty,max=10000"`
	Priority    *string `json:"priority" binding:"omitempty,oneof=P0 P1 P2 P3"`
	AssigneeID  *uint64 `json:"assignee_id"`
	DueDate     *string `json:"due_date"`
	IssueTypeID *uint64 `json:"issue_type_id"`
}

// TransitionIssueRequest 工单状态流转请求
type TransitionIssueRequest struct {
	Status  string `json:"status" binding:"required,oneof=open in_progress resolved closed reopened"`
	Comment string `json:"comment" binding:"max=1000"`
}

// ListIssuesRequest 工单列表请求
type ListIssuesRequest struct {
	Page        int     `form:"page" binding:"omitempty,min=1"`
	PageSize    int     `form:"page_size" binding:"omitempty,min=1,max=100"`
	ProjectKey  string  `form:"project_key" binding:"omitempty,max=20"`
	Status      string  `form:"status" binding:"omitempty"`
	Priority    string  `form:"priority" binding:"omitempty"`
	AssigneeID  *uint64 `form:"assignee_id"`
	ReporterID  *uint64 `form:"reporter_id"`
	IssueTypeID *uint64 `form:"issue_type_id"`
	Keyword     string  `form:"keyword" binding:"omitempty,max=100"`
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

// ============ 响应 DTO ============

// IssueResponse 工单响应
type IssueResponse struct {
	ID          uint64             `json:"id"`
	IssueKey    string             `json:"issue_key"`
	ProjectID   uint64             `json:"project_id"`
	ProjectKey  string             `json:"project_key"`
	IssueTypeID uint64             `json:"issue_type_id"`
	IssueType   *IssueTypeBrief    `json:"issue_type,omitempty"`
	Title       string             `json:"title"`
	Description string             `json:"description"`
	Priority    string             `json:"priority"`
	Status      string             `json:"status"`
	ReporterID  uint64             `json:"reporter_id"`
	Reporter    *UserBrief         `json:"reporter,omitempty"`
	AssigneeID  *uint64            `json:"assignee_id"`
	Assignee    *UserBrief         `json:"assignee,omitempty"`
	ParentID    *uint64            `json:"parent_id"`
	ParentKey   string             `json:"parent_key,omitempty"`
	DueDate     *time.Time         `json:"due_date"`
	ResolvedAt  *time.Time         `json:"resolved_at"`
	ClosedAt    *time.Time         `json:"closed_at"`
	CreatedAt   time.Time          `json:"created_at"`
	UpdatedAt   time.Time          `json:"updated_at"`
	Comments    []*CommentResponse `json:"comments,omitempty"`
	Watchers    []*UserBrief       `json:"watchers,omitempty"`
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
