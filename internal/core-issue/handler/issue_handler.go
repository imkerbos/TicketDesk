// Package handler 提供工单模块的 HTTP 处理器
package handler

import (
	"context"
	"errors"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/kerbos/ticketdesk/internal/api/response"
	"github.com/kerbos/ticketdesk/internal/core-issue/dto"
	"github.com/kerbos/ticketdesk/internal/core-issue/service"
)

// FieldServiceInterface 字段服务接口（避免循环依赖）
type FieldServiceInterface interface {
	GetIssueIDsByEpicLink(ctx context.Context, epicID uint64) ([]uint64, error)
}

// IssueHandler 工单处理器
type IssueHandler struct {
	issueService service.IssueService
	fieldService FieldServiceInterface
}

// NewIssueHandler 创建工单处理器实例
func NewIssueHandler(issueService service.IssueService) *IssueHandler {
	return &IssueHandler{
		issueService: issueService,
	}
}

// SetFieldService 设置字段服务（避免循环依赖）
func (h *IssueHandler) SetFieldService(fieldService FieldServiceInterface) {
	h.fieldService = fieldService
}

// ctxWithUser 将 Gin context 中的 user_id 注入到标准 Go context 中
func ctxWithUser(c *gin.Context) context.Context {
	ctx := c.Request.Context()
	userID := c.GetUint64("user_id")
	if userID > 0 {
		ctx = context.WithValue(ctx, service.ContextUserIDKey, userID)
	}
	return ctx
}

// HandleCreateIssue 创建工单
func (h *IssueHandler) HandleCreateIssue(c *gin.Context) {
	var req dto.CreateIssueRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数错误: "+err.Error())
		return
	}

	userID := c.GetUint64("user_id")
	result, err := h.issueService.CreateIssue(c.Request.Context(), &req, userID)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrProjectNotFound):
			response.NotFound(c, err.Error())
		case errors.Is(err, service.ErrIssueTypeNotFound):
			response.BadRequest(c, err.Error())
		default:
			response.InternalError(c, "创建工单失败")
		}
		return
	}

	response.Created(c, result)
}

// HandleGetIssue 获取工单详情
func (h *IssueHandler) HandleGetIssue(c *gin.Context) {
	key := c.Param("key")

	result, err := h.issueService.GetIssue(c.Request.Context(), key)
	if err != nil {
		if errors.Is(err, service.ErrIssueNotFound) {
			response.NotFound(c, err.Error())
			return
		}
		response.InternalError(c, "获取工单失败")
		return
	}

	response.Success(c, result)
}

// HandleUpdateIssue 更新工单
func (h *IssueHandler) HandleUpdateIssue(c *gin.Context) {
	key := c.Param("key")

	var req dto.UpdateIssueRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数错误: "+err.Error())
		return
	}

	result, err := h.issueService.UpdateIssue(ctxWithUser(c), key, &req)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrIssueNotFound):
			response.NotFound(c, err.Error())
		case errors.Is(err, service.ErrIssueTypeNotFound):
			response.BadRequest(c, err.Error())
		default:
			response.InternalError(c, "更新工单失败")
		}
		return
	}

	response.Success(c, result)
}

// HandleDeleteIssue 删除工单
func (h *IssueHandler) HandleDeleteIssue(c *gin.Context) {
	key := c.Param("key")

	err := h.issueService.DeleteIssue(ctxWithUser(c), key)
	if err != nil {
		if errors.Is(err, service.ErrIssueNotFound) {
			response.NotFound(c, err.Error())
			return
		}
		response.InternalError(c, "删除工单失败")
		return
	}

	response.Success(c, gin.H{"message": "工单删除成功"})
}

