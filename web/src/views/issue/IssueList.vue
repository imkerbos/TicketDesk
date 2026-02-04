<template>
  <div class="issue-list-container">
    <!-- 页面头部 -->
    <div class="page-header">
      <div class="header-info">
        <div class="header-icon">
          <el-icon><Tickets /></el-icon>
        </div>
        <div class="header-text">
          <h1 class="header-title">工单管理</h1>
          <p class="header-desc">跟踪和管理所有工单任务</p>
        </div>
      </div>
      <el-button type="primary" @click="handleCreate" class="header-btn">
        <el-icon><Plus /></el-icon>
        创建工单
      </el-button>
    </div>

    <!-- 过滤器 -->
    <el-card shadow="never" class="filter-card">
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
          <el-select v-model="queryParams.project_key" placeholder="项目" clearable class="filter-select" @change="handleQuery">
            <el-option v-for="p in projects" :key="p.project_key" :label="p.name" :value="p.project_key" />
          </el-select>
          <el-select v-model="queryParams.status" placeholder="状态" clearable class="filter-select" @change="handleQuery">
            <el-option label="待处理" value="open" />
            <el-option label="进行中" value="in_progress" />
            <el-option label="已解决" value="resolved" />
            <el-option label="已关闭" value="closed" />
          </el-select>
          <el-select v-model="queryParams.priority" placeholder="优先级" clearable class="filter-select-sm" @change="handleQuery">
            <el-option label="P0 - 紧急" value="P0" />
            <el-option label="P1 - 高" value="P1" />
            <el-option label="P2 - 中" value="P2" />
            <el-option label="P3 - 低" value="P3" />
          </el-select>
          <el-select v-model="queryParams.assignee_id" placeholder="指派人" clearable filterable class="filter-select" @change="handleQuery">
            <el-option v-for="u in users" :key="u.id" :label="u.display_name" :value="u.id" />
          </el-select>
        </div>
        <div class="filter-right">
          <el-button :icon="Refresh" @click="handleReset">重置</el-button>
        </div>
      </div>
    </el-card>

    <!-- 工具栏 + 内容 -->
    <el-card shadow="never" class="table-card">
      <div class="toolbar">
        <div class="toolbar-left">
          <el-radio-group v-model="viewMode" @change="handleViewModeChange" class="view-toggle">
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
          @row-click="handleRowClick"
        >
          <el-table-column prop="issue_key" label="工单号" width="120" fixed>
            <template #default="{ row }">
              <el-link type="primary" underline="never" @click.stop="$router.push(`/issues/${row.issue_key}`)">
                <span class="issue-key-text">{{ row.issue_key }}</span>
              </el-link>
            </template>
          </el-table-column>
          <el-table-column prop="title" label="标题" min-width="280">
            <template #default="{ row }">
              <div class="issue-title-cell">
                <div class="priority-dot" :class="row.priority"></div>
                <span class="title-text">{{ row.title }}</span>
              </div>
            </template>
          </el-table-column>
          <el-table-column prop="project_key" label="项目" width="110">
            <template #default="{ row }">
              <el-tag size="small" effect="plain" type="info">{{ row.project_key }}</el-tag>
            </template>
          </el-table-column>
          <el-table-column label="类型" width="90">
            <template #default="{ row }">
              <span class="type-text">{{ row.issue_type?.display_name || '-' }}</span>
            </template>
          </el-table-column>
          <el-table-column prop="priority" label="优先级" width="80" align="center">
            <template #default="{ row }">
              <el-tag :type="getPriorityType(row.priority)" size="small" effect="dark" class="priority-tag">
                {{ row.priority }}
              </el-tag>
            </template>
          </el-table-column>
          <el-table-column prop="status" label="状态" width="100" align="center">
            <template #default="{ row }">
              <div class="status-badge" :class="row.status">
                <span class="status-dot"></span>
                <span>{{ getStatusText(row.status) }}</span>
              </div>
            </template>
          </el-table-column>
          <el-table-column label="指派人" width="120">
            <template #default="{ row }">
              <div v-if="row.assignee" class="assignee-cell">
                <div class="mini-avatar">{{ row.assignee.display_name?.charAt(0) }}</div>
                <span>{{ row.assignee.display_name }}</span>
              </div>
              <span v-else class="text-muted">未指派</span>
            </template>
          </el-table-column>
          <el-table-column prop="created_at" label="创建时间" width="160">
            <template #default="{ row }">
              <div class="time-cell">
                <el-icon><Clock /></el-icon>
                <span>{{ formatTime(row.created_at) }}</span>
              </div>
            </template>
          </el-table-column>
        </el-table>

        <div class="pagination-wrapper">
          <el-pagination
            v-model:current-page="queryParams.page"
            v-model:page-size="queryParams.page_size"
            :total="total"
            :page-sizes="[10, 20, 50, 100]"
            layout="total, sizes, prev, pager, next, jumper"
            @size-change="handleQuery"
            @current-change="handlePageChange"
          />
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
                @click="$router.push(`/issues/${issue.issue_key}`)"
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
    <el-dialog v-model="createDialogVisible" title="创建工单" width="600px" destroy-on-close class="create-dialog">
      <el-form ref="createFormRef" :model="createForm" :rules="createRules" label-position="top">
        <el-row :gutter="16">
          <el-col :span="12">
            <el-form-item label="项目" prop="project_key">
              <el-select v-model="createForm.project_key" placeholder="请选择项目" style="width: 100%" @change="handleProjectChange">
                <el-option v-for="p in projects" :key="p.project_key" :label="p.name" :value="p.project_key" />
              </el-select>
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="类型" prop="issue_type_id">
              <el-select v-model="createForm.issue_type_id" placeholder="请选择类型" style="width: 100%" :disabled="!createForm.project_key">
                <el-option v-for="t in issueTypes" :key="t.id" :label="t.display_name" :value="t.id" />
              </el-select>
            </el-form-item>
          </el-col>
        </el-row>
        <el-form-item label="标题" prop="title">
          <el-input v-model="createForm.title" placeholder="请输入工单标题" maxlength="200" show-word-limit />
        </el-form-item>
        <el-row :gutter="16">
          <el-col :span="12">
            <el-form-item label="优先级" prop="priority">
              <el-select v-model="createForm.priority" placeholder="请选择优先级" style="width: 100%">
                <el-option label="P0 - 紧急" value="P0" />
                <el-option label="P1 - 高" value="P1" />
                <el-option label="P2 - 中" value="P2" />
                <el-option label="P3 - 低" value="P3" />
              </el-select>
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="指派人">
              <el-select v-model="createForm.assignee_id" placeholder="请选择指派人" style="width: 100%" clearable filterable>
                <el-option v-for="u in users" :key="u.id" :label="u.display_name" :value="u.id" />
              </el-select>
            </el-form-item>
          </el-col>
        </el-row>
        <el-form-item label="描述">
          <el-input v-model="createForm.description" type="textarea" :rows="4" placeholder="请输入工单描述" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="createDialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="createLoading" @click="submitCreate">
          <el-icon><Check /></el-icon>
          创建工单
        </el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, computed, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage, type FormInstance, type FormRules } from 'element-plus'
