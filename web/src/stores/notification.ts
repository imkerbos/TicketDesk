import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import axios from 'axios'
import { ElNotification } from 'element-plus'
import type { NotificationItem, WSMessage } from '@/types/notification'
import {
  getNotificationList,
  getUnreadCount as fetchUnreadCount,
  markAsRead as apiMarkAsRead,
  markAllAsRead as apiMarkAllAsRead,
} from '@/api/notification'

export const useNotificationStore = defineStore('notification', () => {
  // 状态
  const notifications = ref<NotificationItem[]>([])
  const unreadCount = ref(0)
  const ws = ref<WebSocket | null>(null)
  const connected = ref(false)
  const total = ref(0)
  let heartbeatTimer: ReturnType<typeof setInterval> | null = null
  let reconnectTimer: ReturnType<typeof setTimeout> | null = null
  let reconnectAttempts = 0
  const MAX_RECONNECT_DELAY = 60000 // 最大重连间隔 60 秒
  const MAX_RECONNECT_FAILURES = 5  // 连续失败上限，超过后停止重连

  // 检查 JWT 是否过期（提前 60 秒判定，留网络延迟和时钟偏差余量）
  const isTokenExpired = (token: string): boolean => {
    try {
      const payload = JSON.parse(atob(token.split('.')[1]))
      return payload.exp * 1000 < Date.now() + 60000
    } catch {
      return true
    }
  }

  // 刷新 token
  const refreshAccessToken = async (): Promise<string | null> => {
    const refreshToken = localStorage.getItem('refresh_token')
    if (!refreshToken) return null

    try {
      const res = await axios.post('/api/v1/auth/refresh', { refresh_token: refreshToken })
      const { access_token, refresh_token: newRefreshToken } = res.data.data
      localStorage.setItem('token', access_token)
      if (newRefreshToken) {
        localStorage.setItem('refresh_token', newRefreshToken)
      }
      return access_token
    } catch {
      return null
    }
  }

  // 确保获取有效 token（forceRefresh 为 true 时跳过本地过期检查，直接刷新）
  const ensureFreshToken = async (forceRefresh = false): Promise<string | null> => {
    const token = localStorage.getItem('token')
    if (!token) return null

    // 不强制刷新且 token 未过期，直接返回
    if (!forceRefresh && !isTokenExpired(token)) return token

    // token 已过期或强制刷新（如 WS 连接被服务端拒绝）
    return await refreshAccessToken()
  }

  // 计算属性
  const recentNotifications = computed(() => notifications.value.slice(0, 10))

  // 获取通知列表
  const fetchNotifications = async (page = 1, pageSize = 20) => {
    try {
      const res = await getNotificationList({ page, page_size: pageSize })
      const data = res.data.data
      if (page === 1) {
        notifications.value = data.items || []
      } else {
        notifications.value.push(...(data.items || []))
      }
      total.value = data.total
    } catch {
      // 静默处理
    }
  }

  // 获取未读数量
  const fetchUnreadCountData = async () => {
    try {
      const res = await fetchUnreadCount()
      unreadCount.value = res.data.data.count
    } catch {
      // 静默处理
    }
  }

  // 标记为已读
  const markAsRead = async (id: number) => {
    try {
      await apiMarkAsRead(id)
      const notif = notifications.value.find((n) => n.id === id)
      if (notif && !notif.is_read) {
        notif.is_read = true
        unreadCount.value = Math.max(0, unreadCount.value - 1)
      }
    } catch {
      // 静默处理
    }
  }

  // 全部标记为已读
  const markAllAsRead = async () => {
    try {
      await apiMarkAllAsRead()
      notifications.value.forEach((n) => {
        n.is_read = true
      })
      unreadCount.value = 0
    } catch {
      // 静默处理
    }
  }

  // 请求通知权限
  const requestNotificationPermission = async () => {
    if (!('Notification' in window)) {
      return false
    }

    if (Notification.permission === 'granted') {
      return true
    }

    if (Notification.permission !== 'denied') {
      const permission = await Notification.requestPermission()
      return permission === 'granted'
    }

    return false
  }

  // 显示浏览器通知
  const showBrowserNotification = (notif: NotificationItem) => {
    if (!('Notification' in window)) {
      return
    }

    if (window.Notification.permission !== 'granted') {
      return
    }

    try {
      const browserNotif = new window.Notification(notif.title, {
        body: notif.content,
        icon: '/favicon.ico',
        badge: '/favicon.ico',
        tag: `notification-${notif.id}`,
        requireInteraction: true, // 持续显示，不自动关闭
        silent: false,
      })

      // 点击通知时跳转到对应页面并标记为已读
      browserNotif.onclick = () => {
        markAsRead(notif.id) // 标记为已读
        window.focus()
        if (notif.entity_type === 'issue' && notif.entity_key) {
          window.location.href = `/issues/${notif.entity_key}`
        }
        browserNotif.close()
      }

      // 监听页面可见性变化，当用户切换回标签页时，5秒后自动关闭通知
      const handleVisibilityChange = () => {
        if (!document.hidden) {
          // 用户切换回标签页，5秒后关闭通知
          setTimeout(() => {
            browserNotif.close()
            document.removeEventListener('visibilitychange', handleVisibilityChange)
          }, 5000)
        }
      }

      document.addEventListener('visibilitychange', handleVisibilityChange)

      // 通知关闭时清理事件监听
      browserNotif.onclose = () => {
        document.removeEventListener('visibilitychange', handleVisibilityChange)
      }
    } catch {
      // 静默处理
    }
  }

  // 处理 WebSocket 消息
  const handleWSMessage = (message: WSMessage) => {
    if (message.type === 'notification') {
      const notif = message.data as NotificationItem

      // 添加到列表头部
      notifications.value.unshift(notif)
      unreadCount.value++

      // 显示浏览器系统通知（跨标签页）
      showBrowserNotification(notif)

      // 同时显示页面内通知（当前标签页）
      ElNotification({
        title: notif.title,
        message: notif.content,
        type: 'info',
        duration: 4000,
        position: 'bottom-right',
        onClick: () => {
          if (notif.entity_type === 'issue' && notif.entity_key) {
            window.location.href = `/issues/${notif.entity_key}`
          }
        },
      })
    }
  }

  // 启动心跳
  const startHeartbeat = () => {
    stopHeartbeat()
    heartbeatTimer = setInterval(() => {
      if (ws.value && ws.value.readyState === WebSocket.OPEN) {
        ws.value.send(JSON.stringify({ type: 'ping' }))
      }
    }, 30000)
  }

  // 停止心跳
  const stopHeartbeat = () => {
    if (heartbeatTimer) {
      clearInterval(heartbeatTimer)
      heartbeatTimer = null
    }
  }

  // 连接 WebSocket
  const connectWebSocket = async (forceRefresh = false) => {
    // 确保 token 有效，过期或连接曾被拒绝时强制刷新
    const token = await ensureFreshToken(forceRefresh)
    if (!token) return

    // 请求通知权限
    await requestNotificationPermission()

    // 关闭现有连接
    disconnectWebSocket()

    const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:'
    const host = window.location.host
    const wsUrl = `${protocol}//${host}/ws?token=${token}`

    const socket = new WebSocket(wsUrl)
    let opened = false

    socket.onopen = () => {
      opened = true
      connected.value = true
      reconnectAttempts = 0 // 连接成功，重置重连计数
      startHeartbeat()
      // 连接后刷新数据
      fetchUnreadCountData()
      fetchNotifications()
    }

    socket.onmessage = (event) => {
      try {
        const message = JSON.parse(event.data) as WSMessage
        handleWSMessage(message)
      } catch {
        // 忽略无法解析的消息
      }
    }

    socket.onclose = () => {
      connected.value = false
      stopHeartbeat()
      ws.value = null

      // 连接从未建立（如 401）→ 视为认证失败
      const authFailed = !opened

      reconnectAttempts++

      // 连续失败超限，停止重连
      if (reconnectAttempts > MAX_RECONNECT_FAILURES) return

      // 指数退避重连：5s → 10s → 20s → 40s → 60s（上限）
      const delay = Math.min(5000 * Math.pow(2, reconnectAttempts - 1), MAX_RECONNECT_DELAY)
      reconnectTimer = setTimeout(() => {
        // 认证失败时强制刷新 token，避免用同一个被拒绝的 token 反复重试
        connectWebSocket(authFailed)
      }, delay)
    }

    socket.onerror = () => {
      socket.close()
    }

    ws.value = socket
  }

  // 断开 WebSocket
  const disconnectWebSocket = () => {
    stopHeartbeat()
    if (reconnectTimer) {
      clearTimeout(reconnectTimer)
      reconnectTimer = null
    }
    if (ws.value) {
      ws.value.close()
      ws.value = null
    }
    connected.value = false
  }

  return {
    // 状态
    notifications,
    unreadCount,
    connected,
    total,
    // 计算属性
    recentNotifications,
    // 方法
    fetchNotifications,
    fetchUnreadCount: fetchUnreadCountData,
    markAsRead,
    markAllAsRead,
    connectWebSocket,
    disconnectWebSocket,
    requestNotificationPermission,
  }
})
