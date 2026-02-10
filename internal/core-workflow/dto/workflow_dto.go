// Package dto 定义工作流模块的数据传输对象
package dto

import "time"

// ============ 请求 DTO ============

// CreateWorkflowRequest 创建工作流请求
type CreateWorkflowRequest struct {
	ProjectID   *uint64 `json:"project_id"`
	Name        string  `json:"name" binding:"required,min=1,max=100"`
	Description string  `json:"description" binding:"max=500"`
}

// UpdateWorkflowRequest 更新工作流请求
type UpdateWorkflowRequest struct {
	Name        *string `json:"name" binding:"omitempty,min=1,max=100"`
	Description *string `json:"description" binding:"omitempty,max=500"`
	Status      *int8   `json:"status" binding:"omitempty,oneof=0 1"`
}

// CreateNodeRequest 创建节点请求
type CreateNodeRequest struct {
	Name      string      `json:"name" binding:"required,min=1,max=100"`
	NodeType  string      `json:"node_type" binding:"required,oneof=start end approval work system"`
	Config    *NodeConfig `json:"config"`
	PositionX int         `json:"position_x"`
	PositionY int         `json:"position_y"`
}

// UpdateNodeRequest 更新节点请求
type UpdateNodeRequest struct {
	Name      *string     `json:"name" binding:"omitempty,min=1,max=100"`
	Config    *NodeConfig `json:"config"`
	PositionX *int        `json:"position_x"`
	PositionY *int        `json:"position_y"`
}

// NodeConfig 节点配置
type NodeConfig struct {
	// 审批节点配置
	ApprovalType string   `json:"approval_type,omitempty"` // single, countersign, or_sign
	Approvers    []uint64 `json:"approvers,omitempty"`     // 审批人 ID 列表
	ApproverRole string   `json:"approver_role,omitempty"` // 审批角色

	// 工作节点配置
	AssigneeType string   `json:"assignee_type,omitempty"` // user, role, reporter, project_lead
	Assignees    []uint64 `json:"assignees,omitempty"`     // 指派人 ID 列表
	AssigneeRole string   `json:"assignee_role,omitempty"` // 指派角色

	// 系统节点配置
	Action     string            `json:"action,omitempty"`      // 系统动作
	Parameters map[string]string `json:"parameters,omitempty"`  // 动作参数

	// 通用配置
	TimeoutHours int    `json:"timeout_hours,omitempty"` // 超时时间（小时）
	Description  string `json:"description,omitempty"`   // 节点说明
	TargetStatus string `json:"target_status,omitempty"` // 进入该节点时工单目标状态（open, in_progress, resolved, closed）
}

// CreateEdgeRequest 创建边请求
type CreateEdgeRequest struct {
	SourceNodeID  uint64 `json:"source_node_id" binding:"required"`
	TargetNodeID  uint64 `json:"target_node_id" binding:"required"`
	ConditionExpr string `json:"condition_expr" binding:"max=500"`
}

// UpdateEdgeRequest 更新边请求
type UpdateEdgeRequest struct {
	ConditionExpr *string `json:"condition_expr" binding:"omitempty,max=500"`
}

// ListWorkflowsRequest 工作流列表请求
type ListWorkflowsRequest struct {
	Page      int     `form:"page" binding:"omitempty,min=1"`
	PageSize  int     `form:"page_size" binding:"omitempty,min=1,max=100"`
	ProjectID *uint64 `form:"project_id"`
	Status    *int8   `form:"status" binding:"omitempty,oneof=0 1"`
	Keyword   string  `form:"keyword" binding:"omitempty,max=50"`
}

// GetDefaultPage 获取默认页码
func (r *ListWorkflowsRequest) GetDefaultPage() int {
	if r.Page <= 0 {
		return 1
	}
	return r.Page
}

// GetDefaultPageSize 获取默认每页数量
func (r *ListWorkflowsRequest) GetDefaultPageSize() int {
	if r.PageSize <= 0 {
		return 20
	}
	return r.PageSize
}

// ============ 响应 DTO ============

// WorkflowResponse 工作流响应
type WorkflowResponse struct {
	ID          uint64          `json:"id"`
	ProjectID   *uint64         `json:"project_id"`
	Name        string          `json:"name"`
	Description string          `json:"description"`
	Status      int8            `json:"status"`
	Nodes       []*NodeResponse `json:"nodes,omitempty"`
	Edges       []*EdgeResponse `json:"edges,omitempty"`
	CreatedAt   time.Time       `json:"created_at"`
	UpdatedAt   time.Time       `json:"updated_at"`
}

