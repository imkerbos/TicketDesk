<template>
  <div class="config-panel" v-if="node">
    <div class="panel-header">
      <h3 class="panel-title">节点配置</h3>
      <el-button text size="small" @click="$emit('close')">
        <el-icon><Close /></el-icon>
      </el-button>
    </div>

    <el-scrollbar class="panel-body">
      <el-form label-position="top" size="default">
        <!-- 基本信息 -->
        <div class="config-section">
          <div class="section-title">基本信息</div>
          <el-form-item label="节点名称">
            <el-input
              v-model="localName"
              placeholder="请输入节点名称"
              @change="emitUpdate"
            />
          </el-form-item>
          <el-form-item label="节点类型">
            <el-tag :type="nodeTypeTagType" size="large">{{ nodeTypeText }}</el-tag>
          </el-form-item>
        </div>

        <!-- 审批节点配置 -->
        <template v-if="node.data.nodeType === 'approval'">
          <div class="config-section">
            <div class="section-title">审批配置</div>
            <el-form-item label="审批类型">
              <el-select v-model="localConfig.approval_type" placeholder="请选择审批类型" style="width: 100%" @change="emitUpdate">
                <el-option label="单人审批" value="single" />
                <el-option label="会签（所有人通过）" value="countersign" />
                <el-option label="或签（任一人通过）" value="or_sign" />
              </el-select>
            </el-form-item>
            <el-form-item label="审批人">
              <el-select
                v-model="localConfig.approvers"
                multiple
                filterable
                placeholder="请选择审批人"
                style="width: 100%"
                @change="emitUpdate"
              >
                <el-option v-for="u in users" :key="u.id" :label="u.display_name" :value="u.id" />
              </el-select>
            </el-form-item>
            <el-form-item label="审批角色（可选）">
              <el-input v-model="localConfig.approver_role" placeholder="如: project_lead" @change="emitUpdate" />
            </el-form-item>
          </div>
        </template>

        <!-- 工作节点配置 -->
        <template v-if="node.data.nodeType === 'work'">
          <div class="config-section">
            <div class="section-title">指派配置</div>
            <el-form-item label="指派类型">
              <el-select v-model="localConfig.assignee_type" placeholder="请选择指派类型" style="width: 100%" @change="emitUpdate">
                <el-option label="指定用户" value="user" />
                <el-option label="指定角色" value="role" />
                <el-option label="报告人" value="reporter" />
                <el-option label="项目负责人" value="project_lead" />
              </el-select>
            </el-form-item>
            <el-form-item v-if="localConfig.assignee_type === 'user'" label="指派人">
              <el-select
                v-model="localConfig.assignees"
                multiple
                filterable
                placeholder="请选择指派人"
                style="width: 100%"
                @change="emitUpdate"
              >
                <el-option v-for="u in users" :key="u.id" :label="u.display_name" :value="u.id" />
              </el-select>
            </el-form-item>
            <el-form-item v-if="localConfig.assignee_type === 'role'" label="角色名称">
              <el-input v-model="localConfig.assignee_role" placeholder="请输入角色名称" @change="emitUpdate" />
            </el-form-item>
          </div>
        </template>

        <!-- 系统节点配置 -->
        <template v-if="node.data.nodeType === 'system'">
          <div class="config-section">
            <div class="section-title">系统动作</div>
            <el-form-item label="动作名称">
              <el-input v-model="localConfig.action" placeholder="请输入系统动作" @change="emitUpdate" />
            </el-form-item>
          </div>
        </template>

        <!-- 通用配置（开始/结束节点不显示） -->
        <template v-if="node.data.nodeType !== 'start' && node.data.nodeType !== 'end'">
          <div class="config-section">
            <div class="section-title">通用配置</div>
            <el-form-item label="目标状态">
              <el-select
                v-model="localConfig.target_status"
                placeholder="进入该节点时工单状态"
                style="width: 100%"
                clearable
                @change="emitUpdate"
              >
                <el-option label="未开始 (open)" value="open" />
                <el-option label="进行中 (in_progress)" value="in_progress" />
                <el-option label="待确认 (pending_review)" value="pending_review" />
                <el-option label="已解决 (resolved)" value="resolved" />
                <el-option label="已关闭 (closed)" value="closed" />
              </el-select>
            </el-form-item>
            <el-form-item label="超时时间（小时）">
              <el-input-number
                v-model="localConfig.timeout_hours"
                :min="0"
                :max="720"
                placeholder="0 表示不超时"
                style="width: 100%"
                @change="emitUpdate"
              />
            </el-form-item>
            <el-form-item label="节点说明">
              <el-input
                v-model="localConfig.description"
                type="textarea"
                :rows="3"
                placeholder="节点说明"
                @change="emitUpdate"
              />
            </el-form-item>
          </div>
        </template>

        <!-- 删除按钮（开始/结束节点不可删除） -->
        <div class="config-section" v-if="node.data.nodeType !== 'start' && node.data.nodeType !== 'end'">
          <el-button type="danger" plain style="width: 100%" @click="$emit('delete', node.id)">
            <el-icon><Delete /></el-icon>
            删除节点
          </el-button>
        </div>
      </el-form>
    </el-scrollbar>
  </div>
