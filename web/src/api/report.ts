import request from '@/utils/request'

// 获取仪表盘统计
export function getDashboardStats(params?: { project_key?: string }) {
  return request({
    url: '/reports/dashboard',
    method: 'get',
    params,
  })
}

// 获取工单统计
export function getIssueStats(params?: {
  project_key?: string
  start_date?: string
  end_date?: string
  group_by?: 'day' | 'week' | 'month'
}) {
  return request({
    url: '/reports/issues',
    method: 'get',
    params,
  })
}

// 获取 SLA 报表
export function getSLAReport(params: {
  project_key?: string
  start_date: string
  end_date: string
}) {
  return request({
    url: '/reports/sla',
    method: 'get',
    params,
  })
}

// 获取告警统计
export function getAlertStats(params?: {
  project_key?: string
  start_date?: string
  end_date?: string
  group_by?: 'day' | 'week' | 'month'
}) {
  return request({
    url: '/reports/alerts',
    method: 'get',
    params,
  })
}

// 获取用户绩效
export function getUserPerformance(params?: {
  project_key?: string
  start_date?: string
  end_date?: string
}) {
  return request({
    url: '/reports/user-performance',
    method: 'get',
    params,
  })
}
