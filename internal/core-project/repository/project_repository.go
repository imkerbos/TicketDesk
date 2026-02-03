// Package repository 提供项目数据访问层
package repository

import (
	"context"

	"github.com/kerbos/ticketdesk/internal/model"
	"gorm.io/gorm"
)

// ProjectRepository 项目数据访问接口
type ProjectRepository interface {
	Create(ctx context.Context, project *model.Project) error
	GetByID(ctx context.Context, id uint64) (*model.Project, error)
	GetByKey(ctx context.Context, key string) (*model.Project, error)
	Update(ctx context.Context, project *model.Project) error
	Delete(ctx context.Context, id uint64) error
	List(ctx context.Context, offset, limit int, keyword string, status *int8) ([]*model.Project, int64, error)
	ExistsByKey(ctx context.Context, key string) (bool, error)
	GetNextIssueNumber(ctx context.Context, projectID uint64) (int64, error)
}

// projectRepository 项目数据访问实现
type projectRepository struct {
	db *gorm.DB
}

// NewProjectRepository 创建项目数据访问实例
func NewProjectRepository(db *gorm.DB) ProjectRepository {
	return &projectRepository{db: db}
}

// Create 创建项目
func (r *projectRepository) Create(ctx context.Context, project *model.Project) error {
	return r.db.WithContext(ctx).Create(project).Error
}

// GetByID 根据 ID 获取项目
func (r *projectRepository) GetByID(ctx context.Context, id uint64) (*model.Project, error) {
	var project model.Project
	err := r.db.WithContext(ctx).First(&project, id).Error
	if err != nil {
		return nil, err
	}
	return &project, nil
}

// GetByKey 根据项目 Key 获取项目
func (r *projectRepository) GetByKey(ctx context.Context, key string) (*model.Project, error) {
	var project model.Project
	err := r.db.WithContext(ctx).Where("project_key = ?", key).First(&project).Error
	if err != nil {
		return nil, err
	}
	return &project, nil
}

// Update 更新项目
func (r *projectRepository) Update(ctx context.Context, project *model.Project) error {
	return r.db.WithContext(ctx).Save(project).Error
}

// Delete 软删除项目
func (r *projectRepository) Delete(ctx context.Context, id uint64) error {
	return r.db.WithContext(ctx).Delete(&model.Project{}, id).Error
}

// List 分页查询项目列表
func (r *projectRepository) List(ctx context.Context, offset, limit int, keyword string, status *int8) ([]*model.Project, int64, error) {
	var projects []*model.Project
	var total int64

	query := r.db.WithContext(ctx).Model(&model.Project{})

	// 关键字搜索
	if keyword != "" {
		query = query.Where("project_key LIKE ? OR name LIKE ? OR description LIKE ?",
			"%"+keyword+"%", "%"+keyword+"%", "%"+keyword+"%")
	}

	// 状态过滤
	if status != nil {
		query = query.Where("status = ?", *status)
	}

	// 统计总数
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 分页查询
	if err := query.Offset(offset).Limit(limit).Order("id DESC").Find(&projects).Error; err != nil {
		return nil, 0, err
	}

	return projects, total, nil
}

// ExistsByKey 检查项目 Key 是否存在
func (r *projectRepository) ExistsByKey(ctx context.Context, key string) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&model.Project{}).Where("project_key = ?", key).Count(&count).Error
	return count > 0, err
}

// GetNextIssueNumber 获取项目下一个工单编号
func (r *projectRepository) GetNextIssueNumber(ctx context.Context, projectID uint64) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&model.Issue{}).Where("project_id = ?", projectID).Count(&count).Error
	return count + 1, err
}

// ProjectMemberRepository 项目成员数据访问接口
type ProjectMemberRepository interface {
	Create(ctx context.Context, member *model.ProjectMember) error
	GetByID(ctx context.Context, id uint64) (*model.ProjectMember, error)
	GetByProjectAndUser(ctx context.Context, projectID, userID uint64) (*model.ProjectMember, error)
	Update(ctx context.Context, member *model.ProjectMember) error
	Delete(ctx context.Context, id uint64) error
	DeleteByProjectAndUser(ctx context.Context, projectID, userID uint64) error
	ListByProject(ctx context.Context, projectID uint64) ([]*model.ProjectMember, error)
	ListByUser(ctx context.Context, userID uint64) ([]*model.ProjectMember, error)
	CountByProject(ctx context.Context, projectID uint64) (int64, error)
	IsMember(ctx context.Context, projectID, userID uint64) (bool, error)
	GetUserRole(ctx context.Context, projectID, userID uint64) (string, error)
}

// projectMemberRepository 项目成员数据访问实现
type projectMemberRepository struct {
	db *gorm.DB
}

// NewProjectMemberRepository 创建项目成员数据访问实例
func NewProjectMemberRepository(db *gorm.DB) ProjectMemberRepository {
	return &projectMemberRepository{db: db}
}

// Create 创建项目成员
func (r *projectMemberRepository) Create(ctx context.Context, member *model.ProjectMember) error {
	return r.db.WithContext(ctx).Create(member).Error
}

// GetByID 根据 ID 获取项目成员
func (r *projectMemberRepository) GetByID(ctx context.Context, id uint64) (*model.ProjectMember, error) {
	var member model.ProjectMember
	err := r.db.WithContext(ctx).First(&member, id).Error
	if err != nil {
		return nil, err
	}
	return &member, nil
}

