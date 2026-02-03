<template>
  <div v-loading="loading" class="issue-detail-container">
    <template v-if="issue">
      <!-- 头部信息 -->
      <div class="issue-header">
        <div class="header-main">
          <div class="issue-breadcrumb">
            <el-breadcrumb separator="/">
              <el-breadcrumb-item :to="{ path: '/issues' }">工单列表</el-breadcrumb-item>
              <el-breadcrumb-item :to="{ path: '/issues?project_key=' + issue.project_key }">{{ issue.project_key }}</el-breadcrumb-item>
              <el-breadcrumb-item>{{ issue.issue_key }}</el-breadcrumb-item>
            </el-breadcrumb>
          </div>
          <h1 class="issue-title">{{ issue.title }}</h1>
          <div class="issue-meta">
            <el-tag :type="getPriorityType(issue.priority)" size="small">{{ issue.priority }}</el-tag>
            <el-tag :type="getStatusType(issue.status)" size="small">{{ getStatusText(issue.status) }}</el-tag>
            <span class="meta-item">
              <el-icon><User /></el-icon>
              创建者: {{ issue.reporter?.display_name || '未知' }}
            </span>
            <span class="meta-item">
              <el-icon><Clock /></el-icon>
              {{ formatTime(issue.created_at) }}
            </span>
          </div>
        </div>
        <div class="header-actions">
          <el-dropdown v-if="issue && getAvailableTransitions(issue.status).length > 0" trigger="click" @command="handleTransition">
            <el-button type="primary">
              状态流转 <el-icon class="el-icon--right"><ArrowDown /></el-icon>
            </el-button>
            <template #dropdown>
              <el-dropdown-menu>
                <el-dropdown-item v-for="t in getAvailableTransitions(issue.status)" :key="t.status" :command="t.status">
                  {{ t.label }}
                </el-dropdown-item>
              </el-dropdown-menu>
            </template>
          </el-dropdown>
          <el-button :icon="Edit" @click="handleEdit">编辑</el-button>
        </div>
      </div>

      <el-row :gutter="20">
        <!-- 左侧：描述和评论 -->
        <el-col :xs="24" :lg="16">
          <!-- 描述 -->
          <el-card shadow="never" class="content-card">
            <template #header>
              <span class="card-title">描述</span>
            </template>
            <div v-if="issue.description" class="description-content">
              {{ issue.description }}
            </div>
            <div v-else class="empty-description">暂无描述</div>
          </el-card>

          <!-- 评论 -->
          <el-card shadow="never" class="content-card">
            <template #header>
              <div class="card-header-with-action">
                <span class="card-title">评论 ({{ comments.length }})</span>
              </div>
            </template>

            <!-- 添加评论 -->
            <div class="add-comment">
              <el-input
                v-model="newComment"
                type="textarea"
                :rows="3"
                placeholder="添加评论..."
              />
              <div class="comment-actions">
                <el-button type="primary" :loading="commentLoading" :disabled="!newComment.trim()" @click="submitComment">
                  发表评论
                </el-button>
              </div>
            </div>

            <!-- 评论列表 -->
            <div class="comment-list">
              <div v-for="comment in comments" :key="comment.id" class="comment-item">
                <el-avatar :size="36" class="comment-avatar">
                  {{ comment.user?.display_name?.charAt(0) || '?' }}
                </el-avatar>
                <div class="comment-content">
                  <div class="comment-header">
                    <span class="comment-author">{{ comment.user?.display_name || '未知用户' }}</span>
                    <span class="comment-time">{{ formatTime(comment.created_at) }}</span>
                  </div>
                  <div class="comment-body">{{ comment.content }}</div>
                </div>
              </div>
              <div v-if="comments.length === 0" class="empty-comments">
                暂无评论
              </div>
            </div>
          </el-card>

          <!-- 活动记录 -->
          <el-card shadow="never" class="content-card">
            <template #header>
              <span class="card-title">活动记录</span>
            </template>
            <el-timeline v-if="activities.length > 0" class="activity-timeline">
              <el-timeline-item
                v-for="activity in activities"
                :key="activity.id"
                :timestamp="formatTime(activity.created_at)"
                placement="top"
              >
                <div class="activity-content">
                  <span class="activity-user">{{ activity.user_name }}</span>
                  <span class="activity-action">{{ activity.action }}</span>
                  <template v-if="activity.field">
                    <span class="activity-field">{{ activity.field }}</span>
                    <span v-if="activity.old_value" class="activity-old-value">{{ activity.old_value }}</span>
                    <el-icon v-if="activity.old_value"><ArrowRight /></el-icon>
                    <span v-if="activity.new_value" class="activity-new-value">{{ activity.new_value }}</span>
                  </template>
                </div>
              </el-timeline-item>
            </el-timeline>
            <div v-else class="empty-activities">暂无活动记录</div>
          </el-card>
        </el-col>

        <!-- 右侧：详细信息 -->
        <el-col :xs="24" :lg="8">
          <!-- 基本信息 -->
          <el-card shadow="never" class="info-card">
            <template #header>
              <span class="card-title">详细信息</span>
            </template>
            <div class="info-list">
              <div class="info-item">
                <span class="info-label">状态</span>
                <span class="info-value">
                  <el-tag :type="getStatusType(issue.status)" size="small">
                    {{ getStatusText(issue.status) }}
                  </el-tag>
                </span>
              </div>
              <div class="info-item">
                <span class="info-label">优先级</span>
                <span class="info-value">
                  <el-tag :type="getPriorityType(issue.priority)" size="small">
                    {{ issue.priority }}
                  </el-tag>
                </span>
              </div>
              <div class="info-item">
                <span class="info-label">项目</span>
                <span class="info-value">
                  <el-link type="primary" @click="$router.push(`/issues?project_key=${issue.project_key}`)">
                    {{ issue.project_key }}
                  </el-link>
                </span>
              </div>
              <div class="info-item">
                <span class="info-label">类型</span>
                <span class="info-value">{{ issue.issue_type?.display_name || '-' }}</span>
              </div>
              <div class="info-item">
                <span class="info-label">指派人</span>
                <span class="info-value">
                  <template v-if="issue.assignee">
                    <el-avatar :size="20" class="assignee-avatar">
                      {{ issue.assignee.display_name?.charAt(0) || '?' }}
                    </el-avatar>
                    {{ issue.assignee.display_name }}
                  </template>
                  <span v-else class="text-muted">未指派</span>
                </span>
              </div>
              <div class="info-item">
                <span class="info-label">创建者</span>
                <span class="info-value">{{ issue.reporter?.display_name || '未知' }}</span>
              </div>
              <div class="info-item">
                <span class="info-label">创建时间</span>
                <span class="info-value">{{ formatTime(issue.created_at) }}</span>
              </div>
              <div class="info-item">
                <span class="info-label">更新时间</span>
                <span class="info-value">{{ formatTime(issue.updated_at) }}</span>
              </div>
              <div v-if="issue.due_date" class="info-item">
                <span class="info-label">截止时间</span>
                <span class="info-value">{{ formatDate(issue.due_date) }}</span>
              </div>
            </div>
          </el-card>

          <!-- 关注人 -->
          <el-card shadow="never" class="info-card">
            <template #header>
              <div class="card-header-with-action">
                <span class="card-title">关注人 ({{ watchers.length }})</span>
                <el-button link type="primary" size="small" @click="handleAddWatcher">
                  <el-icon><Plus /></el-icon>
                </el-button>
              </div>
            </template>
            <div class="watcher-list">
              <div v-for="watcher in watchers" :key="watcher.id" class="watcher-item">
                <el-avatar :size="28" class="watcher-avatar">
                  {{ watcher.user?.display_name?.charAt(0) || '?' }}
                </el-avatar>
                <span class="watcher-name">{{ watcher.user?.display_name || '未知用户' }}</span>
              </div>
              <div v-if="watchers.length === 0" class="empty-watchers">
                暂无关注人
              </div>
            </div>
          </el-card>
        </el-col>
      </el-row>
    </template>

    <!-- 编辑对话框 -->
    <el-dialog v-model="editDialogVisible" title="编辑工单" width="600px" destroy-on-close>
      <el-form ref="editFormRef" :model="editForm" :rules="editRules" label-width="80px">
        <el-form-item label="标题" prop="title">
          <el-input v-model="editForm.title" maxlength="200" show-word-limit />
        </el-form-item>
        <el-form-item label="优先级" prop="priority">
          <el-select v-model="editForm.priority" style="width: 100%">
            <el-option label="P0 - 紧急" value="P0" />
            <el-option label="P1 - 高" value="P1" />
            <el-option label="P2 - 中" value="P2" />
            <el-option label="P3 - 低" value="P3" />
          </el-select>
        </el-form-item>
        <el-form-item label="指派人">
          <el-select v-model="editForm.assignee_id" placeholder="请选择指派人" style="width: 100%" clearable filterable>
            <el-option v-for="u in users" :key="u.id" :label="u.display_name" :value="u.id" />
          </el-select>
        </el-form-item>
        <el-form-item label="描述">
          <el-input v-model="editForm.description" type="textarea" :rows="4" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="editDialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="editLoading" @click="submitEdit">保存</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { useRoute } from 'vue-router'
