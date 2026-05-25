// Package service 提供附件业务逻辑层
package service

import (
	"context"
	"errors"
	"fmt"
	"mime/multipart"
	"path/filepath"
	"slices"
	"strings"
	"time"

	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/kerbos/ticketdesk/internal/core-issue/dto"
	"github.com/kerbos/ticketdesk/internal/core-issue/repository"
	userRepo "github.com/kerbos/ticketdesk/internal/core-user/repository"
	"github.com/kerbos/ticketdesk/internal/model"
	"github.com/kerbos/ticketdesk/pkg/logger"
	"github.com/kerbos/ticketdesk/pkg/storage"
)

// 业务错误定义
var (
	ErrAttachmentNotFound = errors.New("附件不存在")
	ErrFileTooLarge       = errors.New("文件大小超过限制")
	ErrInvalidFileType    = errors.New("不支持的文件类型")
)

// 附件校验常量（包级，供 issue_service 与 attachment_service 共享）
const MaxAttachmentSize int64 = 10 * 1024 * 1024 // 10MB

// allowedAttachmentExts 允许的附件扩展名（小写），不对外暴露，通过 IsAllowedAttachmentExt 访问
var allowedAttachmentExts = []string{
	".jpg", ".jpeg", ".png", ".gif", ".bmp", ".webp",
	".pdf", ".doc", ".docx", ".xls", ".xlsx", ".ppt", ".pptx",
	".txt", ".md", ".csv",
	".zip", ".rar", ".7z", ".tar", ".gz",
	".log", ".json", ".xml", ".yaml", ".yml",
}

// imageAttachmentExts 图片扩展名（小写），不对外暴露，通过 IsImageAttachmentExt 访问
var imageAttachmentExts = []string{".jpg", ".jpeg", ".png", ".gif", ".bmp", ".webp"}

// IsAllowedAttachmentExt 判断扩展名是否允许（参数应为小写带点格式，如 ".png"）
func IsAllowedAttachmentExt(ext string) bool {
	return slices.Contains(allowedAttachmentExts, ext)
}

// IsImageAttachmentExt 判断扩展名是否为图片
func IsImageAttachmentExt(ext string) bool {
	return slices.Contains(imageAttachmentExts, ext)
}

// AttachmentService 附件服务接口
type AttachmentService interface {
	UploadAttachment(ctx context.Context, issueKey string, file *multipart.FileHeader, userID uint64) (*dto.AttachmentResponse, error)
	ListAttachments(ctx context.Context, issueKey string) ([]*dto.AttachmentResponse, error)
	DeleteAttachment(ctx context.Context, issueKey string, attachmentID, userID uint64) error
	GetAttachmentPath(ctx context.Context, attachmentID uint64) (string, error)
	SetActivityLogger(activityLogger ActivityLogger)

	// SaveFile 仅落盘, 不写数据库; 返回相对路径
	SaveFile(fileHeader *multipart.FileHeader) (string, error)

	// CreateRecordInTx 在传入事务中写入 issue_attachments 记录
	CreateRecordInTx(ctx context.Context, tx *gorm.DB, issueID uint64, fileHeader *multipart.FileHeader, relPath string, userID uint64) (*model.IssueAttachment, error)

	// DeletePath 删除存储中的文件 (用于事务回滚清理)
	DeletePath(relPath string) error
}

// attachmentService 附件服务实现
type attachmentService struct {
	attachmentRepo repository.AttachmentRepository
	issueRepo      repository.IssueRepository
	userRepo       userRepo.UserRepository
	storage        *storage.LocalStorage
	activityLogger ActivityLogger // 活动日志记录器
}

// NewAttachmentService 创建附件服务实例
func NewAttachmentService(
	attachmentRepo repository.AttachmentRepository,
	issueRepo repository.IssueRepository,
	userRepo userRepo.UserRepository,
	storage *storage.LocalStorage,
) AttachmentService {
	return &attachmentService{
		attachmentRepo: attachmentRepo,
		issueRepo:      issueRepo,
		userRepo:       userRepo,
		storage:        storage,
		activityLogger: nil, // 默认为 nil，可通过 SetActivityLogger 设置
	}
}

