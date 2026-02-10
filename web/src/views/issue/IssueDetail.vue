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
              :disabled="!canOperateWorkflow && workflowInstance.status !== 'active'"
              class="workflow-action-btn"
            >
              <el-icon><Promotion /></el-icon>
              {{ workflowActionBtnText }}
              <el-icon class="el-icon--right"><ArrowDown /></el-icon>
            </el-button>
            <template #dropdown>
              <el-dropdown-menu>
                <!-- 审批节点操作 -->
                <template v-if="workflowInstance.current_node?.node_type === 'approval' && workflowInstance.status === 'active'">
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
                <template v-else-if="isWorkNode && workflowInstance.status === 'active'">
                  <el-dropdown-item command="complete">
                    <el-icon style="color: #409eff"><Check /></el-icon>
                    完成节点
                  </el-dropdown-item>
                </template>
                <!-- 工作流已结束提示 -->
                <template v-else-if="workflowInstance.status !== 'active'">
                  <el-dropdown-item disabled>
                    工作流已{{ workflowInstance.status === 'completed' ? '完成' : '取消' }}
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

          <!-- 工作流卡片 -->
          <el-card v-if="workflowInstance" shadow="never" class="content-card workflow-card">
            <template #header>
              <div class="card-header-group">
                <div class="card-icon workflow">
                  <el-icon><Promotion /></el-icon>
                </div>
                <span class="card-title">工作流</span>
                <el-tag :type="getWorkflowStatusType(workflowInstance.status)" size="small" effect="dark">
                  {{ getWorkflowStatusText(workflowInstance.status) }}
                </el-tag>
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

            <!-- 审批记录 -->
            <div v-if="workflowInstance.approvals && workflowInstance.approvals.length > 0" class="workflow-approvals">
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

            <!-- 审批操作按钮 -->
            <div v-if="isCurrentUserApprover && workflowInstance.status === 'active'" class="workflow-actions">
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
              <div class="actions-label">节点操作</div>
              <div class="actions-row">
                <el-input v-model="completeComment" placeholder="完成备注（可选）" size="default" style="flex: 1; margin-right: 12px;" />
                <el-button type="primary" :loading="completeLoading" @click="handleComplete">
                  <el-icon><Check /></el-icon>
                  完成节点
                </el-button>
              </div>
            </div>

            <!-- 流转历史时间线 -->
            <div v-if="workflowHistoryList.length > 0" class="workflow-history">
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
                  <template v-if="field.value !== null && field.value !== undefined && field.value !== ''">
                    <!-- Epic Link 特殊显示 -->
                    <template v-if="field.field_type === 'epic_link' && typeof field.value === 'object' && field.value.issue_key">
                      <router-link :to="`/issues/${field.value.issue_key}`" class="epic-link">
                        <el-icon><Link /></el-icon>
                        {{ field.value.issue_key }}: {{ field.value.title }}
                      </router-link>
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
                  <div v-for="comment in comments" :key="comment.id" class="comment-item">
                    <div class="comment-avatar">
                      {{ comment.user?.display_name?.charAt(0) || '?' }}
                    </div>
                    <div class="comment-body">
                      <div class="comment-header">
                        <span class="comment-author">{{ comment.user?.display_name || '未知用户' }}</span>
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
                            <el-option label="开发" value="开发" />
                            <el-option label="测试" value="测试" />
                            <el-option label="调试" value="调试" />
                            <el-option label="文档" value="文档" />
                            <el-option label="其他" value="其他" />
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

          <!-- 附件 -->
          <el-card shadow="never" class="content-card">
            <template #header>
              <div class="card-header-group">
                <div class="card-icon attachment">
                  <el-icon><Paperclip /></el-icon>
                </div>
                <span class="card-title">附件</span>
                <span class="card-count">{{ attachments.length }}</span>
              </div>
            </template>
            <div class="attachment-section">
              <AttachmentUpload v-if="issue" :issue-key="issue.issue_key" @success="loadAttachments(issue.issue_key)" />
              <AttachmentList v-if="issue" :issue-key="issue.issue_key" :attachments="attachments" @refresh="loadAttachments(issue.issue_key)" />
            </div>
          </el-card>

          <!-- 活动记录 -->
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
                  {{ issue.epic_key }}
                </el-link>
              </div>
              <div class="info-item">
                <span class="info-label">状态</span>
                <div class="status-badge sm" :class="issue.status">
                  <span class="status-dot"></span>
                  <span>{{ getStatusText(issue.status) }}</span>
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
            <el-form-item label="优先级" prop="priority">
              <el-select v-model="editForm.priority" style="width: 100%">
                <el-option label="P0 - 紧急" value="P0" />
                <el-option label="P1 - 高" value="P1" />
                <el-option label="P2 - 中" value="P2" />
                <el-option label="P3 - 低" value="P3" />
              </el-select>
            </el-form-item>
          </el-col>
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
        <el-row :gutter="20">
          <el-col :span="12">
            <el-form-item label="指派人">
              <div style="display: flex; gap: 8px;">
                <el-select v-model="editForm.assignee_id" placeholder="请选择" style="flex: 1" clearable filterable>
                  <el-option v-for="u in users" :key="u.id" :label="u.display_name" :value="u.id" />
                </el-select>
                <el-button @click="assignToMe">分配给我</el-button>
              </div>
            </el-form-item>
          </el-col>
        </el-row>
        <el-form-item label="描述">
          <el-input v-model="editForm.description" type="textarea" :rows="3" />
        </el-form-item>
        <el-form-item label="Epic">
          <el-select v-model="editForm.epic_id" placeholder="请选择Epic" style="width: 100%" clearable filterable>
            <el-option v-for="epic in epicOptions" :key="epic.id" :label="`${epic.issue_key}: ${epic.title}`" :value="epic.id" />
          </el-select>
        </el-form-item>
        <el-row :gutter="20">
          <el-col :span="12">
            <el-form-item label="预计开始时间">
              <el-date-picker
                v-model="editForm.planned_start_date"
                type="date"
                placeholder="请选择预计开始时间"
                style="width: 100%"
                value-format="YYYY-MM-DD"
              />
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="预计交付时间">
              <el-date-picker
                v-model="editForm.planned_end_date"
                type="date"
                placeholder="请选择预计交付时间"
                style="width: 100%"
                value-format="YYYY-MM-DD"
              />
            </el-form-item>
          </el-col>
        </el-row>

        <!-- 自定义字段 -->
        <div v-if="editFieldScheme.length > 0" class="edit-custom-fields">
          <div class="section-divider">
            <span>扩展字段</span>
          </div>
          <el-row :gutter="20">
            <el-col
              v-for="item in editFieldScheme"
              :key="item.field_id"
              :span="item.field?.field_type === 'textarea' ? 24 : 12"
            >
              <el-form-item :label="item.field?.field_name" :required="item.is_required">
                <FieldRenderer
                  v-if="item.field && issue"
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
        <el-row :gutter="20">
          <el-col :span="12">
            <el-form-item label="优先级" prop="priority">
              <el-select v-model="createSubtaskForm.priority" placeholder="请选择优先级" style="width: 100%">
                <el-option label="P0 - 紧急" value="P0" />
                <el-option label="P1 - 高" value="P1" />
                <el-option label="P2 - 中" value="P2" />
                <el-option label="P3 - 低" value="P3" />
              </el-select>
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="指派人">
              <el-select v-model="createSubtaskForm.assignee_id" placeholder="请选择指派人" style="width: 100%" clearable filterable>
                <el-option v-for="u in users" :key="u.id" :label="u.display_name" :value="u.id" />
              </el-select>
            </el-form-item>
          </el-col>
        </el-row>
        <el-form-item label="描述">
          <el-input v-model="createSubtaskForm.description" type="textarea" :rows="3" placeholder="请输入子任务描述" />
        </el-form-item>

        <!-- 动态自定义字段 -->
        <div v-if="createSubtaskFieldScheme.length > 0" class="custom-fields-section">
          <div class="section-divider">
            <span>扩展字段</span>
          </div>
          <el-row :gutter="20">
            <el-col
              v-for="item in createSubtaskFieldScheme"
              :key="item.field_id"
              :span="item.field?.field_type === 'textarea' ? 24 : 12"
            >
              <el-form-item :label="item.field?.field_name" :required="item.is_required">
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
    <el-dialog v-model="diagramVisible" title="工单流程" width="800px" destroy-on-close>
      <div v-loading="diagramLoading" class="workflow-diagram">
        <div v-if="diagramNodes.length > 0" class="diagram-flow">
          <template v-for="(node, index) in diagramNodes" :key="node.id">
            <div class="diagram-node" :class="getNodeDiagramClass(node)">
              <div class="diagram-node-icon">
                <span v-if="node.node_type === 'start'">▶</span>
                <span v-else-if="node.node_type === 'end'">◉</span>
                <span v-else-if="node.node_type === 'approval'">✓</span>
                <span v-else-if="node.node_type === 'work'">⚙</span>
                <span v-else>●</span>
              </div>
              <div class="diagram-node-name">{{ node.name }}</div>
              <div class="diagram-node-type">{{ getNodeTypeText(node.node_type) }}</div>
              <div v-if="getNodeDiagramClass(node).visited" class="diagram-node-check">✓</div>
            </div>
            <div v-if="index < diagramNodes.length - 1" class="diagram-arrow" :class="{ visited: isEdgeVisited(node.id, diagramNodes[index + 1]?.id) }">
              <div class="arrow-line"></div>
              <div class="arrow-head"></div>
            </div>
          </template>
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
  User, Clock, Edit, ArrowDown, ArrowRight, Plus, Document,
  ChatLineRound, InfoFilled, View, Check, Link, Delete, Paperclip, Promotion
} from '@element-plus/icons-vue'
import {
  getIssueDetail, updateIssue, deleteIssue, createIssue, getIssueList,
  getIssueComments, addIssueComment, getIssueActivities, getIssueWatchers,
  getWorklogs, addWorklog, updateWorklog, deleteWorklog,
  addIssueWatcher, removeIssueWatcher, getEpicIssues, getSubtasks,
} from '@/api/issue'
import { listAttachments } from '@/api/attachment'
import { getWorkflowInstance, getWorkflowHistory, approveWorkflow, rejectWorkflow, completeWorkflow, getWorkflowNodes, getWorkflowEdges } from '@/api/workflow'
import type { WorkflowInstance, WorkflowHistory, WorkflowNode, WorkflowEdge } from '@/types/workflow'
import type { Attachment } from '@/types/attachment'
import AttachmentUpload from '@/components/attachment/AttachmentUpload.vue'
import AttachmentList from '@/components/attachment/AttachmentList.vue'
import { getAllUsers } from '@/api/user'
import { getIssueFieldValues, getFieldScheme } from '@/api/field'
import { getAllProjects, getProjectIssueTypes } from '@/api/project'
import { useUserStore } from '@/stores/user'
import type { Issue, IssueComment, IssueActivity, IssueWatcher, UpdateIssueRequest, Worklog, CreateWorklogRequest, CreateIssueRequest, IssuePriority, CustomFieldValue } from '@/types/issue'
import type { UserOption } from '@/types/user'
import type { FieldValue, FieldSchemeItem } from '@/types/field'
import type { Project, ProjectIssueType } from '@/types/project'
import FieldRenderer from '@/components/field/FieldRenderer.vue'
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