</template>

<script setup lang="ts">
import { ref, watch, computed, reactive } from 'vue'
import { Close, Delete } from '@element-plus/icons-vue'
import type { NodeConfig } from '@/types/workflow'

interface FlowNode {
  id: string
  data: {
    label: string
    nodeType: string
    config?: NodeConfig
    backendId?: number
  }
}

interface UserOption {
  id: number
  display_name: string
}

const props = defineProps<{
  node: FlowNode | null
  users: UserOption[]
}>()

const emit = defineEmits<{
  (e: 'close'): void
  (e: 'update', payload: { id: string; name: string; config: NodeConfig }): void
  (e: 'delete', id: string): void
}>()

const localName = ref('')
const localConfig = reactive<NodeConfig>({
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
  target_status: undefined,
})

// 当选中节点变化时，同步数据到本地
watch(
  () => props.node,
  (newNode) => {
    if (newNode) {
      localName.value = newNode.data.label || ''
      const cfg = newNode.data.config || {}
      Object.assign(localConfig, {
        approval_type: cfg.approval_type || undefined,
        approvers: cfg.approvers || [],
        approver_role: cfg.approver_role || '',
        assignee_type: cfg.assignee_type || undefined,
        assignees: cfg.assignees || [],
        assignee_role: cfg.assignee_role || '',
        action: cfg.action || '',
        parameters: cfg.parameters || {},
        timeout_hours: cfg.timeout_hours || 0,
        description: cfg.description || '',
        target_status: cfg.target_status || undefined,
      })
    }
  },
  { immediate: true, deep: true }
)

const emitUpdate = () => {
  if (!props.node) return
  emit('update', {
    id: props.node.id,
    name: localName.value,
    config: { ...localConfig },
  })
}

const nodeTypeText = computed(() => {
  const map: Record<string, string> = {
    start: '开始节点',
    end: '结束节点',
    approval: '审批节点',
    work: '工作节点',
    system: '系统节点',
  }
  return props.node ? map[props.node.data.nodeType] || props.node.data.nodeType : ''
})

const nodeTypeTagType = computed((): 'primary' | 'success' | 'info' | 'warning' | 'danger' => {
  const map: Record<string, 'primary' | 'success' | 'info' | 'warning' | 'danger'> = {
    start: 'success',
    end: 'info',
    approval: 'warning',
    work: 'primary',
    system: 'danger',
  }
  return props.node ? (map[props.node.data.nodeType] || 'info') : 'info'
})
</script>

<style scoped lang="scss">
.config-panel {
  width: 300px;
  background: #fff;
  border-left: 1px solid #e5e7eb;
  display: flex;
  flex-direction: column;
  flex-shrink: 0;
  overflow: hidden;
}

.panel-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 16px;
  border-bottom: 1px solid #e5e7eb;

  .panel-title {
    font-size: 15px;
    font-weight: 600;
    color: #1f2937;
    margin: 0;
  }
}

.panel-body {
  flex: 1;
  overflow-y: auto;
  padding: 16px;
}

.config-section {
  margin-bottom: 20px;

  .section-title {
    font-size: 13px;
    font-weight: 600;
    color: #374151;
    margin-bottom: 12px;
    padding-bottom: 8px;
    border-bottom: 1px solid #f3f4f6;
  }
}

:deep(.el-form-item) {
  margin-bottom: 14px;
}

:deep(.el-form-item__label) {
  font-size: 12px;
  color: #6b7280;
  padding-bottom: 4px !important;
}
</style>
