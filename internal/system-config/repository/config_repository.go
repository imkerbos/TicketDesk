// Package repository 提供系统配置数据访问层
package repository

import (
	"context"

	"gorm.io/gorm"

	"github.com/kerbos/ticketdesk/internal/model"
)

// ConfigRepository 系统配置数据访问接口
type ConfigRepository interface {
	GetByKey(ctx context.Context, key string) (*model.SystemConfig, error)
	GetByCategory(ctx context.Context, category string) ([]*model.SystemConfig, error)
	GetAll(ctx context.Context) ([]*model.SystemConfig, error)
	Upsert(ctx context.Context, config *model.SystemConfig) error
	BatchUpsert(ctx context.Context, configs []*model.SystemConfig) error
	Delete(ctx context.Context, key string) error
}

// configRepository 系统配置数据访问实现
type configRepository struct {
	db *gorm.DB
}

// NewConfigRepository 创建系统配置数据访问实例
func NewConfigRepository(db *gorm.DB) ConfigRepository {
	return &configRepository{db: db}
}

// GetByKey 根据键获取配置
func (r *configRepository) GetByKey(ctx context.Context, key string) (*model.SystemConfig, error) {
	var config model.SystemConfig
	err := r.db.WithContext(ctx).Where("config_key = ?", key).First(&config).Error
	if err != nil {
		return nil, err
	}
	return &config, nil
}

// GetByCategory 根据分类获取配置
func (r *configRepository) GetByCategory(ctx context.Context, category string) ([]*model.SystemConfig, error) {
	var configs []*model.SystemConfig
	err := r.db.WithContext(ctx).Where("category = ?", category).Find(&configs).Error
	return configs, err
}

// GetAll 获取所有配置
func (r *configRepository) GetAll(ctx context.Context) ([]*model.SystemConfig, error) {
	var configs []*model.SystemConfig
	err := r.db.WithContext(ctx).Order("category, config_key").Find(&configs).Error
	return configs, err
}

// Upsert 创建或更新配置
func (r *configRepository) Upsert(ctx context.Context, config *model.SystemConfig) error {
	return r.db.WithContext(ctx).
		Where("config_key = ?", config.ConfigKey).
		Assign(model.SystemConfig{
			ConfigValue: config.ConfigValue,
			ConfigType:  config.ConfigType,
			Category:    config.Category,
			Description: config.Description,
			IsSecret:    config.IsSecret,
			UpdatedBy:   config.UpdatedBy,
		}).
		FirstOrCreate(config).Error
}

// BatchUpsert 批量创建或更新配置
func (r *configRepository) BatchUpsert(ctx context.Context, configs []*model.SystemConfig) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for _, config := range configs {
			err := tx.Where("config_key = ?", config.ConfigKey).
				Assign(model.SystemConfig{
					ConfigValue: config.ConfigValue,
					ConfigType:  config.ConfigType,
					Category:    config.Category,
					Description: config.Description,
					IsSecret:    config.IsSecret,
					UpdatedBy:   config.UpdatedBy,
				}).
				FirstOrCreate(config).Error
			if err != nil {
				return err
			}
		}
		return nil
	})
}

// Delete 硬删除配置
func (r *configRepository) Delete(ctx context.Context, key string) error {
	return r.db.WithContext(ctx).Unscoped().Where("config_key = ?", key).Delete(&model.SystemConfig{}).Error
}

// WebhookRepository Webhook 数据访问接口
type WebhookRepository interface {
	Create(ctx context.Context, webhook *model.Webhook) error
	GetByID(ctx context.Context, id uint64) (*model.Webhook, error)
	Update(ctx context.Context, webhook *model.Webhook) error
	Delete(ctx context.Context, id uint64) error
	List(ctx context.Context, offset, limit int, status *int8, event string) ([]*model.Webhook, int64, error)
	GetByEvent(ctx context.Context, event string) ([]*model.Webhook, error)
	GetAllEnabled(ctx context.Context) ([]*model.Webhook, error)
}

// webhookRepository Webhook 数据访问实现
type webhookRepository struct {
	db *gorm.DB
}

// NewWebhookRepository 创建 Webhook 数据访问实例
func NewWebhookRepository(db *gorm.DB) WebhookRepository {
	return &webhookRepository{db: db}
}

// Create 创建 Webhook
func (r *webhookRepository) Create(ctx context.Context, webhook *model.Webhook) error {
	return r.db.WithContext(ctx).Create(webhook).Error
}

// GetByID 根据 ID 获取 Webhook
func (r *webhookRepository) GetByID(ctx context.Context, id uint64) (*model.Webhook, error) {
	var webhook model.Webhook
	err := r.db.WithContext(ctx).First(&webhook, id).Error
	if err != nil {
		return nil, err
	}
	return &webhook, nil
}

