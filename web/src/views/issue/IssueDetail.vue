<template>
  <div v-loading="loading" class="issue-detail-container">
    <template v-if="issue">
      <!-- 头部信息 -->
      <div class="issue-header">
        <div class="header-left">
          <div class="issue-breadcrumb">
            <el-breadcrumb separator="/">
              <el-breadcrumb-item :to="{ path: '/issues' }">工单列表</el-breadcrumb-item>
              <el-breadcrumb-item :to="{ path: '/issues?project_key=' + issue.project_key }">{{ issue.project_key }}</el-breadcrumb-item>
              <el-breadcrumb-item>{{ issue.issue_key }}</el-breadcrumb-item>
            </el-breadcrumb>
          </div>
          <h1 class="issue-title">{{ issue.title }}</h1>
          <div class="issue-meta">
            <el-tag :type="getPriorityType(issue.priority)" size="small" effect="dark">{{ issue.priority }}</el-tag>
            <div class="status-badge" :class="issue.status">
              <span class="status-dot"></span>
              <span>{{ getStatusText(issue.status) }}</span>
            </div>
            <span class="meta-item">
              <el-icon><User /></el-icon>
              {{ issue.reporter?.display_name || '未知' }}
            </span>
            <span class="meta-item">
              <el-icon><Clock /></el-icon>
              {{ formatTime(issue.created_at) }}
            </span>
          </div>
        </div>
        <div class="header-actions">
          <el-dropdown v-if="issue && getAvailableTransitions(issue.status).length > 0" trigger="click" @command="handleTransition">
            <el-button type="primary" class="transition-btn">
              <el-icon><Switch /></el-icon>
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
              <div class="card-header-group">
                <div class="card-icon desc">
                  <el-icon><Document /></el-icon>
                </div>
                <span class="card-title">描述</span>
              </div>
            </template>
            <div v-if="issue.description" class="description-content">
              {{ issue.description }}
            </div>
            <div v-else class="empty-placeholder">暂无描述</div>
          </el-card>

          <!-- 评论 -->
          <el-card shadow="never" class="content-card">
            <template #header>
              <div class="card-header-with-action">
                <div class="card-header-group">
                  <div class="card-icon comment">
                    <el-icon><ChatLineRound /></el-icon>
                  </div>
                  <span class="card-title">评论</span>
                  <span class="card-count">{{ comments.length }}</span>
                </div>
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
                  <el-icon><ChatLineRound /></el-icon>
                  发表评论
                </el-button>
              </div>
            </div>

            <!-- 评论列表 -->
            <div class="comment-list">
              <div v-for="comment in comments" :key="comment.id" class="comment-item">
                <div class="comment-avatar">
                  {{ comment.user?.display_name?.charAt(0) || '?' }}
                </div>
                <div class="comment-body">
                  <div class="comment-header">
                    <span class="comment-author">{{ comment.user?.display_name || '未知用户' }}</span>
                    <span class="comment-time">{{ formatTime(comment.created_at) }}</span>
                  </div>
                  <div class="comment-text">{{ comment.content }}</div>
                </div>
              </div>
              <div v-if="comments.length === 0" class="empty-placeholder">
                暂无评论
              </div>
            </div>
          </el-card>

          <!-- 活动记录 -->
          <el-card shadow="never" class="content-card">
            <template #header>
              <div class="card-header-group">
                <div class="card-icon activity">
                  <el-icon><Clock /></el-icon>
                </div>
                <span class="card-title">活动记录</span>
              </div>
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
            <div v-else class="empty-placeholder">暂无活动记录</div>
          </el-card>
        </el-col>

        <!-- 右侧：详细信息 -->
        <el-col :xs="24" :lg="8">
          <!-- 基本信息 -->
          <el-card shadow="never" class="info-card">
            <template #header>
              <div class="card-header-group">
                <div class="card-icon info">
                  <el-icon><InfoFilled /></el-icon>
                </div>
                <span class="card-title">详细信息</span>
              </div>
            </template>
            <div class="info-list">
              <div class="info-item">
                <span class="info-label">状态</span>
                <span class="info-value">
                  <div class="status-badge sm" :class="issue.status">
                    <span class="status-dot"></span>
                    <span>{{ getStatusText(issue.status) }}</span>
                  </div>
                </span>
              </div>
              <div class="info-item">
                <span class="info-label">优先级</span>
                <span class="info-value">
                  <el-tag :type="getPriorityType(issue.priority)" size="small" effect="dark">{{ issue.priority }}</el-tag>
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
                    <div class="mini-avatar">{{ issue.assignee.display_name?.charAt(0) || '?' }}</div>
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
                <div class="card-header-group">
                  <div class="card-icon watcher">
                    <el-icon><View /></el-icon>
                  </div>
                  <span class="card-title">关注人</span>
                  <span class="card-count">{{ watchers.length }}</span>
                </div>
                <el-button link type="primary" size="small" @click="handleAddWatcher">
                  <el-icon><Plus /></el-icon>
                </el-button>
              </div>
            </template>
            <div class="watcher-list">
              <div v-for="watcher in watchers" :key="watcher.id" class="watcher-item">
                <div class="mini-avatar">{{ watcher.user?.display_name?.charAt(0) || '?' }}</div>
                <span class="watcher-name">{{ watcher.user?.display_name || '未知用户' }}</span>
              </div>
              <div v-if="watchers.length === 0" class="empty-placeholder sm">
                暂无关注人
              </div>
            </div>
          </el-card>
        </el-col>
      </el-row>
    </template>

    <!-- 编辑对话框 -->
    <el-dialog v-model="editDialogVisible" title="编辑工单" width="600px" destroy-on-close>
      <el-form ref="editFormRef" :model="editForm" :rules="editRules" label-position="top">
        <el-form-item label="标题" prop="title">
          <el-input v-model="editForm.title" maxlength="200" show-word-limit />
        </el-form-item>
        <el-row :gutter="16">
          <el-col :span="12">
            <el-form-item label="优先级" prop="priority">
              <el-select v-model="editForm.priority" style="width: 100%">
                <el-option label="P0 - 紧急" value="P0" />
                <el-option label="P1 - 高" value="P1" />
                <el-option label="P2 - 中" value="P2" />
                <el-option label="P3 - 低" value="P3" />
              </el-select>
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="指派人">
              <el-select v-model="editForm.assignee_id" placeholder="请选择" style="width: 100%" clearable filterable>
                <el-option v-for="u in users" :key="u.id" :label="u.display_name" :value="u.id" />
              </el-select>
            </el-form-item>
          </el-col>
        </el-row>
        <el-form-item label="描述">
          <el-input v-model="editForm.description" type="textarea" :rows="4" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="editDialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="editLoading" @click="submitEdit">
          <el-icon><Check /></el-icon>
          保存
        </el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { useRoute } from 'vue-router'
