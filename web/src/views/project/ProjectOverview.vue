<template>
  <div class="project-overview-container">
    <!-- 页面头部（头部+导航一体化） -->
    <div class="overview-hero">
      <div class="hero-top">
        <div class="header-info">
          <div class="project-icon" :style="{ background: getProjectColor(projectKey) }">
            {{ projectKey.substring(0, 2).toUpperCase() }}
          </div>
          <div class="header-text">
            <div class="project-name-row">
              <h1 class="project-name">{{ project?.name || projectKey }}</h1>
              <span class="project-key-badge">{{ projectKey }}</span>
            </div>
            <p v-if="project?.description" class="project-desc">{{ project.description }}</p>
            <p v-else class="project-desc muted">暂无描述</p>
          </div>
        </div>
        <router-link :to="`/projects/${projectKey}/settings`" class="settings-link">
          <el-button :icon="Setting" size="small" class="settings-btn">设置</el-button>
        </router-link>
      </div>
      <div class="hero-nav">
        <router-link :to="`/projects/${projectKey}`" class="nav-item active">
          <el-icon><DataAnalysis /></el-icon>
          概览
        </router-link>
        <router-link :to="`/projects/${projectKey}/board`" class="nav-item">
          <el-icon><Grid /></el-icon>
          看板
        </router-link>
        <router-link :to="`/projects/${projectKey}/settings`" class="nav-item">
          <el-icon><Setting /></el-icon>
          设置
        </router-link>
      </div>
    </div>

    <div v-loading="loading" class="overview-content">
      <!-- 统计卡片 -->
      <el-row :gutter="16" class="stat-row">
        <el-col :xs="12" :sm="6">
          <div class="stat-card" @click="goBoard('open')">
            <div class="stat-icon-wrapper todo">
              <el-icon :size="20"><Tickets /></el-icon>
            </div>
            <div class="stat-info">
              <div class="stat-value">{{ stats.open }}</div>
              <div class="stat-label">待处理</div>
            </div>
          </div>
        </el-col>
        <el-col :xs="12" :sm="6">
          <div class="stat-card" @click="goBoard('in_progress')">
            <div class="stat-icon-wrapper progress">
              <el-icon :size="20"><Loading /></el-icon>
            </div>
            <div class="stat-info">
              <div class="stat-value">{{ stats.inProgress }}</div>
              <div class="stat-label">进行中</div>
            </div>
          </div>
        </el-col>
        <el-col :xs="12" :sm="6">
          <div class="stat-card" @click="goBoard('pending_review')">
            <div class="stat-icon-wrapper review">
              <el-icon :size="20"><View /></el-icon>
            </div>
            <div class="stat-info">
              <div class="stat-value">{{ stats.pendingReview }}</div>
              <div class="stat-label">待确认</div>
            </div>
          </div>
        </el-col>
        <el-col :xs="12" :sm="6">
          <div class="stat-card" @click="goBoard('resolved')">
            <div class="stat-icon-wrapper done">
              <el-icon :size="20"><CircleCheck /></el-icon>
            </div>
            <div class="stat-info">
              <div class="stat-value">{{ stats.resolved }}</div>
              <div class="stat-label">已完成</div>
            </div>
          </div>
        </el-col>
      </el-row>

      <!-- 主体两栏 -->
      <el-row :gutter="20">
        <!-- 左：最近工单 -->
        <el-col :xs="24" :lg="16">
          <el-card shadow="never" class="content-card">
            <template #header>
              <div class="card-header">
                <div class="card-title-group">
                  <div class="card-icon">
                    <el-icon><Tickets /></el-icon>
                  </div>
                  <span class="card-title">最近工单</span>
                </div>
                <router-link :to="`/projects/${projectKey}/board`" class="view-all-link">
                  进入看板 <el-icon><ArrowRight /></el-icon>
                </router-link>
              </div>
            </template>
            <div v-if="recentIssues.length === 0" class="empty-placeholder">暂无工单</div>
            <div v-else class="issue-list">
              <div
                v-for="item in recentIssues"
                :key="item.id"
                class="issue-item"
                @click="goIssue(item.issue_key)"
              >
                <div class="issue-left">
                  <div class="priority-dot" :class="item.priority"></div>
                  <span class="issue-key">{{ item.issue_key }}</span>
                  <span class="issue-title-text">{{ item.title }}</span>
                </div>
                <div class="issue-right">
                  <div class="status-badge" :class="item.status">
                    <span class="status-dot"></span>
                    <span>{{ getStatusText(item.status) }}</span>
                  </div>
                  <div v-if="item.assignee" class="mini-avatar" :title="item.assignee.display_name">
                    {{ item.assignee.display_name?.charAt(0) || '?' }}
                  </div>
                </div>
              </div>
            </div>
          </el-card>
        </el-col>

        <!-- 右：团队成员 -->
        <el-col :xs="24" :lg="8">
          <el-card shadow="never" class="content-card">
            <template #header>
              <div class="card-header">
                <div class="card-title-group">
                  <div class="card-icon">
                    <el-icon><User /></el-icon>
                  </div>
                  <span class="card-title">团队成员</span>
                  <span class="card-count">({{ members.length }})</span>
                </div>
              </div>
            </template>
            <div v-if="members.length === 0" class="empty-placeholder">暂无成员</div>
            <div v-else class="member-list">
              <div v-for="member in members" :key="member.id" class="member-item">
                <div class="member-avatar">
                  {{ member.user?.display_name?.charAt(0) || '?' }}
                </div>
                <div class="member-info">
                  <span class="member-name">{{ member.user?.display_name || '未知用户' }}</span>
                  <span class="member-role">{{ member.role_name || member.role }}</span>
                </div>
              </div>
            </div>
          </el-card>
        </el-col>
      </el-row>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, computed, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { Tickets, User, Setting, ArrowRight, View, CircleCheck, Loading, DataAnalysis, Grid } from '@element-plus/icons-vue'
