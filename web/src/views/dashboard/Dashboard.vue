<template>
  <div class="dashboard-container">
    <!-- 页面头部 -->
    <div class="page-header">
      <div class="header-info">
        <div class="header-icon">
          <el-icon><DataBoard /></el-icon>
        </div>
        <div class="header-text">
          <h1 class="header-title">工作台</h1>
          <p class="header-desc">欢迎回来，这里是你的工作概览</p>
        </div>
      </div>
      <div class="header-actions">
        <el-button @click="$router.push('/alerts')">
          <el-icon><Bell /></el-icon>
          查看告警
        </el-button>
        <el-button @click="$router.push('/issues')">
          <el-icon><Tickets /></el-icon>
          所有工单
        </el-button>
        <el-button type="primary" @click="handleCreateIssue">
          <el-icon><Plus /></el-icon>
          创建工单
        </el-button>
      </div>
    </div>

    <!-- 统计卡片 -->
    <el-row :gutter="20" class="stat-row">
      <el-col :xs="12" :sm="6">
        <div class="stat-card todo" @click="$router.push('/issues?filter=my-todo')">
          <div class="stat-icon-wrapper">
            <el-icon :size="22"><Tickets /></el-icon>
          </div>
          <div class="stat-content">
            <div class="stat-value">{{ animatedStats.myTodo }}</div>
            <div class="stat-label">我的待办</div>
          </div>
        </div>
      </el-col>
      <el-col :xs="12" :sm="6">
        <div class="stat-card created" @click="$router.push('/issues?filter=my-created')">
          <div class="stat-icon-wrapper">
            <el-icon :size="22"><Edit /></el-icon>
          </div>
          <div class="stat-content">
            <div class="stat-value">{{ animatedStats.myCreated }}</div>
            <div class="stat-label">我创建的</div>
          </div>
        </div>
      </el-col>
      <el-col :xs="12" :sm="6">
        <div class="stat-card alert" @click="$router.push('/alerts?status=firing')">
          <div class="stat-icon-wrapper">
            <el-icon :size="22"><Bell /></el-icon>
          </div>
          <div class="stat-content">
            <div class="stat-value">{{ animatedStats.pendingAlerts }}</div>
            <div class="stat-label">待确认告警</div>
          </div>
        </div>
      </el-col>
      <el-col :xs="12" :sm="6">
        <div class="stat-card done">
          <div class="stat-icon-wrapper">
            <el-icon :size="22"><CircleCheck /></el-icon>
          </div>
          <div class="stat-content">
            <div class="stat-value">{{ animatedStats.weekDone }}</div>
            <div class="stat-label">本周已完成</div>
          </div>
        </div>
      </el-col>
    </el-row>

    <!-- 主内容区 -->
    <el-row :gutter="20">
      <!-- 左侧：工单列表 -->
      <el-col :xs="24" :lg="14">
        <!-- 待办工单 -->
        <el-card shadow="never" class="content-card">
          <template #header>
            <div class="card-header">
              <div class="card-title-group">
                <div class="card-icon todo">
                  <el-icon><Tickets /></el-icon>
                </div>
                <div>
                  <span class="card-title">我的待办工单</span>
                  <span class="card-count">{{ animatedStats.myTodo }}</span>
                </div>
              </div>
              <el-button link type="primary" @click="$router.push('/issues?filter=my-todo')">
                查看全部 <el-icon><ArrowRight /></el-icon>
              </el-button>
            </div>
          </template>
          <div v-loading="loadingTodo">
            <div v-if="todoIssues.length === 0" class="empty-state">
              <el-empty description="暂无待办工单" :image-size="80" />
            </div>
            <div v-else class="issue-list">
              <div
                v-for="issue in todoIssues"
                :key="issue.id"
                class="issue-item"
                @click="$router.push(`/issues/${issue.issue_key}`)"
              >
                <div class="issue-left">
                  <div class="priority-indicator" :class="issue.priority"></div>
                  <div class="issue-info">
                    <div class="issue-main">
                      <span class="issue-key">{{ issue.issue_key }}</span>
                      <span class="issue-title">{{ issue.title }}</span>
                    </div>
                    <div class="issue-meta">
                      <el-tag size="small" type="info" effect="plain">{{ issue.project_key }}</el-tag>
                      <span class="meta-divider">·</span>
                      <span class="meta-text">{{ getStatusText(issue.status) }}</span>
                    </div>
                  </div>
                </div>
                <el-tag :type="getPriorityType(issue.priority)" size="small" effect="dark">
                  {{ issue.priority }}
                </el-tag>
              </div>
            </div>
          </div>
        </el-card>

        <!-- 我创建的工单 -->
        <el-card shadow="never" class="content-card">
          <template #header>
            <div class="card-header">
              <div class="card-title-group">
                <div class="card-icon created">
                  <el-icon><Edit /></el-icon>
                </div>
                <div>
                  <span class="card-title">我创建的工单</span>
                  <span class="card-count">{{ animatedStats.myCreated }}</span>
                </div>
              </div>
              <el-button link type="primary" @click="$router.push('/issues?filter=my-created')">
                查看全部 <el-icon><ArrowRight /></el-icon>
              </el-button>
            </div>
          </template>
          <div v-loading="loadingCreated">
            <div v-if="createdIssues.length === 0" class="empty-state">
              <el-empty description="暂无创建的工单" :image-size="80" />
            </div>
            <div v-else class="issue-list">
              <div
                v-for="issue in createdIssues"
                :key="issue.id"
                class="issue-item"
                @click="$router.push(`/issues/${issue.issue_key}`)"
              >
                <div class="issue-left">
                  <div class="priority-indicator" :class="issue.priority"></div>
                  <div class="issue-info">
                    <div class="issue-main">
                      <span class="issue-key">{{ issue.issue_key }}</span>
                      <span class="issue-title">{{ issue.title }}</span>
                    </div>
                    <div class="issue-meta">
                      <el-tag :type="getStatusType(issue.status)" size="small" effect="plain">
                        {{ getStatusText(issue.status) }}
                      </el-tag>
                      <span class="meta-divider">·</span>
                      <span v-if="issue.assignee" class="assignee-tag">
                        <el-icon><User /></el-icon>
                        {{ issue.assignee.display_name }}
                      </span>
                      <span v-else class="meta-text unassigned">未指派</span>
                    </div>
                  </div>
                </div>
                <el-tag :type="getPriorityType(issue.priority)" size="small" effect="dark">
                  {{ issue.priority }}
                </el-tag>
              </div>
            </div>
          </div>
        </el-card>
      </el-col>

      <!-- 右侧：告警和活动 -->
      <el-col :xs="24" :lg="10">
        <!-- 待确认告警 -->
        <el-card shadow="never" class="content-card">
          <template #header>
            <div class="card-header">
              <div class="card-title-group">
                <div class="card-icon alert">
                  <el-icon><Bell /></el-icon>
                </div>
                <div>
                  <span class="card-title">待确认告警</span>
                  <span class="card-count danger">{{ animatedStats.pendingAlerts }}</span>
                </div>
              </div>
              <el-button link type="primary" @click="$router.push('/alerts?status=firing')">
                查看全部 <el-icon><ArrowRight /></el-icon>
              </el-button>
            </div>
          </template>
          <div v-loading="loadingAlerts">
            <div v-if="pendingAlerts.length === 0" class="empty-state">
              <el-empty description="暂无待确认告警" :image-size="60" />
            </div>
            <div v-else class="alert-list">
              <div
                v-for="alert in pendingAlerts"
                :key="alert.id"
                class="alert-item"
                @click="$router.push(`/alerts/${alert.id}`)"
              >
                <div class="alert-severity-bar" :class="alert.severity"></div>
                <div class="alert-content">
                  <div class="alert-name">{{ alert.alert_name }}</div>
                  <div class="alert-time">
                    <el-icon><Clock /></el-icon>
                    {{ formatTime(alert.starts_at) }}
                  </div>
                </div>
                <el-tag :type="getSeverityTagType(alert.severity)" size="small" effect="dark">
                  {{ getSeverityText(alert.severity) }}
                </el-tag>
              </div>
            </div>
          </div>
        </el-card>

        <!-- 最近活动 -->
        <el-card shadow="never" class="content-card">
          <template #header>
            <div class="card-header">
              <div class="card-title-group">
                <div class="card-icon activity">
                  <el-icon><Clock /></el-icon>
                </div>
                <span class="card-title">最近活动</span>
              </div>
            </div>
          </template>
          <div v-loading="loadingActivities">
            <div v-if="recentActivities.length === 0" class="empty-state">
              <el-empty description="暂无活动记录" :image-size="60" />
            </div>
            <el-timeline v-else class="activity-timeline">
              <el-timeline-item
                v-for="activity in recentActivities"
                :key="activity.id"
                :timestamp="formatTime(activity.created_at)"
                placement="top"
              >
                <div class="activity-content">
                  <span class="activity-user">{{ activity.user_name }}</span>
                  <span class="activity-action">{{ activity.action }}</span>
                  <el-link
                    v-if="activity.entity_key"
                    type="primary"
                    @click.stop="$router.push(`/issues/${activity.entity_key}`)"
                  >
                    {{ activity.entity_key }}
                  </el-link>
                </div>
              </el-timeline-item>
            </el-timeline>
          </div>
        </el-card>
      </el-col>
    </el-row>

    <!-- 创建工单对话框 -->
    <CreateIssueDialog
      v-model="createDialogVisible"
      @created="(key: string) => router.push(`/issues/${key}`)"
    />
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted, watch } from 'vue'
import { useRouter } from 'vue-router'
import {
  Tickets, Edit, Bell, CircleCheck, Plus, ArrowRight, Clock,
  DataBoard, User
} from '@element-plus/icons-vue'
import { getMyTodoIssues, getMyCreatedIssues } from '@/api/issue'
import { getAlertList } from '@/api/alert'
import { getRecentActivities } from '@/api/activity'
import type { Issue } from '@/types/issue'
import type { Alert } from '@/types/alert'
import type { Activity } from '@/api/activity'
import CreateIssueDialog from '@/components/CreateIssueDialog.vue'
import dayjs from 'dayjs'

