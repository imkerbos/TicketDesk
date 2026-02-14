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
      <el-button type="primary" @click="handleCreate" class="header-btn">
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
        <el-table-column prop="issue_type_name" label="工单类型" width="100">
          <template #default="{ row }">
            <span class="type-text">{{ row.issue_type_name }}</span>
          </template>
        </el-table-column>
        <el-table-column label="标签匹配器" min-width="200">
          <template #default="{ row }">
            <el-tag
              v-for="(matcher, index) in row.label_matchers.slice(0, 2)"
              :key="index"
              size="small"
              style="margin-right: 4px"
            >
              {{ matcher.key }} {{ matcher.operator }} {{ matcher.value }}
            </el-tag>
            <el-tag v-if="row.label_matchers.length > 2" size="small">
              +{{ row.label_matchers.length - 2 }}
            </el-tag>
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
      <el-form :model="form" :rules="rules" ref="formRef" label-position="top">
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
            <el-form-item label="项目" prop="project_id">
              <el-select v-model="form.project_id" placeholder="请选择项目" style="width: 100%">
                <el-option label="示例项目" :value="1" />
              </el-select>
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="工单类型" prop="issue_type_id">
              <el-select v-model="form.issue_type_id" placeholder="请选择工单类型" style="width: 100%">
                <el-option label="故障" :value="4" />
                <el-option label="任务" :value="2" />
              </el-select>
            </el-form-item>
          </el-col>
        </el-row>
        <el-row :gutter="16">
          <el-col :span="24">
            <el-form-item label="标签匹配器" prop="label_matchers">
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
            <el-form-item label="优先级" prop="priority">
              <el-select v-model="form.priority" placeholder="请选择优先级" style="width: 100%">
                <el-option label="P0 - 紧急" value="P0" />
                <el-option label="P1 - 高" value="P1" />
                <el-option label="P2 - 中" value="P2" />
                <el-option label="P3 - 低" value="P3" />
              </el-select>
            </el-form-item>
          </el-col>
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
        </el-row>
        <el-row :gutter="16">
          <el-col :span="24">
            <el-form-item label="自动解决" prop="auto_resolve">
              <el-switch v-model="form.auto_resolve" />
              <div class="form-tip">
                告警恢复时自动解决关联的工单
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
} from '@/api/alert'
import type { AlertRule, LabelMatcher } from '@/types/alert'

const loading = ref(false)
const ruleList = ref<AlertRule[]>([])
const total = ref(0)
const queryParams = reactive({
  page: 1,
  page_size: 20,
})

const dialogVisible = ref(false)
const dialogTitle = ref('创建规则')
const formRef = ref<FormInstance>()
const form = reactive({
  id: 0,
  name: '',
  description: '',
  project_id: undefined as number | undefined,
  issue_type_id: undefined as number | undefined,
  label_matchers: [] as LabelMatcher[],
  priority: 'P2' as 'P0' | 'P1' | 'P2' | 'P3',
  assignee_id: undefined as number | undefined,
  auto_resolve: true,
  merge_window: 3600,
})

const rules: FormRules = {
  name: [{ required: true, message: '请输入规则名称', trigger: 'blur' }],
  project_id: [{ required: true, message: '请选择项目', trigger: 'change' }],
  issue_type_id: [{ required: true, message: '请选择工单类型', trigger: 'change' }],
  label_matchers: [{ required: true, message: '请添加至少一个标签匹配器', trigger: 'change' }],
  priority: [{ required: true, message: '请选择优先级', trigger: 'change' }],
}

const loadData = async () => {
  loading.value = true
  try {
    const { data } = await getAlertRuleList(queryParams)
    ruleList.value = data.data.items
    total.value = data.data.total
  } catch (error) {
    console.error('Failed to load rules:', error)
  } finally {
    loading.value = false
  }
}

const handleCreate = () => {
  dialogTitle.value = '创建规则'
  form.id = 0
  form.name = ''
  form.description = ''
  form.project_id = undefined
  form.issue_type_id = undefined
  form.label_matchers = [{ key: 'alertname', operator: '==', value: '' }]
  form.priority = 'P2'
  form.assignee_id = undefined
  form.auto_resolve = true
  form.merge_window = 3600
  dialogVisible.value = true
}

const handleEdit = (row: AlertRule) => {
  dialogTitle.value = '编辑规则'
  form.id = row.id
  form.name = row.name
  form.description = row.description
  form.project_id = row.project_id
  form.issue_type_id = row.issue_type_id
  form.label_matchers = JSON.parse(JSON.stringify(row.label_matchers))
  form.priority = row.priority
  form.assignee_id = row.assignee_id
  form.auto_resolve = row.auto_resolve
  form.merge_window = row.merge_window
  dialogVisible.value = true
}

const handleSubmit = async () => {
  if (!formRef.value) return
  await formRef.value.validate(async (valid) => {
    if (!valid) return

    try {
      if (form.id) {
        await updateAlertRule(form.id, form)
        ElMessage.success('更新成功')
      } else {
        // Validate required fields before creating
        if (!form.project_id || !form.issue_type_id) {
          ElMessage.error('请选择项目和工单类型')
          return
        }
        await createAlertRule({
          ...form,
          project_id: form.project_id,
          issue_type_id: form.issue_type_id,
        })
        ElMessage.success('创建成功')
      }
      dialogVisible.value = false
      loadData()
    } catch (error) {
      console.error('Failed to submit:', error)
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
      console.error('Failed to delete:', error)
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

onMounted(() => {
  loadData()
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
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  border-radius: 12px;
  color: #fff;

  .header-info {
    display: flex;
    align-items: center;
    gap: 16px;
  }

  .header-icon {
    width: 56px;
    height: 56px;
    background: rgba(255, 255, 255, 0.2);
    border-radius: 14px;
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
    color: #fff;
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
      background: #f8fafc;
      font-weight: 600;
      color: #374151;
    }
  }

  .type-text {
    font-size: 13px;
    color: #6b7280;
  }

  .priority-badge {
    display: inline-flex;
    align-items: center;
    gap: 5px;
    padding: 3px 10px;
    border-radius: 20px;
    font-size: 12px;
    font-weight: 500;

    .priority-dot {
      width: 6px;
      height: 6px;
      border-radius: 50%;
    }

    &.P0 {
      background: #fef2f2;
      color: #dc2626;
      .priority-dot { background: #ef4444; }
    }
    &.P1 {
      background: #fff7ed;
      color: #c2410c;
      .priority-dot { background: #f59e0b; }
    }
    &.P2 {
      background: #eff6ff;
      color: #1d4ed8;
      .priority-dot { background: #3b82f6; }
    }
    &.P3 {
      background: #f3f4f6;
      color: #6b7280;
      .priority-dot { background: #9ca3af; }
    }
  }

  .status-badge {
    display: inline-flex;
    align-items: center;
    gap: 5px;
    padding: 3px 10px;
    border-radius: 20px;
    font-size: 12px;

    .status-dot {
      width: 6px;
      height: 6px;
      border-radius: 50%;
    }

    &.enabled {
      background: #ecfdf5;
      color: #059669;
      .status-dot { background: #10b981; }
    }
    &.disabled {
      background: #f3f4f6;
      color: #6b7280;
      .status-dot { background: #9ca3af; }
    }
  }

  .pagination-wrapper {
    padding-top: 20px;
    display: flex;
    justify-content: flex-end;
    border-top: 1px solid #f0f0f0;
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
    background: #f8fafc;
    border-radius: 8px;
    border: 1px solid #e5e7eb;

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
  color: #909399;
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
