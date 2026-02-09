<template>
  <div class="alert-list-container">
    <!-- 页面头部 -->
    <div class="page-header">
      <div class="header-info">
        <div class="header-icon">
          <el-icon><Bell /></el-icon>
        </div>
        <div class="header-text">
          <h1 class="header-title">告警列表</h1>
          <p class="header-desc">监控和管理所有告警事件</p>
        </div>
      </div>
    </div>

    <!-- 统计卡片 -->
    <el-row :gutter="16" class="stat-row">
      <el-col :xs="12" :sm="8" :md="4">
        <div class="stat-card total" :class="{ active: !queryParams.status && !queryParams.severity }" @click="handleStatClick()">
          <div class="stat-icon-wrapper">
            <el-icon :size="22"><Bell /></el-icon>
          </div>
          <div class="stat-content">
            <div class="stat-value">{{ stats.total }}</div>
            <div class="stat-label">总告警数</div>
          </div>
        </div>
      </el-col>
      <el-col :xs="12" :sm="8" :md="4">
        <div class="stat-card firing" :class="{ active: queryParams.status === 'firing' && !queryParams.severity }" @click="handleStatClick('firing')">
          <div class="stat-icon-wrapper">
            <el-icon :size="22"><Promotion /></el-icon>
          </div>
          <div class="stat-content">
            <div class="stat-value">{{ stats.firing }}</div>
            <div class="stat-label">活跃告警</div>
          </div>
        </div>
      </el-col>
      <el-col :xs="12" :sm="8" :md="4">
        <div class="stat-card resolved" :class="{ active: queryParams.status === 'resolved' && !queryParams.severity }" @click="handleStatClick('resolved')">
          <div class="stat-icon-wrapper">
            <el-icon :size="22"><CircleCheck /></el-icon>
          </div>
          <div class="stat-content">
            <div class="stat-value">{{ stats.resolved }}</div>
            <div class="stat-label">已解决</div>
          </div>
        </div>
      </el-col>
      <el-col :xs="12" :sm="8" :md="4">
        <div class="stat-card critical" :class="{ active: queryParams.severity === 'critical' && !queryParams.status }" @click="handleStatClick(undefined, 'critical')">
          <div class="stat-icon-wrapper">
            <el-icon :size="22"><WarningFilled /></el-icon>
          </div>
          <div class="stat-content">
            <div class="stat-value">{{ stats.critical }}</div>
            <div class="stat-label">严重</div>
          </div>
        </div>
      </el-col>
      <el-col :xs="12" :sm="8" :md="4">
        <div class="stat-card warning" :class="{ active: queryParams.severity === 'warning' && !queryParams.status }" @click="handleStatClick(undefined, 'warning')">
          <div class="stat-icon-wrapper">
            <el-icon :size="22"><Warning /></el-icon>
          </div>
          <div class="stat-content">
            <div class="stat-value">{{ stats.warning }}</div>
            <div class="stat-label">警告</div>
          </div>
        </div>
      </el-col>
      <el-col :xs="12" :sm="8" :md="4">
        <div class="stat-card info" :class="{ active: queryParams.severity === 'info' && !queryParams.status }" @click="handleStatClick(undefined, 'info')">
          <div class="stat-icon-wrapper">
            <el-icon :size="22"><InfoFilled /></el-icon>
          </div>
          <div class="stat-content">
            <div class="stat-value">{{ stats.info }}</div>
            <div class="stat-label">信息</div>
          </div>
        </div>
      </el-col>
    </el-row>

    <!-- 筛选器 -->
    <el-card shadow="never" class="filter-card">
      <div class="filter-content">
        <div class="filter-left">
          <el-input
            v-model="queryParams.alert_name"
            placeholder="搜索告警名称"
            clearable
            class="search-input"
            @clear="handleQuery"
            @keyup.enter="handleQuery"
          >
            <template #prefix>
              <el-icon><Search /></el-icon>
            </template>
          </el-input>
          <el-select v-model="queryParams.status" placeholder="状态" clearable class="filter-select" @change="handleQuery">
            <el-option label="触发中" value="firing" />
            <el-option label="已解决" value="resolved" />
          </el-select>
          <el-select v-model="queryParams.severity" placeholder="严重程度" clearable class="filter-select" @change="handleQuery">
            <el-option label="严重" value="critical" />
            <el-option label="警告" value="warning" />
            <el-option label="信息" value="info" />
          </el-select>
        </div>
        <div class="filter-right">
          <el-button :icon="Refresh" @click="handleReset">重置</el-button>
        </div>
      </div>
    </el-card>

    <!-- 内容区 -->
    <el-card shadow="never" class="table-card">
      <div class="toolbar">
        <div class="toolbar-left">
          <el-radio-group v-model="viewMode" @change="handleViewModeChange" class="view-toggle">
            <el-radio-button value="list">
              <el-icon><List /></el-icon>
              列表
            </el-radio-button>
            <el-radio-button value="group">
              <el-icon><Grid /></el-icon>
              分组
            </el-radio-button>
          </el-radio-group>
        </div>
        <div v-if="viewMode === 'group'" class="toolbar-right">
          <span class="group-label">分组字段:</span>
          <el-select v-model="groupBy" size="default" class="group-select" @change="loadGroupData">
            <el-option label="集群" value="cluster" />
            <el-option label="命名空间" value="namespace" />
            <el-option label="服务" value="service" />
            <el-option label="实例" value="instance" />
          </el-select>
        </div>
      </div>

      <!-- 列表视图 -->
      <div v-if="viewMode === 'list'">
        <el-table
          v-loading="loading"
          :data="alertList"
          style="width: 100%"
          :row-class-name="() => 'clickable-row'"
          @row-click="handleRowClick"
        >
          <el-table-column prop="alert_name" label="告警名称" min-width="200">
            <template #default="{ row }">
              <div class="alert-name-cell">
                <div class="severity-indicator" :class="row.severity"></div>
                <div class="alert-info">
                  <div class="alert-name">{{ row.alert_name }}</div>
                  <div class="alert-fingerprint">{{ row.fingerprint.substring(0, 16) }}...</div>
                </div>
              </div>
            </template>
          </el-table-column>
          <el-table-column prop="severity" label="严重程度" width="100" align="center">
            <template #default="{ row }">
              <el-tag :type="getSeverityType(row.severity)" size="small" effect="dark">
                {{ getSeverityText(row.severity) }}
              </el-tag>
            </template>
          </el-table-column>
          <el-table-column prop="status" label="状态" width="100" align="center">
            <template #default="{ row }">
              <div class="status-badge" :class="row.status">
                <span class="status-dot"></span>
                <span>{{ getStatusText(row.status) }}</span>
              </div>
            </template>
          </el-table-column>
          <el-table-column label="标签" min-width="200">
            <template #default="{ row }">
              <div class="labels-cell">
                <el-tag
                  v-for="(value, key) in getMainLabels(row.labels)"
                  :key="key"
                  size="small"
                  effect="plain"
                  type="info"
                  class="label-tag"
                >
                  {{ key }}={{ value }}
                </el-tag>
              </div>
            </template>
          </el-table-column>
          <el-table-column prop="issue_key" label="关联工单" width="120" align="center">
            <template #default="{ row }">
              <el-link v-if="row.issue_key" type="primary" @click.stop="$router.push(`/issues/${row.issue_key}`)">
                {{ row.issue_key }}
              </el-link>
              <span v-else class="text-muted">-</span>
            </template>
          </el-table-column>
          <el-table-column prop="starts_at" label="开始时间" width="160">
            <template #default="{ row }">
              <div class="time-cell">
                <el-icon><Clock /></el-icon>
                <span>{{ formatTime(row.starts_at) }}</span>
              </div>
            </template>
          </el-table-column>
          <el-table-column label="操作" width="160" fixed="right" align="center">
            <template #default="{ row }">
              <div class="action-buttons">
                <el-tooltip v-if="row.status === 'firing' && !row.ack_at" content="确认" placement="top">
                  <el-button link type="primary" @click.stop="handleAck(row)">
                    <el-icon><Check /></el-icon>
                  </el-button>
                </el-tooltip>
                <el-tooltip v-if="row.status === 'firing'" content="解决" placement="top">
                  <el-button link type="success" @click.stop="handleResolve(row)">
                    <el-icon><CircleCheck /></el-icon>
                  </el-button>
                </el-tooltip>
                <el-tooltip content="详情" placement="top">
                  <el-button link type="info" @click.stop="handleViewDetail(row)">
                    <el-icon><View /></el-icon>
                  </el-button>
                </el-tooltip>
              </div>
            </template>
          </el-table-column>
        </el-table>

        <div class="pagination-wrapper">
          <el-pagination
            v-model:current-page="queryParams.page"
            v-model:page-size="queryParams.page_size"
            :total="total"
            :page-sizes="[10, 20, 50, 100]"
            layout="total, sizes, prev, pager, next, jumper"
            @size-change="handleQuery"
            @current-change="handleQuery"
          />
        </div>
      </div>

      <!-- 分组视图 -->
      <div v-if="viewMode === 'group'" v-loading="loading">
        <el-row :gutter="20">
          <el-col
            v-for="group in groupData"
            :key="group.group_value"
            :xs="24"
            :sm="12"
            :md="8"
            :lg="6"
            class="group-col"
          >
            <div class="group-card">
              <div class="group-name">{{ group.group_value }}</div>
              <div class="group-total">{{ group.count }}</div>
              <div class="group-divider"></div>
              <div class="group-severity-row">
                <div class="severity-item critical">
                  <span class="severity-dot"></span>
                  <span>严重 {{ group.severity.critical || 0 }}</span>
                </div>
                <div class="severity-item warning">
                  <span class="severity-dot"></span>
                  <span>警告 {{ group.severity.warning || 0 }}</span>
                </div>
                <div class="severity-item info">
                  <span class="severity-dot"></span>
                  <span>信息 {{ group.severity.info || 0 }}</span>
                </div>
              </div>
            </div>
          </el-col>
        </el-row>
      </div>
    </el-card>

    <!-- 确认对话框 -->
    <el-dialog v-model="ackDialogVisible" title="确认告警" width="460px" class="alert-dialog">
      <div class="dialog-icon-header">
        <div class="dialog-icon ack">
          <el-icon><Check /></el-icon>
        </div>
        <p class="dialog-tip">确认此告警已知悉并开始处理</p>
      </div>
      <el-form :model="ackForm" label-position="top">
        <el-form-item label="备注">
          <el-input
            v-model="ackForm.comment"
            type="textarea"
            :rows="3"
            placeholder="请输入确认备注（可选）"
          />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="ackDialogVisible = false">取消</el-button>
        <el-button type="primary" @click="confirmAck">
          <el-icon><Check /></el-icon>
          确认
        </el-button>
      </template>
    </el-dialog>

    <!-- 解决对话框 -->
    <el-dialog v-model="resolveDialogVisible" title="解决告警" width="460px" class="alert-dialog">
      <div class="dialog-icon-header">
        <div class="dialog-icon resolve">
          <el-icon><CircleCheck /></el-icon>
        </div>
        <p class="dialog-tip">标记此告警为已解决</p>
      </div>
      <el-form :model="resolveForm" label-position="top">
        <el-form-item label="备注">
          <el-input
            v-model="resolveForm.comment"
            type="textarea"
            :rows="3"
            placeholder="请输入解决备注（可选）"
          />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="resolveDialogVisible = false">取消</el-button>
        <el-button type="success" @click="confirmResolve">
          <el-icon><CircleCheck /></el-icon>
          解决
        </el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import {
  Search, Refresh, List, Grid, Bell, Clock, Check, CircleCheck,
  View, WarningFilled, Warning, Promotion, InfoFilled
} from '@element-plus/icons-vue'
import { getAlertList, getAlertStats, ackAlert, resolveAlert, getAlertGroup } from '@/api/alert'
import type { Alert, AlertGroupItem, AlertStatsResponse } from '@/types/alert'
import dayjs from 'dayjs'

