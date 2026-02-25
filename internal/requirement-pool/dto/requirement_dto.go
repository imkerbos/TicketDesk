package dto

import (
	"fmt"
	"time"

	"github.com/kerbos/ticketdesk/internal/model"
)

// 支持的日期时间格式（优先匹配带时间的格式）
var dateTimeFormats = []string{
	"2006-01-02 15:04:05",
	"2006-01-02T15:04:05Z07:00",
	"2006-01-02T15:04:05",
	"2006-01-02",
}

// ParseDateTime 解析日期时间字符串，支持多种格式
func ParseDateTime(s string) (time.Time, error) {
	for _, layout := range dateTimeFormats {
		if t, err := time.Parse(layout, s); err == nil {
			return t, nil
		}
	}
	return time.Time{}, fmt.Errorf("不支持的日期格式: %s，支持格式: 2006-01-02 或 2006-01-02 15:04:05", s)
}

// CreateRequirementRequest 创建需求请求
type CreateRequirementRequest struct {
	PoolID          uint64                    `json:"pool_id" binding:"required"`
	Title           string                    `json:"title" binding:"required,min=1,max=500"`
	Description     string                    `json:"description" binding:"max=10000"`
	Priority        model.RequirementPriority `json:"priority" binding:"required,oneof=P0 P1 P2 P3"`
	Category        model.RequirementCategory `json:"category" binding:"required,oneof=feature optimization bugfix security infrastructure other"`
	ReporterID      *uint64                   `json:"reporter_id"`
	AssigneeID      *uint64                   `json:"assignee_id"`
	StartDate       *string                   `json:"start_date"`       // 格式: 2006-01-02 或 2006-01-02 15:04:05
	EndDate         *string                   `json:"end_date"`         // 格式: 2006-01-02 或 2006-01-02 15:04:05
	TargetProjectID *uint64                   `json:"target_project_id"`
	Tags            []string                  `json:"tags"`
}

// UpdateRequirementRequest 更新需求请求
type UpdateRequirementRequest struct {
	Title           *string                    `json:"title" binding:"omitempty,min=1,max=500"`
	Description     *string                    `json:"description" binding:"omitempty,max=10000"`
	Priority        *model.RequirementPriority `json:"priority" binding:"omitempty,oneof=P0 P1 P2 P3"`
	Status          *model.RequirementStatus   `json:"status" binding:"omitempty,oneof=pending_review planning in_progress completed on_hold rejected"`
	Category        *model.RequirementCategory `json:"category" binding:"omitempty,oneof=feature optimization bugfix security infrastructure other"`
	ReporterID      *uint64                    `json:"reporter_id"`
	AssigneeID      *uint64                    `json:"assignee_id"`
	StartDate       *string                    `json:"start_date"`      // 格式: 2006-01-02 或 2006-01-02 15:04:05
	EndDate         *string                    `json:"end_date"`        // 格式: 2006-01-02 或 2006-01-02 15:04:05
	Progress        *string                    `json:"progress" binding:"omitempty,max=10000"`
	Result          *string                    `json:"result" binding:"omitempty,max=10000"`
	TargetProjectID *uint64                    `json:"target_project_id"`
	Tags            []string                   `json:"tags"`
}

// RequirementResponse 需求响应
type RequirementResponse struct {
	ID                   uint64                    `json:"id"`
	PoolID               uint64                    `json:"pool_id"`
	PoolName             string                    `json:"pool_name"`
	Title                string                    `json:"title"`
	Description          string                    `json:"description"`
	Priority             model.RequirementPriority `json:"priority"`
	Status               model.RequirementStatus   `json:"status"`
	Category             model.RequirementCategory `json:"category"`
	ReporterID           *uint64                   `json:"reporter_id,omitempty"`
	ReporterName         string                    `json:"reporter_name,omitempty"`
	AssigneeID           *uint64                   `json:"assignee_id,omitempty"`
	AssigneeName         string                    `json:"assignee_name,omitempty"`
	StartDate            *time.Time                `json:"start_date,omitempty"`
	EndDate              *time.Time                `json:"end_date,omitempty"`
	Progress             string                    `json:"progress,omitempty"`
	Result               string                    `json:"result,omitempty"`
	ConvertedIssueID     *uint64                   `json:"converted_issue_id,omitempty"`
	ConvertedIssueKey    string                    `json:"converted_issue_key,omitempty"`
	ConvertedIssueStatus string                    `json:"converted_issue_status,omitempty"`
	TargetProjectID      *uint64                   `json:"target_project_id,omitempty"`
	TargetProjectName    string                    `json:"target_project_name,omitempty"`
	CreatedBy            uint64                    `json:"created_by"`
	CreatorName          string                    `json:"creator_name"`
	Tags                 []string                  `json:"tags"`
	CommentCount         int64                     `json:"comment_count"`
	CreatedAt            time.Time                 `json:"created_at"`
	UpdatedAt            time.Time                 `json:"updated_at"`
}

