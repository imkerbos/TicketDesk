// Package repository 提供字段方案数据访问层
package repository

import (
	"context"

	"github.com/kerbos/ticketdesk/internal/model"
	"gorm.io/gorm"
)

// SchemeRepository 字段方案仓储接口
type SchemeRepository interface {
	Create(ctx context.Context, scheme *model.IssueTypeFieldScheme) error
	Update(ctx context.Context, scheme *model.IssueTypeFieldScheme) error
	Delete(ctx context.Context, id uint64) error
	DeleteByProjectAndType(ctx context.Context, projectID, issueTypeID uint64) error
	GetByID(ctx context.Context, id uint64) (*model.IssueTypeFieldScheme, error)
	GetByProjectTypeAndField(ctx context.Context, projectID, issueTypeID, fieldID uint64) (*model.IssueTypeFieldScheme, error)
	ListByProjectAndType(ctx context.Context, projectID, issueTypeID uint64) ([]*model.IssueTypeFieldScheme, error)
	BatchCreate(ctx context.Context, schemes []*model.IssueTypeFieldScheme) error
	WithTx(tx *gorm.DB) SchemeRepository
}

// schemeRepository 字段方案仓储实现
type schemeRepository struct {
	db *gorm.DB
}

// NewSchemeRepository 创建字段方案仓储实例
func NewSchemeRepository(db *gorm.DB) SchemeRepository {
	return &schemeRepository{db: db}
}

// Create 创建字段方案
func (r *schemeRepository) Create(ctx context.Context, scheme *model.IssueTypeFieldScheme) error {
	return r.db.WithContext(ctx).Create(scheme).Error
}

// Update 更新字段方案
func (r *schemeRepository) Update(ctx context.Context, scheme *model.IssueTypeFieldScheme) error {
	return r.db.WithContext(ctx).Save(scheme).Error
}

// Delete 删除字段方案
func (r *schemeRepository) Delete(ctx context.Context, id uint64) error {
	return r.db.WithContext(ctx).Delete(&model.IssueTypeFieldScheme{}, id).Error
}

// DeleteByProjectAndType 硬删除项目工单类型的所有字段方案（替换场景，避免软删除与唯一键冲突）
func (r *schemeRepository) DeleteByProjectAndType(ctx context.Context, projectID, issueTypeID uint64) error {
	return r.db.WithContext(ctx).Unscoped().
		Where("project_id = ? AND issue_type_id = ?", projectID, issueTypeID).
		Delete(&model.IssueTypeFieldScheme{}).Error
}

// GetByID 根据ID获取字段方案
func (r *schemeRepository) GetByID(ctx context.Context, id uint64) (*model.IssueTypeFieldScheme, error) {
	var scheme model.IssueTypeFieldScheme
	err := r.db.WithContext(ctx).First(&scheme, id).Error
	if err != nil {
		return nil, err
	}
	return &scheme, nil
}

// GetByProjectTypeAndField 获取指定项目、类型、字段的方案
func (r *schemeRepository) GetByProjectTypeAndField(ctx context.Context, projectID, issueTypeID, fieldID uint64) (*model.IssueTypeFieldScheme, error) {
	var scheme model.IssueTypeFieldScheme
	err := r.db.WithContext(ctx).
		Where("project_id = ? AND issue_type_id = ? AND field_id = ?", projectID, issueTypeID, fieldID).
		First(&scheme).Error
	if err != nil {
		return nil, err
	}
	return &scheme, nil
}

// ListByProjectAndType 获取项目工单类型的字段方案列表
func (r *schemeRepository) ListByProjectAndType(ctx context.Context, projectID, issueTypeID uint64) ([]*model.IssueTypeFieldScheme, error) {
	var schemes []*model.IssueTypeFieldScheme
	err := r.db.WithContext(ctx).
		Where("project_id = ? AND issue_type_id = ?", projectID, issueTypeID).
		Order("sort_order ASC, id ASC").
		Find(&schemes).Error
	return schemes, err
}

// BatchCreate 批量创建字段方案
func (r *schemeRepository) BatchCreate(ctx context.Context, schemes []*model.IssueTypeFieldScheme) error {
	if len(schemes) == 0 {
		return nil
	}
	return r.db.WithContext(ctx).Create(&schemes).Error
}

// WithTx 返回使用指定事务的仓储实例
func (r *schemeRepository) WithTx(tx *gorm.DB) SchemeRepository {
	return &schemeRepository{db: tx}
}