const router = useRouter()

const loading = ref(false)
const alertList = ref<Alert[]>([])
const total = ref(0)
const viewMode = ref<'list' | 'group'>('list')
const groupBy = ref('cluster')
const groupData = ref<AlertGroupItem[]>([])

const stats = reactive<AlertStatsResponse>({ total: 0, firing: 0, resolved: 0, critical: 0, warning: 0, info: 0 })

const queryParams = reactive({
  page: 1, page_size: 20,
  status: undefined as 'firing' | 'resolved' | undefined,
  severity: undefined as 'critical' | 'warning' | 'info' | undefined,
  alert_name: undefined as string | undefined,
})

const ackDialogVisible = ref(false)
const ackForm = reactive({ id: 0, comment: '' })
const resolveDialogVisible = ref(false)
const resolveForm = reactive({ id: 0, comment: '' })

const loadData = async () => {
  loading.value = true
  try {
    const { data } = await getAlertList(queryParams)
    alertList.value = data.data.items
    total.value = data.data.total
  } catch (error) {
    console.error('Failed to load alerts:', error)
  } finally {
    loading.value = false
  }
}

const loadStats = async () => {
  try {
    const { data } = await getAlertStats()
    Object.assign(stats, data.data)
  } catch (error) {
    console.error('Failed to load alert stats:', error)
  }
}

