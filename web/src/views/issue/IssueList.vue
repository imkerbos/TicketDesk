<template>
  <div class="issue-list-container">
    <!-- 页面头部 -->
    <TdPageHeader>
      <template #leading>
        <div class="page-header-icon">
          <el-icon :size="20"><Tickets /></el-icon>
        </div>
      </template>
      <template #title>工单管理</template>
      <template #subtitle>跟踪和管理所有工单任务</template>
      <template #actions>
        <el-button type="primary" @click="handleCreate">
          <el-icon><Plus /></el-icon>创建工单
        </el-button>
      </template>
    </TdPageHeader>

    <!-- 分类 Tab -->
    <el-tabs v-model="queryParams.category" class="category-tabs" @tab-change="handleCategoryChange">
      <el-tab-pane label="全部工单" name="" />
      <el-tab-pane label="常规工单" name="normal" />
      <el-tab-pane label="告警工单" name="alert" />
    </el-tabs>

    <!-- 过滤器 -->
    <el-card shadow="never" class="filter-card">
      <div class="quick-filter-bar">
        <div class="quick-filter-left">
          <span class="quick-filter-label">快捷筛选</span>
          <el-button-group>
            <el-button size="small" :type="activeQuickFilter === '' ? 'primary' : 'default'" @click="applyQuickFilter('')">全部</el-button>
            <el-button size="small" :type="activeQuickFilter === 'my-todo' ? 'primary' : 'default'" @click="applyQuickFilter('my-todo')">
              我的待办
            </el-button>
            <el-button size="small" :type="activeQuickFilter === 'my-created' ? 'primary' : 'default'" @click="applyQuickFilter('my-created')">
              我创建的
            </el-button>
          </el-button-group>
          <el-tag
            v-if="activeQuickFilterLabel"
            type="info"
            effect="plain"
            closable
            class="active-filter-tag"
            @close="applyQuickFilter('')"
          >
            {{ activeQuickFilterLabel }}
          </el-tag>
        </div>
        <div class="quick-filter-right">
          <el-select
            v-model="selectedSavedViewId"
            clearable
            placeholder="已保存视图"
            class="saved-view-select"
            @change="handleSavedViewChange"
          >
            <el-option v-for="view in savedViews" :key="view.id" :label="view.name" :value="view.id" />
          </el-select>
          <el-button @click="handleSaveCurrentView">保存当前视图</el-button>
          <el-button type="danger" plain :disabled="!selectedSavedViewId" @click="handleDeleteCurrentView">删除视图</el-button>
        </div>
      </div>
      <div class="filter-content">
        <div class="filter-left">
          <el-input
            v-model="queryParams.keyword"
            placeholder="搜索工单标题"
            clearable
            class="search-input"
            @clear="handleQuery"
            @keyup.enter="handleQuery"
          >
            <template #prefix>
              <el-icon><Search /></el-icon>
            </template>
          </el-input>
          <el-select v-model="queryParams.project_key" placeholder="项目" clearable class="filter-select" @change="handleProjectFilterChange">
            <el-option v-for="p in projects" :key="p.project_key" :label="p.name" :value="p.project_key" />
          </el-select>
          <el-select v-model="queryParams.status" placeholder="状态" clearable class="filter-select" @change="handleQuery">
            <el-option label="待处理" value="open" />
            <el-option label="进行中" value="in_progress" />
            <el-option label="待确认" value="pending_review" />
            <el-option label="已完成" value="resolved" />
            <el-option label="已终止" value="closed" />
            <el-option label="已合并" value="merged" />
          </el-select>
          <el-select v-model="queryParams.priority" placeholder="优先级" clearable class="filter-select-sm" @change="handleQuery">
            <el-option label="P0 - 紧急" value="P0" />
            <el-option label="P1 - 高" value="P1" />
            <el-option label="P2 - 中" value="P2" />
            <el-option label="P3 - 低" value="P3" />
          </el-select>
          <el-select v-model="queryParams.assignee_id" placeholder="指派人" clearable filterable class="filter-select" @change="handleAssigneeFilterChange">
            <el-option v-for="u in users" :key="u.id" :label="u.display_name" :value="u.id" />
          </el-select>
          <el-select v-model="queryParams.reporter_id" placeholder="报告人" clearable filterable class="filter-select" @change="handleReporterFilterChange">
            <el-option v-for="u in users" :key="u.id" :label="u.display_name" :value="u.id" />
          </el-select>
          <el-select v-model="queryParams.issue_type_id" placeholder="工单类型" clearable class="filter-select" @change="handleQuery">
            <el-option v-for="t in filterIssueTypes" :key="t.id" :label="t.display_name" :value="t.id" />
          </el-select>
          <el-select v-model="queryParams.epic_id" placeholder="Epic" clearable filterable class="filter-select" @change="handleQuery">
            <el-option v-for="e in filterEpics" :key="e.id" :label="`${e.issue_key}: ${e.title}`" :value="e.id" />
          </el-select>
          <el-date-picker
            v-model="dateRange"
            type="daterange"
            range-separator="至"
            start-placeholder="开始日期"
            end-placeholder="结束日期"
            value-format="YYYY-MM-DD"
            class="date-range-picker"
            @change="handleDateRangeChange"
          />
        </div>
        <div class="filter-right">
          <el-button type="primary" :icon="Search" @click="handleQuery">查询</el-button>
          <el-button :icon="Refresh" @click="handleReset">重置</el-button>
        </div>
      </div>
    </el-card>

    <!-- KPI 统计卡片 -->
    <div class="stats-row">
      <div class="stat-card stat-total">
        <div class="stat-value">{{ issueStats.total }}</div>
        <div class="stat-label">工单总数</div>
      </div>
      <div class="stat-card stat-resolved">
        <div class="stat-value">{{ issueStats.resolved }}</div>
        <div class="stat-label">已完成</div>
      </div>
      <div class="stat-card stat-rate">
        <div class="stat-value">{{ issueStats.completion_rate }}%</div>
        <div class="stat-label">完成率</div>
      </div>
      <div class="stat-card stat-hours">
        <div class="stat-value">{{ issueStats.avg_resolve_hours }}h</div>
        <div class="stat-label">平均解决时长</div>
      </div>
    </div>

    <!-- 工具栏 + 内容 -->
    <el-card shadow="never" class="table-card">
      <div class="toolbar">
        <div class="toolbar-left">
          <el-radio-group v-model="viewMode" class="view-toggle" @change="handleViewModeChange">
            <el-radio-button value="table">
              <el-icon><List /></el-icon>
              表格
            </el-radio-button>
            <el-radio-button value="kanban">
              <el-icon><Grid /></el-icon>
              看板
            </el-radio-button>
          </el-radio-group>
          <span class="total-count">共 {{ total }} 条</span>
        </div>
      </div>

      <!-- 表格视图 -->
      <div v-if="viewMode === 'table'" class="table-view">
        <el-table
          v-loading="loading"
          :data="issueList"
          style="width: 100%"
          :row-class-name="() => 'clickable-row'"
          :default-sort="defaultSort as any"
          @row-click="handleRowClick"
          @sort-change="handleSortChange"
        >
          <el-table-column prop="issue_key" label="工单号" width="120" fixed sortable="custom" :sort-orders="['ascending', 'descending', null]">
            <template #default="{ row }">
              <el-link type="primary" underline="never" @click.stop="navigateTo(`/issues/${row.issue_key}`, $event)">
                <span class="issue-key-text">{{ row.issue_key }}</span>
              </el-link>
            </template>
          </el-table-column>
          <el-table-column prop="title" label="标题" min-width="280">
            <template #default="{ row }">
              <div class="issue-title-cell">
                <div class="priority-dot" :class="row.priority"></div>
                <span class="title-text">{{ row.title }}</span>
                <el-tag v-if="getDueDateStatus(row) === 'overdue'" type="danger" size="small" class="due-tag">已超时</el-tag>
                <el-tag v-else-if="getDueDateStatus(row) === 'due_soon'" type="warning" size="small" class="due-tag">即将超时</el-tag>
              </div>
            </template>
          </el-table-column>
          <el-table-column prop="project_key" label="项目" width="110">
            <template #default="{ row }">
              <el-tag size="small" effect="plain" type="info">{{ row.project_key }}</el-tag>
            </template>
          </el-table-column>
          <el-table-column prop="issue_type_id" label="类型" width="90" sortable="custom" :sort-orders="['ascending', 'descending', null]">
            <template #default="{ row }">
              <span class="type-text">{{ row.issue_type?.display_name || '-' }}</span>
            </template>
          </el-table-column>
          <el-table-column prop="priority" label="优先级" width="80" align="center" sortable="custom" :sort-orders="['ascending', 'descending', null]">
            <template #default="{ row }">
              <el-tag :type="getPriorityType(row.priority)" size="small" effect="dark" class="priority-tag">
                {{ row.priority }}
              </el-tag>
            </template>
          </el-table-column>
          <el-table-column prop="status" label="状态" width="100" align="center" sortable="custom" :sort-orders="['ascending', 'descending', null]">
            <template #default="{ row }">
              <div class="status-badge" :class="row.status">
                <span class="status-dot"></span>
                <span>{{ getStatusText(row.status) }}</span>
              </div>
            </template>
          </el-table-column>
          <el-table-column label="指派人" width="110" show-overflow-tooltip>
            <template #default="{ row }">
              <div v-if="row.assignee" class="assignee-cell">
                <div class="mini-avatar">{{ row.assignee.display_name?.charAt(0) }}</div>
                <span>{{ row.assignee.display_name }}</span>
              </div>
              <span v-else class="text-muted">未指派</span>
            </template>
          </el-table-column>
          <el-table-column label="报告人" width="110" show-overflow-tooltip>
            <template #default="{ row }">
              <div v-if="row.reporter" class="assignee-cell">
                <div class="mini-avatar">{{ row.reporter.display_name?.charAt(0) }}</div>
                <span>{{ row.reporter.display_name }}</span>
              </div>
              <span v-else class="text-muted">未知</span>
            </template>
          </el-table-column>
          <el-table-column prop="created_at" label="创建时间" width="150" show-overflow-tooltip sortable="custom" :sort-orders="['ascending', 'descending', null]">
            <template #default="{ row }">
              <div class="time-cell">
                <el-icon><Clock /></el-icon>
                <span>{{ formatTime(row.created_at) }}</span>
              </div>
            </template>
          </el-table-column>
          <el-table-column label="操作" width="90" fixed="right" align="center">
            <template #default="{ row }">
              <el-button type="danger" link @click.stop="handleDeleteIssue(row)">
                <el-icon><Delete /></el-icon>
                删除
              </el-button>
            </template>
          </el-table-column>
        </el-table>

        <div class="pagination-wrapper">
          <el-pagination
            v-model:current-page="queryParams.page"
            v-model:page-size="queryParams.page_size"
            :total="total"
            :page-sizes="[10, 20, 50, 100]"
            layout="slot, sizes, prev, pager, next, jumper"
            @size-change="handleQuery"
            @current-change="handlePageChange"
          >
            <span class="el-pagination__total">共 {{ hasMore ? `${total.toLocaleString()}+` : total.toLocaleString() }} 条</span>
          </el-pagination>
        </div>
      </div>

      <!-- 看板视图 -->
      <div v-if="viewMode === 'kanban'" v-loading="loading" class="kanban-view">
        <div class="kanban-board">
          <div v-for="column in kanbanColumns" :key="column.status" class="kanban-column">
            <div class="column-header" :class="column.status">
              <div class="column-title-group">
                <span class="column-dot" :class="column.status"></span>
                <span class="column-title">{{ column.label }}</span>
              </div>
              <span class="column-count">{{ column.issues.length }}</span>
            </div>
            <div class="column-content">
              <div
                v-for="issue in column.issues"
                :key="issue.id"
                class="kanban-card"
                @click="navigateTo(`/issues/${issue.issue_key}`, $event)"
              >
                <div class="card-header">
                  <span class="issue-key">{{ issue.issue_key }}</span>
                  <el-tag :type="getPriorityType(issue.priority)" size="small" effect="dark">
                    {{ issue.priority }}
                  </el-tag>
                </div>
                <div class="card-title">{{ issue.title }}</div>
                <div class="card-footer">
                  <el-tag size="small" effect="plain" type="info">{{ issue.project_key }}</el-tag>
                  <div v-if="issue.assignee" class="card-assignee">
                    <div class="mini-avatar sm">{{ issue.assignee.display_name?.charAt(0) || '?' }}</div>
                    <span>{{ issue.assignee.display_name }}</span>
                  </div>
                </div>
              </div>
              <div v-if="column.issues.length === 0" class="empty-column">
                暂无工单
              </div>
            </div>
          </div>
        </div>
      </div>
    </el-card>

    <!-- 创建工单对话框 -->
    <CreateIssueDialog
      v-model="createDialogVisible"
      :default-project-key="queryParams.project_key || ''"
      :projects="projects"
      @created="(key: string) => router.push(`/issues/${key}`)"
    />
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, computed, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Search, Refresh, Plus, List, Grid, Tickets, Clock, Delete } from '@element-plus/icons-vue'
import { getIssueList, getIssueListStats, deleteIssue } from '@/api/issue'
import { getAllProjects, getAllIssueTypes, getProjectIssueTypes } from '@/api/project'
import { getAllUsers } from '@/api/user'
import type { Issue, IssueStatus, IssuePriority, KanbanColumn, IssueListStats } from '@/types/issue'
import type { Project, ProjectIssueType } from '@/types/project'
import type { UserOption } from '@/types/user'
import { useUserStore } from '@/stores/user'
import CreateIssueDialog from '@/components/CreateIssueDialog.vue'
import dayjs from 'dayjs'

