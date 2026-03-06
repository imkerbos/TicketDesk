// Package service 提供附件业务逻辑层
package service

import (
	"context"
	"errors"
	"fmt"
	"mime/multipart"
	"path/filepath"
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

// AttachmentService 附件服务接口
type AttachmentService interface {
	UploadAttachment(ctx context.Context, issueKey string, file *multipart.FileHeader, userID uint64) (*dto.AttachmentResponse, error)
	ListAttachments(ctx context.Context, issueKey string) ([]*dto.AttachmentResponse, error)
	DeleteAttachment(ctx context.Context, issueKey string, attachmentID, userID uint64) error
	GetAttachmentPath(ctx context.Context, attachmentID uint64) (string, error)
	SetActivityLogger(activityLogger ActivityLogger)
}

// attachmentService 附件服务实现
type attachmentService struct {
	attachmentRepo repository.AttachmentRepository
	issueRepo      repository.IssueRepository
	userRepo       userRepo.UserRepository
	storage        *storage.LocalStorage
	maxFileSize    int64
	allowedTypes   []string
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
		maxFileSize:    10 * 1024 * 1024, // 10MB
		allowedTypes: []string{
			".jpg", ".jpeg", ".png", ".gif", ".bmp", ".webp", // 图片
			".pdf", ".doc", ".docx", ".xls", ".xlsx", ".ppt", ".pptx", // 文档
			".txt", ".md", ".csv", // 文本
			".zip", ".rar", ".7z", ".tar", ".gz", // 压缩包
			".log", ".json", ".xml", ".yaml", ".yml", // 配置文件
		},
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

	// 验证文件大小
	if fileHeader.Size > s.maxFileSize {
		return nil, ErrFileTooLarge
	}

	// 验证文件类型
	ext := strings.ToLower(filepath.Ext(fileHeader.Filename))
	if !s.isAllowedType(ext) {
		return nil, ErrInvalidFileType
	}

	// 打开文件
	file, err := fileHeader.Open()
	if err != nil {
		return nil, fmt.Errorf("打开文件失败: %w", err)
	}
	defer file.Close()

	// 生成唯一文件名（时间戳 + 原文件名）
	timestamp := time.Now().Format("20060102150405")
	uniqueFilename := fmt.Sprintf("%s_%s", timestamp, fileHeader.Filename)

	// 保存文件
	relPath, err := s.storage.Save(file, uniqueFilename)
	if err != nil {
		logger.Error("failed to save file", zap.Error(err))
		return nil, fmt.Errorf("保存文件失败: %w", err)
	}

	// 判断是否为图片
	isImage := s.isImageType(ext)

	// 创建附件记录
	attachment := &model.IssueAttachment{
		IssueID:    issue.ID,
		FileName:   fileHeader.Filename,
		FilePath:   relPath,
		FileSize:   fileHeader.Size,
		FileType:   ext,
		IsImage:    isImage,
		UploadedBy: userID,
	}

	if err := s.attachmentRepo.Create(ctx, attachment); err != nil {
		// 删除已保存的文件
		s.storage.Delete(relPath)
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
		s.activityLogger.LogActivity(
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
		s.activityLogger.LogActivity(
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

// isAllowedType 检查文件类型是否允许
func (s *attachmentService) isAllowedType(ext string) bool {
	for _, allowed := range s.allowedTypes {
		if ext == allowed {
			return true
		}
	}
	return false
}

// isImageType 判断是否为图片类型
func (s *attachmentService) isImageType(ext string) bool {
	imageTypes := []string{".jpg", ".jpeg", ".png", ".gif", ".bmp", ".webp"}
	for _, imgType := range imageTypes {
		if ext == imgType {
			return true
		}
	}
	return false
}
