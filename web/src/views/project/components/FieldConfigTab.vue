<template>
  <div class="field-config-container">
    <!-- 顶部操作栏 -->
    <div class="config-header">
      <div class="header-left">
        <el-segmented v-model="viewMode" :options="viewModeOptions" />
      </div>
      <div class="header-right">
        <el-button type="primary" @click="openFieldDialog()">
          <el-icon><Plus /></el-icon>
          创建自定义字段
        </el-button>
      </div>
    </div>

    <!-- 字段列表视图 -->
    <div v-if="viewMode === 'fields'" v-loading="fieldsLoading" class="fields-view">
      <div class="fields-section">
        <div class="section-title">
          <el-icon><Collection /></el-icon>
          <span>系统字段</span>
          <el-tag size="small" type="info">{{ systemFields.length }}</el-tag>
        </div>
        <div class="fields-grid">
          <div v-for="field in systemFields" :key="field.id" class="field-card system">
            <div class="field-icon" :class="getFieldTypeClass(field.field_type)">
              <el-icon><component :is="getFieldTypeIcon(field.field_type)" /></el-icon>
            </div>
            <div class="field-info">
              <div class="field-name">{{ field.field_name }}</div>
              <div class="field-meta">
                <span class="field-key">{{ field.field_key }}</span>
                <el-tag size="small" type="info">{{ getFieldTypeLabel(field.field_type) }}</el-tag>
              </div>
              <div v-if="field.description" class="field-desc">{{ field.description }}</div>
            </div>
            <div class="field-badge">
              <el-tag size="small">系统</el-tag>
            </div>
          </div>
        </div>
      </div>

      <div class="fields-section">
        <div class="section-title">
          <el-icon><EditPen /></el-icon>
          <span>自定义字段</span>
          <el-tag size="small" type="success">{{ customFields.length }}</el-tag>
        </div>
        <div v-if="customFields.length > 0" class="fields-grid">
          <div v-for="field in customFields" :key="field.id" class="field-card custom">
            <div class="field-icon" :class="getFieldTypeClass(field.field_type)">
              <el-icon><component :is="getFieldTypeIcon(field.field_type)" /></el-icon>
            </div>
            <div class="field-info">
              <div class="field-name">{{ field.field_name }}</div>
              <div class="field-meta">
                <span class="field-key">{{ field.field_key }}</span>
                <el-tag size="small" type="success">{{ getFieldTypeLabel(field.field_type) }}</el-tag>
              </div>
              <div v-if="field.description" class="field-desc">{{ field.description }}</div>
            </div>
            <div class="field-actions">
              <el-button size="small" @click="openFieldDialog(field)">
                <el-icon><Edit /></el-icon>
              </el-button>
              <el-button size="small" type="danger" @click="handleDeleteField(field)">
                <el-icon><Delete /></el-icon>
              </el-button>
            </div>
          </div>
        </div>
        <el-empty v-else description="暂无自定义字段" :image-size="80">
          <el-button type="primary" @click="openFieldDialog()">
            <el-icon><Plus /></el-icon>
            创建第一个自定义字段
          </el-button>
        </el-empty>
      </div>
    </div>

    <!-- 字段方案视图（按工单类型配置） -->
    <div v-else-if="viewMode === 'schemes'" v-loading="schemesLoading" class="schemes-view">
      <div class="issue-type-selector">
        <span class="selector-label">选择工单类型:</span>
        <el-select v-model="selectedIssueTypeId" placeholder="请选择工单类型" @change="loadFieldScheme">
          <el-option
            v-for="type in issueTypes"
            :key="type.id"
            :label="type.display_name"
            :value="type.id"
          >
            <div class="type-option">
              <div class="type-color" :style="{ background: type.color || '#3b82f6' }"></div>
              <span>{{ type.display_name }}</span>
            </div>
          </el-option>
        </el-select>
      </div>

      <div v-if="selectedIssueTypeId" class="scheme-config">
        <div class="scheme-header">
          <div class="scheme-title">
            <span>{{ getSelectedIssueTypeName() }} 字段配置</span>
            <el-tag size="small" type="info">{{ currentScheme.length }} 个字段</el-tag>
          </div>
          <el-button type="primary" @click="openSchemeFieldDialog">
            <el-icon><Plus /></el-icon>
            添加字段
          </el-button>
        </div>

        <el-table :data="currentScheme" stripe class="scheme-table">
          <el-table-column label="字段" min-width="200">
            <template #default="{ row }">
              <div class="field-cell">
                <div class="field-icon small" :class="getFieldTypeClass(row.field?.field_type || '')">
                  <el-icon><component :is="getFieldTypeIcon(row.field?.field_type || '')" /></el-icon>
                </div>
                <div class="field-info">
                  <div class="field-name">{{ row.field?.field_name }}</div>
                  <div class="field-key">{{ row.field?.field_key }}</div>
                </div>
              </div>
            </template>
          </el-table-column>
          <el-table-column label="类型" width="120">
            <template #default="{ row }">
              <el-tag size="small" :type="row.field?.is_system ? 'info' : 'success'">
                {{ row.field?.is_system ? '系统' : '自定义' }}
              </el-tag>
            </template>
          </el-table-column>
          <el-table-column label="必填" width="80" align="center">
            <template #default="{ row }">
              <el-switch
                v-model="row.is_required"
                size="small"
                @change="handleSchemeChange(row)"
              />
            </template>
          </el-table-column>
          <el-table-column label="创建时显示" width="100" align="center">
            <template #default="{ row }">
              <el-switch
                v-model="row.is_visible_create"
                size="small"
                @change="handleSchemeChange(row)"
              />
            </template>
          </el-table-column>
          <el-table-column label="编辑时显示" width="100" align="center">
            <template #default="{ row }">
              <el-switch
                v-model="row.is_visible_edit"
                size="small"
                @change="handleSchemeChange(row)"
              />
            </template>
          </el-table-column>
          <el-table-column label="详情显示" width="100" align="center">
            <template #default="{ row }">
              <el-switch
                v-model="row.is_visible_detail"
                size="small"
                @change="handleSchemeChange(row)"
              />
            </template>
          </el-table-column>
          <el-table-column label="排序" width="100" align="center">
            <template #default="{ row }">
              <el-input-number
                v-model="row.sort_order"
                :min="0"
                :max="999"
                size="small"
                controls-position="right"
                @change="handleSchemeChange(row)"
              />
            </template>
          </el-table-column>
          <el-table-column label="操作" width="80" align="center">
            <template #default="{ row }">
              <el-button
                size="small"
                type="danger"
                text
                @click="handleRemoveSchemeField(row)"
              >
                移除
              </el-button>
            </template>
          </el-table-column>
        </el-table>

        <el-empty v-if="currentScheme.length === 0" description="暂未配置字段" :image-size="80">
          <el-button type="primary" @click="openSchemeFieldDialog">
            <el-icon><Plus /></el-icon>
            添加字段
          </el-button>
        </el-empty>
      </div>

      <el-empty v-else description="请选择一个工单类型以配置字段" :image-size="100" />
    </div>

    <!-- 创建/编辑字段对话框 -->
    <el-dialog
      v-model="fieldDialogVisible"
      :title="isEditingField ? '编辑字段' : '创建自定义字段'"
      width="560px"
      destroy-on-close
      class="custom-dialog"
    >
      <el-form ref="fieldFormRef" :model="fieldForm" :rules="fieldRules" label-width="100px">
        <el-form-item label="字段标识" prop="field_key">
          <el-input
            v-model="fieldForm.field_key"
            placeholder="如: custom_status"
            :disabled="isEditingField"
          />
          <div class="form-tip">小写字母和下划线，创建后不可修改</div>
        </el-form-item>
        <el-form-item label="字段名称" prop="field_name">
          <el-input v-model="fieldForm.field_name" placeholder="如: 自定义状态" />
        </el-form-item>
        <el-form-item label="字段类型" prop="field_type">
          <el-select
            v-model="fieldForm.field_type"
            placeholder="请选择字段类型"
            style="width: 100%"
            :disabled="isEditingField"
          >
            <el-option
              v-for="type in fieldTypes"
              :key="type.value"
              :label="type.label"
              :value="type.value"
            >
              <div class="type-option">
                <el-icon><component :is="type.icon" /></el-icon>
                <span>{{ type.label }}</span>
              </div>
            </el-option>
          </el-select>
        </el-form-item>
        <el-form-item v-if="showOptionsInput" label="选项配置">
          <div class="options-editor">
            <div v-for="(opt, index) in fieldOptions" :key="index" class="option-item">
              <el-input v-model="opt.label" placeholder="选项显示名" style="width: 45%" />
              <el-input v-model="opt.value" placeholder="选项值" style="width: 45%" />
              <el-button type="danger" text @click="removeOption(index)">
                <el-icon><Delete /></el-icon>
              </el-button>
            </div>
            <el-button text @click="addOption">
              <el-icon><Plus /></el-icon>
              添加选项
            </el-button>
          </div>
        </el-form-item>
        <el-form-item label="描述">
          <el-input
            v-model="fieldForm.description"
            type="textarea"
            :rows="3"
            placeholder="字段描述"
          />
        </el-form-item>
        <el-form-item label="默认值">
          <el-input v-model="fieldForm.default_value" placeholder="默认值" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="fieldDialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="fieldSubmitLoading" @click="submitField">
          <el-icon><Check /></el-icon>
          确定
        </el-button>
      </template>
    </el-dialog>

    <!-- 添加字段到方案对话框 -->
    <el-dialog
      v-model="schemeFieldDialogVisible"
      title="添加字段到方案"
      width="500px"
      destroy-on-close
      class="custom-dialog"
    >
      <div class="scheme-field-list">
        <div
          v-for="field in availableFieldsForScheme"
          :key="field.id"
          class="scheme-field-item"
          :class="{ selected: selectedFieldIds.includes(field.id) }"
          @click="toggleFieldSelection(field.id)"
        >
          <el-checkbox :model-value="selectedFieldIds.includes(field.id)" />
          <div class="field-icon small" :class="getFieldTypeClass(field.field_type)">
            <el-icon><component :is="getFieldTypeIcon(field.field_type)" /></el-icon>
          </div>
          <div class="field-info">
            <div class="field-name">{{ field.field_name }}</div>
            <div class="field-meta">
              <span class="field-key">{{ field.field_key }}</span>
              <el-tag size="small" :type="field.is_system ? 'info' : 'success'">
                {{ field.is_system ? '系统' : '自定义' }}
              </el-tag>
            </div>
          </div>
        </div>
        <el-empty v-if="availableFieldsForScheme.length === 0" description="所有字段已添加" :image-size="60" />
      </div>
      <template #footer>
        <el-button @click="schemeFieldDialogVisible = false">取消</el-button>
        <el-button
          type="primary"
          :loading="addToSchemeLoading"
          :disabled="selectedFieldIds.length === 0"
          @click="addFieldsToScheme"
        >
          <el-icon><Check /></el-icon>
          添加 ({{ selectedFieldIds.length }})
        </el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, computed, onMounted, watch } from 'vue'
