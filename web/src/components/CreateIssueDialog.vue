<template>
  <el-dialog
    :model-value="modelValue"
    :title="title || '创建工单'"
    width="640px"
    destroy-on-close
    class="create-issue-dialog"
    @update:model-value="$emit('update:modelValue', $event)"
    @open="handleDialogOpen"
  >
    <el-form
      ref="createFormRef"
      :model="createForm"
      :rules="createRules"
      label-position="top"
      class="create-issue-form"
    >
      <el-row :gutter="20">
        <el-col :span="12">
          <el-form-item label="项目" prop="project_key">
            <el-input
              v-if="fixedProjectKey"
              :model-value="fixedProjectKey"
              disabled
            />
            <el-select
              v-else
              v-model="createForm.project_key"
              placeholder="请选择项目"
              style="width: 100%"
              @change="handleProjectChange"
            >
              <el-option
                v-for="p in projectList"
                :key="p.project_key"
                :label="p.name"
                :value="p.project_key"
              />
            </el-select>
          </el-form-item>
        </el-col>
        <el-col :span="12">
          <el-form-item label="类型" prop="issue_type_id">
            <el-select
              v-model="createForm.issue_type_id"
              placeholder="请选择类型"
              style="width: 100%"
              :disabled="!effectiveProjectKey"
              @change="handleIssueTypeChange"
            >
              <el-option
                v-for="t in issueTypes"
                :key="t.id"
                :label="t.display_name"
                :value="t.id"
              />
            </el-select>
          </el-form-item>
        </el-col>
      </el-row>

      <el-form-item label="标题" prop="title">
        <el-input
          v-model="createForm.title"
          :placeholder="parentId ? '请输入子任务标题' : '请输入工单标题'"
          maxlength="200"
          show-word-limit
        />
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
                :project-key="effectiveProjectKey"
              />
            </el-form-item>
          </el-col>
        </el-row>
      </div>
    </el-form>

    <template #footer>
      <el-button @click="$emit('update:modelValue', false)">取消</el-button>
      <el-button type="primary" :loading="createLoading" @click="submitCreate">
        <el-icon><Check /></el-icon>
        {{ title || '创建工单' }}
      </el-button>
    </template>
  </el-dialog>
</template>

<script setup lang="ts">
import { ref, reactive, computed } from 'vue'
import { ElMessage, type FormInstance, type FormRules } from 'element-plus'
import { Check, QuestionFilled } from '@element-plus/icons-vue'
import { createIssue } from '@/api/issue'
import { getAllProjects, getProjectIssueTypes } from '@/api/project'
import { getFieldScheme } from '@/api/field'
import type { CreateIssueRequest } from '@/types/issue'
import type { Project, ProjectIssueType } from '@/types/project'
import type { FieldSchemeItem } from '@/types/field'
import { FieldRenderer } from '@/components/field'
import { extractBuiltinFields } from '@/utils/builtin-fields'

const props = withDefaults(defineProps<{
  modelValue: boolean
  title?: string
  fixedProjectKey?: string
  defaultProjectKey?: string
  parentId?: number
  projects?: Project[]
}>(), {
  title: '',
  fixedProjectKey: '',
  defaultProjectKey: '',
  parentId: undefined,
  projects: undefined,
})

const emit = defineEmits<{
  (e: 'update:modelValue', value: boolean): void
  (e: 'created', issueKey: string): void
}>()

// 内部状态
const createLoading = ref(false)
const createFormRef = ref<FormInstance>()
const internalProjects = ref<Project[]>([])
const issueTypes = ref<ProjectIssueType[]>([])
const fieldScheme = ref<FieldSchemeItem[]>([])
const customFieldValues = ref<Record<number, any>>({})
const fieldSchemeLoading = ref(false)

const createForm = reactive({
  project_key: '',
  issue_type_id: undefined as number | undefined,
  title: '',
})

const createRules = computed<FormRules>(() => {
  const rules: FormRules = {
    issue_type_id: [{ required: true, message: '请选择类型', trigger: 'change' }],
    title: [{ required: true, message: '请输入标题', trigger: 'blur' }],
  }
  if (!props.fixedProjectKey) {
    rules.project_key = [{ required: true, message: '请选择项目', trigger: 'change' }]
  }
  return rules
})

