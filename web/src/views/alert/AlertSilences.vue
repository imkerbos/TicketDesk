<template>
  <div class="alert-silences-container">
    <!-- 页面头部 -->
    <div class="page-header">
      <div class="header-info">
        <div class="header-icon">
          <el-icon><BellFilled /></el-icon>
        </div>
        <div class="header-text">
          <h1 class="header-title">告警静默</h1>
          <p class="header-desc">管理告警静默规则，临时屏蔽特定告警通知</p>
        </div>
      </div>
      <el-button type="primary" class="header-btn" @click="handleCreate">
        <el-icon><Plus /></el-icon>
        创建静默
      </el-button>
    </div>

    <!-- 静默列表 -->
    <el-card shadow="never" class="table-card">
      <el-table
        v-loading="loading"
        :data="silenceList"
        style="width: 100%"
      >
        <el-table-column prop="name" label="静默名称" min-width="150" />
        <el-table-column prop="description" label="描述" min-width="200" show-overflow-tooltip />
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
        <el-table-column label="生效时间" width="160">
          <template #default="{ row }">
            {{ formatTime(row.starts_at) }}
          </template>
        </el-table-column>
        <el-table-column label="结束时间" width="160">
          <template #default="{ row }">
            {{ formatTime(row.ends_at) }}
          </template>
        </el-table-column>
        <el-table-column label="状态" width="100" align="center">
          <template #default="{ row }">
            <div class="status-badge" :class="getStatusClass(row.status)">
              <span class="status-dot"></span>
              <span class="status-text">{{ getStatusText(row.status) }}</span>
            </div>
          </template>
        </el-table-column>
        <el-table-column label="操作" width="180" fixed="right" align="center">
          <template #default="{ row }">
            <div class="action-buttons">
              <el-tooltip v-if="row.status === 0" content="启用" placement="top">
                <el-button link type="success" @click="handleEnable(row)">
                  <el-icon><CircleCheck /></el-icon>
                </el-button>
              </el-tooltip>
              <el-tooltip v-if="row.status === 1" content="关闭" placement="top">
                <el-button link type="warning" @click="handleDisable(row)">
                  <el-icon><CircleClose /></el-icon>
                </el-button>
              </el-tooltip>
              <el-tooltip content="编辑" placement="top">
                <el-button link type="primary" @click="handleEdit(row)">
                  <el-icon><Edit /></el-icon>
                </el-button>
              </el-tooltip>
              <el-tooltip content="删除" placement="top">
                <el-button link type="danger" @click="handleDelete(row)">
                  <el-icon><Delete /></el-icon>
                </el-button>
              </el-tooltip>
            </div>
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
      width="600px"
      @close="handleDialogClose"
    >
      <el-form :model="form" :rules="rules" ref="formRef" label-position="top">
        <el-form-item label="静默名称" prop="name">
          <el-input v-model="form.name" placeholder="请输入静默名称" />
        </el-form-item>
        <el-form-item label="描述" prop="description">
          <el-input v-model="form.description" type="textarea" :rows="2" placeholder="请输入描述" />
        </el-form-item>
        <el-form-item label="标签匹配器" prop="label_matchers">
          <div style="width: 100%">
            <div
              v-for="(matcher, index) in form.label_matchers"
              :key="index"
              class="matcher-row"
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
        <el-row :gutter="16">
          <el-col :span="12">
            <el-form-item label="生效时间" prop="starts_at">
              <el-date-picker
                v-model="form.starts_at"
                type="datetime"
                placeholder="选择开始时间"
                style="width: 100%"
              />
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="结束时间" prop="ends_at">
              <el-date-picker
                v-model="form.ends_at"
                type="datetime"
                placeholder="选择结束时间"
                style="width: 100%"
              />
            </el-form-item>
          </el-col>
        </el-row>
        <el-form-item label="备注" prop="comment">
          <el-input v-model="form.comment" type="textarea" :rows="2" placeholder="请输入静默原因" />
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
import { Plus, Delete, Edit, CircleClose, CircleCheck, BellFilled } from '@element-plus/icons-vue'
import {
  getAlertSilenceList,
  createAlertSilence,
  updateAlertSilence,
  deleteAlertSilence,
  cancelAlertSilence,
} from '@/api/alert'
import type { AlertSilence, LabelMatcher } from '@/types/alert'
import dayjs from 'dayjs'

const loading = ref(false)
const silenceList = ref<AlertSilence[]>([])
const total = ref(0)
const queryParams = reactive({
  page: 1,
  page_size: 20,
})

const dialogVisible = ref(false)
const dialogTitle = ref('创建静默')
const formRef = ref<FormInstance>()
const form = reactive({
  id: 0,
  name: '',
  description: '',
  label_matchers: [] as LabelMatcher[],
  starts_at: '',
  ends_at: '',
  comment: '',
})

const rules: FormRules = {
  name: [{ required: true, message: '请输入静默名称', trigger: 'blur' }],
  label_matchers: [{ required: true, message: '请添加至少一个标签匹配器', trigger: 'change' }],
  starts_at: [{ required: true, message: '请选择开始时间', trigger: 'change' }],
  ends_at: [{ required: true, message: '请选择结束时间', trigger: 'change' }],
}

const loadData = async () => {
  loading.value = true
  try {
    const { data } = await getAlertSilenceList(queryParams)
    silenceList.value = data.data.items
    total.value = data.data.total
  } catch (error) {
    console.error('Failed to load silences:', error)
  } finally {
    loading.value = false
  }
}

