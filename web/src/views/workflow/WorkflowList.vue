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
                  <el-tag size="small" :type="workflow.status === 1 ? 'success' : 'info'">
                    {{ workflow.status === 1 ? '启用' : '禁用' }}
                  </el-tag>
                  <span class="workflow-project">项目: {{ workflow.project_id || '全局' }}</span>
                </div>
              </div>
            </div>
            <div class="workflow-actions">
              <el-button size="small" type="primary" @click="handleDesign(workflow)">
                <el-icon><EditPen /></el-icon>
                设计工作流
              </el-button>
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
          <el-select v-model="form.project_id" placeholder="选择项目（留空为全局）" clearable style="width: 100%">
            <el-option v-for="p in projects" :key="p.id" :label="p.name" :value="p.id" />
          </el-select>
        </el-form-item>
        <el-form-item label="状态">
          <el-switch v-model="form.status" :active-value="1" :inactive-value="0" active-text="启用" inactive-text="禁用" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="submitLoading" @click="submitForm">确定</el-button>
      </template>
    </el-dialog>

    <!-- 节点查看对话框 -->
    <el-dialog v-model="nodesDialogVisible" :title="`${currentWorkflow?.name} - 节点与边管理`" width="900px" destroy-on-close>
      <div v-loading="nodesLoading" class="nodes-content">
        <!-- 节点管理 -->
        <div class="section-header">
          <h4>节点列表</h4>
          <el-button type="primary" size="small" @click="handleAddNode">
            <el-icon><Plus /></el-icon>
            添加节点
          </el-button>
        </div>
        <div class="nodes-list">
          <div v-for="node in workflowNodes" :key="node.id" class="node-item">
            <div class="node-info">
              <div class="node-icon" :class="node.node_type">
                <el-icon v-if="node.node_type === 'start'"><VideoPlay /></el-icon>
                <el-icon v-else-if="node.node_type === 'end'"><CircleCheck /></el-icon>
                <el-icon v-else-if="node.node_type === 'approval'"><Checked /></el-icon>
                <el-icon v-else-if="node.node_type === 'system'"><Setting /></el-icon>
                <el-icon v-else><Operation /></el-icon>
              </div>
              <div class="node-details">
                <div class="node-title-row">
                  <span class="node-name">{{ node.name }}</span>
                  <el-tag size="small">{{ getNodeTypeText(node.node_type) }}</el-tag>
                  <el-tag v-if="node.config?.approval_type" size="small" type="warning">{{ getApprovalTypeText(node.config.approval_type) }}</el-tag>
                  <el-tag v-if="node.config?.assignee_type" size="small" type="success">{{ getAssigneeTypeText(node.config.assignee_type) }}</el-tag>
                </div>
                <p v-if="node.config?.description" class="node-desc">{{ node.config.description }}</p>
              </div>
            </div>
            <div class="node-actions">
              <el-button size="small" @click="handleEditNode(node)">编辑</el-button>
              <el-button size="small" type="danger" @click="handleDeleteNode(node)" :disabled="node.node_type === 'start' || node.node_type === 'end'">删除</el-button>
            </div>
          </div>
          <el-empty v-if="!nodesLoading && workflowNodes.length === 0" description="暂无节点" />
        </div>

        <!-- 边管理 -->
        <div class="section-header" style="margin-top: 24px;">
          <h4>边列表（流转路径）</h4>
          <el-button type="primary" size="small" @click="handleAddEdge" :disabled="workflowNodes.length < 2">
            <el-icon><Plus /></el-icon>
            添加边
          </el-button>
        </div>
        <div v-loading="edgesLoading" class="edges-list">
          <div v-for="edge in workflowEdges" :key="edge.id" class="edge-item">
            <div class="edge-info">
              <span class="edge-node">{{ getNodeName(edge.source_node_id) }}</span>
              <el-icon class="edge-arrow"><Right /></el-icon>
              <span class="edge-node">{{ getNodeName(edge.target_node_id) }}</span>
              <el-tag v-if="edge.condition_expr" size="small" type="info" class="edge-condition">{{ edge.condition_expr }}</el-tag>
            </div>
            <div class="edge-actions">
              <el-button size="small" type="danger" @click="handleDeleteEdge(edge)">删除</el-button>
            </div>
          </div>
          <el-empty v-if="!edgesLoading && workflowEdges.length === 0" description="暂无边" />
        </div>
      </div>
    </el-dialog>

    <!-- 添加/编辑节点对话框 -->
    <el-dialog v-model="nodeDialogVisible" :title="isEditNodeMode ? '编辑节点' : '添加节点'" width="600px" destroy-on-close>
      <el-form ref="nodeFormRef" :model="nodeForm" :rules="nodeFormRules" label-position="top">
        <el-form-item label="节点名称" prop="name">
          <el-input v-model="nodeForm.name" placeholder="请输入节点名称" />
        </el-form-item>
        <el-form-item v-if="!isEditNodeMode" label="节点类型" prop="node_type">
          <el-select v-model="nodeForm.node_type" placeholder="请选择节点类型" style="width: 100%" @change="handleNodeTypeChange">
            <el-option label="审批节点" value="approval" />
            <el-option label="工作节点" value="work" />
            <el-option label="系统节点" value="system" />
          </el-select>
        </el-form-item>
        <el-form-item v-else label="节点类型">
          <el-tag>{{ getNodeTypeText(nodeForm.node_type) }}</el-tag>
        </el-form-item>

        <!-- 审批节点配置 -->
        <template v-if="nodeForm.node_type === 'approval'">
          <el-form-item label="审批类型">
            <el-select v-model="nodeForm.config.approval_type" placeholder="请选择审批类型" style="width: 100%">
              <el-option label="单人审批" value="single" />
              <el-option label="会签（所有人通过）" value="countersign" />
              <el-option label="或签（任一人通过）" value="or_sign" />
            </el-select>
          </el-form-item>
          <el-form-item label="审批人">
            <el-select v-model="nodeForm.config.approvers" multiple placeholder="请选择审批人" style="width: 100%" filterable>
              <el-option v-for="u in allUsers" :key="u.id" :label="u.display_name" :value="u.id" />
            </el-select>
          </el-form-item>
        </template>

        <!-- 工作节点配置 -->
        <template v-if="nodeForm.node_type === 'work'">
          <el-form-item label="指派类型">
            <el-select v-model="nodeForm.config.assignee_type" placeholder="请选择指派类型" style="width: 100%">
              <el-option label="指定用户" value="user" />
              <el-option label="指定角色" value="role" />
              <el-option label="报告人" value="reporter" />
              <el-option label="项目负责人" value="project_lead" />
            </el-select>
          </el-form-item>
          <el-form-item v-if="nodeForm.config.assignee_type === 'user'" label="指派人">
            <el-select v-model="nodeForm.config.assignees" multiple placeholder="请选择指派人" style="width: 100%" filterable>
              <el-option v-for="u in allUsers" :key="u.id" :label="u.display_name" :value="u.id" />
            </el-select>
          </el-form-item>
          <el-form-item v-if="nodeForm.config.assignee_type === 'role'" label="角色名称">
            <el-input v-model="nodeForm.config.assignee_role" placeholder="请输入角色名称" />
          </el-form-item>
        </template>

        <!-- 系统节点配置 -->
        <template v-if="nodeForm.node_type === 'system'">
          <el-form-item label="系统动作">
            <el-input v-model="nodeForm.config.action" placeholder="请输入系统动作" />
          </el-form-item>
        </template>

        <!-- 通用配置 -->
        <el-form-item label="超时时间（小时）">
          <el-input-number v-model="nodeForm.config.timeout_hours" :min="0" :max="720" placeholder="0 表示不超时" />
        </el-form-item>
        <el-form-item label="节点说明">
          <el-input v-model="nodeForm.config.description" type="textarea" :rows="2" placeholder="节点说明" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="nodeDialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="nodeSubmitLoading" @click="submitNodeForm">确定</el-button>
      </template>
    </el-dialog>

    <!-- 添加边对话框 -->
    <el-dialog v-model="edgeDialogVisible" title="添加边" width="500px" destroy-on-close>
      <el-form ref="edgeFormRef" :model="edgeForm" :rules="edgeFormRules" label-position="top">
        <el-form-item label="源节点" prop="source_node_id">
          <el-select v-model="edgeForm.source_node_id" placeholder="请选择源节点" style="width: 100%">
            <el-option v-for="node in workflowNodes" :key="node.id" :label="`${node.name} (${getNodeTypeText(node.node_type)})`" :value="node.id" />
          </el-select>
        </el-form-item>
        <el-form-item label="目标节点" prop="target_node_id">
          <el-select v-model="edgeForm.target_node_id" placeholder="请选择目标节点" style="width: 100%">
            <el-option v-for="node in workflowNodes" :key="node.id" :label="`${node.name} (${getNodeTypeText(node.node_type)})`" :value="node.id" :disabled="node.id === edgeForm.source_node_id" />
          </el-select>
        </el-form-item>
        <el-form-item label="条件表达式（可选）">
          <el-input v-model="edgeForm.condition_expr" placeholder="如: status == 'approved'" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="edgeDialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="edgeSubmitLoading" @click="submitEdgeForm">确定</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage, ElMessageBox, type FormInstance, type FormRules } from 'element-plus'
