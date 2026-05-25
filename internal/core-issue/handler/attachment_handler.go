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
// @Summary 上传附件
// @Description 为指定工单上传附件文件，最大 10MB
// @Tags Attachment
// @Accept multipart/form-data
// @Produce json
// @Security BearerAuth
// @Param key path string true "工单 Key"
// @Param file formData file true "上传文件"
// @Success 201 {object} response.Response{data=dto.AttachmentResponse} "上传成功"
// @Failure 400 {object} response.ErrorResponse "参数错误"
// @Failure 401 {object} response.ErrorResponse "未认证"
// @Failure 404 {object} response.ErrorResponse "工单不存在"
// @Router /api/v1/issues/{key}/attachments [post]
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
// @Summary 列出工单附件
// @Tags Attachment
// @Produce json
// @Security BearerAuth
// @Param key path string true "工单 Key"
// @Success 200 {object} response.Response{data=[]dto.AttachmentResponse} "获取成功"
// @Failure 401 {object} response.ErrorResponse "未认证"
// @Failure 404 {object} response.ErrorResponse "工单不存在"
// @Router /api/v1/issues/{key}/attachments [get]
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
// @Summary 删除附件
// @Tags Attachment
// @Produce json
// @Security BearerAuth
// @Param key path string true "工单 Key"
// @Param id path int true "附件 ID"
// @Success 200 {object} response.Response "删除成功"
// @Failure 400 {object} response.ErrorResponse "无效的附件 ID"
// @Failure 401 {object} response.ErrorResponse "未认证"
// @Failure 403 {object} response.ErrorResponse "无权限"
// @Failure 404 {object} response.ErrorResponse "附件不存在"
// @Router /api/v1/issues/{key}/attachments/{id} [delete]
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
// @Summary 下载附件
// @Tags Attachment
// @Produce application/octet-stream
// @Security BearerAuth
// @Param key path string true "工单 Key"
// @Param id path int true "附件 ID"
// @Success 200 {file} binary "文件内容"
// @Failure 400 {object} response.ErrorResponse "无效的附件 ID"
// @Failure 401 {object} response.ErrorResponse "未认证"
// @Failure 404 {object} response.ErrorResponse "附件不存在"
// @Router /api/v1/issues/{key}/attachments/{id}/download [get]
func (h *AttachmentHandler) HandleDownloadAttachment(c *gin.Context) {
	issueKey := c.Param("key")
	attachmentIDStr := c.Param("id")

	attachmentID, err := strconv.ParseUint(attachmentIDStr, 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的附件ID")
		return
	}

	// 校验附件归属于该工单, 防 IDOR (别的工单的附件 ID 不应能下到)
	filePath, err := h.attachmentService.GetAttachmentPathForIssue(c.Request.Context(), issueKey, attachmentID)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrIssueNotFound):
			response.NotFound(c, "工单不存在")
		case errors.Is(err, service.ErrAttachmentNotFound):
			response.NotFound(c, "附件不存在")
		default:
			response.InternalError(c, "获取附件失败")
		}
		return
	}

	// 直接返回文件
	c.File(filePath)
}
