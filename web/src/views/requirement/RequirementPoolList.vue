<template>
  <div class="requirement-pool-list">
    <div class="page-header">
      <h1>需求池管理</h1>
      <el-button type="primary" @click="handleCreate">
        <el-icon><Plus /></el-icon>
        创建需求池
      </el-button>
    </div>

    <!-- 筛选条件 -->
    <el-card class="filter-card" shadow="never">
      <el-form :inline="true" :model="filters">
        <el-form-item label="类型">
          <el-select v-model="filters.type" placeholder="全部" clearable style="width: 120px">
            <el-option label="全局" value="global" />
            <el-option label="项目级" value="project" />
          </el-select>
        </el-form-item>
        <el-form-item label="状态">
          <el-select v-model="filters.status" placeholder="全部" clearable style="width: 120px">
            <el-option label="活跃" value="active" />
            <el-option label="已归档" value="archived" />
          </el-select>
        </el-form-item>
        <el-form-item>
          <el-button type="primary" @click="loadData">查询</el-button>
          <el-button @click="resetFilters">重置</el-button>
        </el-form-item>
      </el-form>
    </el-card>

    <!-- 需求池列表 -->
    <el-card class="table-card" shadow="never">
      <el-table v-loading="loading" :data="pools" stripe>
        <el-table-column prop="name" label="名称" min-width="200">
          <template #default="{ row }">
            <span class="link" style="cursor: pointer;" @click="handleViewRequirements(row)">
              {{ row.name }}
            </span>
          </template>
        </el-table-column>
        <el-table-column prop="type" label="类型" width="100">
          <template #default="{ row }">
            <el-tag :type="row.type === 'global' ? 'primary' : 'success'" size="small">
              {{ row.type === 'global' ? '全局' : '项目级' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="project_name" label="关联项目" width="150" />
        <el-table-column prop="owner_name" label="负责人" width="120" show-overflow-tooltip />
        <el-table-column prop="requirement_count" label="需求数量" width="100" align="center" />
        <el-table-column prop="status" label="状态" width="100">
          <template #default="{ row }">
            <el-tag :type="row.status === 'active' ? 'success' : 'info'" size="small">
              {{ row.status === 'active' ? '活跃' : '已归档' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="创建时间" width="150" show-overflow-tooltip>
          <template #default="{ row }">
            {{ formatDateTime(row.created_at) }}
          </template>
        </el-table-column>
        <el-table-column label="操作" width="200" fixed="right">
          <template #default="{ row }">
            <el-button link type="primary" size="small" @click="handleEdit(row)">编辑</el-button>
            <el-button link type="primary" size="small" @click="handleViewRequirements(row)">查看需求</el-button>
            <el-button link type="danger" size="small" @click="handleDelete(row)">删除</el-button>
          </template>
        </el-table-column>
      </el-table>

      <!-- 分页 -->
      <div class="pagination">
        <el-pagination
          v-model:current-page="pagination.page"
          v-model:page-size="pagination.page_size"
          :total="pagination.total"
          :page-sizes="[10, 20, 50, 100]"
          layout="total, sizes, prev, pager, next, jumper"
          @size-change="loadData"
          @current-change="loadData"
        />
      </div>
    </el-card>

    <!-- 创建/编辑对话框 -->
    <el-dialog
      v-model="showCreateDialog"
      :title="editingPool ? '编辑需求池' : '创建需求池'"
      width="600px"
      @closed="resetForm"
    >
      <el-form ref="formRef" :model="form" :rules="rules" label-width="100px">
        <el-form-item label="名称" prop="name">
          <el-input v-model="form.name" placeholder="请输入需求池名称" />
        </el-form-item>
        <el-form-item label="描述" prop="description">
          <el-input
            v-model="form.description"
            type="textarea"
            :rows="3"
            placeholder="请输入需求池描述"
          />
        </el-form-item>
        <el-form-item label="类型" prop="type">
          <el-radio-group v-model="form.type" :disabled="!!editingPool">
            <el-radio label="global">全局需求池</el-radio>
            <el-radio label="project">项目级需求池</el-radio>
          </el-radio-group>
        </el-form-item>
        <el-form-item v-if="form.type === 'project'" label="关联项目" prop="project_id">
          <el-select v-model="form.project_id" placeholder="请选择项目" :disabled="!!editingPool" style="width: 100%">
            <el-option
              v-for="project in projects"
              :key="project.id"
              :label="project.name"
              :value="project.id"
            />
          </el-select>
        </el-form-item>
        <el-form-item label="负责人" prop="owner_id">
          <el-select v-model="form.owner_id" placeholder="请选择负责人" filterable style="width: 100%">
            <el-option
              v-for="user in users"
              :key="user.id"
              :label="user.display_name"
              :value="user.id"
            />
          </el-select>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="handleCancel">取消</el-button>
        <el-button type="primary" :loading="submitting" @click="handleSubmit">确定</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage, ElMessageBox, type FormInstance, type FormRules } from 'element-plus'
import { Plus } from '@element-plus/icons-vue'
import {
  getRequirementPoolList,
  createRequirementPool,
  updateRequirementPool,
  deleteRequirementPool,
} from '@/api/requirement'
import { getAllProjects } from '@/api/project'
import { getAllUsers } from '@/api/user'
import type { RequirementPool, CreateRequirementPoolRequest, UpdateRequirementPoolRequest, RequirementPoolType, RequirementPoolStatus } from '@/types/requirement'

const router = useRouter()

// 数据
const pools = ref<RequirementPool[]>([])
const projects = ref<any[]>([])
const users = ref<any[]>([])
const loading = ref(false)
const submitting = ref(false)
const showCreateDialog = ref(false)
const editingPool = ref<RequirementPool | null>(null)
const formRef = ref<FormInstance>()

// 筛选条件
const filters = reactive({
  type: undefined as RequirementPoolType | undefined,
  status: undefined as RequirementPoolStatus | undefined,
})

// 分页
const pagination = reactive({
  page: 1,
  page_size: 20,
  total: 0,
})

// 表单
const form = reactive<CreateRequirementPoolRequest>({
  name: '',
  description: '',
  type: 'global',
  owner_id: undefined,
})

// 表单验证规则
const rules: FormRules = {
  name: [{ required: true, message: '请输入需求池名称', trigger: 'blur' }],
  type: [{ required: true, message: '请选择类型', trigger: 'change' }],
  owner_id: [{
    required: true,
    trigger: 'change',
    validator: (_rule, value, callback) => {
      if (!value || value === 0) {
        callback(new Error('请选择负责人'))
      } else {
        callback()
      }
    }
  }],
  project_id: [
    {
      required: true,
      message: '请选择关联项目',
      trigger: 'change',
      validator: (_rule, value, callback) => {
        if (form.type === 'project' && !value) {
          callback(new Error('项目级需求池必须关联项目'))
        } else {
          callback()
        }
      },
    },
  ],
}

// 格式化日期时间
const formatDateTime = (dateStr: string | undefined) => {
  if (!dateStr) return '-'
  const date = new Date(dateStr)
  const year = date.getFullYear()
  const month = String(date.getMonth() + 1).padStart(2, '0')
  const day = String(date.getDate()).padStart(2, '0')
  const hours = String(date.getHours()).padStart(2, '0')
  const minutes = String(date.getMinutes()).padStart(2, '0')
  return `${year}-${month}-${day} ${hours}:${minutes}`
}

// 加载数据
const loadData = async () => {
  loading.value = true
  try {
    const { data } = await getRequirementPoolList({
      ...filters,
      page: pagination.page,
      page_size: pagination.page_size,
    })
    pools.value = data.data.items
    pagination.total = data.data.total
  } catch {
    ElMessage.error('加载需求池列表失败')
  } finally {
    loading.value = false
  }
}

// 加载项目列表
const loadProjects = async () => {
  try {
    const { data } = await getAllProjects()
    projects.value = data.data
  } catch {
    // ignored
  }
}

// 加载用户列表
const loadUsers = async () => {
  try {
    const { data } = await getAllUsers()
    users.value = data.data
  } catch {
    // ignored
  }
}

// 重置筛选条件
const resetFilters = () => {
  filters.type = undefined
  filters.status = undefined
  pagination.page = 1
  loadData()
}

// 编辑
const handleEdit = (pool: RequirementPool) => {
  editingPool.value = pool
  form.name = pool.name
  form.description = pool.description
  form.type = pool.type
  form.owner_id = pool.owner_id
  form.project_id = pool.project_id
  showCreateDialog.value = true
}

// 查看需求
const handleViewRequirements = (pool: RequirementPool) => {
  router.push(`/requirements?pool_id=${pool.id}`)
}

// 删除
const handleDelete = async (pool: RequirementPool) => {
  try {
    await ElMessageBox.confirm(`确定要删除需求池"${pool.name}"吗？`, '提示', {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      type: 'warning',
    })

    await deleteRequirementPool(pool.id)
    ElMessage.success('删除成功')
    loadData()
  } catch (error: any) {
    if (error !== 'cancel') {
      ElMessage.error(error.response?.data?.message || '删除失败')
    }
  }
}

// 提交表单
const handleSubmit = async () => {
  if (!formRef.value) return

  await formRef.value.validate(async (valid) => {
    if (!valid) return

    submitting.value = true
    try {
      if (editingPool.value) {
        // 更新
        const updateData: UpdateRequirementPoolRequest = {
          name: form.name,
          description: form.description,
          owner_id: form.owner_id,
        }
        await updateRequirementPool(editingPool.value.id, updateData)
        ElMessage.success('更新成功')
      } else {
        // 创建
        await createRequirementPool(form)
        ElMessage.success('创建成功')
      }

      showCreateDialog.value = false
      resetForm()
      loadData()
    } catch (error: any) {
      ElMessage.error(error.response?.data?.message || '操作失败')
    } finally {
      submitting.value = false
    }
  })
}

// 创建需求池
const handleCreate = () => {
  resetForm()
  showCreateDialog.value = true
}

// 取消对话框
const handleCancel = () => {
  showCreateDialog.value = false
}

// 重置表单
const resetForm = () => {
  editingPool.value = null
  form.name = ''
  form.description = ''
  form.type = 'global'
  form.owner_id = undefined
  form.project_id = undefined
  formRef.value?.resetFields()
}

onMounted(() => {
  loadData()
  loadProjects()
  loadUsers()
})
</script>

<style scoped lang="scss">
.requirement-pool-list {
  padding: 24px;
  background: var(--td-bg-page);
  min-height: 100vh;

  .page-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-bottom: 24px;
    padding: 20px 24px;
    background: var(--td-color-primary);
    border-radius: 12px;
    box-shadow: 0 4px 20px rgba(59, 130, 246, 0.3);

    h1 {
      margin: 0;
      font-size: 28px;
      font-weight: 600;
      color: var(--td-text-white);
      text-shadow: 0 2px 4px rgba(0, 0, 0, 0.1);
    }

    :deep(.el-button) {
      background: rgba(255, 255, 255, 0.2);
      border: 1px solid rgba(255, 255, 255, 0.3);
      color: var(--td-text-white);
      backdrop-filter: blur(10px);
      transition: all 150ms ease-out;

      &:hover {
        background: rgba(255, 255, 255, 0.3);
        box-shadow: 0 4px 12px rgba(0, 0, 0, 0.15);
      }
    }
  }

  .filter-card {
    margin-bottom: 20px;
    border-radius: 12px;
    border: none;
    box-shadow: 0 2px 12px rgba(0, 0, 0, 0.08);
    transition: box-shadow 150ms ease-out;

    &:hover {
      box-shadow: 0 4px 20px rgba(0, 0, 0, 0.12);
    }

    :deep(.el-card__body) {
      padding: 20px 24px;
    }

    :deep(.el-form-item) {
      margin-bottom: 0;
    }

    :deep(.el-button--primary) {
      background: var(--td-color-primary);
      border: none;
      transition: all 150ms ease-out;

      &:hover {
        box-shadow: 0 4px 12px rgba(59, 130, 246, 0.4);
      }
    }
  }

  .table-card {
    border-radius: 12px;
    border: none;
    box-shadow: 0 2px 12px rgba(0, 0, 0, 0.08);
    overflow: hidden;

    :deep(.el-card__body) {
      padding: 0;
    }

    :deep(.el-table) {
      border-radius: 12px;

      th {
        background: var(--td-table-header-bg);
        color: var(--td-text-regular);
        font-weight: 600;
        font-size: 14px;
        padding: 16px 12px;
      }

      td {
        padding: 16px 12px;
      }

      .el-table__row {
        transition: all 150ms ease-out;

        &:hover {
          background: var(--td-bg-card-hover) !important;
          box-shadow: 0 2px 8px rgba(59, 130, 246, 0.1);
        }
      }
    }

    .link {
      color: var(--td-color-primary);
      text-decoration: none;
      font-weight: 500;
      transition: all 150ms ease-out;

      &:hover {
        color: var(--td-color-primary-hover);
        text-decoration: underline;
      }
    }

    :deep(.el-tag) {
      border-radius: 6px;
      padding: 4px 12px;
      font-weight: 500;
      border: none;
    }

    :deep(.el-button--primary) {
      color: var(--td-color-primary);

      &:hover {
        color: var(--td-color-primary-hover);
        background: var(--td-tag-primary-bg);
      }
    }

    :deep(.el-button--danger) {
      color: var(--td-color-danger);

      &:hover {
        color: var(--td-color-danger);
        background: var(--td-tag-danger-bg);
      }
    }

    .pagination {
      margin-top: 0;
      padding: 20px 24px;
      display: flex;
      justify-content: flex-end;
      background: var(--td-bg-section);
      border-top: 1px solid var(--td-border-color);
    }
  }

  :deep(.el-dialog) {
    border-radius: 12px;
    box-shadow: 0 8px 32px rgba(0, 0, 0, 0.12);

    .el-dialog__header {
      padding: 20px 24px;
      background: var(--td-color-primary);
      border-radius: 12px 12px 0 0;

      .el-dialog__title {
        color: var(--td-text-white);
        font-weight: 600;
        font-size: 18px;
      }

      .el-dialog__headerbtn .el-dialog__close {
        color: var(--td-text-white);
        font-size: 20px;

        &:hover {
          color: var(--td-text-white);
        }
      }
    }

    .el-dialog__body {
      padding: 24px;
    }

    .el-dialog__footer {
      padding: 16px 24px;
      border-top: 1px solid #e9ecef;

      .el-button--primary {
        background: var(--td-color-primary);
        border: none;
        padding: 10px 24px;
        transition: all 150ms ease-out;

        &:hover {
          box-shadow: 0 4px 12px rgba(59, 130, 246, 0.4);
        }
      }
    }
  }
}
</style>