const loadGroupData = async () => {
  loading.value = true
  try {
    const { data } = await getAlertGroup(groupBy.value, {
      status: queryParams.status, severity: queryParams.severity,
    })
    groupData.value = data.data.items
  } catch (error) {
    console.error('Failed to load group data:', error)
  } finally {
    loading.value = false
  }
}

const handleQuery = () => {
  queryParams.page = 1
  viewMode.value === 'list' ? loadData() : loadGroupData()
}

const handleStatClick = (status?: 'firing' | 'resolved', severity?: 'critical' | 'warning' | 'info') => {
  queryParams.status = status
  queryParams.severity = severity
  handleQuery()
}

const handleReset = () => {
  queryParams.page = 1; queryParams.page_size = 20
  queryParams.status = undefined; queryParams.severity = undefined; queryParams.alert_name = undefined
  handleQuery()
}

const handleViewModeChange = () => { viewMode.value === 'list' ? loadData() : loadGroupData() }
const handleRowClick = (row: Alert) => { router.push(`/alerts/${row.id}`) }
const handleViewDetail = (row: Alert) => { router.push(`/alerts/${row.id}`) }

const handleAck = (row: Alert) => { ackForm.id = row.id; ackForm.comment = ''; ackDialogVisible.value = true }
const confirmAck = async () => {
  try {
    await ackAlert(ackForm.id, ackForm.comment)
    ElMessage.success('确认成功'); ackDialogVisible.value = false; loadData()
  } catch (error) { console.error(error) }
}

