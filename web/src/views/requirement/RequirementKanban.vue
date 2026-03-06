<template>
  <div class="requirement-kanban">
    <div class="page-header">
      <h1>需求看板</h1>
      <div class="header-actions">
        <el-button @click="router.push('/requirements')">
          <el-icon><List /></el-icon>
          列表视图
        </el-button>
        <el-button type="primary" @click="handleCreate">
          <el-icon><Plus /></el-icon>
          创建需求
        </el-button>
      </div>
    </div>

    <!-- 筛选条件 -->
    <el-card class="filter-card" shadow="never">
      <el-form :inline="true" :model="filters">
        <el-form-item label="需求池">
          <el-select v-model="filters.pool_id" placeholder="全部" clearable style="width: 180px" @change="loadKanban">
            <el-option
              v-for="pool in pools"
              :key="pool.id"
              :label="pool.name"
              :value="pool.id"
            />
          </el-select>
        </el-form-item>
        <el-form-item label="分组方式">
          <el-select v-model="filters.group_by" style="width: 140px" @change="loadKanban">
            <el-option label="按状态" value="status" />
            <el-option label="按优先级" value="priority" />
            <el-option label="按负责人" value="assignee" />
            <el-option label="按时间线" value="timeline" />
          </el-select>
        </el-form-item>
      </el-form>
    </el-card>

    <!-- 看板 -->
    <div class="kanban-container" v-loading="loading">
      <div class="kanban-board">
        <div
          v-for="column in kanbanData.columns"
          :key="column.key"
          class="kanban-column"
          :class="{ 'completed-column': column.key === 'completed' }"
        >
          <div class="column-header" :class="{ 'completed-header': column.key === 'completed' }">
            <span class="column-title">{{ column.title }}</span>
            <el-badge :value="column.count" :max="99" type="info" />
          </div>
          <div class="column-content" :class="{ 'completed-content': column.key === 'completed' }">
            <div
              v-for="requirement in column.requirements"
              :key="requirement.id"
              class="kanban-card"
              @click="handleCardClick(requirement)"
            >
              <div class="card-header">
                <el-tag :type="getPriorityType(requirement.priority)" size="small">
                  {{ requirement.priority }}
                </el-tag>
                <el-tag :type="getCategoryType(requirement.category)" size="small">
                  {{ getCategoryLabel(requirement.category) }}
                </el-tag>
                <el-tag :type="getStatusType(requirement.status)" size="small">
                  {{ getStatusLabel(requirement.status) }}
                </el-tag>
              </div>
              <div class="card-title">{{ requirement.title }}</div>
              <div class="card-meta">
                <span v-if="requirement.reporter_name">
                  <el-icon><User /></el-icon>
                  {{ requirement.reporter_name }}
                </span>
                <span v-if="requirement.assignee_name">
                  <el-icon><User /></el-icon>
                  {{ requirement.assignee_name }}
                </span>
                <span v-if="requirement.end_date">
                  <el-icon><Calendar /></el-icon>
                  {{ formatDateTime(requirement.end_date) }}
                </span>
              </div>
              <!-- 关联工单信息 -->
              <div class="card-issue" v-if="requirement.converted_issue_key">
                <el-icon><Link /></el-icon>
                <span class="issue-key">{{ requirement.converted_issue_key }}</span>
                <el-tag
                  v-if="requirement.converted_issue_status"
                  :type="getIssueStatusType(requirement.converted_issue_status)"
                  size="small"
                >
                  {{ getIssueStatusLabel(requirement.converted_issue_status) }}
                </el-tag>
              </div>
              <div class="card-footer">
                <span class="pool-name">{{ requirement.pool_name }}</span>
                <span class="comment-count" v-if="requirement.comment_count > 0">
                  <el-icon><ChatDotRound /></el-icon>
                  {{ requirement.comment_count }}
                </span>
              </div>
            </div>
            <div v-if="column.requirements.length === 0" class="empty-column">
              暂无需求
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- 需求详情抽屉 -->
    <el-drawer
      v-model="showDetailDrawer"
      :title="selectedRequirement?.title"
      size="600px"
    >
      <template v-if="selectedRequirement">
        <el-descriptions :column="2" border>
          <el-descriptions-item label="需求池">{{ selectedRequirement.pool_name }}</el-descriptions-item>
          <el-descriptions-item label="分类">
            <el-tag :type="getCategoryType(selectedRequirement.category)" size="small">
              {{ getCategoryLabel(selectedRequirement.category) }}
            </el-tag>
          </el-descriptions-item>
          <el-descriptions-item label="优先级">
            <el-tag :type="getPriorityType(selectedRequirement.priority)" size="small">
              {{ selectedRequirement.priority }}
            </el-tag>
          </el-descriptions-item>
          <el-descriptions-item label="状态">
            <el-tag :type="getStatusType(selectedRequirement.status)" size="small">
              {{ getStatusLabel(selectedRequirement.status) }}
            </el-tag>
          </el-descriptions-item>
          <el-descriptions-item label="来源">{{ selectedRequirement.reporter_name || '-' }}</el-descriptions-item>
          <el-descriptions-item label="负责人">{{ selectedRequirement.assignee_name || '-' }}</el-descriptions-item>
          <el-descriptions-item label="创建人">{{ selectedRequirement.creator_name }}</el-descriptions-item>
          <el-descriptions-item label="关联工单">
            <div v-if="selectedRequirement.converted_issue_key">
              <router-link
                :to="`/issues/${selectedRequirement.converted_issue_key}`"
                class="link"
              >
                {{ selectedRequirement.converted_issue_key }}
              </router-link>
              <el-tag
                v-if="selectedRequirement.converted_issue_status"
                :type="getIssueStatusType(selectedRequirement.converted_issue_status)"
                size="small"
                style="margin-left: 8px;"
              >
                {{ getIssueStatusLabel(selectedRequirement.converted_issue_status) }}
              </el-tag>
            </div>
            <span v-else>-</span>
          </el-descriptions-item>
          <el-descriptions-item label="开始时间">{{ formatDateTime(selectedRequirement.start_date) }}</el-descriptions-item>
          <el-descriptions-item label="结束时间">{{ formatDateTime(selectedRequirement.end_date) }}</el-descriptions-item>
          <el-descriptions-item label="创建时间">{{ selectedRequirement.created_at }}</el-descriptions-item>
        </el-descriptions>

        <div class="description-section">
          <h4>描述</h4>
          <div class="description-content">{{ selectedRequirement.description || '暂无描述' }}</div>
        </div>

        <div class="description-section" v-if="selectedRequirement.progress">
          <h4>当前进度</h4>
          <div class="description-content">{{ selectedRequirement.progress }}</div>
        </div>

        <div class="description-section" v-if="selectedRequirement.result">
          <h4>结果</h4>
          <div class="description-content">{{ selectedRequirement.result }}</div>
        </div>

        <div class="tags-section" v-if="selectedRequirement.tags && selectedRequirement.tags.length">
          <h4>标签</h4>
          <div class="tags">
            <el-tag v-for="tag in selectedRequirement.tags" :key="tag" size="small">{{ tag }}</el-tag>
          </div>
        </div>

        <div class="actions-section">
          <el-button-group>
            <el-button @click="handleStatusChange('planning')" v-if="selectedRequirement.status === 'pending_review'">
              开始规划
            </el-button>
            <el-button type="primary" @click="handleStatusChange('in_progress')" v-if="selectedRequirement.status === 'pending_review' || selectedRequirement.status === 'planning' || selectedRequirement.status === 'on_hold'">
              开始执行
            </el-button>
            <el-button type="success" @click="handleStatusChange('completed')" v-if="selectedRequirement.status === 'in_progress'">
              完成
            </el-button>
            <el-button type="warning" @click="handleStatusChange('on_hold')" v-if="selectedRequirement.status === 'pending_review' || selectedRequirement.status === 'planning' || selectedRequirement.status === 'in_progress'">
              搁置
            </el-button>
            <el-button type="danger" @click="handleStatusChange('rejected')" v-if="selectedRequirement.status === 'pending_review' || selectedRequirement.status === 'planning' || selectedRequirement.status === 'in_progress'">
              拒绝
            </el-button>
            <el-button @click="handleStatusChange('pending_review')" v-if="selectedRequirement.status === 'rejected' || selectedRequirement.status === 'on_hold'">
              重新评估
            </el-button>
            <el-button type="primary" @click="handleConvert" v-if="selectedRequirement.status !== 'completed' && selectedRequirement.status !== 'rejected' && !selectedRequirement.converted_issue_id">
              转化为工单
            </el-button>
          </el-button-group>
        </div>
      </template>
    </el-drawer>

    <!-- 创建需求对话框 -->
    <el-dialog v-model="showCreateDialog" title="创建需求" width="700px" @closed="resetForm">
      <el-form :model="form" :rules="rules" ref="formRef" label-width="100px">
        <el-form-item label="需求池" prop="pool_id">
          <el-select v-model="form.pool_id" placeholder="请选择需求池" style="width: 100%">
            <el-option
              v-for="pool in pools"
              :key="pool.id"
              :label="pool.name"
              :value="pool.id"
            />
          </el-select>
        </el-form-item>
        <el-form-item label="标题" prop="title">
          <el-input v-model="form.title" placeholder="请输入需求标题" />
        </el-form-item>
        <el-form-item label="描述" prop="description">
          <el-input
            v-model="form.description"
            type="textarea"
            :rows="4"
            placeholder="请输入需求描述"
          />
        </el-form-item>
        <el-row :gutter="20">
          <el-col :span="12">
            <el-form-item label="分类" prop="category">
              <el-select v-model="form.category" placeholder="请选择分类" style="width: 100%">
                <el-option label="功能需求" value="feature" />
                <el-option label="优化改进" value="optimization" />
                <el-option label="故障修复" value="bugfix" />
                <el-option label="安全合规" value="security" />
                <el-option label="基础设施" value="infrastructure" />
                <el-option label="其他" value="other" />
              </el-select>
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="优先级" prop="priority">
              <el-select v-model="form.priority" placeholder="请选择优先级" style="width: 100%">
                <el-option label="P0 - 紧急" value="P0" />
                <el-option label="P1 - 高" value="P1" />
                <el-option label="P2 - 中" value="P2" />
                <el-option label="P3 - 低" value="P3" />
              </el-select>
            </el-form-item>
          </el-col>
        </el-row>
        <el-row :gutter="20">
          <el-col :span="12">
            <el-form-item label="负责人" prop="assignee_id">
              <el-select v-model="form.assignee_id" placeholder="请选择负责人" filterable clearable style="width: 100%">
                <el-option
                  v-for="user in users"
                  :key="user.id"
                  :label="user.display_name"
                  :value="user.id"
                />
              </el-select>
            </el-form-item>
          </el-col>
        </el-row>
      </el-form>
      <template #footer>
        <el-button @click="handleCancel">取消</el-button>
        <el-button type="primary" @click="handleCreateSubmit" :loading="submitting">确定</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage, type FormInstance, type FormRules } from 'element-plus'
