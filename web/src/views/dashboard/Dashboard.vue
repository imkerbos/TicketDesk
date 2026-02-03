<template>
  <div class="dashboard-container">
    <!-- 统计卡片 -->
    <el-row :gutter="20" class="stat-row">
      <el-col :xs="24" :sm="12" :md="6">
        <div class="stat-card todo">
          <div class="stat-icon">
            <el-icon :size="28"><Tickets /></el-icon>
          </div>
          <div class="stat-content">
            <div class="stat-value">{{ stats.myTodo }}</div>
            <div class="stat-label">我的待办</div>
          </div>
        </div>
      </el-col>
      <el-col :xs="24" :sm="12" :md="6">
        <div class="stat-card created">
          <div class="stat-icon">
            <el-icon :size="28"><Edit /></el-icon>
          </div>
          <div class="stat-content">
            <div class="stat-value">{{ stats.myCreated }}</div>
            <div class="stat-label">我创建的</div>
          </div>
        </div>
      </el-col>
      <el-col :xs="24" :sm="12" :md="6">
        <div class="stat-card alert">
          <div class="stat-icon">
            <el-icon :size="28"><Bell /></el-icon>
          </div>
          <div class="stat-content">
            <div class="stat-value">{{ stats.pendingAlerts }}</div>
            <div class="stat-label">待确认告警</div>
          </div>
        </div>
      </el-col>
      <el-col :xs="24" :sm="12" :md="6">
        <div class="stat-card done">
          <div class="stat-icon">
            <el-icon :size="28"><CircleCheck /></el-icon>
          </div>
          <div class="stat-content">
            <div class="stat-value">{{ stats.weekDone }}</div>
            <div class="stat-label">本周已完成</div>
          </div>
        </div>
      </el-col>
    </el-row>

    <!-- 快捷操作 -->
    <div class="quick-actions">
      <el-button type="primary" :icon="Plus" @click="handleCreateIssue">创建工单</el-button>
      <el-button :icon="Tickets" @click="$router.push('/issues')">查看所有工单</el-button>
      <el-button :icon="Bell" @click="$router.push('/alerts')">查看所有告警</el-button>
    </div>

    <!-- 主内容区 -->
    <el-row :gutter="20">
      <!-- 左侧：待办工单 -->
      <el-col :xs="24" :lg="14">
        <el-card shadow="never" class="content-card">
          <template #header>
            <div class="card-header">
              <span class="card-title">
                <el-icon><Tickets /></el-icon>
                我的待办工单
              </span>
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
                <div class="issue-main">
                  <span class="issue-key">{{ issue.issue_key }}</span>
                  <span class="issue-title">{{ issue.title }}</span>
                </div>
                <div class="issue-meta">
                  <el-tag :type="getPriorityType(issue.priority)" size="small">
                    {{ issue.priority }}
                  </el-tag>
                  <el-tag :type="getStatusType(issue.status)" size="small">
                    {{ getStatusText(issue.status) }}
                  </el-tag>
                  <span class="issue-project">{{ issue.project_key }}</span>
                </div>
              </div>
            </div>
          </div>
        </el-card>

        <!-- 我创建的工单 -->
        <el-card shadow="never" class="content-card">
          <template #header>
            <div class="card-header">
              <span class="card-title">
                <el-icon><Edit /></el-icon>
                我创建的工单
              </span>
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
                <div class="issue-main">
                  <span class="issue-key">{{ issue.issue_key }}</span>
                  <span class="issue-title">{{ issue.title }}</span>
                </div>
                <div class="issue-meta">
                  <el-tag :type="getPriorityType(issue.priority)" size="small">
                    {{ issue.priority }}
                  </el-tag>
                  <el-tag :type="getStatusType(issue.status)" size="small">
                    {{ getStatusText(issue.status) }}
                  </el-tag>
                  <span v-if="issue.assignee" class="issue-assignee">
                    → {{ issue.assignee.display_name }}
                  </span>
                  <span v-else class="issue-assignee unassigned">未指派</span>
                </div>
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
              <span class="card-title">
                <el-icon><Bell /></el-icon>
                待确认告警
              </span>
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
                <div class="alert-severity" :class="alert.severity"></div>
                <div class="alert-content">
                  <div class="alert-name">{{ alert.alert_name }}</div>
                  <div class="alert-time">{{ formatTime(alert.starts_at) }}</div>
                </div>
                <el-tag :type="getSeverityTagType(alert.severity)" size="small">
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
              <span class="card-title">
                <el-icon><Clock /></el-icon>
                最近活动
              </span>
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
                    v-if="activity.issue_key"
                    type="primary"
                    @click.stop="$router.push(`/issues/${activity.issue_key}`)"
                  >
                    {{ activity.issue_key }}
                  </el-link>
                </div>
              </el-timeline-item>
            </el-timeline>
          </div>
        </el-card>
      </el-col>
    </el-row>

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
import { ref, reactive, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage, type FormInstance, type FormRules } from 'element-plus'
import {
  Tickets,
  Edit,
  Bell,
  CircleCheck,
  Plus,
  ArrowRight,
  Clock,
} from '@element-plus/icons-vue'
import { getMyTodoIssues, getMyCreatedIssues, createIssue } from '@/api/issue'
import { getAlertList } from '@/api/alert'
import { getAllProjects, getProjectIssueTypes } from '@/api/project'
import { getAllUsers } from '@/api/user'
import type { Issue, CreateIssueRequest } from '@/types/issue'
import type { Alert } from '@/types/alert'
import type { Project, ProjectIssueType } from '@/types/project'
import type { UserOption } from '@/types/user'
import dayjs from 'dayjs'

