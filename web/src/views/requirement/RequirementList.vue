<template>
  <div class="requirement-list">
    <div class="page-header">
      <h1>需求管理</h1>
      <div class="header-actions">
        <el-button @click="router.push('/requirements/kanban')">
          <el-icon><Grid /></el-icon>
          看板视图
        </el-button>
        <el-button type="primary" @click="showCreateDialog = true">
          <el-icon><Plus /></el-icon>
          创建需求
        </el-button>
      </div>
    </div>

    <!-- 筛选条件 -->
    <el-card class="filter-card" shadow="never">
      <el-form :inline="true" :model="filters">
        <el-form-item label="需求池">
          <el-select v-model="filters.pool_id" placeholder="全部" clearable style="width: 180px">
            <el-option
              v-for="pool in pools"
              :key="pool.id"
              :label="pool.name"
              :value="pool.id"
            />
          </el-select>
        </el-form-item>
        <el-form-item label="状态">
          <el-select v-model="filters.status" placeholder="全部" clearable style="width: 120px">
            <el-option label="待评估" value="pending_review" />
            <el-option label="规划中" value="planning" />
            <el-option label="进行中" value="in_progress" />
            <el-option label="已完成" value="completed" />
            <el-option label="已搁置" value="on_hold" />
            <el-option label="已拒绝" value="rejected" />
          </el-select>
        </el-form-item>
        <el-form-item label="分类">
          <el-select v-model="filters.category" placeholder="全部" clearable style="width: 120px">
            <el-option label="功能需求" value="feature" />
            <el-option label="优化改进" value="optimization" />
            <el-option label="故障修复" value="bugfix" />
            <el-option label="安全合规" value="security" />
            <el-option label="基础设施" value="infrastructure" />
            <el-option label="其他" value="other" />
          </el-select>
        </el-form-item>
        <el-form-item label="优先级">
          <el-select v-model="filters.priority" placeholder="全部" clearable style="width: 100px">
            <el-option label="P0" value="P0" />
            <el-option label="P1" value="P1" />
            <el-option label="P2" value="P2" />
            <el-option label="P3" value="P3" />
          </el-select>
        </el-form-item>
        <el-form-item label="关键词">
          <el-input v-model="filters.keyword" placeholder="搜索标题或描述" clearable style="width: 200px" />
        </el-form-item>
        <el-form-item>
          <el-button type="primary" @click="loadData">查询</el-button>
          <el-button @click="resetFilters">重置</el-button>
        </el-form-item>
      </el-form>
    </el-card>

    <!-- 需求列表 -->
    <el-card class="table-card" shadow="never">
      <el-table :data="requirements" v-loading="loading" stripe>
        <el-table-column prop="title" label="需求名称" min-width="250">
          <template #default="{ row }">
            <span class="link" @click="handleViewDetail(row)">
              {{ row.title }}
            </span>
            <div class="tags" v-if="row.tags && row.tags.length">
              <el-tag v-for="tag in row.tags" :key="tag" size="small" type="info">{{ tag }}</el-tag>
            </div>
          </template>
        </el-table-column>
        <el-table-column prop="description" label="需求描述" min-width="200" show-overflow-tooltip>
          <template #default="{ row }">
            {{ row.description || '-' }}
          </template>
        </el-table-column>
        <el-table-column label="来源" width="100" show-overflow-tooltip>
          <template #default="{ row }">
            {{ row.reporter_name || '-' }}
          </template>
        </el-table-column>
        <el-table-column prop="category" label="分类" width="100">
          <template #default="{ row }">
            <el-tag :type="getCategoryType(row.category)" size="small">{{ getCategoryLabel(row.category) }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="priority" label="优先级" width="80" align="center">
          <template #default="{ row }">
            <el-tag :type="getPriorityType(row.priority)" size="small">{{ row.priority }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="负责人" width="100" show-overflow-tooltip>
          <template #default="{ row }">
            {{ row.assignee_name || '-' }}
          </template>
        </el-table-column>
        <el-table-column label="开始时间" width="110" show-overflow-tooltip>
          <template #default="{ row }">
            {{ formatDate(row.start_date) }}
          </template>
        </el-table-column>
        <el-table-column label="结束时间" width="110" show-overflow-tooltip>
          <template #default="{ row }">
            {{ formatDate(row.end_date) }}
          </template>
        </el-table-column>
        <el-table-column prop="status" label="状态" width="100">
          <template #default="{ row }">
            <el-tag :type="getStatusType(row.status)" size="small">{{ getStatusLabel(row.status) }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="progress" label="进度" min-width="150" show-overflow-tooltip>
          <template #default="{ row }">
            {{ row.progress || '-' }}
          </template>
        </el-table-column>
        <el-table-column prop="result" label="结果" min-width="150" show-overflow-tooltip>
          <template #default="{ row }">
            {{ row.result || '-' }}
          </template>
        </el-table-column>
        <el-table-column label="关联工单" width="150" show-overflow-tooltip>
          <template #default="{ row }">
            <div v-if="row.converted_issue_key" style="display: flex; align-items: center; gap: 4px;">
              <router-link
                :to="`/issues/${row.converted_issue_key}`"
                class="link"
                style="flex-shrink: 0;"
              >
                {{ row.converted_issue_key }}
              </router-link>
              <el-tag
                v-if="row.converted_issue_status"
                :type="getIssueStatusType(row.converted_issue_status)"
                size="small"
              >
                {{ getIssueStatusLabel(row.converted_issue_status) }}
              </el-tag>
            </div>
            <span v-else>-</span>
          </template>
        </el-table-column>
        <el-table-column label="操作" width="180" fixed="right">
          <template #default="{ row }">
            <el-button link type="primary" size="small" @click="handleEdit(row)">编辑</el-button>
            <el-button
              link
              type="success"
              size="small"
              @click="handleConvert(row)"
              v-if="row.status !== 'completed' && row.status !== 'rejected'"
            >
              转工单
            </el-button>
            <el-button link type="danger" size="small" @click="handleDelete(row)">删除</el-button>
          </template>
        </el-table-column>
      </el-table>

      <!-- 分页 -->
      <div class="pagination">
        <el-pagination
          v-model:current-page="pagination.page"
          v-model:page-size="pagination.page_size"
          :total="pagination.total"
          :page-sizes="[10, 20, 50, 100]"
          layout="total, sizes, prev, pager, next, jumper"
          @size-change="loadData"
          @current-change="loadData"
        />
      </div>
    </el-card>

    <!-- 创建/编辑对话框 -->
    <el-dialog
      v-model="showCreateDialog"
      :title="editingRequirement ? '编辑需求' : '创建需求'"
      width="700px"
    >
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
            <el-form-item label="来源" prop="reporter_id">
              <el-select v-model="form.reporter_id" placeholder="请选择报告人" filterable clearable style="width: 100%">
                <el-option
                  v-for="user in users"
                  :key="user.id"
                  :label="user.display_name"
                  :value="user.id"
                />
              </el-select>
            </el-form-item>
          </el-col>
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
        <el-row :gutter="20">
          <el-col :span="12">
            <el-form-item label="开始时间" prop="start_date">
              <el-date-picker
                v-model="form.start_date"
                type="date"
                placeholder="选择日期"
                value-format="YYYY-MM-DD"
                style="width: 100%"
              />
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="结束时间" prop="end_date">
              <el-date-picker
                v-model="form.end_date"
                type="date"
                placeholder="选择日期"
                value-format="YYYY-MM-DD"
                style="width: 100%"
              />
            </el-form-item>
          </el-col>
        </el-row>
        <el-form-item label="进度" prop="progress" v-if="editingRequirement">
          <el-input
            v-model="form.progress"
            type="textarea"
            :rows="3"
            placeholder="请输入当前进度描述"
          />
        </el-form-item>
        <el-form-item label="结果" prop="result" v-if="editingRequirement">
          <el-input
            v-model="form.result"
            type="textarea"
            :rows="3"
            placeholder="请输入结果描述"
          />
        </el-form-item>
        <el-form-item label="目标项目" prop="target_project_id">
          <el-select v-model="form.target_project_id" placeholder="请选择目标项目" clearable style="width: 100%">
            <el-option
              v-for="project in projects"
              :key="project.id"
              :label="project.name"
              :value="project.id"
            />
          </el-select>
        </el-form-item>
        <el-form-item label="标签" prop="tags">
          <el-select
            v-model="form.tags"
            multiple
            filterable
            allow-create
            default-first-option
            placeholder="输入标签后回车"
            style="width: 100%"
          />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="showCreateDialog = false">取消</el-button>
        <el-button type="primary" @click="handleSubmit" :loading="submitting">确定</el-button>
      </template>
    </el-dialog>

    <!-- 转化为工单对话框 -->
    <el-dialog v-model="showConvertDialog" title="转化为工单" width="500px">
      <el-form :model="convertForm" :rules="convertRules" ref="convertFormRef" label-width="100px">
        <el-form-item label="目标项目" prop="project_id">
          <el-select v-model="convertForm.project_id" placeholder="请选择项目" style="width: 100%" @change="loadIssueTypes">
            <el-option
              v-for="project in projects"
              :key="project.id"
              :label="project.name"
              :value="project.id"
            />
          </el-select>
        </el-form-item>
        <el-form-item label="工单类型" prop="issue_type_id">
          <el-select v-model="convertForm.issue_type_id" placeholder="请选择工单类型" style="width: 100%">
            <el-option
              v-for="type in issueTypes"
              :key="type.id"
              :label="type.display_name"
              :value="type.id"
            />
          </el-select>
        </el-form-item>
        <el-form-item label="指派给" prop="assignee_id">
          <el-select v-model="convertForm.assignee_id" placeholder="请选择负责人" filterable clearable style="width: 100%">
            <el-option
              v-for="user in users"
              :key="user.id"
              :label="user.display_name"
              :value="user.id"
            />
          </el-select>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="showConvertDialog = false">取消</el-button>
        <el-button type="primary" @click="handleConvertSubmit" :loading="converting">转化</el-button>
      </template>
    </el-dialog>

    <!-- 详情抽屉 -->
    <el-drawer v-model="showDetailDrawer" title="需求详情" size="50%" v-if="selectedRequirement">
      <el-descriptions :column="2" border>
        <el-descriptions-item label="标题" :span="2">{{ selectedRequirement.title }}</el-descriptions-item>
        <el-descriptions-item label="需求池" :span="2">{{ selectedRequirement.pool_name }}</el-descriptions-item>
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
        <el-descriptions-item label="开始时间">{{ formatDate(selectedRequirement.start_date) }}</el-descriptions-item>
        <el-descriptions-item label="结束时间">{{ formatDate(selectedRequirement.end_date) }}</el-descriptions-item>
        <el-descriptions-item label="创建时间" :span="2">{{ formatDateTime(selectedRequirement.created_at) }}</el-descriptions-item>
      </el-descriptions>

      <div class="description-section" style="margin-top: 20px;">
        <h4>需求描述</h4>
        <p>{{ selectedRequirement.description || '暂无描述' }}</p>
      </div>

      <div class="description-section" style="margin-top: 20px;" v-if="selectedRequirement.progress">
        <h4>当前进度</h4>
        <p>{{ selectedRequirement.progress }}</p>
      </div>

      <div class="description-section" style="margin-top: 20px;" v-if="selectedRequirement.result">
        <h4>结果</h4>
        <p>{{ selectedRequirement.result }}</p>
      </div>

      <div class="tags-section" style="margin-top: 20px;" v-if="selectedRequirement.tags && selectedRequirement.tags.length">
        <h4>标签</h4>
        <el-tag v-for="tag in selectedRequirement.tags" :key="tag" style="margin-right: 8px;">{{ tag }}</el-tag>
      </div>

      <template #footer>
        <el-button @click="showDetailDrawer = false">关闭</el-button>
        <el-button type="primary" @click="handleEdit(selectedRequirement)">编辑</el-button>
        <el-button type="danger" @click="handleDelete(selectedRequirement)">删除</el-button>
      </template>
    </el-drawer>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { ElMessage, ElMessageBox, type FormInstance, type FormRules } from 'element-plus'
import { Plus, Grid } from '@element-plus/icons-vue'
import {
  getRequirementList,
  getRequirementPoolList,
  createRequirement,
  updateRequirement,
  deleteRequirement,
  convertToIssue,
} from '@/api/requirement'
import { getAllProjects, getProjectIssueTypes } from '@/api/project'
import { getAllUsers } from '@/api/user'
import type {
  Requirement,
  RequirementPool,
  CreateRequirementRequest,
  UpdateRequirementRequest,
  ConvertToIssueRequest,
  RequirementStatus,
  RequirementPriority,
  RequirementCategory,
} from '@/types/requirement'

const router = useRouter()
const route = useRoute()

// 数据
const requirements = ref<Requirement[]>([])
const pools = ref<RequirementPool[]>([])
const projects = ref<any[]>([])
const users = ref<any[]>([])
const issueTypes = ref<any[]>([])
const loading = ref(false)
const submitting = ref(false)
const converting = ref(false)
const showCreateDialog = ref(false)
const showConvertDialog = ref(false)
const showDetailDrawer = ref(false)
const editingRequirement = ref<Requirement | null>(null)
const convertingRequirement = ref<Requirement | null>(null)
const selectedRequirement = ref<Requirement | null>(null)
const formRef = ref<FormInstance>()
const convertFormRef = ref<FormInstance>()

// 筛选条件
const filters = reactive({
  pool_id: undefined as number | undefined,
  status: undefined as RequirementStatus | undefined,
  priority: undefined as RequirementPriority | undefined,
  category: undefined as RequirementCategory | undefined,
  keyword: '',
})

// 分页
const pagination = reactive({
  page: 1,
  page_size: 20,
  total: 0,
})

// 表单
const form = reactive<CreateRequirementRequest & { progress?: string; result?: string }>({
  pool_id: undefined,
  title: '',
  description: '',
  priority: 'P2',
  category: 'feature',
  reporter_id: undefined,
  assignee_id: undefined,
  start_date: undefined,
  end_date: undefined,
  target_project_id: undefined,
  tags: [],
  progress: undefined,
  result: undefined,
})

// 转化表单
const convertForm = reactive<ConvertToIssueRequest>({
  project_id: undefined,
  issue_type_id: undefined,
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

const convertRules: FormRules = {
  project_id: [{
    required: true,
    trigger: 'change',
    validator: (_rule, value, callback) => {
      if (!value || value === 0) {
        callback(new Error('请选择项目'))
      } else {
        callback()
      }
    }
  }],
  issue_type_id: [{
    required: true,
    trigger: 'change',
    validator: (_rule, value, callback) => {
      if (!value || value === 0) {
        callback(new Error('请选择工单类型'))
      } else {
        callback()
      }
    }
  }],
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

// 格式化日期
const formatDate = (dateStr: string | undefined) => {
  if (!dateStr) return '-'
  const date = new Date(dateStr)
  const year = date.getFullYear()
  const month = String(date.getMonth() + 1).padStart(2, '0')
  const day = String(date.getDate()).padStart(2, '0')
  return `${year}-${month}-${day}`
}

// 加载数据
const loadData = async () => {
  loading.value = true
  try {
    const { data } = await getRequirementList({
      ...filters,
      page: pagination.page,
      page_size: pagination.page_size,
    })
    requirements.value = data.data.items
    pagination.total = data.data.total
  } catch (error) {
    ElMessage.error('加载需求列表失败')
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

// 加载项目列表
const loadProjects = async () => {
  try {
    const { data } = await getAllProjects()
    projects.value = data.data
  } catch (error) {
    console.error('加载项目列表失败', error)
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

// 加载工单类型
const loadIssueTypes = async () => {
  if (!convertForm.project_id) {
    issueTypes.value = []
    return
  }
  try {
    const project = projects.value.find(p => p.id === convertForm.project_id)
    if (project) {
      const { data } = await getProjectIssueTypes(project.project_key)
      issueTypes.value = data.data
    }
  } catch (error) {
    console.error('加载工单类型失败', error)
  }
}

// 重置筛选条件
const resetFilters = () => {
  filters.pool_id = undefined
  filters.status = undefined
  filters.priority = undefined
  filters.category = undefined
  filters.keyword = ''
  pagination.page = 1
  loadData()
}

// 查看详情
const handleViewDetail = (requirement: Requirement) => {
  selectedRequirement.value = requirement
  showDetailDrawer.value = true
}

// 编辑
const handleEdit = (requirement: Requirement) => {
  editingRequirement.value = requirement
  form.pool_id = requirement.pool_id
  form.title = requirement.title
  form.description = requirement.description
  form.priority = requirement.priority
  form.category = requirement.category
  form.reporter_id = requirement.reporter_id
  form.assignee_id = requirement.assignee_id
  form.start_date = requirement.start_date
  form.end_date = requirement.end_date
  form.progress = requirement.progress
  form.result = requirement.result
  form.target_project_id = requirement.target_project_id
  form.tags = requirement.tags || []
  showCreateDialog.value = true
}

// 转化为工单
const handleConvert = (requirement: Requirement) => {
  convertingRequirement.value = requirement
  convertForm.project_id = requirement.target_project_id || undefined
  convertForm.issue_type_id = undefined
  convertForm.assignee_id = requirement.assignee_id
  if (convertForm.project_id) {
    loadIssueTypes()
  }
  showConvertDialog.value = true
}

// 删除
const handleDelete = async (requirement: Requirement) => {
  try {
    await ElMessageBox.confirm(`确定要删除需求"${requirement.title}"吗？`, '提示', {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      type: 'warning',
    })

    await deleteRequirement(requirement.id)
    ElMessage.success('删除成功')
    loadData()
  } catch (error: any) {
    if (error !== 'cancel') {
      ElMessage.error(error.response?.data?.message || '删除失败')
    }
  }
}

// 提交表单
const handleSubmit = async () => {
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
      if (editingRequirement.value) {
        // 更新
        const updateData: UpdateRequirementRequest = {
          title: form.title,
          description: form.description,
          priority: form.priority,
          category: form.category,
          reporter_id: form.reporter_id,
          assignee_id: form.assignee_id,
          start_date: form.start_date ? form.start_date : undefined,
          end_date: form.end_date ? form.end_date : undefined,
          progress: form.progress,
          result: form.result,
          target_project_id: form.target_project_id,
          tags: form.tags,
        }
        await updateRequirement(editingRequirement.value.id, updateData)
        ElMessage.success('更新成功')
      } else {
        // 创建 - 构建请求数据，只包含有值的字段
        const createData: any = {
          pool_id: form.pool_id,
          title: form.title,
          description: form.description || '',
          priority: form.priority,
          category: form.category,
          tags: form.tags || [],
        }

        // 只添加有值的可选字段
        if (form.reporter_id !== undefined) {
          createData.reporter_id = form.reporter_id
        }
        if (form.assignee_id !== undefined) {
          createData.assignee_id = form.assignee_id
        }
        if (form.start_date) {
          createData.start_date = new Date(form.start_date).toISOString()
        }
        if (form.end_date) {
          createData.end_date = new Date(form.end_date).toISOString()
        }
        if (form.target_project_id !== undefined) {
          createData.target_project_id = form.target_project_id
        }

        await createRequirement(createData)
        ElMessage.success('创建成功')
      }

      showCreateDialog.value = false
      resetForm()
      loadData()
    } catch (error: any) {
      ElMessage.error(error.response?.data?.message || '操作失败')
    } finally {
      submitting.value = false
    }
  })
}

// 提交转化
const handleConvertSubmit = async () => {
  if (!convertFormRef.value || !convertingRequirement.value) return

  await convertFormRef.value.validate(async (valid) => {
    if (!valid) return

    // 再次检查必填字段
    if (!convertForm.project_id) {
      ElMessage.error('请选择项目')
      return
    }
    if (!convertForm.issue_type_id) {
      ElMessage.error('请选择工单类型')
      return
    }

    converting.value = true
    try {
      // 构建请求数据，只包含有值的字段
      const convertData: any = {
        project_id: convertForm.project_id,
        issue_type_id: convertForm.issue_type_id,
      }

      if (convertForm.assignee_id !== undefined) {
        convertData.assignee_id = convertForm.assignee_id
      }

      const { data } = await convertToIssue(convertingRequirement.value!.id, convertData)

      // 显示成功消息，包含工单号
      ElMessage.success(`转化成功！工单号：${data.data.issue_key}`)

      showConvertDialog.value = false
      loadData()

      // 询问是否跳转到工单详情
      ElMessageBox.confirm(`需求已转化为工单 ${data.data.issue_key}，是否查看工单详情？`, '转化成功', {
        confirmButtonText: '查看工单',
        cancelButtonText: '留在当前页',
        type: 'success',
      }).then(() => {
        router.push(`/issues/${data.data.issue_key}`)
      }).catch(() => {
        // 用户选择留在当前页，不做任何操作
      })
    } catch (error: any) {
      ElMessage.error(error.response?.data?.message || '转化失败')
    } finally {
      converting.value = false
    }
  })
}

// 重置表单
const resetForm = () => {
  editingRequirement.value = null
  form.pool_id = undefined
  form.title = ''
  form.description = ''
  form.priority = 'P2'
  form.category = 'feature'
  form.reporter_id = undefined
  form.assignee_id = undefined
  form.start_date = undefined
  form.end_date = undefined
  form.progress = undefined
  form.result = undefined
  form.target_project_id = undefined
  form.tags = []
  formRef.value?.resetFields()
}

onMounted(() => {
  // 从 URL 参数获取 pool_id
  if (route.query.pool_id) {
    filters.pool_id = Number(route.query.pool_id)
  }

  loadData()
  loadPools()
  loadProjects()
  loadUsers()
})
</script>

<style scoped lang="scss">
.requirement-list {
  padding: 24px;
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

    :deep(.el-button--primary) {
      background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
      border: none;
      transition: all 0.3s ease;

      &:hover {
        transform: translateY(-2px);
        box-shadow: 0 4px 12px rgba(102, 126, 234, 0.4);
      }
    }
  }

  .table-card {
    border-radius: 12px;
    border: none;
    box-shadow: 0 2px 12px rgba(0, 0, 0, 0.08);
    overflow: hidden;

    :deep(.el-card__body) {
      padding: 0;
    }

    :deep(.el-table) {
      border-radius: 12px;

      th {
        background: linear-gradient(135deg, #f8f9fa 0%, #e9ecef 100%);
        color: #495057;
        font-weight: 600;
        font-size: 14px;
        padding: 16px 12px;
      }

      td {
        padding: 16px 12px;
      }

      .el-table__row {
        transition: all 0.3s ease;

        &:hover {
          background: #f8f9ff !important;
          transform: scale(1.005);
          box-shadow: 0 2px 8px rgba(102, 126, 234, 0.1);
        }
      }
    }

    .link {
      color: #667eea;
      text-decoration: none;
      font-weight: 500;
      font-size: 15px;
      transition: all 0.3s ease;

      &:hover {
        color: #764ba2;
        text-decoration: underline;
      }
    }

    .tags {
      margin-top: 8px;
      display: flex;
      flex-wrap: wrap;
      gap: 6px;

      .el-tag {
        border-radius: 6px;
        padding: 2px 10px;
        font-size: 12px;
        background: linear-gradient(135deg, #e3f2fd 0%, #bbdefb 100%);
        color: #1976d2;
        border: none;
        font-weight: 500;
      }
    }

    :deep(.el-tag) {
      border-radius: 6px;
      padding: 4px 12px;
      font-weight: 500;
      border: none;
    }

    :deep(.el-button--primary) {
      color: #667eea;

      &:hover {
        color: #764ba2;
        background: #f0f3ff;
      }
    }

    :deep(.el-button--success) {
      color: #67c23a;

      &:hover {
        color: #67c23a;
        background: #f0f9ff;
      }
    }

    :deep(.el-button--danger) {
      color: #f56c6c;

      &:hover {
        color: #f56c6c;
        background: #fef0f0;
      }
    }

    .pagination {
      margin-top: 0;
      padding: 20px 24px;
      display: flex;
      justify-content: flex-end;
      background: #fafbfc;
      border-top: 1px solid #e9ecef;
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
