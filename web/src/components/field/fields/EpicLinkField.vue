<template>
  <el-select
    v-model="internalValue"
    :placeholder="field.description || `${field.field_name}`"
    :disabled="disabled"
    :readonly="readonly"
    clearable
    filterable
    remote
    :remote-method="searchEpics"
    :loading="loading"
    style="width: 100%"
    @change="handleChange"
  >
    <el-option
      v-for="epic in epics"
      :key="epic.id"
      :label="`${epic.issue_key} - ${epic.title}`"
      :value="epic.id"
    >
      <div class="epic-option">
        <span class="epic-key">{{ epic.issue_key }}</span>
        <span class="epic-title">{{ epic.title }}</span>
      </div>
    </el-option>
  </el-select>
</template>

<script setup lang="ts">
import { ref, watch, onMounted } from 'vue'
import type { FieldDefinition, FieldSchemeItem } from '@/types/field'
import { getIssueList } from '@/api/issue'

interface Epic {
  id: number
  issue_key: string
  title: string
}

const props = defineProps<{
  field: FieldDefinition
  scheme?: FieldSchemeItem
  projectKey: string
  modelValue?: number
  readonly?: boolean
  disabled?: boolean
}>()

const emit = defineEmits<{
  (e: 'update:modelValue', value: number | undefined): void
  (e: 'change', value: number | undefined): void
}>()

const internalValue = ref(props.modelValue)
const epics = ref<Epic[]>([])
const loading = ref(false)

const loadEpics = async (keyword = '') => {
  if (!props.projectKey) return
  loading.value = true
  try {
    // 查询 Epic 类型的工单
    const { data } = await getIssueList({
      project_key: props.projectKey,
      keyword,
      page: 1,
      page_size: 50,
    })
    // 过滤出 Epic 类型（名称为 Epic）
    epics.value = (data.data?.items || [])
      .filter((item: any) => item.issue_type?.name === 'Epic')
      .map((item: any) => ({
        id: item.id,
        issue_key: item.issue_key,
        title: item.title,
      }))
  } catch (error) {
    console.error('Failed to load epics:', error)
  } finally {
    loading.value = false
  }
}

const searchEpics = (query: string) => {
  loadEpics(query)
}

onMounted(() => {
  loadEpics()
})

watch(() => props.projectKey, () => {
  loadEpics()
})

watch(() => props.modelValue, (newVal) => {
  internalValue.value = newVal
})

watch(internalValue, (newVal) => {
  emit('update:modelValue', newVal)
})

const handleChange = (value: number | undefined) => {
  emit('change', value)
}
</script>

<style scoped>
.epic-option {
  display: flex;
  align-items: center;
  gap: 8px;
}

.epic-key {
  color: var(--el-color-primary);
  font-weight: 500;
}

.epic-title {
  flex: 1;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}
</style>
