<template>
  <div class="reports-container">
    <!-- 页面头部 -->
    <TdPageHeader>
      <template #leading>
        <div class="page-header-icon">
          <el-icon :size="20"><DataAnalysis /></el-icon>
        </div>
      </template>
      <template #title>报表统计</template>
      <template #subtitle>查看工单和告警的统计数据</template>
      <template #actions>
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
        <el-select v-model="selectedProject" placeholder="全部项目" clearable style="width: 200px" filterable @change="handleProjectChange">
          <el-option v-for="p in projects" :key="p.project_key" :label="`${p.project_key} - ${p.name}`" :value="p.project_key" />
        </el-select>
      </template>
    </TdPageHeader>

    <!-- Tab 切换 -->
    <el-tabs v-model="activeTab" class="report-tabs" @tab-change="handleTabChange">
      <el-tab-pane label="工单统计" name="issues">
        <div v-loading="loading.issues" class="tab-content">
          <!-- 汇总卡片 -->
          <el-row :gutter="16" class="summary-row">
            <el-col :xs="12" :sm="6">
              <div class="summary-card accent-default">
                <div class="summary-icon-wrap default"><el-icon :size="20"><Tickets /></el-icon></div>
                <div class="summary-body">
                  <div class="summary-value">{{ issueStats.summary?.total || 0 }}</div>
                  <div class="summary-label">工单总数</div>
                </div>
              </div>
            </el-col>
            <el-col :xs="12" :sm="6">
              <div class="summary-card accent-success">
                <div class="summary-icon-wrap success"><el-icon :size="20"><CircleCheck /></el-icon></div>
                <div class="summary-body">
                  <div class="summary-value">{{ issueStats.summary?.resolved || 0 }}</div>
                  <div class="summary-label">已完成</div>
                </div>
              </div>
            </el-col>
            <el-col :xs="12" :sm="6">
              <div class="summary-card accent-warning">
                <div class="summary-icon-wrap warning"><el-icon :size="20"><Loading /></el-icon></div>
                <div class="summary-body">
                  <div class="summary-value">{{ issueStats.summary?.in_progress || 0 }}</div>
                  <div class="summary-label">进行中</div>
                </div>
              </div>
            </el-col>
            <el-col :xs="12" :sm="6">
              <div class="summary-card accent-info">
                <div class="summary-icon-wrap info"><el-icon :size="20"><Timer /></el-icon></div>
                <div class="summary-body">
                  <div class="summary-value">{{ formatHours(issueStats.summary?.avg_resolve_time || 0) }}</div>
                  <div class="summary-label">平均解决时间</div>
                </div>
              </div>
            </el-col>
          </el-row>

          <!-- 分布图表：第一行 状态 + 优先级 -->
          <el-row :gutter="16" class="stretch-row">
            <el-col :xs="24" :lg="12">
              <el-card shadow="never" class="chart-card">
                <template #header>
                  <div class="card-header">
                    <span class="card-dot" style="background: #67C23A"></span>
                    <span class="card-title">状态分布</span>
                  </div>
                </template>
                <div class="distribution-list">
                  <div v-for="item in issueStats.status_distribution" :key="item.name" class="distribution-item">
                    <div class="dist-header">
                      <span class="dist-dot" :style="{ background: getStatusColor(item.name) }"></span>
                      <span class="dist-name">{{ getStatusLabel(item.name) }}</span>
                      <span class="dist-value">{{ item.value }}<span class="dist-ratio">{{ item.ratio.toFixed(1) }}%</span></span>
                    </div>
                    <el-progress :percentage="item.ratio" :show-text="false" :stroke-width="6" :color="getStatusColor(item.name)" />
                  </div>
                  <el-empty v-if="!issueStats.status_distribution?.length" description="暂无数据" :image-size="60" />
                </div>
              </el-card>
            </el-col>
            <el-col :xs="24" :lg="12">
              <el-card shadow="never" class="chart-card">
                <template #header>
                  <div class="card-header">
                    <span class="card-dot" style="background: #409EFF"></span>
                    <span class="card-title">优先级分布</span>
                  </div>
                </template>
                <div class="distribution-list">
                  <div v-for="item in issueStats.priority_distribution" :key="item.name" class="distribution-item">
                    <div class="dist-header">
                      <span class="dist-dot" :style="{ background: getPriorityColor(item.name) }"></span>
                      <span class="dist-name">{{ item.name }}</span>
                      <span class="dist-value">{{ item.value }}<span class="dist-ratio">{{ item.ratio.toFixed(1) }}%</span></span>
                    </div>
                    <el-progress :percentage="item.ratio" :show-text="false" :stroke-width="6" :color="getPriorityColor(item.name)" />
                  </div>
                  <el-empty v-if="!issueStats.priority_distribution?.length" description="暂无数据" :image-size="60" />
                </div>
              </el-card>
            </el-col>
          </el-row>

          <!-- 分布图表：第二行 工单类型 + 指派人 -->
          <el-row :gutter="16" class="stretch-row">
            <el-col :xs="24" :lg="12">
              <el-card shadow="never" class="chart-card">
                <template #header>
                  <div class="card-header">
                    <span class="card-dot" style="background: var(--td-color-primary)"></span>
                    <span class="card-title">工单类型分布</span>
                  </div>
                </template>
                <div class="distribution-list">
                  <div v-for="(item, idx) in issueStats.type_distribution" :key="item.name" class="distribution-item">
                    <div class="dist-header">
                      <span class="dist-dot" :style="{ background: typeColors[idx % typeColors.length] }"></span>
                      <span class="dist-name">{{ item.name }}</span>
                      <span class="dist-value">{{ item.value }}<span class="dist-ratio">{{ item.ratio.toFixed(1) }}%</span></span>
                    </div>
                    <el-progress :percentage="item.ratio" :show-text="false" :stroke-width="6" :color="typeColors[idx % typeColors.length]" />
                  </div>
                  <el-empty v-if="!issueStats.type_distribution?.length" description="暂无数据" :image-size="60" />
                </div>
              </el-card>
            </el-col>
            <el-col :xs="24" :lg="12">
              <el-card shadow="never" class="chart-card">
                <template #header>
                  <div class="card-header">
                    <span class="card-dot" style="background: #8b5cf6"></span>
                    <span class="card-title">指派人分布</span>
                  </div>
                </template>
                <div class="distribution-list">
                  <div v-for="(item, idx) in issueStats.assignee_distribution" :key="item.name" class="distribution-item">
                    <div class="dist-header">
                      <el-avatar :size="22" class="dist-avatar">{{ item.name?.charAt(0) }}</el-avatar>
                      <span class="dist-name">{{ item.name }}</span>
                      <span class="dist-value">{{ item.value }}<span class="dist-ratio">{{ item.ratio.toFixed(1) }}%</span></span>
                    </div>
                    <el-progress :percentage="item.ratio" :show-text="false" :stroke-width="6" :color="assigneeColors[idx % assigneeColors.length]" />
                  </div>
                  <el-empty v-if="!issueStats.assignee_distribution?.length" description="暂无数据" :image-size="60" />
                </div>
              </el-card>
            </el-col>
          </el-row>

          <!-- 分布图表：第三行 Epic -->
          <el-row :gutter="16">
            <el-col :span="24">
              <el-card shadow="never" class="chart-card">
                <template #header>
                  <div class="card-header">
                    <span class="card-dot" style="background: var(--td-color-warning)"></span>
                    <span class="card-title">Epic 分布</span>
                  </div>
                </template>
                <div class="distribution-list distribution-list--horizontal">
                  <div v-for="(item, idx) in issueStats.epic_distribution" :key="item.name" class="distribution-item">
                    <div class="dist-header">
                      <span class="dist-dot" :style="{ background: epicColors[idx % epicColors.length] }"></span>
                      <span class="dist-name">{{ item.name }}</span>
                      <span class="dist-value">{{ item.value }}<span class="dist-ratio">{{ item.ratio.toFixed(1) }}%</span></span>
                    </div>
                    <el-progress :percentage="item.ratio" :show-text="false" :stroke-width="6" :color="epicColors[idx % epicColors.length]" />
                  </div>
                  <el-empty v-if="!issueStats.epic_distribution?.length" description="暂无数据" :image-size="60" />
                </div>
              </el-card>
            </el-col>
          </el-row>

          <!-- 时间趋势 -->
          <el-card shadow="never" class="chart-card">
            <template #header>
              <div class="card-header">
                <span class="card-dot" style="background: #6366f1"></span>
                <span class="card-title">工单趋势</span>
              </div>
            </template>
            <div class="timeline-list">
              <el-table :data="issueStats.timeline || []" stripe size="small" :header-cell-style="{ background: 'var(--td-table-header-bg)', color: 'var(--td-text-regular)', fontWeight: 600 }">
                <el-table-column label="日期" min-width="120">
                  <template #default="{ row }">
                    <span class="timeline-date">{{ formatDate(row.date) }}</span>
                  </template>
                </el-table-column>
                <el-table-column label="创建" min-width="100">
                  <template #default="{ row }">
                    <span class="timeline-badge created">{{ row.created }}</span>
                  </template>
                </el-table-column>
                <el-table-column label="进行中" min-width="100">
                  <template #default="{ row }">
                    <span class="timeline-badge in-progress">{{ row.in_progress }}</span>
                  </template>
                </el-table-column>
                <el-table-column label="解决" min-width="100">
                  <template #default="{ row }">
                    <span class="timeline-badge resolved">{{ row.resolved }}</span>
                  </template>
                </el-table-column>
                <el-table-column label="关闭" min-width="100">
                  <template #default="{ row }">
                    <span class="timeline-badge closed">{{ row.closed }}</span>
                  </template>
                </el-table-column>
              </el-table>
              <el-empty v-if="!issueStats.timeline?.length" description="暂无趋势数据" :image-size="60" />
            </div>
          </el-card>
        </div>
      </el-tab-pane>

      <el-tab-pane label="SLA 报表" name="sla">
        <div v-loading="loading.sla" class="tab-content">
          <!-- SLA 汇总卡片 -->
          <el-row :gutter="16" class="summary-row">
            <el-col :xs="12" :sm="6">
              <div class="summary-card accent-default">
                <div class="summary-icon-wrap default"><el-icon :size="20"><Tickets /></el-icon></div>
                <div class="summary-body">
                  <div class="summary-value">{{ slaReport.summary?.total_issues || 0 }}</div>
                  <div class="summary-label">工单总数 / 已解决 {{ slaReport.summary?.resolved_issues || 0 }}</div>
                </div>
              </div>
            </el-col>
            <el-col :xs="12" :sm="6">
              <div class="summary-card accent-success">
                <div class="summary-icon-wrap success"><el-icon :size="20"><CircleCheck /></el-icon></div>
                <div class="summary-body">
                  <div class="summary-value">{{ formatPercent(slaReport.summary?.sla_rate || 0) }}</div>
                  <div class="summary-label">达标率（达标 {{ slaReport.summary?.sla_met || 0 }} / 违规 {{ slaReport.summary?.sla_violated || 0 }}）</div>
                </div>
              </div>
            </el-col>
            <el-col :xs="12" :sm="6">
              <div class="summary-card accent-info">
                <div class="summary-icon-wrap info"><el-icon :size="20"><Timer /></el-icon></div>
                <div class="summary-body">
                  <div class="summary-value">{{ formatMinutes(slaReport.summary?.mtta || 0) }}</div>
                  <div class="summary-label">MTTA 平均确认时间</div>
                </div>
              </div>
            </el-col>
            <el-col :xs="12" :sm="6">
              <div class="summary-card accent-warning">
                <div class="summary-icon-wrap warning"><el-icon :size="20"><Clock /></el-icon></div>
                <div class="summary-body">
                  <div class="summary-value">{{ formatMinutes(slaReport.summary?.mttr || 0) }}</div>
                  <div class="summary-label">MTTR 平均解决时间</div>
                </div>
              </div>
            </el-col>
          </el-row>

          <!-- 按优先级 + 按项目 SLA（等高两列） -->
          <el-row :gutter="16" class="stretch-row">
            <el-col :xs="24" :lg="12">
              <el-card shadow="never" class="chart-card">
                <template #header>
                  <div class="card-header">
                    <span class="card-dot" style="background: #409EFF"></span>
                    <span class="card-title">按优先级 SLA 统计</span>
                  </div>
                </template>
                <el-table :data="slaReport.by_priority || []" stripe size="small" :header-cell-style="{ background: 'var(--td-table-header-bg)', color: 'var(--td-text-regular)', fontWeight: 600 }">
                  <el-table-column prop="priority" label="优先级" min-width="70">
                    <template #default="{ row }">
                      <el-tag :type="getPriorityType(row.priority)" effect="dark" size="small">{{ row.priority }}</el-tag>
                    </template>
                  </el-table-column>
                  <el-table-column prop="total" label="总数" min-width="55" />
                  <el-table-column prop="resolved" label="已解决" min-width="55" />
                  <el-table-column label="SLA 目标" min-width="80">
                    <template #default="{ row }">{{ formatMinutes(row.sla_target) }}</template>
                  </el-table-column>
                  <el-table-column label="MTTR" min-width="80">
                    <template #default="{ row }">{{ formatMinutes(row.mttr) }}</template>
                  </el-table-column>
                  <el-table-column label="达标率" min-width="75">
                    <template #default="{ row }">
                      <span :class="getSLARateClass(row.sla_rate)">{{ formatPercent(row.sla_rate) }}</span>
                    </template>
                  </el-table-column>
                </el-table>
              </el-card>
            </el-col>
            <el-col :xs="24" :lg="12">
              <el-card shadow="never" class="chart-card">
                <template #header>
                  <div class="card-header">
                    <span class="card-dot" style="background: #8b5cf6"></span>
                    <span class="card-title">按项目 SLA 统计</span>
                  </div>
                </template>
                <el-table :data="slaReport.by_project || []" stripe size="small" :header-cell-style="{ background: 'var(--td-table-header-bg)', color: 'var(--td-text-regular)', fontWeight: 600 }">
                  <el-table-column label="项目" min-width="140">
                    <template #default="{ row }">
                      <span class="project-key-badge">{{ row.project_key }}</span>
                      <span class="project-name-text">{{ row.project_name }}</span>
                    </template>
                  </el-table-column>
                  <el-table-column prop="total" label="总数" width="55" />
                  <el-table-column prop="resolved" label="已解决" width="55" />
                  <el-table-column label="MTTR" width="80">
                    <template #default="{ row }">{{ formatMinutes(row.mttr) }}</template>
                  </el-table-column>
                  <el-table-column label="达标率" width="75">
                    <template #default="{ row }">
                      <span :class="getSLARateClass(row.sla_rate)">{{ formatPercent(row.sla_rate) }}</span>
                    </template>
                  </el-table-column>
                </el-table>
                <el-empty v-if="!slaReport.by_project?.length" description="暂无数据" :image-size="60" />
              </el-card>
            </el-col>
          </el-row>

          <!-- SLA 超时工单列表 -->
          <el-card shadow="never" class="chart-card">
            <template #header>
              <div class="card-header">
                <span class="card-dot" style="background: #ef4444"></span>
                <span class="card-title">SLA 超时工单</span>
                <span v-if="slaReport.violations?.length" class="card-count">{{ slaReport.violations.length }} 条</span>
              </div>
            </template>
            <el-table :data="slaReport.violations || []" stripe size="small" :header-cell-style="{ background: 'var(--td-table-header-bg)', color: 'var(--td-text-regular)', fontWeight: 600 }">
              <el-table-column prop="issue_key" label="工单号" width="120">
                <template #default="{ row }">
                  <router-link :to="`/issues/${row.issue_key}`" class="issue-link">{{ row.issue_key }}</router-link>
                </template>
              </el-table-column>
              <el-table-column prop="title" label="标题" min-width="200" show-overflow-tooltip />
              <el-table-column prop="priority" label="优先级" width="80">
                <template #default="{ row }">
                  <el-tag :type="getPriorityType(row.priority)" effect="dark" size="small">{{ row.priority }}</el-tag>
                </template>
              </el-table-column>
              <el-table-column prop="assignee_name" label="经办人" width="100" />
              <el-table-column label="SLA 目标" width="100">
                <template #default="{ row }">{{ formatMinutes(row.sla_target) }}</template>
              </el-table-column>
              <el-table-column label="实际耗时" width="100">
                <template #default="{ row }">
                  <span class="text-danger">{{ formatMinutes(row.actual_time) }}</span>
                </template>
              </el-table-column>
              <el-table-column label="超时" width="100">
                <template #default="{ row }">
                  <span class="text-danger">+{{ formatMinutes(row.overdue_by) }}</span>
                </template>
              </el-table-column>
            </el-table>
            <el-empty v-if="!slaReport.violations?.length" description="暂无超时工单" :image-size="60" />
          </el-card>
        </div>
      </el-tab-pane>

      <el-tab-pane label="告警统计" name="alerts">
        <div v-loading="loading.alerts" class="tab-content">
          <!-- 告警汇总卡片 -->
          <el-row :gutter="16" class="summary-row">
            <el-col :xs="12" :sm="6">
              <div class="summary-card accent-default">
                <div class="summary-icon-wrap default"><el-icon :size="20"><Warning /></el-icon></div>
                <div class="summary-body">
                  <div class="summary-value">{{ alertStats.summary?.total || 0 }}</div>
                  <div class="summary-label">告警总数</div>
                </div>
              </div>
            </el-col>
            <el-col :xs="12" :sm="6">
              <div class="summary-card accent-danger">
                <div class="summary-icon-wrap danger"><el-icon :size="20"><Warning /></el-icon></div>
                <div class="summary-body">
                  <div class="summary-value">{{ alertStats.summary?.firing || 0 }}</div>
                  <div class="summary-label">触发中</div>
                </div>
              </div>
            </el-col>
            <el-col :xs="12" :sm="6">
              <div class="summary-card accent-warning">
                <div class="summary-icon-wrap warning"><el-icon :size="20"><CircleCheck /></el-icon></div>
                <div class="summary-body">
                  <div class="summary-value">{{ alertStats.summary?.acked || 0 }}</div>
                  <div class="summary-label">已确认</div>
                </div>
              </div>
            </el-col>
            <el-col :xs="12" :sm="6">
              <div class="summary-card accent-success">
                <div class="summary-icon-wrap success"><el-icon :size="20"><CircleCheck /></el-icon></div>
                <div class="summary-body">
                  <div class="summary-value">{{ alertStats.summary?.resolved || 0 }}</div>
                  <div class="summary-label">已恢复</div>
                </div>
              </div>
            </el-col>
          </el-row>

          <!-- 严重程度分布 + 告警趋势（等高两列） -->
          <el-row :gutter="16" class="stretch-row">
            <el-col :xs="24" :lg="12">
              <el-card shadow="never" class="chart-card">
                <template #header>
                  <div class="card-header">
                    <span class="card-dot" style="background: #ef4444"></span>
                    <span class="card-title">严重程度分布</span>
                  </div>
                </template>
                <div class="distribution-list">
                  <div v-for="item in alertStats.severity_distribution" :key="item.name" class="distribution-item">
                    <div class="dist-header">
                      <span class="dist-dot" :style="{ background: getSeverityColor(item.name) }"></span>
                      <span class="dist-name">{{ getSeverityLabel(item.name) }}</span>
                      <span class="dist-value">{{ item.value }}<span class="dist-ratio">{{ item.ratio.toFixed(1) }}%</span></span>
                    </div>
                    <el-progress :percentage="item.ratio" :show-text="false" :stroke-width="6" :color="getSeverityColor(item.name)" />
                  </div>
                  <el-empty v-if="!alertStats.severity_distribution?.length" description="暂无数据" :image-size="60" />
                </div>
              </el-card>
            </el-col>
            <el-col :xs="24" :lg="12">
              <el-card shadow="never" class="chart-card">
                <template #header>
                  <div class="card-header">
                    <span class="card-dot" style="background: #6366f1"></span>
                    <span class="card-title">告警趋势</span>
                  </div>
                </template>
                <div class="timeline-list">
                  <el-table :data="alertStats.timeline || []" stripe size="small" :header-cell-style="{ background: 'var(--td-table-header-bg)', color: 'var(--td-text-regular)', fontWeight: 600 }">
                    <el-table-column label="日期" min-width="120">
                      <template #default="{ row }">
                        <span class="timeline-date">{{ formatDate(row.date) }}</span>
                      </template>
                    </el-table-column>
                    <el-table-column label="触发" min-width="100">
                      <template #default="{ row }">
                        <span class="timeline-badge created">{{ row.firing }}</span>
                      </template>
                    </el-table-column>
                    <el-table-column label="恢复" min-width="100">
                      <template #default="{ row }">
                        <span class="timeline-badge resolved">{{ row.resolved }}</span>
                      </template>
                    </el-table-column>
                  </el-table>
                  <el-empty v-if="!alertStats.timeline?.length" description="暂无趋势数据" :image-size="60" />
                </div>
              </el-card>
            </el-col>
          </el-row>

          <!-- Top 告警（全宽，含严重程度列） -->
          <el-card shadow="never" class="chart-card">
            <template #header>
              <div class="card-header">
                <span class="card-dot" style="background: #f59e0b"></span>
                <span class="card-title">Top 告警</span>
              </div>
            </template>
            <el-table :data="alertStats.top_alerts || []" stripe size="small" :header-cell-style="{ background: 'var(--td-table-header-bg)', color: 'var(--td-text-regular)', fontWeight: 600 }">
              <el-table-column label="排名" width="60" align="center">
                <template #default="{ $index }">
                  <span :class="['rank-badge', { 'top-3': $index < 3 }]">{{ $index + 1 }}</span>
                </template>
              </el-table-column>
              <el-table-column prop="alert_name" label="告警名称" min-width="200" show-overflow-tooltip />
              <el-table-column prop="severity" label="严重程度" width="100">
                <template #default="{ row }">
                  <el-tag :type="getSeverityTagType(row.severity)" size="small" effect="plain">{{ getSeverityLabel(row.severity) }}</el-tag>
                </template>
              </el-table-column>
              <el-table-column prop="count" label="次数" width="100" align="center">
                <template #default="{ row }">
                  <span class="alert-count-value">{{ row.count }}</span>
                </template>
              </el-table-column>
            </el-table>
            <el-empty v-if="!alertStats.top_alerts?.length" description="暂无数据" :image-size="60" />
          </el-card>
        </div>
      </el-tab-pane>

      <el-tab-pane label="用户绩效" name="performance">
        <div v-loading="loading.performance" class="tab-content">
          <!-- 绩效汇总卡片 -->
          <el-row :gutter="16" class="summary-row">
            <el-col :xs="12" :sm="6">
              <div class="summary-card accent-default">
                <div class="summary-icon-wrap default"><el-icon :size="20"><User /></el-icon></div>
                <div class="summary-body">
                  <div class="summary-value">{{ performanceSummary.userCount }}</div>
                  <div class="summary-label">参与人数</div>
                </div>
              </div>
            </el-col>
            <el-col :xs="12" :sm="6">
              <div class="summary-card accent-info">
                <div class="summary-icon-wrap info"><el-icon :size="20"><Tickets /></el-icon></div>
                <div class="summary-body">
                  <div class="summary-value">{{ performanceSummary.totalAssigned }}</div>
                  <div class="summary-label">总指派工单</div>
                </div>
              </div>
            </el-col>
            <el-col :xs="12" :sm="6">
              <div class="summary-card accent-success">
                <div class="summary-icon-wrap success"><el-icon :size="20"><CircleCheck /></el-icon></div>
                <div class="summary-body">
                  <div class="summary-value">{{ performanceSummary.totalResolved }}</div>
                  <div class="summary-label">总解决工单</div>
                </div>
              </div>
            </el-col>
            <el-col :xs="12" :sm="6">
              <div class="summary-card accent-warning">
                <div class="summary-icon-wrap warning"><el-icon :size="20"><Timer /></el-icon></div>
                <div class="summary-body">
                  <div class="summary-value">{{ formatPercent(performanceSummary.avgResolveRate) }}</div>
                  <div class="summary-label">平均解决率</div>
                </div>
              </div>
            </el-col>
          </el-row>

          <!-- 绩效表格 -->
          <el-card shadow="never" class="chart-card">
            <template #header>
              <div class="card-header">
                <span class="card-dot" style="background: #8b5cf6"></span>
                <span class="card-title">用户处理工单绩效</span>
              </div>
            </template>
            <el-table :data="userPerformance" stripe size="small" :header-cell-style="{ background: 'var(--td-table-header-bg)', color: 'var(--td-text-regular)', fontWeight: 600 }">
              <el-table-column label="排名" width="60" align="center">
                <template #default="{ $index }">
                  <span :class="['rank-badge', { 'top-3': $index < 3 }]">{{ $index + 1 }}</span>
                </template>
              </el-table-column>
              <el-table-column prop="display_name" label="用户" min-width="150">
                <template #default="{ row }">
                  <div class="user-cell">
                    <el-avatar :size="24" class="dist-avatar">{{ (row.display_name || row.username)?.charAt(0) }}</el-avatar>
                    <span>{{ row.display_name || row.username }}</span>
                  </div>
                </template>
              </el-table-column>
              <el-table-column prop="assigned" label="指派工单" width="100" align="center" />
              <el-table-column prop="resolved" label="解决工单" width="100" align="center" />
              <el-table-column label="解决率" width="120">
                <template #default="{ row }">
                  <el-progress
                    :percentage="row.assigned > 0 ? Math.round(row.resolved / row.assigned * 100) : 0"
                    :stroke-width="8"
                    :show-text="true"
                    :format="(p: number) => p + '%'"
                  />
                </template>
              </el-table-column>
              <el-table-column label="平均解决时间" width="120">
                <template #default="{ row }">{{ formatHours(row.avg_resolve_time) }}</template>
              </el-table-column>
            </el-table>
            <el-empty v-if="!userPerformance.length" description="暂无绩效数据" :image-size="60" />
          </el-card>
        </div>
      </el-tab-pane>

      <el-tab-pane label="工时统计" name="worklogs">
        <div v-loading="loading.worklogs" class="tab-content">
          <!-- 汇总卡片 -->
          <el-row :gutter="16" class="summary-row">
            <el-col :xs="12" :sm="6">
              <div class="summary-card accent-default">
                <div class="summary-icon-wrap default"><el-icon :size="20"><Clock /></el-icon></div>
                <div class="summary-body">
                  <div class="summary-value">{{ formatWorklogTime(worklogStats.summary?.total_time_sec || 0) }}</div>
                  <div class="summary-label">总工时</div>
                </div>
              </div>
            </el-col>
            <el-col :xs="12" :sm="6">
              <div class="summary-card accent-success">
                <div class="summary-icon-wrap success"><el-icon :size="20"><Tickets /></el-icon></div>
                <div class="summary-body">
                  <div class="summary-value">{{ worklogStats.summary?.total_entries || 0 }}</div>
                  <div class="summary-label">记录条数</div>
                </div>
              </div>
            </el-col>
            <el-col :xs="12" :sm="6">
              <div class="summary-card accent-warning">
                <div class="summary-icon-wrap warning"><el-icon :size="20"><User /></el-icon></div>
                <div class="summary-body">
                  <div class="summary-value">{{ worklogStats.summary?.active_users || 0 }}</div>
                  <div class="summary-label">参与人数</div>
                </div>
              </div>
            </el-col>
            <el-col :xs="12" :sm="6">
              <div class="summary-card accent-info">
                <div class="summary-icon-wrap info"><el-icon :size="20"><Timer /></el-icon></div>
                <div class="summary-body">
                  <div class="summary-value">{{ formatWorklogTime(worklogStats.summary?.avg_daily_time_sec || 0) }}</div>
                  <div class="summary-label">日均工时</div>
                </div>
              </div>
            </el-col>
          </el-row>

          <el-row :gutter="16" class="stretch-row">
            <!-- 每日工时柱状图 -->
            <el-col :xs="24" :lg="16">
              <el-card shadow="never" class="chart-card">
                <template #header>
                  <div class="card-header">
                    <span class="card-dot" style="background: #6366f1"></span>
                    <span class="card-title">每日工时</span>
                  </div>
                </template>
                <div class="daily-worklog-chart">
                  <div v-if="worklogStats.daily_stats?.length" class="bar-chart">
                    <div v-for="item in worklogStats.daily_stats" :key="item.date" class="bar-item">
                      <div class="bar-value">{{ formatWorklogTime(item.total_time_sec) }}</div>
                      <div class="bar-fill" :style="{ height: getDailyBarHeight(item.total_time_sec) + '%' }"></div>
                      <div class="bar-label">{{ formatShortDate(item.date) }}</div>
                    </div>
                  </div>
                  <el-empty v-else description="暂无工时数据" :image-size="60" />
                </div>
              </el-card>
            </el-col>

            <!-- 工作类型分布 -->
            <el-col :xs="24" :lg="8">
              <el-card shadow="never" class="chart-card">
                <template #header>
                  <div class="card-header">
                    <span class="card-dot" style="background: var(--td-color-warning)"></span>
                    <span class="card-title">工作类型分布</span>
                  </div>
                </template>
                <div class="distribution-list">
                  <div v-for="(item, idx) in worklogStats.type_stats" :key="item.work_type" class="distribution-item">
                    <div class="dist-header">
                      <span class="dist-dot" :style="{ background: worklogTypeColors[idx % worklogTypeColors.length] }"></span>
                      <span class="dist-name">{{ item.work_type }}</span>
                      <span class="dist-value">{{ formatWorklogTime(item.total_time_sec) }}<span class="dist-ratio">{{ item.entry_count }}条</span></span>
                    </div>
                    <el-progress :percentage="getTypePercentage(item.total_time_sec)" :show-text="false" :stroke-width="6" :color="worklogTypeColors[idx % worklogTypeColors.length]" />
                  </div>
                  <el-empty v-if="!worklogStats.type_stats?.length" description="暂无数据" :image-size="60" />
                </div>
              </el-card>
            </el-col>
          </el-row>

          <!-- 个人工时排行 -->
          <el-card shadow="never" class="chart-card">
            <template #header>
              <div class="card-header">
                <span class="card-dot" style="background: #8b5cf6"></span>
                <span class="card-title">个人工时排行</span>
              </div>
            </template>
            <el-table :data="worklogStats.user_stats || []" stripe size="small" :header-cell-style="{ background: 'var(--td-table-header-bg)', color: 'var(--td-text-regular)', fontWeight: 600 }">
              <el-table-column label="排名" width="60" align="center">
                <template #default="{ $index }">
                  <span :class="['rank-badge', { 'top-3': $index < 3 }]">{{ $index + 1 }}</span>
                </template>
              </el-table-column>
              <el-table-column prop="display_name" label="用户" min-width="150">
                <template #default="{ row }">
                  <div class="user-cell">
                    <el-avatar :size="24" class="dist-avatar">{{ row.display_name?.charAt(0) }}</el-avatar>
                    <span>{{ row.display_name }}</span>
                  </div>
                </template>
              </el-table-column>
              <el-table-column label="总工时" min-width="120">
                <template #default="{ row }">
                  <span class="worklog-time-value">{{ formatWorklogTime(row.total_time_sec) }}</span>
                </template>
              </el-table-column>
              <el-table-column prop="entry_count" label="记录数" width="100" align="center" />
              <el-table-column label="占比" width="200">
                <template #default="{ row }">
                  <el-progress :percentage="getUserPercentage(row.total_time_sec)" :stroke-width="8" :show-text="true" :format="(p: number) => p.toFixed(1) + '%'" />
                </template>
              </el-table-column>
            </el-table>
            <el-empty v-if="!worklogStats.user_stats?.length" description="暂无工时数据" :image-size="60" />
          </el-card>

          <!-- 工时明细网格（用户 × 日期） -->
          <el-card shadow="never" class="chart-card">
            <template #header>
              <div class="card-header">
                <span class="card-dot" style="background: #3b82f6"></span>
                <span class="card-title">工时明细</span>
                <div class="grid-month-picker">
                  <el-button text :icon="ArrowLeft" size="small" @click="changeGridMonth(-1)" />
                  <span class="grid-month-label">{{ gridMonthLabel }}</span>
                  <el-button text :icon="ArrowRight" size="small" @click="changeGridMonth(1)" />
                </div>
              </div>
            </template>
            <div class="worklog-grid-wrap">
              <el-table
                v-if="worklogStats.grid?.length"
                :data="worklogGridData"
                stripe
                size="small"
                border
                show-summary
                :summary-method="worklogGridSummary"
                :header-cell-style="{ background: 'var(--td-table-header-bg)', color: 'var(--td-text-regular)', fontWeight: 600, fontSize: '12px', padding: '6px 0' }"
                :cell-style="worklogGridCellStyle"
                class="worklog-grid-table"
              >
                <el-table-column prop="display_name" label="用户" fixed width="120">
                  <template #default="{ row }">
                    <div class="user-cell">
                      <el-avatar :size="22" class="dist-avatar">{{ row.display_name?.charAt(0) }}</el-avatar>
                      <span>{{ row.display_name }}</span>
                    </div>
                  </template>
                </el-table-column>
                <el-table-column
                  v-for="d in (worklogStats.grid_dates || [])"
                  :key="d"
                  :prop="'d_' + d"
                  :label="formatGridDate(d)"
                  min-width="62"
                  align="center"
                  :class-name="isWeekend(d) ? 'weekend-col' : ''"
                >
                  <template #default="{ row }">
                    <el-popover
                      v-if="row['d_' + d]"
                      placement="top"
                      trigger="hover"
                      :width="260"
                      :show-after="200"
                    >
                      <template #reference>
                        <span class="grid-cell-value">{{ formatGridHours(row['d_' + d]) }}</span>
                      </template>
                      <div class="grid-popover">
                        <div class="grid-popover-header">
                          <span>{{ row.display_name }} · {{ dayjs(d).format('M月D日') }}</span>
                          <span class="grid-popover-total">{{ formatGridHours(row['d_' + d]) }}</span>
                        </div>
                        <div v-for="item in row['_details_' + d]" :key="item.issue_key" class="grid-popover-item">
                          <router-link :to="`/issues/${item.issue_key}`" class="grid-popover-key">{{ item.issue_key }}</router-link>
                          <span class="grid-popover-title">{{ item.title }}</span>
                          <span class="grid-popover-time">{{ formatGridHours(item.time_sec) }}</span>
                        </div>
                      </div>
                    </el-popover>
                  </template>
                </el-table-column>
                <el-table-column prop="_total" label="合计" fixed="right" width="80" align="center">
                  <template #default="{ row }">
                    <span class="grid-total-value">{{ formatGridHours(row._total) }}</span>
                  </template>
                </el-table-column>
              </el-table>
              <el-empty v-else description="暂无工时明细" :image-size="60" />
            </div>
          </el-card>
        </div>
      </el-tab-pane>
    </el-tabs>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, computed, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { DataAnalysis, Tickets, CircleCheck, Loading, Timer, Clock, User, Warning, ArrowLeft, ArrowRight } from '@element-plus/icons-vue'
