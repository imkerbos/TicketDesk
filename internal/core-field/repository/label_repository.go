// Package repository 提供标签数据访问层
package repository

import (
	"context"

	"gorm.io/gorm"

	"github.com/kerbos/ticketdesk/internal/model"
)

// LabelRepository 工单标签仓储接口
type LabelRepository interface {
	Create(ctx context.Context, label *model.IssueLabel) error
	Update(ctx context.Context, label *model.IssueLabel) error
	Delete(ctx context.Context, id uint64) error
	GetByID(ctx context.Context, id uint64) (*model.IssueLabel, error)
	GetByName(ctx context.Context, projectID uint64, name string) (*model.IssueLabel, error)
	ListByProject(ctx context.Context, projectID uint64) ([]*model.IssueLabel, error)
	ListByIDs(ctx context.Context, ids []uint64) ([]*model.IssueLabel, error)
}

// labelRepository 工单标签仓储实现
type labelRepository struct {
	db *gorm.DB
}

// NewLabelRepository 创建工单标签仓储实例
func NewLabelRepository(db *gorm.DB) LabelRepository {
	return &labelRepository{db: db}
}

// Create 创建标签
func (r *labelRepository) Create(ctx context.Context, label *model.IssueLabel) error {
	return r.db.WithContext(ctx).Create(label).Error
}

// Update 更新标签
func (r *labelRepository) Update(ctx context.Context, label *model.IssueLabel) error {
	return r.db.WithContext(ctx).Save(label).Error
}

// Delete 硬删除标签
func (r *labelRepository) Delete(ctx context.Context, id uint64) error {
	return r.db.WithContext(ctx).Unscoped().Delete(&model.IssueLabel{}, id).Error
}

// GetByID 根据ID获取标签
func (r *labelRepository) GetByID(ctx context.Context, id uint64) (*model.IssueLabel, error) {
	var label model.IssueLabel
	err := r.db.WithContext(ctx).First(&label, id).Error
	if err != nil {
		return nil, err
	}
	return &label, nil
}

// GetByName 根据项目ID和名称获取标签
func (r *labelRepository) GetByName(ctx context.Context, projectID uint64, name string) (*model.IssueLabel, error) {
	var label model.IssueLabel
	err := r.db.WithContext(ctx).
		Where("project_id = ? AND name = ?", projectID, name).
		First(&label).Error
	if err != nil {
		return nil, err
	}
	return &label, nil
}

// ListByProject 获取项目的所有标签
func (r *labelRepository) ListByProject(ctx context.Context, projectID uint64) ([]*model.IssueLabel, error) {
	var labels []*model.IssueLabel
	err := r.db.WithContext(ctx).
		Where("project_id = ?", projectID).
		Order("name ASC").
		Find(&labels).Error
	return labels, err
}

// ListByIDs 根据ID列表获取标签
func (r *labelRepository) ListByIDs(ctx context.Context, ids []uint64) ([]*model.IssueLabel, error) {
	if len(ids) == 0 {
		return []*model.IssueLabel{}, nil
	}
	var labels []*model.IssueLabel
	err := r.db.WithContext(ctx).
		Where("id IN ?", ids).
		Find(&labels).Error
	return labels, err
}