import { Search, Refresh, Plus, List, Grid, Tickets, Clock, Check } from '@element-plus/icons-vue'
import { getIssueList, createIssue } from '@/api/issue'
import { getAllProjects, getProjectIssueTypes } from '@/api/project'
import { getAllUsers } from '@/api/user'
import type { Issue, IssueStatus, IssuePriority, CreateIssueRequest, KanbanColumn } from '@/types/issue'
import type { Project, ProjectIssueType } from '@/types/project'
import type { UserOption } from '@/types/user'
import dayjs from 'dayjs'

const router = useRouter()

const loading = ref(false)
const issueList = ref<Issue[]>([])
const total = ref(0)
const viewMode = ref<'table' | 'kanban'>('table')
const projects = ref<Project[]>([])
const users = ref<UserOption[]>([])
const issueTypes = ref<ProjectIssueType[]>([])

const queryParams = reactive({
  page: 1,
  page_size: 20,
  project_key: undefined as string | undefined,
  status: undefined as IssueStatus | undefined,
  priority: undefined as IssuePriority | undefined,
  assignee_id: undefined as number | undefined,
  keyword: undefined as string | undefined,
})

const kanbanColumns = computed<KanbanColumn[]>(() => {
  const columns: KanbanColumn[] = [
    { status: 'open', label: '待处理', issues: [] },
    { status: 'in_progress', label: '进行中', issues: [] },
    { status: 'resolved', label: '已解决', issues: [] },
  ]
  issueList.value.forEach((issue) => {
    const column = columns.find((c) => c.status === issue.status)
    if (column) column.issues.push(issue)
  })
  return columns
})