const router = useRouter()

// 统计数据
const stats = reactive({
  myTodo: 0,
  myCreated: 0,
  pendingAlerts: 0,
  weekDone: 0,
})

// 计数动画
const animatedStats = reactive({
  myTodo: 0,
  myCreated: 0,
  pendingAlerts: 0,
  weekDone: 0,
})

const animateTo = (key: keyof typeof animatedStats, to: number) => {
  const from = animatedStats[key]
  if (from === to) {
    animatedStats[key] = to
    return
  }
  if (window.matchMedia('(prefers-reduced-motion: reduce)').matches) {
    animatedStats[key] = to
    return
  }
  const duration = 250
  const startTime = performance.now()
  const step = (now: number) => {
    const progress = Math.min((now - startTime) / duration, 1)
    const eased = 1 - (1 - progress) ** 3
    animatedStats[key] = Math.round(from + (to - from) * eased)
    if (progress < 1) requestAnimationFrame(step)
  }
  requestAnimationFrame(step)
}

watch(() => stats.myTodo, (val) => animateTo('myTodo', val))
watch(() => stats.myCreated, (val) => animateTo('myCreated', val))
watch(() => stats.pendingAlerts, (val) => animateTo('pendingAlerts', val))
watch(() => stats.weekDone, (val) => animateTo('weekDone', val))