import { Plus, List, User, Calendar, ChatDotRound, Link } from '@element-plus/icons-vue'
import {
  getRequirementKanban,
  getRequirementPoolList,
  createRequirement,
  updateRequirement,
} from '@/api/requirement'
import { getAllUsers } from '@/api/user'
import type {
  Requirement,
  RequirementPool,
  CreateRequirementRequest,
  KanbanResponse,
  RequirementStatus,
  RequirementPriority,
  RequirementCategory,
} from '@/types/requirement'

const router = useRouter()

// 数据
const pools = ref<RequirementPool[]>([])
const users = ref<any[]>([])
const kanbanData = ref<KanbanResponse>({ group_by: 'status', columns: [], total: 0 })
const loading = ref(false)
const submitting = ref(false)
const showDetailDrawer = ref(false)
const showCreateDialog = ref(false)
const selectedRequirement = ref<Requirement | null>(null)
const formRef = ref<FormInstance>()

// 筛选条件
const filters = reactive({
  pool_id: undefined as number | undefined,
  group_by: 'status' as 'status' | 'priority' | 'assignee' | 'timeline',
})

// 表单
const form = reactive<CreateRequirementRequest>({
  pool_id: undefined,
  title: '',
  description: '',
  priority: 'P2',
  category: 'feature',
  assignee_id: undefined,
})