import {
  Connection,
  Plus,
  Share,
  View,
  Edit,
  EditPen,
  Delete,
  VideoPlay,
  CircleCheck,
  Checked,
  Operation,
  Setting,
  Right,
} from '@element-plus/icons-vue'
import { getAllProjects } from '@/api/project'
import { getAllUsers } from '@/api/user'
import type { Project } from '@/types/project'
import type { UserOption } from '@/types/user'
import {
  getWorkflowList,
  createWorkflow,
  updateWorkflow,
  deleteWorkflow,
  getWorkflowNodes,
  createNode,
  updateNode,
  deleteNode,
  getWorkflowEdges,
  createEdge,
  deleteEdge,
} from '@/api/workflow'
import type { Workflow, WorkflowNode, WorkflowEdge, NodeConfig } from '@/types/workflow'

const loading = ref(false)
const workflows = ref<Workflow[]>([])
const projects = ref<Project[]>([])
const router = useRouter()
// 工作流表单
const dialogVisible = ref(false)
const isEditMode = ref(false)
const editingId = ref<number | null>(null)
const submitLoading = ref(false)
const formRef = ref<FormInstance>()
const form = reactive({
  name: '',
  description: '',
  project_id: undefined as number | undefined,
  status: 1,
})
const formRules: FormRules = {
  name: [{ required: true, message: '请输入工作流名称', trigger: 'blur' }],
}