// 待办工单
const loadingTodo = ref(false)
const todoIssues = ref<Issue[]>([])

// 我创建的工单
const loadingCreated = ref(false)
const createdIssues = ref<Issue[]>([])

// 待确认告警
const loadingAlerts = ref(false)
const pendingAlerts = ref<Alert[]>([])

// 最近活动
const loadingActivities = ref(false)
const recentActivities = ref<Activity[]>([])

// 创建工单
const createDialogVisible = ref(false)

// 加载待办工单
const loadTodoIssues = async () => {
  loadingTodo.value = true
  try {
    const { data } = await getMyTodoIssues({ page: 1, page_size: 5 })
    todoIssues.value = data.data.items
    stats.myTodo = data.data.total
  } catch (error) {
    console.error('Failed to load todo issues:', error)
  } finally {
    loadingTodo.value = false
  }
}

// 加载我创建的工单
const loadCreatedIssues = async () => {
  loadingCreated.value = true
  try {
    const { data } = await getMyCreatedIssues({ page: 1, page_size: 5 })
    createdIssues.value = data.data.items
    stats.myCreated = data.data.total
  } catch (error) {
    console.error('Failed to load created issues:', error)
  } finally {
    loadingCreated.value = false
  }
}

// 加载待确认告警
const loadPendingAlerts = async () => {
  loadingAlerts.value = true
  try {
    const { data } = await getAlertList({ status: 'firing', page: 1, page_size: 5 })
    pendingAlerts.value = data.data.items
    stats.pendingAlerts = data.data.total
  } catch (error) {
    console.error('Failed to load alerts:', error)
  } finally {
    loadingAlerts.value = false
  }
}

