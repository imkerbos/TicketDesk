<template>
  <div class="alert-rules-container">
    <el-card shadow="never" style="margin-bottom: 20px">
      <div style="display: flex; justify-content: space-between; align-items: center">
        <h2 style="margin: 0">告警规则配置</h2>
        <el-button type="primary" @click="handleCreate">
          <el-icon><Plus /></el-icon>
          创建规则
        </el-button>
      </div>
    </el-card>

    <el-card v-loading="loading" shadow="never">
      <el-table :data="ruleList" stripe>
        <el-table-column prop="name" label="规则名称" min-width="150" />
        <el-table-column prop="project_name" label="项目" width="120" />
        <el-table-column prop="issue_type_name" label="工单类型" width="100" />
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
        <el-table-column prop="priority" label="优先级" width="80">
          <template #default="{ row }">
            <el-tag :type="getPriorityType(row.priority)" size="small">
              {{ row.priority }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="合并窗口" width="100">
          <template #default="{ row }">
            {{ formatMergeWindow(row.merge_window) }}
          </template>
        </el-table-column>
        <el-table-column label="自动解决" width="80">
          <template #default="{ row }">
            <el-tag :type="row.auto_resolve ? 'success' : 'info'" size="small">
              {{ row.auto_resolve ? '是' : '否' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="状态" width="80">
          <template #default="{ row }">
            <el-tag :type="row.status === 1 ? 'success' : 'info'" size="small">
              {{ row.status === 1 ? '启用' : '禁用' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="操作" width="180" fixed="right">
          <template #default="{ row }">
            <el-button link type="primary" size="small" @click="handleEdit(row)">编辑</el-button>
            <el-button link type="danger" size="small" @click="handleDelete(row)">删除</el-button>
          </template>
        </el-table-column>
      </el-table>

      <el-pagination
        v-model:current-page="queryParams.page"
        v-model:page-size="queryParams.page_size"
        :total="total"
        layout="total, prev, pager, next"
        style="margin-top: 20px; justify-content: flex-end"
        @current-change="loadData"
      />
    </el-card>

    <!-- 创建/编辑对话框 -->
    <el-dialog
      v-model="dialogVisible"
      :title="dialogTitle"
      width="700px"
      @close="handleDialogClose"
    >
      <el-form :model="form" :rules="rules" ref="formRef" label-width="120px">
        <el-form-item label="规则名称" prop="name">
          <el-input v-model="form.name" placeholder="请输入规则名称" />
        </el-form-item>
        <el-form-item label="描述" prop="description">
          <el-input v-model="form.description" type="textarea" :rows="2" placeholder="请输入描述" />
        </el-form-item>
        <el-form-item label="项目" prop="project_id">
          <el-select v-model="form.project_id" placeholder="请选择项目" style="width: 100%">
            <el-option label="示例项目" :value="1" />
          </el-select>
        </el-form-item>
        <el-form-item label="工单类型" prop="issue_type_id">
          <el-select v-model="form.issue_type_id" placeholder="请选择工单类型" style="width: 100%">
            <el-option label="故障" :value="4" />
            <el-option label="任务" :value="2" />
          </el-select>
        </el-form-item>
        <el-form-item label="标签匹配器" prop="label_matchers">
          <div style="width: 100%">
            <div
              v-for="(matcher, index) in form.label_matchers"
              :key="index"
              style="display: flex; gap: 8px; margin-bottom: 8px"
            >
              <el-input v-model="matcher.key" placeholder="标签键" style="flex: 1" />
              <el-select v-model="matcher.operator" style="width: 100px">
                <el-option label="==" value="==" />
                <el-option label="!=" value="!=" />
                <el-option label="=~" value="=~" />
                <el-option label="!~" value="!~" />
              </el-select>
              <el-input v-model="matcher.value" placeholder="标签值" style="flex: 1" />
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
        <el-form-item label="优先级" prop="priority">
          <el-select v-model="form.priority" placeholder="请选择优先级" style="width: 100%">
            <el-option label="P0 - 紧急" value="P0" />
            <el-option label="P1 - 高" value="P1" />
            <el-option label="P2 - 中" value="P2" />
            <el-option label="P3 - 低" value="P3" />
          </el-select>
        </el-form-item>
        <el-form-item label="合并窗口" prop="merge_window">
          <el-input-number
            v-model="form.merge_window"
            :min="0"
            :max="86400"
            :step="300"
            style="width: 100%"
          />
          <div style="font-size: 12px; color: #909399; margin-top: 4px">
            单位：秒，0 表示不合并。建议设置为 3600（1小时）
          </div>
        </el-form-item>
        <el-form-item label="自动解决" prop="auto_resolve">
          <el-switch v-model="form.auto_resolve" />
          <div style="font-size: 12px; color: #909399; margin-top: 4px">
            告警恢复时自动解决关联的工单
          </div>
        </el-form-item>
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
import { Plus, Delete } from '@element-plus/icons-vue'
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

const getPriorityType = (priority: string) => {
  const map: Record<string, any> = {
    P0: 'danger',
    P1: 'warning',
    P2: 'primary',
    P3: 'info',
  }
  return map[priority] || 'info'
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
