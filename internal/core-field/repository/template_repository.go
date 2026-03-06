// Package repository 提供字段方案模板数据访问层
package repository

import (
	"context"

	"gorm.io/gorm"

	"github.com/kerbos/ticketdesk/internal/model"
)

// TemplateRepository 字段方案模板仓储接口
type TemplateRepository interface {
	Create(ctx context.Context, tpl *model.FieldSchemeTemplate) error
	Update(ctx context.Context, tpl *model.FieldSchemeTemplate) error
	Delete(ctx context.Context, id uint64) error
	GetByID(ctx context.Context, id uint64) (*model.FieldSchemeTemplate, error)
	GetByName(ctx context.Context, name string) (*model.FieldSchemeTemplate, error)
	List(ctx context.Context) ([]*model.FieldSchemeTemplate, error)
	// 模板项
	ListItemsByTemplate(ctx context.Context, templateID uint64) ([]*model.FieldSchemeTemplateItem, error)
	DeleteItemsByTemplate(ctx context.Context, templateID uint64) error
	BatchCreateItems(ctx context.Context, items []*model.FieldSchemeTemplateItem) error
	CountByFieldID(ctx context.Context, fieldID uint64) (int64, error)
	WithTx(tx *gorm.DB) TemplateRepository
}

// templateRepository 字段方案模板仓储实现
type templateRepository struct {
	db *gorm.DB
}

// NewTemplateRepository 创建字段方案模板仓储实例
func NewTemplateRepository(db *gorm.DB) TemplateRepository {
	return &templateRepository{db: db}
}

// Create 创建模板
func (r *templateRepository) Create(ctx context.Context, tpl *model.FieldSchemeTemplate) error {
	return r.db.WithContext(ctx).Create(tpl).Error
}

// Update 更新模板
func (r *templateRepository) Update(ctx context.Context, tpl *model.FieldSchemeTemplate) error {
	return r.db.WithContext(ctx).Save(tpl).Error
}

// Delete 硬删除模板（级联删除模板项）
func (r *templateRepository) Delete(ctx context.Context, id uint64) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Unscoped().Where("template_id = ?", id).Delete(&model.FieldSchemeTemplateItem{}).Error; err != nil {
			return err
		}
		return tx.Unscoped().Delete(&model.FieldSchemeTemplate{}, id).Error
	})
}

// GetByID 根据ID获取模板
func (r *templateRepository) GetByID(ctx context.Context, id uint64) (*model.FieldSchemeTemplate, error) {
	var tpl model.FieldSchemeTemplate
	err := r.db.WithContext(ctx).First(&tpl, id).Error
	if err != nil {
		return nil, err
	}
	return &tpl, nil
}

// GetByName 根据名称获取模板
func (r *templateRepository) GetByName(ctx context.Context, name string) (*model.FieldSchemeTemplate, error) {
	var tpl model.FieldSchemeTemplate
	err := r.db.WithContext(ctx).Where("name = ?", name).First(&tpl).Error
	if err != nil {
		return nil, err
	}
	return &tpl, nil
}

// List 获取所有模板
func (r *templateRepository) List(ctx context.Context) ([]*model.FieldSchemeTemplate, error) {
	var templates []*model.FieldSchemeTemplate
	err := r.db.WithContext(ctx).Order("id ASC").Find(&templates).Error
	return templates, err
}

// ListItemsByTemplate 获取模板的所有字段项
func (r *templateRepository) ListItemsByTemplate(ctx context.Context, templateID uint64) ([]*model.FieldSchemeTemplateItem, error) {
	var items []*model.FieldSchemeTemplateItem
	err := r.db.WithContext(ctx).
		Where("template_id = ?", templateID).
		Order("sort_order ASC, id ASC").
		Find(&items).Error
	return items, err
}

// DeleteItemsByTemplate 删除模板的所有字段项（硬删除）
func (r *templateRepository) DeleteItemsByTemplate(ctx context.Context, templateID uint64) error {
	return r.db.WithContext(ctx).Unscoped().
		Where("template_id = ?", templateID).
		Delete(&model.FieldSchemeTemplateItem{}).Error
}

// BatchCreateItems 批量创建模板字段项
func (r *templateRepository) BatchCreateItems(ctx context.Context, items []*model.FieldSchemeTemplateItem) error {
	if len(items) == 0 {
		return nil
	}
	return r.db.WithContext(ctx).Create(&items).Error
}

// CountByFieldID 统计引用指定字段的模板项数量
func (r *templateRepository) CountByFieldID(ctx context.Context, fieldID uint64) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&model.FieldSchemeTemplateItem{}).
		Where("field_id = ?", fieldID).
		Count(&count).Error
	return count, err
}

// WithTx 返回使用指定事务的仓储实例
func (r *templateRepository) WithTx(tx *gorm.DB) TemplateRepository {
	return &templateRepository{db: tx}
}
