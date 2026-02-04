// 通知类型
export interface NotificationItem {
  id: number
  user_id: number
  type: string
  title: string
  content: string
  entity_type: string
  entity_id: number
  entity_key: string
  actor_id: number
  actor_name: string
  is_read: boolean
  read_at?: string
  created_at: string
}

// 通知列表请求
export interface NotificationListRequest {
  page?: number
  page_size?: number
  is_read?: boolean
  type?: string
}

// 未读数量响应
export interface UnreadCountResponse {
  count: number
}

// WebSocket 消息
export interface WSMessage {
  type: string
  data: unknown
}
