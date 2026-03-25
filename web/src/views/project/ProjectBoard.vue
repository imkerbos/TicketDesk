<template>
  <div class="project-board-container">
    <!-- 页面头部 -->
    <div class="board-header">
      <div class="header-left">
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
    <el-dialog v-model="createDialogVisible" title="创建工单" width="640px" destroy-on-close class="create-dialog">
      <el-form ref="createFormRef" :model="createForm" :rules="createRules" label-position="top">
        <el-row :gutter="20">
          <el-col :span="12">
            <el-form-item label="项目">
              <el-input :model-value="projectKey" disabled />
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="类型" prop="issue_type_id">
              <el-select v-model="createForm.issue_type_id" placeholder="请选择类型" style="width: 100%" @change="handleIssueTypeChange">
                <el-option v-for="t in issueTypes" :key="t.id" :label="t.display_name" :value="t.id" />
              </el-select>
            </el-form-item>
          </el-col>
        </el-row>
        <el-form-item label="标题" prop="title">
          <el-input v-model="createForm.title" placeholder="请输入工单标题" maxlength="200" show-word-limit />
        </el-form-item>

        <!-- 字段方案驱动 -->
        <div v-if="fieldScheme.length > 0" v-loading="fieldSchemeLoading" class="custom-fields-section">
          <el-row :gutter="20">
            <el-col
              v-for="item in fieldScheme"
              :key="item.field_id"
              :span="getFieldColSpan(item.field?.field_type)"
            >
              <el-form-item :required="item.is_required">
                <template #label>
                  <span>{{ item.field?.field_name }}</span>
                  <el-tooltip v-if="item.field?.description" :content="item.field?.description" placement="top">
                    <el-icon class="field-hint"><QuestionFilled /></el-icon>
                  </el-tooltip>
                </template>
                <FieldRenderer
                  v-if="item.field"
                  v-model="customFieldValues[item.field_id]"
                  :field="item.field"
                  :scheme="item"
                  :project-key="projectKey"
                />
              </el-form-item>
            </el-col>
          </el-row>
        </div>
      </el-form>
      <template #footer>
        <el-button @click="createDialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="createLoading" @click="submitCreate">
          <el-icon><Check /></el-icon>
          创建
        </el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, computed, onMounted, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage, type FormInstance, type FormRules } from 'element-plus'
import { Check, DataAnalysis, Grid, Setting, Tickets, QuestionFilled } from '@element-plus/icons-vue'
import BoardIssueList from './components/BoardIssueList.vue'
import IssueDetail from '@/views/issue/IssueDetail.vue'
import { getProjectDetail, getProjectIssueTypes } from '@/api/project'
import { createIssue } from '@/api/issue'
import { getAllUsers } from '@/api/user'
import { getFieldScheme } from '@/api/field'
import type { Project, ProjectIssueType } from '@/types/project'
import type { UserOption } from '@/types/user'
import type { CreateIssueRequest } from '@/types/issue'
import type { FieldSchemeItem } from '@/types/field'
import { FieldRenderer } from '@/components/field'
import { extractBuiltinFields } from '@/utils/builtin-fields'

const route = useRoute()
const router = useRouter()

const projectKey = computed(() => route.params.key as string)
const selectedIssueKey = ref('')
const project = ref<Project | null>(null)

// 创建工单
const createDialogVisible = ref(false)
const createLoading = ref(false)
const createFormRef = ref<FormInstance>()
const issueTypes = ref<ProjectIssueType[]>([])
const users = ref<UserOption[]>([])
const fieldScheme = ref<FieldSchemeItem[]>([])
const customFieldValues = ref<Record<number, any>>({})
const fieldSchemeLoading = ref(false)

const createForm = reactive({
  issue_type_id: undefined as number | undefined,
  title: '',
})

const createRules: FormRules = {
  issue_type_id: [{ required: true, message: '请选择类型', trigger: 'change' }],
  title: [{ required: true, message: '请输入标题', trigger: 'blur' }],
}

// 初始化选中工单
const initSelectedKey = () => {
  const issueKey = route.params.issueKey as string
  selectedIssueKey.value = issueKey || ''
}

