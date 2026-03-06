// Package handler 提供字段HTTP处理器
package handler

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/kerbos/ticketdesk/internal/api/response"
	"github.com/kerbos/ticketdesk/internal/core-field/dto"
	"github.com/kerbos/ticketdesk/internal/core-field/service"
)

// FieldHandler 字段处理器
type FieldHandler struct {
	fieldService service.FieldService
}

// NewFieldHandler 创建字段处理器实例
func NewFieldHandler(fieldService service.FieldService) *FieldHandler {
	return &FieldHandler{
		fieldService: fieldService,
	}
}

// ============ 字段定义 ============

// CreateField 创建自定义字段
// @Summary 创建自定义字段
// @Description 为项目创建自定义字段
// @Tags Field
// @Accept json
// @Produce json
// @Param key path string true "项目Key"
// @Param request body dto.CreateFieldRequest true "创建字段请求"
// @Success 201 {object} response.Response{data=dto.FieldDefinitionResponse}
// @Failure 400 {object} response.Response
// @Router /api/v1/projects/{key}/fields [post]
// @Security BearerAuth
func (h *FieldHandler) CreateField(c *gin.Context) {
	projectKey := c.Param("key")

	var req dto.CreateFieldRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	result, err := h.fieldService.CreateField(c.Request.Context(), projectKey, &req)
	if err != nil {
		h.handleError(c, err)
		return
	}

	response.Created(c, result)
}

// UpdateField 更新字段定义
// @Summary 更新字段定义
// @Description 更新项目的字段定义
// @Tags Field
// @Accept json
// @Produce json
// @Param key path string true "项目Key"
// @Param id path int true "字段ID"
// @Param request body dto.UpdateFieldRequest true "更新字段请求"
// @Success 200 {object} response.Response{data=dto.FieldDefinitionResponse}
// @Failure 400 {object} response.Response
// @Router /api/v1/projects/{key}/fields/{id} [put]
// @Security BearerAuth
func (h *FieldHandler) UpdateField(c *gin.Context) {
	projectKey := c.Param("key")
	fieldID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的字段ID")
		return
	}

	var req dto.UpdateFieldRequest
	if bindErr := c.ShouldBindJSON(&req); bindErr != nil {
		response.BadRequest(c, bindErr.Error())
		return
	}

	result, err := h.fieldService.UpdateField(c.Request.Context(), projectKey, fieldID, &req)
	if err != nil {
		h.handleError(c, err)
		return
	}

	response.Success(c, result)
}

// DeleteField 删除字段定义
// @Summary 删除字段定义
// @Description 删除项目的自定义字段
// @Tags Field
// @Produce json
// @Param key path string true "项目Key"
// @Param id path int true "字段ID"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.Response
// @Router /api/v1/projects/{key}/fields/{id} [delete]
// @Security BearerAuth
func (h *FieldHandler) DeleteField(c *gin.Context) {
	projectKey := c.Param("key")
	fieldID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的字段ID")
		return
	}

	if err := h.fieldService.DeleteField(c.Request.Context(), projectKey, fieldID); err != nil {
		h.handleError(c, err)
		return
	}

	response.Success(c, nil)
}

// ListFields 获取字段列表
// @Summary 获取字段列表
// @Description 获取项目可用的所有字段（系统字段+自定义字段）
// @Tags Field
// @Produce json
// @Param key path string true "项目Key"
// @Success 200 {object} response.Response{data=[]dto.FieldDefinitionResponse}
// @Failure 400 {object} response.Response
// @Router /api/v1/projects/{key}/fields [get]
// @Security BearerAuth
func (h *FieldHandler) ListFields(c *gin.Context) {
	projectKey := c.Param("key")

	result, err := h.fieldService.ListFields(c.Request.Context(), projectKey)
	if err != nil {
		h.handleError(c, err)
		return
	}

	response.Success(c, result)
}

