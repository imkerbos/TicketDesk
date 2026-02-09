import request, { type ApiResponse } from '@/utils/request'
import type {
  Alert,
  AlertListRequest,
  AlertListResponse,
  AlertStatsResponse,
  AlertRule,
  CreateAlertRuleRequest,
  AlertSilence,
  CreateAlertSilenceRequest,
  AlertGroupResponse,
  AlertDatasource,
  CreateDatasourceRequest,
  UpdateDatasourceRequest,
  TestDatasourceRequest,
  TestDatasourceResponse,
  DatasourceListResponse,
} from '@/types/alert'

// 告警列表
export const getAlertList = (params: AlertListRequest) => {
  return request.get<ApiResponse<AlertListResponse>>('/alerts', { params })
}

// 告警统计
export const getAlertStats = () => {
  return request.get<ApiResponse<AlertStatsResponse>>('/alerts/stats')
}

// 告警详情
export const getAlertDetail = (id: number) => {
  return request.get<ApiResponse<Alert>>(`/alerts/${id}`)
}

// 确认告警
export const ackAlert = (id: number, comment?: string) => {
  return request.post(`/alerts/${id}/ack`, { comment })
}

// 解决告警
export const resolveAlert = (id: number, comment?: string) => {
  return request.post(`/alerts/${id}/resolve`, { comment })
}

// 告警分组统计
export const getAlertGroup = (groupBy: string, params?: { status?: string; severity?: string }) => {
  return request.get<ApiResponse<AlertGroupResponse>>('/alerts/group', {
    params: { group_by: groupBy, ...params },
  })
}

// 告警规则列表
export const getAlertRuleList = (params?: { page?: number; page_size?: number; project_id?: number; status?: number }) => {
  return request.get<ApiResponse<{ items: AlertRule[]; total: number; page: number; page_size: number }>>('/alert-rules', { params })
}

// 创建告警规则
export const createAlertRule = (data: CreateAlertRuleRequest) => {
  return request.post<ApiResponse<AlertRule>>('/alert-rules', data)
}

// 获取告警规则详情
export const getAlertRuleDetail = (id: number) => {
  return request.get<ApiResponse<AlertRule>>(`/alert-rules/${id}`)
}

// 更新告警规则
export const updateAlertRule = (id: number, data: Partial<CreateAlertRuleRequest>) => {
  return request.put<ApiResponse<AlertRule>>(`/alert-rules/${id}`, data)
}

// 删除告警规则
export const deleteAlertRule = (id: number) => {
  return request.delete(`/alert-rules/${id}`)
}

// 告警静默列表
export const getAlertSilenceList = (params?: { page?: number; page_size?: number; status?: number }) => {
  return request.get<ApiResponse<{ items: AlertSilence[]; total: number; page: number; page_size: number }>>('/alert-silences', { params })
}

// 创建告警静默
export const createAlertSilence = (data: CreateAlertSilenceRequest) => {
  return request.post<ApiResponse<AlertSilence>>('/alert-silences', data)
}

// 获取告警静默详情
export const getAlertSilenceDetail = (id: number) => {
  return request.get<ApiResponse<AlertSilence>>(`/alert-silences/${id}`)
}

// 更新告警静默
export const updateAlertSilence = (id: number, data: Partial<CreateAlertSilenceRequest>) => {
  return request.put<ApiResponse<AlertSilence>>(`/alert-silences/${id}`, data)
}

// 删除告警静默
export const deleteAlertSilence = (id: number) => {
  return request.delete(`/alert-silences/${id}`)
}

// 取消告警静默
export const cancelAlertSilence = (id: number) => {
  return request.post(`/alert-silences/${id}/cancel`)
}

// ============ 告警数据源管理 ============

// 数据源列表
export const getDatasourceList = (params?: { page?: number; page_size?: number; type?: string; status?: number }) => {
  return request.get<ApiResponse<DatasourceListResponse>>('/alert-datasources', { params })
}

// 数据源详情
export const getDatasource = (id: number) => {
  return request.get<ApiResponse<AlertDatasource>>(`/alert-datasources/${id}`)
}

// 创建数据源
export const createDatasource = (data: CreateDatasourceRequest) => {
  return request.post<ApiResponse<AlertDatasource>>('/alert-datasources', data)
}

// 更新数据源
export const updateDatasource = (id: number, data: UpdateDatasourceRequest) => {
  return request.put<ApiResponse<AlertDatasource>>(`/alert-datasources/${id}`, data)
}

// 删除数据源
export const deleteDatasource = (id: number) => {
  return request.delete(`/alert-datasources/${id}`)
}

// 测试数据源连接（创建前）
export const testDatasource = (data: TestDatasourceRequest) => {
  return request.post<ApiResponse<TestDatasourceResponse>>('/alert-datasources/test', data)
}

// 测试已保存数据源的连接
export const testDatasourceById = (id: number) => {
  return request.post<ApiResponse<TestDatasourceResponse>>(`/alert-datasources/${id}/test`)
}