// 表单验证规则
const rules: FormRules = {
  pool_id: [{
    required: true,
    trigger: 'change',
    validator: (_rule, value, callback) => {
      if (!value || value === 0) {
        callback(new Error('请选择需求池'))
      } else {
        callback()
      }
    }
  }],
  title: [{ required: true, message: '请输入需求标题', trigger: 'blur' }],
  priority: [{ required: true, message: '请选择优先级', trigger: 'change' }],
  category: [{ required: true, message: '请选择分类', trigger: 'change' }],
}

// 状态映射
type TagType = 'primary' | 'success' | 'warning' | 'info' | 'danger'

const getStatusLabel = (status: RequirementStatus) => {
  const map: Record<RequirementStatus, string> = {
    pending_review: '待评估',
    planning: '规划中',
    in_progress: '进行中',
    completed: '已完成',
    on_hold: '已搁置',
    rejected: '已拒绝',
  }
  return map[status] || status
}

const getStatusType = (status: RequirementStatus): TagType => {
  const map: Record<RequirementStatus, TagType> = {
    pending_review: 'info',
    planning: 'warning',
    in_progress: 'primary',
    completed: 'success',
    on_hold: 'info',
    rejected: 'danger',
  }
  return map[status] || 'info'
}

const getCategoryLabel = (category: RequirementCategory) => {
  const map: Record<RequirementCategory, string> = {
    feature: '功能需求',
    optimization: '优化改进',
    bugfix: '故障修复',
    security: '安全合规',
    infrastructure: '基础设施',
    other: '其他',
  }
  return map[category] || category
}