import { getProjectDetail, getProjectMembers } from '@/api/project'
import { getIssueList } from '@/api/issue'
import type { Project, ProjectMember } from '@/types/project'
import type { Issue } from '@/types/issue'

const route = useRoute()
const router = useRouter()

const projectKey = computed(() => route.params.key as string)
const loading = ref(false)
const project = ref<Project | null>(null)
const recentIssues = ref<Issue[]>([])
const members = ref<ProjectMember[]>([])

const stats = reactive({
  open: 0,
  inProgress: 0,
  pendingReview: 0,
  resolved: 0,
})

const projectColors = ['#3b82f6', '#ef4444', '#3b82f6', '#10b981', '#f59e0b', '#8b5cf6']
const getProjectColor = (key: string) => {
  let hash = 0
  for (let i = 0; i < key.length; i++) hash = key.charCodeAt(i) + ((hash << 5) - hash)
  return projectColors[Math.abs(hash) % projectColors.length]
}

const getStatusText = (status: string) => {
  const map: Record<string, string> = {
    open: '待处理', in_progress: '进行中', pending_review: '待确认',
    resolved: '已完成', closed: '已终止', reopened: '重新打开', merged: '已合并',
  }
  return map[status] || status
}

const loadData = async () => {
  loading.value = true
  try {
    const [projectRes, membersRes, issuesRes] = await Promise.all([
      getProjectDetail(projectKey.value),
      getProjectMembers(projectKey.value),
      getIssueList({ project_key: projectKey.value, page: 1, page_size: 10 }),
    ])
    project.value = projectRes.data.data
    members.value = membersRes.data.data || []
    recentIssues.value = issuesRes.data.data.items || []

    // 统计各状态数量（分别请求）
    await loadStats()
  } catch (e) {
    console.error('Failed to load project overview:', e)
  } finally {
    loading.value = false
  }
}

const loadStats = async () => {
  try {
    const [openRes, progressRes, reviewRes, resolvedRes] = await Promise.all([
      getIssueList({ project_key: projectKey.value, status: 'open' as any, page: 1, page_size: 1 }),
      getIssueList({ project_key: projectKey.value, status: 'in_progress' as any, page: 1, page_size: 1 }),
      getIssueList({ project_key: projectKey.value, status: 'pending_review' as any, page: 1, page_size: 1 }),
      getIssueList({ project_key: projectKey.value, status: 'resolved' as any, page: 1, page_size: 1 }),
    ])
    stats.open = openRes.data.data.total
    stats.inProgress = progressRes.data.data.total
    stats.pendingReview = reviewRes.data.data.total
    stats.resolved = resolvedRes.data.data.total
  } catch (e) {
    console.error('Failed to load stats:', e)
  }
}

