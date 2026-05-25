<template>
  <div class="project-board-container">
    <!-- 页面头部 -->
    <header class="board-header">
      <div class="header-left">
        <button class="back-btn" type="button" :title="`返回 ${project?.name || projectKey} 概览`" @click="$router.push(`/projects/${projectKey}`)">
          <el-icon><ArrowLeft /></el-icon>
        </button>
        <div class="project-identity">
          <div class="project-icon" :style="{ background: getProjectColor(projectKey) }">
            {{ projectKey.substring(0, 2).toUpperCase() }}
          </div>
          <div class="project-info">
            <div class="project-name">{{ project?.name || projectKey }}</div>
            <div class="project-section">工单看板</div>
          </div>
        </div>
      </div>
      <nav class="project-nav">
        <router-link :to="`/projects/${projectKey}`" class="nav-item">
          <el-icon><DataAnalysis /></el-icon>
          <span>概览</span>
        </router-link>
        <router-link :to="`/projects/${projectKey}/board`" class="nav-item active">
          <el-icon><Grid /></el-icon>
          <span>看板</span>
        </router-link>
        <router-link :to="`/projects/${projectKey}/settings`" class="nav-item">
          <el-icon><Setting /></el-icon>
          <span>设置</span>
        </router-link>
      </nav>
    </header>

    <!-- 分���主体 -->
    <div class="board-body">
      <!-- 左侧工单列表 -->
      <div class="board-left">
        <BoardIssueList
          ref="boardListRef"
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
          <TdEmptyState
            preset="first-time"
            tone="primary"
            title="从左侧列表选择一个工单"
            description="或点击「创建」开始一个新工单"
          >
            <el-button type="primary" @click="handleCreateIssue">
              <el-icon><Plus /></el-icon>创建工单
            </el-button>
          </TdEmptyState>
        </div>
      </div>
    </div>

    <!-- 创建工单对话框 -->
    <CreateIssueDialog
      v-model="createDialogVisible"
      :fixed-project-key="projectKey"
      @created="handleIssueCreated"
    />
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ArrowLeft, DataAnalysis, Grid, Setting, Plus } from '@element-plus/icons-vue'
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
const boardListRef = ref<InstanceType<typeof BoardIssueList> | null>(null)

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

// 创建工单后: 刷新列表并选中新工单
const handleIssueCreated = async (issueKey: string) => {
  await boardListRef.value?.reload()
  handleSelectIssue(issueKey)
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
  padding: 0 var(--td-space-5);
  height: 60px;
  background: var(--td-bg-card);
  border-bottom: 1px solid var(--td-border-color);
  flex-shrink: 0;
  gap: var(--td-space-4);
}

.header-left {
  display: flex;
  align-items: center;
  gap: var(--td-space-3);
  min-width: 0;
}

.back-btn {
  width: 32px;
  height: 32px;
  border-radius: var(--td-radius-md);
  display: inline-flex;
  align-items: center;
  justify-content: center;
  background: transparent;
  border: none;
  color: var(--td-text-secondary);
  cursor: pointer;
  flex-shrink: 0;
  font-size: var(--td-font-md);
  transition: var(--td-transition-bg), var(--td-transition-color);

  &:hover {
    background: var(--td-bg-section);
    color: var(--td-color-primary);
  }
}

.project-identity {
  display: flex;
  align-items: center;
  gap: var(--td-space-3);
  min-width: 0;
}

.project-icon {
  width: 36px;
  height: 36px;
  border-radius: var(--td-radius-md);
  display: flex;
  align-items: center;
  justify-content: center;
  color: var(--td-text-white);
  font-weight: var(--td-weight-bold);
  font-size: var(--td-font-sm);
  letter-spacing: var(--td-tracking-wide);
  flex-shrink: 0;
  box-shadow: var(--td-elevation-1);
}

.project-info {
  display: flex;
  flex-direction: column;
  gap: 1px;
  min-width: 0;
}

.project-name {
  font-size: var(--td-font-md);
  font-weight: var(--td-weight-semibold);
  color: var(--td-text-primary);
  line-height: var(--td-leading-tight);
  letter-spacing: var(--td-tracking-tight);
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.project-section {
  font-size: var(--td-font-xs);
  color: var(--td-text-placeholder);
  font-weight: var(--td-weight-medium);
  letter-spacing: var(--td-tracking-wide);
  text-transform: uppercase;
  line-height: 1.4;
}

.project-nav {
  display: flex;
  align-items: center;
  gap: 2px;
  flex-shrink: 0;

  .nav-item {
    display: inline-flex;
    align-items: center;
    gap: var(--td-space-2);
    padding: var(--td-space-2) var(--td-space-3);
    border-radius: var(--td-radius-md);
    font-size: var(--td-font-base);
    font-weight: var(--td-weight-medium);
    color: var(--td-text-secondary);
    text-decoration: none;
    transition: var(--td-transition-bg), var(--td-transition-color);
    position: relative;

    .el-icon {
      font-size: 15px;
    }

    &:hover {
      background: var(--td-bg-section);
      color: var(--td-text-primary);
    }

    &.active,
    &.router-link-exact-active {
      background: var(--td-tag-primary-bg);
      color: var(--td-color-primary);
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
  width: clamp(320px, 28vw, 400px);
  border-right: 1px solid var(--td-border-color);
  overflow: hidden;
  display: flex;
  flex-direction: column;
  background: var(--td-bg-card);
}

.board-right {
  flex: 1;
  overflow-y: auto;
  background: var(--td-bg-page);
  position: relative;

  // 微妙网格装饰背景, 减少纯色空白单调感
  background-image:
    radial-gradient(circle at 1px 1px, var(--td-border-color-light, rgba(255, 255, 255, 0.04)) 1px, transparent 0);
  background-size: 24px 24px;
}

// ---- 空状态容器 ----
.empty-detail {
  display: flex;
  justify-content: center;
  align-items: center;
  height: 100%;
  min-height: 360px;
}

@media (prefers-reduced-motion: reduce) {
  .back-btn,
  .nav-item {
    transition: none;
  }
}

@media (max-width: 768px) {
  .board-header {
    padding: 0 var(--td-space-3);

    .nav-item span {
      display: none;
    }
  }
}
</style>
