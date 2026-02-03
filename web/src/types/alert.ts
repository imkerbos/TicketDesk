// 告警相关类型定义

export interface Alert {
  id: number
  fingerprint: string
  source: string
  alert_name: string
  severity: 'critical' | 'warning' | 'info'
  status: 'firing' | 'resolved'
  labels: Record<string, string>
  annotations: Record<string, string>
  starts_at: string
  ends_at?: string
  issue_id?: number
  issue_key?: string
  ack_at?: string
  ack_by?: number
  ack_by_name?: string
  resolved_at?: string
  resolved_by?: number
  resolved_by_name?: string
  created_at: string
  updated_at: string
}

export interface AlertListRequest {
  page?: number
  page_size?: number
  status?: 'firing' | 'resolved'
  severity?: 'critical' | 'warning' | 'info'
  source?: string
  alert_name?: string
  issue_id?: number
}

export interface AlertListResponse {
  items: Alert[]
  total: number
  page: number
  page_size: number
}

export interface LabelMatcher {
  key: string
  operator: '==' | '!=' | '=~' | '!~'
  value: string
}

export interface AlertRule {
  id: number
  name: string
  description: string
  project_id: number
  project_key: string
  project_name: string
  issue_type_id: number
  issue_type_name: string
  label_matchers: LabelMatcher[]
  priority: 'P0' | 'P1' | 'P2' | 'P3'
  assignee_id?: number
  assignee_name?: string
  auto_resolve: boolean
  merge_window: number
  status: number
  created_at: string
  updated_at: string
}

export interface CreateAlertRuleRequest {
  name: string
  description?: string
  project_id: number
  issue_type_id: number
  label_matchers: LabelMatcher[]
  priority: 'P0' | 'P1' | 'P2' | 'P3'
  assignee_id?: number
  auto_resolve: boolean
  merge_window?: number
}

export interface AlertSilence {
  id: number
  name: string
  description: string
  label_matchers: LabelMatcher[]
  starts_at: string
  ends_at: string
  created_by: number
  created_by_name: string
  comment: string
  status: number // 0-已取消, 1-生效中, 2-已过期
  created_at: string
  updated_at: string
}

export interface CreateAlertSilenceRequest {
  name: string
  description?: string
  label_matchers: LabelMatcher[]
  starts_at: string
  ends_at: string
  comment?: string
}

export interface AlertGroupItem {
  group_value: string
  count: number
  severity: Record<string, number>
  status: Record<string, number>
}

export interface AlertGroupResponse {
  group_by: string
  items: AlertGroupItem[]
  total: number
}