const route = useRoute()
const router = useRouter()
const userStore = useUserStore()

const loading = ref(false)
const issueList = ref<Issue[]>([])
const total = ref(0)
const hasMore = ref(false)
const viewMode = ref<'table' | 'kanban'>('table')
const projects = ref<Project[]>([])
const users = ref<UserOption[]>([])
const filterIssueTypes = ref<ProjectIssueType[]>([])
const filterEpics = ref<Issue[]>([])

// 从 URL query 初始化筛选条件
const initQuery = route.query

const dateRange = ref<[string, string] | null>(
  initQuery.start_date && initQuery.end_date
    ? [initQuery.start_date as string, initQuery.end_date as string]
    : null,
)
const issueStats = reactive<IssueListStats>({
  total: 0,
  resolved: 0,
  completion_rate: 0,
  avg_resolve_hours: 0,
})

type QuickFilterKey = '' | 'my-todo' | 'my-created'
interface SavedViewSnapshot {
  quick_filter: QuickFilterKey
  project_key?: string
  status?: IssueStatus
  priority?: IssuePriority
  assignee_id?: number
  reporter_id?: number
  issue_type_id?: number
  epic_id?: number
  keyword?: string
  category: '' | 'normal' | 'alert'
  sort_by?: string
  order?: 'asc' | 'desc'
}
interface SavedFilterView {
  id: string
  name: string
  snapshot: SavedViewSnapshot
  created_at: string
}