const router = useRouter()

// 统计数据
const stats = reactive({
  myTodo: 0,
  myCreated: 0,
  pendingAlerts: 0,
  weekDone: 0,
})

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
const recentActivities = ref<Array<{
  id: number
  user_name: string
  action: string
  issue_key?: string
  created_at: string
}>>([])

// 创建工单相关
const createDialogVisible = ref(false)
const createLoading = ref(false)
const createFormRef = ref<FormInstance>()
const projects = ref<Project[]>([])
const issueTypes = ref<ProjectIssueType[]>([])
const users = ref<UserOption[]>([])

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

// 加载最近活动（模拟数据，实际需要后端支持）
const loadRecentActivities = async () => {
  loadingActivities.value = true
  try {
    // TODO: 替换为实际的 API 调用
    // const { data } = await getRecentActivities({ page_size: 10 })
    // recentActivities.value = data.data

    // 暂时使用模拟数据
    recentActivities.value = []
  } catch (error) {
    console.error('Failed to load activities:', error)
  } finally {
    loadingActivities.value = false
  }
}

// 打开创建工单对话框
const handleCreateIssue = async () => {
  // 重置表单
  Object.assign(createForm, {
    project_key: '',
    issue_type_id: 0,
    title: '',
    description: '',
    priority: 'P2',
    assignee_id: undefined,
  })

  // 加载项目和用户列表
  try {
    const [projectsRes, usersRes] = await Promise.all([
      getAllProjects(),
      getAllUsers(),
    ])
    projects.value = projectsRes.data.data
    users.value = usersRes.data.data
  } catch (error) {
    console.error('Failed to load form data:', error)
  }

  createDialogVisible.value = true
}

// 项目变更时加载工单类型
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

const getSeverityTagType = (severity: string): TagType => {
  const map: Record<string, TagType> = {
    critical: 'danger',
    warning: 'warning',
    info: 'info',
  }
  return map[severity] || 'info'
}