// HandleListIssues 获取工单列表
func (h *IssueHandler) HandleListIssues(c *gin.Context) {
	var req dto.ListIssuesRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.BadRequest(c, "请求参数错误: "+err.Error())
		return
	}

	// 从中间件注入的项目 ID 列表（非管理员且未指定 project_key 时）
	if projectIDs, exists := c.Get("user_project_ids"); exists {
		if ids, ok := projectIDs.([]uint64); ok {
			req.ProjectIDs = ids
		}
	}

	issues, total, hasMore, err := h.issueService.ListIssues(c.Request.Context(), &req)
	if err != nil {
		if errors.Is(err, service.ErrProjectNotFound) {
			response.NotFound(c, err.Error())
			return
		}
		response.InternalError(c, "获取工单列表失败")
		return
	}

	response.SuccessWithPageHasMore(c, issues, total, req.GetDefaultPage(), req.GetDefaultPageSize(), hasMore)
}

// HandleGetIssueListStats 获取工单列表统计数据
func (h *IssueHandler) HandleGetIssueListStats(c *gin.Context) {
	var req dto.ListIssuesRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.BadRequest(c, "请求参数错误: "+err.Error())
		return
	}

	// 从中间件注入的项目 ID 列表（非管理员且未指定 project_key 时）
	if projectIDs, exists := c.Get("user_project_ids"); exists {
		if ids, ok := projectIDs.([]uint64); ok {
			req.ProjectIDs = ids
		}
	}

	stats, err := h.issueService.GetIssueListStats(c.Request.Context(), &req)
	if err != nil {
		if errors.Is(err, service.ErrProjectNotFound) {
			response.NotFound(c, err.Error())
			return
		}
		response.InternalError(c, "获取工单统计失败")
		return
	}

	response.Success(c, stats)
}

// HandleListIssuesInEpic 获取 Epic 下的所有 Issues
func (h *IssueHandler) HandleListIssuesInEpic(c *gin.Context) {
	epicKey := c.Param("key")

	issues, err := h.issueService.ListIssuesInEpic(c.Request.Context(), epicKey)
	if err != nil {
		if errors.Is(err, service.ErrIssueNotFound) {
			response.NotFound(c, "Epic 不存在")
			return
		}
		response.InternalError(c, "获取 Epic 关联工单失败")
		return
	}

	response.Success(c, issues)
}

// HandleAssignIssue 指派工单
func (h *IssueHandler) HandleAssignIssue(c *gin.Context) {
	key := c.Param("key")

	var req dto.AssignIssueRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数错误: "+err.Error())
		return
	}

	result, err := h.issueService.AssignIssue(ctxWithUser(c), key, req.AssigneeID)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrIssueNotFound):
			response.NotFound(c, err.Error())
		case errors.Is(err, service.ErrUserNotFound):
			response.BadRequest(c, err.Error())
		default:
			response.InternalError(c, "指派工单失败")
		}
		return
	}

	response.Success(c, result)
}

// HandleListSubtasks 获取工单的所有子任务
func (h *IssueHandler) HandleListSubtasks(c *gin.Context) {
	parentKey := c.Param("key")

	subtasks, err := h.issueService.ListSubtasks(c.Request.Context(), parentKey)
	if err != nil {
		if errors.Is(err, service.ErrIssueNotFound) {
			response.NotFound(c, "父工单不存在")
			return
		}
		response.InternalError(c, "获取子任务失败")
		return
	}

	response.Success(c, subtasks)
}

// HandleAddComment 添加评论
func (h *IssueHandler) HandleAddComment(c *gin.Context) {
	key := c.Param("key")

	var req dto.CreateCommentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数错误: "+err.Error())
		return
	}

	userID := c.GetUint64("user_id")
	result, err := h.issueService.AddComment(c.Request.Context(), key, &req, userID)
	if err != nil {
		if errors.Is(err, service.ErrIssueNotFound) {
			response.NotFound(c, err.Error())
			return
		}
		response.InternalError(c, "添加评论失败")
		return
	}

	response.Created(c, result)
}

