<template>
  <div class="alert-list-container">
    <!-- 统计卡片 -->
    <el-row :gutter="20" style="margin-bottom: 20px">
      <el-col :xs="24" :sm="12" :md="6">
        <div class="stat-card">
          <div class="stat-value">{{ stats.total }}</div>
          <div class="stat-label">总告警数</div>
        </div>
      </el-col>
      <el-col :xs="24" :sm="12" :md="6">
        <div class="stat-card critical">
          <div class="stat-value">{{ stats.critical }}</div>
          <div class="stat-label">严重告警</div>
        </div>
      </el-col>
      <el-col :xs="24" :sm="12" :md="6">
        <div class="stat-card warning">
          <div class="stat-value">{{ stats.warning }}</div>
          <div class="stat-label">警告告警</div>
        </div>
      </el-col>
      <el-col :xs="24" :sm="12" :md="6">
        <div class="stat-card info">
          <div class="stat-value">{{ stats.firing }}</div>
          <div class="stat-label">触发中</div>
        </div>
      </el-col>
    </el-row>

    <!-- 过滤器 -->
    <el-card shadow="never" style="margin-bottom: 20px">
      <el-form :inline="true" :model="queryParams">
        <el-form-item label="状态">
          <el-select v-model="queryParams.status" placeholder="全部" clearable style="width: 120px" @change="handleQuery">
            <el-option label="触发中" value="firing" />
            <el-option label="已解决" value="resolved" />
          </el-select>
        </el-form-item>
        <el-form-item label="严重程度">
          <el-select v-model="queryParams.severity" placeholder="全部" clearable style="width: 120px" @change="handleQuery">
            <el-option label="严重" value="critical" />
            <el-option label="警告" value="warning" />
            <el-option label="信息" value="info" />
          </el-select>
        </el-form-item>
        <el-form-item label="告警名称">
          <el-input v-model="queryParams.alert_name" placeholder="搜索告警名称" clearable style="width: 200px" @clear="handleQuery" @keyup.enter="handleQuery" />
        </el-form-item>
        <el-form-item>
          <el-button type="primary" :icon="Search" @click="handleQuery">查询</el-button>
          <el-button :icon="Refresh" @click="handleReset">重置</el-button>
        </el-form-item>
      </el-form>
    </el-card>

    <!-- 视图切换 -->
    <el-card shadow="never">
      <div style="margin-bottom: 16px; display: flex; justify-content: space-between; align-items: center">
        <el-radio-group v-model="viewMode" @change="handleViewModeChange">
          <el-radio-button label="list">
            <el-icon><List /></el-icon>
            列表视图
          </el-radio-button>
          <el-radio-button label="group">
            <el-icon><Grid /></el-icon>
            分组视图
          </el-radio-button>
        </el-radio-group>

        <div v-if="viewMode === 'group'">
          <span style="margin-right: 8px">分组字段:</span>
          <el-select v-model="groupBy" style="width: 150px" @change="loadGroupData">
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
          stripe
          style="width: 100%"
          @row-click="handleRowClick"
        >
          <el-table-column prop="alert_name" label="告警名称" min-width="180">
            <template #default="{ row }">
              <div style="font-weight: 500">{{ row.alert_name }}</div>
              <div style="font-size: 12px; color: #909399; margin-top: 4px">
                {{ row.fingerprint.substring(0, 16) }}...
              </div>
            </template>
          </el-table-column>
          <el-table-column prop="severity" label="严重程度" width="100">
            <template #default="{ row }">
              <el-tag :type="getSeverityType(row.severity)" size="small">
                {{ getSeverityText(row.severity) }}
              </el-tag>
            </template>
          </el-table-column>
          <el-table-column prop="status" label="状态" width="100">
            <template #default="{ row }">
              <el-tag :type="getStatusType(row.status)" size="small">
                {{ getStatusText(row.status) }}
              </el-tag>
            </template>
          </el-table-column>
          <el-table-column label="标签" min-width="200">
            <template #default="{ row }">
              <el-tag
                v-for="(value, key) in getMainLabels(row.labels)"
                :key="key"
                size="small"
                style="margin-right: 4px; margin-bottom: 4px"
              >
                {{ key }}={{ value }}
              </el-tag>
            </template>
          </el-table-column>
          <el-table-column prop="issue_key" label="关联工单" width="120">
            <template #default="{ row }">
              <el-link v-if="row.issue_key" type="primary" :href="`/issues/${row.issue_key}`" target="_blank">
                {{ row.issue_key }}
              </el-link>
              <span v-else style="color: #909399">-</span>
            </template>
          </el-table-column>
          <el-table-column prop="starts_at" label="开始时间" width="160">
            <template #default="{ row }">
              {{ formatTime(row.starts_at) }}
            </template>
          </el-table-column>
          <el-table-column label="操作" width="180" fixed="right">
            <template #default="{ row }">
              <el-button
                v-if="row.status === 'firing' && !row.ack_at"
                link
                type="primary"
                size="small"
                @click.stop="handleAck(row)"
              >
                确认
              </el-button>
              <el-button
                v-if="row.status === 'firing'"
                link
                type="success"
                size="small"
                @click.stop="handleResolve(row)"
              >
                解决
              </el-button>
              <el-button link type="info" size="small" @click.stop="handleViewDetail(row)">
                详情
              </el-button>
            </template>
          </el-table-column>
        </el-table>

        <el-pagination
          v-model:current-page="queryParams.page"
          v-model:page-size="queryParams.page_size"
          :total="total"
          :page-sizes="[10, 20, 50, 100]"
          layout="total, sizes, prev, pager, next, jumper"
          style="margin-top: 20px; justify-content: flex-end"
          @size-change="handleQuery"
          @current-change="handleQuery"
        />
      </div>

      <!-- 分组视图 -->
      <div v-if="viewMode === 'group'">
        <el-row :gutter="20">
          <el-col
            v-for="group in groupData"
            :key="group.group_value"
            :xs="24"
            :sm="12"
            :md="8"
            :lg="6"
            style="margin-bottom: 20px"
          >
            <el-card shadow="hover" :body-style="{ padding: '20px' }">
              <div style="font-size: 16px; font-weight: 500; margin-bottom: 12px">
                {{ group.group_value }}
              </div>
              <div style="font-size: 32px; font-weight: 700; color: #409eff; margin-bottom: 12px">
                {{ group.count }}
              </div>
              <el-divider style="margin: 12px 0" />
              <div style="display: flex; justify-content: space-between; font-size: 12px">
                <div>
                  <el-tag type="danger" size="small">严重: {{ group.severity.critical || 0 }}</el-tag>
                </div>
                <div>
                  <el-tag type="warning" size="small">警告: {{ group.severity.warning || 0 }}</el-tag>
                </div>
                <div>
                  <el-tag type="info" size="small">信息: {{ group.severity.info || 0 }}</el-tag>
                </div>
              </div>
            </el-card>
          </el-col>
        </el-row>
      </div>
    </el-card>

    <!-- 确认对话框 -->
    <el-dialog v-model="ackDialogVisible" title="确认告警" width="500px">
      <el-form :model="ackForm" label-width="80px">
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
        <el-button type="primary" @click="confirmAck">确定</el-button>
      </template>
    </el-dialog>

    <!-- 解决对话框 -->
    <el-dialog v-model="resolveDialogVisible" title="解决告警" width="500px">
      <el-form :model="resolveForm" label-width="80px">
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
        <el-button type="primary" @click="confirmResolve">确定</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { Search, Refresh, List, Grid } from '@element-plus/icons-vue'
