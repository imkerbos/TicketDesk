package handler

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/kerbos/ticketdesk/internal/api/response"
	"github.com/kerbos/ticketdesk/internal/requirement-pool/dto"
	"github.com/kerbos/ticketdesk/internal/requirement-pool/service"
	"go.uber.org/zap"
)

// RequirementHandler 需求 HTTP 处理器
type RequirementHandler struct {
	service service.RequirementService
	logger  *zap.Logger
}

// NewRequirementHandler 创建需求 HTTP 处理器实例
func NewRequirementHandler(
	service service.RequirementService,
	logger *zap.Logger,
) *RequirementHandler {
	return &RequirementHandler{
		service: service,
		logger:  logger,
	}
}

// Create 创建需求
// @Summary 创建需求
// @Description 创建一个新的需求
// @Tags Requirement
// @Accept json
// @Produce json
// @Param request body dto.CreateRequirementRequest true "创建需求请求"
// @Success 201 {object} response.Response{data=dto.RequirementResponse}
// @Failure 400 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/v1/requirements [post]
// @Security BearerAuth
func (h *RequirementHandler) Create(c *gin.Context) {
	var req dto.CreateRequirementRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数错误: "+err.Error())
		return
	}

	userID := h.getUserID(c)

	requirement, err := h.service.Create(c.Request.Context(), &req, userID)
	if err != nil {
		h.logger.Error("failed to create requirement",
			zap.Error(err),
			zap.Uint64("user_id", userID),
		)
		response.InternalError(c, err.Error())
		return
	}

	response.Created(c, requirement)
}

// GetByID 根据ID获取需求
// @Summary 获取需求详情
// @Description 根据ID获取需求详情
// @Tags Requirement
// @Accept json
// @Produce json
// @Param id path int true "需求ID"
// @Success 200 {object} response.Response{data=dto.RequirementResponse}
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/v1/requirements/{id} [get]
// @Security BearerAuth
func (h *RequirementHandler) GetByID(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的需求ID")
		return
	}

	requirement, err := h.service.GetByID(c.Request.Context(), id)
	if err != nil {
		if err.Error() == "需求不存在" {
			response.NotFound(c, err.Error())
			return
		}
		h.logger.Error("failed to get requirement",
			zap.Error(err),
			zap.Uint64("requirement_id", id),
		)
		response.InternalError(c, err.Error())
		return
	}

	response.Success(c, requirement)
}

// Update 更新需求
// @Summary 更新需求
// @Description 更新需求信息
// @Tags Requirement
// @Accept json
// @Produce json
// @Param id path int true "需求ID"
// @Param request body dto.UpdateRequirementRequest true "更新需求请求"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/v1/requirements/{id} [put]
// @Security BearerAuth
func (h *RequirementHandler) Update(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的需求ID")
		return
	}

	var req dto.UpdateRequirementRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数错误: "+err.Error())
		return
	}

	userID := h.getUserID(c)

	if err := h.service.Update(c.Request.Context(), id, &req, userID); err != nil {
		if err.Error() == "需求不存在" {
			response.NotFound(c, err.Error())
			return
		}
		h.logger.Error("failed to update requirement",
			zap.Error(err),
			zap.Uint64("requirement_id", id),
			zap.Uint64("user_id", userID),
		)
		response.InternalError(c, err.Error())
		return
	}

	response.Success(c, gin.H{"message": "更新成功"})
}

// Delete 删除需求
// @Summary 删除需求
// @Description 删除需求（软删除）
// @Tags Requirement
// @Accept json
// @Produce json
// @Param id path int true "需求ID"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/v1/requirements/{id} [delete]
// @Security BearerAuth
func (h *RequirementHandler) Delete(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的需求ID")
		return
	}

	userID := h.getUserID(c)

	if err := h.service.Delete(c.Request.Context(), id, userID); err != nil {
		if err.Error() == "需求不存在" {
			response.NotFound(c, err.Error())
			return
		}
		h.logger.Error("failed to delete requirement",
			zap.Error(err),
			zap.Uint64("requirement_id", id),
			zap.Uint64("user_id", userID),
		)
		response.InternalError(c, err.Error())
		return
	}

	response.Success(c, gin.H{"message": "删除成功"})
}