// ============ 字段方案 ============

// GetFieldScheme 获取工单类型的字段方案
// @Summary 获取字段方案
// @Description 获取工单类型的字段配置方案
// @Tags Field
// @Produce json
// @Param key path string true "项目Key"
// @Param id path int true "工单类型ID"
// @Success 200 {object} response.Response{data=[]dto.FieldSchemeResponse}
// @Failure 400 {object} response.Response
// @Router /api/v1/projects/{key}/issue-types/{id}/field-scheme [get]
// @Security BearerAuth
func (h *FieldHandler) GetFieldScheme(c *gin.Context) {
	projectKey := c.Param("key")
	issueTypeID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的工单类型ID")
		return
	}

	result, err := h.fieldService.GetFieldScheme(c.Request.Context(), projectKey, issueTypeID)
	if err != nil {
		h.handleError(c, err)
		return
	}

	response.Success(c, result)
}

// UpdateFieldScheme 更新工单类型的字段方案
// @Summary 更新字段方案
// @Description 更新工单类型的字段配置方案
// @Tags Field
// @Accept json
// @Produce json
// @Param key path string true "项目Key"
// @Param id path int true "工单类型ID"
// @Param request body dto.UpdateFieldSchemeRequest true "更新字段方案请求"
// @Success 200 {object} response.Response{data=[]dto.FieldSchemeResponse}
// @Failure 400 {object} response.Response
// @Router /api/v1/projects/{key}/issue-types/{id}/field-scheme [put]
// @Security BearerAuth
func (h *FieldHandler) UpdateFieldScheme(c *gin.Context) {
	projectKey := c.Param("key")
	issueTypeID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的工单类型ID")
		return
	}

	var req dto.UpdateFieldSchemeRequest
	if bindErr := c.ShouldBindJSON(&req); bindErr != nil {
		response.BadRequest(c, bindErr.Error())
		return
	}

	result, err := h.fieldService.UpdateFieldScheme(c.Request.Context(), projectKey, issueTypeID, &req)
	if err != nil {
		h.handleError(c, err)
		return
	}

	response.Success(c, result)
}

// ============ 版本管理 ============

// CreateVersion 创建项目版本
// @Summary 创建项目版本
// @Description 为项目创建新版本
// @Tags Version
// @Accept json
// @Produce json
// @Param key path string true "项目Key"
// @Param request body dto.CreateVersionRequest true "创建版本请求"
// @Success 201 {object} response.Response{data=dto.VersionResponse}
// @Failure 400 {object} response.Response
// @Router /api/v1/projects/{key}/versions [post]
// @Security BearerAuth
func (h *FieldHandler) CreateVersion(c *gin.Context) {
	projectKey := c.Param("key")

	var req dto.CreateVersionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	result, err := h.fieldService.CreateVersion(c.Request.Context(), projectKey, &req)
	if err != nil {
		h.handleError(c, err)
		return
	}

	response.Created(c, result)
}

// UpdateVersion 更新项目版本
// @Summary 更新项目版本
// @Description 更新项目的版本信息
// @Tags Version
// @Accept json
// @Produce json
// @Param key path string true "项目Key"
// @Param id path int true "版本ID"
// @Param request body dto.UpdateVersionRequest true "更新版本请求"
// @Success 200 {object} response.Response{data=dto.VersionResponse}
// @Failure 400 {object} response.Response
// @Router /api/v1/projects/{key}/versions/{id} [put]
// @Security BearerAuth
func (h *FieldHandler) UpdateVersion(c *gin.Context) {
	projectKey := c.Param("key")
	versionID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的版本ID")
		return
	}

	var req dto.UpdateVersionRequest
	if bindErr := c.ShouldBindJSON(&req); bindErr != nil {
		response.BadRequest(c, bindErr.Error())
		return
	}

	result, err := h.fieldService.UpdateVersion(c.Request.Context(), projectKey, versionID, &req)
	if err != nil {
		h.handleError(c, err)
		return
	}

	response.Success(c, result)
}

