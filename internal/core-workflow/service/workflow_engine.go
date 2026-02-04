// Package service 提供工作流引擎服务
package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/kerbos/ticketdesk/internal/core-workflow/dto"
	"github.com/kerbos/ticketdesk/internal/core-workflow/repository"
	projectRepo "github.com/kerbos/ticketdesk/internal/core-project/repository"
	userRepo "github.com/kerbos/ticketdesk/internal/core-user/repository"
	"github.com/kerbos/ticketdesk/internal/model"
	"github.com/kerbos/ticketdesk/pkg/logger"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// 工作流引擎错误定义
var (
	ErrWorkflowInstanceNotFound = errors.New("工作流实例不存在")
	ErrNotApprover              = errors.New("当前用户不是审批人")
	ErrAlreadyApproved          = errors.New("已经审批过了")
	ErrNoApproversConfigured    = errors.New("未配置审批人")
)

// WorkflowEngine 工作流引擎接口
type WorkflowEngine interface {
	// 创建工作流实例
	CreateInstance(ctx context.Context, issueID, workflowID uint64) (*model.WorkflowInstance, error)
	// 获取工作流实例
	GetInstance(ctx context.Context, instanceID uint64) (*dto.WorkflowInstanceResponse, error)
	// 获取工单的工作流实例
	GetInstanceByIssueID(ctx context.Context, issueID uint64) (*dto.WorkflowInstanceResponse, error)
	// 审批通过
	Approve(ctx context.Context, instanceID, userID uint64, comment string) error
	// 审批拒绝
	Reject(ctx context.Context, instanceID, userID uint64, comment string) error
	// 获取流转历史
	GetHistory(ctx context.Context, instanceID uint64) ([]*dto.WorkflowHistoryResponse, error)
}

// workflowEngine 工作流引擎实现
type workflowEngine struct {
	instanceRepo       repository.WorkflowInstanceRepository
	historyRepo        repository.WorkflowHistoryRepository
	approvalRepo       repository.ApprovalRecordRepository
	workflowRepo       repository.WorkflowRepository
	nodeRepo           repository.NodeRepository
	edgeRepo           repository.EdgeRepository
	projectRoleRepo    projectRepo.ProjectRoleRepository
	userRepo           userRepo.UserRepository
	db                 *gorm.DB
}

// NewWorkflowEngine 创建工作流引擎实例
func NewWorkflowEngine(
	instanceRepo repository.WorkflowInstanceRepository,
	historyRepo repository.WorkflowHistoryRepository,
	approvalRepo repository.ApprovalRecordRepository,
	workflowRepo repository.WorkflowRepository,
	nodeRepo repository.NodeRepository,
	edgeRepo repository.EdgeRepository,
	projectRoleRepo projectRepo.ProjectRoleRepository,
	userRepo userRepo.UserRepository,
	db *gorm.DB,
) WorkflowEngine {
	return &workflowEngine{
		instanceRepo:    instanceRepo,
		historyRepo:     historyRepo,
		approvalRepo:    approvalRepo,
		workflowRepo:    workflowRepo,
		nodeRepo:        nodeRepo,
		edgeRepo:        edgeRepo,
		projectRoleRepo: projectRoleRepo,
		userRepo:        userRepo,
		db:              db,
	}
}

// CreateInstance 创建工作流实例
func (e *workflowEngine) CreateInstance(ctx context.Context, issueID, workflowID uint64) (*model.WorkflowInstance, error) {
	// 获取工作流的开始节点
	startNode, err := e.nodeRepo.GetStartNode(ctx, workflowID)
	if err != nil {
		logger.Error("failed to get start node", zap.Error(err))
		return nil, fmt.Errorf("获取开始节点失败: %w", err)
	}

	// 创建工作流实例
	instance := &model.WorkflowInstance{
		IssueID:       issueID,
		WorkflowID:    workflowID,
		CurrentNodeID: startNode.ID,
		Status:        "active",
		StartedAt:     time.Now(),
	}

	if err := e.instanceRepo.Create(ctx, instance); err != nil {
		logger.Error("failed to create workflow instance", zap.Error(err))
		return nil, fmt.Errorf("创建工作流实例失败: %w", err)
	}

	// 记录流转历史
	history := &model.WorkflowHistory{
		InstanceID: instance.ID,
		FromNodeID: nil,
		ToNodeID:   startNode.ID,
		Action:     "start",
		OperatorID: 0, // 系统操作
		Comment:    "工作流启动",
		OperatedAt: time.Now(),
	}
	if err := e.historyRepo.Create(ctx, history); err != nil {
		logger.Warn("failed to create workflow history", zap.Error(err))
	}

	// 自动流转到下一个节点
	if err := e.moveToNextNode(ctx, instance, 0); err != nil {
		logger.Error("failed to move to next node", zap.Error(err))
		return nil, fmt.Errorf("流转到下一节点失败: %w", err)
	}

	logger.Info("workflow instance created",
		zap.Uint64("instance_id", instance.ID),
		zap.Uint64("issue_id", issueID),
		zap.Uint64("workflow_id", workflowID),
	)

	return instance, nil
}

