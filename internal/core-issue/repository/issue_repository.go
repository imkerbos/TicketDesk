// Package repository 提供工单数据访问层
package repository

import (
	"context"
	"fmt"

	"github.com/kerbos/ticketdesk/internal/model"
	"gorm.io/gorm"
)

// IssueRepository 工单数据访问接口
type IssueRepository interface {
	Create(ctx context.Context, issue *model.Issue) error
	GetByID(ctx context.Context, id uint64) (*model.Issue, error)
	GetByKey(ctx context.Context, key string) (*model.Issue, error)
	Update(ctx context.Context, issue *model.Issue) error
	Delete(ctx context.Context, id uint64) error
	List(ctx context.Context, filter *IssueFilter, offset, limit int) ([]*model.Issue, int64, error)
	GetNextIssueNumber(ctx context.Context, projectID uint64) (int64, error)
}

// IssueFilter 工单过滤条件
type IssueFilter struct {
	ProjectID   *uint64
	Status      string
	StatusNotIn []string // 排除的状态列表
	Priority    string
	AssigneeID  *uint64
	ReporterID  *uint64
	IssueTypeID *uint64
	Keyword     string
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

// List 分页查询工单列表
func (r *issueRepository) List(ctx context.Context, filter *IssueFilter, offset, limit int) ([]*model.Issue, int64, error) {
	var issues []*model.Issue
	var total int64

	query := r.db.WithContext(ctx).Model(&model.Issue{})

	// 应用过滤条件
	if filter != nil {
		if filter.ProjectID != nil {
			query = query.Where("project_id = ?", *filter.ProjectID)
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
		if filter.Keyword != "" {
			keyword := "%" + filter.Keyword + "%"
			query = query.Where("issue_key LIKE ? OR title LIKE ? OR description LIKE ?", keyword, keyword, keyword)
		}
	}

	// 统计总数
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 分页查询
	if err := query.Offset(offset).Limit(limit).Order("id DESC").Find(&issues).Error; err != nil {
		return nil, 0, err
	}

	return issues, total, nil
}

// GetNextIssueNumber 获取项目下一个工单编号
func (r *issueRepository) GetNextIssueNumber(ctx context.Context, projectID uint64) (int64, error) {
	var maxNum int64
	err := r.db.WithContext(ctx).Model(&model.Issue{}).
		Where("project_id = ?", projectID).
		Count(&maxNum).Error
	return maxNum + 1, err
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