import { getIssueStats, getSLAReport, getAlertStats, getUserPerformance, getWorklogStats } from '@/api/report'
import { getAllProjects } from '@/api/project'
import type { IssueStats, SLAReport, AlertStats, UserPerformance, WorklogStats } from '@/types/report'
import type { Project } from '@/types/project'
import dayjs from 'dayjs'

const route = useRoute()
const router = useRouter()

// 多色调色板
const typeColors = ['#3b82f6', '#8b5cf6', '#ec4899', '#06b6d4', '#10b981', '#f43f5e']
const assigneeColors = ['#8b5cf6', '#6366f1', '#a78bfa', '#7c3aed', '#c084fc', '#818cf8']
const epicColors = ['#f59e0b', '#f97316', '#fbbf24', '#fb923c', '#d97706', '#ea580c']

// 从 URL query 初始化筛选条件
const initQuery = route.query
const defaultStart = dayjs().subtract(30, 'day').format('YYYY-MM-DD')
const defaultEnd = dayjs().format('YYYY-MM-DD')

const dateRange = ref<[string, string]>([
  (initQuery.start_date as string) || defaultStart,
  (initQuery.end_date as string) || defaultEnd,
])

const dateShortcuts = [
  { text: '最近一周', value: () => [dayjs().subtract(7, 'day').toDate(), dayjs().toDate()] },
  { text: '最近一月', value: () => [dayjs().subtract(30, 'day').toDate(), dayjs().toDate()] },
  { text: '最近三月', value: () => [dayjs().subtract(90, 'day').toDate(), dayjs().toDate()] },
]

