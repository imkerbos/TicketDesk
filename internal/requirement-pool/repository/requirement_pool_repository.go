package repository

import (
	"context"

	"gorm.io/gorm"

	"github.com/kerbos/ticketdesk/internal/model"
)

// RequirementPoolRepository 需求池数据访问接口
type RequirementPoolRepository interface {
	Create(ctx context.Context, pool *model.RequirementPool) error
	GetByID(ctx context.Context, id uint64) (*model.RequirementPool, error)
	Update(ctx context.Context, pool *model.RequirementPool) error
	Delete(ctx context.Context, id uint64) error
	List(ctx context.Context, filter *RequirementPoolFilter) ([]*model.RequirementPool, int64, error)
	GetByProjectID(ctx context.Context, projectID uint64) ([]*model.RequirementPool, error)
}

// RequirementPoolFilter 需求池查询过滤器
type RequirementPoolFilter struct {
	Type      *model.RequirementPoolType
	Status    *model.RequirementPoolStatus
	ProjectID *uint64
	OwnerID   *uint64
	Page      int
	PageSize  int
}

// requirementPoolRepository 需求池数据访问实现
type requirementPoolRepository struct {
	db *gorm.DB
}

// NewRequirementPoolRepository 创建需求池数据访问实例
func NewRequirementPoolRepository(db *gorm.DB) RequirementPoolRepository {
	return &requirementPoolRepository{db: db}
}

// Create 创建需求池
func (r *requirementPoolRepository) Create(ctx context.Context, pool *model.RequirementPool) error {
	return r.db.WithContext(ctx).Create(pool).Error
}

// GetByID 根据ID获取需求池
func (r *requirementPoolRepository) GetByID(ctx context.Context, id uint64) (*model.RequirementPool, error) {
	var pool model.RequirementPool
	err := r.db.WithContext(ctx).
		Preload("Owner").
		Preload("Project").
		First(&pool, id).Error
	if err != nil {
		return nil, err
	}
	return &pool, nil
}

// Update 更新需求池
func (r *requirementPoolRepository) Update(ctx context.Context, pool *model.RequirementPool) error {
	return r.db.WithContext(ctx).Save(pool).Error
}

// Delete 删除需求池（硬删除）
func (r *requirementPoolRepository) Delete(ctx context.Context, id uint64) error {
	return r.db.WithContext(ctx).Unscoped().Delete(&model.RequirementPool{}, id).Error
}

// List 获取需求池列表
func (r *requirementPoolRepository) List(ctx context.Context, filter *RequirementPoolFilter) ([]*model.RequirementPool, int64, error) {
	var pools []*model.RequirementPool
	var total int64

	query := r.db.WithContext(ctx).Model(&model.RequirementPool{})

	// 应用过滤条件
	if filter.Type != nil {
		query = query.Where("type = ?", *filter.Type)
	}
	if filter.Status != nil {
		query = query.Where("status = ?", *filter.Status)
	}
	if filter.ProjectID != nil {
		query = query.Where("project_id = ?", *filter.ProjectID)
	}
	if filter.OwnerID != nil {
		query = query.Where("owner_id = ?", *filter.OwnerID)
	}

	// 统计总数
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 分页
	if filter.Page > 0 && filter.PageSize > 0 {
		offset := (filter.Page - 1) * filter.PageSize
		query = query.Offset(offset).Limit(filter.PageSize)
	}

	// 预加载关联数据
	query = query.Preload("Owner").Preload("Project")

	// 查询数据
	if err := query.Order("created_at DESC").Find(&pools).Error; err != nil {
		return nil, 0, err
	}

	return pools, total, nil
}

// GetByProjectID 根据项目ID获取需求池列表
func (r *requirementPoolRepository) GetByProjectID(ctx context.Context, projectID uint64) ([]*model.RequirementPool, error) {
	var pools []*model.RequirementPool
	err := r.db.WithContext(ctx).
		Where("project_id = ? AND status = ?", projectID, model.RequirementPoolStatusActive).
		Preload("Owner").
		Order("created_at DESC").
		Find(&pools).Error
	if err != nil {
		return nil, err
	}
	return pools, nil
}
