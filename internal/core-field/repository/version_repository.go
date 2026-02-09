// Package repository 提供版本数据访问层
package repository

import (
	"context"

	"github.com/kerbos/ticketdesk/internal/model"
	"gorm.io/gorm"
)

// VersionRepository 项目版本仓储接口
type VersionRepository interface {
	Create(ctx context.Context, version *model.ProjectVersion) error
	Update(ctx context.Context, version *model.ProjectVersion) error
	Delete(ctx context.Context, id uint64) error
	GetByID(ctx context.Context, id uint64) (*model.ProjectVersion, error)
	GetByName(ctx context.Context, projectID uint64, name string) (*model.ProjectVersion, error)
	ListByProject(ctx context.Context, projectID uint64) ([]*model.ProjectVersion, error)
	ListByStatus(ctx context.Context, projectID uint64, status string) ([]*model.ProjectVersion, error)
}

// versionRepository 项目版本仓储实现
type versionRepository struct {
	db *gorm.DB
}

// NewVersionRepository 创建项目版本仓储实例
func NewVersionRepository(db *gorm.DB) VersionRepository {
	return &versionRepository{db: db}
}

// Create 创建版本
func (r *versionRepository) Create(ctx context.Context, version *model.ProjectVersion) error {
	return r.db.WithContext(ctx).Create(version).Error
}

// Update 更新版本
func (r *versionRepository) Update(ctx context.Context, version *model.ProjectVersion) error {
	return r.db.WithContext(ctx).Save(version).Error
}

// Delete 删除版本
func (r *versionRepository) Delete(ctx context.Context, id uint64) error {
	return r.db.WithContext(ctx).Delete(&model.ProjectVersion{}, id).Error
}

// GetByID 根据ID获取版本
func (r *versionRepository) GetByID(ctx context.Context, id uint64) (*model.ProjectVersion, error) {
	var version model.ProjectVersion
	err := r.db.WithContext(ctx).First(&version, id).Error
	if err != nil {
		return nil, err
	}
	return &version, nil
}

// GetByName 根据项目ID和名称获取版本
func (r *versionRepository) GetByName(ctx context.Context, projectID uint64, name string) (*model.ProjectVersion, error) {
	var version model.ProjectVersion
	err := r.db.WithContext(ctx).
		Where("project_id = ? AND name = ?", projectID, name).
		First(&version).Error
	if err != nil {
		return nil, err
	}
	return &version, nil
}

// ListByProject 获取项目的所有版本
func (r *versionRepository) ListByProject(ctx context.Context, projectID uint64) ([]*model.ProjectVersion, error) {
	var versions []*model.ProjectVersion
	err := r.db.WithContext(ctx).
		Where("project_id = ?", projectID).
		Order("sort_order ASC, release_date DESC, id DESC").
		Find(&versions).Error
	return versions, err
}

// ListByStatus 获取项目指定状态的版本
func (r *versionRepository) ListByStatus(ctx context.Context, projectID uint64, status string) ([]*model.ProjectVersion, error) {
	var versions []*model.ProjectVersion
	err := r.db.WithContext(ctx).
		Where("project_id = ? AND status = ?", projectID, status).
		Order("sort_order ASC, release_date DESC, id DESC").
		Find(&versions).Error
	return versions, err
}