import { getAlertList, ackAlert, resolveAlert, getAlertGroup } from '@/api/alert'
import type { Alert, AlertGroupItem } from '@/types/alert'
import dayjs from 'dayjs'

const router = useRouter()

// 数据
const loading = ref(false)
const alertList = ref<Alert[]>([])
const total = ref(0)
const viewMode = ref<'list' | 'group'>('list')
const groupBy = ref('cluster')
const groupData = ref<AlertGroupItem[]>([])

// 统计数据
const stats = reactive({
  total: 0,
  critical: 0,
  warning: 0,
  firing: 0,
})

// 查询参数
const queryParams = reactive({
  page: 1,
  page_size: 20,
  status: undefined as string | undefined,
  severity: undefined as string | undefined,
  alert_name: undefined as string | undefined,
})

// 确认对话框
const ackDialogVisible = ref(false)
const ackForm = reactive({
  id: 0,
  comment: '',
})

// 解决对话框
const resolveDialogVisible = ref(false)
const resolveForm = reactive({
  id: 0,
  comment: '',
})

// 加载数据
const loadData = async () => {
  loading.value = true
  try {
    const { data } = await getAlertList(queryParams)
    alertList.value = data.data.items
    total.value = data.data.total

    // 更新统计数据
    updateStats()
  } catch (error) {
    console.error('Failed to load alerts:', error)
  } finally {
    loading.value = false
  }
}