// GetInstance 获取工作流实例
func (e *workflowEngine) GetInstance(ctx context.Context, instanceID uint64) (*dto.WorkflowInstanceResponse, error) {
	instance, err := e.instanceRepo.GetByID(ctx, instanceID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrWorkflowInstanceNotFound
		}
		return nil, fmt.Errorf("查询工作流实例失败: %w", err)
	}

	return e.toInstanceResponse(ctx, instance)
}

// GetInstanceByIssueID 获取工单的工作流实例
func (e *workflowEngine) GetInstanceByIssueID(ctx context.Context, issueID uint64) (*dto.WorkflowInstanceResponse, error) {
	instance, err := e.instanceRepo.GetByIssueID(ctx, issueID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrWorkflowInstanceNotFound
		}
		return nil, fmt.Errorf("查询工作流实例失败: %w", err)
	}

	return e.toInstanceResponse(ctx, instance)
}

// Approve 审批通过
func (e *workflowEngine) Approve(ctx context.Context, instanceID, userID uint64, comment string) error {
	instance, err := e.instanceRepo.GetByID(ctx, instanceID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrWorkflowInstanceNotFound
		}
		return fmt.Errorf("查询工作流实例失败: %w", err)
	}

	// 获取当前节点
	currentNode, err := e.nodeRepo.GetByID(ctx, instance.CurrentNodeID)
	if err != nil {
		return fmt.Errorf("查询当前节点失败: %w", err)
	}

	// 检查节点类型
	if currentNode.NodeType != "approval" {
		return fmt.Errorf("当前节点不是审批节点")
	}

	// 获取审批记录
	record, err := e.approvalRepo.GetByInstanceNodeAndApprover(ctx, instanceID, currentNode.ID, userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrNotApprover
		}
		return fmt.Errorf("查询审批记录失败: %w", err)
	}

	// 检查是否已审批
	if record.Status != "pending" {
		return ErrAlreadyApproved
	}

	// 更新审批记录
	now := time.Now()
	record.Status = "approved"
	record.Comment = comment
	record.ApprovedAt = &now

	if err := e.approvalRepo.Update(ctx, record); err != nil {
		logger.Error("failed to update approval record", zap.Error(err))
		return fmt.Errorf("更新审批记录失败: %w", err)
	}

	// 记录流转历史
	history := &model.WorkflowHistory{
		InstanceID: instanceID,
		FromNodeID: &currentNode.ID,
		ToNodeID:   currentNode.ID,
		Action:     "approve",
		OperatorID: userID,
		Comment:    comment,
		OperatedAt: now,
	}
	if err := e.historyRepo.Create(ctx, history); err != nil {
		logger.Warn("failed to create workflow history", zap.Error(err))
	}

	// 解析节点配置
	var config dto.NodeConfig
	if currentNode.Config != "" {
		if err := json.Unmarshal([]byte(currentNode.Config), &config); err != nil {
			logger.Error("failed to parse node config", zap.Error(err))
			return fmt.Errorf("解析节点配置失败: %w", err)
		}
	}

	// 检查审批是否完成
	complete, err := e.checkApprovalComplete(ctx, instanceID, currentNode.ID, config.ApprovalType)
	if err != nil {
		return fmt.Errorf("检查审批状态失败: %w", err)
	}

	// 如果审批完成，流转到下一个节点
	if complete {
		if err := e.moveToNextNode(ctx, instance, userID); err != nil {
			return fmt.Errorf("流转到下一节点失败: %w", err)
		}
	}

	logger.Info("approval approved",
		zap.Uint64("instance_id", instanceID),
		zap.Uint64("user_id", userID),
		zap.Bool("complete", complete),
	)

	return nil
}