// DeleteVersion 删除项目版本
// @Summary 删除项目版本
// @Description 删除项目的版本
// @Tags Version
// @Produce json
// @Param key path string true "项目Key"
// @Param id path int true "版本ID"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.Response
// @Router /api/v1/projects/{key}/versions/{id} [delete]
// @Security BearerAuth
func (h *FieldHandler) DeleteVersion(c *gin.Context) {
	projectKey := c.Param("key")
	versionID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的版本ID")
		return
	}

	if err := h.fieldService.DeleteVersion(c.Request.Context(), projectKey, versionID); err != nil {
		h.handleError(c, err)
		return
	}

	response.Success(c, nil)
}

// ListVersions 获取项目版本列表
// @Summary 获取版本列表
// @Description 获取项目的所有版本
// @Tags Version
// @Produce json
// @Param key path string true "项目Key"
// @Success 200 {object} response.Response{data=[]dto.VersionResponse}
// @Failure 400 {object} response.Response
// @Router /api/v1/projects/{key}/versions [get]
// @Security BearerAuth
func (h *FieldHandler) ListVersions(c *gin.Context) {
	projectKey := c.Param("key")

	result, err := h.fieldService.ListVersions(c.Request.Context(), projectKey)
	if err != nil {
		h.handleError(c, err)
		return
	}

	response.Success(c, result)
}

// ============ 组件管理 ============

// CreateComponent 创建项目组件
// @Summary 创建项目组件
// @Description 为项目创建新组件
// @Tags Component
// @Accept json
// @Produce json
// @Param key path string true "项目Key"
// @Param request body dto.CreateComponentRequest true "创建组件请求"
// @Success 201 {object} response.Response{data=dto.ComponentResponse}
// @Failure 400 {object} response.Response
// @Router /api/v1/projects/{key}/components [post]
// @Security BearerAuth
func (h *FieldHandler) CreateComponent(c *gin.Context) {
	projectKey := c.Param("key")

	var req dto.CreateComponentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	result, err := h.fieldService.CreateComponent(c.Request.Context(), projectKey, &req)
	if err != nil {
		h.handleError(c, err)
		return
	}

	response.Created(c, result)
}

// UpdateComponent 更新项目组件
// @Summary 更新项目组件
// @Description 更新项目的组件信息
// @Tags Component
// @Accept json
// @Produce json
// @Param key path string true "项目Key"
// @Param id path int true "组件ID"
// @Param request body dto.UpdateComponentRequest true "更新组件请求"
// @Success 200 {object} response.Response{data=dto.ComponentResponse}
// @Failure 400 {object} response.Response
// @Router /api/v1/projects/{key}/components/{id} [put]
// @Security BearerAuth
func (h *FieldHandler) UpdateComponent(c *gin.Context) {
	projectKey := c.Param("key")
	componentID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的组件ID")
		return
	}

	var req dto.UpdateComponentRequest
	if bindErr := c.ShouldBindJSON(&req); bindErr != nil {
		response.BadRequest(c, bindErr.Error())
		return
	}

	result, err := h.fieldService.UpdateComponent(c.Request.Context(), projectKey, componentID, &req)
	if err != nil {
		h.handleError(c, err)
		return
	}

	response.Success(c, result)
}

// DeleteComponent 删除项目组件
// @Summary 删除项目组件
// @Description 删除项目的组件
// @Tags Component
// @Produce json
// @Param key path string true "项目Key"
// @Param id path int true "组件ID"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.Response
// @Router /api/v1/projects/{key}/components/{id} [delete]
// @Security BearerAuth
func (h *FieldHandler) DeleteComponent(c *gin.Context) {
	projectKey := c.Param("key")
	componentID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的组件ID")
		return
	}

	if err := h.fieldService.DeleteComponent(c.Request.Context(), projectKey, componentID); err != nil {
		h.handleError(c, err)
		return
	}

	response.Success(c, nil)
}

