<template>
  <div class="notification-container">
    <!-- 页面头部 -->
    <TdPageHeader>
      <template #leading>
        <div class="page-header-icon">
          <el-icon :size="20"><Bell /></el-icon>
        </div>
      </template>
      <template #title>通知中心</template>
      <template #subtitle>查看和管理您的所有通知消息</template>
    </TdPageHeader>

    <!-- 筛选卡片 -->
    <el-card class="filter-card" shadow="never">
      <div class="filter-content">
        <div class="filter-left">
          <el-radio-group v-model="filterStatus" @change="handleFilterChange">
            <el-radio-button value="all">全部</el-radio-button>
            <el-radio-button value="unread">
              未读
              <el-badge
                v-if="notificationStore.unreadCount > 0"
                :value="notificationStore.unreadCount"
                :max="99"
                class="filter-badge"
              />
            </el-radio-button>
            <el-radio-button value="read">已读</el-radio-button>
          </el-radio-group>

          <el-select v-model="filterType" placeholder="通知类型" clearable class="filter-select" @change="handleFilterChange">
            <el-option label="工单指派" value="issue_assigned" />
            <el-option label="状态变更" value="issue_status_changed" />
            <el-option label="新评论" value="issue_commented" />
            <el-option label="@提及" value="mention" />
            <el-option label="工单更新" value="issue_updated" />
          </el-select>
        </div>

        <div class="filter-right">
          <el-button
            v-if="notificationStore.unreadCount > 0"
            type="primary"
            @click="handleMarkAllAsRead"
          >
            <el-icon><Check /></el-icon>
            全部标记为已读
          </el-button>
        </div>
      </div>
    </el-card>

    <!-- 通知列表 -->
    <el-card class="list-card" shadow="never">
      <div v-if="loading" class="loading-state">
        <el-skeleton :rows="6" animated />
      </div>

      <div v-else-if="notifications.length === 0" class="empty-state">
        <el-empty description="暂无通知">
          <template #image>
            <el-icon :size="80" color="#d1d5db"><BellFilled /></el-icon>
          </template>
        </el-empty>
      </div>

      <div v-else class="notification-list">
        <div
          v-for="item in notifications"
          :key="item.id"
          class="notification-item"
          :class="{ unread: !item.is_read }"
          @click="handleClick(item)"
        >
          <div class="item-left">
            <div class="item-icon" :style="{ background: getIconBg(item.type) }">
              <el-icon :size="18" color="#fff">
                <component :is="getIconName(item.type)" />
              </el-icon>
            </div>
          </div>

          <div class="item-content">
            <div class="item-header">
              <span class="item-title">{{ item.title }}</span>
              <el-tag v-if="item.entity_key" size="small" effect="plain" class="item-tag">
                {{ item.entity_key }}
              </el-tag>
            </div>
            <div v-if="item.content" class="item-body">{{ item.content }}</div>
            <div class="item-footer">
              <span v-if="item.actor_name" class="item-actor">
                <el-icon><User /></el-icon>
                {{ item.actor_name }}
              </span>
              <span class="item-time">
                <el-icon><Clock /></el-icon>
                {{ formatTime(item.created_at) }}
              </span>
              <el-tag v-if="!item.is_read" type="danger" size="small" effect="light">未读</el-tag>
            </div>
          </div>

          <div class="item-actions">
            <el-button
              v-if="!item.is_read"
              type="primary"
              text
              size="small"
              @click.stop="handleMarkRead(item.id)"
            >
              标记已读
            </el-button>
            <el-button
              type="danger"
              text
              size="small"
              @click.stop="handleDelete(item.id)"
            >
              删除
            </el-button>
          </div>
        </div>
      </div>

      <!-- 分页 -->
      <div v-if="total > pageSize" class="pagination-wrapper">
        <el-pagination
          v-model:current-page="currentPage"
          :page-size="pageSize"
          :total="total"
          layout="total, prev, pager, next"
          @current-change="handlePageChange"
        />
      </div>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { Bell, BellFilled, Check, User, Clock } from '@element-plus/icons-vue'
import { useNotificationStore } from '@/stores/notification'
import { getNotificationList, markAsRead, deleteNotification } from '@/api/notification'
import type { NotificationItem } from '@/types/notification'

const router = useRouter()
const notificationStore = useNotificationStore()

const notifications = ref<NotificationItem[]>([])
const loading = ref(false)
const currentPage = ref(1)
const pageSize = 15
const total = ref(0)
const filterStatus = ref('all')
const filterType = ref('')

const fetchData = async () => {
  loading.value = true
  try {
    const params: Record<string, unknown> = {
      page: currentPage.value,
      page_size: pageSize,
    }
    if (filterStatus.value === 'unread') params.is_read = false
    if (filterStatus.value === 'read') params.is_read = true
    if (filterType.value) params.type = filterType.value

    const res = await getNotificationList(params)
    notifications.value = res.data.data.items || []
    total.value = res.data.data.total
  } catch {
    // 静默处理
  } finally {
    loading.value = false
  }
}

const handleClick = (notification: NotificationItem) => {
  if (!notification.is_read) {
    handleMarkRead(notification.id)
  }
  if (notification.entity_type === 'issue' && notification.entity_key) {
    router.push(`/issues/${notification.entity_key}`)
  }
}

const handleMarkRead = async (id: number) => {
  try {
    await markAsRead(id)
    const notif = notifications.value.find((n) => n.id === id)
    if (notif && !notif.is_read) {
      notif.is_read = true
      notificationStore.unreadCount = Math.max(0, notificationStore.unreadCount - 1)
    }
  } catch {
    ElMessage.error('操作失败')
  }
}

