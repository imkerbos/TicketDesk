<template>
  <div class="project-board-container">
    <!-- 页面头部 -->
    <div class="board-header">
      <div class="header-left">
        <el-breadcrumb separator="/">
          <el-breadcrumb-item :to="{ path: '/projects' }">项目列表</el-breadcrumb-item>
          <el-breadcrumb-item :to="{ path: `/projects/${projectKey}` }">{{ project?.name || projectKey }}</el-breadcrumb-item>
          <el-breadcrumb-item>工单看板</el-breadcrumb-item>
        </el-breadcrumb>
      </div>
      <div class="project-nav">
        <router-link
          :to="`/projects/${projectKey}`"
          class="nav-item"
        >
          概览
        </router-link>
        <router-link
          :to="`/projects/${projectKey}/board`"
          class="nav-item active"
        >
          看板
        </router-link>
        <router-link
          :to="`/projects/${projectKey}/settings`"
          class="nav-item"
        >
          设置
        </router-link>
      </div>
    </div>

    <!-- 分屏主体 -->
    <div class="board-body">
      <!-- 左侧工单列表 -->
      <div class="board-left">
        <BoardIssueList
          :project-key="projectKey"
          :selected-key="selectedIssueKey"
          @select="handleSelectIssue"
          @create="handleCreateIssue"
        />
      </div>

      <!-- 右侧详情 -->
      <div class="board-right">
        <IssueDetail
          v-if="selectedIssueKey"
          :key="selectedIssueKey"
          embedded
          :issue-key="selectedIssueKey"
          :on-navigate-issue="handleNavigateIssue"
          :on-deleted="handleDeleted"
        />
        <div v-else class="empty-detail">
          <el-empty description="请从左侧选择工单" :image-size="120" />
        </div>
      </div>
    </div>

    <!-- 创建工单对话框 -->
    <el-dialog v-model="createDialogVisible" title="创建工单" width="640px" destroy-on-close class="create-dialog">
      <el-form ref="createFormRef" :model="createForm" :rules="createRules" label-position="top">
        <el-row :gutter="20">
          <el-col :span="12">
            <el-form-item label="项目">
              <el-input :model-value="projectKey" disabled />
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="类型" prop="issue_type_id">
              <el-select v-model="createForm.issue_type_id" placeholder="请选择类型" style="width: 100%" @change="handleIssueTypeChange">
                <el-option v-for="t in issueTypes" :key="t.id" :label="t.display_name" :value="t.id" />
              </el-select>
            </el-form-item>
          </el-col>
        </el-row>
        <el-form-item label="标题" prop="title">
          <el-input v-model="createForm.title" placeholder="请输入工单标题" maxlength="200" show-word-limit />
        </el-form-item>
        <el-form-item label="描述">
          <el-input v-model="createForm.description" type="textarea" :rows="3" placeholder="请输入描述" />
        </el-form-item>
        <el-row :gutter="20">
          <el-col :span="12">
            <el-form-item label="优先级">
              <el-select v-model="createForm.priority" style="width: 100%">
                <el-option label="P0 - 紧急" value="P0" />
                <el-option label="P1 - 高" value="P1" />
                <el-option label="P2 - 中" value="P2" />
                <el-option label="P3 - 低" value="P3" />
              </el-select>
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="指派人">
              <el-select v-model="createForm.assignee_id" placeholder="请选择" style="width: 100%" filterable clearable>
                <el-option v-for="u in users" :key="u.id" :label="u.display_name" :value="u.id" />
              </el-select>
            </el-form-item>
          </el-col>
        </el-row>
      </el-form>
      <template #footer>
        <el-button @click="createDialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="createLoading" @click="submitCreate">
          <el-icon><Check /></el-icon>
          创建
        </el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, computed, onMounted, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage, type FormInstance, type FormRules } from 'element-plus'
import { Check } from '@element-plus/icons-vue'
import BoardIssueList from './components/BoardIssueList.vue'
import IssueDetail from '@/views/issue/IssueDetail.vue'
import { getProjectDetail, getProjectIssueTypes } from '@/api/project'
import { createIssue } from '@/api/issue'
import { getAllUsers } from '@/api/user'
import type { Project, ProjectIssueType } from '@/types/project'
import type { UserOption } from '@/types/user'
import type { IssuePriority } from '@/types/issue'

const route = useRoute()
const router = useRouter()

const projectKey = computed(() => route.params.key as string)
const selectedIssueKey = ref('')
const project = ref<Project | null>(null)

// 创建工单
const createDialogVisible = ref(false)
const createLoading = ref(false)
const createFormRef = ref<FormInstance>()
const issueTypes = ref<ProjectIssueType[]>([])
const users = ref<UserOption[]>([])

