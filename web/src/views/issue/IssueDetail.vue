<template>
  <div v-loading="loading" class="issue-detail-container">
    <template v-if="issue">
      <!-- 头部信息 -->
      <div class="issue-header">
        <div class="header-left">
          <div class="issue-breadcrumb">
            <el-breadcrumb separator="/">
              <el-breadcrumb-item :to="{ path: '/issues' }">工单列表</el-breadcrumb-item>
              <el-breadcrumb-item :to="{ path: '/issues?project_key=' + issue.project_key }">{{ issue.project_key }}</el-breadcrumb-item>
              <el-breadcrumb-item>{{ issue.issue_key }}</el-breadcrumb-item>
            </el-breadcrumb>
          </div>
          <h1 class="issue-title">{{ issue.title }}</h1>
          <div class="issue-meta">
            <div v-if="issue.issue_type" class="type-badge">
              <el-icon class="type-icon"><Document /></el-icon>
              <span class="type-text">{{ issue.issue_type.display_name }}</span>
            </div>
            <el-tag :type="getPriorityType(issue.priority)" size="small" effect="dark">{{ issue.priority }}</el-tag>
            <div class="status-badge" :class="issue.status">
              <span class="status-dot"></span>
              <span>{{ getStatusText(issue.status) }}</span>
            </div>
            <span class="meta-item">
              <el-icon><User /></el-icon>
              {{ issue.reporter?.display_name || '未知' }}
            </span>
            <span class="meta-item">
              <el-icon><Clock /></el-icon>
              {{ formatTime(issue.created_at) }}
            </span>
          </div>
        </div>
        <div class="header-actions">
          <!-- 工作流快捷操作按钮 -->
          <el-dropdown v-if="workflowInstance" trigger="hover" @command="handleWorkflowCommand">
            <el-button
              :type="canOperateWorkflow ? 'primary' : 'info'"
              :disabled="!canOperateWorkflow && !isWorkflowOperable"
              class="workflow-action-btn"
            >
              <el-icon><Promotion /></el-icon>
              {{ workflowActionBtnText }}
              <el-icon class="el-icon--right"><ArrowDown /></el-icon>
            </el-button>
            <template #dropdown>
              <el-dropdown-menu>
                <!-- 审批节点操作 -->
                <template v-if="workflowInstance.current_node?.node_type === 'approval' && isWorkflowOperable">
                  <el-dropdown-item
                    command="approve"
                    :disabled="!isCurrentUserApprover"
                  >
                    <el-icon style="color: #67c23a"><Check /></el-icon>
                    审批通过
                  </el-dropdown-item>
                  <el-dropdown-item
                    command="reject"
                    :disabled="!isCurrentUserApprover"
                  >
                    <el-icon style="color: #f56c6c"><Delete /></el-icon>
                    审批拒绝
                  </el-dropdown-item>
                </template>
                <!-- 工作节点操作 -->
                <template v-else-if="isWorkNode && isWorkflowOperable">
                  <!-- 有条件分支的工作节点：动态生成操作按钮 -->
                  <template v-if="workNodeHasBranching">
                    <el-dropdown-item
                      v-for="action in workNodeOutgoingActions"
                      :key="action.conditionExpr"
                      :command="`complete-condition:${action.conditionExpr}`"
                    >
                      <el-icon :style="{ color: action.conditionExpr === 'rejected' ? '#f56c6c' : '#67c23a' }">
                        <component :is="action.conditionExpr === 'rejected' ? Delete : Check" />
                      </el-icon>
                      {{ action.label }}{{ action.targetNodeName ? ` → ${action.targetNodeName}` : '' }}
                    </el-dropdown-item>
                  </template>
                  <!-- 无分支的工作节点：显示完成 -->
                  <template v-else>
                    <el-dropdown-item command="complete">
                      <el-icon style="color: #409eff"><Check /></el-icon>
                      {{ nextNodeName ? `流转至: ${nextNodeName}` : '完成节点' }}
                    </el-dropdown-item>
                  </template>
                </template>
                <!-- 工作流可操作但节点信息缺失（如 reviewing 状态且节点被重建） -->
                <template v-else-if="isWorkflowOperable">
                  <el-dropdown-item command="complete">
                    <el-icon style="color: #409eff"><Check /></el-icon>
                    确认完成
                  </el-dropdown-item>
                </template>
                <!-- 工作流已结束提示 -->
                <template v-else>
                  <el-dropdown-item disabled>
                    工作流{{ getWorkflowStatusText(workflowInstance.status) }}
                  </el-dropdown-item>
                </template>
                <el-dropdown-item divided command="view-workflow">
                  <el-icon style="color: #909399"><View /></el-icon>
                  查看工作流
                </el-dropdown-item>
              </el-dropdown-menu>
            </template>
          </el-dropdown>
          <el-button :icon="Edit" @click="handleEdit">编辑</el-button>
          <el-button type="danger" :icon="Delete" @click="handleDelete">删除</el-button>
        </div>
      </div>

      <el-row :gutter="20">
        <!-- 左侧：描述和评论 -->
        <el-col :xs="24" :lg="16">
          <!-- 描述 -->
          <el-card shadow="never" class="content-card">
            <template #header>
              <div class="card-header-group">
                <div class="card-icon desc">
                  <el-icon><Document /></el-icon>
                </div>
                <span class="card-title">描述</span>
              </div>
            </template>
            <div v-if="issue.description" class="description-content">
              {{ issue.description }}
            </div>
            <div v-else class="empty-placeholder">暂无描述</div>
          </el-card>

          <!-- Epic 下的 Issues（仅当当前工单是 Epic 类型时显示）-->
          <el-card v-if="issue.issue_type?.name?.toLowerCase() === 'epic'" shadow="never" class="content-card epic-issues-card">
            <template #header>
              <div class="card-header-group">
                <div class="card-icon epic">
                  <el-icon><Document /></el-icon>
                </div>
                <span class="card-title">长篇故事中的事物 ({{ epicIssues.length }})</span>
              </div>
            </template>
            <div class="epic-issues-list">
              <div v-if="epicIssues.length === 0" class="empty-state">
                <el-empty description="暂无关联的工单" :image-size="80" />
              </div>
              <div v-for="epicIssue in epicIssues" :key="epicIssue.id" class="epic-issue-item">
                <div class="issue-left">
                  <div class="issue-type-icon" :class="epicIssue.issue_type?.name?.toLowerCase() || 'task'">
                    <el-icon><Document /></el-icon>
                  </div>
                  <router-link :to="`/issues/${epicIssue.issue_key}`" class="issue-link">
                    <span class="issue-key">{{ epicIssue.issue_key }}</span>
                  </router-link>
                  <div class="issue-title">{{ epicIssue.title }}</div>
                </div>
                <div class="issue-right">
                  <el-tag :type="getPriorityType(epicIssue.priority)" size="small" effect="dark" class="priority-tag">
                    {{ epicIssue.priority }}
                  </el-tag>
                  <div class="status-badge" :class="epicIssue.status">
                    <span class="status-dot"></span>
                    <span>{{ getStatusText(epicIssue.status) }}</span>
                  </div>
                  <div v-if="epicIssue.assignee" class="assignee-info">
                    <div class="assignee-avatar" :title="epicIssue.assignee.display_name">
                      {{ epicIssue.assignee.display_name?.charAt(0) || '?' }}
                    </div>
                    <span class="assignee-name">{{ epicIssue.assignee.display_name }}</span>
                  </div>
                  <div v-else class="assignee-info">
                    <div class="assignee-avatar unassigned" title="未分配">?</div>
                    <span class="assignee-name unassigned">未分配</span>
                  </div>
                </div>
              </div>
            </div>
          </el-card>

          <!-- 工作流卡片 -->
          <el-card v-if="workflowInstance" shadow="never" class="content-card workflow-card">
            <template #header>
              <div class="card-header-group" style="flex: 1;">
                <div class="card-icon workflow">
                  <el-icon><Promotion /></el-icon>
                </div>
                <span class="card-title">工作流</span>
                <el-tag :type="getWorkflowStatusType(workflowInstance.status)" size="small" effect="dark">
                  {{ getWorkflowStatusText(workflowInstance.status) }}
                </el-tag>
                <el-button
                  v-if="(workflowInstance.approvals && workflowInstance.approvals.length > 0) || workflowHistoryList.length > 0"
                  link
                  size="small"
                  style="margin-left: auto;"
                  @click.stop="workflowExpanded = !workflowExpanded"
                >
                  {{ workflowExpanded ? '收起' : '展开详情' }}
                  <el-icon><component :is="workflowExpanded ? ArrowUp : ArrowDown" /></el-icon>
                </el-button>
              </div>
            </template>

            <!-- 当前节点信息 -->
            <div class="workflow-current-node">
              <div class="current-node-label">当前节点</div>
              <div class="current-node-info">
                <el-tag size="default" effect="plain">
                  {{ workflowInstance.current_node?.name || `节点#${workflowInstance.current_node_id}` }}
                </el-tag>
                <el-tag v-if="workflowInstance.current_node?.node_type" size="small" type="info">
                  {{ workflowInstance.current_node.node_type === 'approval' ? '审批节点' : workflowInstance.current_node.node_type === 'work' ? '工作节点' : workflowInstance.current_node.node_type }}
                </el-tag>
              </div>
            </div>

            <!-- 审批记录（默认折叠） -->
            <div v-show="workflowExpanded" v-if="workflowInstance.approvals && workflowInstance.approvals.length > 0" class="workflow-approvals">
              <div class="approvals-label">审批记录</div>
              <div class="approvals-list">
                <div v-for="approval in workflowInstance.approvals" :key="approval.id" class="approval-item">
                  <div class="approval-user">
                    <div class="mini-avatar">{{ approval.approver_name?.charAt(0) || '?' }}</div>
                    <span>{{ approval.approver_name || `用户#${approval.approver_id}` }}</span>
                  </div>
                  <el-tag :type="getApprovalStatusType(approval.status)" size="small">
                    {{ getApprovalStatusText(approval.status) }}
                  </el-tag>
                  <span v-if="approval.comment" class="approval-comment">{{ approval.comment }}</span>
                </div>
              </div>
            </div>

            <!-- 审批操作按钮（仅审批节点显示） -->
            <div v-if="isCurrentUserApprover && isWorkflowOperable && workflowInstance.current_node?.node_type === 'approval'" class="workflow-actions">
              <div class="actions-label">审批操作</div>
              <div class="actions-row">
                <el-input v-model="approveComment" placeholder="审批意见（可选）" size="default" style="flex: 1; margin-right: 12px;" />
                <el-button type="success" :loading="approveLoading" @click="handleApprove">
                  <el-icon><Check /></el-icon>
                  通过
                </el-button>
                <el-button type="danger" @click="showRejectDialog">
                  <el-icon><Delete /></el-icon>
                  拒绝
                </el-button>
              </div>
            </div>

            <!-- 工作节点完成按钮 -->
            <div v-if="isWorkNode" class="workflow-actions">
              <div class="actions-label">{{ workNodeHasBranching ? '节点操作' : (nextNodeName ? '流转操作' : '节点操作') }}</div>
              <div class="actions-row">
                <el-input v-model="completeComment" placeholder="备注（可选）" size="default" style="flex: 1; margin-right: 12px;" />
                <!-- 有条件分支的工作节点：动态生成操作按钮 -->
                <template v-if="workNodeHasBranching">
                  <el-button
                    v-for="action in workNodeOutgoingActions"
                    :key="action.conditionExpr"
                    :type="action.conditionExpr === 'rejected' ? 'warning' : 'success'"
                    :loading="completeLoading"
                    @click="handleCompleteWithResult(action.conditionExpr)"
                  >
                    <el-icon><component :is="action.conditionExpr === 'rejected' ? Delete : Check" /></el-icon>
                    {{ action.label }}{{ action.targetNodeName ? ` → ${action.targetNodeName}` : '' }}
                  </el-button>
                </template>
                <!-- 无分支的工作节点：显示完成按钮 -->
                <template v-else>
                  <el-button type="primary" :loading="completeLoading" @click="handleComplete">
                    <el-icon><Check /></el-icon>
                    {{ nextNodeName ? `流转至: ${nextNodeName}` : '完成节点' }}
                  </el-button>
                </template>
              </div>
            </div>

            <!-- 流转历史时间线（默认折叠） -->
            <div v-show="workflowExpanded" v-if="workflowHistoryList.length > 0" class="workflow-history">
              <div class="history-label">流转历史</div>
              <el-timeline class="workflow-timeline">
                <el-timeline-item
                  v-for="item in workflowHistoryList"
                  :key="item.id"
                  :timestamp="formatTime(item.operated_at || item.created_at)"
                  placement="top"
                  :type="item.action === 'reject' ? 'danger' : item.action === 'complete' ? 'success' : 'primary'"
                >
                  <div class="history-content">
                    <span class="history-user">{{ item.operator_name || (item.operator_id === 0 ? '系统' : `用户#${item.operator_id}`) }}</span>
                    <span class="history-action">{{ getHistoryActionText(item.action) }}</span>
                    <template v-if="item.to_node">
                      <span class="history-arrow">→</span>
                      <el-tag size="small">{{ item.to_node.name }}</el-tag>
                    </template>
                    <div v-if="item.comment" class="history-comment">{{ item.comment }}</div>
                  </div>
                </el-timeline-item>
              </el-timeline>
            </div>
          </el-card>

          <!-- 扩展字段 - 显示在主体区域 -->
          <el-card v-if="customFields.length > 0" shadow="never" class="content-card custom-fields-card">
            <template #header>
              <div class="card-header-group">
                <div class="card-icon custom">
                  <el-icon><Document /></el-icon>
                </div>
                <span class="card-title">扩展字段</span>
              </div>
            </template>
            <div class="custom-fields-grid">
              <div v-for="field in customFields" :key="field.field_id" class="field-item">
                <div class="field-label">{{ field.field_name }}</div>
                <div class="field-value">
                  <template v-if="isFieldValueSet(field)">
                    <!-- Epic Link 特殊显示 -->
                    <template v-if="field.field_type === 'epic_link' && typeof field.value === 'object' && field.value.issue_key">
                      <router-link :to="`/issues/${field.value.issue_key}`" class="epic-link">
                        <el-icon><Link /></el-icon>
                        {{ field.value.issue_key }}: {{ field.value.title }}
                      </router-link>
                    </template>
                    <!-- 多选类型：标签展示 -->
                    <template v-else-if="Array.isArray(field.value)">
                      <el-tag v-for="item in field.value" :key="item" size="small" style="margin-right: 4px;">{{ item }}</el-tag>
                    </template>
                    <!-- 其他字段 -->
                    <template v-else>
                      {{ field.display_value || field.value }}
                    </template>
                  </template>
                  <span v-else class="empty-value">未设置</span>
                </div>
              </div>
            </div>
          </el-card>

          <!-- 子任务列表 -->
          <el-card v-if="subtasks.length > 0 || issue.issue_type?.name?.toLowerCase() === 'task'" shadow="never" class="content-card subtasks-card">
            <template #header>
              <div class="card-header-with-action">
                <div class="card-header-group">
                  <div class="card-icon subtask">
                    <el-icon><Document /></el-icon>
                  </div>
                  <span class="card-title">子任务 ({{ subtasks.length }})</span>
                </div>
                <el-button link type="primary" size="small" @click="handleCreateSubtask">
                  <el-icon><Plus /></el-icon>
                  创建子任务
                </el-button>
              </div>
            </template>
            <div class="subtasks-list">
              <div v-if="subtasks.length === 0" class="empty-state-compact">
                <span class="empty-text">暂无子任务</span>
              </div>
              <div v-for="subtask in subtasks" :key="subtask.id" class="subtask-item">
                <div class="issue-left">
                  <div class="issue-type-icon" :class="subtask.issue_type?.name?.toLowerCase() || 'task'">
                    <el-icon><Document /></el-icon>
                  </div>
                  <router-link :to="`/issues/${subtask.issue_key}`" class="issue-link">
                    <span class="issue-key">{{ subtask.issue_key }}</span>
                  </router-link>
                  <div class="issue-title">{{ subtask.title }}</div>
                </div>
                <div class="issue-right">
                  <el-tag :type="getPriorityType(subtask.priority)" size="small" effect="dark" class="priority-tag">
                    {{ subtask.priority }}
                  </el-tag>
                  <div class="status-badge" :class="subtask.status">
                    <span class="status-dot"></span>
                    <span>{{ getStatusText(subtask.status) }}</span>
                  </div>
                  <div v-if="subtask.assignee" class="assignee-info">
                    <div class="assignee-avatar" :title="subtask.assignee.display_name">
                      {{ subtask.assignee.display_name?.charAt(0) || '?' }}
                    </div>
                    <span class="assignee-name">{{ subtask.assignee.display_name }}</span>
                  </div>
                  <div v-else class="assignee-info">
                    <div class="assignee-avatar unassigned" title="未分配">?</div>
                    <span class="assignee-name unassigned">未分配</span>
                  </div>
                </div>
              </div>
            </div>
          </el-card>

          <!-- 关联告警 -->
          <el-card v-if="issueAlerts.length > 0" shadow="never" class="content-card alert-card">
            <template #header>
              <div class="card-header-with-action">
                <div class="card-header-group">
                  <div class="card-icon alert">
                    <el-icon><Bell /></el-icon>
                  </div>
                  <span class="card-title">关联告警 ({{ issueAlerts.length }})</span>
                </div>
                <el-button link type="primary" size="small" @click="$router.push(`/alerts?issue_id=${issue!.id}`)">
                  在告警列表中查看
                </el-button>
              </div>
            </template>
            <el-table :data="issueAlerts" style="width: 100%" size="small" :row-class-name="() => 'clickable-row'" @row-click="(row: Alert) => $router.push(`/alerts/${row.id}`)">
              <el-table-column prop="alert_name" label="告警名称" min-width="180">
                <template #default="{ row }">
                  <span class="alert-name-text">{{ row.alert_name }}</span>
                </template>
              </el-table-column>
              <el-table-column prop="severity" label="严重程度" width="90" align="center">
                <template #default="{ row }">
                  <el-tag :type="getAlertSeverityType(row.severity)" size="small" effect="dark">
                    {{ getAlertSeverityText(row.severity) }}
                  </el-tag>
                </template>
              </el-table-column>
              <el-table-column prop="status" label="状态" width="90" align="center">
                <template #default="{ row }">
                  <div class="alert-status-badge" :class="row.status">
                    <span class="status-dot"></span>
                    <span>{{ getAlertStatusText(row.status) }}</span>
                  </div>
                </template>
              </el-table-column>
              <el-table-column label="实例" min-width="140">
                <template #default="{ row }">
                  <span class="text-muted">{{ row.labels?.instance || row.labels?.target_ident || '-' }}</span>
                </template>
              </el-table-column>
              <el-table-column prop="starts_at" label="开始时间" width="150">
                <template #default="{ row }">
                  <span class="text-muted">{{ formatTime(row.starts_at) }}</span>
                </template>
              </el-table-column>
            </el-table>
          </el-card>

          <!-- 附件 -->
          <el-card shadow="never" class="content-card attachment-card">
            <template #header>
              <div class="card-header-with-action">
                <div class="card-header-group">
                  <div class="card-icon attachment">
                    <el-icon><Paperclip /></el-icon>
                  </div>
                  <span class="card-title">附件</span>
                  <span class="card-count">{{ attachments.length }}</span>
                </div>
                <el-button v-if="!showAttachmentUpload && attachments.length === 0" link type="primary" size="small" @click="showAttachmentUpload = true">
                  <el-icon><Plus /></el-icon>
                  上传附件
                </el-button>
              </div>
            </template>
            <div class="attachment-section">
              <AttachmentUpload v-if="issue && (showAttachmentUpload || attachments.length > 0)" :issue-key="issue.issue_key" @success="loadAttachments(issue.issue_key)" />
              <AttachmentList v-if="issue && attachments.length > 0" :issue-key="issue.issue_key" :attachments="attachments" @refresh="loadAttachments(issue.issue_key)" />
              <div v-if="attachments.length === 0 && !showAttachmentUpload" class="empty-placeholder sm">暂无附件</div>
            </div>
          </el-card>

          <!-- 评论和工作日志 -->
          <el-card shadow="never" class="content-card">
            <el-tabs v-model="activeTab" class="detail-tabs">
              <el-tab-pane name="comments">
                <template #label>
                  <span class="tab-label"><el-icon><ChatLineRound /></el-icon> 评论 ({{ comments.length }})</span>
                </template>

                <!-- 添加评论 -->
                <div class="add-comment">
                  <el-input
                    v-model="newComment"
                    type="textarea"
                    :rows="3"
                    placeholder="添加评论..."
                  />
                  <div class="comment-actions">
                    <el-button type="primary" :loading="commentLoading" :disabled="!newComment.trim()" @click="submitComment">
                      <el-icon><ChatLineRound /></el-icon>
                      发表评论
                    </el-button>
                  </div>
                </div>

                <!-- 评论列表 -->
                <div class="comment-list">
                  <div v-for="comment in comments" :key="comment.id" :class="['comment-item', { 'system-comment': comment.user_id === 0 }]">
                    <div :class="['comment-avatar', { 'system-avatar': comment.user_id === 0 }]">
                      {{ comment.user_id === 0 ? '系' : (comment.user?.display_name?.charAt(0) || '?') }}
                    </div>
                    <div class="comment-body">
                      <div class="comment-header">
                        <span :class="['comment-author', { 'system-author': comment.user_id === 0 }]">{{ comment.user_id === 0 ? '系统' : (comment.user?.display_name || '未知用户') }}</span>
                        <span class="comment-time">{{ formatTime(comment.created_at) }}</span>
                      </div>
                      <div class="comment-text">{{ comment.content }}</div>
                    </div>
                  </div>
                  <div v-if="comments.length === 0" class="empty-placeholder">
                    暂无评论
                  </div>
                </div>
              </el-tab-pane>

              <el-tab-pane name="worklogs">
                <template #label>
                  <span class="tab-label"><el-icon><Clock /></el-icon> 工作日志 ({{ worklogs.length }})</span>
                </template>

                <!-- 添加工作日志 -->
                <div class="add-worklog">
                  <el-form :model="worklogForm" label-position="top" size="default">
                    <el-form-item label="工作描述">
                      <el-input
                        v-model="worklogForm.description"
                        type="textarea"
                        :rows="3"
                        placeholder="描述本次工作内容..."
                      />
                    </el-form-item>
                    <el-row :gutter="16">
                      <el-col :span="8">
                        <el-form-item label="工作时长">
                          <el-input v-model="worklogForm.time_spent" placeholder="如: 2h 30m" />
                          <div class="form-hint">格式：1d 2h 30m</div>
                        </el-form-item>
                      </el-col>
                      <el-col :span="8">
                        <el-form-item label="工作日期">
                          <el-date-picker
                            v-model="worklogForm.worked_at"
                            type="datetime"
                            placeholder="选择日期时间"
                            style="width: 100%"
                          />
                        </el-form-item>
                      </el-col>
                      <el-col :span="8">
                        <el-form-item label="工作类型">
                          <el-select v-model="worklogForm.work_type" placeholder="选择类型" style="width: 100%" clearable>
                            <el-option v-for="opt in workTypeOptions" :key="opt.value" :label="opt.label" :value="opt.value" />
                          </el-select>
                        </el-form-item>
                      </el-col>
                    </el-row>
                    <el-button type="primary" :loading="worklogLoading" :disabled="!canSubmitWorklog" @click="submitWorklog">
                      <el-icon><Plus /></el-icon>
                      添加工作日志
                    </el-button>
                  </el-form>
                </div>

                <!-- 工作日志列表 -->
                <div class="worklog-list">
                  <div v-if="totalTimeSpent > 0" class="worklog-summary">
                    <el-icon><Clock /></el-icon>
                    <span>总工作时长：{{ formatTimeSpent(totalTimeSpent) }}</span>
                  </div>
                  <div v-for="worklog in worklogs" :key="worklog.id" class="worklog-item">
                    <div class="worklog-avatar">
                      {{ worklog.user?.display_name?.charAt(0) || '?' }}
                    </div>
                    <div class="worklog-body">
                      <div class="worklog-header">
                        <div class="worklog-meta">
                          <span class="worklog-author">{{ worklog.user?.display_name || '未知用户' }}</span>
                          <el-tag size="small" type="info">{{ worklog.time_spent }}</el-tag>
                          <el-tag v-if="worklog.work_type" size="small" type="success">{{ worklog.work_type }}</el-tag>
                        </div>
                        <div class="worklog-actions">
                          <span class="worklog-time">{{ formatTime(worklog.worked_at) }}</span>
                          <el-button v-if="canEditWorklog(worklog)" link type="primary" size="small" @click="handleEditWorklog(worklog)">
                            编辑
                          </el-button>
                          <el-button v-if="canEditWorklog(worklog)" link type="danger" size="small" @click="handleDeleteWorklog(worklog.id)">
                            删除
                          </el-button>
                        </div>
                      </div>
                      <div class="worklog-text">{{ worklog.description }}</div>
                    </div>
                  </div>
                  <div v-if="worklogs.length === 0" class="empty-placeholder">
                    暂无工作日志
                  </div>
                </div>
              </el-tab-pane>
            </el-tabs>
          </el-card>

          <!-- 评论和工作日志 -->
          <el-card shadow="never" class="content-card">
            <template #header>
              <div class="card-header-group">
                <div class="card-icon activity">
                  <el-icon><Clock /></el-icon>
                </div>
                <span class="card-title">活动记录</span>
              </div>
            </template>
            <el-timeline v-if="activities.length > 0" class="activity-timeline">
              <el-timeline-item
                v-for="activity in activities"
                :key="activity.id"
                :timestamp="formatTime(activity.created_at)"
                placement="top"
              >
                <div class="activity-content">
                  <span class="activity-user">{{ activity.user_name }}</span>
                  <template v-if="activity.details">
                    <span class="activity-details">{{ activity.details }}</span>
                  </template>
                  <template v-else>
                    <span class="activity-action">{{ activity.action }}</span>
                    <template v-if="activity.field">
                      <span class="activity-field">{{ activity.field }}</span>
                      <span v-if="activity.old_value" class="activity-old-value">{{ activity.old_value }}</span>
                      <el-icon v-if="activity.old_value"><ArrowRight /></el-icon>
                      <span v-if="activity.new_value" class="activity-new-value">{{ activity.new_value }}</span>
                    </template>
                  </template>
                </div>
              </el-timeline-item>
            </el-timeline>
            <div v-else class="empty-placeholder">暂无活动记录</div>
          </el-card>
        </el-col>

        <!-- 右侧：详细信息 -->
        <el-col :xs="24" :lg="8">
          <!-- 基本信息 -->
          <el-card shadow="never" class="info-card">
            <template #header>
              <div class="card-header-group">
                <div class="card-icon info">
                  <el-icon><InfoFilled /></el-icon>
                </div>
                <span class="card-title">详细信息</span>
              </div>
            </template>
            <div class="info-list">
              <div v-if="issue.parent_key" class="info-item">
                <span class="info-label">父工单</span>
                <el-link type="primary" @click="$router.push(`/issues/${issue.parent_key}`)">
                  {{ issue.parent_key }}
                </el-link>
              </div>
              <div v-if="issue.epic_key" class="info-item">
                <span class="info-label">Epic</span>
                <el-link type="primary" @click="$router.push(`/issues/${issue.epic_key}`)">
                  <el-icon><Link /></el-icon>
                  {{ issue.epic_key }}{{ issue.epic_title ? ' - ' + issue.epic_title : '' }}
                </el-link>
              </div>
              <div v-if="issue.merged_into_issue_key" class="info-item">
                <span class="info-label">已合并到</span>
                <el-link type="primary" @click="$router.push(`/issues/${issue.merged_into_issue_key}`)">
                  <el-icon><Link /></el-icon>
                  {{ issue.merged_into_issue_key }}
                </el-link>
              </div>
              <div v-if="issue.merged_from_issue_keys?.length" class="info-item">
                <span class="info-label">合并来源</span>
                <div class="merged-from-links">
                  <el-link
                    v-for="key in issue.merged_from_issue_keys"
                    :key="key"
                    type="primary"
                    @click="$router.push(`/issues/${key}`)"
                    style="margin-right: 8px;"
                  >
                    {{ key }}
                  </el-link>
                </div>
              </div>
              <div class="info-item">
                <span class="info-label">状态</span>
                <div class="status-badge sm" :class="issue.status">
                  <span class="status-dot"></span>
                  <span>{{ getStatusText(issue.status) }}</span>
                </div>
              </div>
              <div v-if="slaStatus" class="info-item">
                <span class="info-label">SLA 状态</span>
                <div class="sla-status-wrap">
                  <el-tag v-if="slaStatus.level === 'overdue'" type="danger" size="small" effect="dark">已超时</el-tag>
                  <el-tag v-else-if="slaStatus.level === 'due_soon'" type="warning" size="small" effect="dark">即将超时</el-tag>
                  <el-tag v-else type="success" size="small" effect="dark">进行中</el-tag>
                  <span class="sla-hint">{{ slaStatus.hint }}</span>
                </div>
              </div>
              <div class="info-item">
                <span class="info-label">优先级</span>
                <el-tag :type="getPriorityType(issue.priority)" size="small" effect="dark">{{ issue.priority }}</el-tag>
              </div>
              <div v-if="issue.resolution" class="info-item">
                <span class="info-label">解决结果</span>
                <el-tag size="small" type="success">{{ getResolutionText(issue.resolution) }}</el-tag>
              </div>
              <div class="info-item">
                <span class="info-label">项目</span>
                <el-link type="primary" @click="$router.push(`/issues?project_key=${issue.project_key}`)">
                  {{ issue.project_key }}
                </el-link>
              </div>
              <div class="info-item">
                <span class="info-label">类型</span>
                <span>{{ issue.issue_type?.display_name || '-' }}</span>
              </div>
              <div class="info-item">
                <span class="info-label">指派人</span>
                <div class="assignee-with-action">
                  <template v-if="!editingAssignee">
                    <template v-if="issue.assignee">
                      <div class="user-info">
                        <div class="mini-avatar">{{ issue.assignee.display_name?.charAt(0) || '?' }}</div>
                        <span>{{ issue.assignee.display_name }}</span>
                      </div>
                    </template>
                    <span v-else class="text-muted">未指派</span>
                    <el-button
                      v-if="issue.assignee?.id !== userStore.user?.id"
                      link
                      type="primary"
                      size="small"
                      @click="handleAssignToMe"
                    >
                      分配给我
                    </el-button>
                    <el-button
                      v-if="userStore.isProjectAdmin"
                      link
                      size="small"
                      @click="startEditAssignee"
                    >
                      <el-icon><Edit /></el-icon>
                    </el-button>
                  </template>
                  <template v-else>
                    <el-select
                      v-model="editAssigneeId"
                      placeholder="选择指派人"
                      filterable
                      clearable
                      size="small"
                      style="width: 160px;"
                      @change="handleAssigneeChange"
                    >
                      <el-option v-for="u in users" :key="u.id" :label="u.display_name" :value="u.id" />
                    </el-select>
                    <el-button link size="small" @click="editingAssignee = false">取消</el-button>
                  </template>
                </div>
              </div>
              <div class="info-item">
                <span class="info-label">创建者</span>
                <span>{{ issue.reporter?.display_name || '未知' }}</span>
              </div>
              <div class="info-item">
                <span class="info-label">创建时间</span>
                <span>{{ formatTime(issue.created_at) }}</span>
              </div>
              <div class="info-item">
                <span class="info-label">更新时间</span>
                <span>{{ formatTime(issue.updated_at) }}</span>
              </div>
              <div v-if="issue.due_date" class="info-item">
                <span class="info-label">截止时间</span>
                <span>{{ formatDate(issue.due_date) }}</span>
                <el-tag v-if="dueDateStatus === 'overdue'" type="danger" size="small" style="margin-left: 6px;">已超时</el-tag>
                <el-tag v-else-if="dueDateStatus === 'due_soon'" type="warning" size="small" style="margin-left: 6px;">即将超时</el-tag>
              </div>
              <div v-if="issue.planned_start_date" class="info-item">
                <span class="info-label">预计开始</span>
                <span>{{ formatDate(issue.planned_start_date) }}</span>
              </div>
              <div v-if="issue.planned_end_date" class="info-item">
                <span class="info-label">预计交付</span>
                <span>{{ formatDate(issue.planned_end_date) }}</span>
              </div>
              <div v-if="issue.actual_start_date" class="info-item">
                <span class="info-label">实际开始</span>
                <span>{{ formatTime(issue.actual_start_date) }}</span>
              </div>
              <div v-if="issue.actual_end_date" class="info-item">
                <span class="info-label">实际完成</span>
                <span>{{ formatTime(issue.actual_end_date) }}</span>
              </div>
            </div>
          </el-card>

          <!-- 关注人 -->
          <el-card shadow="never" class="info-card">
            <template #header>
              <div class="card-header-with-action">
                <div class="card-header-group">
                  <div class="card-icon watcher">
                    <el-icon><View /></el-icon>
                  </div>
                  <span class="card-title">关注人</span>
                  <span class="card-count">{{ watchers.length }}</span>
                </div>
                <div class="watcher-header-actions">
                  <el-button v-if="!isWatching" link type="primary" size="small" @click="handleWatchIssue">
                    <el-icon><View /></el-icon>
                    关注
                  </el-button>
                  <el-button v-else link type="danger" size="small" @click="handleUnwatchIssue">
                    取消关注
                  </el-button>
                  <el-button link type="primary" size="small" @click="showAddWatcherDialog">
                    <el-icon><Plus /></el-icon>
                  </el-button>
                </div>
              </div>
            </template>
            <div class="watcher-list">
              <div v-for="watcher in watchers" :key="watcher.id" class="watcher-item">
                <div class="watcher-info">
                  <div class="mini-avatar">{{ watcher.user?.display_name?.charAt(0) || '?' }}</div>
                  <span class="watcher-name">{{ watcher.user?.display_name || '未知用户' }}</span>
                </div>
                <el-button
                  v-if="canRemoveWatcher(watcher)"
                  link
                  type="danger"
                  size="small"
                  @click="handleRemoveWatcher(watcher.user_id)"
                >
                  移除
                </el-button>
              </div>
              <div v-if="watchers.length === 0" class="empty-placeholder sm">
                暂无关注人
              </div>
            </div>
          </el-card>
        </el-col>
      </el-row>
    </template>

    <!-- 添加关注人对话框 -->
    <el-dialog v-model="addWatcherDialogVisible" title="添加关注人" width="400px" destroy-on-close>
      <el-select
        v-model="selectedWatcherUserId"
        placeholder="请选择用户"
        style="width: 100%"
        filterable
      >
        <el-option
          v-for="u in availableWatcherUsers"
          :key="u.id"
          :label="u.display_name"
          :value="u.id"
        />
      </el-select>
      <template #footer>
        <el-button @click="addWatcherDialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="watcherLoading" :disabled="!selectedWatcherUserId" @click="handleAddWatcher">
          添加
        </el-button>
      </template>
    </el-dialog>

    <!-- 拒绝审批对话框 -->
    <el-dialog v-model="rejectDialogVisible" title="拒绝审批" width="450px" destroy-on-close>
      <el-form label-position="top">
        <el-form-item label="拒绝原因（必填）">
          <el-input v-model="rejectComment" type="textarea" :rows="3" placeholder="请输入拒绝原因" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="rejectDialogVisible = false">取消</el-button>
        <el-button type="danger" :loading="rejectLoading" :disabled="!rejectComment.trim()" @click="handleReject">
          确认拒绝
        </el-button>
      </template>
    </el-dialog>

    <!-- 编辑对话框 -->
    <el-dialog v-model="editDialogVisible" title="编辑工单" width="640px" destroy-on-close class="edit-dialog">
      <el-form ref="editFormRef" :model="editForm" :rules="editRules" label-position="top">
        <el-form-item label="标题" prop="title">
          <el-input v-model="editForm.title" maxlength="200" show-word-limit />
        </el-form-item>
        <el-row :gutter="20">
          <el-col :span="12">
            <el-form-item label="解决结果">
              <el-select v-model="editForm.resolution" placeholder="请选择" style="width: 100%" clearable>
                <el-option label="已解决" value="fixed" />
                <el-option label="不予修复" value="wont_fix" />
                <el-option label="重复工单" value="duplicate" />
                <el-option label="无法复现" value="cannot_reproduce" />
                <el-option label="按设计工作" value="works_as_designed" />
                <el-option label="信息不完整" value="incomplete" />
                <el-option label="已完成" value="done" />
              </el-select>
            </el-form-item>
          </el-col>
        </el-row>

        <!-- 所有字段由字段方案驱动 -->
        <div v-if="editFieldScheme.length > 0" class="edit-custom-fields">
          <el-row :gutter="20">
            <el-col
              v-for="item in editFieldScheme"
              :key="item.field_id"
              :span="getEditFieldColSpan(item)"
            >
              <el-form-item :required="item.is_required">
                <template #label>
                  <span>{{ item.field?.field_name }}</span>
                  <el-tooltip v-if="item.field?.description" :content="item.field?.description" placement="top">
                    <el-icon class="field-hint" style="margin-left: 4px; font-size: 14px; color: #c0c4cc; cursor: help;"><QuestionFilled /></el-icon>
                  </el-tooltip>
                </template>
                <!-- assignee 字段：包装"分配给我"按钮 -->
                <div v-if="item.field?.field_key === 'assignee'" style="display: flex; gap: 8px; width: 100%;">
                  <FieldRenderer
                    v-if="item.field && issue"
                    :field="item.field"
                    :scheme="item"
                    :project-key="issue.project_key"
                    v-model="editFieldValues[item.field_id]"
                    style="flex: 1;"
                  />
                  <el-button @click="assignToMeField(item.field_id)">分配给我</el-button>
                </div>
                <FieldRenderer
                  v-else-if="item.field && issue"
                  :field="item.field"
                  :scheme="item"
                  :project-key="issue.project_key"
                  v-model="editFieldValues[item.field_id]"
                />
              </el-form-item>
            </el-col>
          </el-row>
        </div>
      </el-form>
      <template #footer>
        <el-button @click="editDialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="editLoading" @click="submitEdit">
          <el-icon><Check /></el-icon>
          保存
        </el-button>
      </template>
    </el-dialog>

    <!-- 创建子任务对话框 -->
    <el-dialog v-model="createSubtaskDialogVisible" title="创建子任务" width="640px" destroy-on-close class="create-dialog">
      <el-form ref="createSubtaskFormRef" :model="createSubtaskForm" :rules="createSubtaskRules" label-position="top">
        <el-row :gutter="20">
          <el-col :span="12">
            <el-form-item label="项目" prop="project_key">
              <el-select v-model="createSubtaskForm.project_key" placeholder="请选择项目" style="width: 100%" @change="handleSubtaskProjectChange">
                <el-option v-for="p in projects" :key="p.project_key" :label="p.name" :value="p.project_key" />
              </el-select>
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="类型" prop="issue_type_id">
              <el-select v-model="createSubtaskForm.issue_type_id" placeholder="请选择类型" style="width: 100%" :disabled="!createSubtaskForm.project_key" @change="handleSubtaskIssueTypeChange">
                <el-option v-for="t in issueTypes" :key="t.id" :label="t.display_name" :value="t.id" />
              </el-select>
            </el-form-item>
          </el-col>
        </el-row>
        <el-form-item label="标题" prop="title">
          <el-input v-model="createSubtaskForm.title" placeholder="请输入子任务标题" maxlength="200" show-word-limit />
        </el-form-item>

        <!-- 所有字段由字段方案驱动 -->
        <div v-if="createSubtaskFieldScheme.length > 0" class="custom-fields-section">
          <el-row :gutter="20">
            <el-col
              v-for="item in createSubtaskFieldScheme"
              :key="item.field_id"
              :span="getEditFieldColSpan(item)"
            >
              <el-form-item :required="item.is_required">
                <template #label>
                  <span>{{ item.field?.field_name }}</span>
                  <el-tooltip v-if="item.field?.description" :content="item.field?.description" placement="top">
                    <el-icon class="field-hint" style="margin-left: 4px; font-size: 14px; color: #c0c4cc; cursor: help;"><QuestionFilled /></el-icon>
                  </el-tooltip>
                </template>
                <FieldRenderer
                  v-if="item.field"
                  :field="item.field"
                  :scheme="item"
                  :project-key="createSubtaskForm.project_key"
                  v-model="createSubtaskFieldValues[item.field_id]"
                />
              </el-form-item>
            </el-col>
          </el-row>
        </div>
      </el-form>
      <template #footer>
        <el-button @click="createSubtaskDialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="createSubtaskLoading" @click="submitCreateSubtask">
          <el-icon><Check /></el-icon>
          创建子任务
        </el-button>
      </template>
    </el-dialog>

    <!-- 工作流流程图对话框 -->
    <el-dialog v-model="diagramVisible" title="工单流程" width="860px" destroy-on-close>
      <div v-loading="diagramLoading" class="workflow-diagram">
        <div v-if="diagramLayout.nodes.length > 0" class="diagram-graph" :style="{ minHeight: diagramLayout.height + 'px', minWidth: diagramLayout.width + 'px' }">
          <!-- SVG 连线层 -->
          <svg class="diagram-edges" :width="diagramLayout.width" :height="diagramLayout.height">
            <defs>
              <marker id="arrowhead" markerWidth="8" markerHeight="6" refX="8" refY="3" orient="auto">
                <polygon points="0 0, 8 3, 0 6" fill="#c0c4cc" />
              </marker>
              <marker id="arrowhead-visited" markerWidth="8" markerHeight="6" refX="8" refY="3" orient="auto">
                <polygon points="0 0, 8 3, 0 6" fill="#67c23a" />
              </marker>
            </defs>
            <template v-for="edge in diagramLayout.edges" :key="`${edge.from}-${edge.to}`">
              <line
                :x1="edge.x1" :y1="edge.y1" :x2="edge.x2" :y2="edge.y2"
                :stroke="edge.visited ? '#67c23a' : '#c0c4cc'"
                stroke-width="2"
                :marker-end="edge.visited ? 'url(#arrowhead-visited)' : 'url(#arrowhead)'"
              />
              <text
                v-if="edge.label"
                :x="(edge.x1 + edge.x2) / 2"
                :y="(edge.y1 + edge.y2) / 2 - 6"
                text-anchor="middle"
                :fill="edge.visited ? '#67c23a' : '#909399'"
                font-size="11"
              >{{ edge.label }}</text>
            </template>
          </svg>
          <!-- 节点层 -->
          <div
            v-for="ln in diagramLayout.nodes"
            :key="ln.node.id"
            class="diagram-node"
            :class="getNodeDiagramClass(ln.node)"
            :style="{ left: ln.x + 'px', top: ln.y + 'px' }"
          >
            <div class="diagram-node-icon">
              <span v-if="ln.node.node_type === 'start'">▶</span>
              <span v-else-if="ln.node.node_type === 'end'">◉</span>
              <span v-else-if="ln.node.node_type === 'approval'">✓</span>
              <span v-else-if="ln.node.node_type === 'work'">⚙</span>
              <span v-else>●</span>
            </div>
            <div class="diagram-node-name">{{ ln.node.name }}</div>
            <div class="diagram-node-type">{{ getNodeTypeText(ln.node.node_type) }}</div>
            <div v-if="getNodeDiagramClass(ln.node).visited" class="diagram-node-check">✓</div>
          </div>
        </div>
        <div v-else-if="!diagramLoading" class="diagram-empty">
          <el-empty description="暂无流程节点" :image-size="80" />
        </div>
      </div>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, computed, onMounted, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage, ElMessageBox, type FormInstance, type FormRules } from 'element-plus'