const projectList = computed(() => props.projects ?? internalProjects.value)
const effectiveProjectKey = computed(() => props.fixedProjectKey || createForm.project_key)

// 对话框打开时重置状态
const handleDialogOpen = async () => {
  createForm.project_key = props.defaultProjectKey || ''
  createForm.issue_type_id = undefined
  createForm.title = ''
  fieldScheme.value = []
  customFieldValues.value = {}
  issueTypes.value = []

  // 加载项目列表（仅当外部未传入且非固定项目时）
  if (!props.fixedProjectKey && !props.projects) {
    try {
      const { data } = await getAllProjects()
      internalProjects.value = data.data
    } catch {
      // ignored
    }
  }

  // 预选项目时自动加载工单类型
  const initialProject = props.fixedProjectKey || props.defaultProjectKey
  if (initialProject) {
    await handleProjectChange(initialProject)
  }
}

// 项目变更
const handleProjectChange = async (projectKey: string) => {
  createForm.issue_type_id = undefined
  fieldScheme.value = []
  customFieldValues.value = {}
  if (!projectKey) {
    issueTypes.value = []
    return
  }
  try {
    const { data } = await getProjectIssueTypes(projectKey)
    issueTypes.value = data.data
  } catch {
    // ignored
  }
}

// 工单类型变更
const handleIssueTypeChange = async (issueTypeId: number) => {
  fieldScheme.value = []
  customFieldValues.value = {}
  if (!issueTypeId || !effectiveProjectKey.value) return
  fieldSchemeLoading.value = true
  try {
    const { data } = await getFieldScheme(effectiveProjectKey.value, issueTypeId)
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
  } catch {
    // ignored
  } finally {
    fieldSchemeLoading.value = false
  }
}

// 字段列宽
const getFieldColSpan = (fieldType?: string): number => {
  if (fieldType === 'textarea' || fieldType === 'epic_link') {
    return 24
  }
  return 12
}

// 提交创建
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
        project_key: props.fixedProjectKey || createForm.project_key,
        issue_type_id: createForm.issue_type_id,
        title: createForm.title,
        description: builtinValues.description || '',
        priority: builtinValues.priority || 'P2',
        assignee_id: builtinValues.assignee_id || undefined,
        planned_start_date: builtinValues.planned_start_date || undefined,
        planned_end_date: builtinValues.planned_end_date || undefined,
        epic_id: builtinValues.epic_id || undefined,
        parent_id: props.parentId || undefined,
        custom_fields: customFields.length > 0 ? customFields : undefined,
      }
      const { data } = await createIssue(requestData)
      ElMessage.success(props.parentId ? '子任务创建成功' : '创建成功')
      emit('update:modelValue', false)
      emit('created', data.data.issue_key)
    } catch {
      // ignored
    } finally {
      createLoading.value = false
    }
  })
}
</script>

<style scoped lang="scss">
.create-issue-form {
  .el-form-item {
    margin-bottom: 20px;
  }

  .custom-fields-section {
    margin-top: 8px;

    .field-hint {
      margin-left: 4px;
      font-size: 14px;
      color: var(--td-text-disabled);
      cursor: help;
      vertical-align: middle;

      &:hover {
        color: var(--td-color-primary);
      }
    }
  }
}
</style>

<style lang="scss">
.create-issue-dialog {
  .el-dialog__header {
    padding: 16px 20px;
    margin-right: 0;
  }

  .el-dialog__body {
    padding: 20px;
    max-height: 70vh;
    overflow-y: auto;

    &::-webkit-scrollbar {
      width: 6px;
    }

    &::-webkit-scrollbar-thumb {
      background: var(--td-border-color);
      border-radius: 3px;
    }

    &::-webkit-scrollbar-track {
      background: transparent;
    }
  }

  .el-dialog__footer {
    padding: 12px 20px 16px;
  }

  .el-form-item__label {
    font-weight: 500;
    color: var(--td-text-secondary);
    font-size: 13px;
    padding-bottom: 4px;
  }

  .el-input, .el-select, .el-textarea {
    --el-input-bg-color: var(--td-input-bg);
  }

  .el-input__inner, .el-textarea__inner {
    &:focus {
      border-color: var(--td-color-primary);
      box-shadow: 0 0 0 2px rgba(64, 158, 255, 0.1);
    }
  }
}
</style>