// 加载最近活动
const loadRecentActivities = async () => {
  loadingActivities.value = true
  try {
    const { data } = await getRecentActivities(10)
    recentActivities.value = data.data
  } catch (error) {
    console.error('Failed to load activities:', error)
  } finally {
    loadingActivities.value = false
  }
}

// 打开创建工单对话框
const handleCreateIssue = () => {
  createDialogVisible.value = true
}

// 工具函数
type TagType = 'primary' | 'success' | 'warning' | 'info' | 'danger'

const getPriorityType = (priority: string): TagType => {
  const map: Record<string, TagType> = { P0: 'danger', P1: 'warning', P2: 'info', P3: 'info' }
  return map[priority] || 'info'
}

const getStatusType = (status: string): TagType => {
  const map: Record<string, TagType> = { open: 'info', in_progress: 'warning', resolved: 'success', closed: 'info' }
  return map[status] || 'info'
}

const getStatusText = (status: string) => {
  const map: Record<string, string> = { open: '待处理', in_progress: '进行中', pending_review: '待确认', resolved: '已完成', closed: '已终止', merged: '已合并' }
  return map[status] || status
}

const getSeverityTagType = (severity: string): TagType => {
  const map: Record<string, TagType> = { critical: 'danger', warning: 'warning', info: 'info' }
  return map[severity] || 'info'
}

const getSeverityText = (severity: string) => {
  const map: Record<string, string> = { critical: '严重', warning: '警告', info: '信息' }
  return map[severity] || severity
}

const formatTime = (time: string) => {
  return dayjs(time).format('MM-DD HH:mm')
}

// 初始化
onMounted(() => {
  loadTodoIssues()
  loadCreatedIssues()
  loadPendingAlerts()
  loadRecentActivities()
})
</script>

<style scoped lang="scss">
.dashboard-container {
  width: 100%;
}

// 页面头部
.page-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 24px;
  padding: 24px 28px;
  background: var(--td-bg-card);
  border-radius: 8px;
  border: 1px solid var(--td-border-color);
  border-left: 4px solid #3b82f6;

  .header-info {
    display: flex;
    align-items: center;
    gap: 16px;
  }

  .header-icon {
    width: 48px;
    height: 48px;
    background: var(--td-tag-primary-bg);
    border-radius: 10px;
    display: flex;
    align-items: center;
    justify-content: center;
    font-size: 24px;
    color: var(--td-color-primary);
  }

  .header-text {
    .header-title {
      font-size: 22px;
      font-weight: 700;
      margin: 0 0 4px 0;
      color: var(--td-text-primary);
    }
    .header-desc {
      font-size: 14px;
      margin: 0;
      color: var(--td-text-secondary);
    }
  }

  .header-actions {
    display: flex;
    gap: 10px;

    .el-button--primary {
      transition: box-shadow 150ms ease-out;
      &:hover { box-shadow: 0 0 0 3px rgba(59, 130, 246, 0.15); }
      &:active { background-color: var(--td-color-primary-active); border-color: var(--td-color-primary-active); }
    }
  }
}