const getCategoryType = (category: RequirementCategory): TagType => {
  const map: Record<RequirementCategory, TagType> = {
    feature: 'primary',
    optimization: 'success',
    bugfix: 'danger',
    security: 'warning',
    infrastructure: 'info',
    other: 'info',
  }
  return map[category] || 'info'
}

const getPriorityType = (priority: RequirementPriority): TagType => {
  const map: Record<RequirementPriority, TagType> = {
    P0: 'danger',
    P1: 'warning',
    P2: 'primary',
    P3: 'info',
  }
  return map[priority] || 'info'
}

// 工单状态映射
const getIssueStatusLabel = (status: string) => {
  const map: Record<string, string> = {
    open: '待处理',
    'in-progress': '进行中',
    resolved: '已完成',
    closed: '已终止',
    reopened: '重新打开',
  }
  return map[status] || status
}

const getIssueStatusType = (status: string): TagType => {
  const map: Record<string, TagType> = {
    open: 'info',
    'in-progress': 'warning',
    resolved: 'success',
    closed: 'info',
    reopened: 'danger',
  }
  return map[status] || 'info'
}

// 格式化日期时间
const formatDateTime = (dateStr: string | undefined) => {
  if (!dateStr) return '-'
  const date = new Date(dateStr)
  const year = date.getFullYear()
  const month = String(date.getMonth() + 1).padStart(2, '0')
  const day = String(date.getDate()).padStart(2, '0')
  const hours = String(date.getHours()).padStart(2, '0')
  const minutes = String(date.getMinutes()).padStart(2, '0')
  return `${year}-${month}-${day} ${hours}:${minutes}`
}

// 加载看板数据
const loadKanban = async () => {
  loading.value = true
  try {
    const { data } = await getRequirementKanban({
      pool_id: filters.pool_id,
      group_by: filters.group_by,
    })
    kanbanData.value = data.data
  } catch (error) {
    ElMessage.error('加载看板数据失败')
  } finally {
    loading.value = false
  }
}