// 加载分组数据
const loadGroupData = async () => {
  loading.value = true
  try {
    const { data } = await getAlertGroup(groupBy.value, {
      status: queryParams.status,
      severity: queryParams.severity,
    })
    groupData.value = data.data.items
  } catch (error) {
    console.error('Failed to load group data:', error)
  } finally {
    loading.value = false
  }
}

// 更新统计数据
const updateStats = () => {
  stats.total = total.value
  stats.critical = alertList.value.filter((a) => a.severity === 'critical').length
  stats.warning = alertList.value.filter((a) => a.severity === 'warning').length
  stats.firing = alertList.value.filter((a) => a.status === 'firing').length
}

// 查询
const handleQuery = () => {
  queryParams.page = 1
  if (viewMode.value === 'list') {
    loadData()
  } else {
    loadGroupData()
  }
}

// 重置
const handleReset = () => {
  queryParams.page = 1
  queryParams.page_size = 20
  queryParams.status = undefined
  queryParams.severity = undefined
  queryParams.alert_name = undefined
  handleQuery()
}

// 视图模式切换
const handleViewModeChange = () => {
  if (viewMode.value === 'list') {
    loadData()
  } else {
    loadGroupData()
  }
}

// 行点击
const handleRowClick = (row: Alert) => {
  router.push(`/alerts/${row.id}`)
}

// 查看详情
const handleViewDetail = (row: Alert) => {
  router.push(`/alerts/${row.id}`)
}

// 确认告警
const handleAck = (row: Alert) => {
  ackForm.id = row.id
  ackForm.comment = ''
  ackDialogVisible.value = true
}

const confirmAck = async () => {
  try {
    await ackAlert(ackForm.id, ackForm.comment)
    ElMessage.success('确认成功')
    ackDialogVisible.value = false
    loadData()
  } catch (error) {
    console.error('Failed to ack alert:', error)
  }
}

// 解决告警
const handleResolve = (row: Alert) => {
  resolveForm.id = row.id
  resolveForm.comment = ''
  resolveDialogVisible.value = true
}

const confirmResolve = async () => {
  try {
    await resolveAlert(resolveForm.id, resolveForm.comment)
    ElMessage.success('解决成功')
    resolveDialogVisible.value = false
    loadData()
  } catch (error) {
    console.error('Failed to resolve alert:', error)
  }
}

// 工具函数
const getSeverityType = (severity: string) => {
  const map: Record<string, any> = {
    critical: 'danger',
    warning: 'warning',
    info: 'info',
  }
  return map[severity] || 'info'
}

const getSeverityText = (severity: string) => {
  const map: Record<string, string> = {
    critical: '严重',
    warning: '警告',
    info: '信息',
  }
  return map[severity] || severity
}

const getStatusType = (status: string) => {
  return status === 'firing' ? 'danger' : 'success'
}

const getStatusText = (status: string) => {
  const map: Record<string, string> = {
    firing: '触发中',
    resolved: '已解决',
  }
  return map[status] || status
}

const getMainLabels = (labels: Record<string, string>) => {
  const mainKeys = ['cluster', 'namespace', 'service', 'instance']
  const result: Record<string, string> = {}
  mainKeys.forEach((key) => {
    if (labels[key]) {
      result[key] = labels[key]
    }
  })
  return result
}

const formatTime = (time: string) => {
  return dayjs(time).format('YYYY-MM-DD HH:mm:ss')
}

// 初始化
onMounted(() => {
  loadData()
})
</script>

<style scoped lang="scss">
.alert-list-container {
  .stat-card {
    background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
    border-radius: 12px;
    padding: 24px;
    color: #fff;
    box-shadow: 0 4px 20px rgba(102, 126, 234, 0.3);
    transition: transform 0.3s;

    &:hover {
      transform: translateY(-4px);
    }

    &.critical {
      background: linear-gradient(135deg, #f56c6c 0%, #c0392b 100%);
      box-shadow: 0 4px 20px rgba(245, 108, 108, 0.3);
    }

    &.warning {
      background: linear-gradient(135deg, #e6a23c 0%, #d68910 100%);
      box-shadow: 0 4px 20px rgba(230, 162, 60, 0.3);
    }

    &.info {
      background: linear-gradient(135deg, #409eff 0%, #2c7bd9 100%);
      box-shadow: 0 4px 20px rgba(64, 158, 255, 0.3);
    }

    .stat-value {
      font-size: 36px;
      font-weight: 700;
      margin-bottom: 8px;
    }

    .stat-label {
      font-size: 14px;
      opacity: 0.9;
    }
  }

  :deep(.el-table) {
    .el-table__row {
      cursor: pointer;

      &:hover {
        background-color: #f5f7fa;
      }
    }
  }
}
</style>
