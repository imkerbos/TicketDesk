<template>
  <div class="field-config-container">
    <!-- 顶部操作栏 -->
    <div class="config-header">
      <div class="header-left">
        <el-select v-model="selectedIssueTypeId" placeholder="请选择工单类型" style="width: 240px" @change="loadFieldScheme">
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
      <div class="header-right">
        <el-button @click="openApplyTemplateDialog">
          <el-icon><DocumentCopy /></el-icon>
          从模板套用
        </el-button>
        <el-button type="primary" @click="openSchemeFieldDialog">
          <el-icon><Plus /></el-icon>
          添加字段
        </el-button>
      </div>
    </div>

    <!-- 字段方案视图 -->
    <div v-loading="schemesLoading" class="schemes-view">
      <div v-if="selectedIssueTypeId" class="scheme-config">
        <div class="scheme-header">
          <div class="scheme-title">
            <span>{{ getSelectedIssueTypeName() }} 字段配置</span>
            <el-tag size="small" type="info">{{ currentScheme.length }} 个字段</el-tag>
          </div>
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

    <!-- 套用模板对话框 -->
    <el-dialog
      v-model="applyTemplateDialogVisible"
      title="从模板套用字段方案"
      width="520px"
      destroy-on-close
    >
      <div class="apply-form">
        <div class="apply-field">
          <label class="apply-label">选择模板</label>
          <el-select v-model="selectedTemplateId" placeholder="请选择模板" style="width: 100%">
            <el-option
              v-for="tpl in templates"
              :key="tpl.id"
              :label="tpl.name"
              :value="tpl.id"
            >
              <div class="template-option">
                <span>{{ tpl.name }}</span>
                <el-tag size="small" type="info">{{ tpl.item_count }} 个字段</el-tag>
              </div>
            </el-option>
          </el-select>
        </div>
        <div class="apply-field">
          <label class="apply-label">套用模式</label>
          <div class="mode-cards">
            <div
              class="mode-card"
              :class="{ active: applyMode === 'merge' }"
              @click="applyMode = 'merge'"
            >
              <div class="mode-card-radio">
                <div class="radio-dot" :class="{ checked: applyMode === 'merge' }"></div>
              </div>
              <div class="mode-card-content">
                <div class="mode-card-title">合并</div>
                <div class="mode-card-desc">保留现有字段，添加模板中新字段</div>
              </div>
            </div>
            <div
              class="mode-card"
              :class="{ active: applyMode === 'replace' }"
              @click="applyMode = 'replace'"
            >
              <div class="mode-card-radio">
                <div class="radio-dot" :class="{ checked: applyMode === 'replace' }"></div>
              </div>
              <div class="mode-card-content">
                <div class="mode-card-title">替换</div>
                <div class="mode-card-desc">删除现有字段，使用模板字段</div>
              </div>
            </div>
          </div>
        </div>
      </div>
      <template #footer>
        <el-button @click="applyTemplateDialogVisible = false">取消</el-button>
        <el-button
          type="primary"
          :loading="applyTemplateLoading"
          :disabled="!selectedTemplateId"
          @click="handleApplyTemplate"
        >
          套用
        </el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import {
  Plus,
  Check,
  Document,
  List,
  Calendar,
  Select,
  User,
  PriceTag,
  Timer,
  Link,
  Grid,
  DocumentCopy,
} from '@element-plus/icons-vue'
import {
  getFields,
  getFieldScheme,
  updateFieldScheme,
  applyTemplate,
} from '@/api/field'
import { getTemplates } from '@/api/admin-field'
import { getProjectIssueTypes } from '@/api/project'
import type {
  FieldDefinition,
  FieldSchemeItem,
  FieldSchemeTemplate,
} from '@/types/field'
import type { ProjectIssueType } from '@/types/project'

const props = defineProps<{
  projectKey: string
}>()

const route = useRoute()
const router = useRouter()

// 字段列表
const allFields = ref<FieldDefinition[]>([])

// 字段方案
const schemesLoading = ref(false)
const issueTypes = ref<ProjectIssueType[]>([])
const selectedIssueTypeId = ref<number | undefined>(
  route.query.issueTypeId ? Number(route.query.issueTypeId) : undefined
)
const currentScheme = ref<FieldSchemeItem[]>([])

