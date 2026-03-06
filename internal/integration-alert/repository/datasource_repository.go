// Package repository 提供告警数据源数据访问层
package repository

import (
	"context"
	"time"

	"gorm.io/gorm"

	"github.com/kerbos/ticketdesk/internal/integration-alert/dto"
	"github.com/kerbos/ticketdesk/internal/model"
)

// AlertDatasourceRepository 告警数据源仓库接口
type AlertDatasourceRepository interface {
	Create(ctx context.Context, ds *model.AlertDatasource) error
	GetByID(ctx context.Context, id uint64) (*model.AlertDatasource, error)
	GetByName(ctx context.Context, name string) (*model.AlertDatasource, error)
	Update(ctx context.Context, ds *model.AlertDatasource) error
	Delete(ctx context.Context, id uint64) error
	List(ctx context.Context, req *dto.DatasourceListRequest) ([]*model.AlertDatasource, int64, error)
	ListEnabled(ctx context.Context) ([]*model.AlertDatasource, error)
	UpdateCheckResult(ctx context.Context, id uint64, ok bool, msg string) error
}

// alertDatasourceRepository 告警数据源仓库实现
type alertDatasourceRepository struct {
	db *gorm.DB
}

// NewAlertDatasourceRepository 创建告警数据源仓库
func NewAlertDatasourceRepository(db *gorm.DB) AlertDatasourceRepository {
	return &alertDatasourceRepository{db: db}
}

// Create 创建数据源
func (r *alertDatasourceRepository) Create(ctx context.Context, ds *model.AlertDatasource) error {
	return r.db.WithContext(ctx).Create(ds).Error
}

// GetByID 根据 ID 获取数据源
func (r *alertDatasourceRepository) GetByID(ctx context.Context, id uint64) (*model.AlertDatasource, error) {
	var ds model.AlertDatasource
	err := r.db.WithContext(ctx).First(&ds, id).Error
	if err != nil {
		return nil, err
	}
	return &ds, nil
}

// GetByName 根据名称获取数据源
func (r *alertDatasourceRepository) GetByName(ctx context.Context, name string) (*model.AlertDatasource, error) {
	var ds model.AlertDatasource
	err := r.db.WithContext(ctx).Where("name = ?", name).First(&ds).Error
	if err != nil {
		return nil, err
	}
	return &ds, nil
}

// Update 更新数据源
func (r *alertDatasourceRepository) Update(ctx context.Context, ds *model.AlertDatasource) error {
	return r.db.WithContext(ctx).Save(ds).Error
}

// Delete 硬删除数据源
func (r *alertDatasourceRepository) Delete(ctx context.Context, id uint64) error {
	return r.db.WithContext(ctx).Unscoped().Delete(&model.AlertDatasource{}, id).Error
}

// List 获取数据源列表
func (r *alertDatasourceRepository) List(ctx context.Context, req *dto.DatasourceListRequest) ([]*model.AlertDatasource, int64, error) {
	var datasources []*model.AlertDatasource
	var total int64

	query := r.db.WithContext(ctx).Model(&model.AlertDatasource{})

	if req.Type != "" {
		query = query.Where("type = ?", req.Type)
	}
	if req.Status != nil {
		query = query.Where("status = ?", *req.Status)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	page := req.Page
	if page < 1 {
		page = 1
	}
	pageSize := req.PageSize
	if pageSize < 1 {
		pageSize = 20
	}
	offset := (page - 1) * pageSize

	err := query.Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&datasources).Error
	return datasources, total, err
}

// ListEnabled 获取所有启用的数据源
func (r *alertDatasourceRepository) ListEnabled(ctx context.Context) ([]*model.AlertDatasource, error) {
	var datasources []*model.AlertDatasource
	err := r.db.WithContext(ctx).Where("status = ?", 1).Find(&datasources).Error
	return datasources, err
}

// UpdateCheckResult 更新数据源连接检测结果
func (r *alertDatasourceRepository) UpdateCheckResult(ctx context.Context, id uint64, ok bool, msg string) error {
	now := time.Now()
	return r.db.WithContext(ctx).Model(&model.AlertDatasource{}).Where("id = ?", id).Updates(map[string]interface{}{
		"last_check_at":  now,
		"last_check_ok":  ok,
		"last_check_msg": msg,
	}).Error
}