// 节点管理
const nodesDialogVisible = ref(false)
const currentWorkflow = ref<Workflow | null>(null)
const workflowNodes = ref<WorkflowNode[]>([])
const nodesLoading = ref(false)
const allUsers = ref<UserOption[]>([])

// 边管理
const workflowEdges = ref<WorkflowEdge[]>([])
const edgesLoading = ref(false)

// 节点表单
const nodeDialogVisible = ref(false)
const isEditNodeMode = ref(false)
const editingNodeId = ref<number | null>(null)
const nodeSubmitLoading = ref(false)
const nodeFormRef = ref<FormInstance>()

const defaultNodeConfig = (): NodeConfig => ({
  approval_type: undefined,
  approvers: [],
  approver_role: '',
  assignee_type: undefined,
  assignees: [],
  assignee_role: '',
  action: '',
  parameters: {},
  timeout_hours: 0,
  description: '',
})

const nodeForm = reactive({
  name: '',
  node_type: '' as string,
  config: defaultNodeConfig(),
})
const nodeFormRules: FormRules = {
  name: [{ required: true, message: '请输入节点名称', trigger: 'blur' }],
  node_type: [{ required: true, message: '请选择节点类型', trigger: 'change' }],
}

// 边表单
const edgeDialogVisible = ref(false)
const edgeSubmitLoading = ref(false)
const edgeFormRef = ref<FormInstance>()
const edgeForm = reactive({
  source_node_id: undefined as number | undefined,
  target_node_id: undefined as number | undefined,
  condition_expr: '',
})
const edgeFormRules: FormRules = {
  source_node_id: [{ required: true, message: '请选择源节点', trigger: 'change' }],
  target_node_id: [{ required: true, message: '请选择目标节点', trigger: 'change' }],
}