// 统计卡片
.stat-row {
  margin-bottom: 24px;
}

.stat-card {
  position: relative;
  display: flex;
  align-items: center;
  padding: 20px 24px;
  border-radius: 8px;
  background: var(--td-bg-card);
  border: 1px solid var(--td-border-color);
  cursor: pointer;
  transition: border-color 150ms ease-out, box-shadow 150ms ease-out, background-color 150ms ease-out;

  .stat-icon-wrapper {
    width: 44px;
    height: 44px;
    display: flex;
    align-items: center;
    justify-content: center;
    border-radius: 10px;
    margin-right: 16px;
    flex-shrink: 0;
  }

  .stat-content {
    .stat-value {
      font-size: 28px;
      font-weight: 700;
      line-height: 1.2;
      color: var(--td-text-primary);
    }
    .stat-label {
      font-size: 13px;
      color: var(--td-text-secondary);
      margin-top: 2px;
    }
  }

  &.todo {
    border-left: 4px solid #3b82f6;
    .stat-icon-wrapper { background: var(--td-tag-primary-bg); color: var(--td-color-primary); }
    &:hover { border-color: var(--td-color-primary); box-shadow: 0 1px 4px rgba(59, 130, 246, 0.1); }
    &:active { background-color: var(--td-bg-page); }
  }
  &.created {
    border-left: 4px solid #10b981;
    .stat-icon-wrapper { background: var(--td-tag-success-bg); color: var(--td-color-success); }
    &:hover { border-color: var(--td-color-success); box-shadow: 0 1px 4px rgba(16, 185, 129, 0.1); }
    &:active { background-color: var(--td-bg-page); }
  }
  &.alert {
    border-left: 4px solid #ef4444;
    .stat-icon-wrapper { background: var(--td-tag-danger-bg); color: var(--td-color-danger); }
    &:hover { border-color: var(--td-color-danger); box-shadow: 0 1px 4px rgba(239, 68, 68, 0.1); }
    &:active { background-color: var(--td-bg-page); }
  }
  &.done {
    border-left: 4px solid #3b82f6;
    .stat-icon-wrapper { background: var(--td-tag-primary-bg); color: var(--td-color-primary); }
    &:hover { border-color: var(--td-color-primary); box-shadow: 0 1px 4px rgba(59, 130, 246, 0.1); }
    &:active { background-color: var(--td-bg-page); }
  }
}

// 内容卡片
.content-card {
  margin-bottom: 20px;
  border-radius: 8px;
  min-height: 200px;

  :deep(.el-card__header) {
    padding: 16px 20px;
    border-bottom: 1px solid var(--td-divider-color);
  }

  :deep(.el-card__body) {
    padding: 0;
  }
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;

  .card-title-group {
    display: flex;
    align-items: center;
    gap: 10px;
  }

  .card-icon {
    width: 32px;
    height: 32px;
    border-radius: 6px;
    display: flex;
    align-items: center;
    justify-content: center;
    font-size: 16px;

    &.todo { background: var(--td-tag-primary-bg); color: var(--td-color-primary); }
    &.created { background: var(--td-tag-success-bg); color: var(--td-color-success); }
    &.alert { background: var(--td-tag-danger-bg); color: var(--td-color-danger); }
    &.activity { background: var(--td-tag-primary-bg); color: var(--td-color-primary); }
  }

  .card-title {
    font-size: 15px;
    font-weight: 600;
    color: var(--td-text-primary);
  }

  .card-count {
    display: inline-flex;
    align-items: center;
    justify-content: center;
    min-width: 20px;
    height: 20px;
    padding: 0 6px;
    background: var(--td-bg-section);
    border-radius: 10px;
    font-size: 12px;
    font-weight: 600;
    color: var(--td-text-regular);
    margin-left: 8px;

    &.danger {
      background: var(--td-tag-danger-bg);
      color: var(--td-color-danger);
    }
  }

  .el-button .el-icon {
    transition: transform 150ms ease-out;
  }
  .el-button:hover .el-icon {
    transform: translateX(2px);
  }
}

