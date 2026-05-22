<template>
  <el-select
    v-model="internalValue"
    :placeholder="field.description || `${field.field_name}`"
    :disabled="disabled"
    :readonly="readonly"
    clearable
    filterable
    multiple
    style="width: 100%"
    @change="handleChange"
  >
    <el-option
      v-for="component in components"
      :key="component.id"
      :label="component.name"
      :value="component.id"
    >
      <div class="component-option">
        <span class="component-name">{{ component.name }}</span>
        <span v-if="component.lead_user" class="component-lead">
          {{ component.lead_user.display_name || component.lead_user.username }}
        </span>
      </div>
    </el-option>
  </el-select>
</template>

<script setup lang="ts">
import { ref, watch, onMounted } from 'vue'
import type { FieldDefinition, FieldSchemeItem, ProjectComponent } from '@/types/field'
import { getComponents } from '@/api/field'

const props = defineProps<{
  field: FieldDefinition
  scheme?: FieldSchemeItem
  projectKey: string
  modelValue?: number[]
  readonly?: boolean
  disabled?: boolean
}>()

const emit = defineEmits<{
  (e: 'update:modelValue', value: number[]): void
  (e: 'change', value: number[]): void
}>()

// 确保值为数组类型
function ensureArray(val: unknown): number[] {
  if (Array.isArray(val)) return val
  return []
}

const internalValue = ref<number[]>(ensureArray(props.modelValue))
const components = ref<ProjectComponent[]>([])

const loadComponents = async () => {
  if (!props.projectKey) return
  try {
    const res = await getComponents(props.projectKey)
    components.value = res.data.data || []
  } catch {
    // ignored
  }
}

onMounted(() => {
  loadComponents()
})

watch(() => props.projectKey, () => {
  loadComponents()
})

watch(() => props.modelValue, (newVal) => {
  const coerced = ensureArray(newVal)
  if (JSON.stringify(coerced) !== JSON.stringify(internalValue.value)) {
    internalValue.value = coerced
  }
})

watch(internalValue, (newVal) => {
  emit('update:modelValue', newVal)
})

const handleChange = (value: number[]) => {
  emit('change', value)
}
</script>

<style scoped>
.component-option {
  display: flex;
  align-items: center;
  justify-content: space-between;
  width: 100%;
}

.component-name {
  flex: 1;
}

.component-lead {
  color: var(--el-text-color-secondary);
  font-size: 12px;
}
</style>
