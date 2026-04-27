import request, { type ApiResponse, type PageResponse } from '@/utils/request'
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
  IssueListStats,
  ProjectOverviewStats,
  Worklog,
  CreateWorklogRequest,
  UpdateWorklogRequest,
} from '@/types/issue'

// 工单列表
export const getIssueList = (params: IssueListRequest) => {
  return request.get<ApiResponse<IssueListResponse>>('/issues', { params })
}

// 工单详情
export const getIssueDetail = (key: string, config?: any) => {
  return request.get<ApiResponse<Issue>>(`/issues/${key}`, config)
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
  return request.get<ApiResponse<PageResponse<IssueActivity>>>('/activities', {
    params: {
      entity_type: 'issue',
      entity_key: key,
      page: 1,
      page_size: 100,
    },
  })
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

// 工单列表统计（用于全局工单列表 KPI 卡片）
export const getIssueListStats = (params: Record<string, any>) => {
  return request.get<ApiResponse<IssueListStats>>('/issues/stats', { params })
}

// 项目概述统计（按状态分组聚合，1 次请求替代 4 次列表请求）
export const getProjectOverviewStats = (projectKey: string) => {
  return request.get<ApiResponse<ProjectOverviewStats>>('/issues/project-overview-stats', {
    params: { project_key: projectKey },
  })
}

// 我的待办工单
export const getMyTodoIssues = (params?: { page?: number; page_size?: number }) => {
  return request.get<ApiResponse<IssueListResponse>>('/issues/my-todo', { params })
}

// 我创建的工单
export const getMyCreatedIssues = (params?: { page?: number; page_size?: number }) => {
  return request.get<ApiResponse<IssueListResponse>>('/issues/my-created', { params })
}

// 获取 Epic 下的所有 Issues
export const getEpicIssues = (epicKey: string) => {
  return request.get<ApiResponse<Issue[]>>(`/issues/${epicKey}/epic-issues`)
}

// 获取子任务列表
export const getSubtasks = (issueKey: string) => {
  return request.get<ApiResponse<Issue[]>>(`/issues/${issueKey}/subtasks`)
}

// 工作日志列表
export const getWorklogs = (issueKey: string) => {
  return request.get<ApiResponse<Worklog[]>>(`/issues/${issueKey}/worklogs`)
}

// 添加工作日志
export const addWorklog = (issueKey: string, data: CreateWorklogRequest) => {
  return request.post<ApiResponse<Worklog>>(`/issues/${issueKey}/worklogs`, data)
}

// 更新工作日志
export const updateWorklog = (issueKey: string, worklogId: number, data: UpdateWorklogRequest) => {
  return request.put<ApiResponse<Worklog>>(`/issues/${issueKey}/worklogs/${worklogId}`, data)
}

// 删除工作日志
export const deleteWorklog = (issueKey: string, worklogId: number) => {
  return request.delete(`/issues/${issueKey}/worklogs/${worklogId}`)
}
