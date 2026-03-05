<template>
  <div class="project-list-container">
    <!-- 页面头部 -->
    <div class="page-header">
      <div class="header-info">
        <div class="header-icon">
          <el-icon><Folder /></el-icon>
        </div>
        <div class="header-text">
          <h1 class="header-title">项目管理</h1>
          <p class="header-desc">管理所有项目和工作空间</p>
        </div>
      </div>
      <el-button type="primary" @click="handleCreate" class="header-btn">
        <el-icon><Plus /></el-icon>
        创建项目
      </el-button>
    </div>

    <!-- 搜索栏 -->
    <el-card shadow="never" class="filter-card">
      <div class="filter-content">
        <el-input
          v-model="keyword"
          placeholder="搜索项目名称或标识"
          clearable
          class="search-input"
          @clear="handleSearch"
          @keyup.enter="handleSearch"
        >
          <template #prefix>
            <el-icon><Search /></el-icon>
          </template>
        </el-input>
        <el-button :icon="Search" type="primary" @click="handleSearch">搜索</el-button>
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
          <div class="project-card">
            <div class="card-body" @click="handleViewProject(project)">
              <div class="project-header">
                <div class="project-icon" :style="{ background: getProjectColor(project.project_key) }">
                  {{ project.project_key?.substring(0, 2).toUpperCase() || '??' }}
                </div>
                <el-tag
                  :type="project.status === 1 ? 'success' : 'info'"
                  size="small"
                  effect="plain"
                  class="status-tag"
                >
                  {{ project.status === 1 ? '启用' : '禁用' }}
                </el-tag>
              </div>
              <div class="project-name">{{ project.name }}</div>
              <div class="project-key-text">{{ project.project_key }}</div>
              <div class="project-description">
                {{ project.description || '暂无描述' }}
              </div>
            </div>
            <div class="card-footer">
              <div class="footer-left">
                <div class="lead-info">
                  <div class="mini-avatar" v-if="project.lead_user">
                    {{ project.lead_user.display_name?.charAt(0) }}
                  </div>
                  <span class="lead-name">{{ project.lead_user?.display_name || '未指定' }}</span>
                </div>
              </div>
              <div class="footer-right">
                <div class="member-count">
                  <el-icon><User /></el-icon>
                  <span>{{ project.member_count || 0 }}</span>
                </div>
                <router-link :to="`/projects/${project.project_key}/settings`" class="settings-link">
                  <el-button size="small" type="primary" text>
                    <el-icon><Setting /></el-icon>
                    设置
                  </el-button>
                </router-link>
              </div>
            </div>
          </div>
        </el-col>
      </el-row>

      <div v-if="projectList.length === 0 && !loading" class="empty-state">
        <el-empty description="暂无项目">
          <el-button type="primary" @click="handleCreate">
            <el-icon><Plus /></el-icon>
            创建第一个项目
          </el-button>
        </el-empty>
      </div>

      <div v-if="total > 0" class="pagination-wrapper">
        <el-pagination
          v-model:current-page="page"
          v-model:page-size="pageSize"
          :total="total"
          :page-sizes="[12, 24, 48]"
          layout="total, sizes, prev, pager, next"
          @size-change="loadProjects"
          @current-change="loadProjects"
        />
      </div>
    </div>

    <!-- 创建/编辑项目对话框 -->
    <el-dialog
      v-model="dialogVisible"
      :title="isEdit ? '编辑项目' : '创建项目'"
      width="520px"
      destroy-on-close
      class="project-dialog"
    >
      <el-form ref="formRef" :model="form" :rules="rules" label-position="top">
        <el-row :gutter="16">
          <el-col :span="12">
            <el-form-item label="项目标识" prop="project_key">
              <el-input
                v-model="form.project_key"
                placeholder="如: PROJ"
                :disabled="isEdit"
                maxlength="10"
              >
                <template #prefix>
                  <el-icon><Key /></el-icon>
                </template>
              </el-input>
              <div class="form-tip">大写字母开头，创建后不可修改</div>
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="项目名称" prop="name">
              <el-input v-model="form.name" placeholder="请输入项目名称" maxlength="50">
                <template #prefix>
                  <el-icon><Folder /></el-icon>
                </template>
              </el-input>
            </el-form-item>
          </el-col>
        </el-row>
        <el-form-item v-if="!isEdit" label="项目模版">
          <div class="template-cards">
            <div
              class="template-card"
              :class="{ active: form.template === 'standard' }"
              @click="form.template = 'standard'"
            >
              <div class="template-card-icon">
                <el-icon :size="22"><Folder /></el-icon>
              </div>
              <div class="template-card-body">
                <div class="template-card-title">标准模版</div>
                <div class="template-card-desc">预置工单类型、工作流、字段方案</div>
              </div>
            </div>
            <div
              class="template-card"
              :class="{ active: form.template === 'blank' }"
              @click="form.template = 'blank'"
            >
              <div class="template-card-icon">
                <el-icon :size="22"><Plus /></el-icon>
              </div>
              <div class="template-card-body">
                <div class="template-card-title">空项目</div>
                <div class="template-card-desc">不预置配置，完全自定义</div>
              </div>
            </div>
          </div>
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
          <el-icon><Check /></el-icon>
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
import { Search, Plus, User, Folder, Key, Check, Setting } from '@element-plus/icons-vue'
import { getProjectList, createProject, updateProject } from '@/api/project'
import { getAllUsers } from '@/api/user'
import type { Project, CreateProjectRequest } from '@/types/project'
import type { UserOption } from '@/types/user'

