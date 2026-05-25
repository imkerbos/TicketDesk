<template>
  <div
    class="pending-attachment-list"
    :class="{ 'is-drag-active': dragActive }"
    @dragover.prevent="dragActive = true"
    @dragleave.prevent="dragActive = false"
    @drop.prevent="handleDrop"
  >
    <div class="add-area" @click="triggerFileInput">
      <el-icon class="add-icon"><Plus /></el-icon>
      <span class="add-text">点击 / 拖拽 / 粘贴 添加附件</span>
      <span class="add-hint">单文件 ≤ 10MB</span>
    </div>
    <input
      ref="fileInputRef"
      type="file"
      multiple
      style="display: none"
      @change="handleFileInput"
    />

    <div v-if="modelValue.length > 0" class="file-list">
      <div v-for="(file, i) in modelValue" :key="`${file.name}-${file.size}-${i}`" class="file-item">
        <img v-if="isImageFile(file)" :src="previewUrl(file)" class="thumb" alt="" />
        <el-icon v-else class="file-icon"><Document /></el-icon>
        <div class="meta">
          <div class="name" :title="file.name">{{ file.name }}</div>
          <div class="size">{{ formatSize(file.size) }}</div>
        </div>
        <el-icon class="remove" @click.stop="removeFile(i)"><Close /></el-icon>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onBeforeUnmount } from 'vue'
import { Plus, Close, Document } from '@element-plus/icons-vue'
import { validateFile, isImage, formatSize } from '@/utils/attachment'

const props = defineProps<{
  modelValue: File[]
}>()

const emit = defineEmits<{
  (e: 'update:modelValue', files: File[]): void
}>()

const fileInputRef = ref<HTMLInputElement | null>(null)
const dragActive = ref(false)
// 用 File 对象本身作 key, 避免删除后索引偏移导致预览串号
const previewUrls = new Map<File, string>()

const isImageFile = (f: File) => isImage(f)

const previewUrl = (file: File): string => {
  let url = previewUrls.get(file)
  if (!url) {
    url = URL.createObjectURL(file)
    previewUrls.set(file, url)
  }
  return url
}

const triggerFileInput = () => {
  fileInputRef.value?.click()
}

const handleFileInput = (e: Event) => {
  const input = e.target as HTMLInputElement
  if (!input.files) return
  addFiles(Array.from(input.files))
  input.value = '' // 允许重复选择同名文件
}

const handleDrop = (e: DragEvent) => {
  dragActive.value = false
  if (!e.dataTransfer?.files) return
  addFiles(Array.from(e.dataTransfer.files))
}

const addFiles = (files: File[]) => {
  const accepted: File[] = []
  for (const f of files) {
    // 图片粘贴常无文件名, 给默认名
    const file = f.name
      ? f
      : new File([f], `screenshot-${Date.now()}.png`, { type: f.type || 'image/png' })
    if (!validateFile(file)) continue
    accepted.push(file)
  }
  if (accepted.length > 0) {
    emit('update:modelValue', [...props.modelValue, ...accepted])
  }
}

const removeFile = (idx: number) => {
  const file = props.modelValue[idx]
  const url = previewUrls.get(file)
  if (url) {
    URL.revokeObjectURL(url)
    previewUrls.delete(file)
  }
  const next = props.modelValue.filter((_, i) => i !== idx)
  emit('update:modelValue', next)
}

// 暴露 addFiles 供父组件粘贴监听调用
defineExpose({ addFiles })

onBeforeUnmount(() => {
  for (const url of previewUrls.values()) {
    URL.revokeObjectURL(url)
  }
  previewUrls.clear()
})
</script>

<style scoped lang="scss">
.pending-attachment-list {
  border: 1px dashed var(--td-border-color, #e5e7eb);
  border-radius: 6px;
  padding: 12px;
  transition: border-color 150ms ease-out, background-color 150ms ease-out;

  &.is-drag-active {
    border-color: var(--td-color-primary, #3b82f6);
    background-color: rgba(59, 130, 246, 0.04);
  }
}

.add-area {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 8px 12px;
  border-radius: 4px;
  cursor: pointer;
  color: var(--td-text-secondary, #6b7280);
  transition: background-color 150ms ease-out;

  &:hover {
    background-color: var(--el-fill-color-light, #f5f7fa);
  }

  .add-icon {
    font-size: 16px;
  }

  .add-hint {
    margin-left: auto;
    font-size: 12px;
    color: var(--td-text-disabled, #9ca3af);
  }
}

.file-list {
  display: flex;
  flex-direction: column;
  gap: 6px;
  margin-top: 8px;
}

.file-item {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 6px 10px;
  background: var(--el-fill-color-light, #f5f7fa);
  border-radius: 4px;
  transition: background-color 150ms ease-out;

  &:hover {
    background: var(--el-fill-color, #f0f2f5);
  }

  .thumb {
    width: 32px;
    height: 32px;
    object-fit: cover;
    border-radius: 3px;
    flex-shrink: 0;
  }

  .file-icon {
    font-size: 24px;
    color: var(--td-text-secondary, #6b7280);
    flex-shrink: 0;
  }

  .meta {
    flex: 1;
    min-width: 0;

    .name {
      font-size: 13px;
      color: var(--td-text-primary, #1f2937);
      white-space: nowrap;
      overflow: hidden;
      text-overflow: ellipsis;
    }

    .size {
      font-size: 12px;
      color: var(--td-text-disabled, #9ca3af);
    }
  }

  .remove {
    font-size: 16px;
    color: var(--td-text-disabled, #9ca3af);
    cursor: pointer;
    flex-shrink: 0;
    transition: color 150ms ease-out;

    &:hover {
      color: var(--td-color-danger, #ef4444);
    }
  }
}

@media (prefers-reduced-motion: reduce) {
  .pending-attachment-list,
  .add-area,
  .file-item,
  .file-item .remove {
    transition: none;
  }
}
</style>