const selectedProject = ref((initQuery.project_key as string) || '')
const projects = ref<Project[]>([])

const validTabs = ['issues', 'sla', 'alerts', 'performance', 'worklogs']
const activeTab = ref(validTabs.includes(initQuery.tab as string) ? (initQuery.tab as string) : 'issues')

// 筛选条件同步到 URL query params
const syncQueryToUrl = () => {
  const query: Record<string, string> = {}
  if (activeTab.value && activeTab.value !== 'issues') query.tab = activeTab.value
  if (selectedProject.value) query.project_key = selectedProject.value
  // 仅当日期非默认值时持久化
  if (dateRange.value[0] !== defaultStart) query.start_date = dateRange.value[0]
  if (dateRange.value[1] !== defaultEnd) query.end_date = dateRange.value[1]
  router.replace({ query })
}
const loading = reactive({
  issues: false,
  sla: false,
  alerts: false,
  performance: false,
  worklogs: false,
})

const issueStats = ref<Partial<IssueStats>>({})
const slaReport = ref<Partial<SLAReport>>({})
const alertStats = ref<Partial<AlertStats>>({})
const userPerformance = ref<UserPerformance[]>([])
const worklogStats = ref<Partial<WorklogStats>>({})

// 工时明细月份（默认当月）
const gridMonth = ref(dayjs().format('YYYY-MM'))
const gridMonthLabel = computed(() => dayjs(gridMonth.value + '-01').format('YYYY年M月'))

