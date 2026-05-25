<template>
  <el-switch
    v-model="internalValue"
    :disabled="disabled || readonly"
    active-text="是"
    inactive-text="否"
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
  modelValue?: boolean
  readonly?: boolean
  disabled?: boolean
}>()

const emit = defineEmits<{
  (e: 'update:modelValue', value: boolean): void
  (e: 'change', value: boolean): void
}>()

const internalValue = ref<boolean>(!!props.modelValue)

watch(
  () => props.modelValue,
  (newVal) => {
    internalValue.value = !!newVal
  },
)

watch(internalValue, (newVal) => {
  emit('update:modelValue', newVal)
})

const handleChange = (value: boolean) => {
  emit('change', value)
}
</script>
