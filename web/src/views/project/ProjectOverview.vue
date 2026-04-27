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
          <div class="stat-card" @click="goBoard('open,reopened')">
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
          <div class="stat-card" @click="goBoard('resolved,closed')">
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
import { getIssueList, getProjectOverviewStats } from '@/api/issue'
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
    const { data } = await getProjectOverviewStats(projectKey.value)
    stats.open = data.data.pending
    stats.inProgress = data.data.in_progress
    stats.pendingReview = data.data.pending_review
    stats.resolved = data.data.completed
  } catch (e) {
    console.error('Failed to load stats:', e)
  }
}

const goBoard = (status?: string) => {
  const query = status ? { status } : {}
  router.push({ path: `/projects/${projectKey.value}/board`, query })
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
  background: var(--td-bg-card);
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
  color: var(--td-text-white);
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
  color: var(--td-text-primary);
  margin: 0;
  line-height: 1.3;
}

.project-key-badge {
  font-size: 11px;
  font-weight: 600;
  color: var(--td-color-primary);
  background: var(--td-tag-primary-bg);
  padding: 2px 8px;
  border-radius: 4px;
  letter-spacing: 0.03em;
}

.project-desc {
  font-size: 13px;
  color: var(--td-text-secondary);
  margin: 0;
  line-height: 1.5;

  &.muted {
    color: var(--td-text-disabled);
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
  border-top: 1px solid var(--td-divider-color);
  padding-top: 12px;

  .nav-item {
    display: inline-flex;
    align-items: center;
    gap: 5px;
    padding: 6px 14px;
    border-radius: 6px;
    font-size: 13px;
    color: var(--td-text-secondary);
    text-decoration: none;
    transition: background-color 150ms ease-out, color 150ms ease-out;

    .el-icon {
      font-size: 14px;
    }

    &:hover {
      background: var(--td-bg-section);
      color: var(--td-text-primary);
    }

    &.active,
    &.router-link-exact-active {
      background: var(--td-tag-primary-bg);
      color: var(--td-color-primary);
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
  background: var(--td-bg-card);
  border-radius: 12px;
  cursor: pointer;
  border: 1px solid var(--td-divider-color);
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

  &.todo { background: var(--td-tag-warning-border); color: var(--td-color-warning); }
  &.progress { background: var(--td-tag-primary-border); color: var(--td-color-primary); }
  &.review { background: var(--td-tag-indigo-bg); color: var(--td-tag-indigo-text); }
  &.done { background: var(--td-tag-success-border); color: var(--td-color-success); }
}

.stat-info {
  .stat-value {
    font-size: 24px;
    font-weight: 700;
    color: var(--td-text-primary);
    line-height: 1;
  }

  .stat-label {
    font-size: 13px;
    color: var(--td-text-secondary);
    margin-top: 4px;
  }
}

// 内容卡片
.content-card {
  border-radius: 12px;
  margin-bottom: 20px;

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
  background: var(--td-tag-primary-bg);
  color: var(--td-color-primary);
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 14px;
}

.card-title {
  font-size: 15px;
  font-weight: 600;
  color: var(--td-text-primary);
}

.card-count {
  font-size: 13px;
  color: var(--td-text-placeholder);
  font-weight: 400;
}

.view-all-link {
  display: flex;
  align-items: center;
  gap: 4px;
  font-size: 13px;
  color: var(--td-color-primary);
  text-decoration: none;

  &:hover {
    color: var(--td-color-primary-hover);
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
      background: var(--td-bg-page);
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

    &.P0 { background: var(--td-color-danger); }
    &.P1 { background: var(--td-color-warning); }
    &.P2 { background: var(--td-color-primary); }
    &.P3 { background: var(--td-color-success); }
  }

  .issue-key {
    font-size: 13px;
    font-weight: 600;
    color: var(--td-color-primary);
    flex-shrink: 0;
  }

  .issue-title-text {
    font-size: 14px;
    color: var(--td-text-primary);
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

  &.open { background: var(--td-tag-warning-border); color: var(--td-tag-orange-text); .status-dot { background: var(--td-color-warning); } }
  &.in_progress { background: var(--td-tag-primary-border); color: var(--td-tag-primary-text); .status-dot { background: var(--td-color-primary); } }
  &.pending_review { background: var(--td-tag-indigo-bg); color: var(--td-tag-indigo-text); .status-dot { background: var(--td-tag-indigo-text); } }
  &.resolved { background: var(--td-tag-success-border); color: var(--td-tag-success-text); .status-dot { background: var(--td-color-success); } }
  &.closed { background: var(--td-bg-section); color: var(--td-text-regular); .status-dot { background: var(--td-text-secondary); } }
  &.reopened { background: var(--td-tag-danger-border); color: var(--td-color-danger); .status-dot { background: var(--td-color-danger); } }
  &.merged { background: var(--td-tag-purple-bg); color: var(--td-tag-purple-text); .status-dot { background: var(--td-tag-purple-text); } }
}

.mini-avatar {
  width: 26px;
  height: 26px;
  border-radius: 6px;
  background: var(--td-color-primary);
  color: var(--td-text-white);
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
  background: var(--td-color-primary);
  color: var(--td-text-white);
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
    color: var(--td-text-primary);
  }

  .member-role {
    font-size: 12px;
    color: var(--td-text-placeholder);
  }
}

.empty-placeholder {
  padding: 40px;
  text-align: center;
  font-size: 14px;
  color: var(--td-text-placeholder);
}

@media (prefers-reduced-motion: reduce) {
  .stat-card, .issue-item, .nav-item {
    transition: none;
  }
}
</style>