const createDialogVisible = ref(false)
const createLoading = ref(false)
const createFormRef = ref<FormInstance>()

const createForm = reactive<CreateIssueRequest & { issue_type_id?: number }>({
  project_key: '',
  issue_type_id: undefined,
  title: '',
  description: '',
  priority: 'P2',
  assignee_id: undefined,
})

const createRules: FormRules = {
  project_key: [{ required: true, message: '请选择项目', trigger: 'change' }],
  issue_type_id: [{ required: true, message: '请选择类型', trigger: 'change' }],
  title: [{ required: true, message: '请输入标题', trigger: 'blur' }],
  priority: [{ required: true, message: '请选择优先级', trigger: 'change' }],
}

const loadData = async () => {
  loading.value = true
  try {
    const params = { ...queryParams }
    if (viewMode.value === 'kanban') params.page_size = 100
    const { data } = await getIssueList(params)
    issueList.value = data.data.items
    total.value = data.data.total
  } catch (error) {
    console.error('Failed to load issues:', error)
  } finally {
    loading.value = false
  }
}

const loadFilterOptions = async () => {
  try {
    const [projectsRes, usersRes] = await Promise.all([getAllProjects(), getAllUsers()])
    projects.value = projectsRes.data.data
    users.value = usersRes.data.data
  } catch (error) {
    console.error('Failed to load filter options:', error)
  }
}

const handleQuery = () => { queryParams.page = 1; loadData() }
const handlePageChange = () => { loadData() }

const handleReset = () => {
  queryParams.page = 1
  queryParams.page_size = 20
  queryParams.project_key = undefined
  queryParams.status = undefined
  queryParams.priority = undefined
  queryParams.assignee_id = undefined
  queryParams.keyword = undefined
  loadData()
}

const handleViewModeChange = () => { loadData() }
const handleRowClick = (row: Issue) => { router.push(`/issues/${row.issue_key}`) }

const handleCreate = () => {
  Object.assign(createForm, {
    project_key: queryParams.project_key || '',
    issue_type_id: undefined, title: '', description: '', priority: 'P2', assignee_id: undefined,
  })
  if (createForm.project_key) handleProjectChange(createForm.project_key)
  createDialogVisible.value = true
}

const handleProjectChange = async (projectKey: string) => {
  createForm.issue_type_id = undefined
  if (!projectKey) { issueTypes.value = []; return }
  try {
    const { data } = await getProjectIssueTypes(projectKey)
    issueTypes.value = data.data
  } catch (error) {
    console.error('Failed to load issue types:', error)
  }
}

const submitCreate = async () => {
  if (!createFormRef.value) return
  await createFormRef.value.validate(async (valid) => {
    if (!valid) return
    createLoading.value = true
    try {
      // Ensure issue_type_id is set before submitting
      if (!createForm.issue_type_id) {
        ElMessage.error('请选择工单类型')
        createLoading.value = false
        return
      }
      const requestData: CreateIssueRequest = {
        ...createForm,
        issue_type_id: createForm.issue_type_id,
      }
      const { data } = await createIssue(requestData)
      ElMessage.success('创建成功')
      createDialogVisible.value = false
      router.push(`/issues/${data.data.issue_key}`)
    } catch (error) {
      console.error('Failed to create issue:', error)
    } finally {
      createLoading.value = false
    }
  })
}

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
  const map: Record<string, string> = { open: '待处理', in_progress: '进行中', resolved: '已解决', closed: '已关闭' }
  return map[status] || status
}
const formatTime = (time: string) => dayjs(time).format('YYYY-MM-DD HH:mm')

onMounted(() => { loadFilterOptions(); loadData() })
</script>

<style scoped lang="scss">
.issue-list-container {
  width: 100%;
}

// 页面头部
.page-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 24px;
  padding: 24px 32px;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  border-radius: 12px;
  color: #fff;

  .header-info {
    display: flex;
    align-items: center;
    gap: 16px;
  }

  .header-icon {
    width: 56px;
    height: 56px;
    background: rgba(255, 255, 255, 0.2);
    border-radius: 14px;
    display: flex;
    align-items: center;
    justify-content: center;
    font-size: 28px;
  }

  .header-text {
    .header-title { font-size: 22px; font-weight: 600; margin: 0 0 4px 0; }
    .header-desc { font-size: 14px; margin: 0; opacity: 0.9; }
  }

  .header-btn {
    background: rgba(255, 255, 255, 0.2);
    border: 1px solid rgba(255, 255, 255, 0.3);
    color: #fff;
    &:hover { background: rgba(255, 255, 255, 0.3); }
  }
}

