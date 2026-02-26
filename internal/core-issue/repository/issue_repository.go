// Package repository 提供工单数据访问层
package repository

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/kerbos/ticketdesk/internal/model"
	"gorm.io/gorm"
)

// IssueRepository 工单数据访问接口
// ListResult 分页查询结果
type ListResult struct {
	Issues  []*model.Issue
	Total   int64
	HasMore bool // true 表示实际总数超过 maxCount，Total 为封顶值
}

type IssueRepository interface {
	Create(ctx context.Context, issue *model.Issue) error
	GetByID(ctx context.Context, id uint64) (*model.Issue, error)
	GetByKey(ctx context.Context, key string) (*model.Issue, error)
	Update(ctx context.Context, issue *model.Issue) error
	Delete(ctx context.Context, id uint64) error
	List(ctx context.Context, filter *IssueFilter, offset, limit int) (*ListResult, error)
	GetNextIssueNumber(ctx context.Context, projectID uint64) (int64, error)
	ListByIDs(ctx context.Context, ids []uint64) ([]*model.Issue, error)
	ListByParentID(ctx context.Context, parentID uint64) ([]*model.Issue, error)
	ListByEpicID(ctx context.Context, epicID uint64) ([]*model.Issue, error)
	ListByMergedIntoIssueID(ctx context.Context, issueID uint64, result *[]model.Issue) error
}

// IssueFilter 工单过滤条件
type IssueFilter struct {
	ProjectID        *uint64
	ProjectIDs       []uint64 // 限定项目范围（非管理员用户可访问的项目列表）
	LimitByProjects  bool     // 是否启用 ProjectIDs 过滤（区分 nil 和空切片）
	Status           string
	StatusNotIn      []string // 排除的状态列表
	Priority         string
	AssigneeID       *uint64
	ReporterID       *uint64
	IssueTypeID      *uint64
	EpicID           *uint64
	Keyword          string
	Category         string // "normal" = 排除告警工单, "alert" = 仅告警工单
}

// issueRepository 工单数据访问实现
type issueRepository struct {
	db *gorm.DB
}

// NewIssueRepository 创建工单数据访问实例
func NewIssueRepository(db *gorm.DB) IssueRepository {
	return &issueRepository{db: db}
}

// Create 创建工单
func (r *issueRepository) Create(ctx context.Context, issue *model.Issue) error {
	return r.db.WithContext(ctx).Create(issue).Error
}

// GetByID 根据 ID 获取工单
func (r *issueRepository) GetByID(ctx context.Context, id uint64) (*model.Issue, error) {
	var issue model.Issue
	err := r.db.WithContext(ctx).First(&issue, id).Error
	if err != nil {
		return nil, err
	}
	return &issue, nil
}

// GetByKey 根据工单 Key 获取工单
func (r *issueRepository) GetByKey(ctx context.Context, key string) (*model.Issue, error) {
	var issue model.Issue
	err := r.db.WithContext(ctx).Where("issue_key = ?", key).First(&issue).Error
	if err != nil {
		return nil, err
	}
	return &issue, nil
}

// Update 更新工单
func (r *issueRepository) Update(ctx context.Context, issue *model.Issue) error {
	return r.db.WithContext(ctx).Save(issue).Error
}

// Delete 软删除工单
func (r *issueRepository) Delete(ctx context.Context, id uint64) error {
	return r.db.WithContext(ctx).Delete(&model.Issue{}, id).Error
}

// maxCountLimit 分页计数上限，超过此值显示 "10,000+"
// 避免 COUNT(*) 全表扫描，保证任意数据量下计数查询 <2ms
const maxCountLimit = 10001

