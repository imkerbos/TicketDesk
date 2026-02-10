<template>
  <div class="reports-container">
    <!-- 页面头部 -->
    <div class="page-header">
      <div class="header-info">
        <div class="header-icon">
          <el-icon><DataAnalysis /></el-icon>
        </div>
        <div class="header-text">
          <h1 class="header-title">报表统计</h1>
          <p class="header-desc">查看工单和告警的统计数据</p>
        </div>
      </div>
      <div class="header-actions">
        <el-date-picker
          v-model="dateRange"
          type="daterange"
          range-separator="至"
          start-placeholder="开始日期"
          end-placeholder="结束日期"
          format="YYYY-MM-DD"
          value-format="YYYY-MM-DD"
          :shortcuts="dateShortcuts"
          @change="handleDateChange"
        />
        <el-select v-model="selectedProject" placeholder="全部项目" clearable style="width: 160px" @change="loadData">
          <el-option v-for="p in projects" :key="p.project_key" :label="p.name" :value="p.project_key" />
        </el-select>
      </div>
    </div>

    <!-- Tab 切换 -->
    <el-tabs v-model="activeTab" class="report-tabs" @tab-change="handleTabChange">
      <el-tab-pane label="工单统计" name="issues">
        <div v-loading="loading.issues" class="tab-content">
          <!-- 汇总卡片 -->
          <el-row :gutter="20" class="summary-row">
            <el-col :xs="12" :sm="6">
              <div class="summary-card">
                <div class="summary-value">{{ issueStats.summary?.total || 0 }}</div>
                <div class="summary-label">工单总数</div>
              </div>
            </el-col>
            <el-col :xs="12" :sm="6">
              <div class="summary-card success">
                <div class="summary-value">{{ issueStats.summary?.resolved || 0 }}</div>
                <div class="summary-label">已完成</div>
              </div>
            </el-col>
            <el-col :xs="12" :sm="6">
              <div class="summary-card warning">
                <div class="summary-value">{{ issueStats.summary?.in_progress || 0 }}</div>
                <div class="summary-label">进行中</div>
              </div>
            </el-col>
            <el-col :xs="12" :sm="6">
              <div class="summary-card info">
                <div class="summary-value">{{ formatHours(issueStats.summary?.avg_resolve_time || 0) }}</div>
                <div class="summary-label">平均解决时间</div>
              </div>
            </el-col>
          </el-row>

          <!-- 图表 -->
          <el-row :gutter="20">
            <el-col :xs="24" :lg="12">
              <el-card shadow="never" class="chart-card">
                <template #header>
                  <span class="card-title">状态分布</span>
                </template>
                <div class="distribution-list">
                  <div v-for="item in issueStats.status_distribution" :key="item.name" class="distribution-item">
                    <div class="dist-label">
                      <span class="dist-name">{{ getStatusLabel(item.name) }}</span>
                      <span class="dist-value">{{ item.value }}</span>
                    </div>
                    <el-progress :percentage="item.ratio" :show-text="false" :stroke-width="8" :color="getStatusColor(item.name)" />
                  </div>
                </div>
              </el-card>
            </el-col>
            <el-col :xs="24" :lg="12">
              <el-card shadow="never" class="chart-card">
                <template #header>
                  <span class="card-title">优先级分布</span>
                </template>
                <div class="distribution-list">
                  <div v-for="item in issueStats.priority_distribution" :key="item.name" class="distribution-item">
                    <div class="dist-label">
                      <span class="dist-name">{{ item.name }}</span>
                      <span class="dist-value">{{ item.value }}</span>
                    </div>
                    <el-progress :percentage="item.ratio" :show-text="false" :stroke-width="8" :color="getPriorityColor(item.name)" />
                  </div>
                </div>
              </el-card>
            </el-col>
          </el-row>

          <!-- 时间趋势 -->
          <el-card shadow="never" class="chart-card">
            <template #header>
              <span class="card-title">工单趋势</span>
            </template>
            <div class="timeline-list">
              <el-table :data="issueStats.timeline || []" stripe>
                <el-table-column prop="date" label="日期" width="120" />
                <el-table-column prop="created" label="创建" width="100" />
                <el-table-column prop="resolved" label="解决" width="100" />
                <el-table-column prop="closed" label="关闭" width="100" />
              </el-table>
            </div>
          </el-card>
        </div>
      </el-tab-pane>

      <el-tab-pane label="SLA 报表" name="sla">
        <div v-loading="loading.sla" class="tab-content">
          <!-- SLA 汇总 -->
          <el-row :gutter="20" class="summary-row">
            <el-col :xs="12" :sm="6">
              <div class="summary-card">
                <div class="summary-value">{{ slaReport.summary?.total_issues || 0 }}</div>
                <div class="summary-label">工单总数</div>
              </div>
            </el-col>
            <el-col :xs="12" :sm="6">
              <div class="summary-card success">
                <div class="summary-value">{{ slaReport.summary?.sla_met || 0 }}</div>
                <div class="summary-label">SLA 达标</div>
              </div>
            </el-col>
            <el-col :xs="12" :sm="6">
              <div class="summary-card danger">
                <div class="summary-value">{{ slaReport.summary?.sla_violated || 0 }}</div>
                <div class="summary-label">SLA 违规</div>
              </div>
            </el-col>
            <el-col :xs="12" :sm="6">
              <div class="summary-card" :class="getSLARateClass(slaReport.summary?.sla_rate)">
                <div class="summary-value">{{ formatPercent(slaReport.summary?.sla_rate || 0) }}</div>
                <div class="summary-label">达标率</div>
              </div>
            </el-col>
          </el-row>

          <!-- 按优先级 SLA -->
          <el-card shadow="never" class="chart-card">
            <template #header>
              <span class="card-title">按优先级 SLA 统计</span>
            </template>
            <el-table :data="slaReport.by_priority || []" stripe>
              <el-table-column prop="priority" label="优先级" width="100">
                <template #default="{ row }">
                  <el-tag :type="getPriorityType(row.priority)" effect="dark">{{ row.priority }}</el-tag>
                </template>
              </el-table-column>
              <el-table-column prop="total" label="总数" width="80" />
              <el-table-column prop="resolved" label="已完成" width="80" />
              <el-table-column prop="sla_target" label="SLA 目标" width="100">
                <template #default="{ row }">
                  {{ formatMinutes(row.sla_target) }}
                </template>
              </el-table-column>
              <el-table-column prop="mttr" label="平均解决" width="100">
                <template #default="{ row }">
                  {{ formatMinutes(row.mttr) }}
                </template>
              </el-table-column>
              <el-table-column prop="sla_met" label="达标数" width="80" />
              <el-table-column prop="sla_rate" label="达标率" width="100">
                <template #default="{ row }">
                  <span :class="getSLARateClass(row.sla_rate)">{{ formatPercent(row.sla_rate) }}</span>
                </template>
              </el-table-column>
            </el-table>
          </el-card>
        </div>
      </el-tab-pane>

      <el-tab-pane label="告警统计" name="alerts">
        <div v-loading="loading.alerts" class="tab-content">
          <!-- 告警汇总 -->
          <el-row :gutter="20" class="summary-row">
            <el-col :xs="12" :sm="6">
              <div class="summary-card">
                <div class="summary-value">{{ alertStats.summary?.total || 0 }}</div>
                <div class="summary-label">告警总数</div>
              </div>
            </el-col>
            <el-col :xs="12" :sm="6">
              <div class="summary-card danger">
                <div class="summary-value">{{ alertStats.summary?.firing || 0 }}</div>
                <div class="summary-label">触发中</div>
              </div>
            </el-col>
            <el-col :xs="12" :sm="6">
              <div class="summary-card success">
                <div class="summary-value">{{ alertStats.summary?.resolved || 0 }}</div>
                <div class="summary-label">已恢复</div>
              </div>
            </el-col>
            <el-col :xs="12" :sm="6">
              <div class="summary-card info">
                <div class="summary-value">{{ formatMinutes(alertStats.summary?.avg_ack_time || 0) }}</div>
                <div class="summary-label">平均确认时间</div>
              </div>
            </el-col>
          </el-row>

          <el-row :gutter="20">
            <!-- 严重程度分布 -->
            <el-col :xs="24" :lg="12">
              <el-card shadow="never" class="chart-card">
                <template #header>
                  <span class="card-title">严重程度分布</span>
                </template>
                <div class="distribution-list">
                  <div v-for="item in alertStats.severity_distribution" :key="item.name" class="distribution-item">
                    <div class="dist-label">
                      <span class="dist-name">{{ getSeverityLabel(item.name) }}</span>
                      <span class="dist-value">{{ item.value }}</span>
                    </div>
                    <el-progress :percentage="item.ratio" :show-text="false" :stroke-width="8" :color="getSeverityColor(item.name)" />
                  </div>
                </div>
              </el-card>
            </el-col>

            <!-- Top 告警 -->
            <el-col :xs="24" :lg="12">
              <el-card shadow="never" class="chart-card">
                <template #header>
                  <span class="card-title">Top 10 告警</span>
                </template>
                <div class="top-list">
                  <div v-for="(item, index) in alertStats.top_alerts" :key="item.alert_name" class="top-item">
                    <span class="top-rank">{{ index + 1 }}</span>
                    <span class="top-name">{{ item.alert_name }}</span>
                    <span class="top-count">{{ item.count }}</span>
                  </div>
                </div>
              </el-card>
            </el-col>
          </el-row>
        </div>
      </el-tab-pane>

      <el-tab-pane label="用户绩效" name="performance">
        <div v-loading="loading.performance" class="tab-content">
          <el-card shadow="never" class="chart-card">
            <template #header>
              <span class="card-title">用户处理工单绩效</span>
            </template>
            <el-table :data="userPerformance" stripe>
              <el-table-column prop="display_name" label="用户" width="150">
                <template #default="{ row }">
                  {{ row.display_name || row.username }}
                </template>
              </el-table-column>
              <el-table-column prop="assigned" label="指派工单" width="120" />
              <el-table-column prop="resolved" label="解决工单" width="120" />
              <el-table-column label="解决率" width="120">
                <template #default="{ row }">
                  {{ row.assigned > 0 ? formatPercent(row.resolved / row.assigned * 100) : '-' }}
                </template>
              </el-table-column>
              <el-table-column prop="avg_resolve_time" label="平均解决时间" width="120">
                <template #default="{ row }">
                  {{ formatHours(row.avg_resolve_time) }}
                </template>
              </el-table-column>
            </el-table>
          </el-card>
        </div>
      </el-tab-pane>
    </el-tabs>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { DataAnalysis } from '@element-plus/icons-vue'