// 筛选卡片
.filter-card {
  margin-bottom: 20px;
  border-radius: 12px;

  :deep(.el-card__body) { padding: 16px 20px; }

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
    color: #9ca3af;
  }
}

// 表格视图
.table-view {
  :deep(.el-table) {
    border-radius: 8px;

    th.el-table__cell {
      background: #f8fafc;
      font-weight: 600;
      color: #374151;
    }

    .clickable-row {
      cursor: pointer;
      &:hover { background-color: #f9fafb; }
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

      &.P0 { background: #ef4444; }
      &.P1 { background: #f59e0b; }
      &.P2 { background: #3b82f6; }
      &.P3 { background: #9ca3af; }
    }

    .title-text {
      font-weight: 500;
      color: #1f2937;
    }
  }

  .type-text { font-size: 13px; color: #6b7280; }

  .status-badge {
    display: inline-flex;
    align-items: center;
    gap: 5px;
    padding: 3px 10px;
    border-radius: 20px;
    font-size: 12px;

    .status-dot {
      width: 6px;
      height: 6px;
      border-radius: 50%;
    }

    &.open { background: #f3f4f6; color: #6b7280; .status-dot { background: #9ca3af; } }
    &.in_progress { background: #fff7ed; color: #c2410c; .status-dot { background: #f59e0b; } }
    &.resolved { background: #ecfdf5; color: #059669; .status-dot { background: #10b981; } }
    &.closed { background: #f3f4f6; color: #6b7280; .status-dot { background: #9ca3af; } }
  }

  .assignee-cell {
    display: flex;
    align-items: center;
    gap: 8px;
    font-size: 13px;
  }

  .mini-avatar {
    width: 24px;
    height: 24px;
    border-radius: 6px;
    background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
    color: #fff;
    display: flex;
    align-items: center;
    justify-content: center;
    font-size: 11px;
    font-weight: 600;
    flex-shrink: 0;
  }

  .text-muted { color: #d1d5db; font-size: 13px; }

  .time-cell {
    display: flex;
    align-items: center;
    gap: 6px;
    font-size: 13px;
    color: #6b7280;

    .el-icon { font-size: 14px; color: #9ca3af; }
  }

  .pagination-wrapper {
    padding-top: 20px;
    display: flex;
    justify-content: flex-end;
    border-top: 1px solid #f0f0f0;
    margin-top: 16px;
  }
}

// 看板视图
.kanban-view {
  .kanban-board {
    display: flex;
    gap: 16px;
    overflow-x: auto;
    padding-bottom: 16px;
  }

  .kanban-column {
    flex: 1;
    min-width: 300px;
    max-width: 380px;
    background-color: #f8fafc;
    border-radius: 12px;
    display: flex;
    flex-direction: column;
    max-height: calc(100vh - 340px);
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

      &.open { background: #9ca3af; }
      &.in_progress { background: #f59e0b; }
      &.resolved { background: #10b981; }
    }

    .column-title {
      font-weight: 600;
      font-size: 14px;
      color: #374151;
    }

    .column-count {
      background-color: #e5e7eb;
      padding: 2px 10px;
      border-radius: 10px;
      font-size: 12px;
      font-weight: 600;
      color: #6b7280;
    }
  }

  .column-content {
    flex: 1;
    overflow-y: auto;
    padding: 12px;
  }

  .kanban-card {
    background-color: #fff;
    border-radius: 10px;
    padding: 14px;
    margin-bottom: 10px;
    box-shadow: 0 1px 4px rgba(0, 0, 0, 0.06);
    cursor: pointer;
    border: 1px solid #f0f0f0;
    transition: box-shadow 0.2s, transform 0.2s;

    &:hover {
      box-shadow: 0 4px 12px rgba(0, 0, 0, 0.1);
      transform: translateY(-2px);
    }

    .card-header {
      display: flex;
      justify-content: space-between;
      align-items: center;
      margin-bottom: 8px;

      .issue-key {
        color: #667eea;
        font-size: 12px;
        font-weight: 600;
      }
    }

    .card-title {
      font-size: 14px;
      font-weight: 500;
      color: #1f2937;
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
        color: #6b7280;
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
    color: #9ca3af;
    font-size: 14px;
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

  .search-input, .filter-select, .filter-select-sm {
    width: 100% !important;
  }
}
</style>