const goBoard = (_status?: string) => {
  router.push(`/projects/${projectKey.value}/board`)
}

const goIssue = (issueKey: string) => {
  router.push(`/projects/${projectKey.value}/board/${issueKey}`)
}

onMounted(() => {
  loadData()
})
</script>

<style scoped lang="scss">
.project-overview-container {
  width: 100%;
}

// ---- 一体化头部 ----
.overview-hero {
  background: #fff;
  border-radius: 12px;
  margin-bottom: 20px;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.04);
  overflow: hidden;
}

.hero-top {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  padding: 24px 28px 16px;
}

.header-info {
  display: flex;
  align-items: flex-start;
  gap: 16px;
}

.project-icon {
  width: 48px;
  height: 48px;
  border-radius: 12px;
  display: flex;
  align-items: center;
  justify-content: center;
  color: #fff;
  font-weight: 700;
  font-size: 16px;
  flex-shrink: 0;
}

.project-name-row {
  display: flex;
  align-items: center;
  gap: 10px;
  margin-bottom: 4px;
}

.project-name {
  font-size: 20px;
  font-weight: 600;
  color: #1f2937;
  margin: 0;
  line-height: 1.3;
}

.project-key-badge {
  font-size: 11px;
  font-weight: 600;
  color: #3b82f6;
  background: #eff6ff;
  padding: 2px 8px;
  border-radius: 4px;
  letter-spacing: 0.03em;
}

.project-desc {
  font-size: 13px;
  color: #6b7280;
  margin: 0;
  line-height: 1.5;

  &.muted {
    color: #d1d5db;
    font-style: italic;
  }
}

.settings-link {
  text-decoration: none;
  flex-shrink: 0;
}

.settings-btn {
  border-radius: 6px;
}

.hero-nav {
  display: flex;
  gap: 2px;
  padding: 0 28px 12px;
  border-top: 1px solid #f0f0f0;
  padding-top: 12px;

  .nav-item {
    display: inline-flex;
    align-items: center;
    gap: 5px;
    padding: 6px 14px;
    border-radius: 6px;
    font-size: 13px;
    color: #6b7280;
    text-decoration: none;
    transition: background-color 150ms ease-out, color 150ms ease-out;

    .el-icon {
      font-size: 14px;
    }

    &:hover {
      background: #f3f4f6;
      color: #1f2937;
    }

    &.active,
    &.router-link-exact-active {
      background: #eff6ff;
      color: #3b82f6;
      font-weight: 500;
    }
  }
}

// 统计卡片
.stat-row {
  margin-bottom: 20px;
}

.stat-card {
  display: flex;
  align-items: center;
  gap: 14px;
  padding: 20px;
  background: #fff;
  border-radius: 12px;
  cursor: pointer;
  border: 1px solid #f0f0f0;
  transition: box-shadow 150ms ease-out;

  &:hover {
    box-shadow: 0 4px 12px rgba(0, 0, 0, 0.08);
  }
}