// 加载需求池列表
const loadPools = async () => {
  try {
    const { data } = await getRequirementPoolList({ status: 'active', page_size: 100 })
    pools.value = data.data.items
  } catch (error) {
    console.error('加载需求池列表失败', error)
  }
}

// 加载用户列表
const loadUsers = async () => {
  try {
    const { data } = await getAllUsers()
    users.value = data.data
  } catch (error) {
    console.error('加载用户列表失败', error)
  }
}

// 点击卡片
const handleCardClick = (requirement: Requirement) => {
  selectedRequirement.value = requirement
  showDetailDrawer.value = true
}

// 状态变更
const handleStatusChange = async (status: RequirementStatus) => {
  if (!selectedRequirement.value) return

  try {
    await updateRequirement(selectedRequirement.value.id, { status })
    ElMessage.success('状态更新成功')
    showDetailDrawer.value = false
    loadKanban()
  } catch (error: any) {
    ElMessage.error(error.response?.data?.message || '状态更新失败')
  }
}

// 转化为工单
const handleConvert = () => {
  if (!selectedRequirement.value) return
  router.push(`/requirements?convert=${selectedRequirement.value.id}`)
}

// 创建需求
const handleCreateSubmit = async () => {
  if (!formRef.value) return

  await formRef.value.validate(async (valid) => {
    if (!valid) return

    // 再次检查必填字段
    if (!form.pool_id) {
      ElMessage.error('请选择需求池')
      return
    }

    submitting.value = true
    try {
      // 构建请求数据，只包含有值的字段
      const createData: any = {
        pool_id: form.pool_id,
        title: form.title,
        description: form.description || '',
        priority: form.priority,
        category: form.category,
        tags: [],
      }

      // 只添加有值的可选字段
      if (form.assignee_id !== undefined) {
        createData.assignee_id = form.assignee_id
      }

      await createRequirement(createData)
      ElMessage.success('创建成功')
      showCreateDialog.value = false
      resetForm()
      loadKanban()
    } catch (error: any) {
      ElMessage.error(error.response?.data?.message || '创建失败')
    } finally {
      submitting.value = false
    }
  })
}

// 创建需求
const handleCreate = () => {
  resetForm()
  showCreateDialog.value = true
}

// 取消对话框
const handleCancel = () => {
  showCreateDialog.value = false
}

// 重置表单
const resetForm = () => {
  form.pool_id = undefined
  form.title = ''
  form.description = ''
  form.priority = 'P2'
  form.category = 'feature'
  form.assignee_id = undefined
  formRef.value?.resetFields()
}

onMounted(() => {
  loadKanban()
  loadPools()
  loadUsers()
})
</script>