const dashboardFilter = ((initQuery.filter as string) || '') as QuickFilterKey
const currentUserId = userStore.user?.id
const initAssigneeID = initQuery.assignee_id
  ? Number(initQuery.assignee_id)
  : (dashboardFilter === 'my-todo' ? currentUserId : undefined)
const initReporterID = initQuery.reporter_id
  ? Number(initQuery.reporter_id)
  : (dashboardFilter === 'my-created' ? currentUserId : undefined)

const queryParams = reactive({
  page: initQuery.page ? Number(initQuery.page) : 1,
  page_size: initQuery.page_size ? Number(initQuery.page_size) : 20,
  project_key: (initQuery.project_key as string) || undefined,
  status: (initQuery.status as IssueStatus | undefined) || undefined,
  priority: (initQuery.priority as IssuePriority | undefined) || undefined,
  assignee_id: initAssigneeID,
  reporter_id: initReporterID,
  issue_type_id: initQuery.issue_type_id ? Number(initQuery.issue_type_id) : undefined,
  epic_id: initQuery.epic_id ? Number(initQuery.epic_id) : undefined,
  keyword: (initQuery.keyword as string) || undefined,
  category: ((initQuery.category as string) || '') as '' | 'normal' | 'alert',
  sort_by: (initQuery.sort_by as string) || undefined,
  order: (initQuery.order as 'asc' | 'desc' | undefined) || undefined,
  start_date: (initQuery.start_date as string) || undefined,
  end_date: (initQuery.end_date as string) || undefined,
})