import {
  User, Clock, Edit, ArrowDown, ArrowUp, ArrowRight, Plus, Document, Bell,
  ChatLineRound, InfoFilled, View, Check, Link, Delete, Paperclip, Promotion, QuestionFilled
} from '@element-plus/icons-vue'
import {
  getIssueDetail, updateIssue, deleteIssue, createIssue,
  getIssueComments, addIssueComment, getIssueActivities, getIssueWatchers,
  getWorklogs, addWorklog, deleteWorklog,
  addIssueWatcher, removeIssueWatcher, getEpicIssues, getSubtasks,
} from '@/api/issue'
import { getAlertList } from '@/api/alert'
import type { Alert } from '@/types/alert'
import { listAttachments } from '@/api/attachment'
import { getWorkflowInstance, getWorkflowHistory, approveWorkflow, rejectWorkflow, completeWorkflow, getWorkflowNodes, getWorkflowEdges } from '@/api/workflow'
import type { WorkflowInstance, WorkflowHistory, WorkflowNode, WorkflowEdge } from '@/types/workflow'
import type { Attachment } from '@/types/attachment'
import AttachmentUpload from '@/components/attachment/AttachmentUpload.vue'
import AttachmentList from '@/components/attachment/AttachmentList.vue'
import { getAllUsers } from '@/api/user'
import { getPublicConfig } from '@/api/system'
import { getIssueFieldValues, getFieldScheme } from '@/api/field'
import { getAllProjects, getProjectIssueTypes } from '@/api/project'
import { useUserStore } from '@/stores/user'
import type { Issue, IssueComment, IssueActivity, IssueWatcher, IssueResolution, UpdateIssueRequest, Worklog, CreateWorklogRequest, CreateIssueRequest } from '@/types/issue'
import type { UserOption } from '@/types/user'
import type { FieldValue, FieldSchemeItem, FieldTypeValue } from '@/types/field'
import type { Project, ProjectIssueType } from '@/types/project'
import FieldRenderer from '@/components/field/FieldRenderer.vue'
import { isBuiltinField } from '@/types/field'
import { extractBuiltinFields, backfillBuiltinFields } from '@/utils/builtin-fields'
import dayjs from 'dayjs'