// List 获取需求列表
// @Summary 获取需求列表
// @Description 获取需求列表（支持分页和筛选）
// @Tags Requirement
// @Accept json
// @Produce json
// @Param pool_id query int false "需求池ID"
// @Param status query string false "状态"
// @Param priority query string false "优先级"
// @Param assignee_id query int false "负责人ID"
// @Param created_by query int false "创建人ID"
// @Param start_date query string false "开始日期"
// @Param end_date query string false "结束日期"
// @Param keyword query string false "关键词"
// @Param page query int false "页码" default(1)
// @Param page_size query int false "每页数量" default(20)
// @Success 200 {object} response.Response{data=response.PageData}
// @Failure 400 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/v1/requirements [get]
// @Security BearerAuth
func (h *RequirementHandler) List(c *gin.Context) {
	var req dto.RequirementListRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.BadRequest(c, "请求参数错误: "+err.Error())
		return
	}

	requirements, total, err := h.service.List(c.Request.Context(), &req)
	if err != nil {
		h.logger.Error("failed to list requirements",
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

	response.SuccessWithPage(c, requirements, total, page, pageSize)
}

// ConvertToIssue 转化为工单
// @Summary 转化为工单
// @Description 将需求转化为项目工单
// @Tags Requirement
// @Accept json
// @Produce json
// @Param id path int true "需求ID"
// @Param request body dto.ConvertToIssueRequest true "转化为工单请求"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/v1/requirements/{id}/convert [post]
// @Security BearerAuth
func (h *RequirementHandler) ConvertToIssue(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的需求ID")
		return
	}

	var req dto.ConvertToIssueRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数错误: "+err.Error())
		return
	}

	userID := h.getUserID(c)

	result, err := h.service.ConvertToIssue(c.Request.Context(), id, &req, userID)
	if err != nil {
		errMsg := err.Error()
		switch errMsg {
		case "需求不存在":
			response.NotFound(c, errMsg)
		case "该需求已转化为工单，不允许重复转化",
			"已完成的需求不能转化为工单",
			"已拒绝的需求不能转化为工单",
			"工单创建服务未初始化":
			response.BadRequest(c, errMsg)
		default:
			h.logger.Error("failed to convert requirement to issue",
				zap.Error(err),
				zap.Uint64("requirement_id", id),
				zap.Uint64("user_id", userID),
			)
			response.InternalError(c, errMsg)
		}
		return
	}

	response.Success(c, gin.H{
		"message":   "转化成功",
		"issue_id":  result.IssueID,
		"issue_key": result.IssueKey,
	})
}

// AddComment 添加评论
// @Summary 添加评论
// @Description 为需求添加评论
// @Tags Requirement
// @Accept json
// @Produce json
// @Param id path int true "需求ID"
// @Param request body dto.RequirementCommentRequest true "评论请求"
// @Success 201 {object} response.Response{data=dto.RequirementCommentResponse}
// @Failure 400 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/v1/requirements/{id}/comments [post]
// @Security BearerAuth
func (h *RequirementHandler) AddComment(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的需求ID")
		return
	}

	var req dto.RequirementCommentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数错误: "+err.Error())
		return
	}

	userID := h.getUserID(c)

	comment, err := h.service.AddComment(c.Request.Context(), id, &req, userID)
	if err != nil {
		if err.Error() == "需求不存在" {
			response.NotFound(c, err.Error())
			return
		}
		h.logger.Error("failed to add comment",
			zap.Error(err),
			zap.Uint64("requirement_id", id),
			zap.Uint64("user_id", userID),
		)
		response.InternalError(c, err.Error())
		return
	}

	response.Created(c, comment)
}

// GetKanban 获取看板数据
// @Summary 获取看板数据
// @Description 获取需求看板数据（支持多维度分组）
// @Tags Requirement
// @Accept json
// @Produce json
// @Param pool_id query int false "需求池ID"
// @Param group_by query string true "分组方式" Enums(status, priority, assignee, timeline)
// @Param start_date query string false "开始日期"
// @Param end_date query string false "结束日期"
// @Success 200 {object} response.Response{data=dto.KanbanResponse}
// @Failure 400 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/v1/requirements/kanban [get]
// @Security BearerAuth
func (h *RequirementHandler) GetKanban(c *gin.Context) {
	var req dto.KanbanRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.BadRequest(c, "请求参数错误: "+err.Error())
		return
	}

	kanban, err := h.service.GetKanban(c.Request.Context(), &req)
	if err != nil {
		h.logger.Error("failed to get kanban",
			zap.Error(err),
		)
		response.InternalError(c, err.Error())
		return
	}

	response.Success(c, kanban)
}

// GetReport 获取报告数据
// @Summary 获取报告数据
// @Description 获取需求统计报告
// @Tags Requirement
// @Accept json
// @Produce json
// @Param pool_id query int false "需求池ID"
// @Param start_date query string true "开始日期"
// @Param end_date query string true "结束日期"
// @Param group_by query string false "分组方式" Enums(day, week, month)
// @Success 200 {object} response.Response{data=dto.ReportResponse}
// @Failure 400 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/v1/requirements/report [get]
// @Security BearerAuth
func (h *RequirementHandler) GetReport(c *gin.Context) {
	var req dto.ReportRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.BadRequest(c, "请求参数错误: "+err.Error())
		return
	}

	report, err := h.service.GetReport(c.Request.Context(), &req)
	if err != nil {
		h.logger.Error("failed to get report",
			zap.Error(err),
		)
		response.InternalError(c, err.Error())
		return
	}

	response.Success(c, report)
}

// getUserID 从上下文获取用户ID
func (h *RequirementHandler) getUserID(c *gin.Context) uint64 {
	// 从 JWT 中间件设置的上下文中获取用户ID
	if userID, exists := c.Get("user_id"); exists {
		if id, ok := userID.(uint64); ok {
			return id
		}
	}
	return 0
}