// List 分页查询工单列表
func (r *issueRepository) List(ctx context.Context, filter *IssueFilter, offset, limit int) (*ListResult, error) {
	var issues []*model.Issue
	result := &ListResult{}

	query := r.db.WithContext(ctx).Model(&model.Issue{})

	// 应用过滤条件
	if filter != nil {
		if filter.ProjectID != nil {
			query = query.Where("project_id = ?", *filter.ProjectID)
		}
		if filter.LimitByProjects {
			if len(filter.ProjectIDs) == 0 {
				// 用户没有任何项目权限，直接返回空结果
				return result, nil
			}
			query = query.Where("project_id IN ?", filter.ProjectIDs)
		}
		if filter.Status != "" {
			query = query.Where("status = ?", filter.Status)
		}
		if len(filter.StatusNotIn) > 0 {
			query = query.Where("status NOT IN ?", filter.StatusNotIn)
		}
		if filter.Priority != "" {
			query = query.Where("priority = ?", filter.Priority)
		}
		if filter.AssigneeID != nil {
			query = query.Where("assignee_id = ?", *filter.AssigneeID)
		}
		if filter.ReporterID != nil {
			query = query.Where("reporter_id = ?", *filter.ReporterID)
		}
		if filter.IssueTypeID != nil {
			query = query.Where("issue_type_id = ?", *filter.IssueTypeID)
		}
		if filter.EpicID != nil {
			query = query.Where("epic_id = ?", *filter.EpicID)
		}
		if filter.Keyword != "" {
			keyword := "%" + filter.Keyword + "%"
			query = query.Where("issue_key LIKE ? OR title LIKE ?", keyword, keyword)
		}
		if filter.Category == "alert" {
			query = query.Where("issue_type_id IN (SELECT id FROM issue_types WHERE name = 'Alert' AND deleted_at IS NULL)")
		} else if filter.Category == "normal" {
			query = query.Where("issue_type_id NOT IN (SELECT id FROM issue_types WHERE name = 'Alert' AND deleted_at IS NULL)")
		}
	}

	// 封顶计数：最多扫描 maxCountLimit 行，避免百万级 COUNT 全扫描
	var cappedCount int64
	countSubQuery := query.Session(&gorm.Session{}).Select("1").Limit(maxCountLimit)
	if err := r.db.WithContext(ctx).Table("(?) AS capped", countSubQuery).Count(&cappedCount).Error; err != nil {
		return nil, err
	}

	if cappedCount >= int64(maxCountLimit) {
		result.Total = int64(maxCountLimit - 1) // 10000
		result.HasMore = true
	} else {
		result.Total = cappedCount
	}

	// 分页查询
	if err := query.Offset(offset).Limit(limit).Order("id DESC").Find(&issues).Error; err != nil {
		return nil, err
	}
	result.Issues = issues

	return result, nil
}

// GetNextIssueNumber 获取项目下一个工单编号（包含软删除记录，避免唯一键冲突）
func (r *issueRepository) GetNextIssueNumber(ctx context.Context, projectID uint64) (int64, error) {
	var maxNum sql.NullInt64
	// 从 issue_key 中提取最大序号，格式为 "PROJECT-123"
	err := r.db.WithContext(ctx).Unscoped().Model(&model.Issue{}).
		Where("project_id = ?", projectID).
		Select("MAX(CAST(SUBSTRING_INDEX(issue_key, '-', -1) AS UNSIGNED))").
		Scan(&maxNum).Error
	if err != nil {
		return 0, err
	}
	if !maxNum.Valid {
		return 1, nil
	}
	return maxNum.Int64 + 1, nil
}

// ListByIDs 根据 ID 列表批量查询工单
func (r *issueRepository) ListByIDs(ctx context.Context, ids []uint64) ([]*model.Issue, error) {
	var issues []*model.Issue
	err := r.db.WithContext(ctx).Where("id IN ?", ids).Find(&issues).Error
	return issues, err
}

// ListByParentID 根据父工单 ID 获取所有子任务
func (r *issueRepository) ListByParentID(ctx context.Context, parentID uint64) ([]*model.Issue, error) {
	var issues []*model.Issue
	err := r.db.WithContext(ctx).Where("parent_id = ?", parentID).Order("created_at ASC").Find(&issues).Error
	return issues, err
}