import { ElMessage, ElMessageBox, type FormInstance, type FormRules } from 'element-plus'
import {
  Plus,
  Edit,
  Delete,
  Check,
  Collection,
  EditPen,
  Document,
  List,
  Calendar,
  Select,
  User,
  PriceTag,
  Timer,
  Link,
  Grid,
} from '@element-plus/icons-vue'
import {
  getFields,
  createField,
  updateField,
  deleteField,
  getFieldScheme,
  updateFieldScheme,
} from '@/api/field'
import { getProjectIssueTypes } from '@/api/project'
import type {
  FieldDefinition,
  FieldSchemeItem,
  CreateFieldRequest,
  UpdateFieldRequest,
  FieldOption,
} from '@/types/field'
import type { ProjectIssueType } from '@/types/project'
import { getFieldTypeLabel, parseFieldOptions } from '@/types/field'

const props = defineProps<{
  projectKey: string
}>()

// 视图模式
const viewModeOptions = [
  { label: '字段列表', value: 'fields' },
  { label: '字段方案', value: 'schemes' },
]
const viewMode = ref('fields')

// 字段列表
const fieldsLoading = ref(false)
const allFields = ref<FieldDefinition[]>([])
const systemFields = computed(() => allFields.value.filter(f => f.is_system))
const customFields = computed(() => allFields.value.filter(f => !f.is_system))