import { getIssueStats, getSLAReport, getAlertStats, getUserPerformance } from '@/api/report'
import { getAllProjects } from '@/api/project'
import type { IssueStats, SLAReport, AlertStats, UserPerformance } from '@/types/report'
import type { Project } from '@/types/project'
import dayjs from 'dayjs'

// 日期范围
const dateRange = ref<[string, string]>([
  dayjs().subtract(30, 'day').format('YYYY-MM-DD'),
  dayjs().format('YYYY-MM-DD'),
])

const dateShortcuts = [
  { text: '最近一周', value: () => [dayjs().subtract(7, 'day').toDate(), dayjs().toDate()] },
  { text: '最近一月', value: () => [dayjs().subtract(30, 'day').toDate(), dayjs().toDate()] },
  { text: '最近三月', value: () => [dayjs().subtract(90, 'day').toDate(), dayjs().toDate()] },
]

const selectedProject = ref('')
const projects = ref<Project[]>([])

const activeTab = ref('issues')
const loading = reactive({
  issues: false,
  sla: false,
  alerts: false,
  performance: false,
})

const issueStats = ref<Partial<IssueStats>>({})
const slaReport = ref<Partial<SLAReport>>({})
const alertStats = ref<Partial<AlertStats>>({})
const userPerformance = ref<UserPerformance[]>([])

