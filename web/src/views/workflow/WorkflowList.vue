<template>
  <div class="workflow-list-container">
    <!-- 页面头部 -->
    <div class="page-header">
      <div class="header-info">
        <div class="header-icon">
          <el-icon><Connection /></el-icon>
        </div>
        <div class="header-text">
          <h1 class="header-title">工作流管理</h1>
          <p class="header-desc">配置和管理工作流程</p>
        </div>
      </div>
      <el-button type="primary" @click="handleCreate" class="header-btn">
        <el-icon><Plus /></el-icon>
        创建工作流
      </el-button>
    </div>

    <!-- 工作流列表 -->
    <el-card shadow="never" class="workflows-card">
      <div v-loading="loading" class="workflows-list">
        <div v-for="workflow in workflows" :key="workflow.id" class="workflow-item">
          <div class="workflow-header">
            <div class="workflow-info">
              <div class="workflow-icon">
                <el-icon><Share /></el-icon>
              </div>
              <div class="workflow-details">
                <h3 class="workflow-name">{{ workflow.name }}</h3>
                <p v-if="workflow.description" class="workflow-desc">{{ workflow.description }}</p>
                <div class="workflow-meta">
                  <el-tag size="small" :type="workflow.is_active ? 'success' : 'info'">
                    {{ workflow.is_active ? '启用' : '禁用' }}
                  </el-tag>
                  <span class="workflow-project">项目: {{ workflow.project_key || '全局' }}</span>
                </div>
              </div>
            </div>
            <div class="workflow-actions">
              <el-button size="small" @click="handleViewNodes(workflow)">
                <el-icon><View /></el-icon>
                查看节点
              </el-button>
              <el-button size="small" @click="handleEdit(workflow)">
                <el-icon><Edit /></el-icon>
                编辑
              </el-button>
              <el-button size="small" type="danger" @click="handleDelete(workflow)">
                <el-icon><Delete /></el-icon>
                删除
              </el-button>
            </div>
          </div>
        </div>
        <el-empty v-if="!loading && workflows.length === 0" description="暂无工作流" />
      </div>
    </el-card>

    <!-- 创建/编辑工作流对话框 -->
    <el-dialog v-model="dialogVisible" :title="isEditMode ? '编辑工作流' : '创建工作流'" width="600px" destroy-on-close>
      <el-form ref="formRef" :model="form" :rules="formRules" label-position="top">
        <el-form-item label="工作流名称" prop="name">
          <el-input v-model="form.name" placeholder="请输入工作流名称" />
        </el-form-item>
        <el-form-item label="描述">
          <el-input v-model="form.description" type="textarea" :rows="3" placeholder="工作流描述" />
        </el-form-item>
        <el-form-item label="所属项目">
          <el-select v-model="form.project_key" placeholder="选择项目（留空为全局）" clearable style="width: 100%">
            <el-option v-for="p in projects" :key="p.project_key" :label="p.name" :value="p.project_key" />
          </el-select>
        </el-form-item>
        <el-form-item label="状态">
          <el-switch v-model="form.is_active" active-text="启用" inactive-text="禁用" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="submitLoading" @click="submitForm">确定</el-button>
      </template>
    </el-dialog>

    <!-- 节点查看对话框 -->
    <el-dialog v-model="nodesDialogVisible" :title="`${currentWorkflow?.name} - 节点列表`" width="800px" destroy-on-close>
      <div v-loading="nodesLoading" class="nodes-content">
        <el-button type="primary" size="small" @click="handleAddNode" style="margin-bottom: 16px">
          <el-icon><Plus /></el-icon>
          添加节点
        </el-button>
        <div class="nodes-list">
          <div v-for="node in workflowNodes" :key="node.id" class="node-item">
            <div class="node-info">
              <div class="node-icon" :class="node.node_type">
                <el-icon v-if="node.node_type === 'start'"><VideoPlay /></el-icon>
                <el-icon v-else-if="node.node_type === 'end'"><CircleCheck /></el-icon>
                <el-icon v-else-if="node.node_type === 'approval'"><Checked /></el-icon>
                <el-icon v-else><Operation /></el-icon>
              </div>
              <div class="node-details">
                <div class="node-title-row">
                  <span class="node-name">{{ node.name }}</span>
                  <el-tag size="small">{{ getNodeTypeText(node.node_type) }}</el-tag>
                </div>
                <p v-if="node.description" class="node-desc">{{ node.description }}</p>
              </div>
            </div>
            <div class="node-actions">
              <el-button size="small" @click="handleEditNode(node)">编辑</el-button>
              <el-button size="small" type="danger" @click="handleDeleteNode(node)">删除</el-button>
            </div>
          </div>
          <el-empty v-if="!nodesLoading && workflowNodes.length === 0" description="暂无节点" />
        </div>
      </div>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { ElMessage, ElMessageBox, type FormInstance, type FormRules } from 'element-plus'