const router = useRouter()

const loading = ref(false)
const projectList = ref<Project[]>([])
const total = ref(0)
const page = ref(1)
const pageSize = ref(12)
const keyword = ref('')
const users = ref<UserOption[]>([])

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
  template: 'standard',
})

const rules: FormRules = {
  project_key: [
    { required: true, message: '请输入项目标识', trigger: 'blur' },
    { pattern: /^[A-Z][A-Z0-9]*$/, message: '必须以大写字母开头，只含大写字母和数字', trigger: 'blur' },
    { min: 2, max: 10, message: '长度 2-10 个字符', trigger: 'blur' },
  ],
  name: [
    { required: true, message: '请输入项目名称', trigger: 'blur' },
    { max: 50, message: '最多 50 个字符', trigger: 'blur' },
  ],
}

const projectColors = [
  '#3b82f6',
  '#ef4444',
  '#3b82f6',
  '#10b981',
  '#f59e0b',
  '#8b5cf6',
]

const getProjectColor = (key: string) => {
  let hash = 0
  for (let i = 0; i < key.length; i++) hash = key.charCodeAt(i) + ((hash << 5) - hash)
  return projectColors[Math.abs(hash) % projectColors.length]
}

const loadProjects = async () => {
  loading.value = true
  try {
    const { data } = await getProjectList({
      page: page.value, page_size: pageSize.value, keyword: keyword.value || undefined,
    })
    projectList.value = data.data.items
    total.value = data.data.total
  } catch (error) {
    console.error('Failed to load projects:', error)
  } finally {
    loading.value = false
  }
}

const loadUsers = async () => {
  try { const { data } = await getAllUsers(); users.value = data.data } catch (e) { console.error(e) }
}

const handleSearch = () => { page.value = 1; loadProjects() }

const handleViewProject = (project: Project) => {
  router.push(`/issues?project_key=${project.project_key}`)
}

const handleCreate = () => {
  isEdit.value = false
  editingProject.value = null
  Object.assign(form, { project_key: '', name: '', description: '', lead_user_id: undefined, template: 'standard' })
  dialogVisible.value = true
}