const route = useRoute()
const router = useRouter()
const userStore = useUserStore()

const loading = ref(false)
const issue = ref<Issue | null>(null)
const comments = ref<IssueComment[]>([])
const activities = ref<IssueActivity[]>([])
const watchers = ref<IssueWatcher[]>([])
const worklogs = ref<Worklog[]>([])
const customFields = ref<FieldValue[]>([])
const users = ref<UserOption[]>([])
const activeTab = ref('comments')
const epicIssues = ref<Issue[]>([])
const subtasks = ref<Issue[]>([])
const attachments = ref<Attachment[]>([])
const issueAlerts = ref<Alert[]>([])
const showAttachmentUpload = ref(false)

// 工作流相关
const workflowExpanded = ref(false)
const workflowInstance = ref<WorkflowInstance | null>(null)
const workflowHistoryList = ref<WorkflowHistory[]>([])
const approveComment = ref('')
const rejectComment = ref('')
const approveLoading = ref(false)
const rejectLoading = ref(false)
const rejectDialogVisible = ref(false)

// 创建子任务相关
const createSubtaskDialogVisible = ref(false)
const createSubtaskLoading = ref(false)
const createSubtaskFormRef = ref<FormInstance>()
const projects = ref<Project[]>([])
const issueTypes = ref<ProjectIssueType[]>([])
const createSubtaskFieldScheme = ref<FieldSchemeItem[]>([])
const createSubtaskFieldValues = ref<Record<number, any>>({})

