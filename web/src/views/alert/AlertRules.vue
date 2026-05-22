<template>
  <div class="alert-rules-container">
    <!-- 页面头部 -->
    <div class="page-header">
      <div class="header-info">
        <div class="header-icon">
          <el-icon><DocumentChecked /></el-icon>
        </div>
        <div class="header-text">
          <h1 class="header-title">告警规则</h1>
          <p class="header-desc">配置告警自动建单规则和匹配策略</p>
        </div>
      </div>
      <el-button type="primary" class="header-btn" @click="handleCreate">
        <el-icon><Plus /></el-icon>
        创建规则
      </el-button>
    </div>

    <!-- 表格卡片 -->
    <el-card v-loading="loading" shadow="never" class="table-card">
      <el-table :data="ruleList" style="width: 100%">
        <el-table-column prop="name" label="规则名称" min-width="150" />
        <el-table-column prop="project_name" label="项目" width="120">
          <template #default="{ row }">
            <el-tag size="small" effect="plain" type="info">{{ row.project_name }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="datasource_name" label="数据源" width="120">
          <template #default="{ row }">
            <span class="type-text">{{ row.datasource_name || '全部' }}</span>
          </template>
        </el-table-column>
        <el-table-column prop="issue_type_name" label="工单类型" width="100">
          <template #default="{ row }">
            <span class="type-text">{{ row.issue_type_name }}</span>
          </template>
        </el-table-column>
        <el-table-column label="告警等级" width="90" align="center">
          <template #default="{ row }">
            <el-tag
              v-if="getSeverityFromMatchers(row.label_matchers)"
              :type="getSeverityTagType(getSeverityFromMatchers(row.label_matchers))"
              size="small"
              effect="dark"
            >
              {{ getSeverityLabel(getSeverityFromMatchers(row.label_matchers)) }}
            </el-tag>
            <span v-else class="type-text">全部</span>
          </template>
        </el-table-column>
        <el-table-column label="额外匹配" min-width="180">
          <template #default="{ row }">
            <template v-for="(matcher, index) in getExtraMatchers(row.label_matchers)" :key="index">
              <el-tag v-if="index < 2" size="small" style="margin-right: 4px">
                {{ matcher.key }} {{ matcher.operator }} {{ matcher.value }}
              </el-tag>
            </template>
            <el-tag v-if="getExtraMatchers(row.label_matchers).length > 2" size="small">
              +{{ getExtraMatchers(row.label_matchers).length - 2 }}
            </el-tag>
            <span v-if="getExtraMatchers(row.label_matchers).length === 0" class="type-text">-</span>
          </template>
        </el-table-column>
        <el-table-column prop="priority" label="优先级" width="80" align="center">
          <template #default="{ row }">
            <div class="priority-badge" :class="row.priority">
              <span class="priority-dot"></span>
              <span>{{ row.priority }}</span>
            </div>
          </template>
        </el-table-column>
        <el-table-column label="合并窗口" width="100">
          <template #default="{ row }">
            {{ formatMergeWindow(row.merge_window) }}
          </template>
        </el-table-column>
        <el-table-column label="自动解决" width="80" align="center">
          <template #default="{ row }">
            <div class="status-badge" :class="row.auto_resolve ? 'enabled' : 'disabled'">
              <span class="status-dot"></span>
              <span>{{ row.auto_resolve ? '是' : '否' }}</span>
            </div>
          </template>
        </el-table-column>
        <el-table-column label="状态" width="80" align="center">
          <template #default="{ row }">
            <div class="status-badge" :class="row.status === 1 ? 'enabled' : 'disabled'">
              <span class="status-dot"></span>
              <span>{{ row.status === 1 ? '启用' : '禁用' }}</span>
            </div>
          </template>
        </el-table-column>
        <el-table-column label="操作" width="120" fixed="right" align="center">
          <template #default="{ row }">
            <el-tooltip content="编辑" placement="top">
              <el-button link type="primary" :icon="Edit" @click="handleEdit(row)" />
            </el-tooltip>
            <el-tooltip content="删除" placement="top">
              <el-button link type="danger" :icon="Delete" @click="handleDelete(row)" />
            </el-tooltip>
          </template>
        </el-table-column>
      </el-table>

      <div class="pagination-wrapper">
        <el-pagination
          v-model:current-page="queryParams.page"
          v-model:page-size="queryParams.page_size"
          :total="total"
          layout="total, prev, pager, next"
          @current-change="loadData"
        />
      </div>
    </el-card>

    <!-- 创建/编辑对话框 -->
    <el-dialog
      v-model="dialogVisible"
      :title="dialogTitle"
      width="700px"
      @close="handleDialogClose"
    >
      <el-form ref="formRef" :model="form" :rules="rules" label-position="top">
        <el-row :gutter="16">
          <el-col :span="24">
            <el-form-item label="规则名称" prop="name">
              <el-input v-model="form.name" placeholder="请输入规则名称" />
            </el-form-item>
          </el-col>
        </el-row>
        <el-row :gutter="16">
          <el-col :span="24">
            <el-form-item label="描述" prop="description">
              <el-input v-model="form.description" type="textarea" :rows="2" placeholder="请输入描述" />
            </el-form-item>
          </el-col>
        </el-row>
        <el-row :gutter="16">
          <el-col :span="12">
            <el-form-item label="数据源" prop="datasource_id">
              <el-select v-model="form.datasource_id" placeholder="请选择数据源" style="width: 100%" filterable>
                <el-option
                  v-for="ds in datasourceList"
                  :key="ds.id"
                  :label="ds.name"
                  :value="ds.id"
                />
              </el-select>
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="项目" prop="project_id">
              <el-select v-model="form.project_id" placeholder="请选择项目" style="width: 100%" filterable @change="handleProjectChange">
                <el-option
                  v-for="p in projectList"
                  :key="p.id"
                  :label="`${p.project_key} - ${p.name}`"
                  :value="p.id"
                />
              </el-select>
            </el-form-item>
          </el-col>
        </el-row>
        <el-row :gutter="16">
          <el-col :span="12">
            <el-form-item label="工单类型" prop="issue_type_id">
              <el-select v-model="form.issue_type_id" placeholder="请先选择项目" style="width: 100%" :disabled="!form.project_id" filterable>
                <el-option
                  v-for="t in issueTypeList"
                  :key="t.id"
                  :label="t.display_name || t.name"
                  :value="t.id"
                />
              </el-select>
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="告警等级" prop="severity">
              <el-select v-model="form.severity" placeholder="全部等级（不限）" style="width: 100%" clearable>
                <el-option label="严重 (critical)" value="critical" />
                <el-option label="警告 (warning)" value="warning" />
                <el-option label="信息 (info)" value="info" />
              </el-select>
              <div class="form-tip">选择要匹配的告警等级，留空则匹配所有等级</div>
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="工单优先级" prop="priority">
              <el-select v-model="form.priority" placeholder="请选择优先级" style="width: 100%">
                <el-option label="P0 - 紧急" value="P0" />
                <el-option label="P1 - 高" value="P1" />
                <el-option label="P2 - 中" value="P2" />
                <el-option label="P3 - 低" value="P3" />
              </el-select>
              <div class="form-tip">匹配到的告警创建工单时使用的优先级</div>
            </el-form-item>
          </el-col>
        </el-row>
        <el-row :gutter="16">
          <el-col :span="24">
            <el-form-item label="额外标签匹配器（可选）" prop="label_matchers">
              <div class="form-tip" style="margin-bottom: 8px;">
                除告警等级外的额外过滤条件，所有条件需同时满足。常用标签：alertname（告警名称）、instance（实例）、group_name（业务组）。
                操作符：== 精确匹配、!= 不等于、=~ 正则匹配、!~ 正则排除。
              </div>
              <div class="matchers-container">
                <div
                  v-for="(matcher, index) in form.label_matchers"
                  :key="index"
                  class="matcher-item"
                >
                  <el-input v-model="matcher.key" placeholder="标签键" class="matcher-input" />
                  <el-select v-model="matcher.operator" class="matcher-operator">
                    <el-option label="==" value="==" />
                    <el-option label="!=" value="!=" />
                    <el-option label="=~" value="=~" />
                    <el-option label="!~" value="!~" />
                  </el-select>
                  <el-input v-model="matcher.value" placeholder="标签值" class="matcher-input" />
                  <el-button
                    type="danger"
                    :icon="Delete"
                    circle
                    @click="removeMatcher(index)"
                  />
                </div>
                <el-button type="primary" text @click="addMatcher">
                  <el-icon><Plus /></el-icon>
                  添加匹配器
                </el-button>
              </div>
            </el-form-item>
          </el-col>
        </el-row>
        <el-row :gutter="16">
          <el-col :span="12">
            <el-form-item label="合并窗口（秒）" prop="merge_window">
              <el-input-number
                v-model="form.merge_window"
                :min="0"
                :max="86400"
                :step="300"
                style="width: 100%"
              />
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="自动解决" prop="auto_resolve">
              <el-switch v-model="form.auto_resolve" />
              <div class="form-tip">
                开启：告警恢复时自动关闭工单；关闭：告警恢复时工单变为"待确认"，由处理人确认后关闭
              </div>
            </el-form-item>
          </el-col>
        </el-row>
        <div class="form-tip" style="margin-top: -8px; margin-bottom: 16px;">
          合并窗口：0 表示不合并，建议设置为 3600（1小时）
        </div>
      </el-form>
      <template #footer>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button type="primary" @click="handleSubmit">确定</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { ElMessage, ElMessageBox, type FormInstance, type FormRules } from 'element-plus'
import { Plus, Delete, Edit, DocumentChecked } from '@element-plus/icons-vue'
import {
  getAlertRuleList,
  createAlertRule,
  updateAlertRule,
  deleteAlertRule,
  getDatasourceList,
} from '@/api/alert'
import { getProjectList } from '@/api/project'
import type { AlertRule, LabelMatcher, AlertDatasource } from '@/types/alert'

const loading = ref(false)
const ruleList = ref<AlertRule[]>([])
const total = ref(0)
const queryParams = reactive({
  page: 1,
  page_size: 20,
})

// 项目、工单类型和数据源选项
const projectList = ref<{ id: number; name: string; project_key: string }[]>([])
const issueTypeList = ref<{ id: number; name: string; display_name: string }[]>([])
const datasourceList = ref<AlertDatasource[]>([])

const loadProjects = async () => {
  try {
    const { data } = await getProjectList({ page: 1, page_size: 100 })
    projectList.value = data.data.items || []
  } catch {
    // ignored
  }
}

const loadDatasources = async () => {
  try {
    const { data } = await getDatasourceList({ page: 1, page_size: 100 })
    datasourceList.value = data.data.items || []
  } catch {
    // ignored
  }
}

const loadIssueTypes = async (projectKey: string) => {
  try {
    const { getProjectIssueTypes } = await import('@/api/project')
    const { data } = await getProjectIssueTypes(projectKey)
    issueTypeList.value = data.data || []
  } catch {
    // ignored
  }
}

const handleProjectChange = (projectId: number) => {
  form.issue_type_id = undefined
  issueTypeList.value = []
  const project = projectList.value.find(p => p.id === projectId)
  if (project) {
    loadIssueTypes(project.project_key)
  }
}

const dialogVisible = ref(false)
const dialogTitle = ref('创建规则')
const formRef = ref<FormInstance>()
const form = reactive({
  id: 0,
  name: '',
  description: '',
  datasource_id: undefined as number | undefined,
  project_id: undefined as number | undefined,
  issue_type_id: undefined as number | undefined,
  severity: '' as string,
  label_matchers: [] as LabelMatcher[],
  priority: 'P2' as 'P0' | 'P1' | 'P2' | 'P3',
  assignee_id: undefined as number | undefined,
  auto_resolve: false,
  merge_window: 3600,
})

const rules: FormRules = {
  name: [{ required: true, message: '请输入规则名称', trigger: 'blur' }],
  datasource_id: [{ required: true, message: '请选择数据源', trigger: 'change' }],
  project_id: [{ required: true, message: '请选择项目', trigger: 'change' }],
  issue_type_id: [{ required: true, message: '请选择工单类型', trigger: 'change' }],
  priority: [{ required: true, message: '请选择优先级', trigger: 'change' }],
}

const loadData = async () => {
  loading.value = true
  try {
    const { data } = await getAlertRuleList(queryParams)
    ruleList.value = data.data.items
    total.value = data.data.total
  } catch {
    // ignored
  } finally {
    loading.value = false
  }
}

const handleCreate = () => {
  dialogTitle.value = '创建规则'
  form.id = 0
  form.name = ''
  form.description = ''
  form.datasource_id = undefined
  form.project_id = undefined
  form.issue_type_id = undefined
  form.severity = ''
  form.label_matchers = []
  form.priority = 'P2'
  form.assignee_id = undefined
  form.auto_resolve = false
  form.merge_window = 3600
  dialogVisible.value = true
}

const handleEdit = async (row: AlertRule) => {
  dialogTitle.value = '编辑规则'
  form.id = row.id
  form.name = row.name
  form.description = row.description
  form.datasource_id = row.datasource_id
  form.project_id = row.project_id
  form.issue_type_id = row.issue_type_id
  const matchers: LabelMatcher[] = JSON.parse(JSON.stringify(row.label_matchers))
  // 从 label_matchers 中提取 severity 到独立字段
  const severityIdx = matchers.findIndex(m => m.key === 'severity' && m.operator === '==')
  if (severityIdx >= 0) {
    form.severity = matchers[severityIdx].value
    matchers.splice(severityIdx, 1)
  } else {
    form.severity = ''
  }
  form.label_matchers = matchers
  form.priority = row.priority
  form.assignee_id = row.assignee_id
  form.auto_resolve = row.auto_resolve
  form.merge_window = row.merge_window
  // 加载该项目的工单类型
  const project = projectList.value.find(p => p.id === row.project_id)
  if (project) {
    await loadIssueTypes(project.project_key)
  }
  dialogVisible.value = true
}

const handleSubmit = async () => {
  if (!formRef.value) return
  await formRef.value.validate(async (valid) => {
    if (!valid) return

    // 合并 severity 到 label_matchers
    const matchers: LabelMatcher[] = [...form.label_matchers]
    if (form.severity) {
      matchers.unshift({ key: 'severity', operator: '==', value: form.severity })
    }

    const submitData = {
      ...form,
      label_matchers: matchers,
    }

    try {
      if (form.id) {
        await updateAlertRule(form.id, submitData)
        ElMessage.success('更新成功')
      } else {
        if (!form.project_id || !form.issue_type_id || !form.datasource_id) {
          ElMessage.error('请选择数据源、项目和工单类型')
          return
        }
        await createAlertRule({
          ...submitData,
          datasource_id: form.datasource_id!,
          project_id: form.project_id!,
          issue_type_id: form.issue_type_id!,
        })
        ElMessage.success('创建成功')
      }
      dialogVisible.value = false
      loadData()
    } catch {
      // ignored
    }
  })
}

const handleDelete = async (row: AlertRule) => {
  try {
    await ElMessageBox.confirm('确定要删除此规则吗？', '提示', {
      type: 'warning',
    })
    await deleteAlertRule(row.id)
    ElMessage.success('删除成功')
    loadData()
  } catch (error) {
    if (error !== 'cancel') {
      // ignored
    }
  }
}

const addMatcher = () => {
  form.label_matchers.push({ key: '', operator: '==', value: '' })
}

const removeMatcher = (index: number) => {
  form.label_matchers.splice(index, 1)
}

const handleDialogClose = () => {
  formRef.value?.resetFields()
}

const formatMergeWindow = (seconds: number) => {
  if (seconds === 0) return '不合并'
  if (seconds < 60) return `${seconds}秒`
  if (seconds < 3600) return `${Math.floor(seconds / 60)}分钟`
  return `${Math.floor(seconds / 3600)}小时`
}

// 从 label_matchers 中提取 severity 值
const getSeverityFromMatchers = (matchers: LabelMatcher[]) => {
  const m = matchers?.find(m => m.key === 'severity' && m.operator === '==')
  return m?.value || ''
}

// 获取除 severity 外的额外匹配器
const getExtraMatchers = (matchers: LabelMatcher[]) => {
  return (matchers || []).filter(m => !(m.key === 'severity' && m.operator === '=='))
}

// 告警等级标签
const getSeverityLabel = (severity: string) => {
  const map: Record<string, string> = { critical: '严重', warning: '警告', info: '信息' }
  return map[severity] || severity
}

type TagType = 'success' | 'warning' | 'info' | 'danger'

const getSeverityTagType = (severity: string): TagType => {
  const map: Record<string, TagType> = { critical: 'danger', warning: 'warning', info: 'info' }
  return map[severity] || 'info'
}

onMounted(() => {
  loadData()
  loadProjects()
  loadDatasources()
})
</script>

<style scoped lang="scss">
.alert-rules-container {
  width: 100%;
}

// 页面头部
.page-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 24px;
  padding: 24px 32px;
  background: var(--td-color-primary);
  border-radius: 12px;
  color: var(--td-text-white);

  .header-info {
    display: flex;
    align-items: center;
    gap: 16px;
  }

  .header-icon {
    width: 56px;
    height: 56px;
    background: rgba(255, 255, 255, 0.2);
    border-radius: 12px;
    display: flex;
    align-items: center;
    justify-content: center;
    font-size: 28px;
  }

  .header-text {
    .header-title { font-size: 22px; font-weight: 600; margin: 0 0 4px 0; }
    .header-desc { font-size: 14px; margin: 0; opacity: 0.9; }
  }

  .header-btn {
    background: rgba(255, 255, 255, 0.2);
    border: 1px solid rgba(255, 255, 255, 0.3);
    color: var(--td-text-white);
    &:hover { background: rgba(255, 255, 255, 0.3); }
  }
}