// GetByProjectAndUser 根据项目和用户获取成员
func (r *projectMemberRepository) GetByProjectAndUser(ctx context.Context, projectID, userID uint64) (*model.ProjectMember, error) {
	var member model.ProjectMember
	err := r.db.WithContext(ctx).Where("project_id = ? AND user_id = ?", projectID, userID).First(&member).Error
	if err != nil {
		return nil, err
	}
	return &member, nil
}

// Update 更新项目成员
func (r *projectMemberRepository) Update(ctx context.Context, member *model.ProjectMember) error {
	return r.db.WithContext(ctx).Save(member).Error
}

// Delete 删除项目成员
func (r *projectMemberRepository) Delete(ctx context.Context, id uint64) error {
	return r.db.WithContext(ctx).Delete(&model.ProjectMember{}, id).Error
}

// DeleteByProjectAndUser 根据项目和用户删除成员
func (r *projectMemberRepository) DeleteByProjectAndUser(ctx context.Context, projectID, userID uint64) error {
	return r.db.WithContext(ctx).Where("project_id = ? AND user_id = ?", projectID, userID).Delete(&model.ProjectMember{}).Error
}

// ListByProject 获取项目的所有成员
func (r *projectMemberRepository) ListByProject(ctx context.Context, projectID uint64) ([]*model.ProjectMember, error) {
	var members []*model.ProjectMember
	err := r.db.WithContext(ctx).Where("project_id = ?", projectID).Order("role ASC, id ASC").Find(&members).Error
	return members, err
}

// ListByUser 获取用户参与的所有项目成员关系
func (r *projectMemberRepository) ListByUser(ctx context.Context, userID uint64) ([]*model.ProjectMember, error) {
	var members []*model.ProjectMember
	err := r.db.WithContext(ctx).Where("user_id = ?", userID).Find(&members).Error
	return members, err
}

// CountByProject 统计项目成员数量
func (r *projectMemberRepository) CountByProject(ctx context.Context, projectID uint64) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&model.ProjectMember{}).Where("project_id = ?", projectID).Count(&count).Error
	return count, err
}

// IsMember 检查用户是否是项目成员
func (r *projectMemberRepository) IsMember(ctx context.Context, projectID, userID uint64) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&model.ProjectMember{}).
		Where("project_id = ? AND user_id = ?", projectID, userID).Count(&count).Error
	return count > 0, err
}

// GetUserRole 获取用户在项目中的角色
func (r *projectMemberRepository) GetUserRole(ctx context.Context, projectID, userID uint64) (string, error) {
	var member model.ProjectMember
	err := r.db.WithContext(ctx).Where("project_id = ? AND user_id = ?", projectID, userID).First(&member).Error
	if err != nil {
		return "", err
	}
	return member.Role, nil
}

// IssueTypeRepository 工单类型数据访问接口
type IssueTypeRepository interface {
	Create(ctx context.Context, issueType *model.IssueType) error
	GetByID(ctx context.Context, id uint64) (*model.IssueType, error)
	Update(ctx context.Context, issueType *model.IssueType) error
	Delete(ctx context.Context, id uint64) error
	ListByProject(ctx context.Context, projectID *uint64) ([]*model.IssueType, error)
	ListGlobal(ctx context.Context) ([]*model.IssueType, error)
	ListAvailableForProject(ctx context.Context, projectID uint64) ([]*model.IssueType, error)
}

// issueTypeRepository 工单类型数据访问实现
type issueTypeRepository struct {
	db *gorm.DB
}

// NewIssueTypeRepository 创建工单类型数据访问实例
func NewIssueTypeRepository(db *gorm.DB) IssueTypeRepository {
	return &issueTypeRepository{db: db}
}

// Create 创建工单类型
func (r *issueTypeRepository) Create(ctx context.Context, issueType *model.IssueType) error {
	return r.db.WithContext(ctx).Create(issueType).Error
}

// GetByID 根据 ID 获取工单类型
func (r *issueTypeRepository) GetByID(ctx context.Context, id uint64) (*model.IssueType, error) {
	var issueType model.IssueType
	err := r.db.WithContext(ctx).First(&issueType, id).Error
	if err != nil {
		return nil, err
	}
	return &issueType, nil
}

// Update 更新工单类型
func (r *issueTypeRepository) Update(ctx context.Context, issueType *model.IssueType) error {
	return r.db.WithContext(ctx).Save(issueType).Error
}

// Delete 删除工单类型
func (r *issueTypeRepository) Delete(ctx context.Context, id uint64) error {
	return r.db.WithContext(ctx).Delete(&model.IssueType{}, id).Error
}

// ListByProject 获取项目的工单类型
func (r *issueTypeRepository) ListByProject(ctx context.Context, projectID *uint64) ([]*model.IssueType, error) {
	var issueTypes []*model.IssueType
	query := r.db.WithContext(ctx)
	if projectID != nil {
		query = query.Where("project_id = ?", *projectID)
	} else {
		query = query.Where("project_id IS NULL")
	}
	err := query.Order("id ASC").Find(&issueTypes).Error
	return issueTypes, err
}

// ListGlobal 获取全局工单类型
func (r *issueTypeRepository) ListGlobal(ctx context.Context) ([]*model.IssueType, error) {
	return r.ListByProject(ctx, nil)
}

// ListAvailableForProject 获取项目可用的工单类型（全局 + 项目自定义）
func (r *issueTypeRepository) ListAvailableForProject(ctx context.Context, projectID uint64) ([]*model.IssueType, error) {
	var issueTypes []*model.IssueType
	err := r.db.WithContext(ctx).
		Where("project_id IS NULL OR project_id = ?", projectID).
		Order("project_id ASC, id ASC").
		Find(&issueTypes).Error
	return issueTypes, err
}
