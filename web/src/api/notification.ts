import request, { type ApiResponse, type PageResponse } from '@/utils/request'
import type { NotificationItem, NotificationListRequest, UnreadCountResponse } from '@/types/notification'

// 获取通知列表
export const getNotificationList = (params?: NotificationListRequest) => {
  return request.get<ApiResponse<PageResponse<NotificationItem>>>('/notifications', { params })
}

// 获取未读通知数量
export const getUnreadCount = () => {
  return request.get<ApiResponse<UnreadCountResponse>>('/notifications/unread-count')
}

// 标记通知为已读
export const markAsRead = (id: number) => {
  return request.put(`/notifications/${id}/read`)
}

// 全部标记为已读
export const markAllAsRead = () => {
  return request.put('/notifications/read-all')
}

// 删除通知
export const deleteNotification = (id: number) => {
  return request.delete(`/notifications/${id}`)
}
