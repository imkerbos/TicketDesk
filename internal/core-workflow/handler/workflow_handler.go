// Package handler 提供工作流模块的 HTTP 处理器
package handler

import (
	"context"
	"errors"
	"fmt"
	"strconv"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/kerbos/ticketdesk/internal/api/response"
	issueRepo "github.com/kerbos/ticketdesk/internal/core-issue/repository"
	projectRepo "github.com/kerbos/ticketdesk/internal/core-project/repository"
	"github.com/kerbos/ticketdesk/internal/core-workflow/dto"
	"github.com/kerbos/ticketdesk/internal/core-workflow/service"
	"github.com/kerbos/ticketdesk/pkg/logger"
)

// ActivityLogger 活动日志记录接口
type ActivityLogger interface {
	LogActivity(ctx context.Context, userID uint64, userName, action, entityType string, entityID uint64, entityKey, details string) error
}

// WorkflowHandler 工作流处理器
type WorkflowHandler struct {
	workflowService service.WorkflowService
	workflowEngine  service.WorkflowEngine
	issueRepo       issueRepo.IssueRepository
	projectRepo     projectRepo.ProjectRepository
	activityLogger  ActivityLogger
}

// NewWorkflowHandler 创建工作流处理器实例
func NewWorkflowHandler(
	workflowService service.WorkflowService,
	workflowEngine service.WorkflowEngine,
	issueRepository issueRepo.IssueRepository,
	projectRepository projectRepo.ProjectRepository,
	activityLogger ActivityLogger,
) *WorkflowHandler {
	return &WorkflowHandler{
		workflowService: workflowService,
		workflowEngine:  workflowEngine,
		issueRepo:       issueRepository,
		projectRepo:     projectRepository,
		activityLogger:  activityLogger,
	}
}

// logActivity 异步记录活动日志（不阻塞主流程）
func (h *WorkflowHandler) logActivity(userID uint64, userName, action, entityKey, details string, entityID uint64) {
	if h.activityLogger == nil {
		return
	}
	go func() {
		if err := h.activityLogger.LogActivity(context.Background(), userID, userName, action, "issue", entityID, entityKey, details); err != nil {
			logger.Warn("failed to log workflow activity", zap.Error(err))
		}
	}()
}

// ============ 工作流管理 ============

// HandleCreateWorkflow 创建工作流
func (h *WorkflowHandler) HandleCreateWorkflow(c *gin.Context) {
	var req dto.CreateWorkflowRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数错误: "+err.Error())
		return
	}

	result, err := h.workflowService.CreateWorkflow(c.Request.Context(), &req)
	if err != nil {
		response.InternalError(c, "创建工作流失败")
		return
	}

	response.Created(c, result)
}

// HandleGetWorkflow 获取工作流详情
func (h *WorkflowHandler) HandleGetWorkflow(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的工作流 ID")
		return
	}

	result, err := h.workflowService.GetWorkflow(c.Request.Context(), id)
	if err != nil {
		if errors.Is(err, service.ErrWorkflowNotFound) {
			response.NotFound(c, err.Error())
			return
		}
		response.InternalError(c, "获取工作流失败")
		return
	}

	response.Success(c, result)
}

// HandleUpdateWorkflow 更新工作流
func (h *WorkflowHandler) HandleUpdateWorkflow(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的工作流 ID")
		return
	}

	var req dto.UpdateWorkflowRequest
	if bindErr := c.ShouldBindJSON(&req); bindErr != nil {
		response.BadRequest(c, "请求参数错误: "+bindErr.Error())
		return
	}

	result, err := h.workflowService.UpdateWorkflow(c.Request.Context(), id, &req)
	if err != nil {
		if errors.Is(err, service.ErrWorkflowNotFound) {
			response.NotFound(c, err.Error())
			return
		}
		response.InternalError(c, "更新工作流失败")
		return
	}

	response.Success(c, result)
}

// HandleDeleteWorkflow 删除工作流
func (h *WorkflowHandler) HandleDeleteWorkflow(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的工作流 ID")
		return
	}

	err = h.workflowService.DeleteWorkflow(c.Request.Context(), id)
	if err != nil {
		if errors.Is(err, service.ErrWorkflowNotFound) {
			response.NotFound(c, err.Error())
			return
		}
		response.InternalError(c, "删除工作流失败")
		return
	}

	response.Success(c, gin.H{"message": "工作流删除成功"})
}