// Reject 审批拒绝
func (e *workflowEngine) Reject(ctx context.Context, instanceID, userID uint64, comment string) error {
	instance, err := e.instanceRepo.GetByID(ctx, instanceID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrWorkflowInstanceNotFound
		}
		return fmt.Errorf("查询工作流实例失败: %w", err)
	}

	// 获取当前节点
	currentNode, err := e.nodeRepo.GetByID(ctx, instance.CurrentNodeID)
	if err != nil {
		return fmt.Errorf("查询当前节点失败: %w", err)
	}

	// 检查节点类型
	if currentNode.NodeType != "approval" {
		return fmt.Errorf("当前节点不是审批节点")
	}

	// 获取审批记录
	record, err := e.approvalRepo.GetByInstanceNodeAndApprover(ctx, instanceID, currentNode.ID, userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrNotApprover
		}
		return fmt.Errorf("查询审批记录失败: %w", err)
	}

	// 检查是否已审批
	if record.Status != "pending" {
		return ErrAlreadyApproved
	}

	// 更新审批记录
	now := time.Now()
	record.Status = "rejected"
	record.Comment = comment
	record.ApprovedAt = &now

	if err := e.approvalRepo.Update(ctx, record); err != nil {
		logger.Error("failed to update approval record", zap.Error(err))
		return fmt.Errorf("更新审批记录失败: %w", err)
	}

	// 记录流转历史
	history := &model.WorkflowHistory{
		InstanceID: instanceID,
		FromNodeID: &currentNode.ID,
		ToNodeID:   currentNode.ID,
		Action:     "reject",
		OperatorID: userID,
		Comment:    comment,
		OperatedAt: now,
	}
	if err := e.historyRepo.Create(ctx, history); err != nil {
		logger.Warn("failed to create workflow history", zap.Error(err))
	}

	// 拒绝后，工作流实例状态变为 cancelled
	instance.Status = "cancelled"
	completedAt := time.Now()
	instance.CompletedAt = &completedAt

	if err := e.instanceRepo.Update(ctx, instance); err != nil {
		logger.Error("failed to update workflow instance", zap.Error(err))
		return fmt.Errorf("更新工作流实例失败: %w", err)
	}

	logger.Info("approval rejected",
		zap.Uint64("instance_id", instanceID),
		zap.Uint64("user_id", userID),
	)

	return nil
}

// GetHistory 获取流转历史
func (e *workflowEngine) GetHistory(ctx context.Context, instanceID uint64) ([]*dto.WorkflowHistoryResponse, error) {
	histories, err := e.historyRepo.ListByInstance(ctx, instanceID)
	if err != nil {
		return nil, fmt.Errorf("查询流转历史失败: %w", err)
	}

	responses := make([]*dto.WorkflowHistoryResponse, len(histories))
	for i, history := range histories {
		responses[i] = e.toHistoryResponse(ctx, history)
	}

	return responses, nil
}