.stat-icon-wrapper {
  width: 44px;
  height: 44px;
  border-radius: 10px;
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;

  &.todo { background: #fef3c7; color: #f59e0b; }
  &.progress { background: #dbeafe; color: #3b82f6; }
  &.review { background: #e0e7ff; color: #6366f1; }
  &.done { background: #d1fae5; color: #10b981; }
}

.stat-info {
  .stat-value {
    font-size: 24px;
    font-weight: 700;
    color: #1f2937;
    line-height: 1;
  }

  .stat-label {
    font-size: 13px;
    color: #6b7280;
    margin-top: 4px;
  }
}

// 内容卡片
.content-card {
  border-radius: 12px;
  margin-bottom: 20px;

  :deep(.el-card__header) {
    padding: 16px 20px;
    border-bottom: 1px solid #f0f0f0;
  }

  :deep(.el-card__body) {
    padding: 0;
  }
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.card-title-group {
  display: flex;
  align-items: center;
  gap: 8px;
}

.card-icon {
  width: 28px;
  height: 28px;
  border-radius: 6px;
  background: #eff6ff;
  color: #3b82f6;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 14px;
}

.card-title {
  font-size: 15px;
  font-weight: 600;
  color: #1f2937;
}

.card-count {
  font-size: 13px;
  color: #9ca3af;
  font-weight: 400;
}

.view-all-link {
  display: flex;
  align-items: center;
  gap: 4px;
  font-size: 13px;
  color: #3b82f6;
  text-decoration: none;

  &:hover {
    color: #2563eb;
  }
}

// 工单列表
.issue-list {
  .issue-item {
    display: flex;
    justify-content: space-between;
    align-items: center;
    padding: 12px 20px;
    cursor: pointer;
    transition: background-color 150ms ease-out;

    &:hover {
      background: #f9fafb;
    }

    &:not(:last-child) {
      border-bottom: 1px solid #f5f5f5;
    }
  }

  .issue-left {
    display: flex;
    align-items: center;
    gap: 10px;
    flex: 1;
    min-width: 0;
  }

  .priority-dot {
    width: 8px;
    height: 8px;
    border-radius: 50%;
    flex-shrink: 0;

    &.P0 { background: #ef4444; }
    &.P1 { background: #f59e0b; }
    &.P2 { background: #3b82f6; }
    &.P3 { background: #10b981; }
  }

  .issue-key {
    font-size: 13px;
    font-weight: 600;
    color: #3b82f6;
    flex-shrink: 0;
  }

  .issue-title-text {
    font-size: 14px;
    color: #1f2937;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }

  .issue-right {
    display: flex;
    align-items: center;
    gap: 10px;
    flex-shrink: 0;
    margin-left: 12px;
  }
}

// 状态徽章
.status-badge {
  display: inline-flex;
  align-items: center;
  gap: 4px;
  font-size: 11px;
  padding: 2px 8px;
  border-radius: 10px;
  font-weight: 500;

  .status-dot {
    width: 6px;
    height: 6px;
    border-radius: 50%;
  }

  &.open { background: #fef3c7; color: #92400e; .status-dot { background: #f59e0b; } }
  &.in_progress { background: #dbeafe; color: #1e40af; .status-dot { background: #3b82f6; } }
  &.pending_review { background: #e0e7ff; color: #3730a3; .status-dot { background: #6366f1; } }
  &.resolved { background: #d1fae5; color: #065f46; .status-dot { background: #10b981; } }
  &.closed { background: #f3f4f6; color: #374151; .status-dot { background: #6b7280; } }
  &.reopened { background: #fee2e2; color: #991b1b; .status-dot { background: #ef4444; } }
  &.merged { background: #ede9fe; color: #5b21b6; .status-dot { background: #8b5cf6; } }
}

.mini-avatar {
  width: 26px;
  height: 26px;
  border-radius: 6px;
  background: #3b82f6;
  color: #fff;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 11px;
  font-weight: 600;
  flex-shrink: 0;
}

// 成员列表
.member-list {
  padding: 8px 0;
}

.member-item {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 10px 20px;

  &:not(:last-child) {
    border-bottom: 1px solid #f5f5f5;
  }
}

.member-avatar {
  width: 32px;
  height: 32px;
  border-radius: 8px;
  background: #3b82f6;
  color: #fff;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 13px;
  font-weight: 600;
  flex-shrink: 0;
}

.member-info {
  display: flex;
  flex-direction: column;

  .member-name {
    font-size: 14px;
    font-weight: 500;
    color: #1f2937;
  }

  .member-role {
    font-size: 12px;
    color: #9ca3af;
  }
}

.empty-placeholder {
  padding: 40px;
  text-align: center;
  font-size: 14px;
  color: #9ca3af;
}

@media (prefers-reduced-motion: reduce) {
  .stat-card, .issue-item, .nav-item {
    transition: none;
  }
}
</style>