// 工作流相关
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
  description: string
  priority: IssuePriority
  assignee_id: number | undefined
}

const createSubtaskForm = reactive<CreateSubtaskFormData>({
  project_key: '',
  issue_type_id: undefined,
  title: '',
  description: '',
  priority: 'P2',
  assignee_id: undefined,
})

const createSubtaskRules: FormRules = {
  project_key: [{ required: true, message: '请选择项目', trigger: 'change' }],
  issue_type_id: [{ required: true, message: '请选择类型', trigger: 'change' }],
  title: [{ required: true, message: '请输入标题', trigger: 'blur' }],
  priority: [{ required: true, message: '请选择优先级', trigger: 'change' }],
}

const newComment = ref('')
const commentLoading = ref(false)
const editDialogVisible = ref(false)
const editLoading = ref(false)
const editFormRef = ref<FormInstance>()
const editForm = reactive<UpdateIssueRequest>({
  title: '', description: '', priority: undefined, resolution: undefined, epic_id: undefined, assignee_id: undefined,
  planned_start_date: undefined, planned_end_date: undefined,
})
const editRules: FormRules = {
  title: [{ required: true, message: '请输入标题', trigger: 'blur' }],
}
// 编辑用的字段方案和值
const editFieldScheme = ref<FieldSchemeItem[]>([])
const editFieldValues = ref<Record<number, any>>({})
const epicOptions = ref<Issue[]>([])

