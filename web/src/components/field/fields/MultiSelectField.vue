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
      v-for="option in options"
      :key="option.value"
      :label="option.label"
      :value="option.value"
    />
  </el-select>
</template>

<script setup lang="ts">
import { ref, watch, computed } from 'vue'
import type { FieldDefinition, FieldSchemeItem, FieldOption } from '@/types/field'
import { parseFieldOptions } from '@/types/field'

const props = defineProps<{
  field: FieldDefinition
  scheme?: FieldSchemeItem
  projectKey: string
  modelValue?: string[]
  readonly?: boolean
  disabled?: boolean
}>()

const emit = defineEmits<{
  (e: 'update:modelValue', value: string[]): void
  (e: 'change', value: string[]): void
}>()

// 确保值始终为 string[]，处理后端可能返回的各种格式
function ensureStringArray(val: unknown): string[] {
  if (Array.isArray(val)) return val.map(String)
  if (typeof val === 'string' && val) {
    try {
      const parsed = JSON.parse(val)
      return Array.isArray(parsed) ? parsed.map(String) : []
    } catch {
      return []
    }
  }
  return []
}

const internalValue = ref<string[]>(ensureStringArray(props.modelValue))

const options = computed<FieldOption[]>(() => {
  return parseFieldOptions(props.field.options)
})

watch(() => props.modelValue, (newVal) => {
  const coerced = ensureStringArray(newVal)
  // 内容相同时不更新，避免不必要的 el-select 重渲染
  if (JSON.stringify(coerced) !== JSON.stringify(internalValue.value)) {
    internalValue.value = coerced
  }
})

watch(internalValue, (newVal) => {
  emit('update:modelValue', newVal)
})

const handleChange = (value: string[]) => {
  emit('change', value)
}
</script>
