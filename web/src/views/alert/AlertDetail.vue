<template>
  <div class="alert-detail-container">
    <el-page-header style="margin-bottom: 20px" @back="$router.back()">
      <template #content>
        <span style="font-size: 18px; font-weight: 500">告警详情</span>
      </template>
    </el-page-header>

    <el-card v-loading="loading" shadow="never">
      <template #header>
        <div style="display: flex; justify-content: space-between; align-items: center">
          <div>
            <h2 style="margin: 0">{{ alert?.alert_name }}</h2>
            <div style="margin-top: 8px">
              <el-tag :type="getSeverityType(alert?.severity)" size="small" style="margin-right: 8px">
                {{ getSeverityText(alert?.severity) }}
              </el-tag>
              <el-tag :type="getStatusType(alert?.status)" size="small">
                {{ getStatusText(alert?.status) }}
              </el-tag>
            </div>
          </div>
          <div>
            <el-button
              v-if="alert?.status === 'firing' && !alert?.ack_at"
              type="primary"
              @click="handleAck"
            >
              确认告警
            </el-button>
            <el-button
              v-if="alert?.status === 'firing'"
              type="success"
              @click="handleResolve"
            >
              解决告警
            </el-button>
          </div>
        </div>
      </template>

      <el-descriptions :column="2" border>
        <el-descriptions-item label="告警指纹">
          <el-text type="info" size="small">{{ alert?.fingerprint }}</el-text>
        </el-descriptions-item>
        <el-descriptions-item label="告警来源">
          {{ alert?.source }}
        </el-descriptions-item>
        <el-descriptions-item label="开始时间">
          {{ formatTime(alert?.starts_at) }}
        </el-descriptions-item>
        <el-descriptions-item label="结束时间">
          {{ alert?.ends_at ? formatTime(alert.ends_at) : '-' }}
        </el-descriptions-item>
        <el-descriptions-item label="关联工单">
          <el-link
            v-if="alert?.issue_key"
            type="primary"
            :href="`/issues/${alert.issue_key}`"
            target="_blank"
          >
            {{ alert.issue_key }}
          </el-link>
          <span v-else>-</span>
        </el-descriptions-item>
        <el-descriptions-item label="确认信息">
          <div v-if="alert?.ack_at">
            <div>{{ alert.ack_by_name }} 于 {{ formatTime(alert.ack_at) }} 确认</div>
          </div>
          <span v-else>-</span>
        </el-descriptions-item>
      </el-descriptions>

      <el-divider />

      <h3>标签</h3>
      <div style="margin-top: 12px">
        <el-tag
          v-for="(value, key) in alert?.labels"
          :key="key"
          style="margin-right: 8px; margin-bottom: 8px"
        >
          {{ key }} = {{ value }}
        </el-tag>
      </div>

      <el-divider />

      <h3>注解</h3>
      <div style="margin-top: 12px">
        <el-descriptions :column="1" border>
          <el-descriptions-item
            v-for="(value, key) in alert?.annotations"
            :key="key"
            :label="key"
          >
            {{ value }}
          </el-descriptions-item>
        </el-descriptions>
      </div>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { getAlertDetail, ackAlert, resolveAlert } from '@/api/alert'
import type { Alert } from '@/types/alert'
import dayjs from 'dayjs'

const route = useRoute()
// useRouter is available globally in template as $router, but we call it to ensure Vue Router is setup
useRouter()

const loading = ref(false)
const alert = ref<Alert>()

const loadData = async () => {
  loading.value = true
  try {
    const { data } = await getAlertDetail(Number(route.params.id), { _redirectOn404: true })
    alert.value = data.data
  } catch {
    // ignored
  } finally {
    loading.value = false
  }
}

const handleAck = async () => {
  try {
    await ackAlert(alert.value!.id)
    ElMessage.success('确认成功')
    loadData()
  } catch {
    // ignored
  }
}

const handleResolve = async () => {
  try {
    await resolveAlert(alert.value!.id)
    ElMessage.success('解决成功')
    loadData()
  } catch {
    // ignored
  }
}

const getSeverityType = (severity?: string) => {
  const map: Record<string, any> = {
    critical: 'danger',
    warning: 'warning',
    info: 'info',
  }
  return map[severity || ''] || 'info'
}

const getSeverityText = (severity?: string) => {
  const map: Record<string, string> = {
    critical: '严重',
    warning: '警告',
    info: '信息',
  }
  return map[severity || ''] || severity
}

const getStatusType = (status?: string) => {
  return status === 'firing' ? 'danger' : 'success'
}

const getStatusText = (status?: string) => {
  const map: Record<string, string> = {
    firing: '触发中',
    resolved: '已解决',
  }
  return map[status || ''] || status
}

const formatTime = (time?: string) => {
  return time ? dayjs(time).format('YYYY-MM-DD HH:mm:ss') : '-'
}

onMounted(() => {
  loadData()
})
</script>

<style scoped lang="scss">
.alert-detail-container {
  h3 {
    font-size: 16px;
    font-weight: 500;
    margin-bottom: 12px;
  }
}
</style>