interface CreateSubtaskFormData {
  project_key: string
  issue_type_id: number | undefined
  title: string
}

const createSubtaskForm = reactive<CreateSubtaskFormData>({
  project_key: '',
  issue_type_id: undefined,
  title: '',
})

const createSubtaskRules: FormRules = {
  project_key: [{ required: true, message: '请选择项目', trigger: 'change' }],
  issue_type_id: [{ required: true, message: '请选择类型', trigger: 'change' }],
  title: [{ required: true, message: '请输入标题', trigger: 'blur' }],
}

const newComment = ref('')
const commentLoading = ref(false)
const editDialogVisible = ref(false)
const editLoading = ref(false)
const editFormRef = ref<FormInstance>()
const editForm = reactive({
  title: '',
  resolution: undefined as IssueResolution | undefined,
})
const editRules: FormRules = {
  title: [{ required: true, message: '请输入标题', trigger: 'blur' }],
}
// 编辑用的字段方案和值
const editFieldScheme = ref<FieldSchemeItem[]>([])
const editFieldValues = ref<Record<number, any>>({})

const loadIssue = async () => {
  const key = route.params.key as string
  if (!key) return
  loading.value = true
  try {
    const { data } = await getIssueDetail(key, { _redirectOn404: true })
    issue.value = data.data
    // 先加载字段（需要 issue.value 中的 project_key 和 issue_type_id）
    await loadCustomFields(data.data.id)
    // 然后并行加载其他数据
    await Promise.all([
      loadComments(key),
      loadActivities(key),
      loadWatchers(key),
      loadWorklogs(key),
      loadEpicIssues(key),
      loadSubtasks(key),
      loadAttachments(key),
      loadWorkflowData(key),
      loadIssueAlerts(data.data.id),
    ])
  } catch (error: any) {
    console.error('Failed to load issue:', error)
    ElMessage.error('加载工单失败')
  } finally {
    loading.value = false
  }
}

const loadComments = async (key: string) => {
  try { const { data } = await getIssueComments(key); comments.value = data.data } catch (e) { console.error(e) }
}
const loadActivities = async (key: string) => {
  try { const { data } = await getIssueActivities(key); activities.value = data.data.items || [] } catch (e) { console.error(e) }
}
const loadWatchers = async (key: string) => {
  try { const { data } = await getIssueWatchers(key); watchers.value = data.data } catch (e) { console.error(e) }
}
const loadWorklogs = async (key: string) => {
  try { const { data } = await getWorklogs(key); worklogs.value = data.data } catch (e) { console.error(e) }
}
const loadEpicIssues = async (key: string) => {
  // 只有当工单类型是 Epic 时才加载
  if (issue.value?.issue_type?.name?.toLowerCase() !== 'epic') {
    epicIssues.value = []
    return
  }
  try {
    const { data } = await getEpicIssues(key)
    epicIssues.value = data.data || []
  } catch (e) {
    console.error('Failed to load epic issues:', e)
    epicIssues.value = []
  }
}

const loadSubtasks = async (key: string) => {
  try {
    const { data } = await getSubtasks(key)
    subtasks.value = data.data || []
  } catch (e) {
    console.error('Failed to load subtasks:', e)
    subtasks.value = []
  }
}

const loadAttachments = async (key: string) => {
  try {
    const response = await listAttachments(key)
    attachments.value = (response as any).data.data || []
  } catch (e) {
    console.error('Failed to load attachments:', e)
    attachments.value = []
  }
}

// 关联告警加载
const loadIssueAlerts = async (issueId: number) => {
  try {
    const { data } = await getAlertList({ issue_id: issueId, page_size: 50 })
    issueAlerts.value = data.data.items || []
  } catch (e) {
    console.error('Failed to load issue alerts:', e)
    issueAlerts.value = []
  }
}

// 工作流数据加载
const loadWorkflowData = async (key: string) => {
  // 加载工作流实例，404 表示没有工作流，静默处理
  try {
    const { data } = await getWorkflowInstance(key)
    workflowInstance.value = (data as any).data
  } catch (e: any) {
    // 404 或其他错误表示没有关联工作流，不显示卡片
    workflowInstance.value = null
    workflowHistoryList.value = []
    return
  }

  // 加载工作流节点和边（用于显示下一步节点名称）
  try {
    const workflowId = workflowInstance.value!.workflow_id
    const [nodesRes, edgesRes] = await Promise.all([
      getWorkflowNodes(workflowId),
      getWorkflowEdges(workflowId),
    ])
    diagramNodes.value = (nodesRes.data as any).data || []
    diagramEdges.value = (edgesRes.data as any).data || []
  } catch (e: any) {
    // 静默处理
  }

  // 只有工作流实例存在时才加载流转历史
  try {
    const { data } = await getWorkflowHistory(key)
    workflowHistoryList.value = (data as any).data || []
  } catch (e: any) {
    workflowHistoryList.value = []
  }
}

// 判断当前用户是否是审批人（包括系统管理员和项目管理员）
const isCurrentUserApprover = computed(() => {
  if (!workflowInstance.value || !userStore.user) return false
  // 系统管理员或项目管理员可以审批任何节点（与后端 isAdminOrProjectLead 对齐）
  if (userStore.isAdmin || userStore.isProjectAdmin) return true
  // 检查是否在审批人列表中
  const approvals = workflowInstance.value.approvals || []
  return approvals.some(
    a => a.approver_id === userStore.user!.id && a.status === 'pending'
  )
})

// 判断当前节点是否是工作节点（可完成）
const isWorkNode = computed(() => {
  if (!workflowInstance.value || !isWorkflowOperable.value) return false
  const nodeType = workflowInstance.value.current_node?.node_type
  return nodeType === 'work' || nodeType === 'start'
})

// 获取当前工作节点的所有出边（带条件和目标节点名称）
interface OutgoingAction {
  conditionExpr: string  // 条件表达式（如 approved, rejected, 自定义条件名称）
  label: string          // 显示标签
  targetNodeName: string // 目标节点名称
}

const workNodeOutgoingActions = computed<OutgoingAction[]>(() => {
  if (!workflowInstance.value || !isWorkNode.value) return []
  const currentNodeId = workflowInstance.value.current_node_id
  if (!currentNodeId) return []

  const outEdges = diagramEdges.value.filter(e => e.source_node_id === currentNodeId)

  // 预设条件的显示名称
  const presetLabels: Record<string, string> = {
    approved: '通过',
    rejected: '退回',
    confirmed: '确认完成',
    continue: '继续处理',
  }

  return outEdges
    .filter(e => e.condition_expr) // 只取有条件的边
    .map(e => {
      const targetNode = diagramNodes.value.find(n => n.id === e.target_node_id)
      return {
        conditionExpr: e.condition_expr,
        label: presetLabels[e.condition_expr] || e.condition_expr,
        targetNodeName: targetNode?.name || '',
      }
    })
})

// 判断当前工作节点是否有多个条件分支
const workNodeHasBranching = computed(() => {
  return workNodeOutgoingActions.value.length > 0
})

// 获取下一个节点名称（无条件边的目标，用于单一流转）
const nextNodeName = computed(() => {
  if (!workflowInstance.value) return ''
  const currentNodeId = workflowInstance.value.current_node_id
  if (!currentNodeId) return ''

  // 优先找无条件边
  const outEdges = diagramEdges.value.filter(e => e.source_node_id === currentNodeId)
  const unconditionalEdge = outEdges.find(e => !e.condition_expr)
  const targetEdge = unconditionalEdge || outEdges[0]
  if (!targetEdge) return ''

  const targetNode = diagramNodes.value.find(n => n.id === targetEdge.target_node_id)
  return targetNode?.name || ''
})

// 工作流是否处于可操作状态（active 和 reviewing 都可以操作）
const isWorkflowOperable = computed(() => {
  const status = workflowInstance.value?.status
  return status === 'active' || status === 'reviewing'
})

// 当前用户是否可以操作工作流
const canOperateWorkflow = computed(() => {
  return isCurrentUserApprover.value || isWorkNode.value
})

// 工作流快捷按钮文本
const workflowActionBtnText = computed(() => {
  if (!workflowInstance.value) return '工作流'
  const status = workflowInstance.value.status
  if (status === 'completed' || status === 'cancelled') {
    const statusMap: Record<string, string> = { completed: '已完成', cancelled: '已取消' }
    return statusMap[status] || status
  }
  // active 和 reviewing 都显示当前节点名
  const nodeName = workflowInstance.value.current_node?.name || '当前节点'
  return nodeName
})

// 工作流下拉菜单命令处理
const handleWorkflowCommand = (command: string) => {
  // 处理动态条件命令：complete-condition:xxx
  if (command.startsWith('complete-condition:')) {
    const condition = command.replace('complete-condition:', '')
    handleQuickCompleteWithResult(condition)
    return
  }

  switch (command) {
    case 'approve':
      handleQuickApprove()
      break
    case 'reject':
      showRejectDialog()
      break
    case 'complete':
      handleQuickComplete()
      break
    case 'view-workflow':
      showWorkflowDiagram()
      break
  }
}

// 快捷审批通过（从下拉菜单触发，弹确认框）
const handleQuickApprove = async () => {
  try {
    await ElMessageBox.confirm('确认审批通过？', '审批确认', {
      confirmButtonText: '通过',
      cancelButtonText: '取消',
      type: 'success',
    })
    handleApprove()
  } catch {
    // 用户取消
  }
}

// 快捷完成节点（从下拉菜单触发，弹确认框）
const handleQuickComplete = async () => {
  try {
    const confirmMsg = nextNodeName.value
      ? `确认流转至「${nextNodeName.value}」？`
      : '确认完成当前节点？'
    await ElMessageBox.confirm(confirmMsg, '流转确认', {
      confirmButtonText: '确认',
      cancelButtonText: '取消',
      type: 'info',
    })
    handleComplete()
  } catch {
    // 用户取消
  }
}

// 快捷完成节点（带结果，从下拉菜单触发）
const handleQuickCompleteWithResult = async (result: string) => {
  // 从出边动作中找到对应的信息
  const action = workNodeOutgoingActions.value.find(a => a.conditionExpr === result)
  const actionLabel = action?.label || result
  const targetName = action?.targetNodeName || ''
  const confirmMsg = targetName
    ? `确认「${actionLabel}」并流转至「${targetName}」？`
    : `确认「${actionLabel}」？`
  try {
    await ElMessageBox.confirm(confirmMsg, '操作确认', {
      confirmButtonText: '确认',
      cancelButtonText: '取消',
      type: result === 'rejected' ? 'warning' : 'success',
    })
    handleCompleteWithResult(result)
  } catch {
    // 用户取消
  }
}

const completeLoading = ref(false)
const completeComment = ref('')

