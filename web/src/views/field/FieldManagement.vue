<template>
  <div class="field-management-container">
    <!-- 页面头部 -->
    <TdPageHeader>
      <template #leading>
        <div class="page-header-icon">
          <el-icon :size="20"><Grid /></el-icon>
        </div>
      </template>
      <template #title>字段管理</template>
      <template #subtitle>管理全局字段定义和方案模板</template>
    </TdPageHeader>

    <!-- Tab 切换 -->
    <el-card shadow="never" class="content-card">
      <el-tabs v-model="activeTab" class="custom-tabs">
        <!-- ==================== 全局字段 Tab ==================== -->
        <el-tab-pane name="fields">
          <template #label>
            <div class="tab-label">
              <el-icon><Grid /></el-icon>
              <span>全局字段</span>
            </div>
          </template>

          <!-- 字段列表头部 -->
          <div class="tab-header">
            <el-input
              v-model="fieldSearchText"
              placeholder="搜索字段名称或标识..."
              :prefix-icon="Search"
              clearable
              class="search-input"
            />
            <el-button type="primary" class="action-btn" @click="openCreateFieldDialog">
              <el-icon><Plus /></el-icon>
              创建字段
            </el-button>
          </div>

          <!-- 系统字段区 -->
          <div class="field-section">
            <div class="section-header" @click="systemFieldsCollapsed = !systemFieldsCollapsed">
              <div class="section-title-group">
                <el-icon class="collapse-arrow" :class="{ collapsed: systemFieldsCollapsed }"><ArrowDown /></el-icon>
                <span class="section-title">系统字段</span>
                <span class="section-count">{{ filteredSystemFields.length }}</span>
              </div>
              <span class="section-hint">系统内置字段，仅可切换启用状态和排序</span>
            </div>
            <transition name="collapse">
              <div v-show="!systemFieldsCollapsed" class="field-card-list">
                <div v-loading="fieldsLoading">
                  <div
                    v-for="field in filteredSystemFields"
                    :key="field.id"
                    class="field-card"
                  >
                    <div class="field-card-left">
                      <div class="field-type-icon" :class="getFieldTypeClass(field.field_type)">
                        <el-icon><component :is="getFieldTypeIcon(field.field_type)" /></el-icon>
                      </div>
                      <div class="field-card-info">
                        <div class="field-card-name">{{ field.field_name }}</div>
                        <div class="field-card-key">{{ field.field_key }}</div>
                      </div>
                      <div class="field-type-tag" :class="'tag-' + getFieldTypeClass(field.field_type)">
                        {{ getFieldTypeLabel(field.field_type) }}
                      </div>
                    </div>
                    <div class="field-card-right">
                      <div class="field-card-action">
                        <span class="action-label">排序</span>
                        <el-input
                          v-model.number="field.sort_order"
                          type="number"
                          size="small"
                          class="sort-input"
                          @change="handleUpdateFieldSort(field)"
                        />
                      </div>
                      <div class="field-card-action">
                        <el-switch
                          v-model="field.is_active"
                          size="small"
                          @change="handleToggleFieldActive(field)"
                        />
                      </div>
                    </div>
                  </div>
                  <TdEmptyState v-if="!fieldsLoading && filteredSystemFields.length === 0" preset="no-result" title="无匹配的系统字段" />
                </div>
              </div>
            </transition>
          </div>

          <!-- 全局自定义字段区 -->
          <div class="field-section">
            <div class="section-header">
              <div class="section-title-group">
                <span class="section-title">自定义字段</span>
                <span class="section-count">{{ filteredCustomFields.length }}</span>
              </div>
              <span class="section-hint">全局可用的自定义字段，可被所有项目引用</span>
            </div>
            <div v-loading="fieldsLoading" class="field-card-list">
              <div
                v-for="field in filteredCustomFields"
                :key="field.id"
                class="field-card"
              >
                <div class="field-card-left">
                  <div class="field-type-icon" :class="getFieldTypeClass(field.field_type)">
                    <el-icon><component :is="getFieldTypeIcon(field.field_type)" /></el-icon>
                  </div>
                  <div class="field-card-info">
                    <div class="field-card-name">{{ field.field_name }}</div>
                    <div class="field-card-key">{{ field.field_key }}</div>
                  </div>
                  <div class="field-type-tag" :class="'tag-' + getFieldTypeClass(field.field_type)">
                    {{ getFieldTypeLabel(field.field_type) }}
                  </div>
                  <div v-if="field.description" class="field-card-desc">{{ field.description }}</div>
                </div>
                <div class="field-card-right">
                  <el-switch
                    v-model="field.is_active"
                    size="small"
                    @change="handleToggleFieldActive(field)"
                  />
                  <el-button size="small" text type="primary" @click="handleEditField(field)">
                    <el-icon><Edit /></el-icon>
                  </el-button>
                  <el-button size="small" text type="danger" @click="handleDeleteField(field)">
                    <el-icon><Delete /></el-icon>
                  </el-button>
                </div>
              </div>
              <div v-if="!fieldsLoading && filteredCustomFields.length === 0 && !fieldSearchText" class="empty-custom">
                <TdEmptyState preset="first-time" title="暂无自定义字段" description="创建自定义字段，为工单添加更多属性">
                  <el-button type="primary" @click="openCreateFieldDialog">
                    <el-icon><Plus /></el-icon>
                    创建第一个字段
                  </el-button>
                </TdEmptyState>
              </div>
              <TdEmptyState v-if="!fieldsLoading && filteredCustomFields.length === 0 && fieldSearchText" preset="no-result" title="无匹配的自定义字段" />
            </div>
          </div>
        </el-tab-pane>

        <!-- ==================== 方案模板 Tab ==================== -->
        <el-tab-pane name="templates">
          <template #label>
            <div class="tab-label">
              <el-icon><CopyDocument /></el-icon>
              <span>方案模板</span>
            </div>
          </template>

          <!-- 模板列表头部 -->
          <div class="tab-header">
            <div class="tab-header-hint">模板定义了一组字段配置，可快速套用到项目的工单类型</div>
            <el-button type="primary" class="action-btn" @click="openCreateTemplateDialog">
              <el-icon><Plus /></el-icon>
              创建模板
            </el-button>
          </div>

          <!-- 模板卡片列表 -->
          <div v-loading="templatesLoading" class="template-card-list">
            <div
              v-for="tpl in templates"
              :key="tpl.id"
              class="template-card"
              @click="openTemplateDetail(tpl)"
            >
              <div class="template-card-header">
                <div class="template-icon">
                  <el-icon><CopyDocument /></el-icon>
                </div>
                <div class="template-meta">
                  <div class="template-name">{{ tpl.name }}</div>
                  <div class="template-desc">{{ tpl.description || '暂无描述' }}</div>
                </div>
                <el-tag
                  :type="tpl.is_active ? 'success' : 'info'"
                  size="small"
                  class="template-status"
                  effect="light"
                >
                  {{ tpl.is_active ? '启用' : '禁用' }}
                </el-tag>
              </div>
              <div class="template-card-footer">
                <div class="template-stat">
                  <el-icon><Grid /></el-icon>
                  <span>{{ tpl.item_count || 0 }} 个字段</span>
                </div>
                <div class="template-actions" @click.stop>
                  <el-button size="small" type="primary" plain @click="openTemplateDetail(tpl)">
                    <el-icon><Setting /></el-icon>
                    配置字段
                  </el-button>
                  <el-dropdown trigger="click">
                    <el-button size="small" class="more-btn">
                      <el-icon><MoreFilled /></el-icon>
                    </el-button>
                    <template #dropdown>
                      <el-dropdown-menu>
                        <el-dropdown-item @click="handleEditTemplate(tpl)">
                          <el-icon><Edit /></el-icon>
                          编辑信息
                        </el-dropdown-item>
                        <el-dropdown-item divided class="danger-item" @click="handleDeleteTemplate(tpl)">
                          <el-icon><Delete /></el-icon>
                          删除模板
                        </el-dropdown-item>
                      </el-dropdown-menu>
                    </template>
                  </el-dropdown>
                </div>
              </div>
            </div>
            <TdEmptyState v-if="!templatesLoading && templates.length === 0" preset="first-time" title="暂无方案模板" description="方案模板可快速套用到项目的工单类型字段配置">
              <el-button type="primary" @click="openCreateTemplateDialog">
                <el-icon><Plus /></el-icon>
                创建第一个模板
              </el-button>
            </TdEmptyState>
          </div>
        </el-tab-pane>
      </el-tabs>
    </el-card>

    <!-- 创建/编辑字段对话框 -->
    <el-dialog
      v-model="fieldDialogVisible"
      :title="isEditFieldMode ? '编辑字段' : '创建全局字段'"
      width="560px"
      destroy-on-close
      class="custom-dialog"
    >
      <el-form ref="fieldFormRef" :model="fieldForm" :rules="fieldFormRules" label-position="top">
        <el-form-item label="字段标识" prop="field_key">
          <el-input
            v-model="fieldForm.field_key"
            placeholder="如: custom_region (仅英文、数字、下划线)"
            :disabled="isEditFieldMode"
          />
        </el-form-item>
        <el-form-item label="字段名称" prop="field_name">
          <el-input v-model="fieldForm.field_name" placeholder="如: 所在区域" />
        </el-form-item>
        <el-form-item v-if="!isEditFieldMode" label="字段类型" prop="field_type">
          <el-select v-model="fieldForm.field_type" placeholder="请选择字段类型" style="width: 100%">
            <el-option
              v-for="(label, value) in fieldTypeOptions"
              :key="value"
              :label="label"
              :value="value"
            />
          </el-select>
        </el-form-item>
        <el-form-item label="描述">
          <el-input v-model="fieldForm.description" type="textarea" :rows="2" placeholder="字段说明" />
        </el-form-item>
        <el-form-item
          v-if="fieldForm.field_type === 'select' || fieldForm.field_type === 'multiselect'"
          label="选项配置"
        >
          <div class="options-editor">
            <div v-for="(opt, idx) in editingOptions" :key="idx" class="option-row">
              <el-input v-model="opt.label" placeholder="显示名称" size="small" class="option-input" />
              <el-input v-model="opt.value" placeholder="值（留空则同名称）" size="small" class="option-input" />
              <el-button type="danger" text size="small" @click="editingOptions.splice(idx, 1)">
                <el-icon><Delete /></el-icon>
              </el-button>
            </div>
            <el-button size="small" @click="editingOptions.push({ value: '', label: '' })">
              <el-icon><Plus /></el-icon> 添加选项
            </el-button>
          </div>
        </el-form-item>
        <el-form-item label="默认值">
          <el-input v-model="fieldForm.default_value" placeholder="字段默认值（可选）" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="fieldDialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="fieldSubmitLoading" @click="submitFieldForm">确定</el-button>
      </template>
    </el-dialog>

    <!-- 创建/编辑模板对话框 -->
    <el-dialog
      v-model="templateDialogVisible"
      :title="isEditTemplateMode ? '编辑模板' : '创建方案模板'"
      width="500px"
      destroy-on-close
      class="custom-dialog"
    >
      <el-form ref="templateFormRef" :model="templateForm" :rules="templateFormRules" label-position="top">
        <el-form-item label="模板名称" prop="name">
          <el-input v-model="templateForm.name" placeholder="如: Bug 标准模板" />
        </el-form-item>
        <el-form-item label="描述">
          <el-input v-model="templateForm.description" type="textarea" :rows="3" placeholder="模板描述" />
        </el-form-item>
        <el-form-item v-if="isEditTemplateMode" label="状态">
          <el-switch v-model="templateForm.is_active" active-text="启用" inactive-text="禁用" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="templateDialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="templateSubmitLoading" @click="submitTemplateForm">确定</el-button>
      </template>
    </el-dialog>

    <!-- 模板字段配置对话框 -->
    <el-dialog
      v-model="templateDetailVisible"
      :title="`模板字段配置 - ${currentTemplate?.name || ''}`"
      width="940px"
      destroy-on-close
      class="custom-dialog"
    >
      <div v-loading="templateDetailLoading" class="template-detail-content">
        <div class="template-detail-header">
          <div class="template-detail-info">
            <span class="detail-field-count">{{ templateItems.length }} 个字段</span>
            <el-tag v-if="templateItemsDirty" type="warning" size="small" effect="plain">未保存</el-tag>
          </div>
          <el-button type="primary" size="small" @click="openAddTemplateFieldDialog">
            <el-icon><Plus /></el-icon>
            添加字段
          </el-button>
        </div>

        <el-table :data="templateItems" stripe class="detail-table" row-class-name="detail-row">
          <el-table-column label="字段" min-width="200">
            <template #default="{ row }">
              <div class="field-cell">
                <div class="field-type-icon small" :class="getFieldTypeClass(row.field?.field_type || '')">
                  <el-icon><component :is="getFieldTypeIcon(row.field?.field_type || '')" /></el-icon>
                </div>
                <div class="field-card-info">
                  <div class="field-card-name">{{ row.field?.field_name }}</div>
                  <div class="field-card-key">{{ row.field?.field_key }}</div>
                </div>
              </div>
            </template>
          </el-table-column>
          <el-table-column label="必填" width="70" align="center">
            <template #default="{ row }">
              <el-switch v-model="row.is_required" size="small" @change="markTemplateItemsDirty" />
            </template>
          </el-table-column>
          <el-table-column label="创建" width="70" align="center">
            <template #default="{ row }">
              <el-switch v-model="row.is_visible_create" size="small" @change="markTemplateItemsDirty" />
            </template>
          </el-table-column>
          <el-table-column label="编辑" width="70" align="center">
            <template #default="{ row }">
              <el-switch v-model="row.is_visible_edit" size="small" @change="markTemplateItemsDirty" />
            </template>
          </el-table-column>
          <el-table-column label="详情" width="70" align="center">
            <template #default="{ row }">
              <el-switch v-model="row.is_visible_detail" size="small" @change="markTemplateItemsDirty" />
            </template>
          </el-table-column>
          <el-table-column label="排序" width="90" align="center">
            <template #default="{ row }">
              <el-input
                v-model.number="row.sort_order"
                type="number"
                size="small"
                class="sort-input"
                @change="markTemplateItemsDirty"
              />
            </template>
          </el-table-column>
          <el-table-column label="" width="60" align="center">
            <template #default="{ $index }">
              <el-button size="small" type="danger" text circle @click="removeTemplateItem($index)">
                <el-icon><Close /></el-icon>
              </el-button>
            </template>
          </el-table-column>
        </el-table>
        <TdEmptyState v-if="templateItems.length === 0" preset="no-data" title="暂无字段，点击上方按钮添加" />
      </div>
      <template #footer>
        <el-button @click="templateDetailVisible = false">取消</el-button>
        <el-button type="primary" :loading="templateItemsSaving" :disabled="!templateItemsDirty" @click="saveTemplateItems">
          保存配置
        </el-button>
      </template>
    </el-dialog>

    <!-- 添加模板字段选择对话框 -->
    <el-dialog
      v-model="addTemplateFieldVisible"
      title="选择要添加的字段"
      width="620px"
      destroy-on-close
      class="custom-dialog"
    >
      <el-input
        v-model="templateFieldSearch"
        placeholder="搜索字段..."
        :prefix-icon="Search"
        clearable
        style="margin-bottom: 16px"
      />
      <el-table
        :data="availableFieldsForTemplate"
        max-height="400"
        class="select-field-table"
        @selection-change="handleTemplateFieldSelectionChange"
      >
        <el-table-column type="selection" width="46" />
        <el-table-column label="字段" min-width="200">
          <template #default="{ row }">
            <div class="field-cell">
              <div class="field-type-icon small" :class="getFieldTypeClass(row.field_type)">
                <el-icon><component :is="getFieldTypeIcon(row.field_type)" /></el-icon>
              </div>
              <div class="field-card-info">
                <div class="field-card-name">{{ row.field_name }}</div>
                <div class="field-card-key">{{ row.field_key }}</div>
              </div>
            </div>
          </template>
        </el-table-column>
        <el-table-column label="类型" width="110">
          <template #default="{ row }">
            <div class="field-type-tag inline" :class="'tag-' + getFieldTypeClass(row.field_type)">
              {{ getFieldTypeLabel(row.field_type) }}
            </div>
          </template>
        </el-table-column>
        <el-table-column label="来源" width="90" align="center">
          <template #default="{ row }">
            <el-tag size="small" :type="row.is_system ? 'info' : 'success'" effect="plain" round>
              {{ row.is_system ? '系统' : '自定义' }}
            </el-tag>
          </template>
        </el-table-column>
      </el-table>
      <template #footer>
        <el-button @click="addTemplateFieldVisible = false">取消</el-button>
        <el-button type="primary" :disabled="selectedFieldsForTemplate.length === 0" @click="confirmAddTemplateFields">
          添加选中 ({{ selectedFieldsForTemplate.length }})
        </el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage, ElMessageBox, type FormInstance, type FormRules } from 'element-plus'
