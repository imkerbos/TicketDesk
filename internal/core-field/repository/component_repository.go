// Package repository 提供组件数据访问层
package repository

import (
	"context"

	"github.com/kerbos/ticketdesk/internal/model"
	"gorm.io/gorm"
)

// ComponentRepository 项目组件仓储接口
type ComponentRepository interface {
	Create(ctx context.Context, component *model.ProjectComponent) error
	Update(ctx context.Context, component *model.ProjectComponent) error
	Delete(ctx context.Context, id uint64) error
	GetByID(ctx context.Context, id uint64) (*model.ProjectComponent, error)
	GetByName(ctx context.Context, projectID uint64, name string) (*model.ProjectComponent, error)
	ListByProject(ctx context.Context, projectID uint64) ([]*model.ProjectComponent, error)
}

// componentRepository 项目组件仓储实现
type componentRepository struct {
	db *gorm.DB
}

// NewComponentRepository 创建项目组件仓储实例
func NewComponentRepository(db *gorm.DB) ComponentRepository {
	return &componentRepository{db: db}
}

// Create 创建组件
func (r *componentRepository) Create(ctx context.Context, component *model.ProjectComponent) error {
	return r.db.WithContext(ctx).Create(component).Error
}

// Update 更新组件
func (r *componentRepository) Update(ctx context.Context, component *model.ProjectComponent) error {
	return r.db.WithContext(ctx).Save(component).Error
}

// Delete 删除组件
func (r *componentRepository) Delete(ctx context.Context, id uint64) error {
	return r.db.WithContext(ctx).Delete(&model.ProjectComponent{}, id).Error
}

// GetByID 根据ID获取组件
func (r *componentRepository) GetByID(ctx context.Context, id uint64) (*model.ProjectComponent, error) {
	var component model.ProjectComponent
	err := r.db.WithContext(ctx).First(&component, id).Error
	if err != nil {
		return nil, err
	}
	return &component, nil
}

// GetByName 根据项目ID和名称获取组件
func (r *componentRepository) GetByName(ctx context.Context, projectID uint64, name string) (*model.ProjectComponent, error) {
	var component model.ProjectComponent
	err := r.db.WithContext(ctx).
		Where("project_id = ? AND name = ?", projectID, name).
		First(&component).Error
	if err != nil {
		return nil, err
	}
	return &component, nil
}

// ListByProject 获取项目的所有组件
func (r *componentRepository) ListByProject(ctx context.Context, projectID uint64) ([]*model.ProjectComponent, error) {
	var components []*model.ProjectComponent
	err := r.db.WithContext(ctx).
		Where("project_id = ?", projectID).
		Order("name ASC").
		Find(&components).Error
	return components, err
}