const loadProject = async () => {
  try {
    const { data } = await getProjectDetail(projectKey.value)
    project.value = data.data
  } catch (e) {
    console.error('Failed to load project:', e)
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

const handleCreateIssue = async () => {
  // 加载工单类型和用户
  try {
    const [typesRes, usersRes] = await Promise.all([
      getProjectIssueTypes(projectKey.value),
      getAllUsers(),
    ])
    issueTypes.value = typesRes.data.data
    users.value = usersRes.data.data
  } catch (e) {
    console.error('Failed to load options:', e)
    ElMessage.error('加载选项失败')
    return
  }
  // 重置表单
  Object.assign(createForm, {
    issue_type_id: undefined,
    title: '',
  })
  fieldScheme.value = []
  customFieldValues.value = {}
  createDialogVisible.value = true
}

// 根据字段类型决定列宽
const getFieldColSpan = (fieldType?: string): number => {
  if (fieldType === 'textarea' || fieldType === 'epic_link') {
    return 24
  }
  return 12
}

const handleIssueTypeChange = async (issueTypeId: number) => {
  fieldScheme.value = []
  customFieldValues.value = {}
  if (!issueTypeId) return
  fieldSchemeLoading.value = true
  try {
    const { data } = await getFieldScheme(projectKey.value, issueTypeId)
    const schemeItems = data.data || []
    fieldScheme.value = schemeItems.filter(item => item.is_visible_create)
    // 初始化字段默认值
    const arrayFieldTypes = ['multiselect', 'label', 'component']
    fieldScheme.value.forEach(item => {
      const fieldType = item.field?.field_type || ''
      if (arrayFieldTypes.includes(fieldType)) {
        if (item.default_value) {
          try {
            const parsed = JSON.parse(item.default_value)
            customFieldValues.value[item.field_id] = Array.isArray(parsed) ? parsed : []
          } catch {
            customFieldValues.value[item.field_id] = []
          }
        } else {
          customFieldValues.value[item.field_id] = []
        }
      } else if (item.default_value) {
        customFieldValues.value[item.field_id] = item.default_value
      }
    })
  } catch (error) {
    console.error('Failed to load field scheme:', error)
  } finally {
    fieldSchemeLoading.value = false
  }
}

const submitCreate = async () => {
  if (!createFormRef.value) return
  await createFormRef.value.validate(async (valid) => {
    if (!valid) return
    createLoading.value = true
    try {
      if (!createForm.issue_type_id) {
        ElMessage.error('请选择工单类型')
        createLoading.value = false
        return
      }

      // 校验必填字段
      for (const item of fieldScheme.value) {
        if (item.is_required) {
          const val = customFieldValues.value[item.field_id]
          const isEmpty = val === undefined || val === null || val === '' || (Array.isArray(val) && val.length === 0)
          if (isEmpty) {
            ElMessage.error(`请填写 ${item.field?.field_name}`)
            createLoading.value = false
            return
          }
        }
      }

      // 分离内置字段和扩展字段
      const { builtinValues, customFields } = extractBuiltinFields(fieldScheme.value, customFieldValues.value)

      const requestData: CreateIssueRequest = {
        project_key: projectKey.value,
        issue_type_id: createForm.issue_type_id,
        title: createForm.title,
        description: builtinValues.description || '',
        priority: builtinValues.priority || 'P2',
        assignee_id: builtinValues.assignee_id || undefined,
        planned_start_date: builtinValues.planned_start_date || undefined,
        planned_end_date: builtinValues.planned_end_date || undefined,
        epic_id: builtinValues.epic_id || undefined,
        custom_fields: customFields.length > 0 ? customFields : undefined,
      }
      const { data } = await createIssue(requestData)
      ElMessage.success('工单创建成功')
      createDialogVisible.value = false
      // 选中新创建的工单
      const newKey = data.data?.issue_key
      if (newKey) {
        handleSelectIssue(newKey)
      }
    } catch (e) {
      console.error('Failed to create issue:', e)
      ElMessage.error('创建工单失败')
    } finally {
      createLoading.value = false
    }
  })
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
