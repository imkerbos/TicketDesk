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