// ListByEpicID 根据 Epic ID 获取所有关联工单
func (r *issueRepository) ListByEpicID(ctx context.Context, epicID uint64) ([]*model.Issue, error) {
	var issues []*model.Issue
	err := r.db.WithContext(ctx).Where("epic_id = ?", epicID).Order("created_at ASC").Find(&issues).Error
	return issues, err
}

// ListByMergedIntoIssueID 查找所有合并到指定工单的旧工单
func (r *issueRepository) ListByMergedIntoIssueID(ctx context.Context, issueID uint64, result *[]model.Issue) error {
	return r.db.WithContext(ctx).
		Where("merged_into_issue_id = ?", issueID).
		Order("created_at ASC").
		Find(result).Error
}

// CommentRepository 评论数据访问接口
type CommentRepository interface {
	Create(ctx context.Context, comment *model.IssueComment) error
	GetByID(ctx context.Context, id uint64) (*model.IssueComment, error)
	Update(ctx context.Context, comment *model.IssueComment) error
	Delete(ctx context.Context, id uint64) error
	ListByIssue(ctx context.Context, issueID uint64) ([]*model.IssueComment, error)
}

// commentRepository 评论数据访问实现
type commentRepository struct {
	db *gorm.DB
}

// NewCommentRepository 创建评论数据访问实例
func NewCommentRepository(db *gorm.DB) CommentRepository {
	return &commentRepository{db: db}
}

// Create 创建评论
func (r *commentRepository) Create(ctx context.Context, comment *model.IssueComment) error {
	return r.db.WithContext(ctx).Create(comment).Error
}

// GetByID 根据 ID 获取评论
func (r *commentRepository) GetByID(ctx context.Context, id uint64) (*model.IssueComment, error) {
	var comment model.IssueComment
	err := r.db.WithContext(ctx).First(&comment, id).Error
	if err != nil {
		return nil, err
	}
	return &comment, nil
}

// Update 更新评论
func (r *commentRepository) Update(ctx context.Context, comment *model.IssueComment) error {
	return r.db.WithContext(ctx).Save(comment).Error
}

// Delete 删除评论
func (r *commentRepository) Delete(ctx context.Context, id uint64) error {
	return r.db.WithContext(ctx).Delete(&model.IssueComment{}, id).Error
}

// ListByIssue 获取工单的所有评论
func (r *commentRepository) ListByIssue(ctx context.Context, issueID uint64) ([]*model.IssueComment, error) {
	var comments []*model.IssueComment
	err := r.db.WithContext(ctx).Where("issue_id = ?", issueID).Order("id ASC").Find(&comments).Error
	return comments, err
}

// WatcherRepository 关注人数据访问接口
type WatcherRepository interface {
	Create(ctx context.Context, watcher *model.IssueWatcher) error
	Delete(ctx context.Context, id uint64) error
	DeleteByIssueAndUser(ctx context.Context, issueID, userID uint64) error
	ListByIssue(ctx context.Context, issueID uint64) ([]*model.IssueWatcher, error)
	IsWatching(ctx context.Context, issueID, userID uint64) (bool, error)
	GetByIssueAndUser(ctx context.Context, issueID, userID uint64) (*model.IssueWatcher, error)
}

// watcherRepository 关注人数据访问实现
type watcherRepository struct {
	db *gorm.DB
}

// NewWatcherRepository 创建关注人数据访问实例
func NewWatcherRepository(db *gorm.DB) WatcherRepository {
	return &watcherRepository{db: db}
}

// Create 创建关注人
func (r *watcherRepository) Create(ctx context.Context, watcher *model.IssueWatcher) error {
	return r.db.WithContext(ctx).Create(watcher).Error
}

// Delete 删除关注人
func (r *watcherRepository) Delete(ctx context.Context, id uint64) error {
	return r.db.WithContext(ctx).Delete(&model.IssueWatcher{}, id).Error
}

// DeleteByIssueAndUser 根据工单和用户删除关注
func (r *watcherRepository) DeleteByIssueAndUser(ctx context.Context, issueID, userID uint64) error {
	return r.db.WithContext(ctx).Where("issue_id = ? AND user_id = ?", issueID, userID).Delete(&model.IssueWatcher{}).Error
}