// 加载项目列表
const loadProjects = async () => {
  try {
    const { data } = await getAllProjects()
    projects.value = data.data
  } catch (error) {
    console.error('Failed to load projects:', error)
  }
}

// 加载工单统计
const loadIssueStats = async () => {
  loading.issues = true
  try {
    const { data } = await getIssueStats({
      project_key: selectedProject.value || undefined,
      start_date: dateRange.value[0],
      end_date: dateRange.value[1],
    })
    issueStats.value = data.data
  } catch (error) {
    console.error('Failed to load issue stats:', error)
  } finally {
    loading.issues = false
  }
}

// 加载 SLA 报表
const loadSLAReport = async () => {
  loading.sla = true
  try {
    const { data } = await getSLAReport({
      project_key: selectedProject.value || undefined,
      start_date: dateRange.value[0],
      end_date: dateRange.value[1],
    })
    slaReport.value = data.data
  } catch (error) {
    console.error('Failed to load SLA report:', error)
  } finally {
    loading.sla = false
  }
}

// 加载告警统计
const loadAlertStats = async () => {
  loading.alerts = true
  try {
    const { data } = await getAlertStats({
      project_key: selectedProject.value || undefined,
      start_date: dateRange.value[0],
      end_date: dateRange.value[1],
    })
    alertStats.value = data.data
  } catch (error) {
    console.error('Failed to load alert stats:', error)
  } finally {
    loading.alerts = false
  }
}