// 表格卡片
.table-card {
  border-radius: 12px;

  :deep(.el-card__body) { padding: 20px; }

  :deep(.el-table) {
    border-radius: 8px;

    th.el-table__cell {
      background: var(--td-bg-page);
      font-weight: 600;
      color: var(--td-text-regular);
    }
  }

  .type-text {
    font-size: 13px;
    color: var(--td-text-secondary);
  }

  .priority-badge {
    display: inline-flex;
    align-items: center;
    gap: 5px;
    padding: 3px 10px;
    border-radius: 12px;
    font-size: 12px;
    font-weight: 500;

    .priority-dot {
      width: 6px;
      height: 6px;
      border-radius: 50%;
    }

    &.P0 {
      background: var(--td-tag-danger-bg);
      color: var(--td-color-danger);
      .priority-dot { background: var(--td-color-danger); }
    }
    &.P1 {
      background: var(--td-tag-orange-bg);
      color: var(--td-tag-orange-text);
      .priority-dot { background: var(--td-color-warning); }
    }
    &.P2 {
      background: var(--td-tag-primary-bg);
      color: var(--td-color-primary-active);
      .priority-dot { background: var(--td-color-primary); }
    }
    &.P3 {
      background: var(--td-bg-section);
      color: var(--td-text-secondary);
      .priority-dot { background: var(--td-text-placeholder); }
    }
  }

  .status-badge {
    display: inline-flex;
    align-items: center;
    gap: 5px;
    padding: 3px 10px;
    border-radius: 12px;
    font-size: 12px;

    .status-dot {
      width: 6px;
      height: 6px;
      border-radius: 50%;
    }

    &.enabled {
      background: var(--td-tag-success-bg);
      color: var(--td-color-success);
      .status-dot { background: var(--td-color-success); }
    }
    &.disabled {
      background: var(--td-bg-section);
      color: var(--td-text-secondary);
      .status-dot { background: var(--td-text-placeholder); }
    }
  }

  .pagination-wrapper {
    padding-top: 20px;
    display: flex;
    justify-content: flex-end;
    border-top: 1px solid var(--td-divider-color);
    margin-top: 16px;
  }
}

// 对话框样式
.matchers-container {
  width: 100%;

  .matcher-item {
    display: flex;
    gap: 8px;
    margin-bottom: 12px;
    padding: 12px;
    background: var(--td-bg-page);
    border-radius: 8px;
    border: 1px solid var(--td-border-color);

    .matcher-input {
      flex: 1;
    }

    .matcher-operator {
      width: 100px;
    }
  }
}

.form-tip {
  font-size: 12px;
  color: var(--td-color-info);
  margin-top: 4px;
  line-height: 1.5;
}

// 响应式
@media (max-width: 768px) {
  .page-header {
    flex-direction: column;
    align-items: flex-start;
    gap: 16px;
  }
}
</style>