// RequirementListRequest 需求列表请求
type RequirementListRequest struct {
	PoolID     *uint64                    `form:"pool_id"`
	Status     *model.RequirementStatus   `form:"status" binding:"omitempty,oneof=pending_review planning in_progress completed on_hold rejected"`
	Priority   *model.RequirementPriority `form:"priority" binding:"omitempty,oneof=P0 P1 P2 P3"`
	Category   *model.RequirementCategory `form:"category" binding:"omitempty,oneof=feature optimization bugfix security infrastructure other"`
	AssigneeID *uint64                    `form:"assignee_id"`
	CreatedBy  *uint64                    `form:"created_by"`
	StartDate  *time.Time                 `form:"start_date" time_format:"2006-01-02"`
	EndDate    *time.Time                 `form:"end_date" time_format:"2006-01-02"`
	Keyword    string                     `form:"keyword"`
	Page       int                        `form:"page" binding:"omitempty,min=1"`
	PageSize   int                        `form:"page_size" binding:"omitempty,min=1,max=100"`
}

// ConvertToIssueRequest 转化为工单请求
type ConvertToIssueRequest struct {
	ProjectKey  string  `json:"project_key" binding:"required"`
	IssueTypeID uint64  `json:"issue_type_id" binding:"required"`
	AssigneeID  *uint64 `json:"assignee_id"`
}

// ConvertToIssueResponse 转化为工单响应
type ConvertToIssueResponse struct {
	IssueID  uint64 `json:"issue_id"`
	IssueKey string `json:"issue_key"`
}

// RequirementCommentRequest 需求评论请求
type RequirementCommentRequest struct {
	Content string `json:"content" binding:"required,min=1,max=5000"`
}

// RequirementCommentResponse 需求评论响应
type RequirementCommentResponse struct {
	ID            uint64    `json:"id"`
	RequirementID uint64    `json:"requirement_id"`
	UserID        uint64    `json:"user_id"`
	UserName      string    `json:"user_name"`
	Content       string    `json:"content"`
	CreatedAt     time.Time `json:"created_at"`
}

// ========== 看板相关 DTO ==========

// KanbanRequest 看板请求
type KanbanRequest struct {
	PoolID    *uint64    `form:"pool_id"`
	GroupBy   string     `form:"group_by" binding:"required,oneof=status priority assignee timeline"`
	StartDate *time.Time `form:"start_date" time_format:"2006-01-02"`
	EndDate   *time.Time `form:"end_date" time_format:"2006-01-02"`
}

// KanbanColumn 看板列
type KanbanColumn struct {
	Key          string                 `json:"key"`
	Title        string                 `json:"title"`
	Count        int64                  `json:"count"`
	Requirements []*RequirementResponse `json:"requirements"`
}

// KanbanResponse 看板响应
type KanbanResponse struct {
	GroupBy string          `json:"group_by"`
	Columns []*KanbanColumn `json:"columns"`
	Total   int64           `json:"total"`
}

// ========== 统计报告相关 DTO ==========

// ReportRequest 报告请求
type ReportRequest struct {
	PoolID    *uint64    `form:"pool_id"`
	StartDate time.Time  `form:"start_date" binding:"required" time_format:"2006-01-02"`
	EndDate   time.Time  `form:"end_date" binding:"required" time_format:"2006-01-02"`
	GroupBy   string     `form:"group_by" binding:"omitempty,oneof=day week month"`
}

// StatusSummary 状态统计
type StatusSummary struct {
	Status model.RequirementStatus `json:"status"`
	Count  int64                   `json:"count"`
}

// PrioritySummary 优先级统计
type PrioritySummary struct {
	Priority model.RequirementPriority `json:"priority"`
	Count    int64                     `json:"count"`
}

// AssigneeSummary 负责人统计
type AssigneeSummary struct {
	AssigneeID   uint64 `json:"assignee_id"`
	AssigneeName string `json:"assignee_name"`
	Total        int64  `json:"total"`
	Completed    int64  `json:"completed"`
	InProgress   int64  `json:"in_progress"`
}

// TrendData 趋势数据
type TrendData struct {
	Date      string `json:"date"`
	Created   int64  `json:"created"`
	Completed int64  `json:"completed"`
}

// ReportResponse 报告响应
type ReportResponse struct {
	PoolID          *uint64            `json:"pool_id,omitempty"`
	PoolName        string             `json:"pool_name,omitempty"`
	StartDate       time.Time          `json:"start_date"`
	EndDate         time.Time          `json:"end_date"`
	TotalCreated    int64              `json:"total_created"`
	TotalCompleted  int64              `json:"total_completed"`
	TotalRejected   int64              `json:"total_rejected"`
	AvgProcessDays  float64            `json:"avg_process_days"`
	StatusSummary   []*StatusSummary   `json:"status_summary"`
	PrioritySummary []*PrioritySummary `json:"priority_summary"`
	AssigneeSummary []*AssigneeSummary `json:"assignee_summary"`
	TrendData       []*TrendData       `json:"trend_data"`
}

// ExportRequest 导出请求
type ExportRequest struct {
	PoolID    *uint64                    `form:"pool_id"`
	Status    *model.RequirementStatus   `form:"status"`
	Priority  *model.RequirementPriority `form:"priority"`
	StartDate *time.Time                 `form:"start_date" time_format:"2006-01-02"`
	EndDate   *time.Time                 `form:"end_date" time_format:"2006-01-02"`
	Format    string                     `form:"format" binding:"omitempty,oneof=csv xlsx"`
}