// 完成工作节点（默认流转）
const handleComplete = async () => {
  if (!issue.value) return
  completeLoading.value = true
  try {
    await completeWorkflow(issue.value.issue_key, { comment: completeComment.value || undefined })
    ElMessage.success('工作节点已完成')
    completeComment.value = ''
    await loadIssue()
  } catch (error: any) {
    console.error('Failed to complete:', error)
    ElMessage.error(error.response?.data?.message || '完成操作失败')
  } finally {
    completeLoading.value = false
  }
}

// 完成工作节点（带结果：任意条件）
const handleCompleteWithResult = async (result: string) => {
  if (!issue.value) return
  completeLoading.value = true
  try {
    await completeWorkflow(issue.value.issue_key, {
      comment: completeComment.value || undefined,
      result: result,
    })
    // 从出边动作中找到对应的标签
    const action = workNodeOutgoingActions.value.find(a => a.conditionExpr === result)
    const actionLabel = action?.label || result
    ElMessage.success(`已执行: ${actionLabel}`)
    completeComment.value = ''
    await loadIssue()
  } catch (error: any) {
    console.error('Failed to complete with result:', error)
    ElMessage.error(error.response?.data?.message || '操作失败')
  } finally {
    completeLoading.value = false
  }
}

// 审批通过
const handleApprove = async () => {
  if (!issue.value) return
  approveLoading.value = true
  try {
    await approveWorkflow(issue.value.issue_key, { comment: approveComment.value || undefined })
    ElMessage.success('审批通过')
    approveComment.value = ''
    await loadIssue() // 刷新工单（状态已由工作流联动更新）
  } catch (error: any) {
    console.error('Failed to approve:', error)
    ElMessage.error(error.response?.data?.message || '审批操作失败')
  } finally {
    approveLoading.value = false
  }
}

// 打开拒绝对话框
const showRejectDialog = () => {
  rejectComment.value = ''
  rejectDialogVisible.value = true
}

// 审批拒绝
const handleReject = async () => {
  if (!issue.value || !rejectComment.value.trim()) {
    ElMessage.warning('请填写拒绝原因')
    return
  }
  rejectLoading.value = true
  try {
    await rejectWorkflow(issue.value.issue_key, { comment: rejectComment.value })
    ElMessage.success('已拒绝')
    rejectComment.value = ''
    rejectDialogVisible.value = false
    await loadIssue() // 刷新工单（状态已由工作流联动更新）
  } catch (error: any) {
    console.error('Failed to reject:', error)
    ElMessage.error(error.response?.data?.message || '拒绝操作失败')
  } finally {
    rejectLoading.value = false
  }
}

const getWorkflowStatusText = (status: string) => {
  const map: Record<string, string> = {
    active: '进行中',
    completed: '已完成',
    cancelled: '已取消',
    reviewing: '验收中',
  }
  return map[status] || status
}

const getWorkflowStatusType = (status: string): 'success' | 'info' | 'warning' | 'danger' => {
  const map: Record<string, 'success' | 'info' | 'warning' | 'danger'> = {
    active: 'warning',
    completed: 'success',
    cancelled: 'info',
    reviewing: 'warning',
  }
  return map[status] || 'info'
}

const getApprovalStatusText = (status: string) => {
  const map: Record<string, string> = {
    pending: '待审批',
    approved: '已通过',
    rejected: '已拒绝',
  }
  return map[status] || status
}

const getApprovalStatusType = (status: string): 'success' | 'info' | 'warning' | 'danger' => {
  const map: Record<string, 'success' | 'info' | 'warning' | 'danger'> = {
    pending: 'warning',
    approved: 'success',
    rejected: 'danger',
  }
  return map[status] || 'info'
}

const getHistoryActionText = (action: string) => {
  const map: Record<string, string> = {
    start: '启动工作流',
    approve: '审批通过',
    reject: '审批拒绝',
    forward: '流转到下一节点',
    advance: '流转到下一节点',
    complete: '工作流完成',
    cancel: '工作流取消',
  }
  return map[action] || action
}

// 判断字段值是否已设置（支持数组类型）
const isFieldValueSet = (field: FieldValue): boolean => {
  if (field.value === null || field.value === undefined || field.value === '') return false
  if (Array.isArray(field.value) && field.value.length === 0) return false
  return true
}

const loadCustomFields = async (issueId: number) => {
  if (!issue.value) return
  try {
    // 获取该工单类型配置的字段方案
    const schemeRes = await getFieldScheme(issue.value.project_key, issue.value.issue_type_id)
    const scheme = schemeRes.data.data || []

    // 只获取详情页可见的字段，排除内置字段（已在右侧硬编码展示）
    const visibleScheme = scheme.filter((s: FieldSchemeItem) => s.is_visible_detail && !isBuiltinField(s.field?.field_key || ''))

    // 获取已保存的字段值
    const valuesRes = await getIssueFieldValues(issueId)
    const savedValues = valuesRes.data.data || []

    // 合并：用字段方案作为基础，填充已保存的值
    const valueMap = new Map(savedValues.map((v: FieldValue) => [v.field_id, v]))

    customFields.value = visibleScheme.map((item: FieldSchemeItem) => {
      const savedValue = valueMap.get(item.field_id)
      return {
        field_id: item.field_id,
        field_key: item.field?.field_key || '',
        field_name: item.field?.field_name || '',
        field_type: (item.field?.field_type || 'text') as FieldTypeValue,
        value: savedValue?.value ?? null,
        display_value: savedValue?.display_value || ''
      }
    })
  } catch (e) {
    console.error('Failed to load custom fields:', e)
  }
}

const submitComment = async () => {
  if (!issue.value || !newComment.value.trim()) return
  commentLoading.value = true
  try {
    await addIssueComment(issue.value.issue_key, { content: newComment.value })
    ElMessage.success('评论成功')
    newComment.value = ''
    loadComments(issue.value.issue_key)
  } catch (error) { console.error(error) }
  finally { commentLoading.value = false }
}

// 通过 field_id 设置"分配给我"
const assignToMeField = (fieldId: number) => {
  if (userStore.user) {
    editFieldValues.value[fieldId] = userStore.user.id
  }
}

// 编辑/子任务表单的字段列宽
const getEditFieldColSpan = (item: FieldSchemeItem): number => {
  const fieldType = item.field?.field_type || ''
  if (fieldType === 'textarea' || fieldType === 'epic_link') return 24
  return 12
}

const handleAssignToMe = async () => {
  if (!issue.value || !userStore.user) return
  try {
    await updateIssue(issue.value.issue_key, {
      assignee_id: userStore.user.id
    })
    ElMessage.success('已分配给您')
    loadIssue()
  } catch (error) {
    console.error('Failed to assign to me:', error)
    ElMessage.error('分配失败')
  }
}

// 指派人内联编辑（项目管理员）
const editingAssignee = ref(false)
const editAssigneeId = ref<number | undefined>(undefined)

const startEditAssignee = async () => {
  if (users.value.length === 0) {
    try { const { data } = await getAllUsers(); users.value = data.data } catch (e) { console.error(e) }
  }
  editAssigneeId.value = issue.value?.assignee?.id
  editingAssignee.value = true
}

const handleAssigneeChange = async (userId: number | undefined) => {
  if (!issue.value) return
  try {
    await updateIssue(issue.value.issue_key, { assignee_id: userId || 0 })
    ElMessage.success('指派人已更新')
    editingAssignee.value = false
    loadIssue()
  } catch (error) {
    console.error('Failed to update assignee:', error)
    ElMessage.error('更新指派人失败')
  }
}

const handleDelete = async () => {
  if (!issue.value) return
  try {
    await ElMessageBox.confirm(
      `确定要删除工单 ${issue.value.issue_key} 吗？删除后无法恢复。`,
      '删除工单',
      {
        confirmButtonText: '确定删除',
        cancelButtonText: '取消',
        type: 'warning',
        confirmButtonClass: 'el-button--danger',
      }
    )
    await deleteIssue(issue.value.issue_key)
    ElMessage.success('工单已删除')
    router.push('/issues')
  } catch (error) {
    if (error !== 'cancel') {
      console.error('Failed to delete issue:', error)
      ElMessage.error('删除工单失败')
    }
  }
}

const handleCreateSubtask = async () => {
  if (!issue.value) return

  // 加载项目列表和用户列表
  try {
    const [projectsRes, usersRes] = await Promise.all([getAllProjects(), getAllUsers()])
    projects.value = projectsRes.data.data
    users.value = usersRes.data.data
  } catch (error) {
    console.error('Failed to load options:', error)
    ElMessage.error('加载选项失败')
    return
  }

  // 初始化表单，自动填充父工单的项目
  Object.assign(createSubtaskForm, {
    project_key: issue.value.project_key,
    issue_type_id: undefined,
    title: '',
  })
  createSubtaskFieldScheme.value = []
  createSubtaskFieldValues.value = {}

  // 加载项目的工单类型
  await handleSubtaskProjectChange(issue.value.project_key)

  createSubtaskDialogVisible.value = true
}

const handleSubtaskProjectChange = async (projectKey: string) => {
  createSubtaskForm.issue_type_id = undefined
  createSubtaskFieldScheme.value = []
  createSubtaskFieldValues.value = {}
  if (!projectKey) {
    issueTypes.value = []
    return
  }
  try {
    const { data } = await getProjectIssueTypes(projectKey)
    issueTypes.value = data.data
  } catch (error) {
    console.error('Failed to load issue types:', error)
  }
}

const handleSubtaskIssueTypeChange = async (issueTypeId: number) => {
  createSubtaskFieldScheme.value = []
  createSubtaskFieldValues.value = {}
  if (!issueTypeId || !createSubtaskForm.project_key) return
  try {
    const { data } = await getFieldScheme(createSubtaskForm.project_key, issueTypeId)
    const schemeItems = data.data || []
    createSubtaskFieldScheme.value = schemeItems.filter(item => item.is_visible_create)
    const arrayFieldTypes = ['multiselect', 'label', 'component']
    createSubtaskFieldScheme.value.forEach(item => {
      const fieldType = item.field?.field_type || ''
      if (arrayFieldTypes.includes(fieldType)) {
        if (item.default_value) {
          try {
            const parsed = JSON.parse(item.default_value)
            createSubtaskFieldValues.value[item.field_id] = Array.isArray(parsed) ? parsed : []
          } catch {
            createSubtaskFieldValues.value[item.field_id] = []
          }
        } else {
          createSubtaskFieldValues.value[item.field_id] = []
        }
      } else if (item.default_value) {
        createSubtaskFieldValues.value[item.field_id] = item.default_value
      }
    })
  } catch (error) {
    console.error('Failed to load field scheme:', error)
  }
}

const submitCreateSubtask = async () => {
  if (!createSubtaskFormRef.value || !issue.value) return
  await createSubtaskFormRef.value.validate(async (valid) => {
    if (!valid) return
    createSubtaskLoading.value = true
    try {
      // 校验必填字段
      for (const item of createSubtaskFieldScheme.value) {
        if (item.is_required) {
          const val = createSubtaskFieldValues.value[item.field_id]
          const isEmpty = val === undefined || val === null || val === '' || (Array.isArray(val) && val.length === 0)
          if (isEmpty) {
            ElMessage.error(`请填写 ${item.field?.field_name}`)
            createSubtaskLoading.value = false
            return
          }
        }
      }

      // 分离内置字段和扩展字段
      const { builtinValues, customFields } = extractBuiltinFields(createSubtaskFieldScheme.value, createSubtaskFieldValues.value)

      const requestData: CreateIssueRequest = {
        project_key: createSubtaskForm.project_key,
        issue_type_id: createSubtaskForm.issue_type_id!,
        title: createSubtaskForm.title,
        description: builtinValues.description || '',
        priority: builtinValues.priority || 'P2',
        assignee_id: builtinValues.assignee_id || undefined,
        planned_start_date: builtinValues.planned_start_date || undefined,
        planned_end_date: builtinValues.planned_end_date || undefined,
        epic_id: builtinValues.epic_id || undefined,
        parent_id: issue.value!.id,
        custom_fields: customFields.length > 0 ? customFields : undefined,
      }

      await createIssue(requestData)
      ElMessage.success('子任务创建成功')
      createSubtaskDialogVisible.value = false
      await loadSubtasks(issue.value!.issue_key)
    } catch (error) {
      console.error('Failed to create subtask:', error)
      ElMessage.error('创建子任务失败')
    } finally {
      createSubtaskLoading.value = false
    }
  })
}

const handleEdit = async () => {
  if (!issue.value) return
  try { const { data } = await getAllUsers(); users.value = data.data } catch (e) { console.error(e) }

  Object.assign(editForm, {
    title: issue.value.title,
    resolution: issue.value.resolution || undefined,
  })

  // 加载编辑用的字段方案
  try {
    const schemeRes = await getFieldScheme(issue.value.project_key, issue.value.issue_type_id)
    const scheme = schemeRes.data.data || []
    editFieldScheme.value = scheme.filter((s: FieldSchemeItem) => s.is_visible_edit)

    // 初始化字段值
    const arrayFieldTypes = ['multiselect', 'label', 'component']
    editFieldValues.value = {}
    // 先为所有数组类型字段初始化空数组
    editFieldScheme.value.forEach(item => {
      if (item.field && arrayFieldTypes.includes(item.field.field_type)) {
        editFieldValues.value[item.field_id] = []
      }
    })

    // 回填内置字段值（从 issue 对象读取）
    backfillBuiltinFields(issue.value, editFieldScheme.value, editFieldValues.value)

    // 加载 EAV 保存的扩展字段值
    const valuesRes = await getIssueFieldValues(issue.value.id)
    const savedValues = valuesRes.data.data || []
    savedValues.forEach((v: FieldValue) => {
      // 只回填非内置字段（内置字段已从 issue 对象回填）
      if (!isBuiltinField(v.field_key)) {
        editFieldValues.value[v.field_id] = v.value
      }
    })
  } catch (e) { console.error('Failed to load field scheme for edit:', e) }

  editDialogVisible.value = true
}

const submitEdit = async () => {
  if (!editFormRef.value || !issue.value) return
  await editFormRef.value.validate(async (valid) => {
    if (!valid) return
    editLoading.value = true
    try {
      // 校验必填字段
      for (const item of editFieldScheme.value) {
        if (item.is_required) {
          const val = editFieldValues.value[item.field_id]
          const isEmpty = val === undefined || val === null || val === '' || (Array.isArray(val) && val.length === 0)
          if (isEmpty) {
            ElMessage.error(`请填写 ${item.field?.field_name}`)
            editLoading.value = false
            return
          }
        }
      }

      // 分离内置字段和扩展字段
      const { builtinValues, customFields } = extractBuiltinFields(editFieldScheme.value, editFieldValues.value)

      const updateData: UpdateIssueRequest = {
        title: editForm.title,
        resolution: editForm.resolution,
        description: builtinValues.description || '',
        priority: builtinValues.priority || undefined,
        assignee_id: builtinValues.assignee_id || undefined,
        planned_start_date: builtinValues.planned_start_date || undefined,
        planned_end_date: builtinValues.planned_end_date || undefined,
        epic_id: builtinValues.epic_id || undefined,
        custom_fields: customFields.length > 0 ? customFields : undefined,
      }
      await updateIssue(issue.value!.issue_key, updateData)
      ElMessage.success('更新成功')
      editDialogVisible.value = false
      loadIssue()
    } catch (error) { console.error(error) }
    finally { editLoading.value = false }
  })
}

