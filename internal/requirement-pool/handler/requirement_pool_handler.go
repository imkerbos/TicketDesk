package handler

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/kerbos/ticketdesk/internal/api/response"
	"github.com/kerbos/ticketdesk/internal/requirement-pool/dto"
	"github.com/kerbos/ticketdesk/internal/requirement-pool/service"
)

// RequirementPoolHandler 需求池 HTTP 处理器
type RequirementPoolHandler struct {
	service service.RequirementPoolService
	logger  *zap.Logger
}

// NewRequirementPoolHandler 创建需求池 HTTP 处理器实例
func NewRequirementPoolHandler(
	service service.RequirementPoolService,
	logger *zap.Logger,
) *RequirementPoolHandler {
	return &RequirementPoolHandler{
		service: service,
		logger:  logger,
	}
}

// Create 创建需求池
// @Summary 创建需求池
// @Description 创建一个新的需求池
// @Tags RequirementPool
// @Accept json
// @Produce json
// @Param request body dto.CreateRequirementPoolRequest true "创建需求池请求"
// @Success 201 {object} response.Response{data=dto.RequirementPoolResponse}
// @Failure 400 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/v1/requirement-pools [post]
// @Security BearerAuth
func (h *RequirementPoolHandler) Create(c *gin.Context) {
	var req dto.CreateRequirementPoolRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数错误: "+err.Error())
		return
	}

	// 从上下文获取用户ID（假设已通过认证中间件设置）
	userID := h.getUserID(c)

	pool, err := h.service.Create(c.Request.Context(), &req, userID)
	if err != nil {
		h.logger.Error("failed to create requirement pool",
			zap.Error(err),
			zap.Uint64("user_id", userID),
		)
		response.InternalError(c, err.Error())
		return
	}

	response.Created(c, pool)
}

// GetByID 根据ID获取需求池
// @Summary 获取需求池详情
// @Description 根据ID获取需求池详情
// @Tags RequirementPool
// @Accept json
// @Produce json
// @Param id path int true "需求池ID"
// @Success 200 {object} response.Response{data=dto.RequirementPoolResponse}
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/v1/requirement-pools/{id} [get]
// @Security BearerAuth
func (h *RequirementPoolHandler) GetByID(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的需求池ID")
		return
	}

	pool, err := h.service.GetByID(c.Request.Context(), id)
	if err != nil {
		if err.Error() == "需求池不存在" {
			response.NotFound(c, err.Error())
			return
		}
		h.logger.Error("failed to get requirement pool",
			zap.Error(err),
			zap.Uint64("pool_id", id),
		)
		response.InternalError(c, err.Error())
		return
	}

	response.Success(c, pool)
}

// Update 更新需求池
// @Summary 更新需求池
// @Description 更新需求池信息
// @Tags RequirementPool
// @Accept json
// @Produce json
// @Param id path int true "需求池ID"
// @Param request body dto.UpdateRequirementPoolRequest true "更新需求池请求"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/v1/requirement-pools/{id} [put]
// @Security BearerAuth
func (h *RequirementPoolHandler) Update(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的需求池ID")
		return
	}

	var req dto.UpdateRequirementPoolRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数错误: "+err.Error())
		return
	}

	userID := h.getUserID(c)

	if err := h.service.Update(c.Request.Context(), id, &req, userID); err != nil {
		if err.Error() == "需求池不存在" {
			response.NotFound(c, err.Error())
			return
		}
		h.logger.Error("failed to update requirement pool",
			zap.Error(err),
			zap.Uint64("pool_id", id),
			zap.Uint64("user_id", userID),
		)
		response.InternalError(c, err.Error())
		return
	}

	response.Success(c, gin.H{"message": "更新成功"})
}

// Delete 删除需求池
// @Summary 删除需求池
// @Description 删除需求池（软删除）
// @Tags RequirementPool
// @Accept json
// @Produce json
// @Param id path int true "需求池ID"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/v1/requirement-pools/{id} [delete]
// @Security BearerAuth
func (h *RequirementPoolHandler) Delete(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的需求池ID")
		return
	}

	userID := h.getUserID(c)

	if err := h.service.Delete(c.Request.Context(), id, userID); err != nil {
		if err.Error() == "需求池不存在" {
			response.NotFound(c, err.Error())
			return
		}
		h.logger.Error("failed to delete requirement pool",
			zap.Error(err),
			zap.Uint64("pool_id", id),
			zap.Uint64("user_id", userID),
		)
		response.InternalError(c, err.Error())
		return
	}

	response.Success(c, gin.H{"message": "删除成功"})
}

// List 获取需求池列表
// @Summary 获取需求池列表
// @Description 获取需求池列表（支持分页和筛选）
// @Tags RequirementPool
// @Accept json
// @Produce json
// @Param type query string false "类型" Enums(global, project)
// @Param status query string false "状态" Enums(active, archived)
// @Param project_id query int false "项目ID"
// @Param owner_id query int false "负责人ID"
// @Param page query int false "页码" default(1)
// @Param page_size query int false "每页数量" default(20)
// @Success 200 {object} response.Response{data=response.PageData}
// @Failure 400 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/v1/requirement-pools [get]
// @Security BearerAuth
func (h *RequirementPoolHandler) List(c *gin.Context) {
	var req dto.RequirementPoolListRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.BadRequest(c, "请求参数错误: "+err.Error())
		return
	}

	pools, total, err := h.service.List(c.Request.Context(), &req)
	if err != nil {
		h.logger.Error("failed to list requirement pools",
			zap.Error(err),
		)
		response.InternalError(c, err.Error())
		return
	}

	page := req.Page
	if page == 0 {
		page = 1
	}
	pageSize := req.PageSize
	if pageSize == 0 {
		pageSize = 20
	}

	response.SuccessWithPage(c, pools, total, page, pageSize)
}

// getUserID 从上下文获取用户ID
func (h *RequirementPoolHandler) getUserID(c *gin.Context) uint64 {
	// 从 JWT 中间件设置的上下文中获取用户ID
	if userID, exists := c.Get("user_id"); exists {
		if id, ok := userID.(uint64); ok {
			return id
		}
	}
	return 0
}