// 工时统计调色板
const worklogTypeColors = ['#f59e0b', '#6366f1', '#10b981', '#ef4444', '#8b5cf6', '#ec4899']

// 加载项目列表
const loadProjects = async () => {
  try {
    const { data } = await getAllProjects()
    projects.value = data.data
  } catch {
    // ignored
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
  } catch {
    // ignored
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
  } catch {
    // ignored
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
  } catch {
    // ignored
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
  } catch {
    // ignored
  } finally {
    loading.performance = false
  }
}

// 加载工时统计
const loadWorklogStats = async () => {
  loading.worklogs = true
  try {
    const { data } = await getWorklogStats({
      project_key: selectedProject.value || undefined,
      start_date: dateRange.value[0],
      end_date: dateRange.value[1],
      grid_month: gridMonth.value,
    })
    worklogStats.value = data.data
  } catch {
    // ignored
  } finally {
    loading.worklogs = false
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
    case 'worklogs':
      loadWorklogStats()
      break
  }
}

const handleDateChange = () => {
  syncQueryToUrl()
  loadData()
}

const handleTabChange = () => {
  syncQueryToUrl()
  loadData()
}

const handleProjectChange = () => {
  syncQueryToUrl()
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

const formatDate = (date: string) => {
  if (!date) return '-'
  return dayjs(date).format('YYYY-MM-DD')
}

// 状态相关
type TagType = 'primary' | 'success' | 'warning' | 'info' | 'danger'

const getStatusLabel = (status: string) => {
  const map: Record<string, string> = { open: '待处理', in_progress: '进行中', pending_review: '待审核', resolved: '已完成', closed: '已终止', merged: '已合并' }
  return map[status] || status
}

const getStatusColor = (status: string) => {
  const map: Record<string, string> = { open: '#909399', in_progress: '#E6A23C', pending_review: '#6366f1', resolved: '#67C23A', closed: '#409EFF', merged: '#8b5cf6' }
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

const getSeverityTagType = (severity: string): TagType => {
  const map: Record<string, TagType> = { critical: 'danger', warning: 'warning', info: 'info' }
  return map[severity] || 'info'
}

// 用户绩效汇总（从数组计算）
const performanceSummary = computed(() => {
  const list = userPerformance.value
  const userCount = list.length
  const totalAssigned = list.reduce((s, u) => s + u.assigned, 0)
  const totalResolved = list.reduce((s, u) => s + u.resolved, 0)
  const avgResolveRate = totalAssigned > 0 ? (totalResolved / totalAssigned) * 100 : 0
  return { userCount, totalAssigned, totalResolved, avgResolveRate }
})

const getSLARateClass = (rate?: number) => {
  if (!rate) return ''
  if (rate >= 90) return 'success'
  if (rate >= 70) return 'warning'
  return 'danger'
}

// 工时格式化：秒 → 可读字符串
const formatWorklogTime = (seconds: number) => {
  if (!seconds || seconds <= 0) return '0h'
  const hours = Math.floor(seconds / 3600)
  const minutes = Math.floor((seconds % 3600) / 60)
  if (hours >= 8) {
    const days = Math.floor(hours / 8)
    const remainHours = hours % 8
    if (remainHours > 0) return `${days}d ${remainHours}h`
    return `${days}d`
  }
  if (hours > 0 && minutes > 0) return `${hours}h ${minutes}m`
  if (hours > 0) return `${hours}h`
  return `${minutes}m`
}

const formatShortDate = (date: string) => {
  if (!date) return ''
  return dayjs(date).format('MM/DD')
}

// 每日柱状图高度计算
const getDailyBarHeight = (timeSec: number) => {
  const maxTime = Math.max(...(worklogStats.value.daily_stats || []).map(d => d.total_time_sec), 1)
  return Math.max((timeSec / maxTime) * 100, 2)
}

// 类型占比计算
const getTypePercentage = (timeSec: number) => {
  const total = worklogStats.value.summary?.total_time_sec || 1
  return Math.round((timeSec / total) * 100)
}

// 用户占比计算
const getUserPercentage = (timeSec: number) => {
  const total = worklogStats.value.summary?.total_time_sec || 1
  return Math.round((timeSec / total) * 1000) / 10
}

// ========== 工时明细网格 ==========

// 将后端 grid 数据展平为 el-table 可用的行（d_YYYY-MM-DD 作为动态列 prop）
const worklogGridData = computed(() => {
  return (worklogStats.value.grid || []).map((row) => {
    const flat: Record<string, unknown> = {
      user_id: row.user_id,
      display_name: row.display_name,
      _total: row.total_sec,
    }
    for (const [date, sec] of Object.entries(row.daily || {})) {
      flat['d_' + date] = sec
    }
    for (const [date, details] of Object.entries(row.daily_details || {})) {
      flat['_details_' + date] = details
    }
    return flat
  })
})

// 网格日期列头格式
const formatGridDate = (date: string) => {
  const d = dayjs(date)
  const weekDay = ['日', '一', '二', '三', '四', '五', '六'][d.day()]
  return d.format('MM/DD') + '\n' + weekDay
}

// 网格单元格显示小时数
const formatGridHours = (sec: number) => {
  if (!sec) return ''
  const h = sec / 3600
  if (h >= 1) return h % 1 === 0 ? h + 'h' : h.toFixed(1) + 'h'
  return Math.round(sec / 60) + 'm'
}

// 判断是否周末
const isWeekend = (date: string) => {
  const day = dayjs(date).day()
  return day === 0 || day === 6
}

// 合计行
const worklogGridSummary = ({ columns }: { columns: { property: string }[] }) => {
  return columns.map((col, idx) => {
    if (idx === 0) return '合计'
    const prop = col.property
    if (prop === '_total') {
      const total = (worklogStats.value.grid || []).reduce((s, r) => s + r.total_sec, 0)
      return formatGridHours(total)
    }
    if (prop?.startsWith('d_')) {
      const date = prop.slice(2)
      const val = worklogStats.value.grid_totals?.[date]
      return val ? formatGridHours(val) : ''
    }
    return ''
  })
}

// 周末列灰底（通过 class-name 控制，不再需要 inline style）
const worklogGridCellStyle = () => ({})

// 月份切换
const changeGridMonth = (delta: number) => {
  gridMonth.value = dayjs(gridMonth.value + '-01').add(delta, 'month').format('YYYY-MM')
  loadWorklogStats()
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

// 页面头部 icon (TdPageHeader leading slot)
.page-header-icon {
  width: 40px;
  height: 40px;
  background: var(--td-tag-primary-bg);
  border-radius: var(--td-radius-md);
  display: flex;
  align-items: center;
  justify-content: center;
  color: var(--td-color-primary);
  flex-shrink: 0;
}

// Tabs
.report-tabs {
  :deep(.el-tabs__header) {
    margin-bottom: 20px;
  }
  :deep(.el-tabs__item) {
    font-size: 14px;
    font-weight: 500;
  }
}

.tab-content {
  min-height: 400px;
}

// ========== 汇总卡片 ==========
.summary-row {
  margin-bottom: 16px;
}

.summary-card {
  background: var(--td-bg-card);
  border-radius: 12px;
  padding: 18px 20px;
  display: flex;
  align-items: center;
  gap: 14px;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.04), 0 1px 2px rgba(0, 0, 0, 0.06);
  border: 1px solid var(--td-divider-color);
  transition: box-shadow 150ms ease-out, border-color 150ms ease-out;

  &:hover {
    border-color: var(--td-border-color);
    box-shadow: 0 4px 12px rgba(0, 0, 0, 0.08);
  }

  .summary-icon-wrap {
    width: 42px;
    height: 42px;
    border-radius: 10px;
    display: flex;
    align-items: center;
    justify-content: center;
    flex-shrink: 0;

    &.default { background: var(--td-tag-primary-bg); color: #6366f1; }
    &.success { background: var(--td-tag-success-bg); color: var(--td-color-success); }
    &.warning { background: var(--td-tag-warning-bg); color: var(--td-color-warning); }
    &.info    { background: var(--td-tag-primary-bg); color: var(--td-color-primary); }
    &.danger  { background: var(--td-tag-danger-bg); color: var(--td-color-danger); }
  }

  .summary-body {
    min-width: 0;
  }

  .summary-value {
    font-size: 24px;
    font-weight: 700;
    color: var(--td-text-primary);
    line-height: 1.2;
  }

  .summary-label {
    font-size: 12px;
    color: var(--td-text-secondary);
    margin-top: 2px;
    white-space: nowrap;
  }
}

// ========== 等高行 ==========
.stretch-row {
  :deep(.el-col) {
    display: flex;
    flex-direction: column;
  }

  .chart-card {
    flex: 1;
  }
}

// ========== 图表卡片 ==========
.chart-card {
  margin-bottom: 16px;
  border-radius: 12px;
  border: 1px solid var(--td-divider-color);

  :deep(.el-card__header) {
    padding: 14px 20px;
    border-bottom: 1px solid var(--td-divider-color);
  }

  :deep(.el-card__body) {
    padding: 16px 20px;
  }

  .card-header {
    display: flex;
    align-items: center;
    gap: 8px;
  }

  .card-dot {
    width: 8px;
    height: 8px;
    border-radius: 50%;
    flex-shrink: 0;
  }

  .card-title {
    font-size: 14px;
    font-weight: 600;
    color: var(--td-text-primary);
  }
}

// ========== 分布列表 ==========
.distribution-list {
  max-height: 320px;
  overflow-y: auto;

  // Epic 全宽时双列布局
  &--horizontal {
    max-height: none;
    display: grid;
    grid-template-columns: repeat(2, 1fr);
    gap: 0 24px;

    .distribution-item:last-child {
      margin-bottom: 0;
    }
  }

  .distribution-item {
    margin-bottom: 14px;

    &:last-child {
      margin-bottom: 0;
    }

    .dist-header {
      display: flex;
      align-items: center;
      margin-bottom: 6px;
      gap: 8px;
    }

    .dist-dot {
      width: 8px;
      height: 8px;
      border-radius: 50%;
      flex-shrink: 0;
    }

    .dist-avatar {
      background: #8b5cf6;
      font-size: 11px;
      color: var(--td-text-white);
      flex-shrink: 0;
    }

    .dist-name {
      flex: 1;
      font-size: 13px;
      color: var(--td-text-regular);
      overflow: hidden;
      text-overflow: ellipsis;
      white-space: nowrap;
    }

    .dist-value {
      font-size: 13px;
      font-weight: 600;
      color: var(--td-text-primary);
      white-space: nowrap;

      .dist-ratio {
        font-weight: 400;
        font-size: 11px;
        color: var(--td-text-secondary);
        margin-left: 4px;
      }
    }

    :deep(.el-progress-bar__outer) {
      border-radius: 4px;
      background: var(--td-bg-section);
    }

    :deep(.el-progress-bar__inner) {
      border-radius: 4px;
    }
  }
}

// ========== 时间线表格 ==========
.timeline-list {
  max-height: 420px;
  overflow-y: auto;

  .timeline-date {
    font-size: 13px;
    color: var(--td-text-regular);
    font-variant-numeric: tabular-nums;
  }

  .timeline-badge {
    display: inline-flex;
    align-items: center;
    justify-content: center;
    min-width: 32px;
    padding: 2px 10px;
    border-radius: 12px;
    font-size: 12px;
    font-weight: 600;

    &.created {
      background: var(--td-tag-primary-bg);
      color: #6366f1;
    }
    &.in-progress {
      background: var(--td-tag-warning-bg);
      color: var(--td-color-warning);
    }
    &.resolved {
      background: var(--td-tag-success-bg);
      color: var(--td-color-success);
    }
    &.closed {
      background: var(--td-tag-primary-bg);
      color: var(--td-color-primary);
    }
  }

  :deep(.el-table) {
    border-radius: 8px;
    --el-table-border-color: var(--td-border-color-light);
  }
}

// ========== Top 列表 ==========
.top-list {
  .top-item {
    display: flex;
    align-items: center;
    padding: 10px 0;
    border-bottom: 1px solid var(--td-divider-color);

    &:last-child {
      border-bottom: none;
    }

    .top-rank {
      width: 22px;
      height: 22px;
      background: var(--td-bg-section);
      border-radius: 6px;
      display: flex;
      align-items: center;
      justify-content: center;
      font-size: 11px;
      font-weight: 700;
      color: var(--td-text-secondary);
      margin-right: 12px;
      flex-shrink: 0;
    }

    &:nth-child(1) .top-rank { background: var(--td-tag-warning-border); color: var(--td-color-warning); }
    &:nth-child(2) .top-rank { background: var(--td-bg-section); color: var(--td-text-regular); }
    &:nth-child(3) .top-rank { background: var(--td-tag-orange-border); color: var(--td-tag-orange-text); }

    .top-name {
      flex: 1;
      font-size: 13px;
      color: var(--td-text-primary);
      overflow: hidden;
      text-overflow: ellipsis;
      white-space: nowrap;
    }

    .top-count {
      font-size: 13px;
      font-weight: 700;
      color: #6366f1;
    }
  }
}

// ========== 响应式 ==========
@media (max-width: 768px) {
  .summary-card {
    padding: 14px;

    .summary-value {
      font-size: 20px;
    }
  }
}

// 工时统计柱状图
.daily-worklog-chart {
  min-height: 240px;
}

.bar-chart {
  display: flex;
  align-items: flex-end;
  gap: 4px;
  height: 220px;
  padding: 20px 0 0;
  overflow-x: auto;
}

.bar-item {
  display: flex;
  flex-direction: column;
  align-items: center;
  flex: 1;
  min-width: 36px;
  max-width: 60px;
  height: 100%;
  justify-content: flex-end;

  .bar-value {
    font-size: 10px;
    color: var(--td-text-secondary);
    margin-bottom: 4px;
    white-space: nowrap;
  }

  .bar-fill {
    width: 100%;
    max-width: 32px;
    background: var(--td-color-primary);
    border-radius: 4px 4px 0 0;
    min-height: 2px;
    transition: height 150ms ease-out;
  }

  .bar-label {
    font-size: 10px;
    color: var(--td-text-secondary);
    margin-top: 6px;
    white-space: nowrap;
  }
}

.rank-badge {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 24px;
  height: 24px;
  border-radius: 50%;
  font-size: 12px;
  font-weight: 600;
  color: var(--td-text-secondary);
  background: var(--td-bg-section);

  &.top-3 {
    color: var(--td-text-white);
    background: var(--td-color-warning);
  }
}

.user-cell {
  display: flex;
  align-items: center;
  gap: 8px;
}

.worklog-time-value {
  font-weight: 600;
  color: var(--td-text-primary);
}

// ========== SLA / 告警 补充样式 ==========
.card-count {
  margin-left: auto;
  font-size: 12px;
  font-weight: 500;
  color: var(--td-text-secondary);
  background: var(--td-bg-section);
  padding: 2px 8px;
  border-radius: 10px;
}

.issue-link {
  color: var(--td-color-primary);
  text-decoration: none;
  font-weight: 500;

  &:hover {
    text-decoration: underline;
  }
}

.text-danger {
  color: var(--td-color-danger);
  font-weight: 600;
}

.project-key-badge {
  display: inline-block;
  background: var(--td-tag-primary-bg);
  color: #6366f1;
  font-size: 11px;
  font-weight: 600;
  padding: 1px 6px;
  border-radius: 4px;
  margin-right: 6px;
}

.project-name-text {
  font-size: 13px;
  color: var(--td-text-regular);
}

.alert-count-value {
  font-weight: 700;
  color: #6366f1;
}

// ========== accent-danger 补充 ==========
.summary-card.accent-danger {
  .summary-icon-wrap.danger {
    background: var(--td-tag-danger-bg);
    color: var(--td-color-danger);
  }
}

// ========== 工时明细网格 ==========
.grid-month-picker {
  margin-left: auto;
  display: flex;
  align-items: center;
  gap: 4px;

  .grid-month-label {
    font-size: 13px;
    font-weight: 600;
    color: var(--td-text-primary);
    min-width: 80px;
    text-align: center;
  }
}

.worklog-grid-wrap {
  overflow-x: auto;
}

.worklog-grid-table {
  :deep(.el-table__header) th {
    white-space: pre-line;
    line-height: 1.3;
    text-align: center;
  }

  :deep(.el-table__footer) td {
    font-weight: 700;
    color: var(--td-text-primary);
  }

  // 周末列灰底
  :deep(.weekend-col) {
    background: var(--td-bg-section) !important;
  }

  .grid-cell-value {
    font-size: 12px;
    font-weight: 500;
    color: var(--td-text-primary);
    background: var(--td-tag-primary-bg);
    padding: 2px 6px;
    border-radius: 4px;
  }

  .grid-total-value {
    font-size: 12px;
    font-weight: 700;
    color: #6366f1;
  }
}

// ========== 工时明细 Popover ==========
.grid-popover {
  .grid-popover-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    font-size: 13px;
    font-weight: 600;
    color: var(--td-text-primary);
    padding-bottom: 8px;
    margin-bottom: 8px;
    border-bottom: 1px solid var(--td-divider-color);
  }

  .grid-popover-total {
    color: #6366f1;
  }

  .grid-popover-item {
    display: flex;
    align-items: center;
    gap: 8px;
    padding: 4px 0;
    font-size: 12px;

    .grid-popover-key {
      color: var(--td-color-primary);
      text-decoration: none;
      font-weight: 500;
      flex-shrink: 0;

      &:hover {
        text-decoration: underline;
      }
    }

    .grid-popover-title {
      flex: 1;
      color: var(--td-text-regular);
      overflow: hidden;
      text-overflow: ellipsis;
      white-space: nowrap;
    }

    .grid-popover-time {
      font-weight: 600;
      color: var(--td-text-primary);
      flex-shrink: 0;
    }
  }
}
</style>