const activeQuickFilter = ref<QuickFilterKey>(dashboardFilter)
const savedViews = ref<SavedFilterView[]>([])
const selectedSavedViewId = ref((initQuery.view as string) || '')
const activeQuickFilterLabel = computed(() => {
  if (activeQuickFilter.value === 'my-todo') return '已应用：我的待办'
  if (activeQuickFilter.value === 'my-created') return '已应用：我创建的'
  return ''
})
const savedViewsStorageKey = computed(() => `issue_list_saved_views_${userStore.user?.id || 'anonymous'}`)

const buildSnapshot = (): SavedViewSnapshot => ({
  quick_filter: activeQuickFilter.value,
  project_key: queryParams.project_key,
  status: queryParams.status,
  priority: queryParams.priority,
  assignee_id: queryParams.assignee_id,
  reporter_id: queryParams.reporter_id,
  issue_type_id: queryParams.issue_type_id,
  epic_id: queryParams.epic_id,
  keyword: queryParams.keyword,
  category: queryParams.category,
  sort_by: queryParams.sort_by,
  order: queryParams.order,
})

const applySnapshot = (snapshot: SavedViewSnapshot) => {
  activeQuickFilter.value = snapshot.quick_filter || ''
  queryParams.project_key = snapshot.project_key
  queryParams.status = snapshot.status
  queryParams.priority = snapshot.priority
  queryParams.assignee_id = snapshot.assignee_id
  queryParams.reporter_id = snapshot.reporter_id
  queryParams.issue_type_id = snapshot.issue_type_id
  queryParams.epic_id = snapshot.epic_id
  queryParams.keyword = snapshot.keyword
  queryParams.category = snapshot.category || ''
  queryParams.sort_by = snapshot.sort_by
  queryParams.order = snapshot.order
}

const loadSavedViewsFromStorage = () => {
  try {
    const raw = localStorage.getItem(savedViewsStorageKey.value)
    savedViews.value = raw ? JSON.parse(raw) : []
  } catch {
    savedViews.value = []
  }
}

const persistSavedViews = () => {
  localStorage.setItem(savedViewsStorageKey.value, JSON.stringify(savedViews.value))
}

// 筛选条件同步到 URL query params
const syncQueryToUrl = () => {
  const query: Record<string, string> = {}
  if (activeQuickFilter.value) query.filter = activeQuickFilter.value
  if (queryParams.keyword) query.keyword = queryParams.keyword
  if (queryParams.project_key) query.project_key = queryParams.project_key
  if (queryParams.status) query.status = queryParams.status
  if (queryParams.priority) query.priority = queryParams.priority
  if (queryParams.assignee_id) query.assignee_id = String(queryParams.assignee_id)
  if (queryParams.reporter_id) query.reporter_id = String(queryParams.reporter_id)
  if (queryParams.issue_type_id) query.issue_type_id = String(queryParams.issue_type_id)
  if (queryParams.epic_id) query.epic_id = String(queryParams.epic_id)
  if (queryParams.category) query.category = queryParams.category
  if (queryParams.sort_by) query.sort_by = queryParams.sort_by
  if (queryParams.order) query.order = queryParams.order
  if (queryParams.start_date) query.start_date = queryParams.start_date
  if (queryParams.end_date) query.end_date = queryParams.end_date
  if (selectedSavedViewId.value) query.view = selectedSavedViewId.value
  if (queryParams.page > 1) query.page = String(queryParams.page)
  router.replace({ query })
}

const kanbanColumns = computed<KanbanColumn[]>(() => {
  const columns: KanbanColumn[] = [
    { status: 'open', label: '待处理', issues: [] },
    { status: 'in_progress', label: '进行中', issues: [] },
    { status: 'pending_review', label: '待确认', issues: [] },
    { status: 'resolved', label: '已完成', issues: [] },
  ]
  issueList.value.forEach((issue) => {
    const column = columns.find((c) => c.status === issue.status)
    if (column) column.issues.push(issue)
  })
  return columns
})

const createDialogVisible = ref(false)

const loadData = async () => {
  loading.value = true
  try {
    const params = { ...queryParams }
    if (viewMode.value === 'kanban') params.page_size = 100
    const [listRes, statsRes] = await Promise.all([
      getIssueList(params),
      getIssueListStats(params),
    ])
    issueList.value = listRes.data.data.items
    total.value = listRes.data.data.total
    hasMore.value = listRes.data.data.has_more || false
    Object.assign(issueStats, statsRes.data.data)
  } catch {
    // ignored
  } finally {
    loading.value = false
  }
}

