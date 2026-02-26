// Package repository 提供字段数据访问层
package repository

import (
	"context"

	"github.com/kerbos/ticketdesk/internal/model"
	"gorm.io/gorm"
)

// FieldRepository 字段定义仓储接口
type FieldRepository interface {
	Create(ctx context.Context, field *model.FieldDefinition) error
	Update(ctx context.Context, field *model.FieldDefinition) error
	Delete(ctx context.Context, id uint64) error
	GetByID(ctx context.Context, id uint64) (*model.FieldDefinition, error)
	GetByKey(ctx context.Context, projectID *uint64, fieldKey string) (*model.FieldDefinition, error)
	ListByProject(ctx context.Context, projectID uint64) ([]*model.FieldDefinition, error)
	ListSystemFields(ctx context.Context) ([]*model.FieldDefinition, error)
	ListAllAvailable(ctx context.Context, projectID uint64) ([]*model.FieldDefinition, error)
	ListByIDs(ctx context.Context, ids []uint64) ([]*model.FieldDefinition, error)
	ListGlobalFields(ctx context.Context) ([]*model.FieldDefinition, error)
}

// fieldRepository 字段定义仓储实现
type fieldRepository struct {
	db *gorm.DB
}

// NewFieldRepository 创建字段定义仓储实例
func NewFieldRepository(db *gorm.DB) FieldRepository {
	return &fieldRepository{db: db}
}

// Create 创建字段定义
func (r *fieldRepository) Create(ctx context.Context, field *model.FieldDefinition) error {
	return r.db.WithContext(ctx).Create(field).Error
}

// Update 更新字段定义
func (r *fieldRepository) Update(ctx context.Context, field *model.FieldDefinition) error {
	return r.db.WithContext(ctx).Save(field).Error
}

// Delete 删除字段定义
func (r *fieldRepository) Delete(ctx context.Context, id uint64) error {
	return r.db.WithContext(ctx).Delete(&model.FieldDefinition{}, id).Error
}

// GetByID 根据ID获取字段定义
func (r *fieldRepository) GetByID(ctx context.Context, id uint64) (*model.FieldDefinition, error) {
	var field model.FieldDefinition
	err := r.db.WithContext(ctx).First(&field, id).Error
	if err != nil {
		return nil, err
	}
	return &field, nil
}

// GetByKey 根据项目ID和字段Key获取字段定义
func (r *fieldRepository) GetByKey(ctx context.Context, projectID *uint64, fieldKey string) (*model.FieldDefinition, error) {
	var field model.FieldDefinition
	query := r.db.WithContext(ctx).Where("field_key = ?", fieldKey)
	if projectID == nil {
		query = query.Where("project_id IS NULL")
	} else {
		query = query.Where("project_id = ?", *projectID)
	}
	err := query.First(&field).Error
	if err != nil {
		return nil, err
	}
	return &field, nil
}

// ListByProject 获取项目自定义字段列表
func (r *fieldRepository) ListByProject(ctx context.Context, projectID uint64) ([]*model.FieldDefinition, error) {
	var fields []*model.FieldDefinition
	err := r.db.WithContext(ctx).
		Where("project_id = ?", projectID).
		Order("sort_order ASC, id ASC").
		Find(&fields).Error
	return fields, err
}

// ListSystemFields 获取系统字段列表
func (r *fieldRepository) ListSystemFields(ctx context.Context) ([]*model.FieldDefinition, error) {
	var fields []*model.FieldDefinition
	err := r.db.WithContext(ctx).
		Where("project_id IS NULL AND is_system = ?", true).
		Order("sort_order ASC, id ASC").
		Find(&fields).Error
	return fields, err
}

// ListAllAvailable 获取项目可用的所有字段（系统字段+项目自定义字段）
func (r *fieldRepository) ListAllAvailable(ctx context.Context, projectID uint64) ([]*model.FieldDefinition, error) {
	var fields []*model.FieldDefinition
	err := r.db.WithContext(ctx).
		Where("project_id IS NULL OR project_id = ?", projectID).
		Where("is_active = ?", true).
		Order("is_system DESC, sort_order ASC, id ASC").
		Find(&fields).Error
	return fields, err
}

// ListByIDs 根据ID列表批量获取字段定义
func (r *fieldRepository) ListByIDs(ctx context.Context, ids []uint64) ([]*model.FieldDefinition, error) {
	if len(ids) == 0 {
		return nil, nil
	}
	var fields []*model.FieldDefinition
	err := r.db.WithContext(ctx).Where("id IN ?", ids).Find(&fields).Error
	return fields, err
}

// ListGlobalFields 获取全局字段列表（系统字段+全局自定义字段）
func (r *fieldRepository) ListGlobalFields(ctx context.Context) ([]*model.FieldDefinition, error) {
	var fields []*model.FieldDefinition
	err := r.db.WithContext(ctx).
		Where("project_id IS NULL").
		Order("is_system DESC, sort_order ASC, id ASC").
		Find(&fields).Error
	return fields, err
}