import { ElMessage, type FormInstance, type FormRules } from 'element-plus'
import {
  User, Clock, Edit, ArrowDown, ArrowRight, Plus, Document,
  ChatLineRound, InfoFilled, View, Check, Switch
} from '@element-plus/icons-vue'
import {
  getIssueDetail, updateIssue, transitionIssue,
  getIssueComments, addIssueComment, getIssueActivities, getIssueWatchers,
} from '@/api/issue'
import { getAllUsers } from '@/api/user'
import type { Issue, IssueComment, IssueActivity, IssueWatcher, UpdateIssueRequest, IssueStatus } from '@/types/issue'
import type { UserOption } from '@/types/user'
import dayjs from 'dayjs'

const route = useRoute()

const loading = ref(false)
const issue = ref<Issue | null>(null)
const comments = ref<IssueComment[]>([])
const activities = ref<IssueActivity[]>([])
const watchers = ref<IssueWatcher[]>([])
const users = ref<UserOption[]>([])

interface StatusTransition { status: IssueStatus; label: string }

const getAvailableTransitions = (currentStatus: IssueStatus): StatusTransition[] => {
  const transitions: Record<IssueStatus, StatusTransition[]> = {
    open: [{ status: 'in_progress', label: '开始处理' }, { status: 'closed', label: '关闭' }],
    in_progress: [{ status: 'resolved', label: '标记解决' }, { status: 'open', label: '重新打开' }],
    resolved: [{ status: 'closed', label: '关闭' }, { status: 'reopened', label: '重新打开' }],
    closed: [{ status: 'reopened', label: '重新打开' }],
    reopened: [{ status: 'in_progress', label: '开始处理' }, { status: 'closed', label: '关闭' }],
  }
  return transitions[currentStatus] || []
}

const newComment = ref('')
const commentLoading = ref(false)
const editDialogVisible = ref(false)
const editLoading = ref(false)
const editFormRef = ref<FormInstance>()
const editForm = reactive<UpdateIssueRequest>({
  title: '', description: '', priority: undefined, assignee_id: undefined,
})
const editRules: FormRules = {
  title: [{ required: true, message: '请输入标题', trigger: 'blur' }],
}

