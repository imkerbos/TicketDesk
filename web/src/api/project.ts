import request, { type ApiResponse } from '@/utils/request'
import type {
  Project,
  ProjectListRequest,
  ProjectListResponse,
  CreateProjectRequest,
  UpdateProjectRequest,
  ProjectMember,
  AddProjectMemberRequest,
  ProjectIssueType,
  CreateIssueTypeRequest,
} from '@/types/project'

// 项目列表
export const getProjectList = (params?: ProjectListRequest) => {
  return request.get<ApiResponse<ProjectListResponse>>('/projects', { params })
}

// 项目详情
export const getProjectDetail = (key: string) => {
  return request.get<ApiResponse<Project>>(`/projects/${key}`)
}

// 创建项目
export const createProject = (data: CreateProjectRequest) => {
  return request.post<ApiResponse<Project>>('/projects', data)
}

// 更新项目
export const updateProject = (key: string, data: UpdateProjectRequest) => {
  return request.put<ApiResponse<Project>>(`/projects/${key}`, data)
}

// 删除项目
export const deleteProject = (key: string) => {
  return request.delete(`/projects/${key}`)
}

// 项目成员列表
export const getProjectMembers = (key: string) => {
  return request.get<ApiResponse<ProjectMember[]>>(`/projects/${key}/members`)
}

// 添加项目成员
export const addProjectMember = (key: string, data: AddProjectMemberRequest) => {
  return request.post<ApiResponse<ProjectMember>>(`/projects/${key}/members`, data)
}

// 移除项目成员
export const removeProjectMember = (key: string, userId: number) => {
  return request.delete(`/projects/${key}/members/${userId}`)
}

// 更新成员角色
export const updateProjectMemberRole = (key: string, userId: number, role: string) => {
  return request.put(`/projects/${key}/members/${userId}`, { role })
}

// 项目工单类型列表
export const getProjectIssueTypes = (key: string) => {
  return request.get<ApiResponse<ProjectIssueType[]>>(`/projects/${key}/issue-types`)
}

// 创建工单类型
export const createProjectIssueType = (key: string, data: CreateIssueTypeRequest) => {
  return request.post<ApiResponse<ProjectIssueType>>(`/projects/${key}/issue-types`, data)
}

// 更新工单类型
export const updateProjectIssueType = (key: string, typeId: number, data: Partial<CreateIssueTypeRequest>) => {
  return request.put<ApiResponse<ProjectIssueType>>(`/projects/${key}/issue-types/${typeId}`, data)
}

// 删除工单类型
export const deleteProjectIssueType = (key: string, typeId: number) => {
  return request.delete(`/projects/${key}/issue-types/${typeId}`)
}

// 获取所有项目（用于选择器，不分页）
export const getAllProjects = () => {
  return request.get<ApiResponse<Project[]>>('/projects/all')
}