const handleResolve = (row: Alert) => { resolveForm.id = row.id; resolveForm.comment = ''; resolveDialogVisible.value = true }
const confirmResolve = async () => {
  try {
    await resolveAlert(resolveForm.id, resolveForm.comment)
    ElMessage.success('解决成功'); resolveDialogVisible.value = false; loadData()
  } catch (error) { console.error(error) }
}

const getSeverityType = (severity: string) => {
  const map: Record<string, any> = { critical: 'danger', warning: 'warning', info: 'info' }
  return map[severity] || 'info'
}
const getSeverityText = (severity: string) => {
  const map: Record<string, string> = { critical: '严重', warning: '警告', info: '信息' }
  return map[severity] || severity
}
const getStatusText = (status: string) => {
  const map: Record<string, string> = { firing: '触发中', resolved: '已解决' }
  return map[status] || status
}
const getMainLabels = (labels: Record<string, string>) => {
  const mainKeys = ['cluster', 'namespace', 'service', 'instance']
  const result: Record<string, string> = {}
  mainKeys.forEach((key) => { if (labels[key]) result[key] = labels[key] })
  return result
}
const formatTime = (time: string) => dayjs(time).format('YYYY-MM-DD HH:mm')

onMounted(() => { loadData(); loadStats() })
</script>