// ListByIssue 获取工单的所有关注人
func (r *watcherRepository) ListByIssue(ctx context.Context, issueID uint64) ([]*model.IssueWatcher, error) {
	var watchers []*model.IssueWatcher
	err := r.db.WithContext(ctx).Where("issue_id = ?", issueID).Order("id ASC").Find(&watchers).Error
	return watchers, err
}

// IsWatching 检查用户是否关注工单
func (r *watcherRepository) IsWatching(ctx context.Context, issueID, userID uint64) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&model.IssueWatcher{}).
		Where("issue_id = ? AND user_id = ?", issueID, userID).Count(&count).Error
	return count > 0, err
}

// GetByIssueAndUser 根据工单和用户获取关注记录
func (r *watcherRepository) GetByIssueAndUser(ctx context.Context, issueID, userID uint64) (*model.IssueWatcher, error) {
	var watcher model.IssueWatcher
	err := r.db.WithContext(ctx).Where("issue_id = ? AND user_id = ?", issueID, userID).First(&watcher).Error
	if err != nil {
		return nil, err
	}
	return &watcher, nil
}

// WorklogRepository 工作日志数据访问接口
type WorklogRepository interface {
	Create(ctx context.Context, worklog *model.IssueWorklog) error
	GetByID(ctx context.Context, id uint64) (*model.IssueWorklog, error)
	Update(ctx context.Context, worklog *model.IssueWorklog) error
	Delete(ctx context.Context, id uint64) error
	ListByIssue(ctx context.Context, issueID uint64) ([]*model.IssueWorklog, error)
	GetTotalTimeSpent(ctx context.Context, issueID uint64) (int, error)
}

// worklogRepository 工作日志数据访问实现
type worklogRepository struct {
	db *gorm.DB
}

// NewWorklogRepository 创建工作日志数据访问实例
func NewWorklogRepository(db *gorm.DB) WorklogRepository {
	return &worklogRepository{db: db}
}

// Create 创建工作日志
func (r *worklogRepository) Create(ctx context.Context, worklog *model.IssueWorklog) error {
	return r.db.WithContext(ctx).Create(worklog).Error
}

// GetByID 根据 ID 获取工作日志
func (r *worklogRepository) GetByID(ctx context.Context, id uint64) (*model.IssueWorklog, error) {
	var worklog model.IssueWorklog
	err := r.db.WithContext(ctx).First(&worklog, id).Error
	if err != nil {
		return nil, err
	}
	return &worklog, nil
}

// Update 更新工作日志
func (r *worklogRepository) Update(ctx context.Context, worklog *model.IssueWorklog) error {
	return r.db.WithContext(ctx).Save(worklog).Error
}

// Delete 删除工作日志
func (r *worklogRepository) Delete(ctx context.Context, id uint64) error {
	return r.db.WithContext(ctx).Delete(&model.IssueWorklog{}, id).Error
}

// ListByIssue 获取工单的所有工作日志
func (r *worklogRepository) ListByIssue(ctx context.Context, issueID uint64) ([]*model.IssueWorklog, error) {
	var worklogs []*model.IssueWorklog
	err := r.db.WithContext(ctx).
		Where("issue_id = ?", issueID).
		Order("worked_at DESC, id DESC").
		Find(&worklogs).Error
	return worklogs, err
}

// GetTotalTimeSpent 获取工单的总工作时长（秒）
func (r *worklogRepository) GetTotalTimeSpent(ctx context.Context, issueID uint64) (int, error) {
	var total int
	err := r.db.WithContext(ctx).
		Model(&model.IssueWorklog{}).
		Where("issue_id = ?", issueID).
		Select("COALESCE(SUM(time_spent_sec), 0)").
		Scan(&total).Error
	return total, err
}

// GenerateIssueKey 生成工单 Key
func GenerateIssueKey(projectKey string, number int64) string {
	return fmt.Sprintf("%s-%d", projectKey, number)
}