const getSeverityText = (severity: string) => {
  const map: Record<string, string> = {
    critical: '严重',
    warning: '警告',
    info: '信息',
  }
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
  .stat-row {
    margin-bottom: 24px;
  }

  .stat-card {
    display: flex;
    align-items: center;
    padding: 24px;
    border-radius: 12px;
    color: #fff;
    transition: transform 0.3s;

    &:hover {
      transform: translateY(-4px);
    }

    &.todo {
      background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
      box-shadow: 0 4px 20px rgba(102, 126, 234, 0.3);
    }

    &.created {
      background: linear-gradient(135deg, #11998e 0%, #38ef7d 100%);
      box-shadow: 0 4px 20px rgba(17, 153, 142, 0.3);
    }

    &.alert {
      background: linear-gradient(135deg, #f56c6c 0%, #c0392b 100%);
      box-shadow: 0 4px 20px rgba(245, 108, 108, 0.3);
    }

    &.done {
      background: linear-gradient(135deg, #409eff 0%, #2c7bd9 100%);
      box-shadow: 0 4px 20px rgba(64, 158, 255, 0.3);
    }

    .stat-icon {
      width: 56px;
      height: 56px;
      display: flex;
      align-items: center;
      justify-content: center;
      background: rgba(255, 255, 255, 0.2);
      border-radius: 12px;
      margin-right: 16px;
    }

    .stat-content {
      .stat-value {
        font-size: 32px;
        font-weight: 700;
        line-height: 1.2;
      }

      .stat-label {
        font-size: 14px;
        opacity: 0.9;
        margin-top: 4px;
      }
    }
  }

  .quick-actions {
    margin-bottom: 24px;
    display: flex;
    gap: 12px;
  }

  .content-card {
    margin-bottom: 20px;
    min-height: 280px;

    .card-header {
      display: flex;
      justify-content: space-between;
      align-items: center;

      .card-title {
        display: flex;
        align-items: center;
        gap: 8px;
        font-size: 16px;
        font-weight: 600;
      }
    }

    .empty-state {
      padding: 40px 0;
      display: flex;
      align-items: center;
      justify-content: center;
      min-height: 180px;
    }
  }

  .issue-list {
    .issue-item {
      padding: 12px 0;
      border-bottom: 1px solid #f0f0f0;
      cursor: pointer;
      transition: background-color 0.2s;

      &:hover {
        background-color: #fafafa;
        margin: 0 -20px;
        padding: 12px 20px;
      }

      &:last-child {
        border-bottom: none;
      }

      .issue-main {
        display: flex;
        align-items: center;
        gap: 12px;
        margin-bottom: 8px;

        .issue-key {
          color: #409eff;
          font-weight: 500;
          font-size: 13px;
        }

        .issue-title {
          color: #1f2937;
          font-weight: 500;
          overflow: hidden;
          text-overflow: ellipsis;
          white-space: nowrap;
        }
      }

      .issue-meta {
        display: flex;
        align-items: center;
        gap: 8px;
        font-size: 12px;

        .issue-project {
          color: #909399;
        }

        .issue-assignee {
          color: #67c23a;

          &.unassigned {
            color: #909399;
            font-style: italic;
          }
        }
      }
    }
  }

  .alert-list {
    .alert-item {
      display: flex;
      align-items: center;
      padding: 12px 0;
      border-bottom: 1px solid #f0f0f0;
      cursor: pointer;
      transition: background-color 0.2s;

      &:hover {
        background-color: #fafafa;
        margin: 0 -20px;
        padding: 12px 20px;
      }

      &:last-child {
        border-bottom: none;
      }

      .alert-severity {
        width: 4px;
        height: 40px;
        border-radius: 2px;
        margin-right: 12px;

        &.critical {
          background-color: #f56c6c;
        }

        &.warning {
          background-color: #e6a23c;
        }

        &.info {
          background-color: #909399;
        }
      }

      .alert-content {
        flex: 1;
        overflow: hidden;

        .alert-name {
          font-weight: 500;
          color: #1f2937;
          overflow: hidden;
          text-overflow: ellipsis;
          white-space: nowrap;
        }

        .alert-time {
          font-size: 12px;
          color: #909399;
          margin-top: 4px;
        }
      }
    }
  }

  .activity-timeline {
    padding-left: 0;

    :deep(.el-timeline-item__timestamp) {
      font-size: 12px;
    }

    .activity-content {
      font-size: 14px;

      .activity-user {
        font-weight: 500;
        color: #1f2937;
      }

      .activity-action {
        color: #606266;
        margin: 0 4px;
      }
    }
  }
}
</style>
