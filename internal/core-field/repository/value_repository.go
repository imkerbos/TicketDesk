// Package repository 提供字段值数据访问层
package repository

import (
	"context"

	"github.com/kerbos/ticketdesk/internal/model"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// ValueRepository 字段值仓储接口
type ValueRepository interface {
	Create(ctx context.Context, value *model.IssueFieldValue) error
	Update(ctx context.Context, value *model.IssueFieldValue) error
	Delete(ctx context.Context, id uint64) error
	DeleteByIssue(ctx context.Context, issueID uint64) error
	GetByIssueAndField(ctx context.Context, issueID, fieldID uint64) (*model.IssueFieldValue, error)
	ListByIssue(ctx context.Context, issueID uint64) ([]*model.IssueFieldValue, error)
	BatchUpsert(ctx context.Context, values []*model.IssueFieldValue) error
}

// valueRepository 字段值仓储实现
type valueRepository struct {
	db *gorm.DB
}

// NewValueRepository 创建字段值仓储实例
func NewValueRepository(db *gorm.DB) ValueRepository {
	return &valueRepository{db: db}
}

// Create 创建字段值
func (r *valueRepository) Create(ctx context.Context, value *model.IssueFieldValue) error {
	return r.db.WithContext(ctx).Create(value).Error
}

// Update 更新字段值
func (r *valueRepository) Update(ctx context.Context, value *model.IssueFieldValue) error {
	return r.db.WithContext(ctx).Save(value).Error
}

// Delete 删除字段值
func (r *valueRepository) Delete(ctx context.Context, id uint64) error {
	return r.db.WithContext(ctx).Delete(&model.IssueFieldValue{}, id).Error
}

// DeleteByIssue 删除工单的所有字段值
func (r *valueRepository) DeleteByIssue(ctx context.Context, issueID uint64) error {
	return r.db.WithContext(ctx).
		Where("issue_id = ?", issueID).
		Delete(&model.IssueFieldValue{}).Error
}

// GetByIssueAndField 获取工单指定字段的值
func (r *valueRepository) GetByIssueAndField(ctx context.Context, issueID, fieldID uint64) (*model.IssueFieldValue, error) {
	var value model.IssueFieldValue
	err := r.db.WithContext(ctx).
		Where("issue_id = ? AND field_id = ?", issueID, fieldID).
		First(&value).Error
	if err != nil {
		return nil, err
	}
	return &value, nil
}

// ListByIssue 获取工单的所有字段值
func (r *valueRepository) ListByIssue(ctx context.Context, issueID uint64) ([]*model.IssueFieldValue, error) {
	var values []*model.IssueFieldValue
	err := r.db.WithContext(ctx).
		Where("issue_id = ?", issueID).
		Find(&values).Error
	return values, err
}

// BatchUpsert 批量插入或更新字段值（MySQL ON DUPLICATE KEY UPDATE）
func (r *valueRepository) BatchUpsert(ctx context.Context, values []*model.IssueFieldValue) error {
	if len(values) == 0 {
		return nil
	}
	return r.db.WithContext(ctx).
		Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "issue_id"}, {Name: "field_id"}},
			DoUpdates: clause.AssignmentColumns([]string{
				"value_text", "value_number", "value_date", "value_json", "updated_at",
			}),
		}).Create(&values).Error
}
