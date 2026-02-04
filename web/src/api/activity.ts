import request, { type ApiResponse, type PageResponse } from '@/utils/request'

// 活动响应类型
export interface Activity {
  id: number
  user_id: number
  user_name: string
  action: string
  entity_type: string
  entity_id: number
  entity_key?: string
  details?: string
  created_at: string
}

// 活动列表请求参数
export interface ActivityListRequest {
  page?: number
  page_size?: number
  entity_type?: string
  entity_id?: number
  user_id?: number
}

// 获取活动列表
export const getActivityList = (params?: ActivityListRequest) => {
  return request.get<ApiResponse<PageResponse<Activity>>>('/activities', { params })
}

// 获取最近活动
export const getRecentActivities = (limit?: number) => {
  return request.get<ApiResponse<Activity[]>>('/activities/recent', { params: { limit } })
}
