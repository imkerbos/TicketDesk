import request, { type ApiResponse, type PageResponse } from '@/utils/request'
import type {
  RequirementPool,
  RequirementPoolListRequest,
  CreateRequirementPoolRequest,
  UpdateRequirementPoolRequest,
  Requirement,
  RequirementListRequest,
  CreateRequirementRequest,
  UpdateRequirementRequest,
  ConvertToIssueRequest,
  RequirementComment,
  RequirementCommentRequest,
  KanbanRequest,
  KanbanResponse,
  ReportRequest,
  ReportResponse,
  RequirementCategoryDef,
  CreateCategoryRequest,
  UpdateCategoryRequest,
} from '@/types/requirement'

// ============ 需求分类 API ============

// 获取需求分类列表
export const getRequirementCategories = () => {
  return request.get<ApiResponse<RequirementCategoryDef[]>>('/requirement-categories')
}

// 创建需求分类
export const createRequirementCategory = (data: CreateCategoryRequest) => {
  return request.post<ApiResponse<RequirementCategoryDef>>('/requirement-categories', data)
}

// 更新需求分类
export const updateRequirementCategory = (id: number, data: UpdateCategoryRequest) => {
  return request.put<ApiResponse<void>>(`/requirement-categories/${id}`, data)
}

// 删除需求分类
export const deleteRequirementCategory = (id: number) => {
  return request.delete(`/requirement-categories/${id}`)
}

// ============ 需求池 API ============

// 获取需求池列表
export const getRequirementPoolList = (params: RequirementPoolListRequest) => {
  return request.get<ApiResponse<PageResponse<RequirementPool>>>('/requirement-pools', { params })
}

// 获取需求池详情
export const getRequirementPoolDetail = (id: number) => {
  return request.get<ApiResponse<RequirementPool>>(`/requirement-pools/${id}`)
}

// 创建需求池
export const createRequirementPool = (data: CreateRequirementPoolRequest) => {
  return request.post<ApiResponse<RequirementPool>>('/requirement-pools', data)
}

// 更新需求池
export const updateRequirementPool = (id: number, data: UpdateRequirementPoolRequest) => {
  return request.put<ApiResponse<void>>(`/requirement-pools/${id}`, data)
}

// 删除需求池
export const deleteRequirementPool = (id: number) => {
  return request.delete(`/requirement-pools/${id}`)
}

// ============ 需求 API ============

// 获取需求列表
export const getRequirementList = (params: RequirementListRequest) => {
  return request.get<ApiResponse<PageResponse<Requirement>>>('/requirements', { params })
}

// 获取需求详情
export const getRequirementDetail = (id: number) => {
  return request.get<ApiResponse<Requirement>>(`/requirements/${id}`)
}

// 创建需求
export const createRequirement = (data: CreateRequirementRequest) => {
  return request.post<ApiResponse<Requirement>>('/requirements', data)
}

// 更新需求
export const updateRequirement = (id: number, data: UpdateRequirementRequest) => {
  return request.put<ApiResponse<void>>(`/requirements/${id}`, data)
}

// 删除需求
export const deleteRequirement = (id: number) => {
  return request.delete(`/requirements/${id}`)
}

// 转化为工单
export const convertToIssue = (id: number, data: ConvertToIssueRequest) => {
  return request.post<ApiResponse<{ message: string; issue_id: number; issue_key: string }>>(`/requirements/${id}/convert`, data)
}

// 添加评论
export const addRequirementComment = (id: number, data: RequirementCommentRequest) => {
  return request.post<ApiResponse<RequirementComment>>(`/requirements/${id}/comments`, data)
}

// 获取看板数据
export const getRequirementKanban = (params: KanbanRequest) => {
  return request.get<ApiResponse<KanbanResponse>>('/requirements/kanban', { params })
}

// 获取报告数据
export const getRequirementReport = (params: ReportRequest) => {
  return request.get<ApiResponse<ReportResponse>>('/requirements/report', { params })
}
