<template>
  <div class="issue-list-container">
    <!-- 过滤器 -->
    <el-card shadow="never" class="filter-card">
      <el-form :inline="true" :model="queryParams">
        <el-form-item label="项目">
          <el-select v-model="queryParams.project_key" placeholder="全部项目" clearable style="width: 150px" @change="handleQuery">
            <el-option v-for="p in projects" :key="p.project_key" :label="p.name" :value="p.project_key" />
          </el-select>
        </el-form-item>
        <el-form-item label="状态">
          <el-select v-model="queryParams.status" placeholder="全部" clearable style="width: 120px" @change="handleQuery">
            <el-option label="待处理" value="open" />
            <el-option label="进行中" value="in_progress" />
            <el-option label="已解决" value="resolved" />
            <el-option label="已关闭" value="closed" />
          </el-select>
        </el-form-item>
        <el-form-item label="优先级">
          <el-select v-model="queryParams.priority" placeholder="全部" clearable style="width: 120px" @change="handleQuery">
            <el-option label="P0 - 紧急" value="P0" />
            <el-option label="P1 - 高" value="P1" />
            <el-option label="P2 - 中" value="P2" />
            <el-option label="P3 - 低" value="P3" />
          </el-select>
        </el-form-item>
        <el-form-item label="指派人">
          <el-select v-model="queryParams.assignee_id" placeholder="全部" clearable filterable style="width: 150px" @change="handleQuery">
            <el-option v-for="u in users" :key="u.id" :label="u.display_name" :value="u.id" />
          </el-select>
        </el-form-item>
        <el-form-item>
          <el-input v-model="queryParams.keyword" placeholder="搜索标题" clearable style="width: 200px" @clear="handleQuery" @keyup.enter="handleQuery">
            <template #prefix>
              <el-icon><Search /></el-icon>
            </template>
          </el-input>
        </el-form-item>
        <el-form-item>
          <el-button type="primary" :icon="Search" @click="handleQuery">查询</el-button>
          <el-button :icon="Refresh" @click="handleReset">重置</el-button>
        </el-form-item>
      </el-form>
    </el-card>

    <!-- 工具栏 -->
    <el-card shadow="never">
      <div class="toolbar">
        <div class="toolbar-left">
          <el-radio-group v-model="viewMode" @change="handleViewModeChange">
            <el-radio-button value="table">
              <el-icon><List /></el-icon>
              表格视图
            </el-radio-button>
            <el-radio-button value="kanban">
              <el-icon><Grid /></el-icon>
              看板视图
            </el-radio-button>
          </el-radio-group>
        </div>
        <div class="toolbar-right">
          <el-button type="primary" :icon="Plus" @click="handleCreate">创建工单</el-button>
        </div>
      </div>

      <!-- 表格视图 -->
      <div v-if="viewMode === 'table'" class="table-view">
        <el-table
          v-loading="loading"
          :data="issueList"
          stripe
          style="width: 100%"
          @row-click="handleRowClick"
        >
          <el-table-column prop="issue_key" label="工单号" width="120" fixed>
            <template #default="{ row }">
              <el-link type="primary" underline="never" @click.stop="$router.push(`/issues/${row.issue_key}`)">
                {{ row.issue_key }}
              </el-link>
            </template>
          </el-table-column>
          <el-table-column prop="title" label="标题" min-width="250">
            <template #default="{ row }">
              <div class="issue-title-cell">
                <span class="title">{{ row.title }}</span>
              </div>
            </template>
          </el-table-column>
          <el-table-column prop="project_key" label="项目" width="120">
            <template #default="{ row }">
              <el-tag size="small">{{ row.project_key }}</el-tag>
            </template>
          </el-table-column>
          <el-table-column label="类型" width="100">
            <template #default="{ row }">
              {{ row.issue_type?.display_name || '-' }}
            </template>
          </el-table-column>
          <el-table-column prop="priority" label="优先级" width="90">
            <template #default="{ row }">
              <el-tag :type="getPriorityType(row.priority)" size="small">
                {{ row.priority }}
              </el-tag>
            </template>
          </el-table-column>
          <el-table-column prop="status" label="状态" width="100">
            <template #default="{ row }">
              <el-tag :type="getStatusType(row.status)" size="small">
                {{ getStatusText(row.status) }}
              </el-tag>
            </template>
          </el-table-column>
          <el-table-column label="指派人" width="100">
            <template #default="{ row }">
              <span v-if="row.assignee">{{ row.assignee.display_name }}</span>
              <span v-else class="text-muted">未指派</span>
            </template>
          </el-table-column>
          <el-table-column prop="created_at" label="创建时间" width="160">
            <template #default="{ row }">
              {{ formatTime(row.created_at) }}
            </template>
          </el-table-column>
        </el-table>

        <el-pagination
          v-model:current-page="queryParams.page"
          v-model:page-size="queryParams.page_size"
          :total="total"
          :page-sizes="[10, 20, 50, 100]"
          layout="total, sizes, prev, pager, next, jumper"
          class="pagination"
          @size-change="handleQuery"
          @current-change="handlePageChange"
        />
      </div>

      <!-- 看板视图 -->
      <div v-if="viewMode === 'kanban'" v-loading="loading" class="kanban-view">
        <div class="kanban-board">
          <div v-for="column in kanbanColumns" :key="column.status" class="kanban-column">
            <div class="column-header" :class="column.status">
              <span class="column-title">{{ column.label }}</span>
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
                  <el-tag :type="getPriorityType(issue.priority)" size="small">
                    {{ issue.priority }}
                  </el-tag>
                </div>
                <div class="card-title">{{ issue.title }}</div>
                <div class="card-footer">
                  <el-tag size="small">{{ issue.project_key }}</el-tag>
                  <el-avatar v-if="issue.assignee" :size="24" class="assignee-avatar">
                    {{ issue.assignee.display_name?.charAt(0) || '?' }}
                  </el-avatar>
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
    <el-dialog v-model="createDialogVisible" title="创建工单" width="600px" destroy-on-close>
      <el-form ref="createFormRef" :model="createForm" :rules="createRules" label-width="80px">
        <el-form-item label="项目" prop="project_key">
          <el-select v-model="createForm.project_key" placeholder="请选择项目" style="width: 100%" @change="handleProjectChange">
            <el-option v-for="p in projects" :key="p.project_key" :label="p.name" :value="p.project_key" />
          </el-select>
        </el-form-item>
        <el-form-item label="类型" prop="issue_type_id">
          <el-select v-model="createForm.issue_type_id" placeholder="请选择类型" style="width: 100%" :disabled="!createForm.project_key">
            <el-option v-for="t in issueTypes" :key="t.id" :label="t.display_name" :value="t.id" />
          </el-select>
        </el-form-item>
        <el-form-item label="标题" prop="title">
          <el-input v-model="createForm.title" placeholder="请输入工单标题" maxlength="200" show-word-limit />
        </el-form-item>
        <el-form-item label="优先级" prop="priority">
          <el-select v-model="createForm.priority" placeholder="请选择优先级" style="width: 100%">
            <el-option label="P0 - 紧急" value="P0" />
            <el-option label="P1 - 高" value="P1" />
            <el-option label="P2 - 中" value="P2" />
            <el-option label="P3 - 低" value="P3" />
          </el-select>
        </el-form-item>
        <el-form-item label="指派人">
          <el-select v-model="createForm.assignee_id" placeholder="请选择指派人" style="width: 100%" clearable filterable>
            <el-option v-for="u in users" :key="u.id" :label="u.display_name" :value="u.id" />
          </el-select>
        </el-form-item>
        <el-form-item label="描述">
          <el-input v-model="createForm.description" type="textarea" :rows="4" placeholder="请输入工单描述" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="createDialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="createLoading" @click="submitCreate">创建</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, computed, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage, type FormInstance, type FormRules } from 'element-plus'
