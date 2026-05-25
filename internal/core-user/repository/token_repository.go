// Package repository 提供用户模块数据访问层
package repository

import (
	"context"
	"time"

	"gorm.io/gorm"

	"github.com/kerbos/ticketdesk/internal/model"
)

// APITokenRepository API token 数据访问接口
type APITokenRepository interface {
	Create(ctx context.Context, t *model.APIToken) error
	GetByHash(ctx context.Context, hash string) (*model.APIToken, error)
	ListByUserID(ctx context.Context, userID uint64) ([]*model.APIToken, error)
	GetByIDAndUserID(ctx context.Context, id, userID uint64) (*model.APIToken, error)
	UpdateLastUsed(ctx context.Context, id uint64) error
	Delete(ctx context.Context, id uint64) error
	CountByUserID(ctx context.Context, userID uint64) (int64, error)
}

type apiTokenRepository struct {
	db *gorm.DB
}

// NewAPITokenRepository 创建 API token 数据访问实例
func NewAPITokenRepository(db *gorm.DB) APITokenRepository {
	return &apiTokenRepository{db: db}
}

// Create 创建 token 记录
func (r *apiTokenRepository) Create(ctx context.Context, t *model.APIToken) error {
	return r.db.WithContext(ctx).Create(t).Error
}

// GetByHash 查活跃 token（未删且未过期）
func (r *apiTokenRepository) GetByHash(ctx context.Context, hash string) (*model.APIToken, error) {
	var t model.APIToken
	err := r.db.WithContext(ctx).
		Where("token_hash = ?", hash).
		Where("expires_at IS NULL OR expires_at > ?", time.Now()).
		First(&t).Error
	if err != nil {
		return nil, err
	}
	return &t, nil
}

// ListByUserID 查询用户所有 token
func (r *apiTokenRepository) ListByUserID(ctx context.Context, userID uint64) ([]*model.APIToken, error) {
	var tokens []*model.APIToken
	err := r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Order("created_at DESC").
		Find(&tokens).Error
	return tokens, err
}

// GetByIDAndUserID 查询指定用户的指定 token
func (r *apiTokenRepository) GetByIDAndUserID(ctx context.Context, id, userID uint64) (*model.APIToken, error) {
	var t model.APIToken
	err := r.db.WithContext(ctx).
		Where("id = ? AND user_id = ?", id, userID).
		First(&t).Error
	if err != nil {
		return nil, err
	}
	return &t, nil
}

// UpdateLastUsed 更新最后使用时间
func (r *apiTokenRepository) UpdateLastUsed(ctx context.Context, id uint64) error {
	now := time.Now()
	return r.db.WithContext(ctx).
		Model(&model.APIToken{}).
		Where("id = ?", id).
		Update("last_used_at", &now).Error
}

// Delete 软删除 token
func (r *apiTokenRepository) Delete(ctx context.Context, id uint64) error {
	return r.db.WithContext(ctx).Delete(&model.APIToken{}, id).Error
}

// CountByUserID 统计用户 token 数量
func (r *apiTokenRepository) CountByUserID(ctx context.Context, userID uint64) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&model.APIToken{}).
		Where("user_id = ?", userID).
		Count(&count).Error
	return count, err
}
