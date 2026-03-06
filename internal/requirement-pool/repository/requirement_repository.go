package repository

import (
	"context"
	"fmt"
	"time"

	"gorm.io/gorm"

	"github.com/kerbos/ticketdesk/internal/model"
)

// RequirementRepository 需求数据访问接口
type RequirementRepository interface {
	Create(ctx context.Context, req *model.Requirement) error
	GetByID(ctx context.Context, id uint64) (*model.Requirement, error)
	Update(ctx context.Context, req *model.Requirement) error
	UpdateFields(ctx context.Context, id uint64, fields map[string]interface{}) error
	Delete(ctx context.Context, id uint64) error
	List(ctx context.Context, filter *RequirementFilter) ([]*model.Requirement, int64, error)
	GetByPoolID(ctx context.Context, poolID uint64) ([]*model.Requirement, error)
	CountByStatus(ctx context.Context, poolID uint64) (map[model.RequirementStatus]int64, error)
	CountByPriority(ctx context.Context, poolID uint64) (map[model.RequirementPriority]int64, error)
	GetKanbanData(ctx context.Context, filter *KanbanFilter) (map[string][]*model.Requirement, error)
	GetStatistics(ctx context.Context, filter *StatisticsFilter) (*Statistics, error)
}

// RequirementFilter 需求查询过滤器
type RequirementFilter struct {
	PoolID     *uint64
	Status     *model.RequirementStatus
	Priority   *model.RequirementPriority
	Category   *model.RequirementCategory
	AssigneeID *uint64
	CreatedBy  *uint64
	StartDate  *time.Time
	EndDate    *time.Time
	Keyword    string
	Page       int
	PageSize   int
}

// KanbanFilter 看板查询过滤器
type KanbanFilter struct {
	PoolID    *uint64
	GroupBy   string // status, priority, assignee, timeline
	StartDate *time.Time
	EndDate   *time.Time
}

// StatisticsFilter 统计查询过滤器
type StatisticsFilter struct {
	PoolID    *uint64
	StartDate time.Time
	EndDate   time.Time
}

// Statistics 统计数据
type Statistics struct {
	TotalCreated   int64
	TotalCompleted int64
	TotalRejected  int64
	AvgProcessDays float64
}

// requirementRepository 需求数据访问实现
type requirementRepository struct {
	db *gorm.DB
}

// NewRequirementRepository 创建需求数据访问实例
func NewRequirementRepository(db *gorm.DB) RequirementRepository {
	return &requirementRepository{db: db}
}

// Create 创建需求
func (r *requirementRepository) Create(ctx context.Context, req *model.Requirement) error {
	return r.db.WithContext(ctx).Create(req).Error
}

// GetByID 根据ID获取需求
func (r *requirementRepository) GetByID(ctx context.Context, id uint64) (*model.Requirement, error) {
	var req model.Requirement
	err := r.db.WithContext(ctx).
		Preload("Pool").
		Preload("Reporter").
		Preload("Assignee").
		Preload("Creator").
		Preload("ConvertedIssue").
		Preload("TargetProject").
		Preload("Comments").
		Preload("Attachments").
		First(&req, id).Error
	if err != nil {
		return nil, err
	}
	return &req, nil
}

// Update 更新需求
func (r *requirementRepository) Update(ctx context.Context, req *model.Requirement) error {
	return r.db.WithContext(ctx).Save(req).Error
}

// UpdateFields 按字段更新需求（避免 Save + Preload 关联对象干扰）
func (r *requirementRepository) UpdateFields(ctx context.Context, id uint64, fields map[string]interface{}) error {
	return r.db.WithContext(ctx).Model(&model.Requirement{}).Where("id = ?", id).Updates(fields).Error
}

// Delete 删除需求（硬删除）
func (r *requirementRepository) Delete(ctx context.Context, id uint64) error {
	return r.db.WithContext(ctx).Unscoped().Delete(&model.Requirement{}, id).Error
}