// HandleListWorkflows 获取工作流列表
func (h *WorkflowHandler) HandleListWorkflows(c *gin.Context) {
	var req dto.ListWorkflowsRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.BadRequest(c, "请求参数错误: "+err.Error())
		return
	}

	workflows, total, err := h.workflowService.ListWorkflows(c.Request.Context(), &req)
	if err != nil {
		response.InternalError(c, "获取工作流列表失败")
		return
	}

	response.SuccessWithPage(c, workflows, total, req.GetDefaultPage(), req.GetDefaultPageSize())
}

// ============ 节点管理 ============

// HandleCreateNode 创建节点
func (h *WorkflowHandler) HandleCreateNode(c *gin.Context) {
	workflowID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的工作流 ID")
		return
	}

	var req dto.CreateNodeRequest
	if bindErr := c.ShouldBindJSON(&req); bindErr != nil {
		response.BadRequest(c, "请求参数错误: "+bindErr.Error())
		return
	}

	result, err := h.workflowService.CreateNode(c.Request.Context(), workflowID, &req)
	if err != nil {
		if errors.Is(err, service.ErrWorkflowNotFound) {
			response.NotFound(c, err.Error())
			return
		}
		response.InternalError(c, "创建节点失败")
		return
	}

	response.Created(c, result)
}

// HandleGetNode 获取节点详情
func (h *WorkflowHandler) HandleGetNode(c *gin.Context) {
	nodeID, err := strconv.ParseUint(c.Param("node_id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的节点 ID")
		return
	}

	result, err := h.workflowService.GetNode(c.Request.Context(), nodeID)
	if err != nil {
		if errors.Is(err, service.ErrNodeNotFound) {
			response.NotFound(c, err.Error())
			return
		}
		response.InternalError(c, "获取节点失败")
		return
	}

	response.Success(c, result)
}

// HandleUpdateNode 更新节点
func (h *WorkflowHandler) HandleUpdateNode(c *gin.Context) {
	nodeID, err := strconv.ParseUint(c.Param("node_id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的节点 ID")
		return
	}

	var req dto.UpdateNodeRequest
	if bindErr := c.ShouldBindJSON(&req); bindErr != nil {
		response.BadRequest(c, "请求参数错误: "+bindErr.Error())
		return
	}

	result, err := h.workflowService.UpdateNode(c.Request.Context(), nodeID, &req)
	if err != nil {
		if errors.Is(err, service.ErrNodeNotFound) {
			response.NotFound(c, err.Error())
			return
		}
		response.InternalError(c, "更新节点失败")
		return
	}

	response.Success(c, result)
}

// HandleDeleteNode 删除节点
func (h *WorkflowHandler) HandleDeleteNode(c *gin.Context) {
	nodeID, err := strconv.ParseUint(c.Param("node_id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的节点 ID")
		return
	}

	err = h.workflowService.DeleteNode(c.Request.Context(), nodeID)
	if err != nil {
		if errors.Is(err, service.ErrNodeNotFound) {
			response.NotFound(c, err.Error())
			return
		}
		response.BadRequest(c, err.Error())
		return
	}

	response.Success(c, gin.H{"message": "节点删除成功"})
}

// HandleListNodes 获取工作流的所有节点
func (h *WorkflowHandler) HandleListNodes(c *gin.Context) {
	workflowID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的工作流 ID")
		return
	}

	nodes, err := h.workflowService.ListNodes(c.Request.Context(), workflowID)
	if err != nil {
		if errors.Is(err, service.ErrWorkflowNotFound) {
			response.NotFound(c, err.Error())
			return
		}
		response.InternalError(c, "获取节点列表失败")
		return
	}

	response.Success(c, nodes)
}

// ============ 边管理 ============

// HandleCreateEdge 创建边
func (h *WorkflowHandler) HandleCreateEdge(c *gin.Context) {
	workflowID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的工作流 ID")
		return
	}

	var req dto.CreateEdgeRequest
	if bindErr := c.ShouldBindJSON(&req); bindErr != nil {
		response.BadRequest(c, "请求参数错误: "+bindErr.Error())
		return
	}

	result, err := h.workflowService.CreateEdge(c.Request.Context(), workflowID, &req)
	if err != nil {
		if errors.Is(err, service.ErrWorkflowNotFound) {
			response.NotFound(c, err.Error())
			return
		}
		response.BadRequest(c, err.Error())
		return
	}

	response.Created(c, result)
}

