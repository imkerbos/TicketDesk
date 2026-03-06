package model

import (
	"time"
)

// RequirementStatus 需求状态
type RequirementStatus string

const (
	RequirementStatusPendingReview RequirementStatus = "pending_review" // 待评估
	RequirementStatusPlanning      RequirementStatus = "planning"       // 规划中
	RequirementStatusInProgress    RequirementStatus = "in_progress"    // 进行中
	RequirementStatusCompleted     RequirementStatus = "completed"      // 已完成
	RequirementStatusOnHold        RequirementStatus = "on_hold"        // 已搁置
	RequirementStatusRejected      RequirementStatus = "rejected"       // 已拒绝
)

// RequirementCategory 需求分类
type RequirementCategory string

const (
	RequirementCategoryFeature        RequirementCategory = "feature"        // 功能需求
	RequirementCategoryOptimization   RequirementCategory = "optimization"   // 优化改进
	RequirementCategoryBugfix         RequirementCategory = "bugfix"         // 故障修复
	RequirementCategorySecurity       RequirementCategory = "security"       // 安全合规
	RequirementCategoryInfrastructure RequirementCategory = "infrastructure" // 基础设施
	RequirementCategoryOther          RequirementCategory = "other"          // 其他
)

// RequirementPriority 需求优先级
type RequirementPriority string

const (
	RequirementPriorityP0 RequirementPriority = "P0" // 紧急
	RequirementPriorityP1 RequirementPriority = "P1" // 高
	RequirementPriorityP2 RequirementPriority = "P2" // 中
	RequirementPriorityP3 RequirementPriority = "P3" // 低
)

// Requirement 需求
type Requirement struct {
	ID               uint64              `gorm:"primaryKey;autoIncrement" json:"id"`
	PoolID           uint64              `gorm:"not null;index" json:"pool_id"`                                          // 所属需求池
	Title            string              `gorm:"type:varchar(500);not null" json:"title"`                                // 标题
	Description      string              `gorm:"type:text" json:"description"`                                           // 描述
	Priority         RequirementPriority `gorm:"type:varchar(10);not null;index;default:'P2'" json:"priority"`           // 优先级
	Status           RequirementStatus   `gorm:"type:varchar(20);not null;index;default:'pending_review'" json:"status"` // 状态
	Category         RequirementCategory `gorm:"type:varchar(30);not null;index;default:'other'" json:"category"`        // 分类
	ReporterID       *uint64             `gorm:"index" json:"reporter_id,omitempty"`                                     // 需求来源（报告人）
	AssigneeID       *uint64             `gorm:"index" json:"assignee_id,omitempty"`                                     // 负责人ID
	StartDate        *time.Time          `gorm:"type:datetime" json:"start_date,omitempty"`                              // 开始时间
	EndDate          *time.Time          `gorm:"type:datetime;index" json:"end_date,omitempty"`                          // 结束时间
	Progress         string              `gorm:"type:text" json:"progress,omitempty"`                                    // 当前进度描述
	Result           string              `gorm:"type:text" json:"result,omitempty"`                                      // 结果描述
	ConvertedIssueID *uint64             `gorm:"index" json:"converted_issue_id,omitempty"`                              // 转化后的工单ID
	TargetProjectID  *uint64             `gorm:"index" json:"target_project_id,omitempty"`                               // 目标项目ID（转化时使用）
	CreatedBy        uint64              `gorm:"not null;index" json:"created_by"`                                       // 创建人
	Tags             string              `gorm:"type:varchar(500)" json:"tags,omitempty"`                                // 标签（逗号分隔）
	CreatedAt        time.Time           `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt        time.Time           `gorm:"autoUpdateTime" json:"updated_at"`

	// 关联
	Pool           *RequirementPool         `gorm:"foreignKey:PoolID" json:"pool,omitempty"`
	Reporter       *User                    `gorm:"foreignKey:ReporterID" json:"reporter,omitempty"`
	Assignee       *User                    `gorm:"foreignKey:AssigneeID" json:"assignee,omitempty"`
	Creator        *User                    `gorm:"foreignKey:CreatedBy" json:"creator,omitempty"`
	ConvertedIssue *Issue                   `gorm:"foreignKey:ConvertedIssueID" json:"converted_issue,omitempty"`
	TargetProject  *Project                 `gorm:"foreignKey:TargetProjectID" json:"target_project,omitempty"`
	Comments       []*RequirementComment    `gorm:"foreignKey:RequirementID" json:"comments,omitempty"`
	Attachments    []*RequirementAttachment `gorm:"foreignKey:RequirementID" json:"attachments,omitempty"`
}

// TableName 指定表名
func (Requirement) TableName() string {
	return "requirements"
}

// RequirementComment 需求评论
type RequirementComment struct {
	ID            uint64    `gorm:"primaryKey;autoIncrement" json:"id"`
	RequirementID uint64    `gorm:"not null;index" json:"requirement_id"`
	UserID        uint64    `gorm:"not null;index" json:"user_id"`
	Content       string    `gorm:"type:text;not null" json:"content"`
	CreatedAt     time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt     time.Time `gorm:"autoUpdateTime" json:"updated_at"`

	// 关联
	Requirement *Requirement `gorm:"foreignKey:RequirementID" json:"requirement,omitempty"`
	User        *User        `gorm:"foreignKey:UserID" json:"user,omitempty"`
}

// TableName 指定表名
func (RequirementComment) TableName() string {
	return "requirement_comments"
}

// RequirementAttachment 需求附件
type RequirementAttachment struct {
	ID            uint64    `gorm:"primaryKey;autoIncrement" json:"id"`
	RequirementID uint64    `gorm:"not null;index" json:"requirement_id"`
	FileName      string    `gorm:"type:varchar(255);not null" json:"file_name"`
	FilePath      string    `gorm:"type:varchar(500);not null" json:"file_path"`
	FileSize      int64     `gorm:"not null" json:"file_size"`
	FileType      string    `gorm:"type:varchar(100)" json:"file_type"`
	UploadedBy    uint64    `gorm:"not null;index" json:"uploaded_by"`
	CreatedAt     time.Time `gorm:"autoCreateTime" json:"created_at"`

	// 关联
	Requirement *Requirement `gorm:"foreignKey:RequirementID" json:"requirement,omitempty"`
	Uploader    *User        `gorm:"foreignKey:UploadedBy" json:"uploader,omitempty"`
}

// TableName 指定表名
func (RequirementAttachment) TableName() string {
	return "requirement_attachments"
}