const loadIssue = async () => {
  const key = route.params.key as string
  if (!key) return
  loading.value = true
  try {
    const { data } = await getIssueDetail(key)
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
    ])
  } catch (error) {
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

  // 只有工作流实例存在时才加载流转历史
  try {
    const { data } = await getWorkflowHistory(key)
    workflowHistoryList.value = (data as any).data || []
  } catch (e: any) {
    workflowHistoryList.value = []
  }
}

// 判断当前用户是否是审批人
const isCurrentUserApprover = computed(() => {
  if (!workflowInstance.value || !userStore.user) return false
  const approvals = workflowInstance.value.approvals || []
  return approvals.some(
    a => a.approver_id === userStore.user!.id && a.status === 'pending'
  )
})

// 判断当前节点是否是工作节点（可完成）
const isWorkNode = computed(() => {
  if (!workflowInstance.value || workflowInstance.value.status !== 'active') return false
  const nodeType = workflowInstance.value.current_node?.node_type
  return nodeType === 'work' || nodeType === 'start'
})

// 当前用户是否可以操作工作流
const canOperateWorkflow = computed(() => {
  return isCurrentUserApprover.value || isWorkNode.value
})

// 工作流快捷按钮文本
const workflowActionBtnText = computed(() => {
  if (!workflowInstance.value) return '工作流'
  if (workflowInstance.value.status !== 'active') {
    return workflowInstance.value.status === 'completed' ? '已完成' : '已取消'
  }
  const nodeName = workflowInstance.value.current_node?.name || '当前节点'
  return nodeName
})

