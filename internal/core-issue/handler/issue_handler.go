// Package handler 提供工单模块的 HTTP 处理器
package handler

import (
	"errors"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/kerbos/ticketdesk/internal/api/response"
	"github.com/kerbos/ticketdesk/internal/core-issue/dto"
	"github.com/kerbos/ticketdesk/internal/core-issue/service"
)

// IssueHandler 工单处理器
type IssueHandler struct {
	issueService service.IssueService
}

// NewIssueHandler 创建工单处理器实例
func NewIssueHandler(issueService service.IssueService) *IssueHandler {
	return &IssueHandler{issueService: issueService}
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

	result, err := h.issueService.UpdateIssue(c.Request.Context(), key, &req)
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

	err := h.issueService.DeleteIssue(c.Request.Context(), key)
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

	issues, total, err := h.issueService.ListIssues(c.Request.Context(), &req)
	if err != nil {
		if errors.Is(err, service.ErrProjectNotFound) {
			response.NotFound(c, err.Error())
			return
		}
		response.InternalError(c, "获取工单列表失败")
		return
	}

	response.SuccessWithPage(c, issues, total, req.GetDefaultPage(), req.GetDefaultPageSize())
}

// HandleTransitionIssue 工单状态流转
func (h *IssueHandler) HandleTransitionIssue(c *gin.Context) {
	key := c.Param("key")

	var req dto.TransitionIssueRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数错误: "+err.Error())
		return
	}

	userID := c.GetUint64("user_id")
	result, err := h.issueService.TransitionIssue(c.Request.Context(), key, &req, userID)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrIssueNotFound):
			response.NotFound(c, err.Error())
		case errors.Is(err, service.ErrInvalidTransition):
			response.BadRequest(c, err.Error())
		default:
			response.InternalError(c, "状态流转失败")
		}
		return
	}

	response.Success(c, result)
}

// HandleAssignIssue 指派工单
func (h *IssueHandler) HandleAssignIssue(c *gin.Context) {
	key := c.Param("key")

	var req dto.AssignIssueRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数错误: "+err.Error())
		return
	}

	result, err := h.issueService.AssignIssue(c.Request.Context(), key, req.AssigneeID)
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

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))

	issues, total, err := h.issueService.ListMyTodoIssues(c.Request.Context(), userID, page, pageSize)
	if err != nil {
		response.InternalError(c, "获取我的待办工单失败")
		return
	}

	response.SuccessWithPage(c, issues, total, page, pageSize)
}

// HandleListMyCreatedIssues 获取我创建的工单
func (h *IssueHandler) HandleListMyCreatedIssues(c *gin.Context) {
	userID := c.GetUint64("user_id")

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))

	issues, total, err := h.issueService.ListMyCreatedIssues(c.Request.Context(), userID, page, pageSize)
	if err != nil {
		response.InternalError(c, "获取我创建的工单失败")
		return
	}

	response.SuccessWithPage(c, issues, total, page, pageSize)
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
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数错误: "+err.Error())
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