const loadWorkflows = async () => {
  loading.value = true
  try {
    const { data } = await getWorkflowList()
    workflows.value = (data as any).data.items || []
  } catch (error) {
    console.error('Failed to load workflows:', error)
    ElMessage.error('加载工作流列表失败')
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

const handleDesign = (workflow: Workflow) => {
  router.push(`/workflows/${workflow.id}/designer`)
}

const handleCreate = () => {
  isEditMode.value = false
  editingId.value = null
  Object.assign(form, { name: '', description: '', project_id: undefined, status: 1 })
  dialogVisible.value = true
}

const handleEdit = (workflow: Workflow) => {
  isEditMode.value = true
  editingId.value = workflow.id
  Object.assign(form, {
    name: workflow.name,
    description: workflow.description,
    project_id: workflow.project_id,
    status: workflow.status,
  })
  dialogVisible.value = true
}

const submitForm = async () => {
  if (!formRef.value) return
  await formRef.value.validate(async (valid) => {
    if (!valid) return
    submitLoading.value = true
    try {
      if (isEditMode.value && editingId.value) {
        await updateWorkflow(editingId.value, {
          name: form.name,
          description: form.description,
          status: form.status,
        })
        ElMessage.success('更新成功')
      } else {
        await createWorkflow({
          name: form.name,
          description: form.description,
          project_id: form.project_id,
        })
        ElMessage.success('创建成功')
      }
      dialogVisible.value = false
      loadWorkflows()
    } catch (error) {
      console.error('Failed to save workflow:', error)
      ElMessage.error(isEditMode.value ? '更新失败' : '创建失败')
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
    await deleteWorkflow(workflow.id)
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
  await Promise.all([
    loadNodes(workflow.id),
    loadEdges(workflow.id),
    loadAllUsers(),
  ])
}

const loadNodes = async (workflowId: number) => {
  nodesLoading.value = true
  try {
    const { data } = await getWorkflowNodes(workflowId)
    workflowNodes.value = (data as any).data || []
  } catch (error) {
    console.error('Failed to load nodes:', error)
    ElMessage.error('加载节点列表失败')
  } finally {
    nodesLoading.value = false
  }
}

const loadEdges = async (workflowId: number) => {
  edgesLoading.value = true
  try {
    const { data } = await getWorkflowEdges(workflowId)
    workflowEdges.value = (data as any).data || []
  } catch (error) {
    console.error('Failed to load edges:', error)
    ElMessage.error('加载边列表失败')
  } finally {
    edgesLoading.value = false
  }
}

const loadAllUsers = async () => {
  if (allUsers.value.length > 0) return
  try {
    const { data } = await getAllUsers()
    allUsers.value = data.data || []
  } catch (error) {
    console.error('Failed to load users:', error)
  }
}

const handleNodeTypeChange = () => {
  // 切换节点类型时重置配置
  Object.assign(nodeForm.config, defaultNodeConfig())
}

const handleAddNode = () => {
  isEditNodeMode.value = false
  editingNodeId.value = null
  nodeForm.name = ''
  nodeForm.node_type = ''
  Object.assign(nodeForm.config, defaultNodeConfig())
  nodeDialogVisible.value = true
}

const handleEditNode = (node: WorkflowNode) => {
  isEditNodeMode.value = true
  editingNodeId.value = node.id
  nodeForm.name = node.name
  nodeForm.node_type = node.node_type
  Object.assign(nodeForm.config, defaultNodeConfig(), node.config || {})
  nodeDialogVisible.value = true
}

const submitNodeForm = async () => {
  if (!nodeFormRef.value || !currentWorkflow.value) return
  await nodeFormRef.value.validate(async (valid) => {
    if (!valid) return
    nodeSubmitLoading.value = true
    try {
      if (isEditNodeMode.value && editingNodeId.value) {
        await updateNode(currentWorkflow.value!.id, editingNodeId.value, {
          name: nodeForm.name,
          config: nodeForm.config,
        })
        ElMessage.success('节点更新成功')
      } else {
        await createNode(currentWorkflow.value!.id, {
          name: nodeForm.name,
          node_type: nodeForm.node_type as any,
          config: nodeForm.config,
        })
        ElMessage.success('节点创建成功')
      }
      nodeDialogVisible.value = false
      await loadNodes(currentWorkflow.value!.id)
    } catch (error) {
      console.error('Failed to save node:', error)
      ElMessage.error(isEditNodeMode.value ? '更新节点失败' : '创建节点失败')
    } finally {
      nodeSubmitLoading.value = false
    }
  })
}

const handleDeleteNode = async (node: WorkflowNode) => {
  if (!currentWorkflow.value) return
  try {
    await ElMessageBox.confirm(`确定要删除节点 "${node.name}" 吗？关联的边也会被删除。`, '删除确认', {
      type: 'warning',
    })
    await deleteNode(currentWorkflow.value.id, node.id)
    ElMessage.success('删除成功')
    await Promise.all([
      loadNodes(currentWorkflow.value.id),
      loadEdges(currentWorkflow.value.id),
    ])
  } catch (error) {
    if (error !== 'cancel') {
      console.error('Failed to delete node:', error)
      ElMessage.error('删除节点失败')
    }
  }
}

// 边管理
const handleAddEdge = () => {
  edgeForm.source_node_id = undefined
  edgeForm.target_node_id = undefined
  edgeForm.condition_expr = ''
  edgeDialogVisible.value = true
}

const submitEdgeForm = async () => {
  if (!edgeFormRef.value || !currentWorkflow.value) return
  await edgeFormRef.value.validate(async (valid) => {
    if (!valid) return
    edgeSubmitLoading.value = true
    try {
      await createEdge(currentWorkflow.value!.id, {
        source_node_id: edgeForm.source_node_id!,
        target_node_id: edgeForm.target_node_id!,
        condition_expr: edgeForm.condition_expr || undefined,
      })
      ElMessage.success('边创建成功')
      edgeDialogVisible.value = false
      await loadEdges(currentWorkflow.value!.id)
    } catch (error) {
      console.error('Failed to create edge:', error)
      ElMessage.error('创建边失败')
    } finally {
      edgeSubmitLoading.value = false
    }
  })
}

const handleDeleteEdge = async (edge: WorkflowEdge) => {
  if (!currentWorkflow.value) return
  try {
    await ElMessageBox.confirm('确定要删除这条边吗？', '删除确认', {
      type: 'warning',
    })
    await deleteEdge(currentWorkflow.value.id, edge.id)
    ElMessage.success('删除成功')
    await loadEdges(currentWorkflow.value.id)
  } catch (error) {
    if (error !== 'cancel') {
      console.error('Failed to delete edge:', error)
      ElMessage.error('删除边失败')
    }
  }
}

const getNodeName = (nodeId: number) => {
  const node = workflowNodes.value.find(n => n.id === nodeId)
  return node ? node.name : `节点#${nodeId}`
}

const getNodeTypeText = (type: string) => {
  const map: Record<string, string> = {
    start: '开始',
    end: '结束',
    approval: '审批',
    work: '工作',
    system: '系统',
  }
  return map[type] || type
}

const getApprovalTypeText = (type: string) => {
  const map: Record<string, string> = {
    single: '单人审批',
    countersign: '会签',
    or_sign: '或签',
  }
  return map[type] || type
}

const getAssigneeTypeText = (type: string) => {
  const map: Record<string, string> = {
    user: '指定用户',
    role: '指定角色',
    reporter: '报告人',
    project_lead: '项目负责人',
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
  background: #3b82f6;
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
    background: #10b981;
  }
  &.end {
    background: #6b7280;
  }
  &.approval {
    background: #f59e0b;
  }
  &.work {
    background: #3b82f6;
  }
  &.system {
    background: #8b5cf6;
  }
}

.section-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 12px;

  h4 {
    font-size: 15px;
    font-weight: 600;
    color: #1f2937;
    margin: 0;
  }
}

// 边列表
.edges-list {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.edge-item {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 12px 16px;
  background: #f8fafc;
  border-radius: 8px;

  &:hover {
    background: #f1f5f9;
  }
}

.edge-info {
  display: flex;
  align-items: center;
  gap: 10px;
  flex: 1;

  .edge-node {
    font-size: 14px;
    font-weight: 500;
    color: #1f2937;
    padding: 4px 10px;
    background: #e5e7eb;
    border-radius: 6px;
  }

  .edge-arrow {
    color: #9ca3af;
    font-size: 16px;
  }

  .edge-condition {
    margin-left: 8px;
  }
}

.edge-actions {
  display: flex;
  gap: 8px;
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
