import request, { type ApiResponse } from '@/utils/request'
import type {
  Issue,
  IssueListRequest,
  IssueListResponse,
  CreateIssueRequest,
  UpdateIssueRequest,
  IssueComment,
  CreateCommentRequest,
  IssueActivity,
  IssueWatcher,
  IssueStats,
} from '@/types/issue'

// 工单列表
export const getIssueList = (params: IssueListRequest) => {
  return request.get<ApiResponse<IssueListResponse>>('/issues', { params })
}

// 工单详情
export const getIssueDetail = (key: string) => {
  return request.get<ApiResponse<Issue>>(`/issues/${key}`)
}

// 创建工单
export const createIssue = (data: CreateIssueRequest) => {
  return request.post<ApiResponse<Issue>>('/issues', data)
}

// 更新工单
export const updateIssue = (key: string, data: UpdateIssueRequest) => {
  return request.put<ApiResponse<Issue>>(`/issues/${key}`, data)
}

// 删除工单
export const deleteIssue = (key: string) => {
  return request.delete(`/issues/${key}`)
}

// 工单状态流转
export const transitionIssue = (key: string, status: string, comment?: string) => {
  return request.post<ApiResponse<Issue>>(`/issues/${key}/transition`, { status, comment })
}

// 指派工单
export const assignIssue = (key: string, assigneeId: number) => {
  return request.post(`/issues/${key}/assign`, { assignee_id: assigneeId })
}

// 工单评论列表
export const getIssueComments = (key: string) => {
  return request.get<ApiResponse<IssueComment[]>>(`/issues/${key}/comments`)
}

// 添加评论
export const addIssueComment = (key: string, data: CreateCommentRequest) => {
  return request.post<ApiResponse<IssueComment>>(`/issues/${key}/comments`, data)
}

// 删除评论
export const deleteIssueComment = (key: string, commentId: number) => {
  return request.delete(`/issues/${key}/comments/${commentId}`)
}

// 工单活动记录
export const getIssueActivities = (key: string) => {
  return request.get<ApiResponse<IssueActivity[]>>(`/issues/${key}/activities`)
}

// 关注人列表
export const getIssueWatchers = (key: string) => {
  return request.get<ApiResponse<IssueWatcher[]>>(`/issues/${key}/watchers`)
}

// 添加关注人
export const addIssueWatcher = (key: string, userId: number) => {
  return request.post(`/issues/${key}/watchers`, { user_id: userId })
}

// 移除关注人
export const removeIssueWatcher = (key: string, userId: number) => {
  return request.delete(`/issues/${key}/watchers/${userId}`)
}

// 工单统计（用于 Dashboard）
export const getIssueStats = (params?: { project_id?: number; assignee_id?: number }) => {
  return request.get<ApiResponse<IssueStats>>('/issues/stats', { params })
}

// 我的待办工单
export const getMyTodoIssues = (params?: { page?: number; page_size?: number }) => {
  return request.get<ApiResponse<IssueListResponse>>('/issues/my-todo', { params })
}

// 我创建的工单
export const getMyCreatedIssues = (params?: { page?: number; page_size?: number }) => {
  return request.get<ApiResponse<IssueListResponse>>('/issues/my-created', { params })
}