// 加载用户绩效
const loadUserPerformance = async () => {
  loading.performance = true
  try {
    const { data } = await getUserPerformance({
      project_key: selectedProject.value || undefined,
      start_date: dateRange.value[0],
      end_date: dateRange.value[1],
    })
    userPerformance.value = data.data
  } catch (error) {
    console.error('Failed to load user performance:', error)
  } finally {
    loading.performance = false
  }
}

// 根据当前 tab 加载数据
const loadData = () => {
  switch (activeTab.value) {
    case 'issues':
      loadIssueStats()
      break
    case 'sla':
      loadSLAReport()
      break
    case 'alerts':
      loadAlertStats()
      break
    case 'performance':
      loadUserPerformance()
      break
  }
}

const handleDateChange = () => {
  loadData()
}

const handleTabChange = () => {
  loadData()
}

// 格式化函数
const formatHours = (hours: number) => {
  if (hours < 1) return `${Math.round(hours * 60)}分钟`
  if (hours < 24) return `${hours.toFixed(1)}小时`
  return `${(hours / 24).toFixed(1)}天`
}

const formatMinutes = (minutes: number) => {
  if (minutes < 60) return `${Math.round(minutes)}分钟`
  if (minutes < 1440) return `${(minutes / 60).toFixed(1)}小时`
  return `${(minutes / 1440).toFixed(1)}天`
}

const formatPercent = (value: number) => `${value.toFixed(1)}%`

// 状态相关
type TagType = 'primary' | 'success' | 'warning' | 'info' | 'danger'

const getStatusLabel = (status: string) => {
  const map: Record<string, string> = { open: '待处理', in_progress: '进行中', resolved: '已完成', closed: '已终止' }
  return map[status] || status
}

const getStatusColor = (status: string) => {
  const map: Record<string, string> = { open: '#909399', in_progress: '#E6A23C', resolved: '#67C23A', closed: '#409EFF', merged: '#8b5cf6' }
  return map[status] || '#909399'
}

const getPriorityColor = (priority: string) => {
  const map: Record<string, string> = { P0: '#F56C6C', P1: '#E6A23C', P2: '#409EFF', P3: '#909399' }
  return map[priority] || '#909399'
}

const getPriorityType = (priority: string): TagType => {
  const map: Record<string, TagType> = { P0: 'danger', P1: 'warning', P2: 'primary', P3: 'info' }
  return map[priority] || 'info'
}

const getSeverityLabel = (severity: string) => {
  const map: Record<string, string> = { critical: '严重', warning: '警告', info: '信息' }
  return map[severity] || severity
}