const createForm = reactive({
  issue_type_id: undefined as number | undefined,
  title: '',
  description: '',
  priority: 'P2' as IssuePriority,
  assignee_id: undefined as number | undefined,
})

const createRules: FormRules = {
  issue_type_id: [{ required: true, message: '请选择类型', trigger: 'change' }],
  title: [{ required: true, message: '请输入标题', trigger: 'blur' }],
}

// 初始化选中工单
const initSelectedKey = () => {
  const issueKey = route.params.issueKey as string
  selectedIssueKey.value = issueKey || ''
}

const loadProject = async () => {
  try {
    const { data } = await getProjectDetail(projectKey.value)
    project.value = data.data
  } catch (e) {
    console.error('Failed to load project:', e)
  }
}

const handleSelectIssue = (issueKey: string) => {
  selectedIssueKey.value = issueKey
  router.replace(`/projects/${projectKey.value}/board/${issueKey}`)
}

const handleNavigateIssue = (issueKey: string) => {
  selectedIssueKey.value = issueKey
  router.replace(`/projects/${projectKey.value}/board/${issueKey}`)
}

const handleDeleted = () => {
  selectedIssueKey.value = ''
  router.replace(`/projects/${projectKey.value}/board`)
}

const handleCreateIssue = async () => {
  // 加载工单类型和用户
  try {
    const [typesRes, usersRes] = await Promise.all([
      getProjectIssueTypes(projectKey.value),
      getAllUsers(),
    ])
    issueTypes.value = typesRes.data.data
    users.value = usersRes.data.data
  } catch (e) {
    console.error('Failed to load options:', e)
    ElMessage.error('加载选项失败')
    return
  }
  // 重置表单
  Object.assign(createForm, {
    issue_type_id: undefined,
    title: '',
    description: '',
    priority: 'P2' as IssuePriority,
    assignee_id: undefined,
  })
  createDialogVisible.value = true
}

const handleIssueTypeChange = () => {
  // 预留：未来可在此加载字段方案
}

const submitCreate = async () => {
  if (!createFormRef.value) return
  await createFormRef.value.validate(async (valid) => {
    if (!valid) return
    createLoading.value = true
    try {
      const { data } = await createIssue({
        project_key: projectKey.value,
        issue_type_id: createForm.issue_type_id!,
        title: createForm.title,
        description: createForm.description,
        priority: createForm.priority,
        assignee_id: createForm.assignee_id,
      })
      ElMessage.success('工单创建成功')
      createDialogVisible.value = false
      // 选中新创建的工单
      const newKey = data.data?.issue_key
      if (newKey) {
        handleSelectIssue(newKey)
      }
    } catch (e) {
      console.error('Failed to create issue:', e)
      ElMessage.error('创建工单失败')
    } finally {
      createLoading.value = false
    }
  })
}

// 浏览器前进/后退
watch(
  () => route.params.issueKey,
  (newKey) => {
    const key = newKey as string
    if (key !== selectedIssueKey.value) {
      selectedIssueKey.value = key || ''
    }
  }
)

onMounted(() => {
  initSelectedKey()
  loadProject()
})
</script>

<style scoped lang="scss">
.project-board-container {
  display: flex;
  flex-direction: column;
  height: calc(100vh - 64px);
  margin: -24px;
  background: #f5f7fa;
}

.board-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 16px 24px;
  background: #fff;
  border-bottom: 1px solid #e5e7eb;
  flex-shrink: 0;
}

.project-nav {
  display: flex;
  gap: 4px;

  .nav-item {
    padding: 6px 16px;
    border-radius: 6px;
    font-size: 14px;
    color: #6b7280;
    text-decoration: none;
    transition: background-color 150ms ease-out, color 150ms ease-out;

    &:hover {
      background: #f3f4f6;
      color: #1f2937;
    }

    &.active,
    &.router-link-exact-active {
      background: #eff6ff;
      color: #3b82f6;
      font-weight: 500;
    }
  }
}

.board-body {
  display: flex;
  flex: 1;
  overflow: hidden;
}

.board-left {
  width: 380px;
  min-width: 380px;
  border-right: 1px solid #e5e7eb;
  overflow-y: auto;
}

.board-right {
  flex: 1;
  overflow-y: auto;
  background: #f5f7fa;
}

.empty-detail {
  display: flex;
  justify-content: center;
  align-items: center;
  height: 100%;
  background: #fff;
}

@media (prefers-reduced-motion: reduce) {
  .nav-item {
    transition: none;
  }
}
</style>