// List 获取需求列表
func (r *requirementRepository) List(ctx context.Context, filter *RequirementFilter) ([]*model.Requirement, int64, error) {
	var requirements []*model.Requirement
	var total int64

	query := r.db.WithContext(ctx).Model(&model.Requirement{})

	// 应用过滤条件
	if filter.PoolID != nil {
		query = query.Where("pool_id = ?", *filter.PoolID)
	}
	if filter.Status != nil {
		query = query.Where("status = ?", *filter.Status)
	}
	if filter.Priority != nil {
		query = query.Where("priority = ?", *filter.Priority)
	}
	if filter.Category != nil {
		query = query.Where("category = ?", *filter.Category)
	}
	if filter.AssigneeID != nil {
		query = query.Where("assignee_id = ?", *filter.AssigneeID)
	}
	if filter.CreatedBy != nil {
		query = query.Where("created_by = ?", *filter.CreatedBy)
	}
	if filter.StartDate != nil {
		query = query.Where("created_at >= ?", *filter.StartDate)
	}
	if filter.EndDate != nil {
		query = query.Where("created_at <= ?", *filter.EndDate)
	}
	if filter.Keyword != "" {
		keyword := "%" + filter.Keyword + "%"
		query = query.Where("title LIKE ? OR description LIKE ?", keyword, keyword)
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
	query = query.Preload("Pool").
		Preload("Reporter").
		Preload("Assignee").
		Preload("Creator").
		Preload("ConvertedIssue").
		Preload("TargetProject")

	// 查询数据
	if err := query.Order("created_at DESC").Find(&requirements).Error; err != nil {
		return nil, 0, err
	}

	return requirements, total, nil
}

// GetByPoolID 根据需求池ID获取需求列表
func (r *requirementRepository) GetByPoolID(ctx context.Context, poolID uint64) ([]*model.Requirement, error) {
	var requirements []*model.Requirement
	err := r.db.WithContext(ctx).
		Where("pool_id = ?", poolID).
		Preload("Reporter").
		Preload("Assignee").
		Preload("Creator").
		Order("created_at DESC").
		Find(&requirements).Error
	if err != nil {
		return nil, err
	}
	return requirements, nil
}

// CountByStatus 按状态统计需求数量
func (r *requirementRepository) CountByStatus(ctx context.Context, poolID uint64) (map[model.RequirementStatus]int64, error) {
	type StatusCount struct {
		Status model.RequirementStatus
		Count  int64
	}

	var results []StatusCount
	query := r.db.WithContext(ctx).
		Model(&model.Requirement{}).
		Select("status, COUNT(*) as count").
		Group("status")

	if poolID > 0 {
		query = query.Where("pool_id = ?", poolID)
	}

	if err := query.Find(&results).Error; err != nil {
		return nil, err
	}

	counts := make(map[model.RequirementStatus]int64)
	for _, result := range results {
		counts[result.Status] = result.Count
	}

	return counts, nil
}

// CountByPriority 按优先级统计需求数量
func (r *requirementRepository) CountByPriority(ctx context.Context, poolID uint64) (map[model.RequirementPriority]int64, error) {
	type PriorityCount struct {
		Priority model.RequirementPriority
		Count    int64
	}

	var results []PriorityCount
	query := r.db.WithContext(ctx).
		Model(&model.Requirement{}).
		Select("priority, COUNT(*) as count").
		Group("priority")

	if poolID > 0 {
		query = query.Where("pool_id = ?", poolID)
	}

	if err := query.Find(&results).Error; err != nil {
		return nil, err
	}

	counts := make(map[model.RequirementPriority]int64)
	for _, result := range results {
		counts[result.Priority] = result.Count
	}

	return counts, nil
}

// GetKanbanData 获取看板数据
func (r *requirementRepository) GetKanbanData(ctx context.Context, filter *KanbanFilter) (map[string][]*model.Requirement, error) {
	query := r.db.WithContext(ctx).Model(&model.Requirement{})

	if filter.PoolID != nil {
		query = query.Where("pool_id = ?", *filter.PoolID)
	}
	if filter.StartDate != nil {
		query = query.Where("created_at >= ?", *filter.StartDate)
	}
	if filter.EndDate != nil {
		query = query.Where("created_at <= ?", *filter.EndDate)
	}

	var requirements []*model.Requirement
	err := query.Preload("Pool").
		Preload("Reporter").
		Preload("Assignee").
		Preload("Creator").
		Preload("ConvertedIssue").
		Preload("Comments").
		Order("created_at DESC").
		Find(&requirements).Error
	if err != nil {
		return nil, err
	}

	// 根据 GroupBy 分组
	result := make(map[string][]*model.Requirement)
	for _, req := range requirements {
		var key string
		switch filter.GroupBy {
		case "status":
			key = string(req.Status)
		case "priority":
			key = string(req.Priority)
		case "assignee":
			if req.AssigneeID != nil {
				key = fmt.Sprintf("%d", *req.AssigneeID)
			} else {
				key = "unassigned"
			}
		case "timeline":
			if req.EndDate != nil {
				now := time.Now()
				switch {
				case req.EndDate.Before(now.AddDate(0, 0, 7)):
					key = "this_week"
				case req.EndDate.Before(now.AddDate(0, 1, 0)):
					key = "this_month"
				case req.EndDate.Before(now.AddDate(0, 3, 0)):
					key = "this_quarter"
				default:
					key = "later"
				}
			} else {
				key = "no_date"
			}
		default:
			key = "unknown"
		}
		result[key] = append(result[key], req)
	}

	return result, nil
}

// GetStatistics 获取统计数据
func (r *requirementRepository) GetStatistics(ctx context.Context, filter *StatisticsFilter) (*Statistics, error) {
	query := r.db.WithContext(ctx).Model(&model.Requirement{})

	if filter.PoolID != nil {
		query = query.Where("pool_id = ?", *filter.PoolID)
	}
	query = query.Where("created_at BETWEEN ? AND ?", filter.StartDate, filter.EndDate)

	var stats Statistics

	// 统计创建数量
	query.Count(&stats.TotalCreated)

	// 统计完成数量（completed 状态）
	r.db.WithContext(ctx).Model(&model.Requirement{}).
		Where("created_at BETWEEN ? AND ?", filter.StartDate, filter.EndDate).
		Where("status = ?", model.RequirementStatusCompleted).
		Count(&stats.TotalCompleted)

	// 统计拒绝数量
	r.db.WithContext(ctx).Model(&model.Requirement{}).
		Where("created_at BETWEEN ? AND ?", filter.StartDate, filter.EndDate).
		Where("status = ?", model.RequirementStatusRejected).
		Count(&stats.TotalRejected)

	// 计算平均处理天数
	type AvgDays struct {
		AvgDays float64
	}
	var avgDays AvgDays
	r.db.WithContext(ctx).Model(&model.Requirement{}).
		Select("AVG(DATEDIFF(updated_at, created_at)) as avg_days").
		Where("created_at BETWEEN ? AND ?", filter.StartDate, filter.EndDate).
		Where("status = ?", model.RequirementStatusCompleted).
		Scan(&avgDays)
	stats.AvgProcessDays = avgDays.AvgDays

	return &stats, nil
}
