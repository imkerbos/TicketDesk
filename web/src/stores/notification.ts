import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
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
      console.log('[Notification] Browser does not support notifications')
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
    console.log('[Browser Notification] Attempting to show notification:', notif)
    console.log('[Browser Notification] Notification support:', 'Notification' in window)
    console.log('[Browser Notification] Permission:', window.Notification?.permission)

    if (!('Notification' in window)) {
      console.warn('[Browser Notification] Browser does not support notifications')
      return
    }

    if (window.Notification.permission !== 'granted') {
      console.warn('[Browser Notification] Permission not granted:', window.Notification.permission)
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

      console.log('[Browser Notification] Notification created successfully')

      // 点击通知时跳转到对应页面并标记为已读
      browserNotif.onclick = () => {
        console.log('[Browser Notification] Notification clicked')
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
    } catch (error) {
      console.error('[Browser Notification] Failed to create notification:', error)
    }
  }

  // 处理 WebSocket 消息
  const handleWSMessage = (message: WSMessage) => {
    console.log('[Notification Store] Received WS message:', message)

    if (message.type === 'notification') {
      const notif = message.data as NotificationItem
      console.log('[Notification Store] Processing notification:', notif)

      // 添加到列表头部
      notifications.value.unshift(notif)
      unreadCount.value++

      // 显示浏览器系统通知（跨标签页）
      showBrowserNotification(notif)

      // 同时显示页面内通知（当前标签页）
      console.log('[Notification Store] Showing ElNotification')
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
  const connectWebSocket = async () => {
    const token = localStorage.getItem('token')
    if (!token) return

    // 请求通知权限
    await requestNotificationPermission()

    // 关闭现有连接
    disconnectWebSocket()

    const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:'
    const host = window.location.host
    const wsUrl = `${protocol}//${host}/api/v1/ws?token=${token}`

    const socket = new WebSocket(wsUrl)

    socket.onopen = () => {
      connected.value = true
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

      // 5秒后重连
      reconnectTimer = setTimeout(() => {
        const currentToken = localStorage.getItem('token')
        if (currentToken) {
          connectWebSocket()
        }
      }, 5000)
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

  // 测试浏览器通知（用于调试）
  const testBrowserNotification = () => {
    console.log('[Test] Testing browser notification...')
    console.log('[Test] Permission:', window.Notification?.permission)

    if (!('Notification' in window)) {
      console.error('[Test] Browser does not support notifications')
      return
    }

    if (window.Notification.permission !== 'granted') {
      console.error('[Test] Permission not granted. Current permission:', window.Notification.permission)
      return
    }

    try {
      const testNotif = new window.Notification('测试通知', {
        body: '这是一条测试通知，用于验证浏览器通知功能是否正常工作',
        icon: '/favicon.ico',
        tag: 'test-notification',
      })

      console.log('[Test] Test notification created successfully')

      testNotif.onclick = () => {
        console.log('[Test] Test notification clicked')
        testNotif.close()
      }

      setTimeout(() => testNotif.close(), 4000)
    } catch (error) {
      console.error('[Test] Failed to create test notification:', error)
    }
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
    testBrowserNotification,
  }
})