import {
  Connection,
  Plus,
  Share,
  View,
  Edit,
  Delete,
  VideoPlay,
  CircleCheck,
  Checked,
  Operation,
} from '@element-plus/icons-vue'
import { getAllProjects } from '@/api/project'
import type { Project } from '@/types/project'

// 临时类型定义（后续需要在 types/workflow.ts 中定义）
interface Workflow {
  id: number
  name: string
  description: string
  project_key?: string
  is_active: boolean
  created_at: string
  updated_at: string
}

interface WorkflowNode {
  id: number
  workflow_id: number
  name: string
  node_type: string
  description: string
  position_x: number
  position_y: number
}

const loading = ref(false)
const workflows = ref<Workflow[]>([])
const projects = ref<Project[]>([])

// 工作流表单
const dialogVisible = ref(false)
const isEditMode = ref(false)
const editingId = ref<number | null>(null)
const submitLoading = ref(false)
const formRef = ref<FormInstance>()
const form = reactive({
  name: '',
  description: '',
  project_key: undefined as string | undefined,
  is_active: true,
})
const formRules: FormRules = {
  name: [{ required: true, message: '请输入工作流名称', trigger: 'blur' }],
}

// 节点管理
const nodesDialogVisible = ref(false)
const currentWorkflow = ref<Workflow | null>(null)
const workflowNodes = ref<WorkflowNode[]>([])
const nodesLoading = ref(false)

const loadWorkflows = async () => {
  loading.value = true
  try {
    // TODO: 调用实际的工作流列表 API
    // const { data } = await getWorkflowList()
    // workflows.value = data.data
    workflows.value = []
    ElMessage.info('工作流列表功能开发中')
  } catch (error) {
    console.error('Failed to load workflows:', error)
  } finally {
    loading.value = false
  }
}

const loadProjects = async () => {
  try {
    const { data } = await getAllProjects()
    projects.value = data.data
  } catch (error) {
    console.error('Failed to load projects:', error)
  }
}

const handleCreate = () => {
  isEditMode.value = false
  editingId.value = null
  Object.assign(form, { name: '', description: '', project_key: undefined, is_active: true })
  dialogVisible.value = true
}

const handleEdit = (workflow: Workflow) => {
  isEditMode.value = true
  editingId.value = workflow.id
  Object.assign(form, {
    name: workflow.name,
    description: workflow.description,
    project_key: workflow.project_key,
    is_active: workflow.is_active,
  })
  dialogVisible.value = true
}

const submitForm = async () => {
  if (!formRef.value) return
  await formRef.value.validate(async (valid) => {
    if (!valid) return
    submitLoading.value = true
    try {
      // TODO: 调用实际的创建/更新 API
      ElMessage.success(isEditMode.value ? '更新成功' : '创建成功')
      dialogVisible.value = false
      loadWorkflows()
    } catch (error) {
      console.error('Failed to save workflow:', error)
    } finally {
      submitLoading.value = false
    }
  })
}

const handleDelete = async (workflow: Workflow) => {
  try {
    await ElMessageBox.confirm(`确定要删除工作流 "${workflow.name}" 吗？`, '删除确认', {
      type: 'warning',
    })
    // TODO: 调用实际的删除 API
    ElMessage.success('删除成功')
    loadWorkflows()
  } catch (error) {
    if (error !== 'cancel') {
      console.error('Failed to delete workflow:', error)
    }
  }
}