import { ElMessage, type FormInstance, type FormRules } from 'element-plus'
import { User, Clock, Edit, ArrowDown, ArrowRight, Plus } from '@element-plus/icons-vue'
import {
  getIssueDetail,
  updateIssue,
  transitionIssue,
  getIssueComments,
  addIssueComment,
  getIssueActivities,
  getIssueWatchers,
} from '@/api/issue'
import { getAllUsers } from '@/api/user'
import type { Issue, IssueComment, IssueActivity, IssueWatcher, UpdateIssueRequest, IssueStatus } from '@/types/issue'
import type { UserOption } from '@/types/user'
import dayjs from 'dayjs'

const route = useRoute()

// 数据
const loading = ref(false)
const issue = ref<Issue | null>(null)
const comments = ref<IssueComment[]>([])
const activities = ref<IssueActivity[]>([])
const watchers = ref<IssueWatcher[]>([])
const users = ref<UserOption[]>([])

// 可用的状态转换
interface StatusTransition {
  status: IssueStatus
  label: string
}

const getAvailableTransitions = (currentStatus: IssueStatus): StatusTransition[] => {
  const transitions: Record<IssueStatus, StatusTransition[]> = {
    open: [
      { status: 'in_progress', label: '开始处理' },
      { status: 'closed', label: '关闭' },
    ],
    in_progress: [
      { status: 'resolved', label: '标记解决' },
      { status: 'open', label: '重新打开' },
    ],
    resolved: [
      { status: 'closed', label: '关闭' },
      { status: 'reopened', label: '重新打开' },
    ],
    closed: [
      { status: 'reopened', label: '重新打开' },
    ],
    reopened: [
      { status: 'in_progress', label: '开始处理' },
      { status: 'closed', label: '关闭' },
    ],
  }
  return transitions[currentStatus] || []
}

