<template>
  <div
    class="notification-item"
    :class="{ unread: !notification.is_read }"
    @click="$emit('click', notification)"
  >
    <div class="notif-icon">
      <el-icon :size="16" :color="iconColor">
        <component :is="iconName" />
      </el-icon>
    </div>
    <div class="notif-body">
      <div class="notif-title">{{ notification.title }}</div>
      <div class="notif-content" v-if="notification.content">{{ notification.content }}</div>
      <div class="notif-meta">
        <span v-if="notification.actor_name" class="notif-actor">{{ notification.actor_name }}</span>
        <span class="notif-time">{{ timeAgo(notification.created_at) }}</span>
      </div>
    </div>
    <div class="notif-dot" v-if="!notification.is_read"></div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import type { NotificationItem } from '@/types/notification'

const props = defineProps<{
  notification: NotificationItem
}>()

defineEmits<{
  click: [notification: NotificationItem]
}>()

const iconName = computed(() => {
  switch (props.notification.type) {
    case 'issue_assigned':
      return 'UserFilled'
    case 'issue_status_changed':
      return 'Switch'
    case 'issue_commented':
      return 'ChatDotRound'
    case 'mention':
      return 'ChatLineSquare'
    case 'issue_updated':
      return 'Edit'
    default:
      return 'Bell'
  }
})

const iconColor = computed(() => {
  switch (props.notification.type) {
    case 'issue_assigned':
      return '#3b82f6'
    case 'issue_status_changed':
      return '#f59e0b'
    case 'issue_commented':
      return '#10b981'
    case 'mention':
      return '#8b5cf6'
    case 'issue_updated':
      return '#6366f1'
    default:
      return '#9ca3af'
  }
})

const timeAgo = (dateStr: string): string => {
  const now = new Date()
  const date = new Date(dateStr)
  const diffMs = now.getTime() - date.getTime()
  const diffSec = Math.floor(diffMs / 1000)
  const diffMin = Math.floor(diffSec / 60)
  const diffHour = Math.floor(diffMin / 60)
  const diffDay = Math.floor(diffHour / 24)

  if (diffSec < 60) return '刚刚'
  if (diffMin < 60) return `${diffMin}分钟前`
  if (diffHour < 24) return `${diffHour}小时前`
  if (diffDay < 30) return `${diffDay}天前`
  return date.toLocaleDateString()
}
</script>

<style scoped>
.notification-item {
  display: flex;
  align-items: flex-start;
  gap: 10px;
  padding: 12px 16px;
  cursor: pointer;
  transition: background-color 0.2s;
  border-bottom: 1px solid #f0f0f0;
  position: relative;
}

.notification-item:hover {
  background-color: #f5f7fa;
}

.notification-item.unread {
  background-color: #f0f7ff;
}

.notification-item.unread:hover {
  background-color: #e6f0fa;
}

.notif-icon {
  flex-shrink: 0;
  width: 32px;
  height: 32px;
  border-radius: 50%;
  background: #f5f7fa;
  display: flex;
  align-items: center;
  justify-content: center;
  margin-top: 2px;
}

.notif-body {
  flex: 1;
  min-width: 0;
}

.notif-title {
  font-size: 13px;
  color: #1f2937;
  line-height: 1.4;
  font-weight: 500;
}

.notif-content {
  font-size: 12px;
  color: #6b7280;
  margin-top: 2px;
  line-height: 1.4;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.notif-meta {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-top: 4px;
  font-size: 11px;
  color: #9ca3af;
}

.notif-actor {
  color: #6b7280;
}

.notif-dot {
  flex-shrink: 0;
  width: 8px;
  height: 8px;
  border-radius: 50%;
  background: #3b82f6;
  margin-top: 6px;
}
</style>