// 字段方案
const schemesLoading = ref(false)
const issueTypes = ref<ProjectIssueType[]>([])
const selectedIssueTypeId = ref<number | undefined>()
const currentScheme = ref<FieldSchemeItem[]>([])

// 字段类型配置
const fieldTypes = [
  { value: 'text', label: '单行文本', icon: Document },
  { value: 'textarea', label: '多行文本', icon: Document },
  { value: 'number', label: '数字', icon: Grid },
  { value: 'date', label: '日期', icon: Calendar },
  { value: 'select', label: '单选下拉', icon: Select },
  { value: 'multiselect', label: '多选下拉', icon: List },
  { value: 'user', label: '用户选择', icon: User },
  { value: 'version', label: '版本选择', icon: PriceTag },
  { value: 'component', label: '组件选择', icon: Grid },
  { value: 'label', label: '标签选择', icon: PriceTag },
  { value: 'epic_link', label: 'Epic 链接', icon: Link },
  { value: 'time_estimate', label: '时间预估', icon: Timer },
]

// 字段对话框
const fieldDialogVisible = ref(false)
const isEditingField = ref(false)
const editingFieldId = ref<number | null>(null)
const fieldSubmitLoading = ref(false)
const fieldFormRef = ref<FormInstance>()
const fieldForm = reactive<CreateFieldRequest>({
  field_key: '',
  field_name: '',
  field_type: 'text',
  description: '',
  options: '',
  default_value: '',
})
const fieldOptions = ref<FieldOption[]>([])
const fieldRules: FormRules = {
  field_key: [
    { required: true, message: '请输入字段标识', trigger: 'blur' },
    { pattern: /^[a-z_]+$/, message: '只能包含小写字母和下划线', trigger: 'blur' },
  ],
  field_name: [{ required: true, message: '请输入字段名称', trigger: 'blur' }],
  field_type: [{ required: true, message: '请选择字段类型', trigger: 'change' }],
}

