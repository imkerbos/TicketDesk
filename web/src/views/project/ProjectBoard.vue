<template>
  <div class="project-board-container">
    <!-- 页面头部 -->
    <div class="board-header">
      <div class="header-left">
        <el-button class="back-btn" circle size="small" @click="$router.push(`/projects/${projectKey}`)">
          <el-icon><ArrowLeft /></el-icon>
        </el-button>
        <div class="project-identity">
          <div class="project-icon" :style="{ background: getProjectColor(projectKey) }">
            {{ projectKey.substring(0, 2).toUpperCase() }}
          </div>
          <div class="project-info">
            <el-breadcrumb separator="/" class="header-breadcrumb">
              <el-breadcrumb-item :to="{ path: '/projects' }">项目</el-breadcrumb-item>
              <el-breadcrumb-item>{{ project?.name || projectKey }}</el-breadcrumb-item>
            </el-breadcrumb>
            <h2 class="project-title">工单看板</h2>
          </div>
        </div>
      </div>
      <div class="project-nav">
        <router-link :to="`/projects/${projectKey}`" class="nav-item">
          <el-icon><DataAnalysis /></el-icon>
          概览
        </router-link>
        <router-link :to="`/projects/${projectKey}/board`" class="nav-item active">
          <el-icon><Grid /></el-icon>
          看板
        </router-link>
        <router-link :to="`/projects/${projectKey}/settings`" class="nav-item">
          <el-icon><Setting /></el-icon>
          设置
        </router-link>
      </div>
    </div>

    <!-- 分���主体 -->
    <div class="board-body">
      <!-- 左侧工单列表 -->
      <div class="board-left">
        <BoardIssueList
          :project-key="projectKey"
          :selected-key="selectedIssueKey"
          :initial-status="initialStatus"
          @select="handleSelectIssue"
          @create="handleCreateIssue"
        />
      </div>

      <!-- 右侧详情 -->
      <div class="board-right">
        <IssueDetail
          v-if="selectedIssueKey"
          :key="selectedIssueKey"
          embedded
          :issue-key="selectedIssueKey"
          :on-navigate-issue="handleNavigateIssue"
          :on-deleted="handleDeleted"
        />
        <div v-else class="empty-detail">
          <div class="empty-inner">
            <div class="empty-icon">
              <el-icon :size="48"><Tickets /></el-icon>
            </div>
            <p class="empty-text">从左侧列表选择一个工单</p>
            <p class="empty-hint">或者创建一个新工单开始工作</p>
          </div>
        </div>
      </div>
    </div>

    <!-- 创建工单对话框 -->
    <CreateIssueDialog
      v-model="createDialogVisible"
      :fixed-project-key="projectKey"
      @created="handleSelectIssue"
    />
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ArrowLeft, DataAnalysis, Grid, Setting, Tickets } from '@element-plus/icons-vue'
import BoardIssueList from './components/BoardIssueList.vue'
import IssueDetail from '@/views/issue/IssueDetail.vue'
import { getProjectDetail } from '@/api/project'
import type { Project } from '@/types/project'
import CreateIssueDialog from '@/components/CreateIssueDialog.vue'

const route = useRoute()
const router = useRouter()

const projectKey = computed(() => route.params.key as string)
const initialStatus = computed(() => (route.query.status as string) || '')
const selectedIssueKey = ref('')
const project = ref<Project | null>(null)

// 创建工单
const createDialogVisible = ref(false)

// 初始化选中工单
const initSelectedKey = () => {
  const issueKey = route.params.issueKey as string
  selectedIssueKey.value = issueKey || ''
}

const loadProject = async () => {
  try {
    const { data } = await getProjectDetail(projectKey.value)
    project.value = data.data
  } catch {
    // ignored
  }
}

const handleSelectIssue = (issueKey: string) => {
  selectedIssueKey.value = issueKey
  router.replace(`/projects/${projectKey.value}/board/${issueKey}`)
}

