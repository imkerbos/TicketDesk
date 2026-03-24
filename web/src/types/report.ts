// 仪表盘统计响应
export interface DashboardStats {
  issue_stats: {
    total_open: number
    total_in_progress: number
    total_resolved: number
    total_closed: number
    today_created: number
    week_created: number
    week_resolved: number
  }
  alert_stats: {
    total_firing: number
    total_acked: number
    total_resolved: number
    today_created: number
    week_created: number
  }
  project_stats: {
    total_projects: number
    total_members: number
  }
}

// 工单统计响应
export interface IssueStats {
  summary: {
    total: number
    open: number
    in_progress: number
    resolved: number
    closed: number
    avg_resolve_time: number
  }
  timeline: TimelineItem[]
  priority_distribution: DistributionItem[]
  type_distribution: DistributionItem[]
  status_distribution: DistributionItem[]
  assignee_distribution: DistributionItem[]
  epic_distribution: DistributionItem[]
}

// 时间线统计项
export interface TimelineItem {
  date: string
  created: number
  in_progress: number
  resolved: number
  closed: number
}

// 分布统计项
export interface DistributionItem {
  name: string
  value: number
  ratio: number
}

// SLA 报表响应
export interface SLAReport {
  summary: {
    total_issues: number
    resolved_issues: number
    mtta: number
    mttr: number
    sla_met: number
    sla_violated: number
    sla_rate: number
  }
  by_priority: SLAPriorityStats[]
  by_project: SLAProjectStats[]
  violations: SLAViolation[]
}

// SLA 违规工单
export interface SLAViolation {
  issue_key: string
  title: string
  priority: string
  assignee_name: string
  sla_target: number
  actual_time: number
  overdue_by: number
}

// 按优先级的 SLA 统计
export interface SLAPriorityStats {
  priority: string
  total: number
  resolved: number
  mtta: number
  mttr: number
  sla_target: number
  sla_met: number
  sla_rate: number
}

// 按项目的 SLA 统计
export interface SLAProjectStats {
  project_key: string
  project_name: string
  total: number
  resolved: number
  mttr: number
  sla_rate: number
}

// 告警统计响应
export interface AlertStats {
  summary: {
    total: number
    firing: number
    acked: number
    resolved: number
    avg_ack_time: number
  }
  timeline: AlertTimelineItem[]
  severity_distribution: DistributionItem[]
  top_alerts: TopAlertItem[]
}

// 告警时间线统计项
export interface AlertTimelineItem {
  date: string
  firing: number
  resolved: number
}

// 告警排名项
export interface TopAlertItem {
  alert_name: string
  severity: string
  count: number
}

// 用户绩效响应
export interface UserPerformance {
  user_id: number
  username: string
  display_name: string
  assigned: number
  resolved: number
  avg_resolve_time: number
}

// 工时统计响应
export interface WorklogStats {
  summary: WorklogSummary
  daily_stats: DailyWorklogStat[]
  user_stats: UserWorklogStat[]
  type_stats: WorklogTypeStat[]
  grid: WorklogGridRow[]
  grid_dates: string[]
  grid_totals: Record<string, number>
}

// 工时网格行
export interface WorklogGridRow {
  user_id: number
  display_name: string
  total_sec: number
  daily: Record<string, number>
  daily_details: Record<string, WorklogDetail[]>
}

// 工单工时明细
export interface WorklogDetail {
  issue_key: string
  title: string
  time_sec: number
}

export interface WorklogSummary {
  total_time_sec: number
  total_entries: number
  active_users: number
  avg_daily_time_sec: number
}

export interface DailyWorklogStat {
  date: string
  total_time_sec: number
  entry_count: number
}

export interface UserWorklogStat {
  user_id: number
  display_name: string
  total_time_sec: number
  entry_count: number
}

export interface WorklogTypeStat {
  work_type: string
  total_time_sec: number
  entry_count: number
}
