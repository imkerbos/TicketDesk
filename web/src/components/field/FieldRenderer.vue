<template>
  <div class="field-renderer">
    <component
      :is="fieldComponent"
      v-model="internalValue"
      :field="field"
      :scheme="scheme"
      :project-key="projectKey"
      :readonly="readonly"
      :disabled="disabled"
      @change="handleChange"
    />
  </div>
</template>

<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import type { FieldDefinition, FieldSchemeItem, FieldTypeValue } from '@/types/field'
import { FieldType } from '@/types/field'

// 字段组件
import TextField from './fields/TextField.vue'
import TextareaField from './fields/TextareaField.vue'
import NumberField from './fields/NumberField.vue'
import DateField from './fields/DateField.vue'
import SelectField from './fields/SelectField.vue'
import MultiSelectField from './fields/MultiSelectField.vue'
import UserSelectField from './fields/UserSelectField.vue'
import VersionSelectField from './fields/VersionSelectField.vue'
import ComponentSelectField from './fields/ComponentSelectField.vue'
import LabelSelectField from './fields/LabelSelectField.vue'
import EpicLinkField from './fields/EpicLinkField.vue'
import TimeEstimateField from './fields/TimeEstimateField.vue'

// 需要数组值的字段类型
const ARRAY_FIELD_TYPES: string[] = [FieldType.MULTISELECT, FieldType.LABEL, FieldType.COMPONENT]

const props = defineProps<{
  field: FieldDefinition
  scheme?: FieldSchemeItem
  projectKey: string
  modelValue?: any
  readonly?: boolean
  disabled?: boolean
}>()

const emit = defineEmits<{
  (e: 'update:modelValue', value: any): void
  (e: 'change', value: any): void
}>()

// 根据字段类型确保值的类型正确
function coerceValue(fieldType: string, value: any): any {
  if (ARRAY_FIELD_TYPES.includes(fieldType)) {
    if (Array.isArray(value)) return value
    if (typeof value === 'string' && value) {
      try {
        const parsed = JSON.parse(value)
        return Array.isArray(parsed) ? parsed : []
      } catch {
        return []
      }
    }
    return []
  }
  return value
}

const internalValue = ref(coerceValue(props.field.field_type, props.modelValue))

watch(() => props.modelValue, (newVal) => {
  const coerced = coerceValue(props.field.field_type, newVal)
  // 对数组类型做内容比较，避免引用不同但内容相同时触发子组件不必要的重渲染
  if (ARRAY_FIELD_TYPES.includes(props.field.field_type)) {
    if (JSON.stringify(coerced) === JSON.stringify(internalValue.value)) return
  }
  internalValue.value = coerced
})

watch(internalValue, (newVal) => {
  emit('update:modelValue', newVal)
})

const handleChange = (value: any) => {
  emit('change', value)
}

// 根据字段类型返回对应的组件
const fieldComponent = computed(() => {
  const componentMap: Record<FieldTypeValue, any> = {
    [FieldType.TEXT]: TextField,
    [FieldType.TEXTAREA]: TextareaField,
    [FieldType.NUMBER]: NumberField,
    [FieldType.DATE]: DateField,
    [FieldType.SELECT]: SelectField,
    [FieldType.MULTISELECT]: MultiSelectField,
    [FieldType.USER]: UserSelectField,
    [FieldType.VERSION]: VersionSelectField,
    [FieldType.COMPONENT]: ComponentSelectField,
    [FieldType.LABEL]: LabelSelectField,
    [FieldType.EPIC_LINK]: EpicLinkField,
    [FieldType.TIME_ESTIMATE]: TimeEstimateField,
  }
  return componentMap[props.field.field_type as FieldTypeValue] || TextField
})
</script>

<style scoped>
.field-renderer {
  width: 100%;
}
</style>