const getSeverityColor = (severity: string) => {
  const map: Record<string, string> = { critical: '#F56C6C', warning: '#E6A23C', info: '#909399' }
  return map[severity] || '#909399'
}

const getSLARateClass = (rate?: number) => {
  if (!rate) return ''
  if (rate >= 90) return 'success'
  if (rate >= 70) return 'warning'
  return 'danger'
}

// 初始化
onMounted(() => {
  loadProjects()
  loadData()
})
</script>

<style scoped lang="scss">
.reports-container {
  width: 100%;
}

.page-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 24px;
  padding: 24px;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  border-radius: 12px;
  color: #fff;

  .header-info {
    display: flex;
    align-items: center;
    gap: 16px;
  }

  .header-icon {
    width: 48px;
    height: 48px;
    background: rgba(255, 255, 255, 0.2);
    border-radius: 12px;
    display: flex;
    align-items: center;
    justify-content: center;
    font-size: 24px;
  }

  .header-title {
    font-size: 20px;
    font-weight: 700;
    margin: 0 0 4px;
  }

  .header-desc {
    font-size: 14px;
    margin: 0;
    opacity: 0.9;
  }

  .header-actions {
    display: flex;
    gap: 12px;
  }
}

.report-tabs {
  :deep(.el-tabs__header) {
    margin-bottom: 20px;
  }
}

.tab-content {
  min-height: 400px;
}

.summary-row {
  margin-bottom: 20px;
}

.summary-card {
  background: #fff;
  border-radius: 12px;
  padding: 20px;
  text-align: center;
  box-shadow: 0 2px 12px rgba(0, 0, 0, 0.05);
  transition: transform 0.2s;

  &:hover {
    transform: translateY(-2px);
  }

  .summary-value {
    font-size: 28px;
    font-weight: 700;
    color: #1f2937;
    line-height: 1.2;
  }

  .summary-label {
    font-size: 13px;
    color: #6b7280;
    margin-top: 8px;
  }

  &.success .summary-value { color: #67C23A; }
  &.warning .summary-value { color: #E6A23C; }
  &.danger .summary-value { color: #F56C6C; }
  &.info .summary-value { color: #409EFF; }
}

.chart-card {
  margin-bottom: 20px;
  border-radius: 12px;

  :deep(.el-card__header) {
    padding: 16px 20px;
    border-bottom: 1px solid #f0f0f0;
  }

  .card-title {
    font-size: 15px;
    font-weight: 600;
    color: #1f2937;
  }
}

.distribution-list {
  padding: 16px 0;

  .distribution-item {
    margin-bottom: 16px;

    &:last-child {
      margin-bottom: 0;
    }

    .dist-label {
      display: flex;
      justify-content: space-between;
      margin-bottom: 8px;

      .dist-name {
        font-size: 14px;
        color: #4b5563;
      }

      .dist-value {
        font-size: 14px;
        font-weight: 600;
        color: #1f2937;
      }
    }
  }
}

.top-list {
  padding: 8px 0;

  .top-item {
    display: flex;
    align-items: center;
    padding: 12px 0;
    border-bottom: 1px solid #f0f0f0;

    &:last-child {
      border-bottom: none;
    }

    .top-rank {
      width: 24px;
      height: 24px;
      background: #f3f4f6;
      border-radius: 6px;
      display: flex;
      align-items: center;
      justify-content: center;
      font-size: 12px;
      font-weight: 600;
      color: #6b7280;
      margin-right: 12px;
    }

    .top-name {
      flex: 1;
      font-size: 14px;
      color: #1f2937;
      overflow: hidden;
      text-overflow: ellipsis;
      white-space: nowrap;
    }

    .top-count {
      font-size: 14px;
      font-weight: 600;
      color: #667eea;
    }
  }
}

.timeline-list {
  max-height: 400px;
  overflow-y: auto;
}

@media (max-width: 768px) {
  .page-header {
    flex-direction: column;
    gap: 16px;
    padding: 16px;

    .header-actions {
      width: 100%;
      flex-direction: column;
    }
  }
}
</style>
