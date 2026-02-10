<template>
  <el-dialog v-model="visible" title="连接配置" width="450px" destroy-on-close @close="$emit('close')">
    <el-form label-position="top">
      <el-form-item label="源节点">
        <el-input :model-value="sourceNodeName" disabled />
      </el-form-item>
      <el-form-item label="目标节点">
        <el-input :model-value="targetNodeName" disabled />
      </el-form-item>

      <!-- 流转条件：统一使用预设 + 自定义 -->
      <el-form-item label="流转条件">
        <!-- 条件模式选择 -->
        <el-radio-group v-model="conditionMode" style="margin-bottom: 8px; width: 100%">
          <el-radio-button value="none">无条件</el-radio-button>
          <el-radio-button value="preset">预设条件</el-radio-button>
          <el-radio-button value="custom">自定义</el-radio-button>
        </el-radio-group>

        <!-- 预设条件选择 -->
        <el-select
          v-if="conditionMode === 'preset'"
          v-model="localCondition"
          placeholder="请选择流转条件"
          style="width: 100%"
        >
          <el-option-group label="常用条件">
            <el-option label="通过 / 完成" value="approved" />
            <el-option label="拒绝 / 退回" value="rejected" />
          </el-option-group>
        </el-select>

        <!-- 自定义条件输入 -->
        <el-input
          v-if="conditionMode === 'custom'"
          v-model="localCondition"
          placeholder="输入自定义条件名称，如：验收通过、需要修改、转交等"
          clearable
        />

        <!-- 提示信息 -->
        <div class="form-hint">
          <template v-if="conditionMode === 'none'">
            无条件流转：当源节点完成后，直接流转到「{{ targetNodeName }}」
          </template>
          <template v-else-if="conditionMode === 'preset' || conditionMode === 'custom'">
            <template v-if="localCondition">
              当操作「{{ conditionDisplayText }}」时，流程将流转到「{{ targetNodeName }}」
            </template>
            <template v-else>
              请设置流转条件
            </template>
          </template>
        </div>
      </el-form-item>

      <!-- 显示标签 -->
      <el-form-item label="连接标签（可选）">
        <el-input
          v-model="localLabel"
          placeholder="显示在连接线上的文字（留空则自动使用条件名称）"
          clearable
        />
        <div class="form-hint">显示在连接线上的文字，如：通过、拒绝、验收完成等</div>
      </el-form-item>
    </el-form>
    <template #footer>
      <el-button @click="visible = false">取消</el-button>
      <el-button type="danger" plain @click="handleDelete">删除此连接</el-button>
      <el-button type="primary" @click="handleSave">保存</el-button>
    </template>
  </el-dialog>
</template>

<script setup lang="ts">
import { ref, watch, computed } from 'vue'

interface FlowEdge {
  id: string
  source: string
  target: string
  label?: string
  data?: {
    conditionExpr?: string
    backendId?: number
  }
}

interface FlowNode {
  id: string
  data: {
    label: string
    nodeType?: string
  }
}

const props = defineProps<{
  edge: FlowEdge | null
  nodes: FlowNode[]
}>()

const emit = defineEmits<{
  (e: 'close'): void
  (e: 'update', payload: { id: string; conditionExpr: string; label: string }): void
  (e: 'delete', id: string): void
}>()

const visible = ref(false)
const localCondition = ref('')
const localLabel = ref('')
const conditionMode = ref<'none' | 'preset' | 'custom'>('none')

// 预设条件的显示名称映射
const presetDisplayNames: Record<string, string> = {
  approved: '通过',
  rejected: '拒绝/退回',
}

// 判断条件是否是预设条件
const isPresetCondition = (condition: string) => {
  return condition === 'approved' || condition === 'rejected'
}

// 条件的显示文本
const conditionDisplayText = computed(() => {
  if (!localCondition.value) return ''
  return presetDisplayNames[localCondition.value] || localCondition.value
})

watch(
  () => props.edge,
  (newEdge) => {
    if (newEdge) {
      visible.value = true
      const condition = newEdge.data?.conditionExpr || ''
      localCondition.value = condition
      localLabel.value = (newEdge.label as string) || ''

      // 根据已有条件确定模式
      if (!condition) {
        conditionMode.value = 'none'
      } else if (isPresetCondition(condition)) {
        conditionMode.value = 'preset'
      } else {
        conditionMode.value = 'custom'
      }
    } else {
      visible.value = false
    }
  },
  { immediate: true }
)

// 当模式切换时，清空条件
watch(conditionMode, (newMode, oldMode) => {
  if (newMode === 'none') {
    localCondition.value = ''
  } else if (newMode === 'preset' && oldMode !== 'preset') {
    // 如果从其他模式切到预设，且当前值不是预设值，清空
    if (!isPresetCondition(localCondition.value)) {
      localCondition.value = ''
    }
  } else if (newMode === 'custom' && oldMode !== 'custom') {
    // 如果从预设切到自定义，且当前值是预设值，清空让用户输入
    if (isPresetCondition(localCondition.value)) {
      localCondition.value = ''
    }
  }
})

const sourceNodeName = computed(() => {
  if (!props.edge) return ''
  const node = props.nodes.find(n => n.id === props.edge!.source)
  return node?.data.label || props.edge.source
})

const targetNodeName = computed(() => {
  if (!props.edge) return ''
  const node = props.nodes.find(n => n.id === props.edge!.target)
  return node?.data.label || props.edge.target
})

const handleSave = () => {
  if (!props.edge) return

  const condition = conditionMode.value === 'none' ? '' : localCondition.value

  // 自动生成标签：如果用户没有填标签，使用条件的显示名称
  let label = localLabel.value
  if (!label && condition) {
    label = presetDisplayNames[condition] || condition
  }

  emit('update', {
    id: props.edge.id,
    conditionExpr: condition,
    label: label,
  })
  visible.value = false
}

const handleDelete = () => {
  if (!props.edge) return
  emit('delete', props.edge.id)
  visible.value = false
}
</script>

<style scoped lang="scss">
.form-hint {
  font-size: 11px;
  color: #9ca3af;
  margin-top: 4px;
}
</style>
