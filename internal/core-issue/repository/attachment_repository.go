// Package repository 提供附件数据访问层
package repository

import (
	"context"

	"gorm.io/gorm"

	"github.com/kerbos/ticketdesk/internal/model"
)

// AttachmentRepository 附件数据访问接口
type AttachmentRepository interface {
	Create(ctx context.Context, attachment *model.IssueAttachment) error
	GetByID(ctx context.Context, id uint64) (*model.IssueAttachment, error)
	ListByIssueID(ctx context.Context, issueID uint64) ([]*model.IssueAttachment, error)
	Delete(ctx context.Context, id uint64) error
}

// attachmentRepository 附件数据访问实现
type attachmentRepository struct {
	db *gorm.DB
}

// NewAttachmentRepository 创建附件数据访问实例
func NewAttachmentRepository(db *gorm.DB) AttachmentRepository {
	return &attachmentRepository{db: db}
}

// Create 创建附件记录
func (r *attachmentRepository) Create(ctx context.Context, attachment *model.IssueAttachment) error {
	return r.db.WithContext(ctx).Create(attachment).Error
}

// GetByID 根据ID获取附件
func (r *attachmentRepository) GetByID(ctx context.Context, id uint64) (*model.IssueAttachment, error) {
	var attachment model.IssueAttachment
	err := r.db.WithContext(ctx).First(&attachment, id).Error
	if err != nil {
		return nil, err
	}
	return &attachment, nil
}

// ListByIssueID 获取工单的所有附件
func (r *attachmentRepository) ListByIssueID(ctx context.Context, issueID uint64) ([]*model.IssueAttachment, error) {
	var attachments []*model.IssueAttachment
	err := r.db.WithContext(ctx).
		Where("issue_id = ?", issueID).
		Order("created_at DESC").
		Find(&attachments).Error
	return attachments, err
}

// Delete 硬删除附件记录
func (r *attachmentRepository) Delete(ctx context.Context, id uint64) error {
	return r.db.WithContext(ctx).Unscoped().Delete(&model.IssueAttachment{}, id).Error
}