const loadIssue = async () => {
  const key = route.params.key as string
  if (!key) return
  loading.value = true
  try {
    const { data } = await getIssueDetail(key)
    issue.value = data.data
    await Promise.all([loadComments(key), loadActivities(key), loadWatchers(key)])
  } catch (error) {
    console.error('Failed to load issue:', error)
    ElMessage.error('加载工单失败')
  } finally {
    loading.value = false
  }
}

const loadComments = async (key: string) => {
  try { const { data } = await getIssueComments(key); comments.value = data.data } catch (e) { console.error(e) }
}
const loadActivities = async (key: string) => {
  try { const { data } = await getIssueActivities(key); activities.value = data.data.items || [] } catch (e) { console.error(e) }
}
const loadWatchers = async (key: string) => {
  try { const { data } = await getIssueWatchers(key); watchers.value = data.data } catch (e) { console.error(e) }
}

const handleTransition = async (status: IssueStatus) => {
  if (!issue.value) return
  try {
    await transitionIssue(issue.value.issue_key, status)
    ElMessage.success('状态更新成功')
    loadIssue()
  } catch (error) { console.error('Failed to transition:', error) }
}

const submitComment = async () => {
  if (!issue.value || !newComment.value.trim()) return
  commentLoading.value = true
  try {
    await addIssueComment(issue.value.issue_key, { content: newComment.value })
    ElMessage.success('评论成功')
    newComment.value = ''
    loadComments(issue.value.issue_key)
  } catch (error) { console.error(error) }
  finally { commentLoading.value = false }
}

const handleEdit = async () => {
  if (!issue.value) return
  try { const { data } = await getAllUsers(); users.value = data.data } catch (e) { console.error(e) }
  Object.assign(editForm, {
    title: issue.value.title, description: issue.value.description,
    priority: issue.value.priority, assignee_id: issue.value.assignee_id,
  })
  editDialogVisible.value = true
}

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
    } catch (error) { console.error(error) }
    finally { editLoading.value = false }
  })
}

const handleAddWatcher = () => { ElMessage.info('功能开发中') }

type TagType = 'primary' | 'success' | 'warning' | 'info' | 'danger'
const getPriorityType = (priority: string): TagType => {
  const map: Record<string, TagType> = { P0: 'danger', P1: 'warning', P2: 'info', P3: 'info' }
  return map[priority] || 'info'
}
const getStatusText = (status: string) => {
  const map: Record<string, string> = { open: '待处理', in_progress: '进行中', resolved: '已解决', closed: '已关闭', reopened: '重新打开' }
  return map[status] || status
}
const formatTime = (time: string) => dayjs(time).format('YYYY-MM-DD HH:mm')
const formatDate = (date: string) => dayjs(date).format('YYYY-MM-DD')

onMounted(() => { loadIssue() })
</script>

<style scoped lang="scss">
.issue-detail-container {
  width: 100%;
}

// 头部
.issue-header {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  margin-bottom: 24px;
  padding: 28px 32px;
  background: #fff;
  border-radius: 12px;
  box-shadow: 0 2px 12px rgba(0, 0, 0, 0.04);

  .header-left { flex: 1; }

  .issue-breadcrumb { margin-bottom: 16px; }

  .issue-title {
    font-size: 24px;
    font-weight: 700;
    margin: 0 0 16px 0;
    color: #1f2937;
    line-height: 1.4;
  }

  .issue-meta {
    display: flex;
    align-items: center;
    gap: 14px;
    color: #6b7280;
    font-size: 14px;
    flex-wrap: wrap;

    .meta-item {
      display: flex;
      align-items: center;
      gap: 5px;
    }
  }

  .header-actions {
    display: flex;
    gap: 10px;
    flex-shrink: 0;
    margin-left: 20px;
  }
}