// Watcher related
const addWatcherDialogVisible = ref(false)
const selectedWatcherUserId = ref<number | undefined>(undefined)
const watcherLoading = ref(false)

const isWatching = computed(() => {
  if (!userStore.user) return false
  return watchers.value.some(w => w.user_id === userStore.user?.id)
})

const availableWatcherUsers = computed(() => {
  const watcherUserIds = new Set(watchers.value.map(w => w.user_id))
  return users.value.filter(u => !watcherUserIds.has(u.id))
})

const canRemoveWatcher = (watcher: IssueWatcher) => {
  // 可以移除自己，或者有管理权限
  return watcher.user_id === userStore.user?.id
}

const showAddWatcherDialog = async () => {
  if (users.value.length === 0) {
    try {
      const { data } = await getAllUsers()
      users.value = data.data
    } catch (error) {
      console.error(error)
      ElMessage.error('加载用户列表失败')
      return
    }
  }
  selectedWatcherUserId.value = undefined
  addWatcherDialogVisible.value = true
}

const handleAddWatcher = async () => {
  if (!issue.value || !selectedWatcherUserId.value) return
  watcherLoading.value = true
  try {
    await addIssueWatcher(issue.value.issue_key, selectedWatcherUserId.value)
    ElMessage.success('添加关注人成功')
    addWatcherDialogVisible.value = false
    loadWatchers(issue.value.issue_key)
  } catch (error: any) {
    console.error(error)
    if (error.response?.data?.message?.includes('已经关注')) {
      ElMessage.warning('该用户已经关注此工单')
    } else {
      ElMessage.error('添加关注人失败')
    }
  } finally {
    watcherLoading.value = false
  }
}

const handleRemoveWatcher = async (userId: number) => {
  if (!issue.value) return
  try {
    await ElMessageBox.confirm('确定要移除此关注人吗？', '提示', {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      type: 'warning',
    })
    await removeIssueWatcher(issue.value.issue_key, userId)
    ElMessage.success('移除成功')
    loadWatchers(issue.value.issue_key)
  } catch (error) {
    if (error !== 'cancel') {
      console.error(error)
      ElMessage.error('移除失败')
    }
  }
}

const handleWatchIssue = async () => {
  if (!issue.value || !userStore.user) return
  try {
    await addIssueWatcher(issue.value.issue_key, userStore.user.id)
    ElMessage.success('关注成功')
    loadWatchers(issue.value.issue_key)
  } catch (error: any) {
    console.error(error)
    if (error.response?.data?.message?.includes('已经关注')) {
      ElMessage.warning('您已经关注此工单')
    } else {
      ElMessage.error('关注失败')
    }
  }
}

const handleUnwatchIssue = async () => {
  if (!issue.value || !userStore.user) return
  try {
    await removeIssueWatcher(issue.value.issue_key, userStore.user.id)
    ElMessage.success('取消关注成功')
    loadWatchers(issue.value.issue_key)
  } catch (error) {
    console.error(error)
    ElMessage.error('取消关注失败')
  }
}

// Worklog related
const defaultWorkTypes = [
  { value: '开发', label: '开发' }, { value: '测试', label: '测试' },
  { value: '调试', label: '调试' }, { value: '文档', label: '文档' },
  { value: '故障排查', label: '故障排查' }, { value: '监控运维', label: '监控运维' },
  { value: '部署发布', label: '部署发布' }, { value: '配置变更', label: '配置变更' },
  { value: '巡检', label: '巡检' }, { value: '安全响应', label: '安全响应' },
  { value: '其他', label: '其他' },
]
const workTypeOptions = ref<{ value: string; label: string }[]>(defaultWorkTypes)

const loadWorkTypeOptions = async () => {
  try {
    const res = await getPublicConfig('worklog.work_types')
    const parsed = JSON.parse(res.data.data.config_value || '[]')
    if (Array.isArray(parsed) && parsed.length > 0) {
      workTypeOptions.value = parsed
    }
  } catch {
    // 加载失败时使用默认值，不影响使用
  }
}

const worklogLoading = ref(false)
const worklogForm = reactive<CreateWorklogRequest>({
  description: '',
  time_spent: '',
  worked_at: new Date().toISOString(),
  work_type: '',
})

const canSubmitWorklog = computed(() => {
  return worklogForm.description.trim() && worklogForm.time_spent.trim() && worklogForm.worked_at
})

const totalTimeSpent = computed(() => {
  return worklogs.value.reduce((sum, w) => sum + w.time_spent_sec, 0)
})

const canEditWorklog = (worklog: Worklog) => {
  return worklog.user_id === userStore.user?.id
}

const submitWorklog = async () => {
  if (!issue.value || !canSubmitWorklog.value) return
  worklogLoading.value = true
  try {
    const workedAtDate = new Date(worklogForm.worked_at)

    await addWorklog(issue.value.issue_key, {
      description: worklogForm.description,
      time_spent: worklogForm.time_spent,
      worked_at: workedAtDate.toISOString(),
      work_type: worklogForm.work_type || undefined,
    })
    ElMessage.success('工作日志添加成功')
    Object.assign(worklogForm, {
      description: '',
      time_spent: '',
      worked_at: new Date().toISOString(),
      work_type: '',
    })
    loadWorklogs(issue.value.issue_key)
  } catch (error) {
    console.error(error)
    ElMessage.error('添加工作日志失败')
  } finally {
    worklogLoading.value = false
  }
}

const handleEditWorklog = (_worklog: Worklog) => {
  ElMessage.info('编辑功能开发中')
}

const handleDeleteWorklog = async (worklogId: number) => {
  if (!issue.value) return
  try {
    await ElMessageBox.confirm('确定要删除这条工作日志吗？', '提示', {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      type: 'warning',
    })
    await deleteWorklog(issue.value.issue_key, worklogId)
    ElMessage.success('删除成功')
    loadWorklogs(issue.value.issue_key)
  } catch (error) {
    if (error !== 'cancel') {
      console.error(error)
      ElMessage.error('删除失败')
    }
  }
}

const formatTimeSpent = (seconds: number) => {
  const days = Math.floor(seconds / (8 * 3600))
  const hours = Math.floor((seconds % (8 * 3600)) / 3600)
  const minutes = Math.floor((seconds % 3600) / 60)

  const parts = []
  if (days > 0) parts.push(`${days}d`)
  if (hours > 0) parts.push(`${hours}h`)
  if (minutes > 0) parts.push(`${minutes}m`)

  return parts.length > 0 ? parts.join(' ') : '0m'
}

type TagType = 'primary' | 'success' | 'warning' | 'info' | 'danger'
const getPriorityType = (priority: string): TagType => {
  const map: Record<string, TagType> = { P0: 'danger', P1: 'warning', P2: 'info', P3: 'success' }
  return map[priority] || 'info'
}
const getStatusText = (status: string) => {
  const map: Record<string, string> = { open: '待处理', in_progress: '进行中', pending_review: '待确认', resolved: '已完成', closed: '已终止', reopened: '重新打开', merged: '已合并' }
  return map[status] || status
}

const getResolutionText = (resolution: string) => {
  const map: Record<string, string> = {
    fixed: '已解决',
    wont_fix: '不予修复',
    duplicate: '重复工单',
    cannot_reproduce: '无法复现',
    works_as_designed: '按设计工作',
    incomplete: '信息不完整',
    done: '已完成'
  }
  return map[resolution] || resolution
}

const formatTime = (time: string) => dayjs(time).format('YYYY-MM-DD HH:mm')
const formatDate = (date: string) => dayjs(date).format('YYYY-MM-DD')

// SLA 目标（分钟），与后端 slaTargets 保持一致
const SLA_TARGETS: Record<string, number> = { P0: 60, P1: 240, P2: 1440, P3: 4320 }

// SLA 超时状态计算（基于优先级 SLA 或 due_date）
const slaStatus = computed<{ level: 'overdue' | 'due_soon' | 'normal'; hint: string } | null>(() => {
  if (!issue.value) return null
  // 已完成/已关闭/已合并不显示
  if (['resolved', 'closed', 'merged'].includes(issue.value.status)) return null

  const now = dayjs()

  // 优先使用预计交付时间，没有则用优先级默认 SLA
  let deadline: ReturnType<typeof dayjs>
  let slaMinutes: number

  if (issue.value.planned_end_date) {
    deadline = dayjs(issue.value.planned_end_date).endOf('day')
    slaMinutes = deadline.diff(dayjs(issue.value.created_at), 'minute')
  } else {
    slaMinutes = SLA_TARGETS[issue.value.priority] || SLA_TARGETS.P2
    deadline = dayjs(issue.value.created_at).add(slaMinutes, 'minute')
  }

  const minutesLeft = deadline.diff(now, 'minute', true)
  const threshold = slaMinutes * 0.25 // 剩余不到 25% 时即将超时

  if (minutesLeft < 0) {
    const overMinutes = Math.abs(Math.round(minutesLeft))
    return { level: 'overdue', hint: `已超时 ${formatSLADuration(overMinutes)}` }
  }
  if (minutesLeft < threshold) {
    return { level: 'due_soon', hint: `剩余 ${formatSLADuration(Math.round(minutesLeft))}` }
  }
  return { level: 'normal', hint: `剩余 ${formatSLADuration(Math.round(minutesLeft))}` }
})

// 格式化 SLA 时长
const formatSLADuration = (minutes: number): string => {
  if (minutes < 60) return `${minutes}分钟`
  const hours = Math.floor(minutes / 60)
  const mins = minutes % 60
  if (hours < 24) {
    return mins > 0 ? `${hours}小时${mins}分钟` : `${hours}小时`
  }
  const days = Math.floor(hours / 24)
  const remainHours = hours % 24
  return remainHours > 0 ? `${days}天${remainHours}小时` : `${days}天`
}

// 截止时间超时状态（仅用于 due_date 字段旁的标签）
const dueDateStatus = computed<'overdue' | 'due_soon' | 'normal' | null>(() => {
  if (!issue.value?.due_date) return null
  if (['resolved', 'closed'].includes(issue.value.status)) return null
  const now = dayjs()
  const due = dayjs(issue.value.due_date)
  const hoursLeft = due.diff(now, 'hour', true)
  if (hoursLeft < 0) return 'overdue'
  if (hoursLeft < 24) return 'due_soon'
  return 'normal'
})

const getAlertSeverityType = (severity: string) => {
  const map: Record<string, TagType> = { critical: 'danger', warning: 'warning', info: 'info' }
  return map[severity] || 'info'
}
const getAlertSeverityText = (severity: string) => {
  const map: Record<string, string> = { critical: '严重', warning: '警告', info: '信息' }
  return map[severity] || severity
}
const getAlertStatusText = (status: string) => {
  const map: Record<string, string> = { firing: '触发中', resolved: '已解决' }
  return map[status] || status
}

onMounted(() => {
  loadIssue()
  loadWorkTypeOptions()
})

// 监听路由参数变化，当切换到不同的 Issue 时重新加载数据
watch(() => route.params.key, (newKey, oldKey) => {
  if (newKey && newKey !== oldKey) {
    loadIssue()
  }
})

// ============ 工作流流程图 ============

const diagramVisible = ref(false)
const diagramLoading = ref(false)
const diagramNodes = ref<WorkflowNode[]>([])
const diagramEdges = ref<WorkflowEdge[]>([])

// 计算流程图布局（支持分支）
interface LayoutNode { node: WorkflowNode; x: number; y: number }
interface LayoutEdge { from: number; to: number; x1: number; y1: number; x2: number; y2: number; visited: boolean; label: string }

const diagramLayout = computed(() => {
  const nodes = diagramNodes.value
  const edges = diagramEdges.value
  const result = { nodes: [] as LayoutNode[], edges: [] as LayoutEdge[], width: 0, height: 0 }

  if (nodes.length === 0) return result

  const nodeW = 110
  const nodeH = 80
  const gapX = 160
  const gapY = 120
  const padX = 40
  const padY = 30

  // BFS 分层
  const nodeMap = new Map<number, WorkflowNode>()
  nodes.forEach(n => nodeMap.set(n.id, n))

  const adjacency = new Map<number, number[]>()
  nodes.forEach(n => adjacency.set(n.id, []))
  edges.forEach(e => {
    if (adjacency.has(e.source_node_id)) {
      adjacency.get(e.source_node_id)!.push(e.target_node_id)
    }
  })

  // 找开始节点
  const startNode = nodes.find(n => n.node_type === 'start')
  if (!startNode) return result

  const levels = new Map<number, number>()
  const visited = new Set<number>()
  const queue: { id: number; level: number }[] = [{ id: startNode.id, level: 0 }]
  visited.add(startNode.id)

  while (queue.length > 0) {
    const { id, level } = queue.shift()!
    levels.set(id, Math.max(levels.get(id) || 0, level))
    for (const next of (adjacency.get(id) || [])) {
      if (!visited.has(next)) {
        visited.add(next)
        queue.push({ id: next, level: level + 1 })
      }
    }
  }

  // 未连接的节点
  const maxLevel = Math.max(...Array.from(levels.values()), 0)
  nodes.forEach(n => {
    if (!levels.has(n.id)) levels.set(n.id, maxLevel + 1)
  })

  // 按层分组
  const levelGroups = new Map<number, number[]>()
  for (const [nodeId, level] of levels) {
    if (!levelGroups.has(level)) levelGroups.set(level, [])
    levelGroups.get(level)!.push(nodeId)
  }

  // 计算节点位置（水平布局）
  const nodePositions = new Map<number, { x: number; y: number }>()
  const totalLevels = Math.max(...Array.from(levelGroups.keys())) + 1

  for (let lvl = 0; lvl < totalLevels; lvl++) {
    const group = levelGroups.get(lvl) || []
    const startY = padY + (group.length > 1 ? 0 : (gapY - nodeH) / 2)

    group.forEach((nodeId, idx) => {
      const x = padX + lvl * gapX
      const y = startY + idx * gapY
      nodePositions.set(nodeId, { x, y })
    })
  }

  // 生成布局节点
  for (const [nodeId, pos] of nodePositions) {
    const node = nodeMap.get(nodeId)
    if (node) {
      result.nodes.push({ node, x: pos.x, y: pos.y })
    }
  }

  // 生成布局边
  const presetConditionLabels: Record<string, string> = { approved: '通过', rejected: '拒绝', confirmed: '确认完成', continue: '继续处理' }
  for (const edge of edges) {
    const fromPos = nodePositions.get(edge.source_node_id)
    const toPos = nodePositions.get(edge.target_node_id)
    if (!fromPos || !toPos) continue

    result.edges.push({
      from: edge.source_node_id,
      to: edge.target_node_id,
      x1: fromPos.x + nodeW,
      y1: fromPos.y + nodeH / 2,
      x2: toPos.x,
      y2: toPos.y + nodeH / 2,
      visited: visitedEdgePairs.value.has(`${edge.source_node_id}-${edge.target_node_id}`),
      label: presetConditionLabels[edge.condition_expr] || edge.condition_expr || '',
    })
  }

  // 计算画布尺寸
  let maxX = 0, maxY = 0
  for (const pos of nodePositions.values()) {
    maxX = Math.max(maxX, pos.x + nodeW)
    maxY = Math.max(maxY, pos.y + nodeH)
  }
  result.width = maxX + padX
  result.height = maxY + padY

  return result
})