// 工作流下拉菜单命令处理
const handleWorkflowCommand = (command: string) => {
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
    await ElMessageBox.confirm('确认完成当前节点？', '完成确认', {
      confirmButtonText: '完成',
      cancelButtonText: '取消',
      type: 'info',
    })
    handleComplete()
  } catch {
    // 用户取消
  }
}

const completeLoading = ref(false)
const completeComment = ref('')

// 完成工作节点
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
  }
  return map[status] || status
}

const getWorkflowStatusType = (status: string): 'success' | 'info' | 'warning' | 'danger' => {
  const map: Record<string, 'success' | 'info' | 'warning' | 'danger'> = {
    active: 'warning',
    completed: 'success',
    cancelled: 'info',
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

const loadCustomFields = async (issueId: number) => {
  if (!issue.value) return
  try {
    // 获取该工单类型配置的字段方案
    const schemeRes = await getFieldScheme(issue.value.project_key, issue.value.issue_type_id)
    const scheme = schemeRes.data.data || []

    // 只获取详情页可见的字段
    const visibleScheme = scheme.filter((s: FieldSchemeItem) => s.is_visible_detail)

    // 获取已保存的字段值
    const valuesRes = await getIssueFieldValues(issueId)
    const savedValues = valuesRes.data.data || []

    // 合并：用字段方案作为基础，填充已保存的值
    const valueMap = new Map(savedValues.map((v: FieldValue) => [v.field_id, v]))

    customFields.value = visibleScheme.map((item: FieldSchemeItem) => {
      const savedValue = valueMap.get(item.field_id)
      const result = {
        field_id: item.field_id,
        field_key: item.field?.field_key || '',
        field_name: item.field?.field_name || '',
        field_type: item.field?.field_type || '',
        value: savedValue?.value ?? null,
        display_value: savedValue?.display_value || ''
      }
      console.log(`字段 ${item.field?.field_name}:`, result)
      return result
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

const assignToMe = () => {
  if (userStore.user) {
    editForm.assignee_id = userStore.user.id
  }
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
    description: '',
    priority: 'P2',
    assignee_id: undefined,
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
    createSubtaskFieldScheme.value = schemeItems.filter(item => item.is_visible_create && item.field?.field_key !== 'epic_link')
    createSubtaskFieldScheme.value.forEach(item => {
      if (item.default_value) {
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
      const customFields = Object.entries(createSubtaskFieldValues.value)
        .filter(([_, value]) => value !== undefined && value !== null && value !== '')
        .map(([fieldId, value]) => ({
          field_id: parseInt(fieldId),
          value: value,
        }))

      const requestData: CreateIssueRequest = {
        ...createSubtaskForm,
        issue_type_id: createSubtaskForm.issue_type_id!,
        parent_id: issue.value!.id,  // 设置父工单ID
        custom_fields: customFields.length > 0 ? customFields : undefined,
      }

      const { data } = await createIssue(requestData)
      ElMessage.success('子任务创建成功')
      createSubtaskDialogVisible.value = false
      // 重新加载子任务列表
      await loadSubtasks(issue.value.issue_key)
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

  // 加载Epic选项（获取当前项目的所有Epic类型工单）
  try {
    const { data } = await getIssueList({
      project_key: issue.value.project_key,
      page: 1,
      page_size: 100
    })
    // 过滤出Epic类型的工单
    epicOptions.value = data.data.items.filter(i => i.issue_type?.name?.toLowerCase() === 'epic')
  } catch (e) {
    console.error('Failed to load epic options:', e)
  }

  Object.assign(editForm, {
    title: issue.value.title, description: issue.value.description,
    priority: issue.value.priority, resolution: issue.value.resolution || undefined,
    epic_id: issue.value.epic_id, assignee_id: issue.value.assignee_id,
    planned_start_date: issue.value.planned_start_date ? dayjs(issue.value.planned_start_date).format('YYYY-MM-DD') : undefined,
    planned_end_date: issue.value.planned_end_date ? dayjs(issue.value.planned_end_date).format('YYYY-MM-DD') : undefined,
  })

  // 加载编辑用的字段方案
  try {
    const schemeRes = await getFieldScheme(issue.value.project_key, issue.value.issue_type_id)
    const scheme = schemeRes.data.data || []
    editFieldScheme.value = scheme.filter((s: FieldSchemeItem) => s.is_visible_edit && s.field?.field_key !== 'epic_link')

    // 加载当前字段值
    const valuesRes = await getIssueFieldValues(issue.value.id)
    const savedValues = valuesRes.data.data || []
    editFieldValues.value = {}
    savedValues.forEach((v: FieldValue) => {
      editFieldValues.value[v.field_id] = v.value
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
      // 构建自定义字段值
      const customFields = Object.entries(editFieldValues.value)
        .filter(([_, value]) => value !== undefined && value !== null && value !== '')
        .map(([fieldId, value]) => ({
          field_id: parseInt(fieldId),
          value: value,
        }))

      const updateData: UpdateIssueRequest = {
        ...editForm,
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
    const workedAtDate = worklogForm.worked_at instanceof Date
      ? worklogForm.worked_at
      : new Date(worklogForm.worked_at)

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

const handleEditWorklog = (worklog: Worklog) => {
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
  const map: Record<string, string> = { open: '待处理', in_progress: '进行中', resolved: '已完成', closed: '已终止', reopened: '重新打开', merged: '已合并' }
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

onMounted(() => { loadIssue() })

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

// 拓扑排序：按 edges 的 source→target 关系排序节点
const topologicalSort = (nodes: WorkflowNode[], edges: WorkflowEdge[]): WorkflowNode[] => {
  if (nodes.length === 0) return []

  const nodeMap = new Map<number, WorkflowNode>()
  nodes.forEach(n => nodeMap.set(n.id, n))

  const inDegree = new Map<number, number>()
  const adjacency = new Map<number, number[]>()
  nodes.forEach(n => {
    inDegree.set(n.id, 0)
    adjacency.set(n.id, [])
  })

  edges.forEach(e => {
    if (nodeMap.has(e.source_node_id) && nodeMap.has(e.target_node_id)) {
      inDegree.set(e.target_node_id, (inDegree.get(e.target_node_id) || 0) + 1)
      adjacency.get(e.source_node_id)?.push(e.target_node_id)
    }
  })

  // BFS 拓扑排序
  const queue: number[] = []
  inDegree.forEach((deg, id) => {
    if (deg === 0) queue.push(id)
  })

  const sorted: WorkflowNode[] = []
  while (queue.length > 0) {
    const id = queue.shift()!
    const node = nodeMap.get(id)
    if (node) sorted.push(node)
    for (const next of (adjacency.get(id) || [])) {
      const newDeg = (inDegree.get(next) || 1) - 1
      inDegree.set(next, newDeg)
      if (newDeg === 0) queue.push(next)
    }
  }

  // 如果有未排序的节点（可能有环），追加到末尾
  if (sorted.length < nodes.length) {
    nodes.forEach(n => {
      if (!sorted.find(s => s.id === n.id)) sorted.push(n)
    })
  }

  return sorted
}

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

  return {
    current: isCurrent && instanceStatus === 'active',
    visited: isVisited || (isCurrent && instanceStatus === 'completed'),
    cancelled: instanceStatus === 'cancelled' && isCurrent,
    pending: !isCurrent && !isVisited,
    'node-start': node.node_type === 'start',
    'node-end': node.node_type === 'end',
    'node-approval': node.node_type === 'approval',
    'node-work': node.node_type === 'work',
  }
}

// 判断边是否已访问
const isEdgeVisited = (fromId: number, toId: number | undefined) => {
  if (!toId) return false
  return visitedEdgePairs.value.has(`${fromId}-${toId}`)
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
  diagramLoading.value = true
  try {
    const workflowId = workflowInstance.value!.workflow_id
    const [nodesRes, edgesRes] = await Promise.all([
      getWorkflowNodes(workflowId),
      getWorkflowEdges(workflowId),
    ])
    const nodes = (nodesRes.data as any).data || []
    const edges = (edgesRes.data as any).data || []
    diagramEdges.value = edges
    diagramNodes.value = topologicalSort(nodes, edges)
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
      background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
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

  &.desc { background: linear-gradient(135deg, #667eea 0%, #764ba2 100%); }
  &.comment { background: linear-gradient(135deg, #11998e 0%, #38ef7d 100%); }
  &.activity { background: linear-gradient(135deg, #4facfe 0%, #00f2fe 100%); }
  &.info { background: linear-gradient(135deg, #f093fb 0%, #f5576c 100%); }
  &.watcher { background: linear-gradient(135deg, #fbbf24 0%, #f59e0b 100%); }
  &.custom { background: linear-gradient(135deg, #a78bfa 0%, #8b5cf6 100%); }
  &.epic { background: linear-gradient(135deg, #4facfe 0%, #00f2fe 100%); }
  &.subtask { background: linear-gradient(135deg, #fa709a 0%, #fee140 100%); }
  &.attachment { background: linear-gradient(135deg, #f59e0b 0%, #d97706 100%); }
  &.workflow { background: linear-gradient(135deg, #8b5cf6 0%, #6d28d9 100%); }
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
      background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
      color: #fff;
      display: flex;
      align-items: center;
      justify-content: center;
      font-size: 15px;
      font-weight: 600;
      flex-shrink: 0;
    }

    .comment-body {
      flex: 1;

      .comment-header {
        display: flex;
        align-items: center;
        gap: 12px;
        margin-bottom: 8px;

        .comment-author { font-weight: 600; color: #1f2937; font-size: 14px; }
        .comment-time { font-size: 12px; color: #9ca3af; }
      }

      .comment-text {
        color: #374151;
        line-height: 1.6;
        white-space: pre-wrap;
        font-size: 14px;
      }
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
  }
}

.mini-avatar {
  width: 22px;
  height: 22px;
  border-radius: 6px;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
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
      background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
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
            background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
            color: white;
          }

          &.bug {
            background: linear-gradient(135deg, #f093fb 0%, #f5576c 100%);
            color: white;
          }

          &.epic {
            background: linear-gradient(135deg, #4facfe 0%, #00f2fe 100%);
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
            background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
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
            background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
            color: white;
          }

          &.bug {
            background: linear-gradient(135deg, #f093fb 0%, #f5576c 100%);
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
            background: linear-gradient(135deg, #fa709a 0%, #fee140 100%);
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
  overflow-x: auto;
  padding: 20px 0;

  .diagram-flow {
    display: flex;
    align-items: center;
    justify-content: center;
    gap: 0;
    min-width: max-content;
    padding: 10px 20px;
  }

  .diagram-node {
    position: relative;
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    min-width: 100px;
    padding: 12px 16px;
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
      max-width: 120px;
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

  .diagram-arrow {
    display: flex;
    align-items: center;
    margin: 0 4px;

    .arrow-line {
      width: 32px;
      height: 2px;
      background: #dcdfe6;
    }

    .arrow-head {
      width: 0;
      height: 0;
      border-top: 6px solid transparent;
      border-bottom: 6px solid transparent;
      border-left: 8px solid #dcdfe6;
    }

    &.visited {
      .arrow-line { background: #67c23a; }
      .arrow-head { border-left-color: #67c23a; }
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
</style>