// 状态徽章
.status-badge {
  display: inline-flex;
  align-items: center;
  gap: 5px;
  padding: 3px 12px;
  border-radius: 20px;
  font-size: 13px;
  font-weight: 500;

  .status-dot {
    width: 7px;
    height: 7px;
    border-radius: 50%;
  }

  &.open { background: #f3f4f6; color: #6b7280; .status-dot { background: #9ca3af; } }
  &.in_progress { background: #fff7ed; color: #c2410c; .status-dot { background: #f59e0b; } }
  &.resolved { background: #ecfdf5; color: #059669; .status-dot { background: #10b981; } }
  &.closed { background: #f3f4f6; color: #6b7280; .status-dot { background: #9ca3af; } }
  &.reopened { background: #fef2f2; color: #dc2626; .status-dot { background: #ef4444; } }

  &.sm { padding: 2px 10px; font-size: 12px; .status-dot { width: 6px; height: 6px; } }
}

// 通用卡片
.content-card, .info-card {
  margin-bottom: 20px;
  border-radius: 12px;

  :deep(.el-card__header) {
    padding: 16px 20px;
    border-bottom: 1px solid #f0f0f0;
  }
}

.card-header-group {
  display: flex;
  align-items: center;
  gap: 10px;
}

.card-header-with-action {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.card-icon {
  width: 32px;
  height: 32px;
  border-radius: 8px;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 16px;
  color: #fff;

  &.desc { background: linear-gradient(135deg, #667eea 0%, #764ba2 100%); }
  &.comment { background: linear-gradient(135deg, #11998e 0%, #38ef7d 100%); }
  &.activity { background: linear-gradient(135deg, #4facfe 0%, #00f2fe 100%); }
  &.info { background: linear-gradient(135deg, #f093fb 0%, #f5576c 100%); }
  &.watcher { background: linear-gradient(135deg, #fbbf24 0%, #f59e0b 100%); }
}

.card-title { font-size: 15px; font-weight: 600; color: #1f2937; }

.card-count {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  min-width: 20px;
  height: 20px;
  padding: 0 6px;
  background: #e5e7eb;
  border-radius: 10px;
  font-size: 12px;
  font-weight: 600;
  color: #4b5563;
}

.empty-placeholder {
  color: #9ca3af;
  text-align: center;
  padding: 32px 0;
  font-size: 14px;

  &.sm { padding: 20px 0; }
}

// 描述
.description-content {
  white-space: pre-wrap;
  line-height: 1.8;
  color: #374151;
  padding: 20px;
}

// 评论区
.add-comment {
  padding: 20px;
  border-bottom: 1px solid #f0f0f0;

  .comment-actions {
    margin-top: 12px;
    text-align: right;
  }
}

.comment-list {
  .comment-item {
    display: flex;
    gap: 14px;
    padding: 16px 20px;
    border-bottom: 1px solid #f5f5f5;

    &:last-child { border-bottom: none; }

    .comment-avatar {
      width: 38px;
      height: 38px;
      border-radius: 10px;
      background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
      color: #fff;
      display: flex;
      align-items: center;
      justify-content: center;
      font-size: 15px;
      font-weight: 600;
      flex-shrink: 0;
    }

    .comment-body {
      flex: 1;

      .comment-header {
        display: flex;
        align-items: center;
        gap: 12px;
        margin-bottom: 8px;

        .comment-author { font-weight: 600; color: #1f2937; font-size: 14px; }
        .comment-time { font-size: 12px; color: #9ca3af; }
      }

      .comment-text {
        color: #374151;
        line-height: 1.6;
        white-space: pre-wrap;
        font-size: 14px;
      }
    }
  }
}

// 活动时间线
.activity-timeline {
  padding: 20px;

  .activity-content {
    font-size: 14px;

    .activity-user { font-weight: 600; color: #1f2937; }
    .activity-action { color: #606266; margin: 0 4px; }
    .activity-field { color: #909399; margin: 0 4px; }
    .activity-old-value { text-decoration: line-through; color: #f56c6c; margin: 0 4px; }
    .activity-new-value { color: #67c23a; margin: 0 4px; }
  }
}

// 右侧信息
.info-list {
  .info-item {
    display: flex;
    justify-content: space-between;
    align-items: center;
    padding: 12px 20px;
    border-bottom: 1px solid #f5f5f5;

    &:last-child { border-bottom: none; }

    .info-label {
      color: #6b7280;
      font-size: 13px;
      flex-shrink: 0;
    }

    .info-value {
      color: #1f2937;
      font-size: 13px;
      text-align: right;
      display: flex;
      align-items: center;
      gap: 8px;
      justify-content: flex-end;
    }
  }
}

.mini-avatar {
  width: 22px;
  height: 22px;
  border-radius: 6px;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  color: #fff;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 10px;
  font-weight: 600;
  flex-shrink: 0;
}

.text-muted { color: #9ca3af; }

// 关注人列表
.watcher-list {
  .watcher-item {
    display: flex;
    align-items: center;
    gap: 10px;
    padding: 10px 20px;
    border-bottom: 1px solid #f5f5f5;

    &:last-child { border-bottom: none; }

    .watcher-name { font-size: 14px; color: #374151; }
  }
}

// 响应式
@media (max-width: 768px) {
  .issue-header {
    flex-direction: column;
    gap: 16px;

    .header-actions {
      margin-left: 0;
      width: 100%;
    }
  }
}
</style>