const submitForm = async () => {
  if (!formRef.value) return
  await formRef.value.validate(async (valid) => {
    if (!valid) return
    submitLoading.value = true
    try {
      if (isEdit.value && editingProject.value) {
        await updateProject(editingProject.value.project_key, {
          name: form.name, description: form.description, lead_user_id: form.lead_user_id,
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

onMounted(() => { loadProjects(); loadUsers() })
</script>

<style scoped lang="scss">
.project-list-container {
  width: 100%;
}

// 页面头部
.page-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 24px;
  padding: 24px 32px;
  background: #3b82f6;
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

// 筛选
.filter-card {
  margin-bottom: 20px;
  border-radius: 12px;

  :deep(.el-card__body) { padding: 16px 20px; }

  .filter-content {
    display: flex;
    gap: 12px;
  }

  .search-input { width: 320px; }
}

// 项目卡片
.project-grid {
  .project-col { margin-bottom: 20px; }
}

.project-card {
  background: #fff;
  border-radius: 12px;
  height: 100%;
  cursor: pointer;
  border: 1px solid #f0f0f0;
  display: flex;
  flex-direction: column;
  transition: transform 0.3s, box-shadow 0.3s;
  overflow: hidden;

  &:hover {
    transform: translateY(-4px);
    box-shadow: 0 8px 24px rgba(0, 0, 0, 0.1);
  }

  .card-body {
    padding: 20px;
    flex: 1;
  }

  .project-header {
    display: flex;
    justify-content: space-between;
    align-items: flex-start;
    margin-bottom: 16px;
  }

  .project-icon {
    width: 48px;
    height: 48px;
    border-radius: 12px;
    display: flex;
    align-items: center;
    justify-content: center;
    color: #fff;
    font-weight: 700;
    font-size: 16px;
  }

  .project-name {
    font-size: 16px;
    font-weight: 600;
    color: #1f2937;
    margin-bottom: 4px;
  }

  .project-key-text {
    font-size: 12px;
    color: #9ca3af;
    margin-bottom: 12px;
  }

  .project-description {
    font-size: 13px;
    color: #6b7280;
    display: -webkit-box;
    -webkit-line-clamp: 2;
    -webkit-box-orient: vertical;
    overflow: hidden;
    min-height: 38px;
    line-height: 1.5;
  }

  .card-footer {
    display: flex;
    justify-content: space-between;
    align-items: center;
    padding: 14px 20px;
    border-top: 1px solid #f5f5f5;
    background: #fafafa;
  }

  .lead-info {
    display: flex;
    align-items: center;
    gap: 8px;

    .lead-name {
      font-size: 13px;
      color: #6b7280;
    }
  }

  .mini-avatar {
    width: 24px;
    height: 24px;
    border-radius: 6px;
    background: #3b82f6;
    color: #fff;
    display: flex;
    align-items: center;
    justify-content: center;
    font-size: 11px;
    font-weight: 600;
    flex-shrink: 0;
  }

  .member-count {
    display: flex;
    align-items: center;
    gap: 4px;
    font-size: 13px;
    color: #9ca3af;
    margin-right: 8px;
  }

  .footer-right {
    display: flex;
    align-items: center;
    gap: 8px;
  }
}

.empty-state { padding: 60px 0; }

.pagination-wrapper {
  margin-top: 20px;
  display: flex;
  justify-content: center;
}

.form-tip {
  font-size: 12px;
  color: #9ca3af;
  margin-top: 4px;
}

.template-cards {
  display: flex;
  gap: 12px;
  width: 100%;
}

.template-card {
  flex: 1;
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 14px 16px;
  border: 1.5px solid #e5e7eb;
  border-radius: 8px;
  cursor: pointer;
  transition: all 0.2s;

  &:hover {
    border-color: #a0cfff;
    background: #f0f7ff;
  }

  &.active {
    border-color: #409eff;
    background: #ecf5ff;
  }
}

.template-card-icon {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 40px;
  height: 40px;
  border-radius: 8px;
  background: #f3f4f6;
  color: #6b7280;
  flex-shrink: 0;

  .template-card.active & {
    background: #d9ecff;
    color: #409eff;
  }
}

.template-card-title {
  font-size: 14px;
  font-weight: 600;
  color: #1f2937;
  line-height: 1.4;
}

.template-card-desc {
  font-size: 12px;
  color: #9ca3af;
  line-height: 1.4;
  margin-top: 2px;
}

.settings-link {
  text-decoration: none;
  color: inherit;
}

// 响应式
@media (max-width: 768px) {
  .page-header {
    flex-direction: column;
    align-items: flex-start;
    gap: 16px;
  }

  .filter-card .search-input { width: 100%; }
}
</style>