// HandleListComments 获取评论列表
func (h *IssueHandler) HandleListComments(c *gin.Context) {
	key := c.Param("key")

	comments, err := h.issueService.ListComments(c.Request.Context(), key)
	if err != nil {
		if errors.Is(err, service.ErrIssueNotFound) {
			response.NotFound(c, err.Error())
			return
		}
		response.InternalError(c, "获取评论列表失败")
		return
	}

	response.Success(c, comments)
}

// HandleDeleteComment 删除评论
func (h *IssueHandler) HandleDeleteComment(c *gin.Context) {
	commentID, err := strconv.ParseUint(c.Param("comment_id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的评论 ID")
		return
	}

	userID := c.GetUint64("user_id")
	err = h.issueService.DeleteComment(c.Request.Context(), commentID, userID)
	if err != nil {
		if errors.Is(err, service.ErrCommentNotFound) {
			response.NotFound(c, err.Error())
			return
		}
		response.BadRequest(c, err.Error())
		return
	}

	response.Success(c, gin.H{"message": "评论删除成功"})
}

// HandleAddWatcher 添加关注人
func (h *IssueHandler) HandleAddWatcher(c *gin.Context) {
	key := c.Param("key")

	var req dto.AddWatcherRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		// 如果没有传 user_id，默认关注自己
		userID := c.GetUint64("user_id")
		req.UserID = userID
	}

	err := h.issueService.AddWatcher(c.Request.Context(), key, req.UserID)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrIssueNotFound):
			response.NotFound(c, err.Error())
		case errors.Is(err, service.ErrAlreadyWatching):
			response.BadRequest(c, err.Error())
		default:
			response.InternalError(c, "添加关注失败")
		}
		return
	}

	response.Success(c, gin.H{"message": "关注成功"})
}

// HandleRemoveWatcher 移除关注人
func (h *IssueHandler) HandleRemoveWatcher(c *gin.Context) {
	key := c.Param("key")
	userID, err := strconv.ParseUint(c.Param("user_id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的用户 ID")
		return
	}

	err = h.issueService.RemoveWatcher(c.Request.Context(), key, userID)
	if err != nil {
		if errors.Is(err, service.ErrIssueNotFound) {
			response.NotFound(c, err.Error())
			return
		}
		response.InternalError(c, "取消关注失败")
		return
	}

	response.Success(c, gin.H{"message": "取消关注成功"})
}

// HandleListWatchers 获取关注人列表
func (h *IssueHandler) HandleListWatchers(c *gin.Context) {
	key := c.Param("key")

	watchers, err := h.issueService.ListWatchers(c.Request.Context(), key)
	if err != nil {
		if errors.Is(err, service.ErrIssueNotFound) {
			response.NotFound(c, err.Error())
			return
		}
		response.InternalError(c, "获取关注人列表失败")
		return
	}

	response.Success(c, watchers)
}

// HandleListMyTodoIssues 获取我的待办工单
func (h *IssueHandler) HandleListMyTodoIssues(c *gin.Context) {
	userID := c.GetUint64("user_id")

	page, err := strconv.Atoi(c.DefaultQuery("page", "1"))
	if err != nil || page <= 0 {
		page = 1
	}
	pageSize, err := strconv.Atoi(c.DefaultQuery("page_size", "10"))
	if err != nil || pageSize <= 0 {
		pageSize = 10
	}

	// 从中间件注入的项目 ID 列表（非管理员用户）
	var projectIDs []uint64
	if ids, exists := c.Get("user_project_ids"); exists {
		if v, ok := ids.([]uint64); ok {
			projectIDs = v
		}
	}

	issues, total, hasMore, err := h.issueService.ListMyTodoIssues(c.Request.Context(), userID, page, pageSize, projectIDs)
	if err != nil {
		response.InternalError(c, "获取我的待办工单失败")
		return
	}

	response.SuccessWithPageHasMore(c, issues, total, page, pageSize, hasMore)
}

