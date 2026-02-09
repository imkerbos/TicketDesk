<template>
  <el-select
    v-model="internalValue"
    :placeholder="field.description || `${field.field_name}`"
    :disabled="disabled"
    :readonly="readonly"
    clearable
    filterable
    :multiple="multiple"
    style="width: 100%"
    @change="handleChange"
  >
    <el-option-group
      v-for="group in versionGroups"
      :key="group.label"
      :label="group.label"
    >
      <el-option
        v-for="version in group.versions"
        :key="version.id"
        :label="version.name"
        :value="version.id"
      >
        <div class="version-option">
          <span class="version-name">{{ version.name }}</span>
          <el-tag v-if="version.status === 'released'" type="success" size="small">已发布</el-tag>
          <el-tag v-else-if="version.status === 'archived'" type="info" size="small">已归档</el-tag>
        </div>
      </el-option>
    </el-option-group>
  </el-select>
</template>

<script setup lang="ts">
import { ref, watch, onMounted, computed } from 'vue'
import type { FieldDefinition, FieldSchemeItem, ProjectVersion } from '@/types/field'
import { getVersions } from '@/api/field'

const props = defineProps<{
  field: FieldDefinition
  scheme?: FieldSchemeItem
  projectKey: string
  modelValue?: number | number[]
  readonly?: boolean
  disabled?: boolean
  multiple?: boolean
}>()

const emit = defineEmits<{
  (e: 'update:modelValue', value: number | number[] | undefined): void
  (e: 'change', value: number | number[] | undefined): void
}>()

const internalValue = ref(props.modelValue)
const versions = ref<ProjectVersion[]>([])

const versionGroups = computed(() => {
  const unreleased = versions.value.filter(v => v.status === 'unreleased')
  const released = versions.value.filter(v => v.status === 'released')
  const archived = versions.value.filter(v => v.status === 'archived')

  const groups = []
  if (unreleased.length > 0) {
    groups.push({ label: '未发布', versions: unreleased })
  }
  if (released.length > 0) {
    groups.push({ label: '已发布', versions: released })
  }
  if (archived.length > 0) {
    groups.push({ label: '已归档', versions: archived })
  }
  return groups
})

const loadVersions = async () => {
  if (!props.projectKey) return
  try {
    const data = await getVersions(props.projectKey)
    versions.value = data
  } catch (error) {
    console.error('Failed to load versions:', error)
  }
}

onMounted(() => {
  loadVersions()
})

watch(() => props.projectKey, () => {
  loadVersions()
})

watch(() => props.modelValue, (newVal) => {
  internalValue.value = newVal
})

watch(internalValue, (newVal) => {
  emit('update:modelValue', newVal)
})

const handleChange = (value: number | number[] | undefined) => {
  emit('change', value)
}
</script>

<style scoped>
.version-option {
  display: flex;
  align-items: center;
  justify-content: space-between;
  width: 100%;
}

.version-name {
  flex: 1;
}
</style>