const loadFilterOptions = async () => {
  try {
    const [projectsRes, usersRes] = await Promise.all([getAllProjects(), getAllUsers()])
    projects.value = projectsRes.data.data
    users.value = usersRes.data.data
  } catch {
    // ignored
  }
}

const loadFilterIssueTypes = async () => {
  try {
    if (queryParams.project_key) {
      const { data } = await getProjectIssueTypes(queryParams.project_key)
      filterIssueTypes.value = data.data || []
    } else {
      const { data } = await getAllIssueTypes()
      filterIssueTypes.value = data.data || []
    }
  } catch {
    // ignored
  }
}

const loadFilterEpics = async () => {
  try {
    const epicType = filterIssueTypes.value.find(t => t.name.toLowerCase() === 'epic')
    if (!epicType) {
      filterEpics.value = []
      return
    }
    const params: Record<string, any> = { issue_type_id: epicType.id, page_size: 100 }
    if (queryParams.project_key) {
      params.project_key = queryParams.project_key
    }
    const { data } = await getIssueList(params)
    filterEpics.value = data.data.items || []
  } catch {
    // ignored
  }
}

const clearSelectedSavedView = () => {
  selectedSavedViewId.value = ''
}

const normalizeQuickFilter = () => {
  if (activeQuickFilter.value === 'my-todo' && queryParams.assignee_id !== currentUserId) {
    activeQuickFilter.value = ''
  }
  if (activeQuickFilter.value === 'my-created' && queryParams.reporter_id !== currentUserId) {
    activeQuickFilter.value = ''
  }
}

const applyQuickFilter = (filter: QuickFilterKey) => {
  if ((filter === 'my-todo' || filter === 'my-created') && !currentUserId) {
    ElMessage.warning('未识别当前用户，无法应用快捷筛选')
    return
  }
  activeQuickFilter.value = filter
  if (filter === 'my-todo') {
    queryParams.assignee_id = currentUserId
    queryParams.reporter_id = undefined
  } else if (filter === 'my-created') {
    queryParams.reporter_id = currentUserId
    queryParams.assignee_id = undefined
  } else {
    queryParams.assignee_id = undefined
    queryParams.reporter_id = undefined
  }
  clearSelectedSavedView()
  queryParams.page = 1
  syncQueryToUrl()
  loadData()
}

const handleSaveCurrentView = async () => {
  try {
    const promptResult = await ElMessageBox.prompt('请输入视图名称', '保存当前视图', {
      inputValue: activeQuickFilterLabel.value ? activeQuickFilterLabel.value.replace('已应用：', '') : '',
      inputPlaceholder: '例如：我的高优先级待办',
      confirmButtonText: '保存',
      cancelButtonText: '取消',
      inputValidator: (val: string) => !!val?.trim() || '视图名称不能为空',
    })
    if (typeof promptResult === 'string' || !promptResult || !('value' in promptResult)) return
    const name = promptResult.value.trim()
    const id = `${Date.now()}_${Math.random().toString(36).slice(2, 8)}`
    savedViews.value.unshift({
      id,
      name,
      snapshot: buildSnapshot(),
      created_at: new Date().toISOString(),
    })
    persistSavedViews()
    selectedSavedViewId.value = id
    syncQueryToUrl()
    ElMessage.success('视图已保存')
  } catch {
    // 用户取消
  }
}

const handleSavedViewChange = async (viewID?: string) => {
  if (!viewID) {
    selectedSavedViewId.value = ''
    syncQueryToUrl()
    return
  }
  const view = savedViews.value.find(v => v.id === viewID)
  if (!view) {
    ElMessage.warning('视图不存在')
    selectedSavedViewId.value = ''
    syncQueryToUrl()
    return
  }
  selectedSavedViewId.value = view.id
  applySnapshot(view.snapshot)
  queryParams.page = 1
  await loadFilterIssueTypes()
  await loadFilterEpics()
  syncQueryToUrl()
  loadData()
}

const handleDeleteCurrentView = async () => {
  if (!selectedSavedViewId.value) return
  const view = savedViews.value.find(v => v.id === selectedSavedViewId.value)
  if (!view) return
  try {
    await ElMessageBox.confirm(`确定删除视图「${view.name}」吗？`, '删除视图', { type: 'warning' })
    savedViews.value = savedViews.value.filter(v => v.id !== selectedSavedViewId.value)
    selectedSavedViewId.value = ''
    persistSavedViews()
    syncQueryToUrl()
    ElMessage.success('视图已删除')
  } catch {
    // 用户取消
  }
}

const handleCategoryChange = () => {
  clearSelectedSavedView()
  queryParams.page = 1
  syncQueryToUrl()
  loadData()
}
const handleQuery = () => {
  clearSelectedSavedView()
  normalizeQuickFilter()
  queryParams.page = 1
  syncQueryToUrl()
  loadData()
}
const handlePageChange = () => { syncQueryToUrl(); loadData() }

const handleAssigneeFilterChange = () => {
  clearSelectedSavedView()
  normalizeQuickFilter()
  queryParams.page = 1
  syncQueryToUrl()
  loadData()
}

