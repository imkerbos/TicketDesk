// 工单相关类型定义

export type IssuePriority = 'P0' | 'P1' | 'P2' | 'P3'
export type IssueStatus = 'open' | 'in_progress' | 'resolved' | 'closed' | 'reopened' | 'merged' | 'pending_review'
export type IssueResolution = 'fixed' | 'wont_fix' | 'duplicate' | 'cannot_reproduce' | 'works_as_designed' | 'incomplete' | 'done' | ''

// 用户简要信息
export interface UserBrief {
  id: number
  username: string
  display_name: string
  avatar_url?: string
}

// 工单类型简要信息
export interface IssueTypeBrief {
  id: number
  name: string
  display_name: string
  icon: string
  color: string
}

export interface Issue {
  id: number
  issue_key: string
  project_id: number
  project_key: string
  issue_type_id: number
  issue_type?: IssueTypeBrief
  title: string
  description: string
  priority: IssuePriority
  status: IssueStatus
  resolution: IssueResolution
  assignee_id?: number
  assignee?: UserBrief
  reporter_id: number
  reporter?: UserBrief
  parent_id?: number
  parent_key?: string
  epic_id?: number
  epic_key?: string
  epic_title?: string
  merged_into_issue_id?: number
  merged_into_issue_key?: string
  merged_from_issue_keys?: string[]
  due_date?: string
  planned_start_date?: string
  planned_end_date?: string
  actual_start_date?: string
  actual_end_date?: string
  resolved_at?: string
  closed_at?: string
  created_at: string
  updated_at: string
  comments?: IssueComment[]
  watchers?: UserBrief[]
}

export interface IssueListRequest {
  page?: number
  page_size?: number
  project_key?: string
  status?: IssueStatus
  priority?: IssuePriority
  assignee_id?: number
  reporter_id?: number
  issue_type_id?: number
  epic_id?: number
  keyword?: string
  category?: '' | 'normal' | 'alert'
}

export interface IssueListResponse {
  items: Issue[]
  total: number
  page: number
  page_size: number
  has_more?: boolean // true 表示实际总数超过 total，显示 "10,000+"
}

// Custom field value for issue creation/update
export interface CustomFieldValue {
  field_id: number
  value: any
}

export interface CreateIssueRequest {
  project_key: string
  issue_type_id: number
  title: string
  description?: string
  priority: IssuePriority
  assignee_id?: number
  due_date?: string
  planned_start_date?: string
  planned_end_date?: string
  parent_id?: number
  epic_id?: number
  custom_fields?: CustomFieldValue[]
}

export interface UpdateIssueRequest {
  title?: string
  description?: string
  priority?: IssuePriority
  resolution?: IssueResolution
  assignee_id?: number
  epic_id?: number
  due_date?: string
  planned_start_date?: string
  planned_end_date?: string
  issue_type_id?: number
  custom_fields?: CustomFieldValue[]
}

export interface TransitionRequest {
  status: IssueStatus
  comment?: string
}

export interface IssueComment {
  id: number
  issue_id: number
  user_id: number
  user?: UserBrief
  content: string
  created_at: string
  updated_at: string
}

export interface CreateCommentRequest {
  content: string
}

export interface IssueType {
  id: number
  project_id?: number
  name: string
  display_name: string
  description: string
  icon: string
  color: string
  created_at: string
  updated_at: string
}

export interface IssueTransition {
  id: number
  name: string
  from_status: IssueStatus
  to_status: IssueStatus
}

export interface IssueActivity {
  id: number
  issue_id: number
  user_id: number
  user_name: string
  action: string
  field?: string
  old_value?: string
  new_value?: string
  details?: string
  created_at: string
}

export interface IssueWatcher {
  id: number
  user_id: number
  user?: UserBrief
}

// 工单统计
export interface IssueStats {
  total: number
  open: number
  in_progress: number
  resolved: number
  closed: number
  by_priority: Record<IssuePriority, number>
}

// 看板列
export interface KanbanColumn {
  status: IssueStatus
  label: string
  issues: Issue[]
}

// 工作日志
export interface Worklog {
  id: number
  issue_id: number
  user_id: number
  user?: UserBrief
  description: string
  time_spent: string
  time_spent_sec: number
  worked_at: string
  work_type?: string
  created_at: string
  updated_at: string
}

// 创建工作日志请求
export interface CreateWorklogRequest {
  description: string
  time_spent: string
  worked_at: string
  work_type?: string
}

// 更新工作日志请求
export interface UpdateWorklogRequest {
  description: string
  time_spent: string
  worked_at: string
  work_type?: string
}