const handleNavigateIssue = (issueKey: string) => {
  selectedIssueKey.value = issueKey
  router.replace(`/projects/${projectKey.value}/board/${issueKey}`)
}

const handleDeleted = () => {
  selectedIssueKey.value = ''
  router.replace(`/projects/${projectKey.value}/board`)
}

const handleCreateIssue = () => {
  createDialogVisible.value = true
}

const projectColors = ['#3b82f6', '#ef4444', '#10b981', '#f59e0b', '#8b5cf6', '#ec4899']
const getProjectColor = (key: string) => {
  let hash = 0
  for (let i = 0; i < key.length; i++) hash = key.charCodeAt(i) + ((hash << 5) - hash)
  return projectColors[Math.abs(hash) % projectColors.length]
}

// 浏览器前进/后退
watch(
  () => route.params.issueKey,
  (newKey) => {
    const key = newKey as string
    if (key !== selectedIssueKey.value) {
      selectedIssueKey.value = key || ''
    }
  }
)

onMounted(() => {
  initSelectedKey()
  loadProject()
})
</script>

<style scoped lang="scss">
.project-board-container {
  display: flex;
  flex-direction: column;
  height: calc(100vh - 64px);
  margin: -24px;
  background: var(--td-bg-page);
}

// ---- 头部 ----
.board-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 0 24px;
  height: 56px;
  background: var(--td-bg-card);
  border-bottom: 1px solid var(--td-border-color);
  flex-shrink: 0;
}

.header-left {
  display: flex;
  align-items: center;
  gap: 8px;
}

.back-btn {
  color: var(--td-text-secondary);
  border-color: var(--td-border-color);
  flex-shrink: 0;

  &:hover {
    color: var(--td-color-primary);
    border-color: var(--td-color-primary);
  }
}

.project-identity {
  display: flex;
  align-items: center;
  gap: 12px;
}

.project-icon {
  width: 32px;
  height: 32px;
  border-radius: 8px;
  display: flex;
  align-items: center;
  justify-content: center;
  color: var(--td-text-white);
  font-weight: 700;
  font-size: 12px;
  flex-shrink: 0;
}

.project-info {
  display: flex;
  flex-direction: column;
  gap: 1px;
}

.header-breadcrumb {
  :deep(.el-breadcrumb__inner) {
    font-size: 12px;
    color: var(--td-text-placeholder);
  }

  :deep(.el-breadcrumb__separator) {
    font-size: 12px;
    color: var(--td-text-disabled);
  }
}

.project-title {
  font-size: 15px;
  font-weight: 600;
  color: var(--td-text-primary);
  margin: 0;
  line-height: 1.2;
}

.project-nav {
  display: flex;
  align-items: center;
  gap: 2px;
  height: 100%;

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

// ---- 分屏主体 ----
.board-body {
  display: flex;
  flex: 1;
  overflow: hidden;
}

.board-left {
  width: 360px;
  min-width: 360px;
  border-right: 1px solid var(--td-border-color);
  overflow: hidden;
  display: flex;
  flex-direction: column;
}

.board-right {
  flex: 1;
  overflow-y: auto;
  background: var(--td-bg-section);
}

// ---- 空状态 ----
.empty-detail {
  display: flex;
  justify-content: center;
  align-items: center;
  height: 100%;
  background: var(--td-bg-section);
}

.empty-inner {
  text-align: center;
}

.empty-icon {
  width: 80px;
  height: 80px;
  margin: 0 auto 20px;
  background: var(--td-color-primary-light);
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  color: var(--td-color-primary);
}

.empty-text {
  font-size: 15px;
  color: var(--td-text-regular);
  margin: 0 0 6px;
  font-weight: 500;
}

.empty-hint {
  font-size: 13px;
  color: var(--td-text-placeholder);
  margin: 0;
}

@media (prefers-reduced-motion: reduce) {
  .nav-item {
    transition: none;
  }
}
</style>