// HandleListMyCreatedIssues 获取我创建的工单
func (h *IssueHandler) HandleListMyCreatedIssues(c *gin.Context) {
	userID := c.GetUint64("user_id")

	page, err := strconv.Atoi(c.DefaultQuery("page", "1"))
	if err != nil || page <= 0 {
		page = 1
	}
	pageSize, err := strconv.Atoi(c.DefaultQuery("page_size", "10"))
	if err != nil || pageSize <= 0 {
		pageSize = 10
	}

	// 从中间件注入的项目 ID 列表（非管理员用户）
	var projectIDs []uint64
	if ids, exists := c.Get("user_project_ids"); exists {
		if v, ok := ids.([]uint64); ok {
			projectIDs = v
		}
	}

	issues, total, hasMore, err := h.issueService.ListMyCreatedIssues(c.Request.Context(), userID, page, pageSize, projectIDs)
	if err != nil {
		response.InternalError(c, "获取我创建的工单失败")
		return
	}

	response.SuccessWithPageHasMore(c, issues, total, page, pageSize, hasMore)
}

// ============ 工作日志相关处理器 ============

// HandleAddWorklog 添加工作日志
func (h *IssueHandler) HandleAddWorklog(c *gin.Context) {
	issueKey := c.Param("key")
	userID := c.GetUint64("user_id")

	var req dto.CreateWorklogRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数错误: "+err.Error())
		return
	}

	result, err := h.issueService.AddWorklog(c.Request.Context(), issueKey, &req, userID)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrIssueNotFound):
			response.NotFound(c, err.Error())
		case errors.Is(err, service.ErrInvalidTimeFormat):
			response.BadRequest(c, err.Error())
		default:
			response.InternalError(c, "添加工作日志失败")
		}
		return
	}

	response.Created(c, result)
}

// HandleUpdateWorklog 更新工作日志
func (h *IssueHandler) HandleUpdateWorklog(c *gin.Context) {
	worklogIDStr := c.Param("worklog_id")
	worklogID, err := strconv.ParseUint(worklogIDStr, 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的工作日志ID")
		return
	}

	userID := c.GetUint64("user_id")

	var req dto.UpdateWorklogRequest
	if bindErr := c.ShouldBindJSON(&req); bindErr != nil {
		response.BadRequest(c, "请求参数错误: "+bindErr.Error())
		return
	}

	result, err := h.issueService.UpdateWorklog(c.Request.Context(), worklogID, &req, userID)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrWorklogNotFound):
			response.NotFound(c, err.Error())
		case errors.Is(err, service.ErrUnauthorized):
			response.Forbidden(c, "无权限操作")
		case errors.Is(err, service.ErrInvalidTimeFormat):
			response.BadRequest(c, err.Error())
		default:
			response.InternalError(c, "更新工作日志失败")
		}
		return
	}

	response.Success(c, result)
}

// HandleDeleteWorklog 删除工作日志
func (h *IssueHandler) HandleDeleteWorklog(c *gin.Context) {
	worklogIDStr := c.Param("worklog_id")
	worklogID, err := strconv.ParseUint(worklogIDStr, 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的工作日志ID")
		return
	}

	userID := c.GetUint64("user_id")

	err = h.issueService.DeleteWorklog(c.Request.Context(), worklogID, userID)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrWorklogNotFound):
			response.NotFound(c, err.Error())
		case errors.Is(err, service.ErrUnauthorized):
			response.Forbidden(c, "无权限操作")
		default:
			response.InternalError(c, "删除工作日志失败")
		}
		return
	}

	response.Success(c, nil)
}

// HandleListWorklogs 获取工作日志列表
func (h *IssueHandler) HandleListWorklogs(c *gin.Context) {
	issueKey := c.Param("key")

	worklogs, err := h.issueService.ListWorklogs(c.Request.Context(), issueKey)
	if err != nil {
		if errors.Is(err, service.ErrIssueNotFound) {
			response.NotFound(c, err.Error())
			return
		}
		response.InternalError(c, "获取工作日志列表失败")
		return
	}

	response.Success(c, worklogs)
}