const showOptionsInput = computed(() => {
  return ['select', 'multiselect'].includes(fieldForm.field_type)
})

// 字段方案对话框
const schemeFieldDialogVisible = ref(false)
const selectedFieldIds = ref<number[]>([])
const addToSchemeLoading = ref(false)

const availableFieldsForScheme = computed(() => {
  const usedFieldIds = currentScheme.value.map(s => s.field_id)
  return allFields.value.filter(f => !usedFieldIds.includes(f.id))
})

// 加载字段列表
const loadFields = async () => {
  fieldsLoading.value = true
  try {
    const { data } = await getFields(props.projectKey)
    allFields.value = data.data || []
  } catch (error) {
    console.error('Failed to load fields:', error)
  } finally {
    fieldsLoading.value = false
  }
}

// 加载工单类型
const loadIssueTypes = async () => {
  try {
    const { data } = await getProjectIssueTypes(props.projectKey)
    issueTypes.value = data.data
    if (issueTypes.value.length > 0 && !selectedIssueTypeId.value) {
      selectedIssueTypeId.value = issueTypes.value[0].id
    }
  } catch (error) {
    console.error('Failed to load issue types:', error)
  }
}

// 加载字段方案
const loadFieldScheme = async () => {
  if (!selectedIssueTypeId.value) return
  schemesLoading.value = true
  try {
    const { data } = await getFieldScheme(props.projectKey, selectedIssueTypeId.value)
    currentScheme.value = data.data || []
  } catch (error) {
    console.error('Failed to load field scheme:', error)
  } finally {
    schemesLoading.value = false
  }
}

// 获取选中的工单类型名称
const getSelectedIssueTypeName = () => {
  const type = issueTypes.value.find(t => t.id === selectedIssueTypeId.value)
  return type?.display_name || ''
}