// 新评论
const newComment = ref('')
const commentLoading = ref(false)

// 编辑相关
const editDialogVisible = ref(false)
const editLoading = ref(false)
const editFormRef = ref<FormInstance>()
const editForm = reactive<UpdateIssueRequest>({
  title: '',
  description: '',
  priority: undefined,
  assignee_id: undefined,
})

const editRules: FormRules = {
  title: [{ required: true, message: '请输入标题', trigger: 'blur' }],
}

// 加载工单详情
const loadIssue = async () => {
  const key = route.params.key as string
  if (!key) return

  loading.value = true
  try {
    const { data } = await getIssueDetail(key)
    issue.value = data.data

    // 加载相关数据
    await Promise.all([
      loadComments(key),
      loadActivities(key),
      loadWatchers(key),
    ])
  } catch (error) {
    console.error('Failed to load issue:', error)
    ElMessage.error('加载工单失败')
  } finally {
    loading.value = false
  }
}

// 加载评论
const loadComments = async (key: string) => {
  try {
    const { data } = await getIssueComments(key)
    comments.value = data.data
  } catch (error) {
    console.error('Failed to load comments:', error)
  }
}

// 加载活动记录
const loadActivities = async (key: string) => {
  try {
    const { data } = await getIssueActivities(key)
    activities.value = data.data
  } catch (error) {
    console.error('Failed to load activities:', error)
  }
}

// 加载关注人
const loadWatchers = async (key: string) => {
  try {
    const { data } = await getIssueWatchers(key)
    watchers.value = data.data
  } catch (error) {
    console.error('Failed to load watchers:', error)
  }
}

// 状态流转
const handleTransition = async (status: IssueStatus) => {
  if (!issue.value) return

  try {
    await transitionIssue(issue.value.issue_key, status)
    ElMessage.success('状态更新成功')
    loadIssue()
  } catch (error) {
    console.error('Failed to transition:', error)
  }
}

// 提交评论
const submitComment = async () => {
  if (!issue.value || !newComment.value.trim()) return

  commentLoading.value = true
  try {
    await addIssueComment(issue.value.issue_key, { content: newComment.value })
    ElMessage.success('评论成功')
    newComment.value = ''
    loadComments(issue.value.issue_key)
  } catch (error) {
    console.error('Failed to add comment:', error)
  } finally {
    commentLoading.value = false
  }
}

// 打开编辑对话框
const handleEdit = async () => {
  if (!issue.value) return

  // 加载用户列表
  try {
    const { data } = await getAllUsers()
    users.value = data.data
  } catch (error) {
    console.error('Failed to load users:', error)
  }

  Object.assign(editForm, {
    title: issue.value.title,
    description: issue.value.description,
    priority: issue.value.priority,
    assignee_id: issue.value.assignee_id,
  })

  editDialogVisible.value = true
}