const handleMarkAllAsRead = async () => {
  await notificationStore.markAllAsRead()
  notifications.value.forEach((n) => {
    n.is_read = true
  })
  ElMessage.success('已全部标记为已读')
}

const handleDelete = async (id: number) => {
  try {
    await deleteNotification(id)
    const idx = notifications.value.findIndex((n) => n.id === id)
    if (idx >= 0) {
      const notif = notifications.value[idx]
      if (!notif.is_read) {
        notificationStore.unreadCount = Math.max(0, notificationStore.unreadCount - 1)
      }
      notifications.value.splice(idx, 1)
      total.value--
    }
    ElMessage.success('已删除')
  } catch {
    ElMessage.error('删除失败')
  }
}

const handleFilterChange = () => {
  currentPage.value = 1
  fetchData()
}

const handlePageChange = () => {
  fetchData()
}

const getIconName = (type: string): string => {
  const map: Record<string, string> = {
    issue_assigned: 'UserFilled',
    issue_status_changed: 'Switch',
    issue_commented: 'ChatDotRound',
    mention: 'ChatLineSquare',
    issue_updated: 'Edit',
  }
  return map[type] || 'Bell'
}

const getIconBg = (type: string): string => {
  const map: Record<string, string> = {
    issue_assigned: '#3b82f6',
    issue_status_changed: '#f59e0b',
    issue_commented: '#10b981',
    mention: '#8b5cf6',
    issue_updated: '#6366f1',
  }
  return map[type] || '#9ca3af'
}

const formatTime = (dateStr: string): string => {
  const date = new Date(dateStr)
  const now = new Date()
  const diffMs = now.getTime() - date.getTime()
  const diffMin = Math.floor(diffMs / (1000 * 60))
  const diffHour = Math.floor(diffMs / (1000 * 60 * 60))
  const diffDay = Math.floor(diffMs / (1000 * 60 * 60 * 24))

  if (diffMin < 1) return '刚刚'
  if (diffMin < 60) return `${diffMin}分钟前`
  if (diffHour < 24) return `${diffHour}小时前`
  if (diffDay < 7) return `${diffDay}天前`

  return date.toLocaleString('zh-CN', {
    month: '2-digit',
    day: '2-digit',
    hour: '2-digit',
    minute: '2-digit',
  })
}

onMounted(() => {
  fetchData()
  notificationStore.fetchUnreadCount()
})
</script>

<style scoped lang="scss">
.notification-container {
  width: 100%;
}

// 页面头部 icon (TdPageHeader leading slot)
.page-header-icon {
  width: 40px;
  height: 40px;
  background: var(--td-tag-primary-bg);
  border-radius: var(--td-radius-md);
  display: flex;
  align-items: center;
  justify-content: center;
  color: var(--td-color-primary);
  flex-shrink: 0;
}

// 筛选卡片
.filter-card {
  margin-bottom: 20px;
  border-radius: 12px;

  :deep(.el-card__body) {
    padding: 16px 20px;
  }

  .filter-content {
    display: flex;
    justify-content: space-between;
    align-items: center;
    flex-wrap: wrap;
    gap: 12px;
  }

  .filter-left {
    display: flex;
    align-items: center;
    gap: 12px;
    flex-wrap: wrap;
  }

  .filter-select {
    width: 140px;
  }

  .filter-badge {
    margin-left: 4px;
  }
}

// 列表卡片
.list-card {
  border-radius: 12px;

  :deep(.el-card__body) {
    padding: 0;
  }
}

.notification-list {
  display: flex;
  flex-direction: column;
}

.notification-item {
  display: flex;
  align-items: flex-start;
  gap: 16px;
  padding: 20px 24px;
  border-bottom: 1px solid var(--td-divider-color);
  cursor: pointer;
  transition: all 150ms ease-out;

  &:last-child {
    border-bottom: none;
  }

  &:hover {
    background-color: var(--td-bg-page);
  }

  &.unread {
    background-color: var(--td-tag-primary-bg);

    &:hover {
      background-color: var(--td-tag-primary-bg);
    }
  }
}

.item-left {
  flex-shrink: 0;
}

.item-icon {
  width: 42px;
  height: 42px;
  border-radius: 10px;
  display: flex;
  align-items: center;
  justify-content: center;
}

.item-content {
  flex: 1;
  min-width: 0;
}

.item-header {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 6px;
}

.item-title {
  font-size: 15px;
  font-weight: 500;
  color: var(--td-text-primary);
}

.item-tag {
  flex-shrink: 0;
}

.item-body {
  font-size: 13px;
  color: var(--td-text-secondary);
  line-height: 1.5;
  margin-bottom: 8px;
  overflow: hidden;
  text-overflow: ellipsis;
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
}

.item-footer {
  display: flex;
  align-items: center;
  gap: 16px;
  font-size: 12px;
  color: var(--td-text-placeholder);

  .item-actor,
  .item-time {
    display: flex;
    align-items: center;
    gap: 4px;
  }
}

.item-actions {
  flex-shrink: 0;
  display: flex;
  gap: 4px;
  opacity: 0;
  transition: opacity 150ms ease-out;
}

.notification-item:hover .item-actions {
  opacity: 1;
}

.loading-state {
  padding: 24px;
}

.empty-state {
  padding: 60px 0;
}

.pagination-wrapper {
  display: flex;
  justify-content: center;
  padding: 20px;
  border-top: 1px solid var(--td-divider-color);
}
</style>
