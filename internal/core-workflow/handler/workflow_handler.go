// Package handler 提供工作流模块的 HTTP 处理器
package handler

import (
	"errors"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/kerbos/ticketdesk/internal/api/response"
	"github.com/kerbos/ticketdesk/internal/core-workflow/dto"
	"github.com/kerbos/ticketdesk/internal/core-workflow/service"
)

// WorkflowHandler 工作流处理器
type WorkflowHandler struct {
	workflowService service.WorkflowService
}

// NewWorkflowHandler 创建工作流处理器实例
func NewWorkflowHandler(workflowService service.WorkflowService) *WorkflowHandler {
	return &WorkflowHandler{workflowService: workflowService}
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
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数错误: "+err.Error())
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
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数错误: "+err.Error())
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
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数错误: "+err.Error())
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
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数错误: "+err.Error())
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
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数错误: "+err.Error())
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
