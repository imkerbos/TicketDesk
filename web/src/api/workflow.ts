import request from '@/utils/request'
import type {
  Workflow,
  WorkflowNode,
  WorkflowEdge,
  WorkflowInstance,
  WorkflowHistory,
  WorkflowScheme,
  ListWorkflowsRequest,
  CreateWorkflowRequest,
  UpdateWorkflowRequest,
  CreateNodeRequest,
  UpdateNodeRequest,
  CreateEdgeRequest,
  UpdateEdgeRequest,
  ApproveRequest,
  RejectRequest,
  CreateWorkflowSchemeRequest,
} from '@/types/workflow'

// ============ 工作流管理 ============

/**
 * 获取工作流列表
 */
export const getWorkflowList = (params?: ListWorkflowsRequest) => {
  return request.get<{
    items: Workflow[]
    total: number
    page: number
    page_size: number
  }>('/workflows', { params })
}

/**
 * 获取工作流详情
 */
export const getWorkflowDetail = (id: number) => {
  return request.get<Workflow>(`/workflows/${id}`)
}

/**
 * 创建工作流
 */
export const createWorkflow = (data: CreateWorkflowRequest) => {
  return request.post<Workflow>('/workflows', data)
}

/**
 * 更新工作流
 */
export const updateWorkflow = (id: number, data: UpdateWorkflowRequest) => {
  return request.put<Workflow>(`/workflows/${id}`, data)
}

/**
 * 删除工作流
 */
export const deleteWorkflow = (id: number) => {
  return request.delete(`/workflows/${id}`)
}

// ============ 节点管理 ============

/**
 * 获取工作流的所有节点
 */
export const getWorkflowNodes = (workflowId: number) => {
  return request.get<WorkflowNode[]>(`/workflows/${workflowId}/nodes`)
}

/**
 * 获取节点详情
 */
export const getNodeDetail = (workflowId: number, nodeId: number) => {
  return request.get<WorkflowNode>(`/workflows/${workflowId}/nodes/${nodeId}`)
}

/**
 * 创建节点
 */
export const createNode = (workflowId: number, data: CreateNodeRequest) => {
  return request.post<WorkflowNode>(`/workflows/${workflowId}/nodes`, data)
}

/**
 * 更新节点
 */
export const updateNode = (workflowId: number, nodeId: number, data: UpdateNodeRequest) => {
  return request.put<WorkflowNode>(`/workflows/${workflowId}/nodes/${nodeId}`, data)
}

/**
 * 删除节点
 */
export const deleteNode = (workflowId: number, nodeId: number) => {
  return request.delete(`/workflows/${workflowId}/nodes/${nodeId}`)
}

// ============ 边管理 ============

/**
 * 获取工作流的所有边
 */
export const getWorkflowEdges = (workflowId: number) => {
  return request.get<WorkflowEdge[]>(`/workflows/${workflowId}/edges`)
}

/**
 * 创建边
 */
export const createEdge = (workflowId: number, data: CreateEdgeRequest) => {
  return request.post<WorkflowEdge>(`/workflows/${workflowId}/edges`, data)
}

/**
 * 更新边
 */
export const updateEdge = (workflowId: number, edgeId: number, data: UpdateEdgeRequest) => {
  return request.put<WorkflowEdge>(`/workflows/${workflowId}/edges/${edgeId}`, data)
}

/**
 * 删除边
 */
export const deleteEdge = (workflowId: number, edgeId: number) => {
  return request.delete(`/workflows/${workflowId}/edges/${edgeId}`)
}

// ============ 工作流实例管理 ============

/**
 * 获取工单的工作流实例
 */
export const getWorkflowInstance = (issueKey: string) => {
  return request.get<WorkflowInstance>(`/issues/${issueKey}/workflow`, { _silent: true } as any)
}

/**
 * 审批通过
 */
export const approveWorkflow = (issueKey: string, data: ApproveRequest) => {
  return request.post(`/issues/${issueKey}/workflow/approve`, data)
}

/**
 * 审批拒绝
 */
export const rejectWorkflow = (issueKey: string, data: RejectRequest) => {
  return request.post(`/issues/${issueKey}/workflow/reject`, data)
}

/**
 * 完成工作节点
 */
export const completeWorkflow = (issueKey: string, data: ApproveRequest) => {
  return request.post(`/issues/${issueKey}/workflow/complete`, data)
}

/**
 * 获取流转历史
 */
export const getWorkflowHistory = (issueKey: string) => {
  return request.get<WorkflowHistory[]>(`/issues/${issueKey}/workflow/history`, { _silent: true } as any)
}

// ============ 工作流方案管理 ============

/**
 * 获取项目的工作流方案列表
 */
export const getWorkflowSchemes = (projectKey: string) => {
  return request.get<WorkflowScheme[]>(`/projects/${projectKey}/workflow-schemes`)
}

/**
 * 创建工作流方案
 */
export const createWorkflowScheme = (projectKey: string, data: CreateWorkflowSchemeRequest) => {
  return request.post<WorkflowScheme>(`/projects/${projectKey}/workflow-schemes`, data)
}

/**
 * 删除工作流方案
 */
export const deleteWorkflowScheme = (projectKey: string, typeId: number) => {
  return request.delete(`/projects/${projectKey}/workflow-schemes/${typeId}`)
}