// 字段类型图标映射
const getFieldTypeIcon = (fieldType: string) => {
  const iconMap: Record<string, any> = {
    text: Document,
    textarea: Document,
    number: Grid,
    date: Calendar,
    select: Select,
    multiselect: List,
    user: User,
    version: PriceTag,
    component: Grid,
    label: PriceTag,
    epic_link: Link,
    time_estimate: Timer,
  }
  return iconMap[fieldType] || Document
}

// 字段类型样式类
const getFieldTypeClass = (fieldType: string) => {
  const classMap: Record<string, string> = {
    text: 'text',
    textarea: 'textarea',
    number: 'number',
    date: 'date',
    select: 'select',
    multiselect: 'multiselect',
    user: 'user',
    version: 'version',
    component: 'component',
    label: 'label',
    epic_link: 'epic',
    time_estimate: 'time',
  }
  return classMap[fieldType] || 'default'
}

// 打开字段对话框
const openFieldDialog = (field?: FieldDefinition) => {
  if (field) {
    isEditingField.value = true
    editingFieldId.value = field.id
    Object.assign(fieldForm, {
      field_key: field.field_key,
      field_name: field.field_name,
      field_type: field.field_type,
      description: field.description,
      options: field.options,
      default_value: field.default_value,
    })
    fieldOptions.value = parseFieldOptions(field.options)
  } else {
    isEditingField.value = false
    editingFieldId.value = null
    Object.assign(fieldForm, {
      field_key: '',
      field_name: '',
      field_type: 'text',
      description: '',
      options: '',
      default_value: '',
    })
    fieldOptions.value = []
  }
  fieldDialogVisible.value = true
}

// 添加选项
const addOption = () => {
  fieldOptions.value.push({ label: '', value: '' })
}

// 移除选项
const removeOption = (index: number) => {
  fieldOptions.value.splice(index, 1)
}

// 提交字段
const submitField = async () => {
  if (!fieldFormRef.value) return
  await fieldFormRef.value.validate(async (valid) => {
    if (!valid) return
    fieldSubmitLoading.value = true
    try {
      // 处理选项
      if (showOptionsInput.value && fieldOptions.value.length > 0) {
        fieldForm.options = JSON.stringify(fieldOptions.value.filter(o => o.label && o.value))
      } else {
        fieldForm.options = ''
      }

      if (isEditingField.value && editingFieldId.value) {
        const updateReq: UpdateFieldRequest = {
          field_name: fieldForm.field_name,
          description: fieldForm.description,
          options: fieldForm.options,
          default_value: fieldForm.default_value,
        }
        await updateField(props.projectKey, editingFieldId.value, updateReq)
        ElMessage.success('更新成功')
      } else {
        await createField(props.projectKey, fieldForm)
        ElMessage.success('创建成功')
      }
      fieldDialogVisible.value = false
      loadFields()
    } catch (error) {
      console.error('Failed to submit field:', error)
      ElMessage.error(isEditingField.value ? '更新失败' : '创建失败')
    } finally {
      fieldSubmitLoading.value = false
    }
  })
}

// 删除字段
const handleDeleteField = async (field: FieldDefinition) => {
  try {
    await ElMessageBox.confirm(`确定要删除字段 "${field.field_name}" 吗？`, '删除确认', {
      type: 'warning',
    })
    await deleteField(props.projectKey, field.id)
    ElMessage.success('删除成功')
    loadFields()
  } catch (error) {
    if (error !== 'cancel') {
      console.error('Failed to delete field:', error)
      ElMessage.error('删除失败')
    }
  }
}

// 打开添加字段到方案对话框
const openSchemeFieldDialog = () => {
  selectedFieldIds.value = []
  schemeFieldDialogVisible.value = true
}

// 切换字段选择
const toggleFieldSelection = (fieldId: number) => {
  const index = selectedFieldIds.value.indexOf(fieldId)
  if (index > -1) {
    selectedFieldIds.value.splice(index, 1)
  } else {
    selectedFieldIds.value.push(fieldId)
  }
}