// moveToNextNode 流转到下一个节点
func (e *workflowEngine) moveToNextNode(ctx context.Context, instance *model.WorkflowInstance, operatorID uint64) error {
	// 获取当前节点的出边
	edges, err := e.edgeRepo.GetOutgoingEdges(ctx, instance.CurrentNodeID)
	if err != nil {
		return fmt.Errorf("查询出边失败: %w", err)
	}

	if len(edges) == 0 {
		// 没有出边，可能是结束节点
		currentNode, _ := e.nodeRepo.GetByID(ctx, instance.CurrentNodeID)
		if currentNode != nil && currentNode.NodeType == "end" {
			instance.Status = "completed"
			completedAt := time.Now()
			instance.CompletedAt = &completedAt
			return e.instanceRepo.Update(ctx, instance)
		}
		return fmt.Errorf("当前节点没有出边")
	}

	// 简化处理：选择第一条边（实际应该根据条件表达式选择）
	nextEdge := edges[0]
	nextNode, err := e.nodeRepo.GetByID(ctx, nextEdge.TargetNodeID)
	if err != nil {
		return fmt.Errorf("查询下一节点失败: %w", err)
	}

	// 更新实例的当前节点
	fromNodeID := instance.CurrentNodeID
	instance.CurrentNodeID = nextNode.ID

	if err := e.instanceRepo.Update(ctx, instance); err != nil {
		return fmt.Errorf("更新工作流实例失败: %w", err)
	}

	// 记录流转历史
	history := &model.WorkflowHistory{
		InstanceID: instance.ID,
		FromNodeID: &fromNodeID,
		ToNodeID:   nextNode.ID,
		Action:     "forward",
		OperatorID: operatorID,
		Comment:    fmt.Sprintf("流转到节点: %s", nextNode.Name),
		OperatedAt: time.Now(),
	}
	if err := e.historyRepo.Create(ctx, history); err != nil {
		logger.Warn("failed to create workflow history", zap.Error(err))
	}

	// 如果是审批节点，创建审批记录
	if nextNode.NodeType == "approval" {
		if err := e.createApprovalRecords(ctx, instance, nextNode); err != nil {
			logger.Error("failed to create approval records", zap.Error(err))
			return fmt.Errorf("创建审批记录失败: %w", err)
		}
	}

	// 如果是结束节点，标记实例完成
	if nextNode.NodeType == "end" {
		instance.Status = "completed"
		completedAt := time.Now()
		instance.CompletedAt = &completedAt
		if err := e.instanceRepo.Update(ctx, instance); err != nil {
			return fmt.Errorf("更新工作流实例失败: %w", err)
		}
	}

	return nil
}

// createApprovalRecords 创建审批记录
func (e *workflowEngine) createApprovalRecords(ctx context.Context, instance *model.WorkflowInstance, node *model.WorkflowNode) error {
	// 解析节点配置
	var config dto.NodeConfig
	if node.Config != "" {
		if err := json.Unmarshal([]byte(node.Config), &config); err != nil {
			logger.Error("failed to parse node config", zap.Error(err))
			return fmt.Errorf("解析节点配置失败: %w", err)
		}
	}

	// 解析审批人
	approverIDs, err := e.resolveApprovers(ctx, instance, &config)
	if err != nil {
		return err
	}

	if len(approverIDs) == 0 {
		return ErrNoApproversConfigured
	}

	// 创建审批记录
	for _, approverID := range approverIDs {
		record := &model.ApprovalRecord{
			InstanceID: instance.ID,
			NodeID:     node.ID,
			ApproverID: approverID,
			Status:     "pending",
		}
		if err := e.approvalRepo.Create(ctx, record); err != nil {
			logger.Error("failed to create approval record", zap.Error(err))
			return fmt.Errorf("创建审批记录失败: %w", err)
		}
	}

	return nil
}

// resolveApprovers 解析审批人
func (e *workflowEngine) resolveApprovers(ctx context.Context, instance *model.WorkflowInstance, config *dto.NodeConfig) ([]uint64, error) {
	// 1. 直接指定用户
	if len(config.Approvers) > 0 {
		return config.Approvers, nil
	}

	// 2. 按项目角色解析
	if config.ApproverRole != "" {
		// 需要获取项目 ID，这里需要从 issue 获取
		// 简化处理：假设已经有项目 ID
		// 实际应该通过 issue 查询项目
		return nil, fmt.Errorf("角色解析功能待实现")
	}

	return nil, ErrNoApproversConfigured
}

// checkApprovalComplete 检查审批是否完成
func (e *workflowEngine) checkApprovalComplete(ctx context.Context, instanceID, nodeID uint64, approvalType string) (bool, error) {
	records, err := e.approvalRepo.ListByInstanceAndNode(ctx, instanceID, nodeID)
	if err != nil {
		return false, err
	}

	switch approvalType {
	case "single", "or_sign", "":
		// 任意一人通过即完成
		for _, r := range records {
			if r.Status == "approved" {
				return true, nil
			}
		}
		return false, nil

	case "countersign":
		// 所有人必须通过
		for _, r := range records {
			if r.Status != "approved" {
				return false, nil
			}
		}
		return len(records) > 0, nil

	default:
		return false, fmt.Errorf("未知的审批类型: %s", approvalType)
	}
}

