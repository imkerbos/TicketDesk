<template>
  <el-date-picker
    v-model="internalValue"
    type="datetime"
    :placeholder="field.description || `${field.field_name}`"
    :disabled="disabled"
    :readonly="readonly"
    value-format="YYYY-MM-DDTHH:mm"
    format="YYYY-MM-DD HH:mm"
    :default-time="defaultTime"
    style="width: 100%"
    @change="handleChange"
  />
</template>

<script setup lang="ts">
import { ref, watch } from 'vue'
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
// 默认时间 09:00
const defaultTime = new Date(2000, 0, 1, 9, 0, 0)

watch(
  () => props.modelValue,
  (newVal) => {
    internalValue.value = newVal || ''
  },
)

watch(internalValue, (newVal) => {
  emit('update:modelValue', newVal)
})

const handleChange = (value: string) => {
  emit('change', value)
}
</script>
