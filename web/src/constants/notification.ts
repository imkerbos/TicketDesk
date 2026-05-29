import type { NotificationEventType } from '@/types/project'

// 通知事件类型选项（用于渠道订阅配置）
export interface NotificationEventOption {
  value: NotificationEventType
  label: string
  description: string
}

export const NOTIFICATION_EVENT_OPTIONS: NotificationEventOption[] = [
  { value: 'issue.created', label: '工单创建', description: '新工单创建时通知' },
  { value: 'issue.transitioned', label: '工单流转', description: '工单状态变更/审批节点流转时通知' },
  { value: 'issue.assigned', label: '工单指派', description: '工单指派人变更时通知' },
  { value: 'alert.merged', label: '告警合并', description: '告警合并到已有工单时通知' },
]

export const DEFAULT_NOTIFICATION_EVENTS: NotificationEventType[] =
  NOTIFICATION_EVENT_OPTIONS.map((o) => o.value)

export function getNotificationEventLabel(value: string): string {
  return NOTIFICATION_EVENT_OPTIONS.find((o) => o.value === value)?.label || value
}