// ListComponents 获取项目组件列表
// @Summary 获取组件列表
// @Description 获取项目的所有组件
// @Tags Component
// @Produce json
// @Param key path string true "项目Key"
// @Success 200 {object} response.Response{data=[]dto.ComponentResponse}
// @Failure 400 {object} response.Response
// @Router /api/v1/projects/{key}/components [get]
// @Security BearerAuth
func (h *FieldHandler) ListComponents(c *gin.Context) {
	projectKey := c.Param("key")

	result, err := h.fieldService.ListComponents(c.Request.Context(), projectKey)
	if err != nil {
		h.handleError(c, err)
		return
	}

	response.Success(c, result)
}

// ============ 标签管理 ============

// CreateLabel 创建标签
// @Summary 创建标签
// @Description 为项目创建新标签
// @Tags Label
// @Accept json
// @Produce json
// @Param key path string true "项目Key"
// @Param request body dto.CreateLabelRequest true "创建标签请求"
// @Success 201 {object} response.Response{data=dto.LabelResponse}
// @Failure 400 {object} response.Response
// @Router /api/v1/projects/{key}/labels [post]
// @Security BearerAuth
func (h *FieldHandler) CreateLabel(c *gin.Context) {
	projectKey := c.Param("key")

	var req dto.CreateLabelRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	result, err := h.fieldService.CreateLabel(c.Request.Context(), projectKey, &req)
	if err != nil {
		h.handleError(c, err)
		return
	}

	response.Created(c, result)
}

// UpdateLabel 更新标签
// @Summary 更新标签
// @Description 更新项目的标签信息
// @Tags Label
// @Accept json
// @Produce json
// @Param key path string true "项目Key"
// @Param id path int true "标签ID"
// @Param request body dto.UpdateLabelRequest true "更新标签请求"
// @Success 200 {object} response.Response{data=dto.LabelResponse}
// @Failure 400 {object} response.Response
// @Router /api/v1/projects/{key}/labels/{id} [put]
// @Security BearerAuth
func (h *FieldHandler) UpdateLabel(c *gin.Context) {
	projectKey := c.Param("key")
	labelID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的标签ID")
		return
	}

	var req dto.UpdateLabelRequest
	if bindErr := c.ShouldBindJSON(&req); bindErr != nil {
		response.BadRequest(c, bindErr.Error())
		return
	}

	result, err := h.fieldService.UpdateLabel(c.Request.Context(), projectKey, labelID, &req)
	if err != nil {
		h.handleError(c, err)
		return
	}

	response.Success(c, result)
}

// DeleteLabel 删除标签
// @Summary 删除标签
// @Description 删除项目的标签
// @Tags Label
// @Produce json
// @Param key path string true "项目Key"
// @Param id path int true "标签ID"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.Response
// @Router /api/v1/projects/{key}/labels/{id} [delete]
// @Security BearerAuth
func (h *FieldHandler) DeleteLabel(c *gin.Context) {
	projectKey := c.Param("key")
	labelID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的标签ID")
		return
	}

	if err := h.fieldService.DeleteLabel(c.Request.Context(), projectKey, labelID); err != nil {
		h.handleError(c, err)
		return
	}

	response.Success(c, nil)
}

// ListLabels 获取标签列表
// @Summary 获取标签列表
// @Description 获取项目的所有标签
// @Tags Label
// @Produce json
// @Param key path string true "项目Key"
// @Success 200 {object} response.Response{data=[]dto.LabelResponse}
// @Failure 400 {object} response.Response
// @Router /api/v1/projects/{key}/labels [get]
// @Security BearerAuth
func (h *FieldHandler) ListLabels(c *gin.Context) {
	projectKey := c.Param("key")

	result, err := h.fieldService.ListLabels(c.Request.Context(), projectKey)
	if err != nil {
		h.handleError(c, err)
		return
	}

	response.Success(c, result)
}

