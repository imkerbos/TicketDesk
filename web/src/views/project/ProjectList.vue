<template>
  <div class="project-list-container">
    <!-- 搜索和操作栏 -->
    <el-card shadow="never" class="toolbar-card">
      <div class="toolbar">
        <div class="toolbar-left">
          <el-input
            v-model="keyword"
            placeholder="搜索项目名称"
            clearable
            style="width: 300px"
            @clear="handleSearch"
            @keyup.enter="handleSearch"
          >
            <template #prefix>
              <el-icon><Search /></el-icon>
            </template>
          </el-input>
          <el-button type="primary" :icon="Search" @click="handleSearch">搜索</el-button>
        </div>
        <div class="toolbar-right">
          <el-button type="primary" :icon="Plus" @click="handleCreate">创建项目</el-button>
        </div>
      </div>
    </el-card>

    <!-- 项目列表 -->
    <div v-loading="loading" class="project-grid">
      <el-row :gutter="20">
        <el-col
          v-for="project in projectList"
          :key="project.id"
          :xs="24"
          :sm="12"
          :md="8"
          :lg="6"
          class="project-col"
        >
          <el-card shadow="hover" class="project-card" @click="handleViewProject(project)">
            <div class="project-header">
              <div class="project-icon">{{ project.project_key?.substring(0, 2).toUpperCase() || '??' }}</div>
              <div class="project-info">
                <div class="project-name">{{ project.name }}</div>
                <div class="project-key">{{ project.project_key }}</div>
              </div>
            </div>
            <div class="project-description">
              {{ project.description || '暂无描述' }}
            </div>
            <div class="project-stats">
              <div class="stat-item">
                <el-icon><Tickets /></el-icon>
                <span>{{ project.member_count || 0 }} 成员</span>
              </div>
            </div>
            <div class="project-footer">
              <span class="project-lead">
                <el-icon><User /></el-icon>
                {{ project.lead_user?.display_name || '未指定' }}
              </span>
              <el-tag v-if="project.status === 1" type="success" size="small">启用</el-tag>
              <el-tag v-else type="info" size="small">禁用</el-tag>
            </div>
          </el-card>
        </el-col>
      </el-row>

      <div v-if="projectList.length === 0 && !loading" class="empty-state">
        <el-empty description="暂无项目">
          <el-button type="primary" @click="handleCreate">创建第一个项目</el-button>
        </el-empty>
      </div>

      <el-pagination
        v-if="total > 0"
        v-model:current-page="page"
        v-model:page-size="pageSize"
        :total="total"
        :page-sizes="[12, 24, 48]"
        layout="total, sizes, prev, pager, next"
        class="pagination"
        @size-change="loadProjects"
        @current-change="loadProjects"
      />
    </div>

    <!-- 创建/编辑项目对话框 -->
    <el-dialog
      v-model="dialogVisible"
      :title="isEdit ? '编辑项目' : '创建项目'"
      width="500px"
      destroy-on-close
    >
      <el-form ref="formRef" :model="form" :rules="rules" label-width="80px">
        <el-form-item label="项目标识" prop="project_key" :disabled="isEdit">
          <el-input
            v-model="form.project_key"
            placeholder="如: PROJ"
            :disabled="isEdit"
            maxlength="10"
          />
          <div class="form-tip">项目标识用于工单编号前缀，创建后不可修改</div>
        </el-form-item>
        <el-form-item label="项目名称" prop="name">
          <el-input v-model="form.name" placeholder="请输入项目名称" maxlength="50" />
        </el-form-item>
        <el-form-item label="负责人">
          <el-select v-model="form.lead_user_id" placeholder="请选择负责人" style="width: 100%" clearable filterable>
            <el-option v-for="u in users" :key="u.id" :label="u.display_name" :value="u.id" />
          </el-select>
        </el-form-item>
        <el-form-item label="描述">
          <el-input v-model="form.description" type="textarea" :rows="3" placeholder="请输入项目描述" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="submitLoading" @click="submitForm">
          {{ isEdit ? '保存' : '创建' }}
        </el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage, type FormInstance, type FormRules } from 'element-plus'
import { Search, Plus, Tickets, User } from '@element-plus/icons-vue'
import { getProjectList, createProject, updateProject } from '@/api/project'
import { getAllUsers } from '@/api/user'
import type { Project, CreateProjectRequest } from '@/types/project'
import type { UserOption } from '@/types/user'

const router = useRouter()

// 数据
const loading = ref(false)
const projectList = ref<Project[]>([])
const total = ref(0)
const page = ref(1)
const pageSize = ref(12)
const keyword = ref('')
const users = ref<UserOption[]>([])

// 对话框相关
const dialogVisible = ref(false)
const isEdit = ref(false)
const submitLoading = ref(false)
const formRef = ref<FormInstance>()
const editingProject = ref<Project | null>(null)

