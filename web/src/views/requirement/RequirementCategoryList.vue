<template>
  <div class="category-container">
    <!-- 页面头部 -->
    <div class="page-header">
      <div class="header-info">
        <div class="header-icon">
          <el-icon><Collection /></el-icon>
        </div>
        <div class="header-text">
          <h1 class="header-title">需求分类管理</h1>
          <p class="header-desc">管理需求的分类标签，支持自定义分类</p>
        </div>
      </div>
      <el-button type="primary" @click="handleCreate">
        <el-icon><Plus /></el-icon>
        新增分类
      </el-button>
    </div>

    <!-- 分类列表 -->
    <el-card shadow="never" class="table-card">
      <el-table
        v-loading="loading"
        :data="categories"
        style="width: 100%"
        stripe
      >
        <el-table-column prop="sort_order" label="排序" width="80" align="center" />
        <el-table-column label="标识" min-width="180">
          <template #default="{ row }">
            <code class="name-code">{{ row.name }}</code>
          </template>
        </el-table-column>
        <el-table-column prop="label" label="显示名称" min-width="180" />
        <el-table-column label="预览" min-width="160">
          <template #default="{ row }">
            <el-tag :type="row.color" size="small">{{ row.label }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="默认" width="80" align="center">
          <template #default="{ row }">
            <el-tag v-if="row.is_default" type="success" size="small">默认</el-tag>
            <span v-else class="text-muted">-</span>
          </template>
        </el-table-column>
        <el-table-column label="类型" width="100" align="center">
          <template #default="{ row }">
            <el-tag v-if="row.is_system" size="small" type="info">系统</el-tag>
            <el-tag v-else size="small" type="success">自定义</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="操作" width="140" align="center" fixed="right">
          <template #default="{ row }">
            <div class="action-buttons">
              <el-button link type="primary" size="small" @click="handleEdit(row)">编辑</el-button>
              <el-button
                link
                type="danger"
                size="small"
                :disabled="row.is_system"
                @click="handleDelete(row)"
              >
                删除
              </el-button>
            </div>
          </template>
        </el-table-column>
      </el-table>
    </el-card>

    <!-- 创建/编辑对话框 -->
    <el-dialog
      v-model="showDialog"
      :title="editing ? '编辑分类' : '新增分类'"
      width="500px"
      destroy-on-close
      @closed="resetForm"
    >
      <el-form ref="formRef" :model="form" :rules="rules" label-width="100px">
        <el-form-item label="标识" prop="name">
          <el-input
            v-model="form.name"
            placeholder="英文标识，如 feature"
            :disabled="!!editing"
          />
          <div v-if="!editing" class="form-tip">创建后不可修改，建议使用英文小写</div>
        </el-form-item>
        <el-form-item label="显示名称" prop="label">
          <el-input v-model="form.label" placeholder="如 功能需求" />
        </el-form-item>
        <el-form-item label="标签颜色" prop="color">
          <el-select v-model="form.color" placeholder="选择颜色" style="width: 100%">
            <el-option label="蓝色 (primary)" value="primary">
              <el-tag type="primary" size="small">示例</el-tag>
              <span style="margin-left: 8px;">蓝色</span>
            </el-option>
            <el-option label="绿色 (success)" value="success">
              <el-tag type="success" size="small">示例</el-tag>
              <span style="margin-left: 8px;">绿色</span>
            </el-option>
            <el-option label="红色 (danger)" value="danger">
              <el-tag type="danger" size="small">示例</el-tag>
              <span style="margin-left: 8px;">红色</span>
            </el-option>
            <el-option label="橙色 (warning)" value="warning">
              <el-tag type="warning" size="small">示例</el-tag>
              <span style="margin-left: 8px;">橙色</span>
            </el-option>
            <el-option label="灰色 (info)" value="info">
              <el-tag type="info" size="small">示例</el-tag>
              <span style="margin-left: 8px;">灰色</span>
            </el-option>
          </el-select>
        </el-form-item>
        <el-form-item v-if="editing" label="排序" prop="sort_order">
          <el-input-number v-model="form.sort_order" :min="0" :max="999" />
        </el-form-item>
        <el-form-item v-if="editing" label="设为默认">
          <el-switch v-model="form.is_default" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="showDialog = false">取消</el-button>
        <el-button type="primary" :loading="submitting" @click="handleSubmit">确定</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { ElMessage, ElMessageBox, type FormInstance, type FormRules } from 'element-plus'
import { Plus, Collection } from '@element-plus/icons-vue'
import {
  getRequirementCategories,
  createRequirementCategory,
  updateRequirementCategory,
  deleteRequirementCategory,
} from '@/api/requirement'
import type { RequirementCategoryDef } from '@/types/requirement'

const categories = ref<RequirementCategoryDef[]>([])
const loading = ref(false)
const submitting = ref(false)
const showDialog = ref(false)
const editing = ref<RequirementCategoryDef | null>(null)
const formRef = ref<FormInstance>()

const form = reactive({
  name: '',
  label: '',
  color: 'info' as string,
  sort_order: 0,
  is_default: false,
})

const rules: FormRules = {
  name: [
    { required: true, message: '请输入标识', trigger: 'blur' },
    { pattern: /^[a-z][a-z0-9_]*$/, message: '只允许小写字母、数字和下划线，且以字母开头', trigger: 'blur' },
    { min: 1, max: 30, message: '长度 1-30 个字符', trigger: 'blur' },
  ],
  label: [
    { required: true, message: '请输入显示名称', trigger: 'blur' },
    { min: 1, max: 50, message: '长度 1-50 个字符', trigger: 'blur' },
  ],
  color: [{ required: true, message: '请选择颜色', trigger: 'change' }],
}

const loadCategories = async () => {
  loading.value = true
  try {
    const { data } = await getRequirementCategories()
    categories.value = data.data
  } catch {
    ElMessage.error('加载分类列表失败')
  } finally {
    loading.value = false
  }
}

const handleCreate = () => {
  resetForm()
  showDialog.value = true
}

const handleEdit = (cat: RequirementCategoryDef) => {
  editing.value = cat
  form.name = cat.name
  form.label = cat.label
  form.color = cat.color
  form.sort_order = cat.sort_order
  form.is_default = cat.is_default
  showDialog.value = true
}

const handleDelete = async (cat: RequirementCategoryDef) => {
  try {
    await ElMessageBox.confirm(`确定要删除分类"${cat.label}"吗？`, '提示', {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      type: 'warning',
    })
    await deleteRequirementCategory(cat.id)
    ElMessage.success('删除成功')
    loadCategories()
  } catch (error: any) {
    if (error !== 'cancel') {
      ElMessage.error(error.response?.data?.message || '删除失败')
    }
  }
}

const handleSubmit = async () => {
  if (!formRef.value) return

  await formRef.value.validate(async (valid) => {
    if (!valid) return

    submitting.value = true
    try {
      if (editing.value) {
        await updateRequirementCategory(editing.value.id, {
          label: form.label,
          color: form.color,
          sort_order: form.sort_order,
          is_default: form.is_default,
        })
        ElMessage.success('更新成功')
      } else {
        await createRequirementCategory({
          name: form.name,
          label: form.label,
          color: form.color,
        })
        ElMessage.success('创建成功')
      }
      showDialog.value = false
      loadCategories()
    } catch (error: any) {
      ElMessage.error(error.response?.data?.message || '操作失败')
    } finally {
      submitting.value = false
    }
  })
}

const resetForm = () => {
  editing.value = null
  form.name = ''
  form.label = ''
  form.color = 'info'
  form.sort_order = 0
  form.is_default = false
  formRef.value?.resetFields()
}

onMounted(() => {
  loadCategories()
})
</script>

<style scoped lang="scss">
.category-container {
  width: 100%; /* updated */
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
    border-radius: 12px;
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

  :deep(.el-button) {
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

.name-code {
  font-family: 'SF Mono', 'Monaco', 'Menlo', 'Consolas', monospace;
  font-size: 13px;
  color: var(--td-color-primary);
  background: var(--td-tag-primary-bg);
  padding: 2px 8px;
  border-radius: 4px;
}

.text-muted {
  color: var(--td-text-disabled);
  font-size: 13px;
}

.action-buttons {
  display: flex;
  justify-content: center;
  gap: 8px;
}

.form-tip {
  font-size: 12px;
  color: var(--td-text-placeholder);
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

