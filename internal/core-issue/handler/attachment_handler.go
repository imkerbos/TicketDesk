// Package handler 提供附件模块的 HTTP 处理器
package handler

import (
	"errors"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/kerbos/ticketdesk/internal/api/response"
	"github.com/kerbos/ticketdesk/internal/core-issue/service"
)

// AttachmentHandler 附件处理器
type AttachmentHandler struct {
	attachmentService service.AttachmentService
}

// NewAttachmentHandler 创建附件处理器实例
func NewAttachmentHandler(attachmentService service.AttachmentService) *AttachmentHandler {
	return &AttachmentHandler{
		attachmentService: attachmentService,
	}
}

// HandleUploadAttachment 上传附件
func (h *AttachmentHandler) HandleUploadAttachment(c *gin.Context) {
	issueKey := c.Param("key")
	userID := c.GetUint64("user_id")

	// 获取上传的文件
	file, err := c.FormFile("file")
	if err != nil {
		response.BadRequest(c, "请选择要上传的文件")
		return
	}

	result, err := h.attachmentService.UploadAttachment(c.Request.Context(), issueKey, file, userID)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrIssueNotFound):
			response.NotFound(c, "工单不存在")
		case errors.Is(err, service.ErrFileTooLarge):
			response.BadRequest(c, "文件大小超过限制（最大10MB）")
		case errors.Is(err, service.ErrInvalidFileType):
			response.BadRequest(c, "不支持的文件类型")
		default:
			response.InternalError(c, "上传附件失败")
		}
		return
	}

	response.Created(c, result)
}

// HandleListAttachments 获取附件列表
func (h *AttachmentHandler) HandleListAttachments(c *gin.Context) {
	issueKey := c.Param("key")

	result, err := h.attachmentService.ListAttachments(c.Request.Context(), issueKey)
	if err != nil {
		if errors.Is(err, service.ErrIssueNotFound) {
			response.NotFound(c, "工单不存在")
			return
		}
		response.InternalError(c, "获取附件列表失败")
		return
	}

	response.Success(c, result)
}

// HandleDeleteAttachment 删除附件
func (h *AttachmentHandler) HandleDeleteAttachment(c *gin.Context) {
	issueKey := c.Param("key")
	attachmentIDStr := c.Param("id")
	userID := c.GetUint64("user_id")

	attachmentID, err := strconv.ParseUint(attachmentIDStr, 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的附件ID")
		return
	}

	err = h.attachmentService.DeleteAttachment(c.Request.Context(), issueKey, attachmentID, userID)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrIssueNotFound):
			response.NotFound(c, "工单不存在")
		case errors.Is(err, service.ErrAttachmentNotFound):
			response.NotFound(c, "附件不存在")
		case errors.Is(err, service.ErrUnauthorized):
			response.Forbidden(c, "无权限删除此附件")
		default:
			response.InternalError(c, "删除附件失败")
		}
		return
	}

	response.Success(c, nil)
}

// HandleDownloadAttachment 下载附件
func (h *AttachmentHandler) HandleDownloadAttachment(c *gin.Context) {
	attachmentIDStr := c.Param("id")

	attachmentID, err := strconv.ParseUint(attachmentIDStr, 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的附件ID")
		return
	}

	filePath, err := h.attachmentService.GetAttachmentPath(c.Request.Context(), attachmentID)
	if err != nil {
		if errors.Is(err, service.ErrAttachmentNotFound) {
			response.NotFound(c, "附件不存在")
			return
		}
		response.InternalError(c, "获取附件失败")
		return
	}

	// 直接返回文件
	c.File(filePath)
}
