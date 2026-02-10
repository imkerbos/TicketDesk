// 工作流相关类型定义

// 节点类型
export type NodeType = 'start' | 'end' | 'approval' | 'work' | 'system'

// 审批类型
export type ApprovalType = 'single' | 'countersign' | 'or_sign'

// 指派人类型
export type AssigneeType = 'user' | 'role' | 'reporter' | 'project_lead'

// 节点配置
export interface NodeConfig {
  // 审批节点配置
  approval_type?: ApprovalType
  approvers?: number[]
  approver_role?: string

  // 工作节点配置
  assignee_type?: AssigneeType
  assignees?: number[]
  assignee_role?: string

  // 系统节点配置
  action?: string
  parameters?: Record<string, string>

  // 通用配置
  timeout_hours?: number
  description?: string
  target_status?: string // 进入该节点时工单目标状态（open, in_progress, resolved, closed）
}

// 工作流节点
export interface WorkflowNode {
  id: number
  workflow_id: number
  name: string
  node_type: NodeType
  config?: NodeConfig
  position_x: number
  position_y: number
  created_at: string
  updated_at: string
}

// 工作流边
export interface WorkflowEdge {
  id: number
  workflow_id: number
  source_node_id: number
  target_node_id: number
  condition_expr: string
  created_at: string
  updated_at: string
}

// 工作流
export interface Workflow {
  id: number
  project_id?: number
  name: string
  description: string
  status: number // 0: 禁用, 1: 启用
  nodes?: WorkflowNode[]
  edges?: WorkflowEdge[]
  created_at: string
  updated_at: string
}

// 工作流列表请求
export interface ListWorkflowsRequest {
  page?: number
  page_size?: number
  project_id?: number
  status?: number
  keyword?: string
}

// 创建工作流请求
export interface CreateWorkflowRequest {
  project_id?: number
  name: string
  description?: string
}

// 更新工作流请求
export interface UpdateWorkflowRequest {
  name?: string
  description?: string
  status?: number
}

// 创建节点请求
export interface CreateNodeRequest {
  name: string
  node_type: NodeType
  config?: NodeConfig
  position_x?: number
  position_y?: number
}

// 更新节点请求
export interface UpdateNodeRequest {
  name?: string
  config?: NodeConfig
  position_x?: number
  position_y?: number
}

// 创建边请求
export interface CreateEdgeRequest {
  source_node_id: number
  target_node_id: number
  condition_expr?: string
}

// 更新边请求
export interface UpdateEdgeRequest {
  condition_expr?: string
}

// ============ 工作流实例相关类型 ============

// 工作流实例状态
export type WorkflowInstanceStatus = 'active' | 'completed' | 'cancelled'

// 审批记录状态
export type ApprovalStatus = 'pending' | 'approved' | 'rejected'

// 工作流实例
export interface WorkflowInstance {
  id: number
  issue_id: number
  issue_key?: string
  workflow_id: number
  workflow_name?: string
  current_node_id: number
  current_node?: WorkflowNode
  status: WorkflowInstanceStatus
  started_at: string
  completed_at?: string
  approvals?: ApprovalRecord[]
  created_at: string
  updated_at: string
}

// 审批记录
export interface ApprovalRecord {
  id: number
  instance_id: number
  node_id: number
  approver_id: number
  approver_name?: string
  status: ApprovalStatus
  comment: string
  approved_at?: string
  created_at: string
  updated_at: string
}

// 流转历史
export interface WorkflowHistory {
  id: number
  instance_id: number
  from_node_id?: number
  from_node?: WorkflowNode
  to_node_id: number
  to_node?: WorkflowNode
  action: string
  operator_id: number
  operator_name?: string
  comment: string
  operated_at: string
  created_at: string
}

// 工作流方案
export interface WorkflowScheme {
  id: number
  project_id: number
  issue_type_id: number
  issue_type_name?: string
  workflow_id: number
  workflow_name?: string
  created_at: string
  updated_at: string
}

// ============ 工作流实例请求类型 ============

// 审批通过请求
export interface ApproveRequest {
  comment?: string
}

// 审批拒绝请求
export interface RejectRequest {
  comment: string
}

// 创建工作流方案请求
export interface CreateWorkflowSchemeRequest {
  issue_type_id: number
  workflow_id: number
}
