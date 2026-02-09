import request from '@/utils/request'
import type {
  Workflow,
  WorkflowNode,
  WorkflowEdge,
  ListWorkflowsRequest,
  CreateWorkflowRequest,
  UpdateWorkflowRequest,
  CreateNodeRequest,
  UpdateNodeRequest,
  CreateEdgeRequest,
  UpdateEdgeRequest,
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