// toInstanceResponse 转换为实例响应 DTO
func (e *workflowEngine) toInstanceResponse(ctx context.Context, instance *model.WorkflowInstance) (*dto.WorkflowInstanceResponse, error) {
	resp := &dto.WorkflowInstanceResponse{
		ID:            instance.ID,
		IssueID:       instance.IssueID,
		WorkflowID:    instance.WorkflowID,
		CurrentNodeID: instance.CurrentNodeID,
		Status:        instance.Status,
		StartedAt:     instance.StartedAt,
		CompletedAt:   instance.CompletedAt,
		CreatedAt:     instance.CreatedAt,
		UpdatedAt:     instance.UpdatedAt,
	}

	// 获取当前节点信息
	if currentNode, err := e.nodeRepo.GetByID(ctx, instance.CurrentNodeID); err == nil {
		resp.CurrentNode = e.toNodeResponse(currentNode)
	}

	// 获取工作流信息
	if workflow, err := e.workflowRepo.GetByID(ctx, instance.WorkflowID); err == nil {
		resp.WorkflowName = workflow.Name
	}

	// 获取审批记录
	if approvals, err := e.approvalRepo.ListByInstance(ctx, instance.ID); err == nil {
		resp.Approvals = make([]*dto.ApprovalRecordResponse, len(approvals))
		for i, approval := range approvals {
			resp.Approvals[i] = e.toApprovalResponse(ctx, approval)
		}
	}

	return resp, nil
}

// toHistoryResponse 转换为历史响应 DTO
func (e *workflowEngine) toHistoryResponse(ctx context.Context, history *model.WorkflowHistory) *dto.WorkflowHistoryResponse {
	resp := &dto.WorkflowHistoryResponse{
		ID:         history.ID,
		InstanceID: history.InstanceID,
		FromNodeID: history.FromNodeID,
		ToNodeID:   history.ToNodeID,
		Action:     history.Action,
		OperatorID: history.OperatorID,
		Comment:    history.Comment,
		OperatedAt: history.OperatedAt,
		CreatedAt:  history.CreatedAt,
	}

	// 获取操作人信息
	if history.OperatorID > 0 {
		if user, err := e.userRepo.GetByID(ctx, history.OperatorID); err == nil {
			resp.OperatorName = user.DisplayName
		}
	}

	// 获取节点信息
	if history.FromNodeID != nil {
		if node, err := e.nodeRepo.GetByID(ctx, *history.FromNodeID); err == nil {
			resp.FromNode = e.toNodeResponse(node)
		}
	}
	if node, err := e.nodeRepo.GetByID(ctx, history.ToNodeID); err == nil {
		resp.ToNode = e.toNodeResponse(node)
	}

	return resp
}

// toApprovalResponse 转换为审批记录响应 DTO
func (e *workflowEngine) toApprovalResponse(ctx context.Context, approval *model.ApprovalRecord) *dto.ApprovalRecordResponse {
	resp := &dto.ApprovalRecordResponse{
		ID:         approval.ID,
		InstanceID: approval.InstanceID,
		NodeID:     approval.NodeID,
		ApproverID: approval.ApproverID,
		Status:     approval.Status,
		Comment:    approval.Comment,
		ApprovedAt: approval.ApprovedAt,
		CreatedAt:  approval.CreatedAt,
		UpdatedAt:  approval.UpdatedAt,
	}

	// 获取审批人信息
	if user, err := e.userRepo.GetByID(ctx, approval.ApproverID); err == nil {
		resp.ApproverName = user.DisplayName
	}

	return resp
}

// toNodeResponse 转换为节点响应 DTO
func (e *workflowEngine) toNodeResponse(node *model.WorkflowNode) *dto.NodeResponse {
	resp := &dto.NodeResponse{
		ID:         node.ID,
		WorkflowID: node.WorkflowID,
		Name:       node.Name,
		NodeType:   node.NodeType,
		PositionX:  node.PositionX,
		PositionY:  node.PositionY,
		CreatedAt:  node.CreatedAt,
		UpdatedAt:  node.UpdatedAt,
	}

	// 解析配置
	if node.Config != "" {
		var config dto.NodeConfig
		if err := json.Unmarshal([]byte(node.Config), &config); err == nil {
			resp.Config = &config
		}
	}

	return resp
}