import {
  Grid, Plus, Search, ArrowDown, Close, MoreFilled,
  Document, EditPen, List, Edit, Delete, Setting,
  Tickets, Calendar, User, CopyDocument,
} from '@element-plus/icons-vue'
import {
  getGlobalFields,
  createGlobalField,
  updateGlobalField,
  deleteGlobalField,
  getFieldUsage,
  getTemplates,
  createTemplate,
  getTemplate,
  updateTemplate,
  deleteTemplate,
  updateTemplateItems,
} from '@/api/admin-field'
import type {
  FieldDefinition,
  FieldSchemeTemplate,
  FieldSchemeTemplateItem,
  FieldUsage,
  TemplateItemInput,
} from '@/types/field'
import { getFieldTypeLabel, FieldType } from '@/types/field'

// ============ 路由 & Tab 持久化 ============

const route = useRoute()
const router = useRouter()

const validTabs = ['fields', 'templates']
const initialTab = validTabs.includes(route.query.tab as string) ? (route.query.tab as string) : 'fields'
const activeTab = ref(initialTab)

watch(activeTab, (tab) => {
  router.replace({ query: { ...route.query, tab } })
})

// ============ 状态 ============

const systemFieldsCollapsed = ref(false)

// 全局字段
const fieldsLoading = ref(false)
const allFields = ref<FieldDefinition[]>([])
const fieldSearchText = ref('')