<style scoped lang="scss">
.alert-list-container {
  width: 100%;
}

// 页面头部
.page-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 24px;
  padding: 24px 32px;
  background: linear-gradient(135deg, #f5576c 0%, #ff6b6b 100%);
  border-radius: 12px;
  color: #fff;

  .header-info {
    display: flex;
    align-items: center;
    gap: 16px;
  }

  .header-icon {
    width: 56px;
    height: 56px;
    background: rgba(255, 255, 255, 0.2);
    border-radius: 14px;
    display: flex;
    align-items: center;
    justify-content: center;
    font-size: 28px;
  }

  .header-text {
    .header-title { font-size: 22px; font-weight: 600; margin: 0 0 4px 0; }
    .header-desc { font-size: 14px; margin: 0; opacity: 0.9; }
  }
}

// 统计卡片
.stat-row { margin-bottom: 20px; }

.stat-card {
  display: flex;
  align-items: center;
  gap: 14px;
  padding: 16px 18px;
  background: #fff;
  border-radius: 12px;
  box-shadow: 0 2px 12px rgba(0, 0, 0, 0.04);
  transition: all 0.3s;
  cursor: pointer;
  border: 2px solid transparent;

  &:hover { transform: translateY(-2px); box-shadow: 0 4px 16px rgba(0, 0, 0, 0.08); }
  &.active { border-color: currentColor; }

  .stat-icon-wrapper {
    width: 44px; height: 44px; border-radius: 12px;
    display: flex; align-items: center; justify-content: center; color: #fff;
    flex-shrink: 0;
  }

  .stat-content {
    .stat-value { font-size: 24px; font-weight: 700; line-height: 1.2; }
    .stat-label { font-size: 12px; color: #909399; margin-top: 2px; }
  }

  &.total {
    .stat-icon-wrapper { background: linear-gradient(135deg, #667eea 0%, #764ba2 100%); }
    .stat-value { color: #667eea; }
    &.active { border-color: #667eea; }
  }
  &.firing {
    .stat-icon-wrapper { background: linear-gradient(135deg, #f56c6c 0%, #e74c3c 100%); }
    .stat-value { color: #e74c3c; }
    &.active { border-color: #e74c3c; }
  }
  &.resolved {
    .stat-icon-wrapper { background: linear-gradient(135deg, #67c23a 0%, #2ecc71 100%); }
    .stat-value { color: #2ecc71; }
    &.active { border-color: #2ecc71; }
  }
  &.critical {
    .stat-icon-wrapper { background: linear-gradient(135deg, #f56c6c 0%, #c0392b 100%); }
    .stat-value { color: #c0392b; }
    &.active { border-color: #c0392b; }
  }
  &.warning {
    .stat-icon-wrapper { background: linear-gradient(135deg, #e6a23c 0%, #d68910 100%); }
    .stat-value { color: #d68910; }
    &.active { border-color: #d68910; }
  }
  &.info {
    .stat-icon-wrapper { background: linear-gradient(135deg, #909399 0%, #6b7280 100%); }
    .stat-value { color: #6b7280; }
    &.active { border-color: #6b7280; }
  }
}

// 筛选
.filter-card {
  margin-bottom: 20px;
  border-radius: 12px;

  :deep(.el-card__body) { padding: 16px 20px; }

  .filter-content {
    display: flex;
    justify-content: space-between;
    align-items: center;
    flex-wrap: wrap;
    gap: 12px;
  }
  .filter-left { display: flex; align-items: center; gap: 10px; flex-wrap: wrap; }
  .search-input { width: 220px; }
  .filter-select { width: 130px; }
}

// 表格卡片
.table-card {
  border-radius: 12px;
  :deep(.el-card__body) { padding: 20px; }
}

.toolbar {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 16px;

  .toolbar-right {
    display: flex;
    align-items: center;
    gap: 8px;

    .group-label { font-size: 13px; color: #6b7280; }
    .group-select { width: 120px; }
  }
}

// 表格
:deep(.el-table) {
  border-radius: 8px;

  th.el-table__cell {
    background: #f8fafc;
    font-weight: 600;
    color: #374151;
  }

  .clickable-row {
    cursor: pointer;
    &:hover { background-color: #f9fafb; }
  }
}

.alert-name-cell {
  display: flex;
  align-items: center;
  gap: 10px;

  .severity-indicator {
    width: 4px;
    height: 36px;
    border-radius: 2px;
    flex-shrink: 0;

    &.critical { background: #ef4444; }
    &.warning { background: #f59e0b; }
    &.info { background: #6b7280; }
  }

  .alert-info {
    .alert-name { font-weight: 500; color: #1f2937; font-size: 14px; }
    .alert-fingerprint { font-size: 12px; color: #9ca3af; margin-top: 2px; }
  }
}

.status-badge {
  display: inline-flex;
  align-items: center;
  gap: 5px;
  padding: 3px 10px;
  border-radius: 20px;
  font-size: 12px;

  .status-dot { width: 6px; height: 6px; border-radius: 50%; }

  &.firing { background: #fef2f2; color: #dc2626; .status-dot { background: #ef4444; } }
  &.resolved { background: #ecfdf5; color: #059669; .status-dot { background: #10b981; } }
}

.labels-cell {
  display: flex;
  flex-wrap: wrap;
  gap: 4px;
}

.text-muted { color: #d1d5db; }

.time-cell {
  display: flex;
  align-items: center;
  gap: 6px;
  font-size: 13px;
  color: #6b7280;
  .el-icon { font-size: 14px; color: #9ca3af; }
}

.action-buttons {
  display: flex;
  justify-content: center;
  gap: 8px;

  .el-button {
    font-size: 16px;
    padding: 4px;
    &:hover { background: #f3f4f6; border-radius: 6px; }
  }
}

.pagination-wrapper {
  padding-top: 20px;
  display: flex;
  justify-content: flex-end;
  border-top: 1px solid #f0f0f0;
  margin-top: 16px;
}

// 分组卡片
.group-col { margin-bottom: 20px; }

.group-card {
  background: #fff;
  border-radius: 12px;
  padding: 20px;
  border: 1px solid #f0f0f0;
  transition: transform 0.3s, box-shadow 0.3s;
  cursor: pointer;

  &:hover {
    transform: translateY(-4px);
    box-shadow: 0 8px 24px rgba(0, 0, 0, 0.1);
  }

  .group-name {
    font-size: 14px;
    font-weight: 600;
    color: #374151;
    margin-bottom: 8px;
  }

  .group-total {
    font-size: 36px;
    font-weight: 700;
    color: #667eea;
    margin-bottom: 12px;
  }

  .group-divider {
    height: 1px;
    background: #f0f0f0;
    margin-bottom: 12px;
  }

  .group-severity-row {
    display: flex;
    justify-content: space-between;
    font-size: 12px;
  }

  .severity-item {
    display: flex;
    align-items: center;
    gap: 4px;

    .severity-dot {
      width: 8px;
      height: 8px;
      border-radius: 50%;
    }

    &.critical { color: #dc2626; .severity-dot { background: #ef4444; } }
    &.warning { color: #d97706; .severity-dot { background: #f59e0b; } }
    &.info { color: #6b7280; .severity-dot { background: #9ca3af; } }
  }
}

// 对话框
.alert-dialog {
  .dialog-icon-header {
    text-align: center;
    margin-bottom: 20px;

    .dialog-icon {
      width: 56px; height: 56px;
      margin: 0 auto 12px;
      border-radius: 14px;
      display: flex; align-items: center; justify-content: center;
      font-size: 24px; color: #fff;

      &.ack { background: linear-gradient(135deg, #667eea 0%, #764ba2 100%); }
      &.resolve { background: linear-gradient(135deg, #11998e 0%, #38ef7d 100%); }
    }

    .dialog-tip { font-size: 14px; color: #6b7280; margin: 0; }
  }
}

// 响应式
@media (max-width: 768px) {
  .filter-card .filter-content { flex-direction: column; align-items: stretch; }
  .filter-card .filter-left { flex-direction: column; }
  .search-input, .filter-select { width: 100% !important; }
}
</style>