const handleReporterFilterChange = () => {
  clearSelectedSavedView()
  normalizeQuickFilter()
  queryParams.page = 1
  syncQueryToUrl()
  loadData()
}

const handleDateRangeChange = (val: [string, string] | null) => {
  clearSelectedSavedView()
  queryParams.start_date = val ? val[0] : undefined
  queryParams.end_date = val ? val[1] : undefined
  queryParams.page = 1
  syncQueryToUrl()
  loadData()
}

// 排序：prop 到后端字段的映射
const sortFieldMap: Record<string, string> = {
  issue_key: 'id',
  issue_type_id: 'issue_type_id',
  priority: 'priority',
  status: 'status',
  created_at: 'created_at',
}

const defaultSort = computed(() => {
  if (!queryParams.sort_by) return {}
  // 反向查找 prop
  const prop = Object.entries(sortFieldMap).find(([, v]) => v === queryParams.sort_by)?.[0]
  if (!prop) return {}
  return { prop, order: queryParams.order === 'asc' ? 'ascending' : 'descending' }
})

const handleSortChange = ({ prop, order }: { prop: string; order: string | null }) => {
  if (!order) {
    queryParams.sort_by = undefined
    queryParams.order = undefined
  } else {
    queryParams.sort_by = sortFieldMap[prop] || prop
    queryParams.order = order === 'ascending' ? 'asc' : 'desc'
  }
  clearSelectedSavedView()
  queryParams.page = 1
  syncQueryToUrl()
  loadData()
}

const handleProjectFilterChange = async () => {
  clearSelectedSavedView()
  normalizeQuickFilter()
  queryParams.issue_type_id = undefined
  queryParams.epic_id = undefined
  queryParams.page = 1
  syncQueryToUrl()
  await loadFilterIssueTypes()
  await loadFilterEpics()
  loadData()
}

const handleReset = async () => {
  clearSelectedSavedView()
  activeQuickFilter.value = ''
  queryParams.page = 1
  queryParams.page_size = 20
  queryParams.project_key = undefined
  queryParams.status = undefined
  queryParams.priority = undefined
  queryParams.assignee_id = undefined
  queryParams.reporter_id = undefined
  queryParams.issue_type_id = undefined
  queryParams.epic_id = undefined
  queryParams.keyword = undefined
  queryParams.category = ''
  queryParams.sort_by = undefined
  queryParams.order = undefined
  queryParams.start_date = undefined
  queryParams.end_date = undefined
  dateRange.value = null
  syncQueryToUrl()
  await loadFilterIssueTypes()
  await loadFilterEpics()
  loadData()
}

const handleViewModeChange = () => { loadData() }
const handleRowClick = (row: Issue, _col: any, event: MouseEvent) => {
  const path = `/issues/${row.issue_key}`
  if (event.metaKey || event.ctrlKey) {
    window.open(router.resolve(path).href, '_blank')
  } else {
    router.push(path)
  }
}

const navigateTo = (path: string, event: MouseEvent) => {
  if (event.metaKey || event.ctrlKey) {
    window.open(router.resolve(path).href, '_blank')
  } else {
    router.push(path)
  }
}

const handleCreate = () => {
  createDialogVisible.value = true
}

const handleDeleteIssue = async (issue: Issue) => {
  try {
    await ElMessageBox.confirm(
      `确定要删除工单 "${issue.issue_key}" 吗？删除后评论、附件、工作流实例与活动记录将一并清理。`,
      '删除确认',
      { type: 'warning' }
    )
    await deleteIssue(issue.issue_key)
    ElMessage.success('删除成功')
    loadData()
  } catch (error) {
    if (error !== 'cancel') {
      // ignored
    }
  }
}

type TagType = 'primary' | 'success' | 'warning' | 'info' | 'danger'
const getPriorityType = (priority: string): TagType => {
  const map: Record<string, TagType> = { P0: 'danger', P1: 'warning', P2: 'info', P3: 'info' }
  return map[priority] || 'info'
}
const getStatusText = (status: string) => {
  const map: Record<string, string> = { open: '待处理', in_progress: '进行中', pending_review: '待确认', resolved: '已完成', closed: '已终止', merged: '已合并' }
  return map[status] || status
}
const formatTime = (time: string) => dayjs(time).format('YYYY-MM-DD HH:mm')

// SLA 目标（分钟），与后端 slaTargets 保持一致
const SLA_TARGETS: Record<string, number> = { P0: 60, P1: 240, P2: 1440, P3: 4320 }

const getDueDateStatus = (row: Issue): 'overdue' | 'due_soon' | null => {
  if (['resolved', 'closed', 'merged'].includes(row.status)) return null
  const now = dayjs()

  let deadline: ReturnType<typeof dayjs>
  let slaMinutes: number

  if (row.planned_end_date) {
    deadline = dayjs(row.planned_end_date).endOf('day')
    slaMinutes = deadline.diff(dayjs(row.created_at), 'minute')
  } else {
    slaMinutes = SLA_TARGETS[row.priority] || SLA_TARGETS.P2
    deadline = dayjs(row.created_at).add(slaMinutes, 'minute')
  }

  const minutesLeft = deadline.diff(now, 'minute', true)
  const threshold = slaMinutes * 0.25

  if (minutesLeft < 0) return 'overdue'
  if (minutesLeft < threshold) return 'due_soon'
  return null
}

