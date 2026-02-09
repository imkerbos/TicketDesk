import request, { type ApiResponse } from '@/utils/request'
import type {
  FieldDefinition,
  FieldSchemeItem,
  FieldValue,
  ProjectVersion,
  ProjectComponent,
  IssueLabel,
  CreateFieldRequest,
  UpdateFieldRequest,
  UpdateFieldSchemeRequest,
  CreateVersionRequest,
  UpdateVersionRequest,
  CreateComponentRequest,
  UpdateComponentRequest,
  CreateLabelRequest,
  UpdateLabelRequest,
} from '@/types/field'

// ============ 字段定义 API ============

// 获取项目可用的所有字段
export function getFields(projectKey: string) {
  return request.get<ApiResponse<FieldDefinition[]>>(`/projects/${projectKey}/fields`)
}

// 创建自定义字段
export function createField(projectKey: string, data: CreateFieldRequest) {
  return request.post<ApiResponse<FieldDefinition>>(`/projects/${projectKey}/fields`, data)
}

// 更新字段
export function updateField(projectKey: string, fieldId: number, data: UpdateFieldRequest) {
  return request.put<ApiResponse<FieldDefinition>>(`/projects/${projectKey}/fields/${fieldId}`, data)
}

// 删除字段
export function deleteField(projectKey: string, fieldId: number) {
  return request.delete(`/projects/${projectKey}/fields/${fieldId}`)
}

// ============ 字段方案 API ============

// 获取工单类型的字段方案
export function getFieldScheme(projectKey: string, issueTypeId: number) {
  return request.get<ApiResponse<FieldSchemeItem[]>>(`/projects/${projectKey}/issue-types/${issueTypeId}/field-scheme`)
}

// 更新工单类型的字段方案
export function updateFieldScheme(projectKey: string, issueTypeId: number, data: UpdateFieldSchemeRequest) {
  return request.put<ApiResponse<FieldSchemeItem[]>>(`/projects/${projectKey}/issue-types/${issueTypeId}/field-scheme`, data)
}

// ============ 版本管理 API ============

// 获取项目版本列表
export function getVersions(projectKey: string) {
  return request.get<ApiResponse<ProjectVersion[]>>(`/projects/${projectKey}/versions`)
}

// 创建版本
export function createVersion(projectKey: string, data: CreateVersionRequest) {
  return request.post<ApiResponse<ProjectVersion>>(`/projects/${projectKey}/versions`, data)
}

// 更新版本
export function updateVersion(projectKey: string, versionId: number, data: UpdateVersionRequest) {
  return request.put<ApiResponse<ProjectVersion>>(`/projects/${projectKey}/versions/${versionId}`, data)
}

// 删除版本
export function deleteVersion(projectKey: string, versionId: number) {
  return request.delete(`/projects/${projectKey}/versions/${versionId}`)
}

// ============ 组件管理 API ============

// 获取项目组件列表
export function getComponents(projectKey: string) {
  return request.get<ApiResponse<ProjectComponent[]>>(`/projects/${projectKey}/components`)
}

// 创建组件
export function createComponent(projectKey: string, data: CreateComponentRequest) {
  return request.post<ApiResponse<ProjectComponent>>(`/projects/${projectKey}/components`, data)
}

// 更新组件
export function updateComponent(projectKey: string, componentId: number, data: UpdateComponentRequest) {
  return request.put<ApiResponse<ProjectComponent>>(`/projects/${projectKey}/components/${componentId}`, data)
}

// 删除组件
export function deleteComponent(projectKey: string, componentId: number) {
  return request.delete(`/projects/${projectKey}/components/${componentId}`)
}

// ============ 标签管理 API ============

// 获取项目标签列表
export function getLabels(projectKey: string) {
  return request.get<ApiResponse<IssueLabel[]>>(`/projects/${projectKey}/labels`)
}

// 创建标签
export function createLabel(projectKey: string, data: CreateLabelRequest) {
  return request.post<ApiResponse<IssueLabel>>(`/projects/${projectKey}/labels`, data)
}

// 更新标签
export function updateLabel(projectKey: string, labelId: number, data: UpdateLabelRequest) {
  return request.put<ApiResponse<IssueLabel>>(`/projects/${projectKey}/labels/${labelId}`, data)
}

// 删除标签
export function deleteLabel(projectKey: string, labelId: number) {
  return request.delete(`/projects/${projectKey}/labels/${labelId}`)
}

// ============ 工单字段值 API ============

// 获取工单的自定义字段值
export function getIssueFieldValues(issueId: number) {
  return request.get<ApiResponse<FieldValue[]>>(`/issue-field-values/${issueId}`)
}
