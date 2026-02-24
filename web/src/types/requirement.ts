// 需求池类型
export type RequirementPoolType = 'global' | 'project'

// 需求池状态
export type RequirementPoolStatus = 'active' | 'archived'

// 需求状态
export type RequirementStatus = 'pending_review' | 'planning' | 'in_progress' | 'completed' | 'on_hold' | 'rejected'

// 需求优先级
export type RequirementPriority = 'P0' | 'P1' | 'P2' | 'P3'

// 需求分类
export type RequirementCategory = 'feature' | 'optimization' | 'bugfix' | 'security' | 'infrastructure' | 'other'

// 需求池
export interface RequirementPool {
  id: number
  name: string
  description: string
  type: RequirementPoolType
  project_id?: number
  project_name?: string
  owner_id: number
  owner_name: string
  status: RequirementPoolStatus
  requirement_count: number
  created_at: string
  updated_at: string
}

// 需求
export interface Requirement {
  id: number
  pool_id: number
  pool_name: string
  title: string
  description: string
  priority: RequirementPriority
  status: RequirementStatus
  category: RequirementCategory
  reporter_id?: number
  reporter_name?: string
  assignee_id?: number
  assignee_name?: string
  start_date?: string
  end_date?: string
  progress?: string
  result?: string
  converted_issue_id?: number
  converted_issue_key?: string
  converted_issue_status?: string
  target_project_id?: number
  target_project_name?: string
  created_by: number
  creator_name: string
  tags: string[]
  comment_count: number
  created_at: string
  updated_at: string
}

// 需求评论
export interface RequirementComment {
  id: number
  requirement_id: number
  user_id: number
  user_name: string
  content: string
  created_at: string
}

// 看板列
export interface KanbanColumn {
  key: string
  title: string
  count: number
  requirements: Requirement[]
}

// 看板响应
export interface KanbanResponse {
  group_by: string
  columns: KanbanColumn[]
  total: number
}

// 状态统计
export interface StatusSummary {
  status: RequirementStatus
  count: number
}

// 优先级统计
export interface PrioritySummary {
  priority: RequirementPriority
  count: number
}

// 负责人统计
export interface AssigneeSummary {
  assignee_id: number
  assignee_name: string
  total: number
  completed: number
  in_progress: number
}

// 趋势数据
export interface TrendData {
  date: string
  created: number
  completed: number
}

// 报告响应
export interface ReportResponse {
  pool_id?: number
  pool_name?: string
  start_date: string
  end_date: string
  total_created: number
  total_completed: number
  total_rejected: number
  avg_process_days: number
  status_summary: StatusSummary[]
  priority_summary: PrioritySummary[]
  assignee_summary: AssigneeSummary[]
  trend_data: TrendData[]
}

// ============ 请求类型 ============

// 需求池列表请求
export interface RequirementPoolListRequest {
  type?: RequirementPoolType
  status?: RequirementPoolStatus
  project_id?: number
  owner_id?: number
  page?: number
  page_size?: number
}

// 创建需求池请求
export interface CreateRequirementPoolRequest {
  name: string
  description?: string
  type: RequirementPoolType
  project_id?: number
  owner_id?: number
}

// 更新需求池请求
export interface UpdateRequirementPoolRequest {
  name?: string
  description?: string
  owner_id?: number
  status?: RequirementPoolStatus
}

// 需求列表请求
export interface RequirementListRequest {
  pool_id?: number
  status?: RequirementStatus
  priority?: RequirementPriority
  category?: RequirementCategory
  assignee_id?: number
  created_by?: number
  start_date?: string
  end_date?: string
  keyword?: string
  page?: number
  page_size?: number
}

// 创建需求请求
export interface CreateRequirementRequest {
  pool_id?: number
  title: string
  description?: string
  priority: RequirementPriority
  category: RequirementCategory
  reporter_id?: number
  assignee_id?: number
  start_date?: string
  end_date?: string
  target_project_id?: number
  tags?: string[]
}

// 更新需求请求
export interface UpdateRequirementRequest {
  title?: string
  description?: string
  priority?: RequirementPriority
  status?: RequirementStatus
  category?: RequirementCategory
  reporter_id?: number
  assignee_id?: number
  start_date?: string
  end_date?: string
  progress?: string
  result?: string
  target_project_id?: number
  tags?: string[]
}

// 转化为工单请求
export interface ConvertToIssueRequest {
  project_key: string
  issue_type_id: number
  assignee_id?: number
}

// 需求评论请求
export interface RequirementCommentRequest {
  content: string
}

// 看板请求
export interface KanbanRequest {
  pool_id?: number
  group_by: 'status' | 'priority' | 'assignee' | 'timeline'
  start_date?: string
  end_date?: string
}

// 报告请求
export interface ReportRequest {
  pool_id?: number
  start_date: string
  end_date: string
  group_by?: 'day' | 'week' | 'month'
}
