<template>
  <div class="attachment-upload">
    <el-upload
      ref="uploadRef"
      :action="uploadUrl"
      :headers="uploadHeaders"
      :on-success="handleSuccess"
      :on-error="handleError"
      :before-upload="beforeUpload"
      :show-file-list="false"
      :drag="true"
      multiple
    >
      <el-icon class="el-icon--upload"><upload-filled /></el-icon>
      <div class="el-upload__text">
        将文件拖到此处，或<em>点击上传</em>
      </div>
      <template #tip>
        <div class="el-upload__tip">
          支持图片、文档、压缩包等文件，单个文件不超过10MB
        </div>
      </template>
    </el-upload>
  </div>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue'
import { ElMessage } from 'element-plus'
import { UploadFilled } from '@element-plus/icons-vue'
import type { UploadInstance } from 'element-plus'

const props = defineProps<{
  issueKey: string
}>()

const emit = defineEmits<{
  success: []
}>()

const uploadRef = ref<UploadInstance>()

const uploadUrl = computed(() => {
  return `/api/v1/issues/${props.issueKey}/attachments`
})

const uploadHeaders = computed(() => {
  const token = localStorage.getItem('token')
  return {
    Authorization: `Bearer ${token}`
  }
})

const beforeUpload = (file: File) => {
  // 检查文件大小（10MB）
  const maxSize = 10 * 1024 * 1024
  if (file.size > maxSize) {
    ElMessage.error('文件大小不能超过10MB')
    return false
  }

  // 检查文件类型
  const allowedTypes = [
    '.jpg', '.jpeg', '.png', '.gif', '.bmp', '.webp',
    '.pdf', '.doc', '.docx', '.xls', '.xlsx', '.ppt', '.pptx',
    '.txt', '.md', '.csv',
    '.zip', '.rar', '.7z', '.tar', '.gz',
    '.log', '.json', '.xml', '.yaml', '.yml'
  ]

  const ext = file.name.substring(file.name.lastIndexOf('.')).toLowerCase()
  if (!allowedTypes.includes(ext)) {
    ElMessage.error('不支持的文件类型')
    return false
  }

  return true
}

const handleSuccess = () => {
  ElMessage.success('上传成功')
  emit('success')
}

const handleError = () => {
  ElMessage.error('上传失败')
}
</script>

<style scoped>
.attachment-upload {
  margin-bottom: 20px;
}

.el-icon--upload {
  font-size: 67px;
  color: #8c939d;
  margin: 40px 0 16px;
  line-height: 50px;
}

.el-upload__text {
  color: #606266;
  font-size: 14px;
  text-align: center;
}

.el-upload__text em {
  color: #409eff;
  font-style: normal;
}

.el-upload__tip {
  font-size: 12px;
  color: #909399;
  margin-top: 7px;
}
</style>