import { Search, Refresh, Plus, List, Grid } from '@element-plus/icons-vue'
import { getIssueList, createIssue } from '@/api/issue'
import { getAllProjects, getProjectIssueTypes } from '@/api/project'
import { getAllUsers } from '@/api/user'
import type { Issue, IssueStatus, IssuePriority, CreateIssueRequest, KanbanColumn } from '@/types/issue'
import type { Project, ProjectIssueType } from '@/types/project'
import type { UserOption } from '@/types/user'
import dayjs from 'dayjs'

const router = useRouter()

// 数据
const loading = ref(false)
const issueList = ref<Issue[]>([])
const total = ref(0)
const viewMode = ref<'table' | 'kanban'>('table')
const projects = ref<Project[]>([])
const users = ref<UserOption[]>([])
const issueTypes = ref<ProjectIssueType[]>([])

// 查询参数
const queryParams = reactive({
  page: 1,
  page_size: 20,
  project_key: undefined as string | undefined,
  status: undefined as IssueStatus | undefined,
  priority: undefined as IssuePriority | undefined,
  assignee_id: undefined as number | undefined,
  keyword: undefined as string | undefined,
})

// 看板列配置
const kanbanColumns = computed<KanbanColumn[]>(() => {
  const columns: KanbanColumn[] = [
    { status: 'open', label: '待处理', issues: [] },
    { status: 'in_progress', label: '进行中', issues: [] },
    { status: 'resolved', label: '已解决', issues: [] },
  ]

  issueList.value.forEach((issue) => {
    const column = columns.find((c) => c.status === issue.status)
    if (column) {
      column.issues.push(issue)
    }
  })

  return columns
})