// 同步状态到 URL
const syncStateToUrl = () => {
  const query: Record<string, string> = { ...route.query as Record<string, string> }
  if (selectedIssueTypeId.value) {
    query.issueTypeId = String(selectedIssueTypeId.value)
  } else {
    delete query.issueTypeId
  }
  router.replace({ query })
}

// 字段方案对话框
const schemeFieldDialogVisible = ref(false)
const selectedFieldIds = ref<number[]>([])
const addToSchemeLoading = ref(false)

// 套用模板
const applyTemplateDialogVisible = ref(false)
const templates = ref<FieldSchemeTemplate[]>([])
const selectedTemplateId = ref<number | undefined>()
const applyMode = ref<'replace' | 'merge'>('merge')
const applyTemplateLoading = ref(false)

const availableFieldsForScheme = computed(() => {
  const usedFieldIds = currentScheme.value.map(s => s.field_id)
  return allFields.value.filter(f => !usedFieldIds.includes(f.id))
})

// 加载字段列表
const loadFields = async () => {
  try {
    const { data } = await getFields(props.projectKey)
    allFields.value = data.data || []
  } catch {
    // ignored
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
  } catch {
    // ignored
  }
}

// 加载字段方案
const loadFieldScheme = async () => {
  if (!selectedIssueTypeId.value) return
  schemesLoading.value = true
  try {
    const { data } = await getFieldScheme(props.projectKey, selectedIssueTypeId.value)
    currentScheme.value = data.data || []
  } catch {
    // ignored
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
    await loadFieldScheme()
  } catch {
    ElMessage.error('添加失败')
  } finally {
    addToSchemeLoading.value = false
  }
}

// 处理方案变更（失败时回滚）
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
  } catch {
    ElMessage.error('保存失败')
    // 回滚：重新加载后端数据
    await loadFieldScheme()
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
    await loadFieldScheme()
  } catch (error) {
    if (error !== 'cancel') {
      ElMessage.error('移除失败')
    }
  }
}

// 打开套用模板对话框
const openApplyTemplateDialog = async () => {
  if (!selectedIssueTypeId.value) {
    ElMessage.warning('请先选择工单类型')
    return
  }
  try {
    const { data } = await getTemplates()
    templates.value = (data.data || []).filter((t: FieldSchemeTemplate) => t.is_active)
  } catch {
    // ignored
  }
  selectedTemplateId.value = undefined
  applyMode.value = 'merge'
  applyTemplateDialogVisible.value = true
}

// 套用模板
const handleApplyTemplate = async () => {
  if (!selectedIssueTypeId.value || !selectedTemplateId.value) return
  const modeName = applyMode.value === 'replace' ? '替换' : '合并'
  try {
    await ElMessageBox.confirm(
      `确定要以「${modeName}」模式套用此模板吗？${applyMode.value === 'replace' ? '这将删除当前所有字段配置。' : ''}`,
      '套用确认',
      { type: 'warning' }
    )
  } catch {
    return
  }

  applyTemplateLoading.value = true
  try {
    await applyTemplate(props.projectKey, selectedIssueTypeId.value, {
      template_id: selectedTemplateId.value,
      mode: applyMode.value,
    })
    ElMessage.success('模板套用成功')
    applyTemplateDialogVisible.value = false
    await loadFieldScheme()
  } catch {
    ElMessage.error('套用失败')
  } finally {
    applyTemplateLoading.value = false
  }
}

// 监听工单类型选择变化
watch(selectedIssueTypeId, () => {
  syncStateToUrl()
})

// 监听 projectKey 变化
watch(() => props.projectKey, async () => {
  await loadFields()
  await loadIssueTypes()
  if (selectedIssueTypeId.value) {
    await loadFieldScheme()
  }
})

onMounted(async () => {
  await loadFields()
  await loadIssueTypes()
  if (selectedIssueTypeId.value) {
    await loadFieldScheme()
  }
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
  background: var(--td-bg-card);
  border-radius: 12px;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.05);
}