// SetActivityLogger 设置活动日志记录器
func (s *attachmentService) SetActivityLogger(activityLogger ActivityLogger) {
	s.activityLogger = activityLogger
}

// UploadAttachment 上传附件
func (s *attachmentService) UploadAttachment(ctx context.Context, issueKey string, fileHeader *multipart.FileHeader, userID uint64) (*dto.AttachmentResponse, error) {
	// 获取工单
	issue, err := s.issueRepo.GetByKey(ctx, strings.ToUpper(issueKey))
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrIssueNotFound
		}
		return nil, fmt.Errorf("查询工单失败: %w", err)
	}

	// 校验 + 落盘（复用 SaveFile）
	relPath, err := s.SaveFile(fileHeader)
	if err != nil {
		return nil, err
	}

	// 创建附件记录
	ext := strings.ToLower(filepath.Ext(fileHeader.Filename))
	attachment := &model.IssueAttachment{
		IssueID:    issue.ID,
		FileName:   fileHeader.Filename,
		FilePath:   relPath,
		FileSize:   fileHeader.Size,
		FileType:   ext,
		IsImage:    IsImageAttachmentExt(ext),
		UploadedBy: userID,
	}

	if err := s.attachmentRepo.Create(ctx, attachment); err != nil {
		// 删除已保存的文件
		_ = s.storage.Delete(relPath)
		logger.Error("failed to create attachment record", zap.Error(err))
		return nil, fmt.Errorf("创建附件记录失败: %w", err)
	}

	// 记录活动日志
	if s.activityLogger != nil {
		user, _ := s.userRepo.GetByID(ctx, userID)
		userName := "未知用户"
		if user != nil {
			userName = user.DisplayName
		}
		_ = s.activityLogger.LogActivity(
			ctx,
			userID,
			userName,
			"上传附件",
			"issue",
			issue.ID,
			issue.IssueKey,
			fmt.Sprintf("上传了附件: %s", fileHeader.Filename),
		)
	}

	return s.toAttachmentResponse(ctx, attachment), nil
}

// ListAttachments 获取工单的所有附件
func (s *attachmentService) ListAttachments(ctx context.Context, issueKey string) ([]*dto.AttachmentResponse, error) {
	// 获取工单
	issue, err := s.issueRepo.GetByKey(ctx, strings.ToUpper(issueKey))
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrIssueNotFound
		}
		return nil, fmt.Errorf("查询工单失败: %w", err)
	}

	// 获取附件列表
	attachments, err := s.attachmentRepo.ListByIssueID(ctx, issue.ID)
	if err != nil {
		logger.Error("failed to list attachments", zap.Error(err))
		return nil, fmt.Errorf("查询附件列表失败: %w", err)
	}

	// 转换为响应格式
	responses := make([]*dto.AttachmentResponse, len(attachments))
	for i, attachment := range attachments {
		responses[i] = s.toAttachmentResponse(ctx, attachment)
	}

	return responses, nil
}

// DeleteAttachment 删除附件
func (s *attachmentService) DeleteAttachment(ctx context.Context, issueKey string, attachmentID, userID uint64) error {
	// 获取工单
	issue, err := s.issueRepo.GetByKey(ctx, strings.ToUpper(issueKey))
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrIssueNotFound
		}
		return fmt.Errorf("查询工单失败: %w", err)
	}

	// 获取附件
	attachment, err := s.attachmentRepo.GetByID(ctx, attachmentID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrAttachmentNotFound
		}
		return fmt.Errorf("查询附件失败: %w", err)
	}

	// 验证附件属于该工单
	if attachment.IssueID != issue.ID {
		return ErrAttachmentNotFound
	}

	// 验证权限（只有上传者可以删除）
	if attachment.UploadedBy != userID {
		return ErrUnauthorized
	}

	// 删除文件
	if err := s.storage.Delete(attachment.FilePath); err != nil {
		logger.Warn("failed to delete file from storage", zap.Error(err))
	}

	// 删除数据库记录
	if err := s.attachmentRepo.Delete(ctx, attachmentID); err != nil {
		logger.Error("failed to delete attachment record", zap.Error(err))
		return fmt.Errorf("删除附件记录失败: %w", err)
	}

	// 记录活动日志
	if s.activityLogger != nil {
		user, _ := s.userRepo.GetByID(ctx, userID)
		userName := "未知用户"
		if user != nil {
			userName = user.DisplayName
		}
		_ = s.activityLogger.LogActivity(
			ctx,
			userID,
			userName,
			"删除附件",
			"issue",
			issue.ID,
			issue.IssueKey,
			fmt.Sprintf("删除了附件: %s", attachment.FileName),
		)
	}

	return nil
}