// 模板
const templatesLoading = ref(false)
const templates = ref<FieldSchemeTemplate[]>([])

// ============ 计算属性 ============

const filteredSystemFields = computed(() => {
  const search = fieldSearchText.value.toLowerCase()
  return allFields.value
    .filter(f => f.is_system)
    .filter(f => !search || f.field_name.toLowerCase().includes(search) || f.field_key.toLowerCase().includes(search))
    .sort((a, b) => a.sort_order - b.sort_order)
})

const filteredCustomFields = computed(() => {
  const search = fieldSearchText.value.toLowerCase()
  return allFields.value
    .filter(f => !f.is_system)
    .filter(f => !search || f.field_name.toLowerCase().includes(search) || f.field_key.toLowerCase().includes(search))
    .sort((a, b) => a.sort_order - b.sort_order)
})

// ============ 加载数据 ============

const loadFields = async () => {
  fieldsLoading.value = true
  try {
    const { data } = await getGlobalFields()
    allFields.value = data.data || []
  } catch {
    ElMessage.error('加载字段列表失败')
  } finally {
    fieldsLoading.value = false
  }
}

const loadTemplates = async () => {
  templatesLoading.value = true
  try {
    const { data } = await getTemplates()
    templates.value = data.data || []
  } catch {
    ElMessage.error('加载模板列表失败')
  } finally {
    templatesLoading.value = false
  }
}

