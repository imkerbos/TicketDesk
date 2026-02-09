// Package repository 提供字段值数据访问层
package repository

import (
	"context"

	"github.com/kerbos/ticketdesk/internal/model"
	"github.com/kerbos/ticketdesk/pkg/logger"
	"go.uber.org/zap"
	"gorm.io/gorm"
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

// BatchUpsert 批量插入或更新字段值
func (r *valueRepository) BatchUpsert(ctx context.Context, values []*model.IssueFieldValue) error {
	logger.Info("BatchUpsert: starting", zap.Int("values_count", len(values)))
	if len(values) == 0 {
		logger.Info("BatchUpsert: no values to save")
		return nil
	}

	for _, v := range values {
		valueJSON := ""
		if v.ValueJSON != nil {
			valueJSON = *v.ValueJSON
		}
		logger.Info("BatchUpsert: processing value",
			zap.Uint64("issue_id", v.IssueID),
			zap.Uint64("field_id", v.FieldID),
			zap.Any("value_text", v.ValueText),
			zap.Any("value_number", v.ValueNumber),
			zap.Any("value_date", v.ValueDate),
			zap.String("value_json", valueJSON),
		)
		var existing model.IssueFieldValue
		err := r.db.WithContext(ctx).
			Where("issue_id = ? AND field_id = ?", v.IssueID, v.FieldID).
			First(&existing).Error

		if err == gorm.ErrRecordNotFound {
			// 创建新记录
			if err := r.db.WithContext(ctx).Create(v).Error; err != nil {
				logger.Error("BatchUpsert: failed to create", zap.Error(err))
				return err
			}
			logger.Info("BatchUpsert: created new record", zap.Uint64("id", v.ID))
		} else if err != nil {
			logger.Error("BatchUpsert: failed to query existing", zap.Error(err))
			return err
		} else {
			// 更新现有记录
			v.ID = existing.ID
			if err := r.db.WithContext(ctx).Save(v).Error; err != nil {
				logger.Error("BatchUpsert: failed to update", zap.Error(err))
				return err
			}
			logger.Info("BatchUpsert: updated existing record", zap.Uint64("id", v.ID))
		}
	}

	logger.Info("BatchUpsert: completed successfully")
	return nil
}