// 添加字段到方案
const addFieldsToScheme = async () => {
  if (!selectedIssueTypeId.value || selectedFieldIds.value.length === 0) return
  addToSchemeLoading.value = true
  try {
    // 构建新的方案配置
    const newSchemeItems = selectedFieldIds.value.map((fieldId, index) => ({
      field_id: fieldId,
      is_required: false,
      is_visible_create: true,
      is_visible_edit: true,
      is_visible_detail: true,
      sort_order: currentScheme.value.length + index,
    }))

    const allItems = [
      ...currentScheme.value.map(s => ({
        field_id: s.field_id,
        is_required: s.is_required,
        is_visible_create: s.is_visible_create,
        is_visible_edit: s.is_visible_edit,
        is_visible_detail: s.is_visible_detail,
        sort_order: s.sort_order,
        default_value: s.default_value,
      })),
      ...newSchemeItems,
    ]

    await updateFieldScheme(props.projectKey, selectedIssueTypeId.value, { items: allItems })
    ElMessage.success('添加成功')
    schemeFieldDialogVisible.value = false
    loadFieldScheme()
  } catch (error) {
    console.error('Failed to add fields to scheme:', error)
    ElMessage.error('添加失败')
  } finally {
    addToSchemeLoading.value = false
  }
}

// 处理方案变更
const handleSchemeChange = async (_item: FieldSchemeItem) => {
  if (!selectedIssueTypeId.value) return
  try {
    const allItems = currentScheme.value.map(s => ({
      field_id: s.field_id,
      is_required: s.is_required,
      is_visible_create: s.is_visible_create,
      is_visible_edit: s.is_visible_edit,
      is_visible_detail: s.is_visible_detail,
      sort_order: s.sort_order,
      default_value: s.default_value,
    }))
    await updateFieldScheme(props.projectKey, selectedIssueTypeId.value, { items: allItems })
    ElMessage.success('已保存')
  } catch (error) {
    console.error('Failed to update scheme:', error)
    ElMessage.error('保存失败')
  }
}

// 从方案中移除字段
const handleRemoveSchemeField = async (item: FieldSchemeItem) => {
  if (!selectedIssueTypeId.value) return
  try {
    await ElMessageBox.confirm(`确定要从方案中移除字段 "${item.field?.field_name}" 吗？`, '移除确认', {
      type: 'warning',
    })
    const remainingItems = currentScheme.value
      .filter(s => s.field_id !== item.field_id)
      .map(s => ({
        field_id: s.field_id,
        is_required: s.is_required,
        is_visible_create: s.is_visible_create,
        is_visible_edit: s.is_visible_edit,
        is_visible_detail: s.is_visible_detail,
        sort_order: s.sort_order,
        default_value: s.default_value,
      }))
    await updateFieldScheme(props.projectKey, selectedIssueTypeId.value, { items: remainingItems })
    ElMessage.success('移除成功')
    loadFieldScheme()
  } catch (error) {
    if (error !== 'cancel') {
      console.error('Failed to remove field from scheme:', error)
      ElMessage.error('移除失败')
    }
  }
}

// 监听视图模式变化
watch(viewMode, (mode) => {
  if (mode === 'schemes') {
    loadIssueTypes()
    if (selectedIssueTypeId.value) {
      loadFieldScheme()
    }
  }
})

// 监听 projectKey 变化
watch(() => props.projectKey, () => {
  loadFields()
  if (viewMode.value === 'schemes') {
    loadIssueTypes()
  }
})

onMounted(() => {
  loadFields()
})
</script>

<style scoped lang="scss">
.field-config-container {
  min-height: 400px;
}

.config-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 24px;
  padding: 16px 20px;
  background: #fff;
  border-radius: 12px;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.05);
}

.fields-view {
  display: flex;
  flex-direction: column;
  gap: 24px;
}

.fields-section {
  background: #fff;
  border-radius: 12px;
  padding: 20px;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.05);
}

.section-title {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 16px;
  font-weight: 600;
  color: #1f2937;
  margin-bottom: 16px;
  padding-bottom: 12px;
  border-bottom: 1px solid #e5e7eb;
}

.fields-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(320px, 1fr));
  gap: 16px;
}

.field-card {
  display: flex;
  gap: 16px;
  padding: 16px;
  border: 1px solid #e5e7eb;
  border-radius: 10px;
  transition: all 0.2s;

  &:hover {
    box-shadow: 0 4px 12px rgba(0, 0, 0, 0.08);
    border-color: #d1d5db;
  }

  &.system {
    background: #f9fafb;
  }
}