// HandleGetEdge 获取边详情
func (h *WorkflowHandler) HandleGetEdge(c *gin.Context) {
	edgeID, err := strconv.ParseUint(c.Param("edge_id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的边 ID")
		return
	}

	result, err := h.workflowService.GetEdge(c.Request.Context(), edgeID)
	if err != nil {
		if errors.Is(err, service.ErrEdgeNotFound) {
			response.NotFound(c, err.Error())
			return
		}
		response.InternalError(c, "获取边失败")
		return
	}

	response.Success(c, result)
}

// HandleUpdateEdge 更新边
func (h *WorkflowHandler) HandleUpdateEdge(c *gin.Context) {
	edgeID, err := strconv.ParseUint(c.Param("edge_id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的边 ID")
		return
	}

	var req dto.UpdateEdgeRequest
	if bindErr := c.ShouldBindJSON(&req); bindErr != nil {
		response.BadRequest(c, "请求参数错误: "+bindErr.Error())
		return
	}

	result, err := h.workflowService.UpdateEdge(c.Request.Context(), edgeID, &req)
	if err != nil {
		if errors.Is(err, service.ErrEdgeNotFound) {
			response.NotFound(c, err.Error())
			return
		}
		response.InternalError(c, "更新边失败")
		return
	}

	response.Success(c, result)
}

// HandleDeleteEdge 删除边
func (h *WorkflowHandler) HandleDeleteEdge(c *gin.Context) {
	edgeID, err := strconv.ParseUint(c.Param("edge_id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的边 ID")
		return
	}

	err = h.workflowService.DeleteEdge(c.Request.Context(), edgeID)
	if err != nil {
		if errors.Is(err, service.ErrEdgeNotFound) {
			response.NotFound(c, err.Error())
			return
		}
		response.InternalError(c, "删除边失败")
		return
	}

	response.Success(c, gin.H{"message": "边删除成功"})
}

// HandleListEdges 获取工作流的所有边
func (h *WorkflowHandler) HandleListEdges(c *gin.Context) {
	workflowID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的工作流 ID")
		return
	}

	edges, err := h.workflowService.ListEdges(c.Request.Context(), workflowID)
	if err != nil {
		if errors.Is(err, service.ErrWorkflowNotFound) {
			response.NotFound(c, err.Error())
			return
		}
		response.InternalError(c, "获取边列表失败")
		return
	}

	response.Success(c, edges)
}

// ============ 工作流实例管理 ============

// HandleGetInstanceByIssue 获取工单的工作流实例
// @Summary 获取工单的工作流实例
// @Description 根据工单 Key 获取工作流实例
// @Tags Workflow
// @Produce json
// @Param key path string true "工单 Key"
// @Success 200 {object} response.Response{data=dto.WorkflowInstanceResponse}
// @Failure 404 {object} response.ErrorResponse
// @Router /api/v1/issues/{key}/workflow [get]
// @Security BearerAuth
func (h *WorkflowHandler) HandleGetInstanceByIssue(c *gin.Context) {
	issueKey := c.Param("key")
	if issueKey == "" {
		response.BadRequest(c, "工单 Key 不能为空")
		return
	}

	// 通过 issue key 获取 issue
	issue, err := h.issueRepo.GetByKey(c.Request.Context(), issueKey)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			response.NotFound(c, "工单不存在")
			return
		}
		response.InternalError(c, "查询工单失败")
		return
	}

	// 获取工作流实例
	instance, err := h.workflowEngine.GetInstanceByIssueID(c.Request.Context(), issue.ID)
	if err != nil {
		if errors.Is(err, service.ErrWorkflowInstanceNotFound) {
			// 尝试懒加载创建：如果该工单的项目+类型配置了工作流方案，自动补建实例
			newInstance, createErr := h.workflowEngine.TryCreateInstanceForIssue(
				c.Request.Context(), issue.ID, issue.ProjectID, issue.IssueTypeID,
			)
			if createErr != nil || newInstance == nil {
				response.NotFound(c, "该工单没有关联的工作流实例")
				return
			}

			// 更新工单的 WorkflowInstanceID
			issue.WorkflowInstanceID = &newInstance.ID
			if updateErr := h.issueRepo.Update(c.Request.Context(), issue); updateErr != nil {
				logger.Warn("failed to update issue workflow_instance_id", zap.Error(updateErr))
			}

			// 重新获取完整的实例响应
			instance, err = h.workflowEngine.GetInstanceByIssueID(c.Request.Context(), issue.ID)
			if err != nil {
				response.InternalError(c, "获取工作流实例失败")
				return
			}
		} else {
			response.InternalError(c, "获取工作流实例失败")
			return
		}
	}

	// 填充 IssueKey
	instance.IssueKey = issueKey

	response.Success(c, instance)
}