.empty-state {
  padding: 40px 20px;
  display: flex;
  align-items: center;
  justify-content: center;
}

// 工单列表
.issue-list {
  .issue-item {
    display: flex;
    justify-content: space-between;
    align-items: center;
    padding: 14px 20px;
    cursor: pointer;
    transition: background-color 150ms ease-out;
    border-bottom: 1px solid var(--td-divider-color);

    &:last-child { border-bottom: none; }
    &:hover { background-color: var(--td-bg-page); }
    &:active { background-color: var(--td-bg-section); }

    .issue-left {
      display: flex;
      align-items: center;
      gap: 12px;
      flex: 1;
      min-width: 0;
    }

    .priority-indicator {
      width: 4px;
      height: 36px;
      border-radius: 2px;
      flex-shrink: 0;

      &.P0 { background: var(--td-color-danger); }
      &.P1 { background: var(--td-color-warning); }
      &.P2 { background: var(--td-color-primary); }
      &.P3 { background: var(--td-text-placeholder); }
    }

    .issue-info {
      flex: 1;
      min-width: 0;
    }

    .issue-main {
      display: flex;
      align-items: center;
      gap: 10px;
      margin-bottom: 4px;

      .issue-key {
        color: var(--td-color-primary);
        font-weight: 600;
        font-size: 13px;
        flex-shrink: 0;
        &:hover { text-decoration: underline; }
      }
      .issue-title {
        color: var(--td-text-primary);
        font-weight: 500;
        font-size: 14px;
        overflow: hidden;
        text-overflow: ellipsis;
        white-space: nowrap;
      }
    }

    .issue-meta {
      display: flex;
      align-items: center;
      gap: 6px;
      font-size: 12px;

      .meta-divider { color: var(--td-text-disabled); }
      .meta-text { color: var(--td-text-placeholder); }
      .assignee-tag {
        display: flex;
        align-items: center;
        gap: 4px;
        color: var(--td-text-secondary);
      }
      .unassigned { font-style: italic; }
    }
  }
}

// 告警列表
.alert-list {
  .alert-item {
    display: flex;
    align-items: center;
    padding: 14px 20px;
    cursor: pointer;
    transition: background-color 150ms ease-out;
    border-bottom: 1px solid var(--td-divider-color);

    &:last-child { border-bottom: none; }
    &:hover { background-color: var(--td-tag-danger-bg); }
    &:active { background-color: var(--td-tag-danger-bg); }

    .alert-severity-bar {
      width: 4px;
      height: 40px;
      border-radius: 2px;
      margin-right: 14px;
      flex-shrink: 0;

      &.critical { background: var(--td-color-danger); }
      &.warning { background: var(--td-color-warning); }
      &.info { background: var(--td-text-secondary); }
    }

    .alert-content {
      flex: 1;
      min-width: 0;

      .alert-name {
        font-weight: 500;
        color: var(--td-text-primary);
        font-size: 14px;
        overflow: hidden;
        text-overflow: ellipsis;
        white-space: nowrap;
      }
      .alert-time {
        display: flex;
        align-items: center;
        gap: 4px;
        font-size: 12px;
        color: var(--td-text-placeholder);
        margin-top: 4px;
      }
    }
  }
}

// 活动时间线
.activity-timeline {
  padding: 20px;

  :deep(.el-timeline-item__timestamp) {
    font-size: 12px;
  }

  .activity-content {
    font-size: 14px;

    .activity-user {
      font-weight: 600;
      color: var(--td-text-primary);
    }
    .activity-action {
      color: var(--td-text-secondary);
      margin: 0 4px;
    }

    :deep(.el-link):hover {
      text-decoration: underline;
    }
  }
}

// 响应式
@media (max-width: 768px) {
  .page-header {
    flex-direction: column;
    align-items: flex-start;
    gap: 16px;
    padding: 20px;

    .header-actions {
      flex-wrap: wrap;
    }
  }

  .stat-card {
    padding: 16px;
    margin-bottom: 12px;

    .stat-content .stat-value {
      font-size: 24px;
    }
  }
}
</style>