.field-icon {
  width: 44px;
  height: 44px;
  border-radius: 10px;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 20px;
  color: #fff;
  flex-shrink: 0;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);

  &.small {
    width: 32px;
    height: 32px;
    font-size: 14px;
    border-radius: 8px;
  }

  &.text, &.textarea { background: linear-gradient(135deg, #3b82f6 0%, #2563eb 100%); }
  &.number { background: linear-gradient(135deg, #10b981 0%, #059669 100%); }
  &.date { background: linear-gradient(135deg, #f59e0b 0%, #d97706 100%); }
  &.select, &.multiselect { background: linear-gradient(135deg, #8b5cf6 0%, #7c3aed 100%); }
  &.user { background: linear-gradient(135deg, #ec4899 0%, #db2777 100%); }
  &.version, &.label { background: linear-gradient(135deg, #06b6d4 0%, #0891b2 100%); }
  &.component { background: linear-gradient(135deg, #6b7280 0%, #4b5563 100%); }
  &.epic { background: linear-gradient(135deg, #f97316 0%, #ea580c 100%); }
  &.time { background: linear-gradient(135deg, #14b8a6 0%, #0d9488 100%); }
}

.field-info {
  flex: 1;
  min-width: 0;

  .field-name {
    font-size: 14px;
    font-weight: 600;
    color: #1f2937;
    margin-bottom: 4px;
  }

  .field-meta {
    display: flex;
    align-items: center;
    gap: 8px;
    margin-bottom: 4px;
  }

  .field-key {
    font-size: 12px;
    color: #9ca3af;
    font-family: monospace;
  }

  .field-desc {
    font-size: 12px;
    color: #6b7280;
    line-height: 1.4;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }
}

.field-badge {
  flex-shrink: 0;
  align-self: flex-start;
}

.field-actions {
  display: flex;
  gap: 4px;
  flex-shrink: 0;
  align-self: flex-start;
}

// 方案视图
.schemes-view {
  background: #fff;
  border-radius: 12px;
  padding: 20px;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.05);
}

.issue-type-selector {
  display: flex;
  align-items: center;
  gap: 12px;
  margin-bottom: 20px;
  padding-bottom: 16px;
  border-bottom: 1px solid #e5e7eb;

  .selector-label {
    font-weight: 500;
    color: #374151;
  }

  .el-select {
    width: 240px;
  }
}

.type-option {
  display: flex;
  align-items: center;
  gap: 8px;

  .type-color {
    width: 16px;
    height: 16px;
    border-radius: 4px;
  }
}

.scheme-config {
  .scheme-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-bottom: 16px;
  }

  .scheme-title {
    display: flex;
    align-items: center;
    gap: 12px;
    font-size: 15px;
    font-weight: 600;
    color: #1f2937;
  }
}

.scheme-table {
  border-radius: 8px;
  overflow: hidden;

  :deep(.el-table__header) {
    th {
      background: #f9fafb !important;
      font-weight: 600;
      color: #374151;
    }
  }
}

.field-cell {
  display: flex;
  align-items: center;
  gap: 12px;
}

// 对话框样式
.custom-dialog {
  :deep(.el-dialog__header) {
    padding: 20px 24px;
    border-bottom: 1px solid #e5e7eb;
  }

  :deep(.el-dialog__body) {
    padding: 24px;
  }
}

.form-tip {
  font-size: 12px;
  color: #9ca3af;
  margin-top: 4px;
}

.options-editor {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.option-item {
  display: flex;
  gap: 8px;
  align-items: center;
}

// 方案字段选择对话框
.scheme-field-list {
  max-height: 400px;
  overflow-y: auto;
}

.scheme-field-item {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 12px;
  border: 1px solid #e5e7eb;
  border-radius: 8px;
  margin-bottom: 8px;
  cursor: pointer;
  transition: all 0.2s;

  &:hover {
    background: #f9fafb;
    border-color: #d1d5db;
  }

  &.selected {
    background: #f0f9ff;
    border-color: #3b82f6;
  }
}
</style>