// HandleApprove 审批通过
// @Summary 审批通过
// @Description 对工作流实例进行审批通过操作
// @Tags Workflow
// @Accept json
// @Produce json
// @Param key path string true "工单 Key"
// @Param request body dto.ApproveRequest true "审批请求"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Router /api/v1/issues/{key}/workflow/approve [post]
// @Security BearerAuth
func (h *WorkflowHandler) HandleApprove(c *gin.Context) {
	issueKey := c.Param("key")
	if issueKey == "" {
		response.BadRequest(c, "工单 Key 不能为空")
		return
	}

	// 获取当前用户 ID
	userID, exists := c.Get("user_id")
	if !exists {
		response.Unauthorized(c, "未获取到用户信息")
		return
	}

	// 解析请求体
	var req dto.ApproveRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数错误: "+err.Error())
		return
	}

	// 通过 issue key 获取 issue
	issue, err := h.issueRepo.GetByKey(c.Request.Context(), issueKey)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			response.NotFound(c, "工单不存在")
			return
		}
		response.InternalError(c, "查询工单失败")
		return
	}

	// 获取工作流实例
	instance, err := h.workflowEngine.GetInstanceByIssueID(c.Request.Context(), issue.ID)
	if err != nil {
		if errors.Is(err, service.ErrWorkflowInstanceNotFound) {
			response.NotFound(c, "该工单没有关联的工作流实例")
			return
		}
		response.InternalError(c, "获取工作流实例失败")
		return
	}

	// 执行审批通过
	uid := userID.(uint64)
	if err := h.workflowEngine.Approve(c.Request.Context(), instance.ID, uid, req.Comment); err != nil {
		if errors.Is(err, service.ErrNotApprover) {
			response.Forbidden(c, "当前用户不是审批人")
			return
		}
		if errors.Is(err, service.ErrAlreadyApproved) {
			response.BadRequest(c, "已经审批过了")
			return
		}
		response.InternalError(c, "审批操作失败: "+err.Error())
		return
	}

	// 记录活动日志
	userName, _ := c.Get("username")
	nodeName := ""
	if instance.CurrentNode != nil {
		nodeName = instance.CurrentNode.Name
	}
	details := "节点: " + nodeName
	if req.Comment != "" {
		details += " - " + req.Comment
	}
	h.logActivity(uid, fmt.Sprint(userName), "审批通过", issueKey, details, issue.ID)

	response.Success(c, gin.H{"message": "审批通过"})
}

// HandleReject 审批拒绝
// @Summary 审批拒绝
// @Description 对工作流实例进行审批拒绝操作
// @Tags Workflow
// @Accept json
// @Produce json
// @Param key path string true "工单 Key"
// @Param request body dto.RejectRequest true "拒绝请求"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Router /api/v1/issues/{key}/workflow/reject [post]
// @Security BearerAuth
func (h *WorkflowHandler) HandleReject(c *gin.Context) {
	issueKey := c.Param("key")
	if issueKey == "" {
		response.BadRequest(c, "工单 Key 不能为空")
		return
	}

	// 获取当前用户 ID
	userID, exists := c.Get("user_id")
	if !exists {
		response.Unauthorized(c, "未获取到用户信息")
		return
	}

	// 解析请求体
	var req dto.RejectRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数错误: "+err.Error())
		return
	}

	// 通过 issue key 获取 issue
	issue, err := h.issueRepo.GetByKey(c.Request.Context(), issueKey)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			response.NotFound(c, "工单不存在")
			return
		}
		response.InternalError(c, "查询工单失败")
		return
	}

	// 获取工作流实例
	instance, err := h.workflowEngine.GetInstanceByIssueID(c.Request.Context(), issue.ID)
	if err != nil {
		if errors.Is(err, service.ErrWorkflowInstanceNotFound) {
			response.NotFound(c, "该工单没有关联的工作流实例")
			return
		}
		response.InternalError(c, "获取工作流实例失败")
		return
	}

	// 执行审批拒绝
	uid := userID.(uint64)
	if err := h.workflowEngine.Reject(c.Request.Context(), instance.ID, uid, req.Comment); err != nil {
		if errors.Is(err, service.ErrNotApprover) {
			response.Forbidden(c, "当前用户不是审批人")
			return
		}
		if errors.Is(err, service.ErrAlreadyApproved) {
			response.BadRequest(c, "已经审批过了")
			return
		}
		response.InternalError(c, "拒绝操作失败: "+err.Error())
		return
	}

	// 记录活动日志
	userName, _ := c.Get("username")
	nodeName := ""
	if instance.CurrentNode != nil {
		nodeName = instance.CurrentNode.Name
	}
	details := "节点: " + nodeName
	if req.Comment != "" {
		details += " - " + req.Comment
	}
	h.logActivity(uid, fmt.Sprint(userName), "审批拒绝", issueKey, details, issue.ID)

	response.Success(c, gin.H{"message": "审批已拒绝"})
}