<style scoped lang="scss">
.requirement-kanban {
  padding: 24px;
  height: 100%;
  display: flex;
  flex-direction: column;
  background: linear-gradient(135deg, #f5f7fa 0%, #e8eef5 100%);
  min-height: 100vh;

  .page-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-bottom: 24px;
    padding: 20px 24px;
    background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
    border-radius: 12px;
    box-shadow: 0 4px 20px rgba(102, 126, 234, 0.3);

    h1 {
      margin: 0;
      font-size: 28px;
      font-weight: 600;
      color: #fff;
      text-shadow: 0 2px 4px rgba(0, 0, 0, 0.1);
    }

    .header-actions {
      display: flex;
      gap: 12px;

      :deep(.el-button) {
        background: rgba(255, 255, 255, 0.2);
        border: 1px solid rgba(255, 255, 255, 0.3);
        color: #fff;
        backdrop-filter: blur(10px);
        transition: all 0.3s ease;

        &:hover {
          background: rgba(255, 255, 255, 0.3);
          transform: translateY(-2px);
          box-shadow: 0 4px 12px rgba(0, 0, 0, 0.15);
        }

        &.el-button--primary {
          background: rgba(255, 255, 255, 0.95);
          color: #667eea;
          font-weight: 500;

          &:hover {
            background: #fff;
            color: #764ba2;
          }
        }
      }
    }
  }

  .filter-card {
    margin-bottom: 20px;
    flex-shrink: 0;
    border-radius: 12px;
    border: none;
    box-shadow: 0 2px 12px rgba(0, 0, 0, 0.08);
    transition: box-shadow 0.3s ease;

    &:hover {
      box-shadow: 0 4px 20px rgba(0, 0, 0, 0.12);
    }

    :deep(.el-card__body) {
      padding: 20px 24px;
    }

    :deep(.el-form-item) {
      margin-bottom: 0;
    }
  }

  .kanban-container {
    flex: 1;
    overflow: hidden;

    .kanban-board {
      display: flex;
      gap: 20px;
      height: 100%;
      overflow-x: auto;
      padding-bottom: 10px;

      .kanban-column {
        flex: 0 0 320px;
        background: #fff;
        border-radius: 12px;
        display: flex;
        flex-direction: column;
        max-height: 100%;
        box-shadow: 0 2px 12px rgba(0, 0, 0, 0.08);
        transition: all 0.3s ease;

        &:hover {
          box-shadow: 0 4px 20px rgba(0, 0, 0, 0.12);
        }

        &.completed-column {
          background: linear-gradient(135deg, #ecfdf5 0%, #d1fae5 100%);
          border: 2px solid #6ee7b7;
        }

        .column-header {
          padding: 18px 20px;
          background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
          color: white;
          border-radius: 12px 12px 0 0;
          display: flex;
          justify-content: space-between;
          align-items: center;
          font-weight: 600;
          font-size: 15px;
          box-shadow: 0 2px 8px rgba(102, 126, 234, 0.3);

          &.completed-header {
            background: linear-gradient(135deg, #10b981 0%, #059669 100%);
            box-shadow: 0 2px 8px rgba(16, 185, 129, 0.3);

            :deep(.el-badge__content) {
              background: linear-gradient(135deg, #10b981 0%, #059669 100%);
            }
          }

          .column-title {
            font-weight: 600;
            font-size: 15px;
          }

          :deep(.el-badge__content) {
            background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
            border: none;
            font-weight: 600;
          }
        }

        .column-content {
          flex: 1;
          overflow-y: auto;
          padding: 16px;
          background: #fafbfc;

          &.completed-content {
            background: linear-gradient(180deg, #ecfdf5 0%, #f0fdf4 100%);
          }

          .kanban-card {
            background: #fff;
            border-radius: 10px;
            padding: 16px;
            margin-bottom: 12px;
            cursor: pointer;
            box-shadow: 0 2px 8px rgba(0, 0, 0, 0.08);
            border: 1px solid #e9ecef;
            transition: all 0.3s ease;

            &:hover {
              box-shadow: 0 6px 20px rgba(102, 126, 234, 0.2);
              transform: translateY(-4px);
              border-color: #667eea;
            }

            .card-header {
              display: flex;
              gap: 8px;
              margin-bottom: 12px;
              flex-wrap: wrap;

              :deep(.el-tag) {
                border-radius: 6px;
                padding: 4px 10px;
                font-size: 12px;
                font-weight: 600;
                border: none;
              }
            }

            .card-title {
              font-size: 15px;
              font-weight: 600;
              line-height: 1.5;
              margin-bottom: 12px;
              color: #2c3e50;
              display: -webkit-box;
              -webkit-line-clamp: 2;
              -webkit-box-orient: vertical;
              overflow: hidden;
            }

            .card-meta {
              display: flex;
              gap: 16px;
              font-size: 13px;
              color: #6c757d;
              margin-bottom: 12px;

              span {
                display: flex;
                align-items: center;
                gap: 6px;

                .el-icon {
                  font-size: 14px;
                  color: #667eea;
                }
              }
            }

            .card-issue {
              display: flex;
              align-items: center;
              gap: 8px;
              font-size: 12px;
              color: #495057;
              padding: 8px;
              background: linear-gradient(135deg, #e3f2fd 0%, #f3e5f5 100%);
              border-radius: 6px;
              margin-bottom: 8px;

              .el-icon {
                font-size: 14px;
                color: #667eea;
              }

              .issue-key {
                font-weight: 600;
                color: #667eea;
              }

              .el-tag {
                margin-left: auto;
              }
            }

            .card-footer {
              display: flex;
              justify-content: space-between;
              align-items: center;
              font-size: 12px;
              color: #6c757d;
              padding-top: 12px;
              border-top: 1px solid #f0f0f0;

              .pool-name {
                max-width: 180px;
                overflow: hidden;
                text-overflow: ellipsis;
                white-space: nowrap;
                font-weight: 500;
                color: #667eea;
              }

              .comment-count {
                display: flex;
                align-items: center;
                gap: 4px;
                color: #6c757d;

                .el-icon {
                  font-size: 14px;
                }
              }
            }
          }

          .empty-column {
            text-align: center;
            color: #adb5bd;
            padding: 60px 20px;
            font-size: 14px;
            background: #fff;
            border-radius: 8px;
            border: 2px dashed #dee2e6;
          }
        }
      }
    }
  }

  .description-section,
  .tags-section,
  .actions-section {
    margin-top: 24px;

    h4 {
      margin: 0 0 12px 0;
      font-size: 15px;
      color: #495057;
      font-weight: 600;
    }
  }

  .description-content {
    padding: 16px;
    background: linear-gradient(135deg, #f8f9fa 0%, #e9ecef 100%);
    border-radius: 8px;
    white-space: pre-wrap;
    line-height: 1.8;
    color: #495057;
    border: 1px solid #e9ecef;
  }

  .tags {
    display: flex;
    flex-wrap: wrap;
    gap: 8px;

    .el-tag {
      border-radius: 6px;
      padding: 6px 14px;
      font-weight: 500;
      background: linear-gradient(135deg, #e3f2fd 0%, #bbdefb 100%);
      color: #1976d2;
      border: none;
    }
  }

  :deep(.el-drawer) {
    border-radius: 12px 0 0 12px;

    .el-drawer__header {
      padding: 24px;
      margin-bottom: 0;
      border-bottom: 2px solid #e9ecef;
      background: linear-gradient(135deg, #f8f9fa 0%, #e9ecef 100%);

      .el-drawer__title {
        font-size: 18px;
        font-weight: 600;
        color: #2c3e50;
      }
    }

    .el-drawer__body {
      padding: 24px;
    }

    .el-descriptions {
      :deep(.el-descriptions__label) {
        font-weight: 600;
        color: #495057;
      }
    }

    .actions-section {
      :deep(.el-button-group) {
        display: flex;
        gap: 8px;
        flex-wrap: wrap;

        .el-button {
          border-radius: 8px;
          font-weight: 500;
          transition: all 0.3s ease;

          &:hover {
            transform: translateY(-2px);
            box-shadow: 0 4px 12px rgba(0, 0, 0, 0.15);
          }
        }
      }
    }
  }

  :deep(.el-dialog) {
    border-radius: 12px;
    box-shadow: 0 8px 32px rgba(0, 0, 0, 0.12);

    .el-dialog__header {
      padding: 20px 24px;
      background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
      border-radius: 12px 12px 0 0;

      .el-dialog__title {
        color: #fff;
        font-weight: 600;
        font-size: 18px;
      }

      .el-dialog__headerbtn .el-dialog__close {
        color: #fff;
        font-size: 20px;

        &:hover {
          color: #fff;
        }
      }
    }

    .el-dialog__body {
      padding: 24px;
    }

    .el-dialog__footer {
      padding: 16px 24px;
      border-top: 1px solid #e9ecef;

      .el-button--primary {
        background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
        border: none;
        padding: 10px 24px;
        transition: all 0.3s ease;

        &:hover {
          transform: translateY(-2px);
          box-shadow: 0 4px 12px rgba(102, 126, 234, 0.4);
        }
      }
    }
  }
}
</style>