.header-right {
  display: flex;
  gap: 8px;
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

// 方案视图
.schemes-view {
  background: var(--td-bg-card);
  border-radius: 12px;
  padding: 20px;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.05);
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
    color: var(--td-text-primary);
  }
}

.scheme-table {
  border-radius: 8px;
  overflow: hidden;

  :deep(.el-table__header) {
    th {
      background: var(--td-bg-page) !important;
      font-weight: 600;
      color: var(--td-text-regular);
    }
  }
}

.field-cell {
  display: flex;
  align-items: center;
  gap: 12px;
}

.field-icon {
  width: 44px;
  height: 44px;
  border-radius: 10px;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 20px;
  color: var(--td-text-white);
  flex-shrink: 0;
  background: var(--td-color-primary);

  &.small {
    width: 32px;
    height: 32px;
    font-size: 14px;
    border-radius: 8px;
  }

  &.text, &.textarea { background: var(--td-color-primary); }
  &.number { background: var(--td-color-success); }
  &.date { background: var(--td-color-warning); }
  &.select, &.multiselect { background: #8b5cf6; }
  &.user { background: #ec4899; }
  &.version, &.label { background: #06b6d4; }
  &.component { background: var(--td-text-secondary); }
  &.epic { background: #f97316; }
  &.time { background: #14b8a6; }
}

.field-info {
  flex: 1;
  min-width: 0;

  .field-name {
    font-size: 14px;
    font-weight: 600;
    color: var(--td-text-primary);
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
    color: var(--td-text-placeholder);
    font-family: monospace;
  }

  .field-desc {
    font-size: 12px;
    color: var(--td-text-secondary);
    line-height: 1.4;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }
}

// 对话框样式
.custom-dialog {
  :deep(.el-dialog__header) {
    padding: 20px 24px;
    border-bottom: 1px solid var(--td-border-color);
  }

  :deep(.el-dialog__body) {
    padding: 24px;
  }
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
  border: 1px solid var(--td-border-color);
  border-radius: 8px;
  margin-bottom: 8px;
  cursor: pointer;
  transition: all 150ms ease-out;

  &:hover {
    background: var(--td-bg-page);
    border-color: var(--td-text-disabled);
  }

  &.selected {
    background: var(--td-color-primary-light);
    border-color: var(--td-color-primary);
  }
}

// 套用模板对话框
.apply-form {
  display: flex;
  flex-direction: column;
  gap: 20px;
}

.apply-field {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.apply-label {
  font-size: 14px;
  font-weight: 500;
  color: var(--td-text-regular);
}

.template-option {
  display: flex;
  justify-content: space-between;
  align-items: center;
  width: 100%;
}

.mode-cards {
  display: flex;
  flex-direction: column;
  gap: 10px;
}

.mode-card {
  display: flex;
  align-items: flex-start;
  gap: 12px;
  padding: 14px 16px;
  border: 2px solid var(--td-border-color);
  border-radius: 10px;
  cursor: pointer;
  transition: all 150ms ease-out;

  &:hover {
    border-color: var(--td-tag-primary-border);
    background: var(--td-bg-section);
  }

  &.active {
    border-color: var(--td-color-primary);
    background: var(--td-tag-primary-bg);
  }
}

.mode-card-radio {
  padding-top: 2px;
  flex-shrink: 0;
}

.radio-dot {
  width: 18px;
  height: 18px;
  border-radius: 50%;
  border: 2px solid #d1d5db;
  transition: all 150ms ease-out;
  position: relative;

  &.checked {
    border-color: var(--td-color-primary);

    &::after {
      content: '';
      position: absolute;
      top: 3px;
      left: 3px;
      width: 8px;
      height: 8px;
      border-radius: 50%;
      background: var(--td-color-primary);
    }
  }
}

.mode-card-content {
  flex: 1;
}

.mode-card-title {
  font-size: 14px;
  font-weight: 600;
  color: var(--td-text-primary);
  line-height: 1.3;
}

.mode-card-desc {
  font-size: 12px;
  color: var(--td-text-secondary);
  margin-top: 3px;
  line-height: 1.4;
}
</style>