// ============ 字段值 ============

// GetIssueFieldValues 获取工单的字段值
// @Summary 获取工单字段值
// @Description 获取指定工单的所有自定义字段值
// @Tags FieldValue
// @Produce json
// @Param issue_id path int true "工单ID"
// @Success 200 {object} response.Response{data=[]dto.FieldValueResponse}
// @Failure 400 {object} response.Response
// @Router /api/v1/issues/{issue_id}/field-values [get]
// @Security BearerAuth
func (h *FieldHandler) GetIssueFieldValues(c *gin.Context) {
	issueID, err := strconv.ParseUint(c.Param("issue_id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的工单ID")
		return
	}

	result, err := h.fieldService.GetFieldValues(c.Request.Context(), issueID)
	if err != nil {
		h.handleError(c, err)
		return
	}

	response.Success(c, result)
}

// ============ 全局字段管理 ============

// ListGlobalFields 获取全局字段列表
func (h *FieldHandler) ListGlobalFields(c *gin.Context) {
	result, err := h.fieldService.ListGlobalFields(c.Request.Context())
	if err != nil {
		h.handleError(c, err)
		return
	}
	response.Success(c, result)
}

// CreateGlobalField 创建全局自定义字段
func (h *FieldHandler) CreateGlobalField(c *gin.Context) {
	var req dto.CreateFieldRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	result, err := h.fieldService.CreateGlobalField(c.Request.Context(), &req)
	if err != nil {
		h.handleError(c, err)
		return
	}
	response.Created(c, result)
}

// UpdateGlobalField 更新全局字段
func (h *FieldHandler) UpdateGlobalField(c *gin.Context) {
	fieldID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的字段ID")
		return
	}
	var req dto.UpdateFieldRequest
	if bindErr := c.ShouldBindJSON(&req); bindErr != nil {
		response.BadRequest(c, bindErr.Error())
		return
	}
	result, err := h.fieldService.UpdateGlobalField(c.Request.Context(), fieldID, &req)
	if err != nil {
		h.handleError(c, err)
		return
	}
	response.Success(c, result)
}

// DeleteGlobalField 删除全局字段
func (h *FieldHandler) DeleteGlobalField(c *gin.Context) {
	fieldID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的字段ID")
		return
	}
	if err := h.fieldService.DeleteGlobalField(c.Request.Context(), fieldID); err != nil {
		h.handleError(c, err)
		return
	}
	response.Success(c, nil)
}

// GetFieldUsage 获取字段使用情况
func (h *FieldHandler) GetFieldUsage(c *gin.Context) {
	fieldID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的字段ID")
		return
	}
	result, err := h.fieldService.GetFieldUsage(c.Request.Context(), fieldID)
	if err != nil {
		h.handleError(c, err)
		return
	}
	response.Success(c, result)
}

// ============ 方案模板 ============

// ListTemplates 获取所有方案模板
func (h *FieldHandler) ListTemplates(c *gin.Context) {
	result, err := h.fieldService.ListTemplates(c.Request.Context())
	if err != nil {
		h.handleError(c, err)
		return
	}
	response.Success(c, result)
}

// CreateTemplate 创建方案模板
func (h *FieldHandler) CreateTemplate(c *gin.Context) {
	var req dto.CreateTemplateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	userID := c.GetUint64("user_id")
	result, err := h.fieldService.CreateTemplate(c.Request.Context(), userID, &req)
	if err != nil {
		h.handleError(c, err)
		return
	}
	response.Created(c, result)
}

// GetTemplate 获取模板详情
func (h *FieldHandler) GetTemplate(c *gin.Context) {
	templateID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的模板ID")
		return
	}
	result, err := h.fieldService.GetTemplate(c.Request.Context(), templateID)
	if err != nil {
		h.handleError(c, err)
		return
	}
	response.Success(c, result)
}

