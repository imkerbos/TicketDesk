<template>
  <el-dialog v-model="visible" title="边配置" width="450px" destroy-on-close @close="$emit('close')">
    <el-form label-position="top">
      <el-form-item label="源节点">
        <el-input :model-value="sourceNodeName" disabled />
      </el-form-item>
      <el-form-item label="目标节点">
        <el-input :model-value="targetNodeName" disabled />
      </el-form-item>
      <el-form-item label="条件表达式（可选）">
        <el-input
          v-model="localCondition"
          placeholder="如: status == 'approved'"
          clearable
        />
        <div class="form-hint">留空表示无条件流转，直接通过</div>
      </el-form-item>
    </el-form>
    <template #footer>
      <el-button @click="visible = false">取消</el-button>
      <el-button type="danger" plain @click="handleDelete">删除此边</el-button>
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
  data?: {
    conditionExpr?: string
    backendId?: number
  }
}

interface FlowNode {
  id: string
  data: {
    label: string
  }
}

const props = defineProps<{
  edge: FlowEdge | null
  nodes: FlowNode[]
}>()

const emit = defineEmits<{
  (e: 'close'): void
  (e: 'update', payload: { id: string; conditionExpr: string }): void
  (e: 'delete', id: string): void
}>()

const visible = ref(false)
const localCondition = ref('')

watch(
  () => props.edge,
  (newEdge) => {
    if (newEdge) {
      visible.value = true
      localCondition.value = newEdge.data?.conditionExpr || ''
    } else {
      visible.value = false
    }
  },
  { immediate: true }
)

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
  emit('update', {
    id: props.edge.id,
    conditionExpr: localCondition.value,
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