// Update 更新 Webhook
func (r *webhookRepository) Update(ctx context.Context, webhook *model.Webhook) error {
	return r.db.WithContext(ctx).Save(webhook).Error
}

// Delete 硬删除 Webhook
func (r *webhookRepository) Delete(ctx context.Context, id uint64) error {
	return r.db.WithContext(ctx).Unscoped().Delete(&model.Webhook{}, id).Error
}

// List 分页查询 Webhook 列表
func (r *webhookRepository) List(ctx context.Context, offset, limit int, status *int8, event string) ([]*model.Webhook, int64, error) {
	var webhooks []*model.Webhook
	var total int64

	query := r.db.WithContext(ctx).Model(&model.Webhook{})

	if status != nil {
		query = query.Where("status = ?", *status)
	}

	if event != "" {
		query = query.Where("JSON_CONTAINS(events, ?)", `"`+event+`"`)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := query.Offset(offset).Limit(limit).Order("id DESC").Find(&webhooks).Error; err != nil {
		return nil, 0, err
	}

	return webhooks, total, nil
}

// GetByEvent 根据事件类型获取启用的 Webhook
func (r *webhookRepository) GetByEvent(ctx context.Context, event string) ([]*model.Webhook, error) {
	var webhooks []*model.Webhook
	err := r.db.WithContext(ctx).
		Where("status = ?", 1).
		Where("JSON_CONTAINS(events, ?)", `"`+event+`"`).
		Find(&webhooks).Error
	return webhooks, err
}

// GetAllEnabled 获取所有启用的 Webhook
func (r *webhookRepository) GetAllEnabled(ctx context.Context) ([]*model.Webhook, error) {
	var webhooks []*model.Webhook
	err := r.db.WithContext(ctx).Where("status = ?", 1).Find(&webhooks).Error
	return webhooks, err
}

// WebhookLogRepository Webhook 日志数据访问接口
type WebhookLogRepository interface {
	Create(ctx context.Context, log *model.WebhookLog) error
	Update(ctx context.Context, log *model.WebhookLog) error
	GetByID(ctx context.Context, id uint64) (*model.WebhookLog, error)
	List(ctx context.Context, offset, limit int, webhookID uint64, event string, status *int8) ([]*model.WebhookLog, int64, error)
	GetPendingLogs(ctx context.Context, limit int) ([]*model.WebhookLog, error)
}

// webhookLogRepository Webhook 日志数据访问实现
type webhookLogRepository struct {
	db *gorm.DB
}

// NewWebhookLogRepository 创建 Webhook 日志数据访问实例
func NewWebhookLogRepository(db *gorm.DB) WebhookLogRepository {
	return &webhookLogRepository{db: db}
}

// Create 创建日志
func (r *webhookLogRepository) Create(ctx context.Context, log *model.WebhookLog) error {
	return r.db.WithContext(ctx).Create(log).Error
}

// Update 更新日志
func (r *webhookLogRepository) Update(ctx context.Context, log *model.WebhookLog) error {
	return r.db.WithContext(ctx).Save(log).Error
}

// GetByID 根据 ID 获取日志
func (r *webhookLogRepository) GetByID(ctx context.Context, id uint64) (*model.WebhookLog, error) {
	var log model.WebhookLog
	err := r.db.WithContext(ctx).First(&log, id).Error
	if err != nil {
		return nil, err
	}
	return &log, nil
}

// List 分页查询日志
func (r *webhookLogRepository) List(ctx context.Context, offset, limit int, webhookID uint64, event string, status *int8) ([]*model.WebhookLog, int64, error) {
	var logs []*model.WebhookLog
	var total int64

	query := r.db.WithContext(ctx).Model(&model.WebhookLog{})

	if webhookID > 0 {
		query = query.Where("webhook_id = ?", webhookID)
	}

	if event != "" {
		query = query.Where("event = ?", event)
	}

	if status != nil {
		query = query.Where("status = ?", *status)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := query.Offset(offset).Limit(limit).Order("id DESC").Find(&logs).Error; err != nil {
		return nil, 0, err
	}

	return logs, total, nil
}

// GetPendingLogs 获取待发送的日志（用于重试）
func (r *webhookLogRepository) GetPendingLogs(ctx context.Context, limit int) ([]*model.WebhookLog, error) {
	var logs []*model.WebhookLog
	err := r.db.WithContext(ctx).
		Where("status = ? AND retry_count < ?", 0, 3).
		Limit(limit).
		Order("id ASC").
		Find(&logs).Error
	return logs, err
}