// ============ 字段 CRUD ============

const fieldDialogVisible = ref(false)
const isEditFieldMode = ref(false)
const editingFieldId = ref<number | null>(null)
const fieldSubmitLoading = ref(false)
const fieldFormRef = ref<FormInstance>()
const fieldForm = ref({
  field_key: '',
  field_name: '',
  field_type: '' as string,
  description: '',
  options: '',
  default_value: '',
})
// 可视化选项编辑列表
const editingOptions = ref<{ value: string; label: string }[]>([])

// 将选项列表序列化为 JSON（提交时使用）
function serializeOptions(): string {
  const valid = editingOptions.value
    .filter(o => o.label.trim())
    .map(o => ({
      value: o.value.trim() || o.label.trim(),
      label: o.label.trim(),
    }))
  return valid.length > 0 ? JSON.stringify(valid) : ''
}

// 从 JSON 解析选项列表（编辑时使用）
function deserializeOptions(json: string): { value: string; label: string }[] {
  if (!json) return []
  try {
    const parsed = JSON.parse(json)
    if (!Array.isArray(parsed)) return []
    return parsed.map((item: any) => {
      if (typeof item === 'string') return { value: item, label: item }
      return {
        value: String(item.value ?? item.label ?? ''),
        label: String(item.label ?? item.value ?? ''),
      }
    })
  } catch {
    return []
  }
}

const fieldFormRules: FormRules = {
  field_key: [
    { required: true, message: '请输入字段标识', trigger: 'blur' },
    { pattern: /^[a-zA-Z][a-zA-Z0-9_]*$/, message: '只能包含英文字母、数字、下划线，且以字母开头', trigger: 'blur' },
  ],
  field_name: [{ required: true, message: '请输入字段名称', trigger: 'blur' }],
  field_type: [{ required: true, message: '请选择字段类型', trigger: 'change' }],
}