// HandleComplete 完成工作节点
func (h *WorkflowHandler) HandleComplete(c *gin.Context) {
	issueKey := c.Param("key")
	if issueKey == "" {
		response.BadRequest(c, "工单 Key 不能为空")
		return
	}

	userID, exists := c.Get("user_id")
	if !exists {
		response.Unauthorized(c, "未获取到用户信息")
		return
	}

	var req dto.CompleteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数错误: "+err.Error())
		return
	}

	issue, err := h.issueRepo.GetByKey(c.Request.Context(), issueKey)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			response.NotFound(c, "工单不存在")
			return
		}
		response.InternalError(c, "查询工单失败")
		return
	}

	instance, err := h.workflowEngine.GetInstanceByIssueID(c.Request.Context(), issue.ID)
	if err != nil {
		if errors.Is(err, service.ErrWorkflowInstanceNotFound) {
			response.NotFound(c, "该工单没有关联的工作流实例")
			return
		}
		response.InternalError(c, "获取工作流实例失败")
		return
	}

	uid := userID.(uint64)
	if err := h.workflowEngine.Complete(c.Request.Context(), instance.ID, uid, req.Comment, req.Result); err != nil {
		response.InternalError(c, "完成操作失败: "+err.Error())
		return
	}

	// 记录活动日志
	userName, _ := c.Get("username")
	nodeName := ""
	if instance.CurrentNode != nil {
		nodeName = instance.CurrentNode.Name
	}
	// 将条件值转为友好文本
	resultText := "完成"
	if req.Result != "" {
		presetLabels := map[string]string{
			"approved": "通过",
			"rejected": "退回",
		}
		if label, ok := presetLabels[req.Result]; ok {
			resultText = label
		} else {
			resultText = req.Result
		}
	}
	details := "节点: " + nodeName + " (" + resultText + ")"
	if req.Comment != "" {
		details += " - " + req.Comment
	}
	h.logActivity(uid, fmt.Sprint(userName), "完成工作节点", issueKey, details, issue.ID)

	response.Success(c, gin.H{"message": "工作节点已完成"})
}

// HandleGetHistory 获取流转历史
// @Summary 获取流转历史
// @Description 获取工作流实例的流转历史
// @Tags Workflow
// @Produce json
// @Param key path string true "工单 Key"
// @Success 200 {object} response.Response{data=[]dto.WorkflowHistoryResponse}
// @Failure 404 {object} response.ErrorResponse
// @Router /api/v1/issues/{key}/workflow/history [get]
// @Security BearerAuth
func (h *WorkflowHandler) HandleGetHistory(c *gin.Context) {
	issueKey := c.Param("key")
	if issueKey == "" {
		response.BadRequest(c, "工单 Key 不能为空")
		return
	}

	// 通过 issue key 获取 issue
	issue, err := h.issueRepo.GetByKey(c.Request.Context(), issueKey)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			response.NotFound(c, "工单不存在")
			return
		}
		response.InternalError(c, "查询工单失败")
		return
	}

	// 获取工作流实例
	instance, err := h.workflowEngine.GetInstanceByIssueID(c.Request.Context(), issue.ID)
	if err != nil {
		if errors.Is(err, service.ErrWorkflowInstanceNotFound) {
			response.NotFound(c, "该工单没有关联的工作流实例")
			return
		}
		response.InternalError(c, "获取工作流实例失败")
		return
	}

	// 获取流转历史
	history, err := h.workflowEngine.GetHistory(c.Request.Context(), instance.ID)
	if err != nil {
		response.InternalError(c, "获取流转历史失败")
		return
	}

	response.Success(c, history)
}