// GetAttachmentPath 获取附件的完整路径
func (s *attachmentService) GetAttachmentPath(ctx context.Context, attachmentID uint64) (string, error) {
	attachment, err := s.attachmentRepo.GetByID(ctx, attachmentID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return "", ErrAttachmentNotFound
		}
		return "", fmt.Errorf("查询附件失败: %w", err)
	}

	return s.storage.GetFullPath(attachment.FilePath), nil
}

// SaveFile 校验并落盘文件, 返回相对路径; 不写数据库
func (s *attachmentService) SaveFile(fileHeader *multipart.FileHeader) (string, error) {
	if fileHeader.Size > MaxAttachmentSize {
		return "", ErrFileTooLarge
	}
	ext := strings.ToLower(filepath.Ext(fileHeader.Filename))
	if !IsAllowedAttachmentExt(ext) {
		return "", ErrInvalidFileType
	}

	file, err := fileHeader.Open()
	if err != nil {
		return "", fmt.Errorf("打开文件失败: %w", err)
	}
	defer file.Close()

	timestamp := time.Now().Format("20060102150405")
	uniqueFilename := fmt.Sprintf("%s_%s", timestamp, fileHeader.Filename)

	relPath, err := s.storage.Save(file, uniqueFilename)
	if err != nil {
		logger.Error("failed to save file", zap.Error(err))
		return "", fmt.Errorf("保存文件失败: %w", err)
	}
	return relPath, nil
}

// CreateRecordInTx 在传入事务中创建 issue_attachments 记录
// 调用方负责 SaveFile 已成功
func (s *attachmentService) CreateRecordInTx(
	ctx context.Context,
	tx *gorm.DB,
	issueID uint64,
	fileHeader *multipart.FileHeader,
	relPath string,
	userID uint64,
) (*model.IssueAttachment, error) {
	ext := strings.ToLower(filepath.Ext(fileHeader.Filename))
	attachment := &model.IssueAttachment{
		IssueID:    issueID,
		FileName:   fileHeader.Filename,
		FilePath:   relPath,
		FileSize:   fileHeader.Size,
		FileType:   ext,
		IsImage:    IsImageAttachmentExt(ext),
		UploadedBy: userID,
	}
	if err := tx.WithContext(ctx).Create(attachment).Error; err != nil {
		return nil, fmt.Errorf("创建附件记录失败: %w", err)
	}
	return attachment, nil
}

// DeletePath 删除存储中的文件 (用于事务回滚清理)
func (s *attachmentService) DeletePath(relPath string) error {
	return s.storage.Delete(relPath)
}

// toAttachmentResponse 转换为响应格式
func (s *attachmentService) toAttachmentResponse(ctx context.Context, attachment *model.IssueAttachment) *dto.AttachmentResponse {
	resp := &dto.AttachmentResponse{
		ID:         attachment.ID,
		IssueID:    attachment.IssueID,
		FileName:   attachment.FileName,
		FilePath:   attachment.FilePath,
		FileSize:   attachment.FileSize,
		FileType:   attachment.FileType,
		IsImage:    attachment.IsImage,
		UploadedBy: attachment.UploadedBy,
		CreatedAt:  attachment.CreatedAt,
	}

	// 加载上传者信息
	if uploader, err := s.userRepo.GetByID(ctx, attachment.UploadedBy); err == nil {
		resp.Uploader = &dto.UserBrief{
			ID:          uploader.ID,
			Username:    uploader.Username,
			DisplayName: uploader.DisplayName,
			AvatarURL:   uploader.AvatarURL,
		}
	}

	return resp
}