// 获取已访问的节点 ID 集合（从流转历史中提取）
const visitedNodeIds = computed(() => {
  const ids = new Set<number>()
  workflowHistoryList.value.forEach(h => {
    if (h.from_node_id) ids.add(h.from_node_id)
    if (h.to_node_id) ids.add(h.to_node_id)
  })
  return ids
})

// 获取已访问的边（from→to 对）
const visitedEdgePairs = computed(() => {
  const pairs = new Set<string>()
  workflowHistoryList.value.forEach(h => {
    if (h.from_node_id && h.to_node_id) {
      pairs.add(`${h.from_node_id}-${h.to_node_id}`)
    }
  })
  return pairs
})

// 判断节点的流程图样式类
const getNodeDiagramClass = (node: WorkflowNode) => {
  const currentNodeId = workflowInstance.value?.current_node_id
  const isCurrent = node.id === currentNodeId
  const isVisited = visitedNodeIds.value.has(node.id) && !isCurrent
  const instanceStatus = workflowInstance.value?.status

  const isOperable = instanceStatus === 'active' || instanceStatus === 'reviewing'

  return {
    current: isCurrent && isOperable,
    visited: isVisited || (isCurrent && instanceStatus === 'completed'),
    cancelled: instanceStatus === 'cancelled' && isCurrent,
    pending: !isCurrent && !isVisited,
    'node-start': node.node_type === 'start',
    'node-end': node.node_type === 'end',
    'node-approval': node.node_type === 'approval',
    'node-work': node.node_type === 'work',
  }
}

// 获取节点类型文本
const getNodeTypeText = (nodeType: string) => {
  const map: Record<string, string> = {
    start: '开始',
    end: '结束',
    approval: '审批',
    work: '工作',
    system: '系统',
  }
  return map[nodeType] || nodeType
}

// 显示工作流流程图
const showWorkflowDiagram = async () => {
  diagramVisible.value = true
  // 如果节点数据已加载（loadWorkflowData 中已加载），直接显示
  if (diagramNodes.value.length > 0) {
    return
  }
  // 否则重新加载
  diagramLoading.value = true
  try {
    const workflowId = workflowInstance.value!.workflow_id
    const [nodesRes, edgesRes] = await Promise.all([
      getWorkflowNodes(workflowId),
      getWorkflowEdges(workflowId),
    ])
    diagramNodes.value = (nodesRes.data as any).data || []
    diagramEdges.value = (edgesRes.data as any).data || []
  } catch (e) {
    console.error('Failed to load workflow diagram', e)
    ElMessage.error('加载流程图失败')
  } finally {
    diagramLoading.value = false
  }
}
</script>

<style scoped lang="scss">
.issue-detail-container {
  width: 100%;
}

// 头部
.issue-header {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  margin-bottom: 24px;
  padding: 28px 32px;
  background: #fff;
  border-radius: 12px;
  box-shadow: 0 2px 12px rgba(0, 0, 0, 0.04);

  .header-left { flex: 1; }

  .issue-breadcrumb { margin-bottom: 16px; }

  .issue-title {
    font-size: 24px;
    font-weight: 700;
    margin: 0 0 16px 0;
    color: #1f2937;
    line-height: 1.4;
  }

  .issue-meta {
    display: flex;
    align-items: center;
    gap: 14px;
    color: #6b7280;
    font-size: 14px;
    flex-wrap: wrap;

    .type-badge {
      display: inline-flex;
      align-items: center;
      gap: 6px;
      padding: 5px 14px;
      background: #3b82f6;
      color: #fff;
      border-radius: 6px;
      font-size: 13px;
      font-weight: 500;
      box-shadow: 0 2px 4px rgba(102, 126, 234, 0.2);

      .type-icon {
        font-size: 14px;
      }

      .type-text {
        line-height: 1;
      }
    }

    .meta-item {
      display: flex;
      align-items: center;
      gap: 5px;
    }
  }

  .header-actions {
    display: flex;
    gap: 10px;
    flex-shrink: 0;
    margin-left: 20px;

    .workflow-action-btn {
      font-weight: 500;
    }
  }
}