// ============ 工作流方案管理 ============

// HandleCreateScheme 创建工作流方案
// @Summary 创建工作流方案
// @Description 为项目的工单类型配置工作流
// @Tags Workflow
// @Accept json
// @Produce json
// @Param key path string true "项目 Key"
// @Param request body dto.CreateWorkflowSchemeRequest true "创建方案请求"
// @Success 201 {object} response.Response{data=dto.WorkflowSchemeResponse}
// @Failure 400 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Router /api/v1/projects/{key}/workflow-schemes [post]
// @Security BearerAuth
func (h *WorkflowHandler) HandleCreateScheme(c *gin.Context) {
	projectKey := c.Param("key")
	if projectKey == "" {
		response.BadRequest(c, "项目 Key 不能为空")
		return
	}

	// 通过项目 key 获取项目
	project, err := h.projectRepo.GetByKey(c.Request.Context(), projectKey)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			response.NotFound(c, "项目不存在")
			return
		}
		response.InternalError(c, "查询项目失败")
		return
	}

	// 解析请求体
	var req dto.CreateWorkflowSchemeRequest
	if bindErr := c.ShouldBindJSON(&req); bindErr != nil {
		response.BadRequest(c, "请求参数错误: "+bindErr.Error())
		return
	}

	// 创建方案
	result, err := h.workflowService.CreateScheme(c.Request.Context(), project.ID, &req)
	if err != nil {
		if errors.Is(err, service.ErrSchemeExists) {
			response.BadRequest(c, "该工单类型已配置工作流方案")
			return
		}
		if errors.Is(err, service.ErrWorkflowNotFound) {
			response.NotFound(c, "工作流不存在")
			return
		}
		response.InternalError(c, "创建工作流方案失败")
		return
	}

	response.Created(c, result)
}

// HandleListSchemes 获取项目的工作流方案列表
// @Summary 获取工作流方案列表
// @Description 获取项目的所有工作流方案
// @Tags Workflow
// @Produce json
// @Param key path string true "项目 Key"
// @Success 200 {object} response.Response{data=[]dto.WorkflowSchemeResponse}
// @Failure 404 {object} response.ErrorResponse
// @Router /api/v1/projects/{key}/workflow-schemes [get]
// @Security BearerAuth
func (h *WorkflowHandler) HandleListSchemes(c *gin.Context) {
	projectKey := c.Param("key")
	if projectKey == "" {
		response.BadRequest(c, "项目 Key 不能为空")
		return
	}

	// 通过项目 key 获取项目
	project, err := h.projectRepo.GetByKey(c.Request.Context(), projectKey)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			response.NotFound(c, "项目不存在")
			return
		}
		response.InternalError(c, "查询项目失败")
		return
	}

	// 获取方案列表
	schemes, err := h.workflowService.ListSchemes(c.Request.Context(), project.ID)
	if err != nil {
		response.InternalError(c, "获取工作流方案列表失败")
		return
	}

	response.Success(c, schemes)
}

// HandleDeleteScheme 删除工作流方案
// @Summary 删除工作流方案
// @Description 删除项目工单类型的工作流方案
// @Tags Workflow
// @Produce json
// @Param key path string true "项目 Key"
// @Param type_id path int true "工单类型 ID"
// @Success 200 {object} response.Response
// @Failure 404 {object} response.ErrorResponse
// @Router /api/v1/projects/{key}/workflow-schemes/{type_id} [delete]
// @Security BearerAuth
func (h *WorkflowHandler) HandleDeleteScheme(c *gin.Context) {
	projectKey := c.Param("key")
	if projectKey == "" {
		response.BadRequest(c, "项目 Key 不能为空")
		return
	}

	typeID, err := strconv.ParseUint(c.Param("type_id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的工单类型 ID")
		return
	}

	// 通过项目 key 获取项目
	project, err := h.projectRepo.GetByKey(c.Request.Context(), projectKey)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			response.NotFound(c, "项目不存在")
			return
		}
		response.InternalError(c, "查询项目失败")
		return
	}

	// 删除方案
	if err := h.workflowService.DeleteScheme(c.Request.Context(), project.ID, typeID); err != nil {
		if errors.Is(err, service.ErrSchemeNotFound) {
			response.NotFound(c, "工作流方案不存在")
			return
		}
		response.InternalError(c, "删除工作流方案失败")
		return
	}

	response.Success(c, gin.H{"message": "工作流方案删除成功"})
}
