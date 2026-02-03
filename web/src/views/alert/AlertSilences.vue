<template>
  <div class="alert-silences-container">
    <el-card shadow="never" style="margin-bottom: 20px">
      <div style="display: flex; justify-content: space-between; align-items: center">
        <h2 style="margin: 0">告警静默管理</h2>
        <el-button type="primary" @click="handleCreate">
          <el-icon><Plus /></el-icon>
          创建静默
        </el-button>
      </div>
    </el-card>

    <el-card v-loading="loading" shadow="never">
      <el-table :data="silenceList" stripe>
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
        <el-table-column label="状态" width="100">
          <template #default="{ row }">
            <el-tag :type="getStatusType(row.status)" size="small">
              {{ getStatusText(row.status) }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="操作" width="200" fixed="right">
          <template #default="{ row }">
            <el-button link type="primary" size="small" @click="handleEdit(row)">编辑</el-button>
            <el-button
              v-if="row.status === 1"
              link
              type="warning"
              size="small"
              @click="handleCancel(row)"
            >
              取消
            </el-button>
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
      width="600px"
      @close="handleDialogClose"
    >
      <el-form :model="form" :rules="rules" ref="formRef" label-width="100px">
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
        <el-form-item label="生效时间" prop="starts_at">
          <el-date-picker
            v-model="form.starts_at"
            type="datetime"
            placeholder="选择开始时间"
            style="width: 100%"
          />
        </el-form-item>
        <el-form-item label="结束时间" prop="ends_at">
          <el-date-picker
            v-model="form.ends_at"
            type="datetime"
            placeholder="选择结束时间"
            style="width: 100%"
          />
        </el-form-item>
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
import { Plus, Delete } from '@element-plus/icons-vue'
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

const handleCancel = async (row: AlertSilence) => {
  try {
    await ElMessageBox.confirm('确定要取消此静默规则吗？', '提示', {
      type: 'warning',
    })
    await cancelAlertSilence(row.id)
    ElMessage.success('取消成功')
    loadData()
  } catch (error) {
    if (error !== 'cancel') {
      console.error('Failed to cancel:', error)
    }
  }
}

const handleDelete = async (row: AlertSilence) => {
  try {
    await ElMessageBox.confirm('确定要删除此静默规则吗？', '提示', {
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

const getStatusType = (status: number) => {
  const map: Record<number, any> = {
    0: 'info',
    1: 'success',
    2: 'warning',
  }
  return map[status] || 'info'
}

const getStatusText = (status: number) => {
  const map: Record<number, string> = {
    0: '已取消',
    1: '生效中',
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