// 提交编辑
const submitEdit = async () => {
  if (!editFormRef.value || !issue.value) return

  await editFormRef.value.validate(async (valid) => {
    if (!valid) return

    editLoading.value = true
    try {
      await updateIssue(issue.value!.issue_key, editForm)
      ElMessage.success('更新成功')
      editDialogVisible.value = false
      loadIssue()
    } catch (error) {
      console.error('Failed to update issue:', error)
    } finally {
      editLoading.value = false
    }
  })
}

// 添加关注人
const handleAddWatcher = () => {
  // TODO: 实现添加关注人功能
  ElMessage.info('功能开发中')
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

const formatDate = (date: string) => {
  return dayjs(date).format('YYYY-MM-DD')
}

// 初始化
onMounted(() => {
  loadIssue()
})
</script>

<style scoped lang="scss">
.issue-detail-container {
  .issue-header {
    display: flex;
    justify-content: space-between;
    align-items: flex-start;
    margin-bottom: 24px;
    padding: 24px;
    background: #fff;
    border-radius: 8px;

    .header-main {
      flex: 1;

      .issue-breadcrumb {
        margin-bottom: 16px;
      }

      .issue-title {
        font-size: 24px;
        font-weight: 600;
        margin: 0 0 16px 0;
        color: #1f2937;
      }

      .issue-meta {
        display: flex;
        align-items: center;
        gap: 16px;
        color: #6b7280;
        font-size: 14px;

        .meta-item {
          display: flex;
          align-items: center;
          gap: 4px;
        }
      }
    }

    .header-actions {
      display: flex;
      gap: 12px;
    }
  }

  .content-card {
    margin-bottom: 20px;

    .card-title {
      font-weight: 600;
    }

    .card-header-with-action {
      display: flex;
      justify-content: space-between;
      align-items: center;
    }

    .description-content {
      white-space: pre-wrap;
      line-height: 1.8;
      color: #374151;
    }

    .empty-description,
    .empty-comments,
    .empty-activities {
      color: #909399;
      text-align: center;
      padding: 20px 0;
    }
  }

  .add-comment {
    margin-bottom: 20px;
    padding-bottom: 20px;
    border-bottom: 1px solid #f0f0f0;

    .comment-actions {
      margin-top: 12px;
      text-align: right;
    }
  }

  .comment-list {
    .comment-item {
      display: flex;
      gap: 12px;
      padding: 16px 0;
      border-bottom: 1px solid #f0f0f0;

      &:last-child {
        border-bottom: none;
      }

      .comment-avatar {
        background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
        flex-shrink: 0;
      }

      .comment-content {
        flex: 1;

        .comment-header {
          display: flex;
          align-items: center;
          gap: 12px;
          margin-bottom: 8px;

          .comment-author {
            font-weight: 500;
            color: #1f2937;
          }

          .comment-time {
            font-size: 12px;
            color: #909399;
          }
        }

        .comment-body {
          color: #374151;
          line-height: 1.6;
          white-space: pre-wrap;
        }
      }
    }
  }

  .activity-timeline {
    padding-left: 0;

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

      .activity-field {
        color: #909399;
        margin: 0 4px;
      }

      .activity-old-value {
        text-decoration: line-through;
        color: #f56c6c;
        margin: 0 4px;
      }

      .activity-new-value {
        color: #67c23a;
        margin: 0 4px;
      }
    }
  }

  .info-card {
    margin-bottom: 20px;

    .card-title {
      font-weight: 600;
    }

    .card-header-with-action {
      display: flex;
      justify-content: space-between;
      align-items: center;
    }
  }

  .info-list {
    .info-item {
      display: flex;
      justify-content: space-between;
      align-items: flex-start;
      padding: 12px 0;
      border-bottom: 1px solid #f0f0f0;

      &:last-child {
        border-bottom: none;
      }

      .info-label {
        color: #6b7280;
        font-size: 14px;
        flex-shrink: 0;
        width: 80px;
      }

      .info-value {
        color: #1f2937;
        font-size: 14px;
        text-align: right;
        display: flex;
        align-items: center;
        gap: 8px;
        flex-wrap: wrap;
        justify-content: flex-end;

        .assignee-avatar {
          background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
          font-size: 10px;
        }

        .text-muted {
          color: #909399;
        }
      }
    }
  }

  .watcher-list {
    .watcher-item {
      display: flex;
      align-items: center;
      gap: 8px;
      padding: 8px 0;

      .watcher-avatar {
        background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
        font-size: 12px;
      }

      .watcher-name {
        font-size: 14px;
        color: #374151;
      }
    }

    .empty-watchers {
      color: #909399;
      text-align: center;
      padding: 20px 0;
      font-size: 14px;
    }
  }
}
</style>