// 创建工单相关
const createDialogVisible = ref(false)
const createLoading = ref(false)
const createFormRef = ref<FormInstance>()

const createForm = reactive<CreateIssueRequest>({
  project_key: '',
  issue_type_id: 0,
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

// 加载数据
const loadData = async () => {
  loading.value = true
  try {
    const params = { ...queryParams }
    // 看板模式下加载更多数据
    if (viewMode.value === 'kanban') {
      params.page_size = 100
    }
    const { data } = await getIssueList(params)
    issueList.value = data.data.items
    total.value = data.data.total
  } catch (error) {
    console.error('Failed to load issues:', error)
  } finally {
    loading.value = false
  }
}

// 加载筛选选项
const loadFilterOptions = async () => {
  try {
    const [projectsRes, usersRes] = await Promise.all([
      getAllProjects(),
      getAllUsers(),
    ])
    projects.value = projectsRes.data.data
    users.value = usersRes.data.data
  } catch (error) {
    console.error('Failed to load filter options:', error)
  }
}

// 查询
const handleQuery = () => {
  queryParams.page = 1
  loadData()
}

// 分页变化
const handlePageChange = () => {
  loadData()
}

// 重置
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

// 视图模式切换
const handleViewModeChange = () => {
  loadData()
}

// 行点击
const handleRowClick = (row: Issue) => {
  router.push(`/issues/${row.issue_key}`)
}

// 打开创建对话框
const handleCreate = () => {
  Object.assign(createForm, {
    project_key: queryParams.project_key || '',
    issue_type_id: 0,
    title: '',
    description: '',
    priority: 'P2',
    assignee_id: undefined,
  })

  if (createForm.project_key) {
    handleProjectChange(createForm.project_key)
  }

  createDialogVisible.value = true
}

// 项目变更
const handleProjectChange = async (projectKey: string) => {
  createForm.issue_type_id = 0
  if (!projectKey) {
    issueTypes.value = []
    return
  }

  try {
    const { data } = await getProjectIssueTypes(projectKey)
    issueTypes.value = data.data
  } catch (error) {
    console.error('Failed to load issue types:', error)
  }
}

// 提交创建
const submitCreate = async () => {
  if (!createFormRef.value) return

  await createFormRef.value.validate(async (valid) => {
    if (!valid) return

    createLoading.value = true
    try {
      const { data } = await createIssue(createForm)
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

// 工具函数
type TagType = 'primary' | 'success' | 'warning' | 'info' | 'danger'

const getPriorityType = (priority: string): TagType => {
  const map: Record<string, TagType> = {
    P0: 'danger',
    P1: 'warning',
    P2: 'info',
    P3: 'info',
  }
  return map[priority] || 'info'
}

const getStatusType = (status: string): TagType => {
  const map: Record<string, TagType> = {
    open: 'info',
    in_progress: 'warning',
    resolved: 'success',
    closed: 'info',
  }
  return map[status] || 'info'
}

const getStatusText = (status: string) => {
  const map: Record<string, string> = {
    open: '待处理',
    in_progress: '进行中',
    resolved: '已解决',
    closed: '已关闭',
  }
  return map[status] || status
}

const formatTime = (time: string) => {
  return dayjs(time).format('YYYY-MM-DD HH:mm')
}

// 初始化
onMounted(() => {
  loadFilterOptions()
  loadData()
})
</script>

<style scoped lang="scss">
.issue-list-container {
  .filter-card {
    margin-bottom: 20px;
  }

  .toolbar {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-bottom: 20px;
  }

  .table-view {
    .issue-title-cell {
      .title {
        font-weight: 500;
      }
    }

    .text-muted {
      color: #909399;
    }

    .pagination {
      margin-top: 20px;
      justify-content: flex-end;
    }

    :deep(.el-table) {
      .el-table__row {
        cursor: pointer;

        &:hover {
          background-color: #f5f7fa;
        }
      }
    }
  }

  .kanban-view {
    .kanban-board {
      display: flex;
      gap: 16px;
      overflow-x: auto;
      padding-bottom: 16px;
    }

    .kanban-column {
      flex: 1;
      min-width: 280px;
      max-width: 350px;
      background-color: #f5f7fa;
      border-radius: 8px;
      display: flex;
      flex-direction: column;
      max-height: calc(100vh - 300px);
    }

    .column-header {
      display: flex;
      justify-content: space-between;
      align-items: center;
      padding: 12px 16px;
      border-bottom: 3px solid;
      border-radius: 8px 8px 0 0;

      &.open {
        border-color: #909399;
        background-color: rgba(144, 147, 153, 0.1);
      }

      &.in_progress {
        border-color: #e6a23c;
        background-color: rgba(230, 162, 60, 0.1);
      }

      &.resolved {
        border-color: #67c23a;
        background-color: rgba(103, 194, 58, 0.1);
      }

      .column-title {
        font-weight: 600;
        font-size: 14px;
      }

      .column-count {
        background-color: rgba(0, 0, 0, 0.1);
        padding: 2px 8px;
        border-radius: 10px;
        font-size: 12px;
      }
    }

    .column-content {
      flex: 1;
      overflow-y: auto;
      padding: 12px;
    }

    .kanban-card {
      background-color: #fff;
      border-radius: 8px;
      padding: 12px;
      margin-bottom: 8px;
      box-shadow: 0 1px 3px rgba(0, 0, 0, 0.1);
      cursor: pointer;
      transition: box-shadow 0.2s, transform 0.2s;

      &:hover {
        box-shadow: 0 4px 12px rgba(0, 0, 0, 0.15);
        transform: translateY(-2px);
      }

      .card-header {
        display: flex;
        justify-content: space-between;
        align-items: center;
        margin-bottom: 8px;

        .issue-key {
          color: #409eff;
          font-size: 12px;
          font-weight: 500;
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
      }

      .card-footer {
        display: flex;
        justify-content: space-between;
        align-items: center;

        .assignee-avatar {
          background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
          font-size: 12px;
        }
      }
    }

    .empty-column {
      text-align: center;
      padding: 20px;
      color: #909399;
      font-size: 14px;
    }
  }
}
</style>