const form = reactive<CreateProjectRequest>({
  project_key: '',
  name: '',
  description: '',
  lead_user_id: undefined,
})

const rules: FormRules = {
  project_key: [
    { required: true, message: '请输入项目标识', trigger: 'blur' },
    { pattern: /^[A-Z][A-Z0-9]*$/, message: '项目标识必须以大写字母开头，只能包含大写字母和数字', trigger: 'blur' },
    { min: 2, max: 10, message: '长度为 2-10 个字符', trigger: 'blur' },
  ],
  name: [
    { required: true, message: '请输入项目名称', trigger: 'blur' },
    { max: 50, message: '最多 50 个字符', trigger: 'blur' },
  ],
}

// 加载项目列表
const loadProjects = async () => {
  loading.value = true
  try {
    const { data } = await getProjectList({
      page: page.value,
      page_size: pageSize.value,
      keyword: keyword.value || undefined,
    })
    projectList.value = data.data.items
    total.value = data.data.total
  } catch (error) {
    console.error('Failed to load projects:', error)
  } finally {
    loading.value = false
  }
}

// 加载用户列表
const loadUsers = async () => {
  try {
    const { data } = await getAllUsers()
    users.value = data.data
  } catch (error) {
    console.error('Failed to load users:', error)
  }
}

// 搜索
const handleSearch = () => {
  page.value = 1
  loadProjects()
}

// 查看项目
const handleViewProject = (project: Project) => {
  router.push(`/issues?project_id=${project.id}`)
}

// 创建项目
const handleCreate = () => {
  isEdit.value = false
  editingProject.value = null
  Object.assign(form, {
    project_key: '',
    name: '',
    description: '',
    lead_user_id: undefined,
  })
  dialogVisible.value = true
}

// 提交表单
const submitForm = async () => {
  if (!formRef.value) return

  await formRef.value.validate(async (valid) => {
    if (!valid) return

    submitLoading.value = true
    try {
      if (isEdit.value && editingProject.value) {
        await updateProject(editingProject.value.project_key, {
          name: form.name,
          description: form.description,
          lead_user_id: form.lead_user_id,
        })
        ElMessage.success('更新成功')
      } else {
        await createProject(form)
        ElMessage.success('创建成功')
      }
      dialogVisible.value = false
      loadProjects()
    } catch (error) {
      console.error('Failed to submit:', error)
    } finally {
      submitLoading.value = false
    }
  })
}

// 初始化
onMounted(() => {
  loadProjects()
  loadUsers()
})
</script>

<style scoped lang="scss">
.project-list-container {
  .toolbar-card {
    margin-bottom: 20px;
  }

  .toolbar {
    display: flex;
    justify-content: space-between;
    align-items: center;

    .toolbar-left {
      display: flex;
      gap: 12px;
    }
  }

  .project-grid {
    .project-col {
      margin-bottom: 20px;
    }
  }

  .project-card {
    cursor: pointer;
    height: 100%;
    transition: transform 0.2s, box-shadow 0.2s;

    &:hover {
      transform: translateY(-4px);
      box-shadow: 0 8px 24px rgba(0, 0, 0, 0.12);
    }

    .project-header {
      display: flex;
      align-items: center;
      gap: 12px;
      margin-bottom: 12px;

      .project-icon {
        width: 48px;
        height: 48px;
        background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
        border-radius: 8px;
        display: flex;
        align-items: center;
        justify-content: center;
        color: #fff;
        font-weight: 700;
        font-size: 16px;
      }

      .project-info {
        .project-name {
          font-size: 16px;
          font-weight: 600;
          color: #1f2937;
        }

        .project-key {
          font-size: 12px;
          color: #909399;
          margin-top: 4px;
        }
      }
    }

    .project-description {
      font-size: 14px;
      color: #6b7280;
      margin-bottom: 16px;
      display: -webkit-box;
      -webkit-line-clamp: 2;
      -webkit-box-orient: vertical;
      overflow: hidden;
      min-height: 42px;
    }

    .project-stats {
      display: flex;
      gap: 20px;
      margin-bottom: 16px;
      padding-bottom: 16px;
      border-bottom: 1px solid #f0f0f0;

      .stat-item {
        display: flex;
        align-items: center;
        gap: 4px;
        font-size: 13px;
        color: #6b7280;
      }
    }

    .project-footer {
      display: flex;
      justify-content: space-between;
      align-items: center;

      .project-lead {
        display: flex;
        align-items: center;
        gap: 4px;
        font-size: 13px;
        color: #6b7280;
      }
    }
  }

  .empty-state {
    padding: 60px 0;
  }

  .pagination {
    margin-top: 20px;
    justify-content: center;
  }

  .form-tip {
    font-size: 12px;
    color: #909399;
    margin-top: 4px;
  }
}
</style>