onMounted(async () => {
  loadSavedViewsFromStorage()
  if (selectedSavedViewId.value) {
    const view = savedViews.value.find(v => v.id === selectedSavedViewId.value)
    if (view) {
      applySnapshot(view.snapshot)
    } else {
      selectedSavedViewId.value = ''
    }
  }
  await loadFilterOptions()
  await loadFilterIssueTypes()
  await loadFilterEpics()
  loadData()
})
</script>

<style scoped lang="scss">
.issue-list-container {
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

// 分类 Tab
.category-tabs {
  margin-bottom: 16px;

  :deep(.el-tabs__header) {
    margin: 0;
  }

  :deep(.el-tabs__nav-wrap::after) {
    height: 1px;
    background-color: var(--td-border-color);
  }

  :deep(.el-tabs__item) {
    font-size: 14px;
    font-weight: 500;
    padding: 0 20px;
  }
}

// 筛选卡片
.filter-card {
  margin-bottom: 20px;
  border-radius: 12px;

  :deep(.el-card__body) { padding: 16px 20px; }

  .quick-filter-bar {
    display: flex;
    justify-content: space-between;
    align-items: center;
    gap: 12px;
    margin-bottom: 14px;
    padding-bottom: 12px;
    border-bottom: 1px dashed #e5e7eb;
  }

  .quick-filter-left {
    display: flex;
    align-items: center;
    gap: 10px;
    flex-wrap: wrap;
  }

  .quick-filter-label {
    font-size: 13px;
    color: var(--td-text-secondary);
    font-weight: 500;
  }

  .active-filter-tag {
    margin-left: 2px;
  }

  .quick-filter-right {
    display: flex;
    align-items: center;
    gap: 8px;
    flex-wrap: wrap;
    justify-content: flex-end;
  }

  .saved-view-select {
    width: 180px;
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
    gap: 10px;
    flex-wrap: wrap;
  }

  .search-input { width: 220px; }
  .filter-select { width: 140px; }
  .filter-select-sm { width: 120px; }
}

// 表格卡片
.table-card {
  border-radius: 12px;

  :deep(.el-card__body) { padding: 20px; }
}

.toolbar {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 16px;

  .toolbar-left {
    display: flex;
    align-items: center;
    gap: 16px;
  }

  .total-count {
    font-size: 13px;
    color: var(--td-text-placeholder);
  }
}

// 表格视图
.table-view {
  :deep(.el-table) {
    border-radius: 8px;

    th.el-table__cell {
      background: var(--td-bg-page);
      font-weight: 600;
      color: var(--td-text-regular);
    }

    .clickable-row {
      cursor: pointer;
      transition: background-color 150ms ease-out;
      &:hover { background-color: var(--td-bg-page); }
      &:active { background-color: var(--td-bg-section); }
    }
  }

  .issue-key-text {
    font-weight: 600;
    font-size: 13px;
  }

  .issue-title-cell {
    display: flex;
    align-items: center;
    gap: 8px;

    .priority-dot {
      width: 8px;
      height: 8px;
      border-radius: 50%;
      flex-shrink: 0;

      &.P0 { background: var(--td-color-danger); }
      &.P1 { background: var(--td-color-warning); }
      &.P2 { background: var(--td-color-primary); }
      &.P3 { background: var(--td-text-placeholder); }
    }

    .title-text {
      font-weight: 500;
      color: var(--td-text-primary);
    }

    .due-tag {
      flex-shrink: 0;
      margin-left: 4px;
    }
  }

  .type-text { font-size: 13px; color: var(--td-text-secondary); }

  .status-badge {
    display: inline-flex;
    align-items: center;
    gap: 5px;
    padding: 3px 10px;
    border-radius: 12px;
    font-size: 12px;

    .status-dot {
      width: 6px;
      height: 6px;
      border-radius: 50%;
    }

    &.open { background: var(--td-bg-section); color: var(--td-text-secondary); .status-dot { background: var(--td-text-placeholder); } }
    &.in_progress { background: var(--td-tag-orange-bg); color: var(--td-tag-orange-text); .status-dot { background: var(--td-color-warning); } }
    &.pending_review { background: var(--td-tag-warning-bg); color: var(--td-tag-orange-text); .status-dot { background: var(--td-color-warning); } }
    &.resolved { background: var(--td-tag-success-bg); color: var(--td-color-success); .status-dot { background: var(--td-color-success); } }
    &.closed { background: var(--td-bg-section); color: var(--td-text-secondary); .status-dot { background: var(--td-text-placeholder); } }
    &.merged { background: var(--td-tag-purple-bg); color: var(--td-tag-purple-text); .status-dot { background: var(--td-tag-purple-text); } }
  }

  .assignee-cell {
    display: flex;
    align-items: center;
    gap: 8px;
    font-size: 13px;
    white-space: nowrap;
    overflow: hidden;

    span {
      overflow: hidden;
      text-overflow: ellipsis;
    }
  }

  .mini-avatar {
    width: 24px;
    height: 24px;
    border-radius: 6px;
    background: var(--td-tag-primary-bg);
    color: var(--td-color-primary);
    display: flex;
    align-items: center;
    justify-content: center;
    font-size: 11px;
    font-weight: 600;
    flex-shrink: 0;
  }

  .text-muted { color: var(--td-text-disabled); font-size: 13px; }

  .time-cell {
    display: flex;
    align-items: center;
    gap: 6px;
    font-size: 13px;
    color: var(--td-text-secondary);
    white-space: nowrap;

    .el-icon { font-size: 14px; color: var(--td-text-placeholder); }
  }

  .pagination-wrapper {
    padding-top: 20px;
    display: flex;
    justify-content: flex-end;
    border-top: 1px solid var(--td-divider-color);
    margin-top: 16px;
  }
}

// 看板视图
.kanban-view {
  min-height: calc(100vh - 400px);

  .kanban-board {
    display: flex;
    gap: 20px;
    overflow-x: auto;
    padding-bottom: 16px;
    height: calc(100vh - 400px);
  }

  .kanban-column {
    flex: 1;
    min-width: 320px;
    background-color: var(--td-bg-page);
    border-radius: 12px;
    display: flex;
    flex-direction: column;
    height: 100%;
  }

  .column-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    padding: 14px 16px;
    border-radius: 12px 12px 0 0;

    .column-title-group {
      display: flex;
      align-items: center;
      gap: 8px;
    }

    .column-dot {
      width: 10px;
      height: 10px;
      border-radius: 50%;

      &.open { background: var(--td-text-placeholder); }
      &.in_progress { background: var(--td-color-warning); }
      &.pending_review { background: var(--td-color-warning); }
      &.resolved { background: var(--td-color-success); }
    }

    .column-title {
      font-weight: 600;
      font-size: 14px;
      color: var(--td-text-regular);
    }

    .column-count {
      background-color: var(--td-border-color);
      padding: 2px 10px;
      border-radius: 10px;
      font-size: 12px;
      font-weight: 600;
      color: var(--td-text-secondary);
    }
  }

  .column-content {
    flex: 1;
    overflow-y: auto;
    padding: 12px;
  }

  .kanban-card {
    background-color: var(--td-bg-card);
    border-radius: 10px;
    padding: 14px;
    margin-bottom: 10px;
    box-shadow: 0 1px 4px rgba(0, 0, 0, 0.06);
    cursor: pointer;
    border: 1px solid var(--td-divider-color);
    transition: box-shadow 150ms ease-out, border-color 150ms ease-out;

    &:hover {
      box-shadow: 0 2px 8px rgba(0, 0, 0, 0.08);
      border-color: var(--td-color-primary);
    }

    &:active {
      background-color: var(--td-bg-page);
    }

    .card-header {
      display: flex;
      justify-content: space-between;
      align-items: center;
      margin-bottom: 8px;

      .issue-key {
        color: var(--td-color-primary);
        font-size: 12px;
        font-weight: 600;
      }
    }

    .card-title {
      font-size: 14px;
      font-weight: 500;
      color: var(--td-text-primary);
      margin-bottom: 12px;
      display: -webkit-box;
      -webkit-line-clamp: 2;
      -webkit-box-orient: vertical;
      overflow: hidden;
      line-height: 1.5;
    }

    .card-footer {
      display: flex;
      justify-content: space-between;
      align-items: center;

      .card-assignee {
        display: flex;
        align-items: center;
        gap: 6px;
        font-size: 12px;
        color: var(--td-text-secondary);
      }

      .mini-avatar.sm {
        width: 22px;
        height: 22px;
        border-radius: 6px;
        font-size: 10px;
      }
    }
  }

  .empty-column {
    text-align: center;
    padding: 24px;
    color: var(--td-text-placeholder);
    font-size: 14px;
  }
}