// WorkflowBrief 工作流简要信息
type WorkflowBrief struct {
	ID          uint64  `json:"id"`
	ProjectID   *uint64 `json:"project_id"`
	Name        string  `json:"name"`
	Description string  `json:"description"`
	Status      int8    `json:"status"`
}

// NodeResponse 节点响应
type NodeResponse struct {
	ID         uint64      `json:"id"`
	WorkflowID uint64      `json:"workflow_id"`
	Name       string      `json:"name"`
	NodeType   string      `json:"node_type"`
	Config     *NodeConfig `json:"config,omitempty"`
	PositionX  int         `json:"position_x"`
	PositionY  int         `json:"position_y"`
	CreatedAt  time.Time   `json:"created_at"`
	UpdatedAt  time.Time   `json:"updated_at"`
}

// EdgeResponse 边响应
type EdgeResponse struct {
	ID            uint64    `json:"id"`
	WorkflowID    uint64    `json:"workflow_id"`
	SourceNodeID  uint64    `json:"source_node_id"`
	TargetNodeID  uint64    `json:"target_node_id"`
	ConditionExpr string    `json:"condition_expr"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

// ============ 工作流实例相关 DTO ============

// ApproveRequest 审批请求
type ApproveRequest struct {
	Comment string `json:"comment" binding:"max=1000"`
}

// RejectRequest 拒绝请求
type RejectRequest struct {
	Comment string `json:"comment" binding:"required,max=1000"`
}

// WorkflowInstanceResponse 工作流实例响应
type WorkflowInstanceResponse struct {
	ID            uint64                    `json:"id"`
	IssueID       uint64                    `json:"issue_id"`
	IssueKey      string                    `json:"issue_key,omitempty"`
	WorkflowID    uint64                    `json:"workflow_id"`
	WorkflowName  string                    `json:"workflow_name,omitempty"`
	CurrentNodeID uint64                    `json:"current_node_id"`
	CurrentNode   *NodeResponse             `json:"current_node,omitempty"`
	Status        string                    `json:"status"`
	StartedAt     time.Time                 `json:"started_at"`
	CompletedAt   *time.Time                `json:"completed_at,omitempty"`
	Approvals     []*ApprovalRecordResponse `json:"approvals,omitempty"`
	CreatedAt     time.Time                 `json:"created_at"`
	UpdatedAt     time.Time                 `json:"updated_at"`
}

// WorkflowHistoryResponse 流转历史响应
type WorkflowHistoryResponse struct {
	ID           uint64        `json:"id"`
	InstanceID   uint64        `json:"instance_id"`
	FromNodeID   *uint64       `json:"from_node_id,omitempty"`
	FromNode     *NodeResponse `json:"from_node,omitempty"`
	ToNodeID     uint64        `json:"to_node_id"`
	ToNode       *NodeResponse `json:"to_node,omitempty"`
	Action       string        `json:"action"`
	OperatorID   uint64        `json:"operator_id"`
	OperatorName string        `json:"operator_name,omitempty"`
	Comment      string        `json:"comment"`
	OperatedAt   time.Time     `json:"operated_at"`
	CreatedAt    time.Time     `json:"created_at"`
}

// ApprovalRecordResponse 审批记录响应
type ApprovalRecordResponse struct {
	ID           uint64     `json:"id"`
	InstanceID   uint64     `json:"instance_id"`
	NodeID       uint64     `json:"node_id"`
	ApproverID   uint64     `json:"approver_id"`
	ApproverName string     `json:"approver_name,omitempty"`
	Status       string     `json:"status"`
	Comment      string     `json:"comment"`
	ApprovedAt   *time.Time `json:"approved_at,omitempty"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
}

// WorkflowSchemeResponse 工作流方案响应
type WorkflowSchemeResponse struct {
	ID            uint64    `json:"id"`
	ProjectID     uint64    `json:"project_id"`
	IssueTypeID   uint64    `json:"issue_type_id"`
	IssueTypeName string    `json:"issue_type_name,omitempty"`
	WorkflowID    uint64    `json:"workflow_id"`
	WorkflowName  string    `json:"workflow_name,omitempty"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

// CreateWorkflowSchemeRequest 创建工作流方案请求
type CreateWorkflowSchemeRequest struct {
	IssueTypeID uint64 `json:"issue_type_id" binding:"required"`
	WorkflowID  uint64 `json:"workflow_id" binding:"required"`
}