const handleViewNodes = async (workflow: Workflow) => {
  currentWorkflow.value = workflow
  nodesDialogVisible.value = true
  await loadNodes(workflow.id)
}

const loadNodes = async (workflowId: number) => {
  nodesLoading.value = true
  try {
    // TODO: 调用实际的节点列表 API
    workflowNodes.value = []
  } catch (error) {
    console.error('Failed to load nodes:', error)
  } finally {
    nodesLoading.value = false
  }
}

const handleAddNode = () => {
  ElMessage.info('节点添加功能开发中')
}

const handleEditNode = (node: WorkflowNode) => {
  ElMessage.info('节点编辑功能开发中')
}

const handleDeleteNode = async (node: WorkflowNode) => {
  try {
    await ElMessageBox.confirm(`确定要删除节点 "${node.name}" 吗？`, '删除确认', {
      type: 'warning',
    })
    ElMessage.success('删除成功')
    if (currentWorkflow.value) {
      await loadNodes(currentWorkflow.value.id)
    }
  } catch (error) {
    if (error !== 'cancel') {
      console.error('Failed to delete node:', error)
    }
  }
}

const getNodeTypeText = (type: string) => {
  const map: Record<string, string> = {
    start: '开始',
    end: '结束',
    approval: '审批',
    work: '工作',
  }
  return map[type] || type
}

onMounted(() => {
  loadWorkflows()
  loadProjects()
})
</script>

<style scoped lang="scss">
.workflow-list-container {
  width: 100%;
}

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
    color: #fff;
    &:hover {
      background: rgba(255, 255, 255, 0.3);
    }
  }
}

.workflows-card {
  border-radius: 12px;

  :deep(.el-card__body) {
    padding: 24px;
  }
}

.workflows-list {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.workflow-item {
  background: #f8fafc;
  border-radius: 12px;
  padding: 20px;
  transition: box-shadow 0.2s;

  &:hover {
    box-shadow: 0 4px 12px rgba(0, 0, 0, 0.08);
  }
}

.workflow-header {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
}

.workflow-info {
  display: flex;
  gap: 16px;
  flex: 1;
}

.workflow-icon {
  width: 48px;
  height: 48px;
  border-radius: 12px;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 24px;
  color: #fff;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
}

.workflow-details {
  flex: 1;

  .workflow-name {
    font-size: 16px;
    font-weight: 600;
    color: #1f2937;
    margin: 0 0 4px 0;
  }

  .workflow-desc {
    font-size: 13px;
    color: #6b7280;
    margin: 0 0 8px 0;
  }

  .workflow-meta {
    display: flex;
    align-items: center;
    gap: 12px;
  }

  .workflow-project {
    font-size: 12px;
    color: #9ca3af;
  }
}

.workflow-actions {
  display: flex;
  gap: 8px;
}

// 节点列表
.nodes-content {
  min-height: 300px;
}

.nodes-list {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.node-item {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 16px;
  background: #f8fafc;
  border-radius: 8px;

  &:hover {
    background: #f1f5f9;
  }
}

.node-info {
  display: flex;
  gap: 12px;
  flex: 1;
}

.node-icon {
  width: 40px;
  height: 40px;
  border-radius: 8px;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 20px;
  color: #fff;

  &.start {
    background: linear-gradient(135deg, #10b981 0%, #059669 100%);
  }
  &.end {
    background: linear-gradient(135deg, #6b7280 0%, #4b5563 100%);
  }
  &.approval {
    background: linear-gradient(135deg, #f59e0b 0%, #d97706 100%);
  }
  &.work {
    background: linear-gradient(135deg, #3b82f6 0%, #2563eb 100%);
  }
}

.node-details {
  flex: 1;

  .node-title-row {
    display: flex;
    align-items: center;
    gap: 8px;
    margin-bottom: 4px;
  }

  .node-name {
    font-size: 14px;
    font-weight: 500;
    color: #1f2937;
  }

  .node-desc {
    font-size: 12px;
    color: #9ca3af;
    margin: 0;
  }
}

.node-actions {
  display: flex;
  gap: 8px;
}
</style>