// 日期范围选择器
.date-range-picker {
  width: 240px;
}

// KPI 统计卡片
.stats-row {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: 16px;
  margin-bottom: 20px;
}

.stat-card {
  padding: 16px 20px;
  border-radius: 10px;
  background: var(--td-bg-card);
  border: 1px solid var(--td-border-color);

  .stat-value {
    font-size: 24px;
    font-weight: 700;
    line-height: 1.2;
    margin-bottom: 4px;
  }

  .stat-label {
    font-size: 13px;
    color: var(--td-text-secondary);
  }

  &.stat-total {
    border-left: 3px solid #3b82f6;
    .stat-value { color: #3b82f6; }
  }

  &.stat-resolved {
    border-left: 3px solid #10b981;
    .stat-value { color: #10b981; }
  }

  &.stat-rate {
    border-left: 3px solid #f59e0b;
    .stat-value { color: #f59e0b; }
  }

  &.stat-hours {
    border-left: 3px solid #6b7280;
    .stat-value { color: #6b7280; }
  }
}

// 响应式
@media (max-width: 768px) {
  .page-header {
    flex-direction: column;
    align-items: flex-start;
    gap: 16px;
  }

  .filter-card .filter-content {
    flex-direction: column;
    align-items: stretch;
  }

  .filter-card .filter-left {
    flex-direction: column;
  }

  .filter-card .quick-filter-bar {
    flex-direction: column;
    align-items: stretch;
  }

  .filter-card .quick-filter-right {
    justify-content: flex-start;
  }

  .filter-card .saved-view-select {
    width: 100%;
  }

  .search-input, .filter-select, .filter-select-sm {
    width: 100% !important;
  }

  .date-range-picker {
    width: 100% !important;
  }

  .stats-row {
    grid-template-columns: repeat(2, 1fr);
  }
}
</style>