// 字段类型下拉选项 (与 FieldType 常量同源, 避免类型遗漏)
const fieldTypeOptions: Record<string, string> = {
  [FieldType.TEXT]: getFieldTypeLabel(FieldType.TEXT),
  [FieldType.TEXTAREA]: getFieldTypeLabel(FieldType.TEXTAREA),
  [FieldType.NUMBER]: getFieldTypeLabel(FieldType.NUMBER),
  [FieldType.DATE]: getFieldTypeLabel(FieldType.DATE),
  [FieldType.DATETIME]: getFieldTypeLabel(FieldType.DATETIME),
  [FieldType.SELECT]: getFieldTypeLabel(FieldType.SELECT),
  [FieldType.MULTISELECT]: getFieldTypeLabel(FieldType.MULTISELECT),
  [FieldType.USER]: getFieldTypeLabel(FieldType.USER),
  [FieldType.MULTIUSER]: getFieldTypeLabel(FieldType.MULTIUSER),
  [FieldType.LABEL]: getFieldTypeLabel(FieldType.LABEL),
  [FieldType.URL]: getFieldTypeLabel(FieldType.URL),
  [FieldType.CHECKBOX]: getFieldTypeLabel(FieldType.CHECKBOX),
}

const openCreateFieldDialog = () => {
  isEditFieldMode.value = false
  editingFieldId.value = null
  fieldForm.value = {
    field_key: '',
    field_name: '',
    field_type: '',
    description: '',
    options: '',
    default_value: '',
  }
  editingOptions.value = []
  fieldDialogVisible.value = true
}

const handleEditField = (field: FieldDefinition) => {
  isEditFieldMode.value = true
  editingFieldId.value = field.id
  fieldForm.value = {
    field_key: field.field_key,
    field_name: field.field_name,
    field_type: field.field_type,
    description: field.description || '',
    options: field.options || '',
    default_value: field.default_value || '',
  }
  editingOptions.value = deserializeOptions(field.options)
  fieldDialogVisible.value = true
}

const submitFieldForm = async () => {
  if (!fieldFormRef.value) return
  await fieldFormRef.value.validate(async (valid) => {
    if (!valid) return
    fieldSubmitLoading.value = true
    try {
      // select/multiselect 类型的选项从可视化编辑器序列化
      const optionsJson = (fieldForm.value.field_type === 'select' || fieldForm.value.field_type === 'multiselect')
        ? serializeOptions()
        : fieldForm.value.options

      if (isEditFieldMode.value && editingFieldId.value) {
        await updateGlobalField(editingFieldId.value, {
          field_name: fieldForm.value.field_name,
          description: fieldForm.value.description,
          options: optionsJson,
          default_value: fieldForm.value.default_value,
        })
        ElMessage.success('更新成功')
      } else {
        await createGlobalField({
          field_key: fieldForm.value.field_key,
          field_name: fieldForm.value.field_name,
          field_type: fieldForm.value.field_type as any,
          description: fieldForm.value.description,
          options: optionsJson,
          default_value: fieldForm.value.default_value,
        })
        ElMessage.success('创建成功')
      }
      fieldDialogVisible.value = false
      await loadFields()
    } catch {
      ElMessage.error(isEditFieldMode.value ? '更新字段失败' : '创建字段失败')
    } finally {
      fieldSubmitLoading.value = false
    }
  })
}

const handleToggleFieldActive = async (field: FieldDefinition) => {
  try {
    await updateGlobalField(field.id, { is_active: field.is_active })
  } catch {
    field.is_active = !field.is_active
    ElMessage.error('更新状态失败')
  }
}

const handleUpdateFieldSort = async (field: FieldDefinition) => {
  try {
    await updateGlobalField(field.id, { sort_order: field.sort_order })
  } catch {
    ElMessage.error('更新排序失败')
    await loadFields()
  }
}

const handleDeleteField = async (field: FieldDefinition) => {
  try {
    const { data } = await getFieldUsage(field.id)
    const usage: FieldUsage = data.data

    let warningMsg = `确定要删除字段 "${field.field_name}" 吗？`
    if (usage.scheme_count > 0 || usage.value_count > 0 || usage.template_count > 0) {
      warningMsg = `字段 "${field.field_name}" 正在被使用：\n` +
        `- ${usage.scheme_count} 个字段方案引用\n` +
        `- ${usage.value_count} 条字段值记录\n` +
        `- ${usage.template_count} 个模板引用\n\n` +
        `删除后以上关联数据将一并清除，确定继续？`
    }

    await ElMessageBox.confirm(warningMsg, '删除确认', {
      type: 'warning',
      confirmButtonText: '确认删除',
      cancelButtonText: '取消',
    })

    await deleteGlobalField(field.id)
    ElMessage.success('删除成功')
    await loadFields()
  } catch (error) {
    if (error !== 'cancel') {
      ElMessage.error('删除字段失败')
    }
  }
}

// ============ 模板 CRUD ============

const templateDialogVisible = ref(false)
const isEditTemplateMode = ref(false)
const editingTemplateId = ref<number | null>(null)
const templateSubmitLoading = ref(false)
const templateFormRef = ref<FormInstance>()
const templateForm = ref({
  name: '',
  description: '',
  is_active: true,
})
const templateFormRules: FormRules = {
  name: [{ required: true, message: '请输入模板名称', trigger: 'blur' }],
}

const openCreateTemplateDialog = () => {
  isEditTemplateMode.value = false
  editingTemplateId.value = null
  templateForm.value = { name: '', description: '', is_active: true }
  templateDialogVisible.value = true
}

const handleEditTemplate = (tpl: FieldSchemeTemplate) => {
  isEditTemplateMode.value = true
  editingTemplateId.value = tpl.id
  templateForm.value = {
    name: tpl.name,
    description: tpl.description || '',
    is_active: tpl.is_active,
  }
  templateDialogVisible.value = true
}