// UpdateTemplate 更新模板
func (h *FieldHandler) UpdateTemplate(c *gin.Context) {
	templateID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的模板ID")
		return
	}
	var req dto.UpdateTemplateRequest
	if bindErr := c.ShouldBindJSON(&req); bindErr != nil {
		response.BadRequest(c, bindErr.Error())
		return
	}
	result, err := h.fieldService.UpdateTemplate(c.Request.Context(), templateID, &req)
	if err != nil {
		h.handleError(c, err)
		return
	}
	response.Success(c, result)
}

// DeleteTemplate 删除模板
func (h *FieldHandler) DeleteTemplate(c *gin.Context) {
	templateID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的模板ID")
		return
	}
	if err := h.fieldService.DeleteTemplate(c.Request.Context(), templateID); err != nil {
		h.handleError(c, err)
		return
	}
	response.Success(c, nil)
}

// UpdateTemplateItems 更新模板字段项
func (h *FieldHandler) UpdateTemplateItems(c *gin.Context) {
	templateID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的模板ID")
		return
	}
	var req dto.UpdateTemplateItemsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	if err := h.fieldService.UpdateTemplateItems(c.Request.Context(), templateID, &req); err != nil {
		h.handleError(c, err)
		return
	}
	response.Success(c, nil)
}

// ApplyTemplate 套用模板到工单类型
func (h *FieldHandler) ApplyTemplate(c *gin.Context) {
	projectKey := c.Param("key")
	issueTypeID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的工单类型ID")
		return
	}
	var req dto.ApplyTemplateRequest
	if bindErr := c.ShouldBindJSON(&req); bindErr != nil {
		response.BadRequest(c, bindErr.Error())
		return
	}
	result, err := h.fieldService.ApplyTemplate(c.Request.Context(), projectKey, issueTypeID, &req)
	if err != nil {
		h.handleError(c, err)
		return
	}
	response.Success(c, result)
}

// handleError 统一错误处理
func (h *FieldHandler) handleError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, service.ErrProjectNotFound):
		response.NotFound(c, "项目不存在")
	case errors.Is(err, service.ErrFieldNotFound):
		response.NotFound(c, "字段不存在")
	case errors.Is(err, service.ErrFieldKeyExists):
		response.Error(c, http.StatusConflict, "CONFLICT", "字段Key已存在")
	case errors.Is(err, service.ErrCannotDeleteSystem):
		response.Forbidden(c, "不能删除系统字段")
	case errors.Is(err, service.ErrCannotModifySystem):
		response.Forbidden(c, "不能修改系统字段")
	case errors.Is(err, service.ErrIssueTypeNotFound):
		response.NotFound(c, "工单类型不存在")
	case errors.Is(err, service.ErrVersionNotFound):
		response.NotFound(c, "版本不存在")
	case errors.Is(err, service.ErrVersionNameExists):
		response.Error(c, http.StatusConflict, "CONFLICT", "版本名称已存在")
	case errors.Is(err, service.ErrComponentNotFound):
		response.NotFound(c, "组件不存在")
	case errors.Is(err, service.ErrComponentNameExists):
		response.Error(c, http.StatusConflict, "CONFLICT", "组件名称已存在")
	case errors.Is(err, service.ErrLabelNotFound):
		response.NotFound(c, "标签不存在")
	case errors.Is(err, service.ErrLabelNameExists):
		response.Error(c, http.StatusConflict, "CONFLICT", "标签名称已存在")
	case errors.Is(err, service.ErrTemplateNotFound):
		response.NotFound(c, "模板不存在")
	case errors.Is(err, service.ErrTemplateNameExists):
		response.Error(c, http.StatusConflict, "CONFLICT", "模板名称已存在")
	default:
		response.InternalError(c, err.Error())
	}
}
