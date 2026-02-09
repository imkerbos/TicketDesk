<template>
  <el-input
    v-model="internalValue"
    :placeholder="field.description || '例如: 2h 30m, 1d 4h'"
    :disabled="disabled"
    :readonly="readonly"
    clearable
    @change="handleChange"
  >
    <template #suffix>
      <el-tooltip
        content="支持格式: 1w 2d 3h 4m (周/天/小时/分钟)"
        placement="top"
      >
        <el-icon><QuestionFilled /></el-icon>
      </el-tooltip>
    </template>
  </el-input>
</template>

<script setup lang="ts">
import { ref, watch } from 'vue'
import { QuestionFilled } from '@element-plus/icons-vue'
import type { FieldDefinition, FieldSchemeItem } from '@/types/field'

const props = defineProps<{
  field: FieldDefinition
  scheme?: FieldSchemeItem
  projectKey: string
  modelValue?: string
  readonly?: boolean
  disabled?: boolean
}>()

const emit = defineEmits<{
  (e: 'update:modelValue', value: string): void
  (e: 'change', value: string): void
}>()

const internalValue = ref(props.modelValue || '')

watch(() => props.modelValue, (newVal) => {
  internalValue.value = newVal || ''
})

watch(internalValue, (newVal) => {
  emit('update:modelValue', newVal)
})

const handleChange = (value: string) => {
  emit('change', value)
}
</script>