// 状态徽章
.status-badge {
  display: inline-flex;
  align-items: center;
  gap: 5px;
  padding: 3px 12px;
  border-radius: 20px;
  font-size: 13px;
  font-weight: 500;

  .status-dot {
    width: 7px;
    height: 7px;
    border-radius: 50%;
  }

  &.open { background: #f3f4f6; color: #6b7280; .status-dot { background: #9ca3af; } }
  &.in_progress { background: #fff7ed; color: #c2410c; .status-dot { background: #f59e0b; } }
  &.pending_review { background: #fffbeb; color: #b45309; .status-dot { background: #f59e0b; } }
  &.resolved { background: #ecfdf5; color: #059669; .status-dot { background: #10b981; } }
  &.closed { background: #f3f4f6; color: #6b7280; .status-dot { background: #9ca3af; } }
  &.reopened { background: #fef2f2; color: #dc2626; .status-dot { background: #ef4444; } }
  &.merged { background: #f3e8ff; color: #7c3aed; .status-dot { background: #8b5cf6; } }

  &.sm { padding: 2px 10px; font-size: 12px; .status-dot { width: 6px; height: 6px; } }
}

// 通用卡片
.content-card, .info-card {
  margin-bottom: 20px;
  border-radius: 12px;

  :deep(.el-card__header) {
    padding: 16px 20px;
    border-bottom: 1px solid #f0f0f0;
  }
}

.card-header-group {
  display: flex;
  align-items: center;
  gap: 10px;
}

.card-header-with-action {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.card-icon {
  width: 32px;
  height: 32px;
  border-radius: 8px;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 16px;
  color: #fff;

  &.desc { background: #3b82f6; }
  &.comment { background: #10b981; }
  &.activity { background: #3b82f6; }
  &.info { background: #ef4444; }
  &.watcher { background: #f59e0b; }
  &.custom { background: #8b5cf6; }
  &.epic { background: #3b82f6; }
  &.subtask { background: #f59e0b; }
  &.attachment { background: #f59e0b; }
  &.workflow { background: #8b5cf6; }
  &.alert { background: #ef4444; }
}

.card-title { font-size: 15px; font-weight: 600; color: #1f2937; }

.card-count {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  min-width: 20px;
  height: 20px;
  padding: 0 6px;
  background: #e5e7eb;
  border-radius: 10px;
  font-size: 12px;
  font-weight: 600;
  color: #4b5563;
}

.empty-placeholder {
  color: #9ca3af;
  text-align: center;
  padding: 32px 0;
  font-size: 14px;

  &.sm { padding: 20px 0; }
}

// 描述
.description-content {
  white-space: pre-wrap;
  line-height: 1.8;
  color: #374151;
  padding: 20px;
}

// 扩展字段
.custom-fields-card {
  .custom-fields-grid {
    display: grid;
    grid-template-columns: repeat(auto-fill, minmax(300px, 1fr));
    gap: 20px;
    padding: 20px;

    .field-item {
      .field-label {
        font-size: 12px;
        font-weight: 500;
        color: #6b7280;
        margin-bottom: 6px;
        text-transform: uppercase;
        letter-spacing: 0.5px;
      }

      .field-value {
        font-size: 14px;
        color: #1f2937;
        line-height: 1.6;
        word-break: break-word;

        .empty-value {
          color: #9ca3af;
          font-style: italic;
        }

        .epic-link {
          display: inline-flex;
          align-items: center;
          gap: 6px;
          color: #667eea;
          text-decoration: none;
          font-weight: 500;
          transition: all 0.2s;

          &:hover {
            color: #764ba2;
            text-decoration: underline;
          }

          .el-icon {
            font-size: 14px;
          }
        }
      }
    }
  }
}

// 评论区
.add-comment {
  padding: 20px;
  border-bottom: 1px solid #f0f0f0;

  .comment-actions {
    margin-top: 12px;
    text-align: right;
  }
}

.comment-list {
  .comment-item {
    display: flex;
    gap: 14px;
    padding: 16px 20px;
    border-bottom: 1px solid #f5f5f5;

    &:last-child { border-bottom: none; }

    .comment-avatar {
      width: 38px;
      height: 38px;
      border-radius: 10px;
      background: #3b82f6;
      color: #fff;
      display: flex;
      align-items: center;
      justify-content: center;
      font-size: 15px;
      font-weight: 600;
      flex-shrink: 0;

      &.system-avatar {
        background: #8b5cf6;
        font-size: 14px;
      }
    }

    .comment-body {
      flex: 1;

      .comment-header {
        display: flex;
        align-items: center;
        gap: 12px;
        margin-bottom: 8px;

        .comment-author { font-weight: 600; color: #1f2937; font-size: 14px; }
        .system-author { color: #7c3aed; }
        .comment-time { font-size: 12px; color: #9ca3af; }
      }

      .comment-text {
        color: #374151;
        line-height: 1.6;
        white-space: pre-wrap;
        font-size: 14px;
      }
    }

    &.system-comment {
      background: #faf5ff;
      border-radius: 8px;
      margin: 4px 0;
    }
  }
}

// 活动时间线
.activity-timeline {
  padding: 20px;

  .activity-content {
    font-size: 14px;

    .activity-user { font-weight: 600; color: #1f2937; }
    .activity-action { color: #606266; margin: 0 4px; }
    .activity-field { color: #909399; margin: 0 4px; }
    .activity-old-value { text-decoration: line-through; color: #f56c6c; margin: 0 4px; }
    .activity-new-value { color: #67c23a; margin: 0 4px; }
  }
}

// 右侧信息
.info-list {
  .info-item {
    display: grid;
    grid-template-columns: 80px 1fr;
    align-items: center;
    gap: 12px;
    padding: 12px 20px;
    border-bottom: 1px solid #f5f5f5;

    &:last-child { border-bottom: none; }

    .info-label {
      color: #9ca3af;
      font-size: 12px;
      font-weight: 500;
      text-align: left;
    }

    .user-info {
      display: flex;
      align-items: center;
      gap: 8px;
      color: #1f2937;
      font-size: 13px;
    }

    .assignee-with-action {
      display: flex;
      align-items: center;
      gap: 8px;
      justify-content: space-between;
      width: 100%;
    }

    > span:not(.info-label),
    > .el-tag,
    > .el-link,
    > .status-badge {
      color: #1f2937;
      font-size: 13px;
      justify-self: start;
    }

    .sla-status-wrap {
      display: flex;
      align-items: center;
      gap: 6px;
      flex-wrap: nowrap;
      min-width: 0;

      .sla-hint {
        font-size: 12px;
        color: #9ca3af;
        white-space: nowrap;
      }
    }
  }
}

.mini-avatar {
  width: 22px;
  height: 22px;
  border-radius: 6px;
  background: #3b82f6;
  color: #fff;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 10px;
  font-weight: 600;
  flex-shrink: 0;
}

.text-muted { color: #9ca3af; }

// 关注人列表
.watcher-list {
  .watcher-item {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 10px;
    padding: 10px 20px;
    border-bottom: 1px solid #f5f5f5;

    &:last-child { border-bottom: none; }

    .watcher-info {
      display: flex;
      align-items: center;
      gap: 10px;
      flex: 1;
    }

    .watcher-name { font-size: 14px; color: #374151; }
  }
}

.watcher-header-actions {
  display: flex;
  align-items: center;
  gap: 8px;
}

// 工作日志
.detail-tabs {
  :deep(.el-tabs__header) {
    margin-bottom: 20px;
  }

  .tab-label {
    display: flex;
    align-items: center;
    gap: 6px;

    .el-icon { font-size: 16px; }
  }

  .tab-badge {
    margin-left: 4px;
  }
}

.add-worklog {
  padding: 20px;
  background: #f9fafb;
  border-radius: 8px;
  margin-bottom: 20px;

  .form-hint {
    font-size: 12px;
    color: #9ca3af;
    margin-top: 4px;
  }
}

.worklog-list {
  .worklog-summary {
    display: flex;
    align-items: center;
    gap: 8px;
    padding: 12px 16px;
    background: #ecfdf5;
    border-radius: 8px;
    margin-bottom: 16px;
    font-size: 14px;
    font-weight: 500;
    color: #059669;

    .el-icon { font-size: 16px; }
  }

  .worklog-item {
    display: flex;
    gap: 12px;
    padding: 16px 0;
    border-bottom: 1px solid #f0f0f0;

    &:last-child { border-bottom: none; }

    .worklog-avatar {
      width: 36px;
      height: 36px;
      border-radius: 8px;
      background: #3b82f6;
      color: #fff;
      display: flex;
      align-items: center;
      justify-content: center;
      font-size: 14px;
      font-weight: 600;
      flex-shrink: 0;
    }

    .worklog-body {
      flex: 1;
      min-width: 0;

      .worklog-header {
        display: flex;
        justify-content: space-between;
        align-items: flex-start;
        margin-bottom: 8px;
        gap: 12px;

        .worklog-meta {
          display: flex;
          align-items: center;
          gap: 8px;
          flex-wrap: wrap;

          .worklog-author {
            font-weight: 600;
            color: #1f2937;
            font-size: 14px;
          }
        }

        .worklog-actions {
          display: flex;
          align-items: center;
          gap: 8px;
          flex-shrink: 0;

          .worklog-time {
            font-size: 13px;
            color: #9ca3af;
          }
        }
      }

      .worklog-text {
        color: #4b5563;
        font-size: 14px;
        line-height: 1.6;
        white-space: pre-wrap;
        word-break: break-word;
      }
    }
  }
}

// 工作流卡片
.workflow-card {
  .workflow-current-node {
    padding: 16px 20px;
    border-bottom: 1px solid #f0f0f0;

    .current-node-label {
      font-size: 12px;
      font-weight: 500;
      color: #6b7280;
      margin-bottom: 8px;
      text-transform: uppercase;
      letter-spacing: 0.5px;
    }

    .current-node-info {
      display: flex;
      align-items: center;
      gap: 8px;
    }
  }

  .workflow-approvals {
    padding: 16px 20px;
    border-bottom: 1px solid #f0f0f0;

    .approvals-label {
      font-size: 12px;
      font-weight: 500;
      color: #6b7280;
      margin-bottom: 10px;
      text-transform: uppercase;
      letter-spacing: 0.5px;
    }

    .approvals-list {
      display: flex;
      flex-direction: column;
      gap: 8px;
    }

    .approval-item {
      display: flex;
      align-items: center;
      gap: 10px;
      padding: 8px 12px;
      background: #f9fafb;
      border-radius: 8px;

      .approval-user {
        display: flex;
        align-items: center;
        gap: 8px;
        font-size: 13px;
        color: #1f2937;
        font-weight: 500;
      }

      .approval-comment {
        font-size: 12px;
        color: #6b7280;
        margin-left: auto;
      }
    }
  }

  .workflow-actions {
    padding: 16px 20px;
    border-bottom: 1px solid #f0f0f0;

    .actions-label {
      font-size: 12px;
      font-weight: 500;
      color: #6b7280;
      margin-bottom: 10px;
      text-transform: uppercase;
      letter-spacing: 0.5px;
    }

    .actions-row {
      display: flex;
      align-items: center;
    }
  }

  .workflow-history {
    padding: 16px 20px;

    .history-label {
      font-size: 12px;
      font-weight: 500;
      color: #6b7280;
      margin-bottom: 12px;
      text-transform: uppercase;
      letter-spacing: 0.5px;
    }

    .workflow-timeline {
      padding-left: 4px;
    }

    .history-content {
      font-size: 13px;

      .history-user {
        font-weight: 600;
        color: #1f2937;
      }

      .history-action {
        color: #6b7280;
        margin: 0 4px;
      }

      .history-arrow {
        color: #9ca3af;
        margin: 0 4px;
      }

      .history-comment {
        margin-top: 4px;
        font-size: 12px;
        color: #6b7280;
        padding: 6px 10px;
        background: #f9fafb;
        border-radius: 6px;
      }
    }
  }
}

// 响应式
@media (max-width: 768px) {
  .issue-header {
    flex-direction: column;
    gap: 16px;

    .header-actions {
      margin-left: 0;
      width: 100%;
    }
  }
}

// 编辑对话框
.edit-custom-fields {
  margin-top: 8px;

  .section-divider {
    display: flex;
    align-items: center;
    margin-bottom: 16px;
    padding-top: 8px;
    border-top: 1px dashed #e4e7ed;

    span {
      font-size: 13px;
      font-weight: 500;
      color: #606266;
    }
  }
}

:deep(.edit-dialog) {
  .el-dialog__body {
    max-height: 70vh;
    overflow-y: auto;
  }
}

:deep(.create-dialog) {
  .el-dialog__body {
    max-height: 70vh;
    overflow-y: auto;
  }

  .custom-fields-section {
    margin-top: 16px;
    padding-top: 16px;
    border-top: 1px dashed #e4e7ed;

    .section-divider {
      margin-bottom: 16px;
      font-size: 13px;
      font-weight: 500;
      color: #606266;
    }
  }
}

// Epic Issues 列表
.epic-issues-card {
  .epic-issues-list {
    padding: 8px;

    .empty-state {
      padding: 40px 0;
    }

    .epic-issue-item {
      display: flex;
      justify-content: space-between;
      align-items: center;
      padding: 14px 16px;
      margin-bottom: 8px;
      background: #ffffff;
      border-radius: 8px;
      border: 1px solid #e5e7eb;
      transition: all 0.2s ease;
      cursor: pointer;

      &:hover {
        background: #f9fafb;
        border-color: #667eea;
        box-shadow: 0 2px 8px rgba(102, 126, 234, 0.1);
        transform: translateY(-1px);
      }

      &:last-child {
        margin-bottom: 0;
      }

      .issue-left {
        flex: 1;
        display: flex;
        align-items: center;
        gap: 12px;
        min-width: 0;

        .issue-type-icon {
          width: 24px;
          height: 24px;
          border-radius: 4px;
          display: flex;
          align-items: center;
          justify-content: center;
          font-size: 14px;
          flex-shrink: 0;

          &.task {
            background: #3b82f6;
            color: white;
          }

          &.bug {
            background: #ef4444;
            color: white;
          }

          &.epic {
            background: #3b82f6;
            color: white;
          }
        }

        .issue-link {
          text-decoration: none;
          flex-shrink: 0;

          .issue-key {
            font-size: 13px;
            font-weight: 600;
            color: #667eea;
            transition: all 0.2s;

            &:hover {
              color: #5568d3;
              text-decoration: underline;
            }
          }
        }

        .issue-title {
          font-size: 14px;
          color: #1f2937;
          font-weight: 500;
          overflow: hidden;
          text-overflow: ellipsis;
          white-space: nowrap;
          flex: 1;
          min-width: 0;
        }
      }

      .issue-right {
        display: flex;
        align-items: center;
        gap: 10px;
        flex-shrink: 0;

        .priority-tag {
          font-size: 11px;
          font-weight: 600;
          padding: 2px 8px;
          border-radius: 4px;
        }

        .status-badge {
          display: flex;
          align-items: center;
          gap: 6px;
          padding: 4px 12px;
          border-radius: 12px;
          font-size: 12px;
          font-weight: 500;
          background: #f3f4f6;
          color: #6b7280;

          .status-dot {
            width: 6px;
            height: 6px;
            border-radius: 50%;
            background: currentColor;
          }

          &.open {
            background: #dbeafe;
            color: #1e40af;
          }

          &.in_progress {
            background: #fef3c7;
            color: #92400e;
          }

          &.resolved {
            background: #d1fae5;
            color: #065f46;
          }

          &.closed {
            background: #e5e7eb;
            color: #4b5563;
          }

          &.pending_review {
            background: #fffbeb;
            color: #b45309;
          }

          &.merged {
            background: #f3e8ff;
            color: #7c3aed;
          }
        }

        .assignee-info {
          display: flex;
          align-items: center;
          gap: 8px;
          flex-shrink: 0;

          .assignee-avatar {
            width: 28px;
            height: 28px;
            border-radius: 50%;
            background: #3b82f6;
            color: white;
            display: flex;
            align-items: center;
            justify-content: center;
            font-size: 12px;
            font-weight: 600;
            transition: all 0.2s;
            flex-shrink: 0;

            &.unassigned {
              background: #e5e7eb;
              color: #9ca3af;
            }
          }

          .assignee-name {
            font-size: 13px;
            color: #4b5563;
            font-weight: 500;
            white-space: nowrap;

            &.unassigned {
              color: #9ca3af;
              font-style: italic;
            }
          }

          &:hover .assignee-avatar {
            transform: scale(1.1);
            box-shadow: 0 2px 8px rgba(102, 126, 234, 0.3);
          }
        }
      }
    }
  }
}

// 子任务列表（复用 Epic Issues 样式）
.subtasks-card {
  .subtasks-list {
    padding: 8px;

    .empty-state-compact {
      padding: 12px 16px;
      text-align: center;

      .empty-text {
        color: #9ca3af;
        font-size: 13px;
      }
    }

    .subtask-item {
      display: flex;
      justify-content: space-between;
      align-items: center;
      padding: 14px 16px;
      margin-bottom: 8px;
      background: #ffffff;
      border-radius: 8px;
      border: 1px solid #e5e7eb;
      transition: all 0.2s ease;
      cursor: pointer;

      &:hover {
        background: #f9fafb;
        border-color: #fa709a;
        box-shadow: 0 2px 8px rgba(250, 112, 154, 0.1);
        transform: translateY(-1px);
      }

      &:last-child {
        margin-bottom: 0;
      }

      .issue-left {
        flex: 1;
        display: flex;
        align-items: center;
        gap: 12px;
        min-width: 0;

        .issue-type-icon {
          width: 24px;
          height: 24px;
          border-radius: 4px;
          display: flex;
          align-items: center;
          justify-content: center;
          font-size: 14px;
          flex-shrink: 0;

          &.task {
            background: #3b82f6;
            color: white;
          }

          &.bug {
            background: #ef4444;
            color: white;
          }
        }

        .issue-link {
          text-decoration: none;
          flex-shrink: 0;

          .issue-key {
            font-size: 13px;
            font-weight: 600;
            color: #fa709a;
            transition: all 0.2s;

            &:hover {
              color: #f5576c;
              text-decoration: underline;
            }
          }
        }

        .issue-title {
          font-size: 14px;
          color: #1f2937;
          font-weight: 500;
          overflow: hidden;
          text-overflow: ellipsis;
          white-space: nowrap;
          flex: 1;
          min-width: 0;
        }
      }

      .issue-right {
        display: flex;
        align-items: center;
        gap: 10px;
        flex-shrink: 0;

        .priority-tag {
          font-size: 11px;
          font-weight: 600;
          padding: 2px 8px;
          border-radius: 4px;
        }

        .status-badge {
          display: flex;
          align-items: center;
          gap: 6px;
          padding: 4px 12px;
          border-radius: 12px;
          font-size: 12px;
          font-weight: 500;
          background: #f3f4f6;
          color: #6b7280;

          .status-dot {
            width: 6px;
            height: 6px;
            border-radius: 50%;
            background: currentColor;
          }

          &.open {
            background: #dbeafe;
            color: #1e40af;
          }

          &.in_progress {
            background: #fef3c7;
            color: #92400e;
          }

          &.resolved {
            background: #d1fae5;
            color: #065f46;
          }

          &.closed {
            background: #e5e7eb;
            color: #4b5563;
          }

          &.pending_review {
            background: #fffbeb;
            color: #b45309;
          }

          &.merged {
            background: #f3e8ff;
            color: #7c3aed;
          }
        }

        .assignee-info {
          display: flex;
          align-items: center;
          gap: 8px;
          flex-shrink: 0;

          .assignee-avatar {
            width: 28px;
            height: 28px;
            border-radius: 50%;
            background: #f59e0b;
            color: white;
            display: flex;
            align-items: center;
            justify-content: center;
            font-size: 12px;
            font-weight: 600;
            transition: all 0.2s;
            flex-shrink: 0;

            &.unassigned {
              background: #e5e7eb;
              color: #9ca3af;
            }
          }

          .assignee-name {
            font-size: 13px;
            color: #4b5563;
            font-weight: 500;
            white-space: nowrap;

            &.unassigned {
              color: #9ca3af;
              font-style: italic;
            }
          }

          &:hover .assignee-avatar {
            transform: scale(1.1);
            box-shadow: 0 2px 8px rgba(250, 112, 154, 0.3);
          }
        }
      }
    }
  }
}

// 附件区域
.attachment-section {
  padding: 20px;
}

// 工作流流程图
.workflow-diagram {
  min-height: 120px;
  overflow: auto;
  padding: 20px 0;

  .diagram-graph {
    position: relative;
    margin: 0 auto;
  }

  .diagram-edges {
    position: absolute;
    top: 0;
    left: 0;
    pointer-events: none;
  }

  .diagram-node {
    position: absolute;
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    width: 110px;
    height: 80px;
    border-radius: 10px;
    border: 2px solid #dcdfe6;
    background: #f5f7fa;
    transition: all 0.3s ease;
    cursor: default;

    .diagram-node-icon {
      font-size: 20px;
      margin-bottom: 4px;
      color: #909399;
    }

    .diagram-node-name {
      font-size: 13px;
      font-weight: 600;
      color: #303133;
      white-space: nowrap;
      max-width: 100px;
      overflow: hidden;
      text-overflow: ellipsis;
    }

    .diagram-node-type {
      font-size: 11px;
      color: #909399;
      margin-top: 2px;
    }

    .diagram-node-check {
      position: absolute;
      top: -8px;
      right: -8px;
      width: 20px;
      height: 20px;
      border-radius: 50%;
      background: #67c23a;
      color: #fff;
      font-size: 12px;
      display: flex;
      align-items: center;
      justify-content: center;
      font-weight: bold;
    }

    // 节点类型图标颜色
    &.node-start .diagram-node-icon { color: #67c23a; }
    &.node-end .diagram-node-icon { color: #909399; }
    &.node-approval .diagram-node-icon { color: #e6a23c; }
    &.node-work .diagram-node-icon { color: #409eff; }

    // 已完成节点
    &.visited {
      border-color: #67c23a;
      background: #f0f9eb;

      .diagram-node-name { color: #67c23a; }
      .diagram-node-icon { color: #67c23a; }
    }

    // 当前节点
    &.current {
      border-color: #409eff;
      background: #ecf5ff;
      box-shadow: 0 0 0 3px rgba(64, 158, 255, 0.2);
      animation: pulse-border 2s ease-in-out infinite;

      .diagram-node-name { color: #409eff; }
      .diagram-node-icon { color: #409eff; }
    }

    // 被取消节点
    &.cancelled {
      border-color: #f56c6c;
      background: #fef0f0;

      .diagram-node-name { color: #f56c6c; }
      .diagram-node-icon { color: #f56c6c; }
    }

    // 未到达节点
    &.pending {
      border-color: #dcdfe6;
      background: #f5f7fa;
      opacity: 0.7;
    }
  }

  .diagram-empty {
    display: flex;
    justify-content: center;
    padding: 20px;
  }
}

@keyframes pulse-border {
  0%, 100% {
    box-shadow: 0 0 0 3px rgba(64, 158, 255, 0.2);
  }
  50% {
    box-shadow: 0 0 0 6px rgba(64, 158, 255, 0.1);
  }
}

// 关联告警样式
.alert-card {
  :deep(.el-table) {
    .clickable-row { cursor: pointer; &:hover { background-color: #f9fafb; } }
  }

  .alert-name-text { font-weight: 500; color: #1f2937; font-size: 13px; }

  .alert-status-badge {
    display: inline-flex;
    align-items: center;
    gap: 4px;
    padding: 2px 8px;
    border-radius: 20px;
    font-size: 12px;

    .status-dot { width: 6px; height: 6px; border-radius: 50%; }

    &.firing { background: #fef2f2; color: #dc2626; .status-dot { background: #ef4444; } }
    &.resolved { background: #ecfdf5; color: #059669; .status-dot { background: #10b981; } }
  }
}
</style>
