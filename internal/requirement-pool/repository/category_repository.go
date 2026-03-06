package repository

import (
	"context"

	"gorm.io/gorm"

	"github.com/kerbos/ticketdesk/internal/model"
)

// CategoryRepository 需求分类数据访问接口
type CategoryRepository interface {
	List(ctx context.Context) ([]*model.RequirementCategoryDef, error)
	GetByID(ctx context.Context, id uint64) (*model.RequirementCategoryDef, error)
	GetByName(ctx context.Context, name string) (*model.RequirementCategoryDef, error)
	Create(ctx context.Context, cat *model.RequirementCategoryDef) error
	Update(ctx context.Context, cat *model.RequirementCategoryDef) error
	UpdateFields(ctx context.Context, id uint64, fields map[string]any) error
	Delete(ctx context.Context, id uint64) error
	CountRequirementsByCategory(ctx context.Context, categoryName string) (int64, error)
}

// categoryRepository 需求分类数据访问实现
type categoryRepository struct {
	db *gorm.DB
}

// NewCategoryRepository 创建需求分类数据访问实例
func NewCategoryRepository(db *gorm.DB) CategoryRepository {
	return &categoryRepository{db: db}
}

// List 获取所有分类（按排序）
func (r *categoryRepository) List(ctx context.Context) ([]*model.RequirementCategoryDef, error) {
	var categories []*model.RequirementCategoryDef
	err := r.db.WithContext(ctx).Order("sort_order ASC, id ASC").Find(&categories).Error
	return categories, err
}

// GetByID 根据ID获取分类
func (r *categoryRepository) GetByID(ctx context.Context, id uint64) (*model.RequirementCategoryDef, error) {
	var cat model.RequirementCategoryDef
	err := r.db.WithContext(ctx).First(&cat, id).Error
	return &cat, err
}

// GetByName 根据名称获取分类
func (r *categoryRepository) GetByName(ctx context.Context, name string) (*model.RequirementCategoryDef, error) {
	var cat model.RequirementCategoryDef
	err := r.db.WithContext(ctx).Where("name = ?", name).First(&cat).Error
	return &cat, err
}

// Create 创建分类
func (r *categoryRepository) Create(ctx context.Context, cat *model.RequirementCategoryDef) error {
	return r.db.WithContext(ctx).Create(cat).Error
}

// Update 更新分类
func (r *categoryRepository) Update(ctx context.Context, cat *model.RequirementCategoryDef) error {
	return r.db.WithContext(ctx).Save(cat).Error
}

// UpdateFields 按字段更新分类
func (r *categoryRepository) UpdateFields(ctx context.Context, id uint64, fields map[string]any) error {
	return r.db.WithContext(ctx).Model(&model.RequirementCategoryDef{}).Where("id = ?", id).Updates(fields).Error
}

// Delete 删除分类（硬删除）
func (r *categoryRepository) Delete(ctx context.Context, id uint64) error {
	return r.db.WithContext(ctx).Unscoped().Delete(&model.RequirementCategoryDef{}, id).Error
}

// CountRequirementsByCategory 统计指定分类下的需求数量
func (r *categoryRepository) CountRequirementsByCategory(ctx context.Context, categoryName string) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&model.Requirement{}).Where("category = ?", categoryName).Count(&count).Error
	return count, err
}