const handleCreate = () => {
  dialogTitle.value = '创建静默'
  form.id = 0
  form.name = ''
  form.description = ''
  form.label_matchers = [{ key: '', operator: '==', value: '' }]
  form.starts_at = ''
  form.ends_at = ''
  form.comment = ''
  dialogVisible.value = true
}

const handleEdit = (row: AlertSilence) => {
  dialogTitle.value = '编辑静默'
  form.id = row.id
  form.name = row.name
  form.description = row.description
  form.label_matchers = JSON.parse(JSON.stringify(row.label_matchers))
  form.starts_at = row.starts_at
  form.ends_at = row.ends_at
  form.comment = row.comment
  dialogVisible.value = true
}

const handleSubmit = async () => {
  if (!formRef.value) return
  await formRef.value.validate(async (valid) => {
    if (!valid) return

    try {
      if (form.id) {
        await updateAlertSilence(form.id, form)
        ElMessage.success('更新成功')
      } else {
        await createAlertSilence(form)
        ElMessage.success('创建成功')
      }
      dialogVisible.value = false
      loadData()
    } catch (error) {
      console.error('Failed to submit:', error)
    }
  })
}

// 启用静默规则
const handleEnable = async (row: AlertSilence) => {
  try {
    await ElMessageBox.confirm(
      '启用后，匹配的告警将不会自动创建工单。确定要启用此静默规则吗？',
      '启用静默',
      { type: 'info', confirmButtonText: '启用', cancelButtonText: '取消' }
    )
    await updateAlertSilence(row.id, { status: 1 } as any)
    ElMessage.success('已启用')
    loadData()
  } catch (error) {
    if (error !== 'cancel') {
      console.error('Failed to enable:', error)
    }
  }
}

// 关闭静默规则
const handleDisable = async (row: AlertSilence) => {
  try {
    await ElMessageBox.confirm('关闭后，匹配的告警将恢复正常建单。确定要关闭此静默规则吗？', '关闭静默', {
      type: 'warning',
      confirmButtonText: '关闭',
      cancelButtonText: '取消',
    })
    await cancelAlertSilence(row.id)
    ElMessage.success('已关闭')
    loadData()
  } catch (error) {
    if (error !== 'cancel') {
      console.error('Failed to disable:', error)
    }
  }
}

const handleDelete = async (row: AlertSilence) => {
  try {
    await ElMessageBox.confirm('确定要删除此静默规则吗？删除后不可恢复。', '删除静默', {
      type: 'warning',
    })
    await deleteAlertSilence(row.id)
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

const getStatusClass = (status: number) => {
  const map: Record<number, string> = {
    0: 'disabled',
    1: 'active',
    2: 'expired',
  }
  return map[status] || 'disabled'
}

const getStatusText = (status: number) => {
  const map: Record<number, string> = {
    0: '已关闭',
    1: '已启用',
    2: '已过期',
  }
  return map[status] || '未知'
}

const formatTime = (time: string) => {
  return dayjs(time).format('YYYY-MM-DD HH:mm')
}

onMounted(() => {
  loadData()
})
</script>

<style scoped lang="scss">
.alert-silences-container {
  width: 100%;
}

// 页面头部
.page-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 24px;
  padding: 24px;
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
    border-radius: 14px;
    display: flex;
    align-items: center;
    justify-content: center;
    font-size: 28px;
  }

  .header-text {
    .header-title {
      font-size: 22px;
      font-weight: 600;
      margin: 0 0 4px 0;
    }

    .header-desc {
      font-size: 14px;
      margin: 0;
      opacity: 0.9;
    }
  }

  .header-btn {
    background: rgba(255, 255, 255, 0.2);
    border: 1px solid rgba(255, 255, 255, 0.3);
    color: var(--td-text-white);

    &:hover {
      background: rgba(255, 255, 255, 0.3);
    }
  }
}

// 表格卡片
.table-card {
  border-radius: 12px;

  :deep(.el-card__body) {
    padding: 0;
  }

  :deep(.el-table) {
    border-radius: 12px;

    th.el-table__cell {
      background: var(--td-bg-page);
      font-weight: 600;
      color: var(--td-text-regular);
    }
  }
}

// 状态徽章
.status-badge {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  padding: 4px 10px;
  border-radius: 20px;
  font-size: 12px;

  .status-dot {
    width: 6px;
    height: 6px;
    border-radius: 50%;
  }

  &.active {
    background: var(--td-tag-success-bg);
    color: var(--td-color-success);

    .status-dot {
      background: var(--td-color-success);
    }
  }

  &.disabled {
    background: var(--td-bg-section);
    color: var(--td-text-secondary);

    .status-dot {
      background: var(--td-text-placeholder);
    }
  }

  &.expired {
    background: var(--td-tag-warning-border);
    color: var(--td-color-warning);

    .status-dot {
      background: var(--td-color-warning);
    }
  }
}

// 操作按钮
.action-buttons {
  display: flex;
  justify-content: center;
  gap: 8px;

  .el-button {
    font-size: 16px;
    padding: 4px;

    &:hover {
      background: var(--td-bg-section);
      border-radius: 6px;
    }
  }
}

// 分页
.pagination-wrapper {
  padding: 20px;
  display: flex;
  justify-content: flex-end;
  border-top: 1px solid var(--td-divider-color);
}

// 匹配器行
.matcher-row {
  display: flex;
  gap: 8px;
  margin-bottom: 8px;
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