const submitTemplateForm = async () => {
  if (!templateFormRef.value) return
  await templateFormRef.value.validate(async (valid) => {
    if (!valid) return
    templateSubmitLoading.value = true
    try {
      if (isEditTemplateMode.value && editingTemplateId.value) {
        await updateTemplate(editingTemplateId.value, {
          name: templateForm.value.name,
          description: templateForm.value.description,
          is_active: templateForm.value.is_active,
        })
        ElMessage.success('更新成功')
      } else {
        await createTemplate({
          name: templateForm.value.name,
          description: templateForm.value.description,
        })
        ElMessage.success('创建成功')
      }
      templateDialogVisible.value = false
      await loadTemplates()
    } catch {
      ElMessage.error(isEditTemplateMode.value ? '更新模板失败' : '创建模板失败')
    } finally {
      templateSubmitLoading.value = false
    }
  })
}

const handleDeleteTemplate = async (tpl: FieldSchemeTemplate) => {
  try {
    await ElMessageBox.confirm(`确定要删除模板 "${tpl.name}" 吗？`, '删除确认', {
      type: 'warning',
    })
    await deleteTemplate(tpl.id)
    ElMessage.success('删除成功')
    await loadTemplates()
  } catch (error) {
    if (error !== 'cancel') {
      ElMessage.error('删除模板失败')
    }
  }
}

// ============ 模板字段配置 ============

const templateDetailVisible = ref(false)
const templateDetailLoading = ref(false)
const currentTemplate = ref<FieldSchemeTemplate | null>(null)
const templateItems = ref<FieldSchemeTemplateItem[]>([])
const templateItemsDirty = ref(false)
const templateItemsSaving = ref(false)

const openTemplateDetail = async (tpl: FieldSchemeTemplate) => {
  currentTemplate.value = tpl
  templateDetailVisible.value = true
  templateDetailLoading.value = true
  templateItemsDirty.value = false
  try {
    const { data } = await getTemplate(tpl.id)
    templateItems.value = data.data?.items || []
  } catch {
    ElMessage.error('加载模板详情失败')
  } finally {
    templateDetailLoading.value = false
  }
}

const markTemplateItemsDirty = () => {
  templateItemsDirty.value = true
}

const removeTemplateItem = (index: number) => {
  templateItems.value.splice(index, 1)
  templateItemsDirty.value = true
}

const saveTemplateItems = async () => {
  if (!currentTemplate.value) return
  templateItemsSaving.value = true
  try {
    const items: TemplateItemInput[] = templateItems.value.map((item, idx) => ({
      field_id: item.field_id,
      is_required: item.is_required,
      is_visible_create: item.is_visible_create,
      is_visible_edit: item.is_visible_edit,
      is_visible_detail: item.is_visible_detail,
      sort_order: item.sort_order ?? idx,
      default_value: item.default_value || '',
    }))
    await updateTemplateItems(currentTemplate.value.id, items)
    ElMessage.success('保存成功')
    templateItemsDirty.value = false
    await loadTemplates()
  } catch {
    ElMessage.error('保存模板配置失败')
  } finally {
    templateItemsSaving.value = false
  }
}

// 添加字段到模板
const addTemplateFieldVisible = ref(false)
const templateFieldSearch = ref('')
const selectedFieldsForTemplate = ref<FieldDefinition[]>([])

const availableFieldsForTemplate = computed(() => {
  const existingFieldIds = new Set(templateItems.value.map(i => i.field_id))
  const search = templateFieldSearch.value.toLowerCase()
  return allFields.value
    .filter(f => f.is_active && !existingFieldIds.has(f.id))
    .filter(f => !search || f.field_name.toLowerCase().includes(search) || f.field_key.toLowerCase().includes(search))
})

const openAddTemplateFieldDialog = () => {
  templateFieldSearch.value = ''
  selectedFieldsForTemplate.value = []
  addTemplateFieldVisible.value = true
  if (allFields.value.length === 0) {
    loadFields()
  }
}

const handleTemplateFieldSelectionChange = (selection: FieldDefinition[]) => {
  selectedFieldsForTemplate.value = selection
}

const confirmAddTemplateFields = () => {
  const maxSort = templateItems.value.reduce((max, item) => Math.max(max, item.sort_order || 0), 0)
  for (let i = 0; i < selectedFieldsForTemplate.value.length; i++) {
    const field = selectedFieldsForTemplate.value[i]
    templateItems.value.push({
      id: 0,
      template_id: currentTemplate.value?.id || 0,
      field_id: field.id,
      field: field,
      is_required: false,
      is_visible_create: true,
      is_visible_edit: true,
      is_visible_detail: true,
      sort_order: maxSort + i + 1,
      default_value: '',
    })
  }
  templateItemsDirty.value = true
  addTemplateFieldVisible.value = false
}

// ============ 字段图标/样式工具 ============

const getFieldTypeClass = (fieldType: string) => {
  const classMap: Record<string, string> = {
    text: 'type-text',
    textarea: 'type-textarea',
    number: 'type-number',
    date: 'type-date',
    select: 'type-select',
    multiselect: 'type-multiselect',
    user: 'type-user',
    version: 'type-version',
    component: 'type-component',
    label: 'type-label',
    epic_link: 'type-epic',
    time_estimate: 'type-time',
  }
  return classMap[fieldType] || 'type-text'
}

const getFieldTypeIcon = (fieldType: string) => {
  const iconMap: Record<string, any> = {
    text: Document,
    textarea: EditPen,
    number: Tickets,
    date: Calendar,
    select: List,
    multiselect: List,
    user: User,
    version: Tickets,
    component: Grid,
    label: Tickets,
    epic_link: Tickets,
    time_estimate: Calendar,
  }
  return iconMap[fieldType] || Document
}

// ============ 初始化 ============

onMounted(async () => {
  await Promise.all([loadFields(), loadTemplates()])
})
</script>

<style scoped lang="scss">
.field-management-container {
  width: 100%;
}

// ============ 页面头部 icon (TdPageHeader leading slot) ============
.page-header-icon {
  width: 40px;
  height: 40px;
  background: var(--td-tag-primary-bg);
  border-radius: var(--td-radius-md);
  display: flex;
  align-items: center;
  justify-content: center;
  color: var(--td-color-primary);
  flex-shrink: 0;
}

// ============ 内容卡片 ============
.content-card {
  border: none;
  box-shadow: var(--td-elevation-1);
  transition: var(--td-transition-shadow);
  border-radius: var(--td-radius-lg);

  &:hover { box-shadow: var(--td-elevation-2); }

  :deep(.el-card__body) {
    padding: 0;
  }
}

.custom-tabs {
  :deep(.el-tabs__header) {
    padding: 0 28px;
    margin-bottom: 0;
    border-bottom: 1px solid var(--td-divider-color);
  }

  :deep(.el-tabs__content) {
    padding: 24px 28px 28px;
  }

  :deep(.el-tabs__item) {
    height: 56px;
    line-height: 56px;
    font-size: 14px;
  }

  :deep(.el-tabs__nav-wrap::after) {
    height: 1px;
  }
}

.tab-label {
  display: flex;
  align-items: center;
  gap: 6px;
}

// ============ Tab 头部 ============
.tab-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 24px;
}

.tab-header-hint {
  font-size: 13px;
  color: var(--td-text-placeholder);
}

.search-input {
  width: 280px;

  :deep(.el-input__wrapper) {
    border-radius: 10px;
    box-shadow: 0 0 0 1px var(--td-border-color) inset;

    &:hover, &.is-focus {
      box-shadow: 0 0 0 1px var(--td-color-primary) inset;
    }
  }
}

.action-btn {
  border-radius: 10px;
  padding: 10px 20px;
  font-weight: 500;
}

// ============ 字段分区 ============
.field-section {
  margin-bottom: 28px;

  &:last-child {
    margin-bottom: 0;
  }
}

.section-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 12px 16px;
  background: var(--td-bg-page);
  border-radius: 10px;
  margin-bottom: 12px;
  cursor: pointer;
  transition: background 150ms ease-out;
  user-select: none;

  &:hover {
    background: var(--td-bg-section);
  }
}

.section-title-group {
  display: flex;
  align-items: center;
  gap: 10px;
}

.collapse-arrow {
  font-size: 14px;
  color: var(--td-text-secondary);
  transition: transform 150ms ease-out;

  &.collapsed {
    transform: rotate(-90deg);
  }
}

.section-title {
  font-size: 15px;
  font-weight: 600;
  color: var(--td-text-primary);
}

.section-count {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  min-width: 22px;
  height: 22px;
  padding: 0 7px;
  font-size: 12px;
  font-weight: 600;
  color: var(--td-text-secondary);
  background: var(--td-border-color);
  border-radius: 11px;
}

.section-hint {
  font-size: 12px;
  color: var(--td-text-placeholder);
}

// ============ 字段卡片列表 ============
.field-card-list {
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.field-card {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 14px 18px;
  background: var(--td-bg-card);
  border: 1px solid var(--td-divider-color);
  border-radius: 10px;
  transition: all 150ms ease-out;

  &:hover {
    border-color: var(--td-tag-primary-border);
    background: var(--td-bg-section);
    box-shadow: 0 2px 8px rgba(59, 130, 246, 0.06);
  }
}

.field-card-left {
  display: flex;
  align-items: center;
  gap: 12px;
  flex: 1;
  min-width: 0;
}

.field-card-right {
  display: flex;
  align-items: center;
  gap: 8px;
  flex-shrink: 0;
}

.field-card-action {
  display: flex;
  align-items: center;
  gap: 6px;
}

.action-label {
  font-size: 12px;
  color: var(--td-text-placeholder);
  white-space: nowrap;
}

// ============ 字段类型图标 ============
.field-type-icon {
  width: 38px;
  height: 38px;
  border-radius: 10px;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 18px;
  color: var(--td-text-white);
  flex-shrink: 0;

  &.small {
    width: 32px;
    height: 32px;
    font-size: 15px;
    border-radius: 8px;
  }

  &.type-text      { background: var(--td-color-primary); }
  &.type-textarea   { background: #6366f1; }
  &.type-number     { background: var(--td-color-warning); }
  &.type-date       { background: var(--td-color-success); }
  &.type-select     { background: #8b5cf6; }
  &.type-multiselect { background: #a78bfa; }
  &.type-user       { background: #ec4899; }
  &.type-version    { background: #06b6d4; }
  &.type-component  { background: #14b8a6; }
  &.type-label      { background: #f97316; }
  &.type-epic       { background: #6366f1; }
  &.type-time       { background: #64748b; }
}

// ============ 字段类型标签 ============
.field-type-tag {
  display: inline-flex;
  align-items: center;
  padding: 2px 10px;
  border-radius: 6px;
  font-size: 12px;
  font-weight: 500;
  white-space: nowrap;
  flex-shrink: 0;

  &.inline {
    padding: 3px 10px;
  }

  &.tag-type-text       { background: var(--td-tag-primary-border); color: var(--td-color-primary-hover); }
  &.tag-type-textarea   { background: var(--td-tag-indigo-bg); color: var(--td-tag-indigo-text); }
  &.tag-type-number     { background: var(--td-tag-warning-bg); color: var(--td-tag-orange-text); }
  &.tag-type-date       { background: var(--td-tag-success-bg); color: var(--td-tag-success-text); }
  &.tag-type-select     { background: var(--td-tag-purple-bg); color: var(--td-tag-purple-text); }
  &.tag-type-multiselect { background: var(--td-tag-purple-bg); color: var(--td-tag-purple-text); }
  &.tag-type-user       { background: var(--td-tag-pink-bg); color: var(--td-tag-pink-text); }
  &.tag-type-version    { background: var(--td-tag-cyan-bg); color: var(--td-tag-cyan-text); }
  &.tag-type-component  { background: var(--td-tag-teal-bg); color: var(--td-tag-teal-text); }
  &.tag-type-label      { background: var(--td-tag-orange-bg); color: var(--td-tag-orange-text); }
  &.tag-type-epic       { background: var(--td-tag-indigo-bg); color: var(--td-tag-indigo-text); }
  &.tag-type-time       { background: var(--td-bg-section); color: var(--td-text-regular); }
}

// ============ 字段信息 ============
.field-card-info {
  min-width: 0;

  .field-card-name {
    font-size: 14px;
    font-weight: 500;
    color: var(--td-text-primary);
    line-height: 1.4;
  }

  .field-card-key {
    font-size: 12px;
    color: var(--td-text-placeholder);
    font-family: 'SF Mono', 'Menlo', 'Monaco', monospace;
    line-height: 1.3;
  }
}

.field-card-desc {
  font-size: 12px;
  color: var(--td-text-placeholder);
  max-width: 240px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.sort-input {
  width: 64px;

  :deep(.el-input__wrapper) {
    padding: 0 8px;
  }

  // 隐藏 number 输入框的上下箭头
  :deep(input[type="number"]) {
    -moz-appearance: textfield;
    text-align: center;

    &::-webkit-outer-spin-button,
    &::-webkit-inner-spin-button {
      -webkit-appearance: none;
      margin: 0;
    }
  }
}

.empty-custom {
  padding: 24px 0;
}

// ============ 模板卡片 ============
.template-card-list {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(360px, 1fr));
  gap: 16px;
}

.template-card {
  background: var(--td-bg-card);
  border: 1px solid var(--td-border-color);
  border-radius: 12px;
  padding: 20px;
  cursor: pointer;
  transition: all 150ms ease-out;

  &:hover {
    border-color: var(--td-tag-primary-border);
    box-shadow: 0 4px 16px rgba(59, 130, 246, 0.1);
  }
}

.template-card-header {
  display: flex;
  align-items: flex-start;
  gap: 14px;
  margin-bottom: 16px;
}

.template-icon {
  width: 44px;
  height: 44px;
  background: var(--td-tag-purple-bg);
  color: var(--td-tag-purple-text);
  border-radius: 12px;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 22px;
  flex-shrink: 0;
}

.template-meta {
  flex: 1;
  min-width: 0;
}

.template-name {
  font-size: 15px;
  font-weight: 600;
  color: var(--td-text-primary);
  margin-bottom: 4px;
  line-height: 1.3;
}

.template-desc {
  font-size: 13px;
  color: var(--td-text-placeholder);
  line-height: 1.4;
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
  overflow: hidden;
}

.template-status {
  flex-shrink: 0;
  margin-top: 2px;
}

.template-card-footer {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding-top: 14px;
  border-top: 1px solid var(--td-border-color);
}

.template-stat {
  display: flex;
  align-items: center;
  gap: 6px;
  font-size: 13px;
  color: var(--td-text-secondary);

  .el-icon {
    font-size: 14px;
    color: var(--td-text-placeholder);
  }
}

.template-actions {
  display: flex;
  align-items: center;
  gap: 8px;
}

.more-btn {
  padding: 6px;
  min-width: auto;

  :deep(.el-icon) {
    margin: 0;
  }
}

.danger-item {
  color: var(--td-color-danger) !important;

  .el-icon {
    color: var(--td-color-danger);
  }
}

// ============ 模板详情弹窗 ============
.template-detail-content {
  min-height: 200px;
}

.template-detail-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 16px;
}

.template-detail-info {
  display: flex;
  align-items: center;
  gap: 10px;
}

.detail-field-count {
  font-size: 13px;
  color: var(--td-text-secondary);
  font-weight: 500;
}

.detail-table {
  :deep(.el-table__header th) {
    background: var(--td-bg-page);
    font-weight: 600;
    font-size: 13px;
    color: var(--td-text-secondary);
  }
}

.field-cell {
  display: flex;
  align-items: center;
  gap: 10px;
}

// ============ 折叠动画 ============
.collapse-enter-active,
.collapse-leave-active {
  transition: all 150ms ease-out;
  overflow: hidden;
}
.collapse-enter-from,
.collapse-leave-to {
  opacity: 0;
  max-height: 0;
}
.collapse-enter-to,
.collapse-leave-from {
  opacity: 1;
  max-height: 2000px;
}

// ============ 对话框美化 ============
.custom-dialog {
  :deep(.el-dialog) {
    border-radius: 12px;
  }

  :deep(.el-dialog__header) {
    padding: 20px 24px 16px;
    border-bottom: 1px solid var(--td-divider-color);
    margin-right: 0;
  }

  :deep(.el-dialog__body) {
    padding: 20px 24px;
  }

  :deep(.el-dialog__footer) {
    padding: 16px 24px 20px;
    border-top: 1px solid var(--td-divider-color);
  }
}

.options-editor {
  width: 100%;
}

.option-row {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 8px;
}

.option-input {
  flex: 1;
}
</style>
