<template>
  <div class="project-settings-container">
    <!-- 页面头部 -->
    <TdPageHeader>
      <template #leading>
        <el-button class="back-btn" circle @click="$router.push(`/projects/${projectKey}`)">
          <el-icon><ArrowLeft /></el-icon>
        </el-button>
        <div class="page-header-icon">
          <el-icon :size="20"><Setting /></el-icon>
        </div>
      </template>
      <template #title>项目设置</template>
      <template #subtitle>{{ projectKey }} - 配置项目信息和权限</template>
    </TdPageHeader>

    <!-- 标签页 -->
    <el-card shadow="never" class="settings-card">
      <el-tabs v-model="activeTab" class="settings-tabs">
        <!-- 基本信息 -->
        <el-tab-pane name="basic">
          <template #label>
            <span class="tab-label">
              <el-icon><Document /></el-icon>
              基本信息
            </span>
          </template>
          <div v-loading="loading" class="tab-content">
            <el-form ref="basicFormRef" :model="basicForm" :rules="basicRules" label-width="120px" class="basic-form">
              <el-form-item label="项目标识">
                <el-input v-model="basicForm.project_key" disabled>
                  <template #prefix>
                    <el-icon><Key /></el-icon>
                  </template>
                </el-input>
                <div class="form-tip">
                  <el-icon><InfoFilled /></el-icon>
                  <span>项目标识创建后不可修改</span>
                </div>
              </el-form-item>
              <el-form-item label="项目名称" prop="name">
                <el-input v-model="basicForm.name" placeholder="请输入项目名称" maxlength="50" show-word-limit>
                  <template #prefix>
                    <el-icon><Folder /></el-icon>
                  </template>
                </el-input>
              </el-form-item>
              <el-form-item label="项目描述">
                <el-input
                  v-model="basicForm.description"
                  type="textarea"
                  :rows="4"
                  placeholder="请输入项目描述"
                  maxlength="500"
                  show-word-limit
                />
              </el-form-item>
              <el-form-item label="项目负责人">
                <el-select v-model="basicForm.lead_user_id" placeholder="请选择负责人" filterable style="width: 100%" popper-class="user-select-popper">
                  <template #prefix>
                    <el-icon><User /></el-icon>
                  </template>
                  <el-option v-for="u in users" :key="u.id" :label="u.display_name" :value="u.id">
                    <div class="user-option">
                      <div class="user-option-avatar">{{ u.display_name?.charAt(0) || '?' }}</div>
                      <div class="user-option-info">
                        <span class="user-option-name">{{ u.display_name }}</span>
                        <span class="user-option-username">@{{ u.username }}</span>
                      </div>
                    </div>
                  </el-option>
                </el-select>
              </el-form-item>
              <el-form-item label="项目状态">
                <div class="status-field">
                  <el-switch
                    v-model="basicForm.status"
                    :active-value="1"
                    :inactive-value="0"
                    active-text="启用"
                    inactive-text="禁用"
                    active-color="#10b981"
                    inactive-color="#9ca3af"
                  />
                  <div class="form-tip">
                    <el-icon><InfoFilled /></el-icon>
                    <span>禁用后项目将不可访问</span>
                  </div>
                </div>
              </el-form-item>
              <div class="form-actions">
                <el-button type="primary" :loading="saveLoading" size="large" @click="saveBasicInfo">
                  <el-icon><Check /></el-icon>
                  保存更改
                </el-button>
              </div>
            </el-form>
          </div>
        </el-tab-pane>

        <!-- 项目成员 -->
        <el-tab-pane name="members">
          <template #label>
            <span class="tab-label">
              <el-icon><User /></el-icon>
              项目成员
            </span>
          </template>
          <div v-loading="membersLoading" class="tab-content">
            <div class="section-header">
              <div class="section-title">
                <el-icon><UserFilled /></el-icon>
                <span>成员列表</span>
                <el-tag size="small" type="info" style="margin-left: 8px">{{ members.length }} 人</el-tag>
              </div>
              <el-button type="primary" @click="openMemberDialog">
                <el-icon><Plus /></el-icon>
                添加成员
              </el-button>
            </div>
            <el-table :data="members" stripe class="members-table">
              <el-table-column label="成员" min-width="200">
                <template #default="{ row }">
                  <div class="member-cell">
                    <div class="member-avatar" :class="row.role">
                      {{ row.user?.display_name?.charAt(0) || '?' }}
                    </div>
                    <div class="member-info">
                      <div class="member-name">{{ row.user?.display_name }}</div>
                      <div class="member-username">@{{ row.user?.username }}</div>
                    </div>
                  </div>
                </template>
              </el-table-column>
              <el-table-column label="角色" width="120">
                <template #default="{ row }">
                  <el-tag :type="getRoleType(row.role)" size="small">
                    {{ row.role_name || row.role }}
                  </el-tag>
                </template>
              </el-table-column>
              <el-table-column label="操作" width="200" align="right">
                <template #default="{ row }">
                  <el-dropdown trigger="click" @command="(cmd: string) => handleEditMemberRole(row, cmd)">
                    <el-button size="small" type="primary" text>
                      更改角色
                      <el-icon class="el-icon--right"><ArrowLeft style="transform: rotate(-90deg)" /></el-icon>
                    </el-button>
                    <template #dropdown>
                      <el-dropdown-menu>
                        <el-dropdown-item command="owner" :disabled="row.role === 'owner'">
                          <el-icon><Star /></el-icon>
                          所有者
                        </el-dropdown-item>
                        <el-dropdown-item
                          v-for="r in roles"
                          :key="r.role_key"
                          :command="r.role_key"
                          :disabled="row.role === r.role_key"
                        >
                          {{ r.role_name }}
                        </el-dropdown-item>
                      </el-dropdown-menu>
                    </template>
                  </el-dropdown>
                  <el-button size="small" type="danger" text @click="handleRemoveMember(row)">
                    移除
                  </el-button>
                </template>
              </el-table-column>
            </el-table>
            <TdEmptyState v-if="!membersLoading && members.length === 0" preset="no-data" title="暂无成员" />
          </div>
        </el-tab-pane>

        <!-- 项目角色 -->
        <el-tab-pane name="roles">
          <template #label>
            <span class="tab-label">
              <el-icon><Avatar /></el-icon>
              项目角色
            </span>
          </template>
          <div v-loading="rolesLoading" class="tab-content">
            <div class="section-header">
              <div class="section-title">
                <el-icon><Avatar /></el-icon>
                <span>角色列表</span>
                <el-tag size="small" type="info" style="margin-left: 8px">{{ roles.length }} 个</el-tag>
              </div>
              <el-button type="primary" @click="openRoleDialog">
                <el-icon><Plus /></el-icon>
                创建角色
              </el-button>
            </div>
            <div class="roles-grid">
              <div v-for="role in roles" :key="role.id" class="role-card" :class="getRoleIconClass(role.role_key)">
                <!-- 卡片头部 -->
                <div class="role-card-top">
                  <div class="role-icon" :class="getRoleIconClass(role.role_key)">
                    <el-icon :size="22"><Avatar /></el-icon>
                  </div>
                  <div class="role-card-title">
                    <div class="role-name-row">
                      <span class="role-name">{{ role.role_name }}</span>
                      <el-tag v-if="role.is_system" size="small" effect="plain" round>系统</el-tag>
                    </div>
                    <div class="role-key">{{ role.role_key }}</div>
                  </div>
                </div>
                <!-- 描述 -->
                <div class="role-desc">{{ role.description || '暂无描述' }}</div>
                <!-- 统计标签 -->
                <div class="role-stats">
                  <div class="role-stat-item">
                    <el-icon :size="13"><User /></el-icon>
                    <span>{{ role.member_count || 0 }} 成员</span>
                  </div>
                  <div class="role-stat-item">
                    <el-icon :size="13"><Lock /></el-icon>
                    <span>{{ role.permissions?.length || 0 }} 项权限</span>
                  </div>
                </div>
                <!-- 操作按钮 -->
                <div class="role-actions">
                  <el-button size="small" class="role-action-btn" @click="handleManageRoleMembers(role)">
                    <el-icon><User /></el-icon>
                    成员
                  </el-button>
                  <el-button size="small" class="role-action-btn" @click="handleConfigPermissions(role)">
                    <el-icon><Lock /></el-icon>
                    权限
                  </el-button>
                  <el-button v-if="!role.is_system" size="small" class="role-action-btn" @click="handleEditRole(role)">
                    <el-icon><Edit /></el-icon>
                    编辑
                  </el-button>
                  <el-button v-if="!role.is_system" size="small" type="danger" plain class="role-action-btn" @click="handleDeleteRole(role)">
                    <el-icon><Delete /></el-icon>
                    删除
                  </el-button>
                </div>
              </div>
            </div>
            <TdEmptyState v-if="!rolesLoading && roles.length === 0" preset="no-data" title="暂无角色" />
          </div>
        </el-tab-pane>

        <!-- 工单类型 -->
        <el-tab-pane name="issue-types">
          <template #label>
            <span class="tab-label">
              <el-icon><Tickets /></el-icon>
              工单类型
            </span>
          </template>
          <div v-loading="issueTypesLoading" class="tab-content">
            <div class="section-header">
              <div class="section-title">
                <el-icon><Tickets /></el-icon>
                <span>类型列表</span>
                <el-tag size="small" type="info" style="margin-left: 8px">{{ issueTypes.length }} 个</el-tag>
              </div>
              <el-button type="primary" @click="openIssueTypeDialog">
                <el-icon><Plus /></el-icon>
                添加类型
              </el-button>
            </div>
            <div class="issue-types-grid">
              <div v-for="type in issueTypes" :key="type.id" class="issue-type-card">
                <div class="type-icon" :style="{ background: type.color || '#3b82f6' }">
                  <el-icon :size="24">
                    <component :is="getIssueTypeIcon(type.icon)" />
                  </el-icon>
                </div>
                <div class="type-info">
                  <div class="type-name">{{ type.display_name }}</div>
                  <div class="type-key">{{ type.name }}</div>
                  <div v-if="type.description" class="type-desc">{{ type.description }}</div>
                </div>
                <div class="type-actions">
                  <el-button size="small" @click="handleEditIssueTypeClick(type)">
                    <el-icon><Edit /></el-icon>
                  </el-button>
                  <el-button size="small" type="danger" @click="handleDeleteIssueType(type)">
                    <el-icon><Delete /></el-icon>
                  </el-button>
                </div>
              </div>
            </div>
            <TdEmptyState v-if="!issueTypesLoading && issueTypes.length === 0" preset="no-data" title="暂无工单类型" />
          </div>
        </el-tab-pane>

        <!-- 字段配置 -->
        <el-tab-pane name="fields">
          <template #label>
            <span class="tab-label">
              <el-icon><Grid /></el-icon>
              字段配置
            </span>
          </template>
          <div class="tab-content">
            <FieldConfigTab :project-key="projectKey" />
          </div>
        </el-tab-pane>

        <!-- 通知渠道 -->
        <el-tab-pane name="notifications">
          <template #label>
            <span class="tab-label">
              <el-icon><Bell /></el-icon>
              通知渠道
            </span>
          </template>
          <div v-loading="channelsLoading" class="tab-content">
            <div class="section-header">
              <div class="section-title">
                <el-icon><Bell /></el-icon>
                <span>通知渠道</span>
                <el-tag size="small" type="info" style="margin-left: 8px">{{ channels.length }} 个</el-tag>
              </div>
              <el-button type="primary" @click="openChannelDialog">
                <el-icon><Plus /></el-icon>
                添加渠道
              </el-button>
            </div>

            <!-- 内置通知提示 -->
            <div class="notification-builtin-card">
              <div class="builtin-icon">
                <el-icon :size="20"><Connection /></el-icon>
              </div>
              <div class="builtin-content">
                <div class="builtin-title">站内实时通知</div>
                <div class="builtin-desc">WebSocket 实时推送默认开启，工单相关事件会自动通知到站内消息中心，无需额外配置。</div>
              </div>
              <el-tag type="success" size="small" effect="dark" class="builtin-tag">默认开启</el-tag>
            </div>

            <!-- 每日日报配置区 -->
            <div class="digest-card">
              <div class="digest-card-header">
                <div class="digest-card-title">
                  <el-icon :size="18"><Bell /></el-icon>
                  <span>定时日报</span>
                  <el-tag v-if="digestForm.enabled" type="success" size="small" effect="dark" class="builtin-tag">已启用</el-tag>
                  <el-tag v-else type="info" size="small" effect="plain" class="builtin-tag">未启用</el-tag>
                </div>
                <div class="digest-card-desc">每日定时推送项目「未完结工单」到订阅了「每日日报」事件的通知渠道，按指派人分组并 @ 对应人。</div>
              </div>
              <div class="digest-card-body">
                <el-form label-position="top" :model="digestForm" class="digest-form">
                  <el-form-item label="启用定时日报">
                    <el-switch
                      v-model="digestForm.enabled"
                      active-color="#10b981"
                      inactive-color="#d1d5db"
                    />
                  </el-form-item>
                  <el-form-item label="推送时间">
                    <el-time-picker
                      v-model="digestForm.timeValue"
                      format="HH:mm"
                      value-format="HH:mm"
                      placeholder="选择时间"
                      :clearable="false"
                      style="width: 160px"
                    />
                    <span class="digest-form-tip">每日固定时间推送（按下方时区计算）</span>
                  </el-form-item>
                  <el-form-item label="时区">
                    <el-select v-model="digestForm.tz" style="width: 240px">
                      <el-option label="Asia/Shanghai (UTC+8)" value="Asia/Shanghai" />
                      <el-option label="Asia/Tokyo (UTC+9)" value="Asia/Tokyo" />
                      <el-option label="Asia/Singapore (UTC+8)" value="Asia/Singapore" />
                      <el-option label="UTC" value="UTC" />
                      <el-option label="America/Los_Angeles" value="America/Los_Angeles" />
                      <el-option label="America/New_York" value="America/New_York" />
                      <el-option label="Europe/London" value="Europe/London" />
                    </el-select>
                  </el-form-item>
                  <el-form-item label="覆盖范围">
                    <el-radio-group v-model="digestForm.scope">
                      <el-radio value="all_open">全部未完结（含未指派单独分组）</el-radio>
                      <el-radio value="assigned_only">仅含已指派</el-radio>
                    </el-radio-group>
                  </el-form-item>
                  <el-form-item>
                    <el-button type="primary" :loading="digestSaving" @click="saveDigestConfig">
                      <el-icon><Check /></el-icon>
                      保存日报设置
                    </el-button>
                    <el-button :loading="digestRunning" :disabled="!digestForm.enabled" @click="triggerDigestNow">
                      <el-icon><Promotion /></el-icon>
                      立即触发一次
                    </el-button>
                    <div class="digest-form-tip" style="margin-top: 8px">
                      还需在下方至少一个通知渠道勾选「每日日报」事件才能收到推送
                    </div>
                  </el-form-item>
                </el-form>
              </div>
            </div>

            <!-- 外部渠道说明 -->
            <div class="notification-section-label">
              <span class="section-label-text">外部通知渠道</span>
              <span class="section-label-desc">工单创建、状态变更、指派、评论等事件会自动推送到已启用的渠道</span>
            </div>

            <!-- 渠道列表 -->
            <div v-if="channels.length > 0" class="channels-list">
              <div v-for="channel in channels" :key="channel.id" class="channel-row" :class="{ disabled: !channel.enabled }">
                <div class="channel-row-left">
                  <div class="channel-icon" :class="channel.channel_type">
                    <span v-if="channel.channel_type === 'lark'" class="channel-icon-text">飞书</span>
                    <span v-else-if="channel.channel_type === 'telegram'" class="channel-icon-text">TG</span>
                  </div>
                  <div class="channel-row-info">
                    <div class="channel-row-name">
                      <span>{{ channel.name }}</span>
                      <el-tag
                        :type="channel.channel_type === 'lark' ? 'primary' : 'info'"
                        size="small"
                        effect="plain"
                        class="channel-type-tag"
                      >
                        {{ channel.channel_type === 'lark' ? '飞书群机器人' : 'Telegram Bot' }}
                      </el-tag>
                    </div>
                    <div class="channel-row-detail">
                      <span class="channel-config-text">
                        <template v-if="channel.channel_type === 'lark'">
                          Webhook: {{ maskUrl((channel.config as any)?.webhook_url) }}
                        </template>
                        <template v-else-if="channel.channel_type === 'telegram'">
                          Chat ID: {{ (channel.config as any)?.chat_id || '-' }}
                        </template>
                      </span>
                    </div>
                    <div v-if="channel.event_types && channel.event_types.length > 0" class="channel-row-events">
                      <el-tag
                        v-for="ev in channel.event_types"
                        :key="ev"
                        size="small"
                        type="info"
                        effect="plain"
                        class="channel-event-tag"
                      >
                        {{ getNotificationEventLabel(ev) }}
                      </el-tag>
                    </div>
                  </div>
                </div>
                <div class="channel-row-right">
                  <span class="channel-status-indicator" :class="channel.enabled ? 'active' : 'inactive'">
                    <span class="status-dot"></span>
                    {{ channel.enabled ? '运行中' : '已停用' }}
                  </span>
                  <el-button
                    size="small"
                    type="success"
                    plain
                    :loading="testingChannelId === channel.id"
                    @click="handleTestChannel(channel)"
                  >
                    <el-icon><Promotion /></el-icon>
                    测试
                  </el-button>
                  <el-button size="small" plain @click="handleEditChannel(channel)">
                    <el-icon><Edit /></el-icon>
                    编辑
                  </el-button>
                  <el-button size="small" plain type="danger" @click="handleDeleteChannel(channel)">
                    <el-icon><Delete /></el-icon>
                    删除
                  </el-button>
                </div>
              </div>
            </div>

            <!-- 空状态 -->
            <div v-if="!channelsLoading && channels.length === 0" class="channels-empty">
              <div class="empty-illustration">
                <div class="empty-icon-wrapper">
                  <el-icon :size="48"><Bell /></el-icon>
                </div>
                <div class="empty-decorations">
                  <span class="deco deco-1"></span>
                  <span class="deco deco-2"></span>
                  <span class="deco deco-3"></span>
                </div>
              </div>
              <div class="empty-title">还没有配置外部通知渠道</div>
              <div class="empty-desc">添加飞书或 Telegram 通知渠道，让团队第一时间收到工单动态</div>
              <el-button type="primary" @click="openChannelDialog">
                <el-icon><Plus /></el-icon>
                添加第一个渠道
              </el-button>
            </div>
          </div>
        </el-tab-pane>

        <!-- 工作流配置 -->
        <el-tab-pane name="workflow">
          <template #label>
            <span class="tab-label">
              <el-icon><Connection /></el-icon>
              工作流配置
            </span>
          </template>
          <div v-loading="schemesLoading" class="tab-content">
            <div class="section-header">
              <div class="section-title">
                <el-icon><Connection /></el-icon>
                <span>工作流方案</span>
                <el-tag size="small" type="info" style="margin-left: 8px">{{ schemes.length }} 个</el-tag>
              </div>
              <el-button type="primary" @click="openSchemeDialog">
                <el-icon><Plus /></el-icon>
                添加配置
              </el-button>
            </div>

            <div class="workflow-scheme-tip">
              <el-icon><InfoFilled /></el-icon>
              <span>工作流方案将工单类型与工作流绑定。创建对应类型的工单时，会自动启动关联的工作流实例。</span>
            </div>

            <div v-if="schemes.length > 0" class="schemes-list">
              <div v-for="scheme in schemes" :key="scheme.id" class="scheme-row">
                <div class="scheme-row-left">
                  <div class="scheme-type-icon" :style="{ background: getIssueTypeColor(scheme.issue_type_id) }">
                    <el-icon><Tickets /></el-icon>
                  </div>
                  <div class="scheme-row-info">
                    <div class="scheme-row-name">
                      <span class="scheme-type-name">{{ scheme.issue_type_name || `类型#${scheme.issue_type_id}` }}</span>
                      <el-icon class="scheme-arrow"><ArrowLeft style="transform: rotate(180deg)" /></el-icon>
                      <span class="scheme-workflow-name">{{ scheme.workflow_name || `工作流#${scheme.workflow_id}` }}</span>
                    </div>
                    <div class="scheme-row-detail">
                      创建于 {{ formatSchemeTime(scheme.created_at) }}
                    </div>
                  </div>
                </div>
                <div class="scheme-row-right">
                  <el-button size="small" type="danger" plain @click="handleDeleteScheme(scheme)">
                    <el-icon><Delete /></el-icon>
                    删除
                  </el-button>
                </div>
              </div>
            </div>

            <TdEmptyState v-if="!schemesLoading && schemes.length === 0" preset="first-time" title="暂未配置工作流方案" description="绑定工单类型与工作流，自动启动流程实例">
              <el-button type="primary" @click="openSchemeDialog">
                <el-icon><Plus /></el-icon>
                添加第一个配置
              </el-button>
            </TdEmptyState>
          </div>
        </el-tab-pane>

        <!-- 危险操作 -->
        <el-tab-pane name="danger">
          <template #label>
            <span class="tab-label danger-tab-label">
              <el-icon><Warning /></el-icon>
              危险操作
            </span>
          </template>
          <div class="tab-content danger-content">
            <div class="danger-card">
              <div class="danger-card-title">
                <el-icon><Warning /></el-icon>
                <span>删除项目</span>
              </div>
              <p class="danger-card-desc">
                删除后将永久清理项目、工单及关联数据，此操作不可恢复。请务必确认当前项目不再使用。
              </p>
              <ul class="danger-card-list">
                <li>项目 Key：{{ basicForm.project_key || projectKey }}</li>
                <li>相关工单、工作流、角色、成员、通知将一并删除</li>
                <li>关联的告警记录将一并删除</li>
              </ul>
              <el-button type="danger" :loading="dangerDeleting" @click="handleDeleteProject">
                <el-icon><Delete /></el-icon>
                删除此项目
              </el-button>
            </div>
          </div>
        </el-tab-pane>
      </el-tabs>
    </el-card>

    <!-- 批量添加成员对话框 -->
    <el-dialog v-model="memberDialogVisible" title="添加成员" width="680px" destroy-on-close class="custom-dialog batch-member-dialog">
      <div class="batch-member-content">
        <!-- 搜索和角色选择 -->
        <div class="batch-member-toolbar">
          <el-input
            v-model="memberSearchKeyword"
            placeholder="搜索用户名、显示名称"
            clearable
            class="member-search-input"
          >
            <template #prefix>
              <el-icon><Search /></el-icon>
            </template>
          </el-input>
          <el-select v-model="memberForm.role" style="width: 140px">
            <el-option
              v-for="r in roles"
              :key="r.role_key"
              :label="r.role_name"
              :value="r.role_key"
            />
          </el-select>
        </div>
        <!-- 已选提示 -->
        <div v-if="selectedUserIds.length > 0" class="batch-member-selected-hint">
          已选择 <strong>{{ selectedUserIds.length }}</strong> 个用户
          <el-button type="primary" link size="small" @click="selectedUserIds = []">清空</el-button>
        </div>
        <!-- 用户列表 -->
        <div class="batch-member-list">
          <el-table
            ref="memberTableRef"
            :data="filteredAvailableUsers"
            max-height="400"
            row-key="id"
            size="small"
            @selection-change="handleMemberSelectionChange"
          >
            <el-table-column type="selection" width="40" :reserve-selection="true" />
            <el-table-column label="用户" min-width="200">
              <template #default="{ row }">
                <div class="user-option">
                  <div class="user-option-avatar">{{ row.display_name?.charAt(0) || '?' }}</div>
                  <div class="user-option-info">
                    <span class="user-option-name">{{ row.display_name }}</span>
                    <span class="user-option-username">@{{ row.username }}</span>
                  </div>
                </div>
              </template>
            </el-table-column>
          </el-table>
          <TdEmptyState v-if="filteredAvailableUsers.length === 0 && memberSearchKeyword" preset="no-result" title="无匹配用户" />
        </div>
      </div>
      <template #footer>
        <el-button @click="memberDialogVisible = false">取消</el-button>
        <el-button
          type="primary"
          :loading="addMemberLoading"
          :disabled="selectedUserIds.length === 0"
          @click="submitAddMember"
        >
          <el-icon><Check /></el-icon>
          添加{{ selectedUserIds.length > 0 ? ` (${selectedUserIds.length})` : '' }}
        </el-button>
      </template>
    </el-dialog>

    <!-- 创建/编辑角色对话框 -->
    <el-dialog
      v-model="roleDialogVisible"
      :title="isEditingRole ? '编辑角色' : '创建角色'"
      width="500px"
      destroy-on-close
      class="custom-dialog"
    >
      <el-form ref="roleFormRef" :model="roleForm" :rules="roleRules" label-width="80px">
        <el-form-item label="角色标识" prop="role_key">
          <el-input v-model="roleForm.role_key" placeholder="如: developers" :disabled="isEditingRole" />
          <div class="form-tip">小写字母和下划线，创建后不可修改</div>
        </el-form-item>
        <el-form-item label="角色名称" prop="role_name">
          <el-input v-model="roleForm.role_name" placeholder="如: 开发人员" />
        </el-form-item>
        <el-form-item label="描述">
          <el-input v-model="roleForm.description" type="textarea" :rows="3" placeholder="角色描述" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="roleDialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="roleSubmitLoading" @click="submitRole">
          <el-icon><Check /></el-icon>
          确定
        </el-button>
      </template>
    </el-dialog>

    <!-- 角色成员管理对话框 -->
    <el-dialog
      v-model="memberMgmtDialogVisible"
      :title="`${currentMgmtRole?.role_name} - 成员管理`"
      width="600px"
      destroy-on-close
      class="custom-dialog"
      @close="closeMemberMgmtDialog"
    >
      <div class="member-mgmt-content">
        <div class="member-mgmt-add">
          <el-select v-model="selectedRoleMemberUserId" placeholder="选择用户" filterable style="flex: 1" popper-class="user-select-popper">
            <el-option v-for="u in availableRoleUsers" :key="u.id" :label="u.display_name" :value="u.id">
              <div class="user-option">
                <div class="user-option-avatar">{{ u.display_name?.charAt(0) || '?' }}</div>
                <div class="user-option-info">
                  <span class="user-option-name">{{ u.display_name }}</span>
                  <span class="user-option-username">@{{ u.username }}</span>
                </div>
              </div>
            </el-option>
          </el-select>
          <el-button type="primary" :loading="addRoleMemberLoading" :disabled="!selectedRoleMemberUserId" @click="handleAddRoleMember">
            <el-icon><Plus /></el-icon>
            添加成员
          </el-button>
        </div>
        <el-divider />
        <div v-loading="roleMembersLoading" class="member-mgmt-list">
          <div v-for="member in roleMembers" :key="member.id" class="member-mgmt-item">
            <div class="member-mgmt-info">
              <div class="member-mgmt-avatar">{{ member.user?.display_name?.charAt(0) || '?' }}</div>
              <div class="member-mgmt-details">
                <span class="member-mgmt-name">{{ member.user?.display_name }}</span>
                <span class="member-mgmt-username">@{{ member.user?.username }}</span>
              </div>
            </div>
            <el-button size="small" type="danger" text @click="handleRemoveRoleMember(member)">
              <el-icon><Delete /></el-icon>
              移除
            </el-button>
          </div>
          <TdEmptyState v-if="!roleMembersLoading && roleMembers.length === 0" preset="no-data" title="暂无成员" />
        </div>
      </div>
    </el-dialog>

    <!-- 角色权限配置对话框 -->
    <el-dialog
      v-model="permDialogVisible"
      :title="`${currentPermRole?.role_name} - 权限配置`"
      width="640px"
      destroy-on-close
      class="custom-dialog"
    >
      <div v-loading="permLoading" class="perm-config-content">
        <div v-for="group in PERMISSION_GROUPS" :key="group.module" class="perm-group">
          <div class="perm-group-header">
            <el-checkbox
              :model-value="isModuleAllChecked(group)"
              :indeterminate="isModuleIndeterminate(group)"
              @change="handleToggleModuleAll(group)"
            >
              {{ group.module }}
            </el-checkbox>
          </div>
          <div class="perm-group-items">
            <el-checkbox
              v-for="perm in group.permissions"
              :key="perm.key"
              :model-value="rolePermissions.includes(perm.key)"
              @change="(val: string | number | boolean) => handleTogglePermission(perm.key, !!val)"
            >
              {{ perm.label }}
            </el-checkbox>
          </div>
        </div>
      </div>
      <template #footer>
        <el-button @click="permDialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="permSaveLoading" @click="handleSavePermissions">
          <el-icon><Check /></el-icon>
          保存
        </el-button>
      </template>
    </el-dialog>

    <!-- 创建/编辑工单类型对话框 -->
    <el-dialog
      v-model="issueTypeDialogVisible"
      :title="isEditingIssueType ? '编辑工单类型' : '创建工单类型'"
      width="560px"
      destroy-on-close
      class="custom-dialog"
    >
      <el-form ref="issueTypeFormRef" :model="issueTypeForm" :rules="issueTypeRules" label-width="80px">
        <el-form-item label="类型标识" prop="name">
          <el-input v-model="issueTypeForm.name" placeholder="如: epic, bug, task" :disabled="isEditingIssueType" />
          <div class="form-tip-small">
            <el-icon><InfoFilled /></el-icon>
            <span>内部标识，只能包含小写字母，创建后不可修改。例如: epic, bug, task</span>
          </div>
        </el-form-item>
        <el-form-item label="显示名称" prop="display_name">
          <el-input v-model="issueTypeForm.display_name" placeholder="如: Epic, 缺陷, 任务" />
          <div class="form-tip-small">用户看到的名称，可以使用任意字符</div>
        </el-form-item>
        <el-form-item label="图标">
          <div class="icon-selector">
            <div
              v-for="iconItem in availableIcons"
              :key="iconItem.name"
              class="icon-option"
              :class="{ active: issueTypeForm.icon === iconItem.name }"
              @click="issueTypeForm.icon = iconItem.name"
            >
              <el-icon :size="20">
                <component :is="iconItem.icon" />
              </el-icon>
              <span class="icon-label">{{ iconItem.label }}</span>
            </div>
          </div>
        </el-form-item>
        <el-form-item label="颜色">
          <div class="color-selector">
            <el-color-picker v-model="issueTypeForm.color" />
            <div class="color-presets">
              <div
                v-for="color in ['#3b82f6', '#10b981', '#f59e0b', '#ef4444', '#8b5cf6', '#ec4899', '#6b7280', '#06b6d4']"
                :key="color"
                class="color-preset"
                :style="{ background: color }"
                :class="{ active: issueTypeForm.color === color }"
                @click="issueTypeForm.color = color"
              ></div>
            </div>
          </div>
        </el-form-item>
        <el-form-item label="描述">
          <el-input v-model="issueTypeForm.description" type="textarea" :rows="3" placeholder="类型描述" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="issueTypeDialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="issueTypeSubmitLoading" @click="submitIssueType">
          <el-icon><Check /></el-icon>
          确定
        </el-button>
      </template>
    </el-dialog>

    <!-- 创建/编辑通知渠道对话框 -->
    <el-dialog
      v-model="channelDialogVisible"
      :title="isEditingChannel ? '编辑通知渠道' : '添加通知渠道'"
      width="600px"
      destroy-on-close
      class="custom-dialog channel-dialog"
    >
      <el-form ref="channelFormRef" :model="channelForm" :rules="channelRules" label-position="top">
        <!-- 渠道类型选择 - 卡片式 -->
        <el-form-item label="选择渠道类型" prop="channel_type" class="channel-type-form-item">
          <div class="channel-type-cards" :class="{ disabled: isEditingChannel }">
            <div
              class="channel-type-card"
              :class="{ active: channelForm.channel_type === 'lark', disabled: isEditingChannel }"
              @click="!isEditingChannel && (channelForm.channel_type = 'lark')"
            >
              <div class="type-card-icon lark">
                <span>飞书</span>
              </div>
              <div class="type-card-content">
                <div class="type-card-name">飞书群机器人</div>
                <div class="type-card-desc">通过 Webhook 推送消息到飞书群</div>
              </div>
              <div v-if="channelForm.channel_type === 'lark'" class="type-card-check">
                <el-icon><CircleCheck /></el-icon>
              </div>
            </div>
            <div
              class="channel-type-card"
              :class="{ active: channelForm.channel_type === 'telegram', disabled: isEditingChannel }"
              @click="!isEditingChannel && (channelForm.channel_type = 'telegram')"
            >
              <div class="type-card-icon telegram">
                <span>TG</span>
              </div>
              <div class="type-card-content">
                <div class="type-card-name">Telegram Bot</div>
                <div class="type-card-desc">通过 Bot API 推送消息到 Telegram 群组</div>
              </div>
              <div v-if="channelForm.channel_type === 'telegram'" class="type-card-check">
                <el-icon><CircleCheck /></el-icon>
              </div>
            </div>
          </div>
        </el-form-item>

        <!-- 基本信息 -->
        <div class="dialog-section">
          <div class="dialog-section-title">基本信息</div>
          <el-form-item label="渠道名称" prop="name">
            <el-input v-model="channelForm.name" placeholder="如: 运维飞书群、开发 Telegram 群" maxlength="100" show-word-limit>
              <template #prefix>
                <el-icon><Bell /></el-icon>
              </template>
            </el-input>
          </el-form-item>
          <el-form-item label="启用状态">
            <div class="enable-switch-row">
              <el-switch
                v-model="channelForm.enabled"
                active-color="#10b981"
                inactive-color="#d1d5db"
              />
              <span class="enable-switch-label" :class="{ active: channelForm.enabled }">
                {{ channelForm.enabled ? '创建后立即启用' : '创建后暂不启用' }}
              </span>
            </div>
          </el-form-item>
          <el-form-item label="订阅的通知事件" prop="event_types">
            <div class="event-types-wrapper">
              <el-checkbox-group v-model="channelForm.event_types" class="event-types-group">
                <el-checkbox
                  v-for="opt in NOTIFICATION_EVENT_OPTIONS"
                  :key="opt.value"
                  :value="opt.value"
                  class="event-type-checkbox"
                >
                  <span class="event-type-label">{{ opt.label }}</span>
                  <span class="event-type-desc">{{ opt.description }}</span>
                </el-checkbox>
              </el-checkbox-group>
              <div class="form-tip-small">
                <el-icon><InfoFilled /></el-icon>
                <span>勾选哪些事件就只会推送对应事件，至少需选择一项</span>
              </div>
            </div>
          </el-form-item>
          <el-form-item v-if="channelForm.channel_type === 'lark'" label="@ 全员">
            <div class="enable-switch-row">
              <el-switch
                v-model="channelForm.mention_all"
                active-color="#10b981"
                inactive-color="#d1d5db"
              />
              <span class="enable-switch-label" :class="{ active: channelForm.mention_all }">
                {{ channelForm.mention_all ? '通知时 @ 群内所有成员' : '不 @ 全员' }}
              </span>
            </div>
            <div class="form-tip-small">
              <el-icon><InfoFilled /></el-icon>
              <span>开启后飞书卡片末尾会追加 @ 所有人；Telegram 不支持 @ 全员，此开关对 Telegram 渠道无效</span>
            </div>
          </el-form-item>
        </div>

        <!-- 飞书配置 -->
        <div v-if="channelForm.channel_type === 'lark'" class="dialog-section">
          <div class="dialog-section-title">
            <span class="section-title-icon lark">飞书</span>
            连接配置
          </div>
          <el-form-item label="Webhook URL" prop="lark_webhook_url">
            <el-input v-model="channelForm.lark_webhook_url" placeholder="https://open.feishu.cn/open-apis/bot/v2/hook/xxx">
              <template #prefix>
                <el-icon><Connection /></el-icon>
              </template>
            </el-input>
            <div class="form-tip-small">
              <el-icon><InfoFilled /></el-icon>
              <span>在飞书群设置 → 群机器人 → 添加自定义机器人，复制 Webhook 地址</span>
            </div>
          </el-form-item>
          <el-form-item label="签名密钥（可选）">
            <el-input v-model="channelForm.lark_secret" placeholder="开启签名校验时填写" show-password>
              <template #prefix>
                <el-icon><Key /></el-icon>
              </template>
            </el-input>
            <div class="form-tip-small">
              <el-icon><InfoFilled /></el-icon>
              <span>如果机器人开启了签名校验，请填写对应的密钥</span>
            </div>
          </el-form-item>
        </div>

        <!-- Telegram 配置 -->
        <div v-if="channelForm.channel_type === 'telegram'" class="dialog-section">
          <div class="dialog-section-title">
            <span class="section-title-icon telegram">TG</span>
            连接配置
          </div>
          <el-form-item label="Bot Token" prop="telegram_bot_token">
            <el-input v-model="channelForm.telegram_bot_token" placeholder="通过 @BotFather 创建 Bot 获取 Token" show-password>
              <template #prefix>
                <el-icon><Key /></el-icon>
              </template>
            </el-input>
            <div class="form-tip-small">
              <el-icon><InfoFilled /></el-icon>
              <span>在 Telegram 中搜索 @BotFather，发送 /newbot 创建机器人并获取 Token</span>
            </div>
          </el-form-item>
          <el-form-item label="Chat ID" prop="telegram_chat_id">
            <el-input v-model="channelForm.telegram_chat_id" placeholder="群组或频道的 Chat ID（如 -1001234567890）">
              <template #prefix>
                <el-icon><Promotion /></el-icon>
              </template>
            </el-input>
            <div class="form-tip-small">
              <el-icon><InfoFilled /></el-icon>
              <span>将 Bot 添加到群组后，通过 @userinfobot 或 getUpdates API 获取 Chat ID</span>
            </div>
          </el-form-item>
        </div>
      </el-form>
      <template #footer>
        <div class="dialog-footer">
          <el-button size="large" @click="channelDialogVisible = false">取消</el-button>
          <el-button type="primary" :loading="channelSubmitLoading" size="large" @click="submitChannel">
            <el-icon><Check /></el-icon>
            {{ isEditingChannel ? '保存更改' : '添加渠道' }}
          </el-button>
        </div>
      </template>
    </el-dialog>

    <!-- 添加工作流方案对话框 -->
    <el-dialog v-model="schemeDialogVisible" title="添加工作流配置" width="520px" destroy-on-close class="custom-dialog">
      <el-form ref="schemeFormRef" :model="schemeForm" :rules="schemeRules" label-width="100px">
        <el-form-item label="工单类型" prop="issue_type_id">
          <el-select v-model="schemeForm.issue_type_id" placeholder="请选择工单类型" style="width: 100%">
            <el-option
              v-for="t in availableIssueTypes"
              :key="t.id"
              :label="t.display_name"
              :value="t.id"
            >
              <div class="scheme-option">
                <div class="scheme-option-icon" :style="{ background: t.color || '#3b82f6' }">
                  <el-icon :size="14"><component :is="getIssueTypeIcon(t.icon)" /></el-icon>
                </div>
                <span>{{ t.display_name }}</span>
              </div>
            </el-option>
          </el-select>
        </el-form-item>
        <el-form-item label="工作流" prop="workflow_id">
          <el-select v-model="schemeForm.workflow_id" placeholder="请选择工作流" style="width: 100%" filterable>
            <el-option
              v-for="w in allWorkflows"
              :key="w.id"
              :label="w.name"
              :value="w.id"
            >
              <div class="scheme-option">
                <div class="scheme-option-icon workflow">
                  <el-icon :size="14"><Connection /></el-icon>
                </div>
                <div class="scheme-option-info">
                  <span class="scheme-option-name">{{ w.name }}</span>
                  <span v-if="w.description" class="scheme-option-desc">{{ w.description }}</span>
                </div>
              </div>
            </el-option>
          </el-select>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="schemeDialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="schemeSubmitLoading" @click="submitScheme">
          <el-icon><Check /></el-icon>
          确定
        </el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, computed, onMounted, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage, ElMessageBox, type FormInstance, type FormRules } from 'element-plus'
import {
  Setting,
  Plus,
  User,
  UserFilled,
  Avatar,
  ArrowLeft,
  Document,
  InfoFilled,
  Key,
  Folder,
  Check,
  Star,
  Delete,
  Edit,
  Tickets,
  Lightning,
  CircleCheck,
  Warning,
  Flag,
  Promotion,
  Tools,
  Connection,
  Memo,
  Grid,
  Bell,
  Search,
  Lock,
} from '@element-plus/icons-vue'
import FieldConfigTab from './components/FieldConfigTab.vue'
import {
  getProjectDetail,
  deleteProject,
  updateProject,
  getProjectMembers,
  addProjectMember,
  updateProjectMemberRole,
  removeProjectMember,
  getProjectRoles,
  createProjectRole,
  updateProjectRole,
  deleteProjectRole,
  getProjectIssueTypes,
  createProjectIssueType,
  updateProjectIssueType,
  deleteProjectIssueType,
  getNotificationChannels,
  createNotificationChannel,
  updateNotificationChannel,
  deleteNotificationChannel,
  testNotificationChannel,
  runDailyDigest,
  getRoleMembers,
  addRoleMember,
  removeRoleMember,
  getRolePermissions,
  setRolePermissions,
} from '@/api/project'
import { getWorkflowSchemes, createWorkflowScheme, deleteWorkflowScheme, getWorkflowList } from '@/api/workflow'
import type { WorkflowScheme, Workflow } from '@/types/workflow'
import { getAllUsers } from '@/api/user'
import type {
  ProjectMember,
  ProjectRole,
  ProjectIssueType,
  UpdateProjectRequest,
  CreateProjectRoleRequest,
  CreateIssueTypeRequest,
  NotificationChannel,
  NotificationEventType,
  ProjectRoleMember,
} from '@/types/project'
import {
  NOTIFICATION_EVENT_OPTIONS,
  DEFAULT_NOTIFICATION_EVENTS,
  getNotificationEventLabel,
} from '@/constants/notification'
import type { UserOption } from '@/types/user'

const route = useRoute()
const router = useRouter()
const projectKey = computed(() => route.params.key as string)

// 从 URL 查询参数中获取 tab，如果没有则默认为 'basic'
const activeTab = ref((route.query.tab as string) || 'basic')
const loading = ref(false)
const saveLoading = ref(false)
const dangerDeleting = ref(false)
const users = ref<UserOption[]>([])

// 监听 tab 变化，更新 URL
watch(activeTab, (newTab) => {
  router.replace({
    query: { ...route.query, tab: newTab },
  })
})

// 基本信息
const basicFormRef = ref<FormInstance>()
const basicForm = reactive<UpdateProjectRequest & { project_key?: string }>({
  project_key: '',
  name: '',
  description: '',
  lead_user_id: undefined,
  status: 1,
})
const basicRules: FormRules = {
  name: [{ required: true, message: '请输入项目名称', trigger: 'blur' }],
}

// 项目成员
const membersLoading = ref(false)
const members = ref<ProjectMember[]>([])
const memberDialogVisible = ref(false)
const addMemberLoading = ref(false)
const memberTableRef = ref()
const memberSearchKeyword = ref('')
const selectedUserIds = ref<number[]>([])
const memberForm = reactive({
  role: 'developers',
})

const availableUsers = computed(() => {
  const memberIds = members.value.map((m) => m.user_id)
  return users.value.filter((u) => !memberIds.includes(u.id))
})

const filteredAvailableUsers = computed(() => {
  const keyword = memberSearchKeyword.value.trim().toLowerCase()
  if (!keyword) return availableUsers.value
  return availableUsers.value.filter(
    (u) =>
      u.display_name?.toLowerCase().includes(keyword) ||
      u.username?.toLowerCase().includes(keyword),
  )
})

const handleMemberSelectionChange = (rows: UserOption[]) => {
  selectedUserIds.value = rows.map((r) => r.id)
}

// 项目角色
const rolesLoading = ref(false)
const roles = ref<ProjectRole[]>([])
const roleDialogVisible = ref(false)
const isEditingRole = ref(false)
const editingRoleId = ref<number | null>(null)
const roleSubmitLoading = ref(false)
const roleFormRef = ref<FormInstance>()
const roleForm = reactive<CreateProjectRoleRequest>({
  role_key: '',
  role_name: '',
  description: '',
})
const roleRules: FormRules = {
  role_key: [
    { required: true, message: '请输入角色标识', trigger: 'blur' },
    { pattern: /^[a-z_]+$/, message: '只能包含小写字母和下划线', trigger: 'blur' },
  ],
  role_name: [{ required: true, message: '请输入角色名称', trigger: 'blur' }],
}

// 角色成员管理
const memberMgmtDialogVisible = ref(false)
const currentMgmtRole = ref<ProjectRole | null>(null)
const roleMembers = ref<ProjectRoleMember[]>([])
const roleMembersLoading = ref(false)
const selectedRoleMemberUserId = ref<number | null>(null)
const addRoleMemberLoading = ref(false)

const availableRoleUsers = computed(() => {
  const memberIds = roleMembers.value.map((m) => m.user_id)
  return users.value.filter((u) => !memberIds.includes(u.id))
})

// 角色权限配置
const PERMISSION_GROUPS = [
  { module: '项目', permissions: [{ key: 'project:view', label: '访问项目' }, { key: 'project:manage', label: '管理设置' }] },
  { module: '工单', permissions: [{ key: 'issue:view', label: '查看工单' }, { key: 'issue:create', label: '创建工单' }, { key: 'issue:edit', label: '编辑工单' }, { key: 'issue:delete', label: '删除工单' }, { key: 'issue:assign', label: '指派工单' }] },
  { module: '成员', permissions: [{ key: 'member:view', label: '查看成员' }, { key: 'member:manage', label: '管理成员' }] },
  { module: '角色', permissions: [{ key: 'role:view', label: '查看角色' }, { key: 'role:manage', label: '管理角色和权限' }] },
  { module: '工作流', permissions: [{ key: 'workflow:view', label: '查看工作流' }, { key: 'workflow:manage', label: '管理工作流' }] },
  { module: '告警', permissions: [{ key: 'alert:view', label: '查看告警' }, { key: 'alert:manage', label: '管理告警规则' }] },
]
const permDialogVisible = ref(false)
const currentPermRole = ref<ProjectRole | null>(null)
const rolePermissions = ref<string[]>([])
const permLoading = ref(false)
const permSaveLoading = ref(false)

const handleConfigPermissions = async (role: ProjectRole) => {
  currentPermRole.value = role
  permDialogVisible.value = true
  permLoading.value = true
  try {
    const { data } = await getRolePermissions(projectKey.value, role.id)
    rolePermissions.value = data.data || []
  } catch {
    rolePermissions.value = role.permissions || []
  } finally {
    permLoading.value = false
  }
}

const handleTogglePermission = (key: string, checked: boolean) => {
  if (checked) {
    if (!rolePermissions.value.includes(key)) {
      rolePermissions.value = [...rolePermissions.value, key]
    }
  } else {
    rolePermissions.value = rolePermissions.value.filter((k) => k !== key)
  }
}

const isModuleAllChecked = (group: typeof PERMISSION_GROUPS[number]) => {
  return group.permissions.every((p) => rolePermissions.value.includes(p.key))
}

const isModuleIndeterminate = (group: typeof PERMISSION_GROUPS[number]) => {
  const checked = group.permissions.filter((p) => rolePermissions.value.includes(p.key)).length
  return checked > 0 && checked < group.permissions.length
}

const handleToggleModuleAll = (group: typeof PERMISSION_GROUPS[number]) => {
  const allChecked = isModuleAllChecked(group)
  const keys = group.permissions.map((p) => p.key)
  if (allChecked) {
    rolePermissions.value = rolePermissions.value.filter((k) => !keys.includes(k))
  } else {
    const newPerms = [...rolePermissions.value]
    for (const key of keys) {
      if (!newPerms.includes(key)) {
        newPerms.push(key)
      }
    }
    rolePermissions.value = newPerms
  }
}

const handleSavePermissions = async () => {
  if (!currentPermRole.value) return
  permSaveLoading.value = true
  try {
    await setRolePermissions(projectKey.value, currentPermRole.value.id, { permissions: rolePermissions.value })
    ElMessage.success('权限保存成功')
    permDialogVisible.value = false
    loadRoles()
  } catch {
    ElMessage.error('权限保存失败')
  } finally {
    permSaveLoading.value = false
  }
}

// 工单类型
const issueTypesLoading = ref(false)
const issueTypes = ref<ProjectIssueType[]>([])
const issueTypeDialogVisible = ref(false)
const isEditingIssueType = ref(false)
const editingIssueTypeId = ref<number | null>(null)
const issueTypeSubmitLoading = ref(false)
const issueTypeFormRef = ref<FormInstance>()
const issueTypeForm = reactive<CreateIssueTypeRequest>({
  name: '',
  display_name: '',
  description: '',
  icon: 'task',
  color: '#3b82f6',
})
const issueTypeRules: FormRules = {
  name: [
    { required: true, message: '请输入类型标识', trigger: 'blur' },
    { pattern: /^[a-z]+$/, message: '只能包含小写字母', trigger: 'blur' },
  ],
  display_name: [{ required: true, message: '请输入显示名称', trigger: 'blur' }],
}

// 通知渠道
const channelsLoading = ref(false)
const channels = ref<NotificationChannel[]>([])
const channelDialogVisible = ref(false)
const isEditingChannel = ref(false)
const editingChannelId = ref<number | null>(null)
const channelSubmitLoading = ref(false)
const testingChannelId = ref<number | null>(null)
const channelFormRef = ref<FormInstance>()
const channelForm = reactive({
  channel_type: 'lark' as 'lark' | 'telegram',
  name: '',
  enabled: true,
  mention_all: false,
  event_types: [...DEFAULT_NOTIFICATION_EVENTS] as NotificationEventType[],
  lark_webhook_url: '',
  lark_secret: '',
  telegram_bot_token: '',
  telegram_chat_id: '',
})
const channelRules: FormRules = {
  channel_type: [{ required: true, message: '请选择渠道类型', trigger: 'change' }],
  name: [{ required: true, message: '请输入渠道名称', trigger: 'blur' }],
  event_types: [
    {
      type: 'array',
      required: true,
      message: '请至少选择一种通知事件',
      trigger: 'change',
      validator: (_rule, value, callback) => {
        if (!Array.isArray(value) || value.length === 0) {
          callback(new Error('请至少选择一种通知事件'))
          return
        }
        callback()
      },
    },
  ],
  lark_webhook_url: [{ required: true, message: '请输入飞书 Webhook URL', trigger: 'blur' }],
  telegram_bot_token: [{ required: true, message: '请输入 Telegram Bot Token', trigger: 'blur' }],
  telegram_chat_id: [{ required: true, message: '请输入 Telegram Chat ID', trigger: 'blur' }],
}

// 定时日报配置
const digestForm = reactive({
  enabled: false,
  timeValue: '09:00',
  cron: '0 9 * * *',
  tz: 'Asia/Shanghai',
  scope: 'all_open' as 'all_open' | 'assigned_only',
})
const digestSaving = ref(false)
const digestRunning = ref(false)

const cronToTime = (cron: string): string => {
  const parts = cron.trim().split(/\s+/)
  if (parts.length < 2) return '09:00'
  const mm = String(parts[0]).padStart(2, '0')
  const hh = String(parts[1]).padStart(2, '0')
  return `${hh}:${mm}`
}
const timeToCron = (t: string): string => {
  const [hh, mm] = t.split(':').map((s) => parseInt(s, 10))
  return `${isNaN(mm) ? 0 : mm} ${isNaN(hh) ? 9 : hh} * * *`
}

const saveDigestConfig = async () => {
  digestSaving.value = true
  try {
    await updateProject(projectKey.value, {
      daily_digest_enabled: digestForm.enabled,
      daily_digest_cron: timeToCron(digestForm.timeValue),
      daily_digest_tz: digestForm.tz,
      daily_digest_scope: digestForm.scope,
    })
    ElMessage.success('日报设置已保存')
    loadProjectDetail()
  } catch (e: any) {
    ElMessage.error(e?.response?.data?.message || '保存失败')
  } finally {
    digestSaving.value = false
  }
}

const triggerDigestNow = async () => {
  digestRunning.value = true
  try {
    await runDailyDigest(projectKey.value)
    ElMessage.success('日报已触发，请查看订阅了「每日日报」事件的渠道')
  } catch (e: any) {
    ElMessage.error(e?.response?.data?.message || '触发失败')
  } finally {
    digestRunning.value = false
  }
}

const loadProjectDetail = async () => {
  loading.value = true
  try {
    const { data } = await getProjectDetail(projectKey.value, { _redirectOn404: true })
    const project = data.data
    Object.assign(basicForm, {
      project_key: project.project_key,
      name: project.name,
      description: project.description,
      lead_user_id: project.lead_user_id,
      status: project.status,
    })
    // 回填日报配置
    digestForm.enabled = project.daily_digest_enabled === true
    digestForm.cron = project.daily_digest_cron || '0 9 * * *'
    digestForm.timeValue = cronToTime(digestForm.cron)
    digestForm.tz = project.daily_digest_tz || 'Asia/Shanghai'
    digestForm.scope = (project.daily_digest_scope as 'all_open' | 'assigned_only') || 'all_open'
  } catch {
    // ignored
  } finally {
    loading.value = false
  }
}

const loadUsers = async () => {
  try {
    const { data } = await getAllUsers()
    users.value = data.data
  } catch {
    // ignored
  }
}

const loadMembers = async () => {
  membersLoading.value = true
  try {
    const { data } = await getProjectMembers(projectKey.value)
    members.value = data.data
  } catch {
    // ignored
  } finally {
    membersLoading.value = false
  }
}

const loadRoles = async () => {
  rolesLoading.value = true
  try {
    const { data } = await getProjectRoles(projectKey.value)
    roles.value = data.data
  } catch {
    // ignored
  } finally {
    rolesLoading.value = false
  }
}

const loadIssueTypes = async () => {
  issueTypesLoading.value = true
  try {
    const { data } = await getProjectIssueTypes(projectKey.value)
    issueTypes.value = data.data
  } catch {
    // ignored
  } finally {
    issueTypesLoading.value = false
  }
}

const loadChannels = async () => {
  channelsLoading.value = true
  try {
    const { data } = await getNotificationChannels(projectKey.value)
    channels.value = data.data || []
  } catch {
    // ignored
  } finally {
    channelsLoading.value = false
  }
}

const openChannelDialog = () => {
  isEditingChannel.value = false
  editingChannelId.value = null
  Object.assign(channelForm, {
    channel_type: 'lark',
    name: '',
    enabled: true,
    mention_all: false,
    event_types: [...DEFAULT_NOTIFICATION_EVENTS],
    lark_webhook_url: '',
    lark_secret: '',
    telegram_bot_token: '',
    telegram_chat_id: '',
  })
  channelDialogVisible.value = true
}

const handleEditChannel = (channel: NotificationChannel) => {
  isEditingChannel.value = true
  editingChannelId.value = channel.id
  const config = channel.config as any
  const existing = Array.isArray(channel.event_types) && channel.event_types.length > 0
    ? [...channel.event_types]
    : [...DEFAULT_NOTIFICATION_EVENTS]
  Object.assign(channelForm, {
    channel_type: channel.channel_type,
    name: channel.name,
    enabled: channel.enabled,
    mention_all: channel.mention_all === true,
    event_types: existing,
    lark_webhook_url: channel.channel_type === 'lark' ? (config?.webhook_url || '') : '',
    lark_secret: '', // 密钥不回显
    telegram_bot_token: '', // Token 不回显
    telegram_chat_id: channel.channel_type === 'telegram' ? (config?.chat_id || '') : '',
  })
  channelDialogVisible.value = true
}

const handleDeleteChannel = async (channel: NotificationChannel) => {
  try {
    await ElMessageBox.confirm(`确定要删除通知渠道 "${channel.name}" 吗？`, '删除确认', {
      type: 'warning',
    })
    await deleteNotificationChannel(projectKey.value, channel.id)
    ElMessage.success('删除成功')
    loadChannels()
  } catch (error) {
    if (error !== 'cancel') {
      ElMessage.error('删除失败')
    }
  }
}

const handleTestChannel = async (channel: NotificationChannel) => {
  testingChannelId.value = channel.id
  try {
    await testNotificationChannel(projectKey.value, channel.id)
    ElMessage.success('测试消息发送成功，请检查对应渠道')
  } catch (error: any) {
    ElMessage.error(error?.response?.data?.message || '测试发送失败')
  } finally {
    testingChannelId.value = null
  }
}

const submitChannel = async () => {
  if (!channelFormRef.value) return
  await channelFormRef.value.validate(async (valid) => {
    if (!valid) return
    channelSubmitLoading.value = true
    try {
      // 构建 config
      let config: any = {}
      if (channelForm.channel_type === 'lark') {
        config = {
          webhook_url: channelForm.lark_webhook_url,
        }
        if (channelForm.lark_secret) {
          config.secret = channelForm.lark_secret
        }
      } else if (channelForm.channel_type === 'telegram') {
        config = {
          chat_id: channelForm.telegram_chat_id,
        }
        if (channelForm.telegram_bot_token) {
          config.bot_token = channelForm.telegram_bot_token
        }
      }

      if (isEditingChannel.value && editingChannelId.value) {
        const updateData: any = {
          name: channelForm.name,
          enabled: channelForm.enabled,
          event_types: channelForm.event_types,
          mention_all: channelForm.mention_all,
        }
        // 只有填写了配置才更新 config
        const hasConfig = channelForm.channel_type === 'lark'
          ? channelForm.lark_webhook_url
          : channelForm.telegram_chat_id
        if (hasConfig) {
          updateData.config = config
        }
        await updateNotificationChannel(projectKey.value, editingChannelId.value, updateData)
        ElMessage.success('更新成功')
      } else {
        await createNotificationChannel(projectKey.value, {
          channel_type: channelForm.channel_type,
          name: channelForm.name,
          config,
          event_types: channelForm.event_types,
          mention_all: channelForm.mention_all,
          enabled: channelForm.enabled,
        })
        ElMessage.success('创建成功')
      }
      channelDialogVisible.value = false
      loadChannels()
    } catch (error: any) {
      ElMessage.error(error?.response?.data?.message || (isEditingChannel.value ? '更新失败' : '创建失败'))
    } finally {
      channelSubmitLoading.value = false
    }
  })
}

const maskUrl = (url: string | undefined) => {
  if (!url) return '-'
  try {
    const u = new URL(url)
    const path = u.pathname
    if (path.length > 20) {
      return u.origin + path.substring(0, 10) + '...' + path.substring(path.length - 6)
    }
    return url
  } catch {
    if (url.length > 30) {
      return url.substring(0, 15) + '...' + url.substring(url.length - 10)
    }
    return url
  }
}

// 工作流方案
const schemesLoading = ref(false)
const schemes = ref<WorkflowScheme[]>([])
const allWorkflows = ref<Workflow[]>([])
const schemeDialogVisible = ref(false)
const schemeSubmitLoading = ref(false)
const schemeFormRef = ref<FormInstance>()
const schemeForm = reactive({
  issue_type_id: undefined as number | undefined,
  workflow_id: undefined as number | undefined,
})
const schemeRules: FormRules = {
  issue_type_id: [{ required: true, message: '请选择工单类型', trigger: 'change' }],
  workflow_id: [{ required: true, message: '请选择工作流', trigger: 'change' }],
}

// 已绑定的工单类型不再显示
const availableIssueTypes = computed(() => {
  const boundTypeIds = schemes.value.map((s) => s.issue_type_id)
  return issueTypes.value.filter((t) => !boundTypeIds.includes(t.id))
})

const loadSchemes = async () => {
  schemesLoading.value = true
  try {
    const { data } = await getWorkflowSchemes(projectKey.value)
    schemes.value = (data as any).data || []
  } catch {
    // ignored
  } finally {
    schemesLoading.value = false
  }
}

const loadAllWorkflows = async () => {
  try {
    const { data } = await getWorkflowList()
    allWorkflows.value = (data as any).data?.items || []
  } catch {
    // ignored
  }
}

const openSchemeDialog = () => {
  schemeForm.issue_type_id = undefined
  schemeForm.workflow_id = undefined
  // 确保工作流列表已加载
  if (allWorkflows.value.length === 0) {
    loadAllWorkflows()
  }
  schemeDialogVisible.value = true
}

const submitScheme = async () => {
  if (!schemeFormRef.value) return
  await schemeFormRef.value.validate(async (valid) => {
    if (!valid) return
    schemeSubmitLoading.value = true
    try {
      await createWorkflowScheme(projectKey.value, {
        issue_type_id: schemeForm.issue_type_id!,
        workflow_id: schemeForm.workflow_id!,
      })
      ElMessage.success('工作流配置添加成功')
      schemeDialogVisible.value = false
      loadSchemes()
    } catch {
      // ignored
    } finally {
      schemeSubmitLoading.value = false
    }
  })
}

const handleDeleteScheme = async (scheme: WorkflowScheme) => {
  const typeName = scheme.issue_type_name || `类型#${scheme.issue_type_id}`
  try {
    await ElMessageBox.confirm(
      `确定要删除工单类型 "${typeName}" 的工作流配置吗？删除后该类型的新工单将不再自动启动工作流。`,
      '删除确认',
      { type: 'warning' }
    )
    await deleteWorkflowScheme(projectKey.value, scheme.issue_type_id)
    ElMessage.success('删除成功')
    loadSchemes()
  } catch (error) {
    if (error !== 'cancel') {
      ElMessage.error('删除失败')
    }
  }
}

const getIssueTypeColor = (typeId: number) => {
  const t = issueTypes.value.find((it) => it.id === typeId)
  return t?.color || '#3b82f6'
}

const formatSchemeTime = (time: string) => {
  if (!time) return '-'
  const d = new Date(time)
  return `${d.getFullYear()}-${String(d.getMonth() + 1).padStart(2, '0')}-${String(d.getDate()).padStart(2, '0')}`
}

const saveBasicInfo = async () => {
  if (!basicFormRef.value) return
  await basicFormRef.value.validate(async (valid) => {
    if (!valid) return
    saveLoading.value = true
    try {
      await updateProject(projectKey.value, {
        name: basicForm.name,
        description: basicForm.description,
        lead_user_id: basicForm.lead_user_id,
        status: basicForm.status,
      })
      ElMessage.success('保存成功')
    } catch {
      // ignored
    } finally {
      saveLoading.value = false
    }
  })
}

const handleDeleteProject = async () => {
  const expectedKey = String(basicForm.project_key || projectKey.value || '').trim().toUpperCase()
  try {
    const promptResult = await ElMessageBox.prompt(
      `项目删除后不可恢复。请输入项目 Key "${expectedKey}" 以确认删除。`,
      '删除项目',
      {
        type: 'error',
        confirmButtonText: '确认删除',
        cancelButtonText: '取消',
        inputPlaceholder: expectedKey,
        inputValidator: (input) => input.trim().toUpperCase() === expectedKey,
        inputErrorMessage: `请输入正确的项目 Key：${expectedKey}`,
      },
    )
    const inputValue = typeof promptResult === 'string' ? '' : promptResult.value
    if (inputValue.trim().toUpperCase() !== expectedKey) return

    dangerDeleting.value = true
    await deleteProject(projectKey.value)
    ElMessage.success('项目删除成功')
    router.push('/projects')
  } catch (error: any) {
    if (error !== 'cancel' && error !== 'close') {
      ElMessage.error(error?.response?.data?.message || '删除项目失败')
    }
  } finally {
    dangerDeleting.value = false
  }
}

const submitAddMember = async () => {
  if (selectedUserIds.value.length === 0) return
  addMemberLoading.value = true
  try {
    let successCount = 0
    let failCount = 0
    for (const userId of selectedUserIds.value) {
      try {
        await addProjectMember(projectKey.value, {
          user_id: userId,
          role: memberForm.role,
        })
        successCount++
      } catch {
        failCount++
      }
    }
    if (successCount > 0) {
      ElMessage.success(`成功添加 ${successCount} 个成员${failCount > 0 ? `，${failCount} 个失败` : ''}`)
    } else {
      ElMessage.error('添加失败')
    }
    memberDialogVisible.value = false
    loadMembers()
  } finally {
    addMemberLoading.value = false
  }
}

const openMemberDialog = () => {
  memberForm.role = roles.value.length > 0 ? roles.value[roles.value.length - 1].role_key : 'viewers'
  memberSearchKeyword.value = ''
  selectedUserIds.value = []
  memberDialogVisible.value = true
}

const handleEditMemberRole = async (member: ProjectMember, newRole: string) => {
  try {
    await updateProjectMemberRole(projectKey.value, member.user_id, newRole)
    ElMessage.success('角色更新成功')
    loadMembers()
  } catch {
    ElMessage.error('角色更新失败')
  }
}

const handleRemoveMember = async (member: ProjectMember) => {
  try {
    await ElMessageBox.confirm(`确定要移除成员 "${member.user?.display_name}" 吗？`, '移除确认', {
      type: 'warning',
    })
    await removeProjectMember(projectKey.value, member.user_id)
    ElMessage.success('移除成功')
    loadMembers()
  } catch (error) {
    if (error !== 'cancel') {
      // ignored
    }
  }
}

const openRoleDialog = () => {
  isEditingRole.value = false
  editingRoleId.value = null
  Object.assign(roleForm, {
    role_key: '',
    role_name: '',
    description: '',
  })
  roleDialogVisible.value = true
}

const handleEditRole = (role: ProjectRole) => {
  isEditingRole.value = true
  editingRoleId.value = role.id
  Object.assign(roleForm, {
    role_key: role.role_key,
    role_name: role.role_name,
    description: role.description,
  })
  roleDialogVisible.value = true
}

const handleDeleteRole = async (role: ProjectRole) => {
  try {
    await ElMessageBox.confirm(`确定要删除角色 "${role.role_name}" 吗？`, '删除确认', {
      type: 'warning',
    })
    await deleteProjectRole(projectKey.value, role.id)
    ElMessage.success('删除成功')
    loadRoles()
  } catch (error) {
    if (error !== 'cancel') {
      ElMessage.error('删除失败')
    }
  }
}

const handleManageRoleMembers = async (role: ProjectRole) => {
  currentMgmtRole.value = role
  selectedRoleMemberUserId.value = null
  memberMgmtDialogVisible.value = true
  router.replace({ query: { ...route.query, roleId: String(role.id) } })
  await loadRoleMembers(role.id)
}

const loadRoleMembers = async (roleId: number) => {
  roleMembersLoading.value = true
  try {
    const { data } = await getRoleMembers(projectKey.value, roleId)
    roleMembers.value = data.data
  } catch {
    // ignored
  } finally {
    roleMembersLoading.value = false
  }
}

const handleAddRoleMember = async () => {
  if (!selectedRoleMemberUserId.value || !currentMgmtRole.value) return
  addRoleMemberLoading.value = true
  try {
    await addRoleMember(projectKey.value, currentMgmtRole.value.id, { user_id: selectedRoleMemberUserId.value })
    ElMessage.success('添加成功')
    selectedRoleMemberUserId.value = null
    await loadRoleMembers(currentMgmtRole.value.id)
  } catch {
    ElMessage.error('添加失败')
  } finally {
    addRoleMemberLoading.value = false
  }
}

const handleRemoveRoleMember = async (member: ProjectRoleMember) => {
  if (!currentMgmtRole.value) return
  try {
    await ElMessageBox.confirm(`确定要移除成员 "${member.user?.display_name}" 吗？`, '移除确认', {
      type: 'warning',
    })
    await removeRoleMember(projectKey.value, currentMgmtRole.value.id, member.user_id)
    ElMessage.success('移除成功')
    await loadRoleMembers(currentMgmtRole.value.id)
  } catch (error) {
    if (error !== 'cancel') {
      ElMessage.error('移除失败')
    }
  }
}

const closeMemberMgmtDialog = () => {
  memberMgmtDialogVisible.value = false
  currentMgmtRole.value = null
  roleMembers.value = []
  const { roleId: _roleId, ...rest } = route.query
  router.replace({ query: rest })
}

const submitRole = async () => {
  if (!roleFormRef.value) return
  await roleFormRef.value.validate(async (valid) => {
    if (!valid) return
    roleSubmitLoading.value = true
    try {
      if (isEditingRole.value && editingRoleId.value) {
        await updateProjectRole(projectKey.value, editingRoleId.value, {
          role_name: roleForm.role_name,
          description: roleForm.description,
        })
        ElMessage.success('更新成功')
      } else {
        await createProjectRole(projectKey.value, roleForm)
        ElMessage.success('创建成功')
      }
      roleDialogVisible.value = false
      loadRoles()
    } catch {
      ElMessage.error(isEditingRole.value ? '更新失败' : '创建失败')
    } finally {
      roleSubmitLoading.value = false
    }
  })
}

const openIssueTypeDialog = () => {
  isEditingIssueType.value = false
  editingIssueTypeId.value = null
  Object.assign(issueTypeForm, {
    name: '',
    display_name: '',
    description: '',
    icon: 'task',
    color: '#3b82f6',
  })
  issueTypeDialogVisible.value = true
}

const handleEditIssueTypeClick = (issueType: ProjectIssueType) => {
  isEditingIssueType.value = true
  editingIssueTypeId.value = issueType.id
  Object.assign(issueTypeForm, {
    name: issueType.name,
    display_name: issueType.display_name,
    description: issueType.description,
    icon: issueType.icon,
    color: issueType.color,
  })
  issueTypeDialogVisible.value = true
}

const submitIssueType = async () => {
  if (!issueTypeFormRef.value) return
  await issueTypeFormRef.value.validate(async (valid) => {
    if (!valid) return
    issueTypeSubmitLoading.value = true
    try {
      if (isEditingIssueType.value && editingIssueTypeId.value) {
        await updateProjectIssueType(projectKey.value, editingIssueTypeId.value, {
          display_name: issueTypeForm.display_name,
          description: issueTypeForm.description,
          icon: issueTypeForm.icon,
          color: issueTypeForm.color,
        })
        ElMessage.success('更新成功')
      } else {
        await createProjectIssueType(projectKey.value, issueTypeForm)
        ElMessage.success('创建成功')
      }
      issueTypeDialogVisible.value = false
      loadIssueTypes()
    } catch {
      ElMessage.error(isEditingIssueType.value ? '更新失败' : '创建失败')
    } finally {
      issueTypeSubmitLoading.value = false
    }
  })
}

const handleDeleteIssueType = async (issueType: ProjectIssueType) => {
  try {
    await ElMessageBox.confirm(`确定要删除工单类型 "${issueType.display_name}" 吗？`, '删除确认', {
      type: 'warning',
    })
    await deleteProjectIssueType(projectKey.value, issueType.id)
    ElMessage.success('删除成功')
    loadIssueTypes()
  } catch (error) {
    if (error !== 'cancel') {
      // ignored
    }
  }
}

const getRoleType = (role: string) => {
  const map: Record<string, any> = { owner: 'danger', administrators: 'warning', developers: '', testers: 'success', viewers: 'info' }
  return map[role] || 'info'
}

const getRoleIconClass = (roleKey: string) => {
  const classMap: Record<string, string> = {
    administrators: 'admin',
    developers: 'dev',
    testers: 'test',
    viewers: 'view',
  }
  return classMap[roleKey] || 'default'
}

// 工单类型图标映射
const issueTypeIconMap: Record<string, any> = {
  epic: Lightning,
  story: Document,
  task: CircleCheck,
  bug: Warning,
  subtask: Memo,
  improvement: Tools,
  feature: Star,
  change: Connection,
  fault: Flag,
  alert: Promotion,
}

const getIssueTypeIcon = (iconName: string) => {
  return issueTypeIconMap[iconName] || Tickets
}

// 可选的图标列表
const availableIcons = [
  { name: 'epic', label: 'Epic', icon: Lightning },
  { name: 'story', label: 'Story', icon: Document },
  { name: 'task', label: 'Task', icon: CircleCheck },
  { name: 'bug', label: 'Bug', icon: Warning },
  { name: 'subtask', label: 'Subtask', icon: Memo },
  { name: 'improvement', label: 'Improvement', icon: Tools },
  { name: 'feature', label: 'Feature', icon: Star },
  { name: 'change', label: 'Change', icon: Connection },
  { name: 'fault', label: 'Fault', icon: Flag },
  { name: 'alert', label: 'Alert', icon: Promotion },
]

const tryRestoreRoleMemberDialog = () => {
  const roleId = route.query.roleId
  if (roleId && activeTab.value === 'roles') {
    const id = Number(roleId)
    const role = roles.value.find((r) => r.id === id)
    if (role) {
      handleManageRoleMembers(role)
    }
  }
}

// 监听 tab 切换到 roles 时，检查是否需要恢复对话框
watch(activeTab, (newTab) => {
  if (newTab === 'roles') {
    tryRestoreRoleMemberDialog()
  }
})

onMounted(async () => {
  loadProjectDetail()
  loadUsers()
  loadMembers()
  await loadRoles()
  loadIssueTypes()
  loadChannels()
  loadSchemes()
  loadAllWorkflows()
  tryRestoreRoleMemberDialog()
})
</script>

<style scoped lang="scss">
.project-settings-container {
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


.settings-card {
  border: none;
  box-shadow: var(--td-elevation-1);
  transition: var(--td-transition-shadow);
  border-radius: var(--td-radius-lg);

  &:hover { box-shadow: var(--td-elevation-2); }

  :deep(.el-card__body) {
    padding: 0;
  }
}

.settings-tabs {
  :deep(.el-tabs__header) {
    padding: 0 24px;
    margin: 0;
    background: var(--td-bg-page);
  }

  :deep(.el-tabs__nav-wrap) {
    &::after {
      height: 1px;
      background-color: var(--td-border-color);
    }
  }

  :deep(.el-tabs__item) {
    padding: 0 20px;
    height: 48px;
    line-height: 48px;
    font-size: 14px;

    &.is-active {
      color: var(--td-color-primary);
      font-weight: 600;
    }
  }

  :deep(.el-tabs__active-bar) {
    background-color: var(--td-color-primary);
    height: 3px;
  }

  :deep(.el-tabs__content) {
    padding: 0;
  }
}

.tab-label {
  display: flex;
  align-items: center;
  gap: 6px;
}

.danger-tab-label {
  color: var(--td-color-danger);
}

.tab-content {
  padding: 24px;
  min-height: 400px;
  background: var(--td-bg-page);
}

.danger-content {
  display: flex;
  align-items: flex-start;
}

.danger-card {
  width: 100%;
  max-width: 760px;
  background: var(--td-tag-danger-bg);
  border: 1px solid var(--td-tag-danger-border);
  border-radius: 12px;
  padding: 24px;

  .danger-card-title {
    display: flex;
    align-items: center;
    gap: 8px;
    font-size: 18px;
    font-weight: 700;
    color: var(--td-color-danger);
    margin-bottom: 12px;
  }

  .danger-card-desc {
    margin: 0 0 14px 0;
    color: var(--td-color-danger);
    line-height: 1.7;
  }

  .danger-card-list {
    margin: 0 0 20px 0;
    padding-left: 18px;
    color: var(--td-color-danger);
    line-height: 1.8;
  }
}

.basic-form {
  background: var(--td-bg-card);
  padding: 32px;
  border-radius: 12px;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.05);

  :deep(.el-form-item__label) {
    font-weight: 500;
    color: var(--td-text-regular);
  }

  :deep(.el-input__wrapper) {
    border-radius: 8px;
  }

  :deep(.el-textarea__inner) {
    border-radius: 8px;
  }

  .status-field {
    display: flex;
    flex-direction: column;
    gap: 12px;
  }

  .form-tip {
    display: inline-flex;
    align-items: center;
    gap: 6px;
    margin-top: 8px;
    padding: 8px 12px;
    background: var(--td-color-primary-light);
    border-radius: 6px;
    font-size: 12px;
    color: var(--td-color-primary);
    line-height: 1.4;

    .el-icon {
      flex-shrink: 0;
      font-size: 14px;
    }
  }

  .form-actions {
    display: flex;
    justify-content: flex-end;
    gap: 12px;
    margin-top: 24px;
    padding-top: 24px;
    border-top: 1px solid var(--td-border-color);
  }
}

.section-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 20px;

  h3 {
    font-size: 16px;
    font-weight: 600;
    color: var(--td-text-primary);
    margin: 0;
  }
}

// 成员表格
.members-table {
  background: var(--td-bg-card);
  border-radius: 12px;
  overflow: hidden;

  :deep(.el-table__header) {
    th {
      background: var(--td-bg-page) !important;
      font-weight: 600;
      color: var(--td-text-regular);
    }
  }

  :deep(.el-table__row) {
    td {
      padding: 16px 0;
    }
  }
}

.member-cell {
  display: flex;
  align-items: center;
  gap: 12px;
}

.member-avatar {
  width: 40px;
  height: 40px;
  border-radius: 10px;
  background: var(--td-color-primary);
  color: var(--td-text-white);
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 16px;
  font-weight: 600;
  flex-shrink: 0;

  &.owner {
    background: var(--td-color-warning);
  }
  &.admin {
    background: var(--td-color-primary);
  }
  &.member {
    background: var(--td-color-success);
  }
}

.member-info {
  .member-name {
    font-size: 14px;
    font-weight: 600;
    color: var(--td-text-primary);
    margin-bottom: 2px;
  }

  .member-username {
    font-size: 12px;
    color: var(--td-text-placeholder);
  }
}

// 用户单元格
.user-cell {
  display: flex;
  align-items: center;
  gap: 12px;
}

.user-avatar {
  width: 36px;
  height: 36px;
  border-radius: 8px;
  background: var(--td-color-primary);
  color: var(--td-text-white);
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 14px;
  font-weight: 600;
}

.user-info {
  .user-name {
    font-size: 14px;
    font-weight: 500;
    color: var(--td-text-primary);
  }

  .user-username {
    font-size: 12px;
    color: var(--td-text-placeholder);
  }
}

// 角色网格
.roles-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(280px, 1fr));
  gap: 16px;
}

.role-card {
  display: flex;
  flex-direction: column;
  padding: 20px;
  background: var(--td-bg-card);
  border: 1px solid var(--td-border-color);
  border-radius: 12px;
  transition: all 150ms ease-out;
  position: relative;

  &:hover {
    box-shadow: 0 4px 16px rgba(0, 0, 0, 0.08);
    border-color: var(--td-text-disabled);
  }

  // 左侧色条
  &::before {
    content: '';
    position: absolute;
    left: 0;
    top: 16px;
    bottom: 16px;
    width: 3px;
    border-radius: 0 3px 3px 0;
    background: var(--td-color-primary);
  }

  &.admin::before {
    background: var(--td-color-warning);
  }
  &.dev::before {
    background: var(--td-color-primary);
  }
  &.test::before {
    background: var(--td-color-success);
  }
  &.view::before {
    background: var(--td-text-secondary);
  }
}

.role-card-top {
  display: flex;
  align-items: center;
  gap: 14px;
  margin-bottom: 12px;
}

.role-icon {
  width: 44px;
  height: 44px;
  border-radius: 10px;
  display: flex;
  align-items: center;
  justify-content: center;
  color: var(--td-text-white);
  background: var(--td-color-primary);
  flex-shrink: 0;

  &.admin {
    background: var(--td-color-warning);
  }
  &.dev {
    background: var(--td-color-primary);
  }
  &.test {
    background: var(--td-color-success);
  }
  &.view {
    background: var(--td-text-secondary);
  }
}

.role-card-title {
  flex: 1;
  min-width: 0;

  .role-name-row {
    display: flex;
    align-items: center;
    gap: 8px;
  }

  .role-name {
    font-size: 15px;
    font-weight: 600;
    color: var(--td-text-primary);
  }

  .role-key {
    font-size: 12px;
    color: var(--td-text-placeholder);
    font-family: 'SF Mono', 'Fira Code', monospace;
    margin-top: 2px;
  }
}

.role-desc {
  font-size: 13px;
  color: var(--td-text-secondary);
  line-height: 1.6;
  margin-bottom: 14px;
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
  overflow: hidden;
}

.role-stats {
  display: flex;
  gap: 16px;
  margin-bottom: 16px;
  padding: 10px 14px;
  background: var(--td-bg-page);
  border-radius: 8px;

  .role-stat-item {
    display: flex;
    align-items: center;
    gap: 5px;
    font-size: 12px;
    color: var(--td-text-secondary);

    .el-icon {
      color: var(--td-text-placeholder);
    }
  }
}

.role-actions {
  display: flex;
  gap: 8px;
  flex-wrap: wrap;
  border-top: 1px solid var(--td-border-color);
  padding-top: 14px;

  .role-action-btn {
    flex: 1;
    min-width: 0;
  }
}

// 工单类型网格
.issue-types-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(300px, 1fr));
  gap: 16px;
}

.issue-type-card {
  display: flex;
  align-items: center;
  gap: 16px;
  padding: 20px;
  background: var(--td-bg-card);
  border: 1px solid var(--td-border-color);
  border-radius: 12px;
  transition: all 150ms ease-out;

  &:hover {
    box-shadow: 0 4px 12px rgba(0, 0, 0, 0.08);
    border-color: var(--td-text-disabled);
  }
}

// 工单类型图标
.type-icon {
  width: 48px;
  height: 48px;
  border-radius: 12px;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 24px;
  color: var(--td-text-white);
  flex-shrink: 0;
}

.type-info {
  flex: 1;
  min-width: 0;

  .type-name {
    font-size: 15px;
    font-weight: 600;
    color: var(--td-text-primary);
    margin-bottom: 4px;
  }

  .type-key {
    font-size: 12px;
    color: var(--td-text-placeholder);
    font-family: monospace;
    margin-bottom: 4px;
  }

  .type-desc {
    font-size: 13px;
    color: var(--td-text-secondary);
    line-height: 1.5;
  }
}

.type-actions {
  display: flex;
  gap: 8px;
  flex-shrink: 0;
}

// 角色成员管理对话框
.member-mgmt-content {
  .member-mgmt-add {
    display: flex;
    gap: 12px;
    align-items: center;
  }

  .member-mgmt-list {
    min-height: 200px;
    max-height: 400px;
    overflow-y: auto;
  }

  .member-mgmt-item {
    display: flex;
    justify-content: space-between;
    align-items: center;
    padding: 12px;
    border-radius: 8px;
    margin-bottom: 8px;
    background: var(--td-bg-page);

    &:hover {
      background: var(--td-bg-section);
    }
  }

  .member-mgmt-info {
    display: flex;
    align-items: center;
    gap: 12px;
  }

  .member-mgmt-avatar {
    width: 36px;
    height: 36px;
    border-radius: 8px;
    background: var(--td-color-primary);
    color: var(--td-text-white);
    display: flex;
    align-items: center;
    justify-content: center;
    font-size: 14px;
    font-weight: 600;
  }

  .member-mgmt-details {
    display: flex;
    flex-direction: column;
  }

  .member-mgmt-name {
    font-size: 14px;
    font-weight: 500;
    color: var(--td-text-primary);
  }

  .member-mgmt-username {
    font-size: 12px;
    color: var(--td-text-placeholder);
  }
}

// 对话框样式
.custom-dialog {
  :deep(.el-dialog__header) {
    padding: 20px 24px;
    border-bottom: 1px solid var(--td-border-color);
  }

  :deep(.el-dialog__body) {
    padding: 24px;
  }
}

// 批量添加成员对话框
.batch-member-content {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.batch-member-toolbar {
  display: flex;
  gap: 10px;

  .member-search-input {
    flex: 1;
  }
}

.batch-member-selected-hint {
  font-size: 13px;
  color: var(--td-text-secondary);
  display: flex;
  align-items: center;
  gap: 8px;

  strong {
    color: var(--el-color-primary);
  }
}

.batch-member-list {
  border: 1px solid var(--td-border-color);
  border-radius: 8px;
  overflow: hidden;

  :deep(.el-table) {
    --el-table-border-color: #f3f4f6;

    th.el-table__cell {
      background: var(--td-bg-page);
      font-weight: 500;
      font-size: 13px;
    }
  }
}

.user-option {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 4px 0;
}

.user-option-avatar {
  width: 36px;
  height: 36px;
  min-width: 36px;
  border-radius: 8px;
  background: var(--td-color-primary);
  color: var(--td-text-white);
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 14px;
  font-weight: 600;
  flex-shrink: 0;
}

.user-option-info {
  display: flex;
  flex-direction: column;
  flex: 1;
  min-width: 0;

  .user-option-name {
    font-size: 14px;
    font-weight: 500;
    color: var(--td-text-primary);
    line-height: 1.4;
  }

  .user-option-username {
    font-size: 12px;
    color: var(--td-text-placeholder);
    line-height: 1.4;
  }
}

.role-option {
  display: flex;
  align-items: center;
  gap: 8px;
}

// 区块标题
.section-title {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 16px;
  font-weight: 600;
  color: var(--td-text-primary);
}

// 图标选择器
.icon-selector {
  display: grid;
  grid-template-columns: repeat(5, 1fr);
  gap: 8px;
}

.icon-option {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 6px;
  padding: 12px 8px;
  border: 2px solid var(--td-border-color);
  border-radius: 8px;
  cursor: pointer;
  transition: all 150ms ease-out;

  &:hover {
    border-color: var(--td-color-primary);
    background: var(--td-color-primary-light);
  }

  &.active {
    border-color: var(--td-color-primary);
    background: var(--td-color-primary-light);
    box-shadow: 0 0 0 3px rgba(59, 130, 246, 0.1);
  }

  .icon-label {
    font-size: 11px;
    color: var(--td-text-secondary);
    text-align: center;
  }
}

// 颜色选择器
.color-selector {
  display: flex;
  align-items: center;
  gap: 16px;
}

.color-presets {
  display: flex;
  gap: 8px;
}

.color-preset {
  width: 32px;
  height: 32px;
  border-radius: 6px;
  cursor: pointer;
  transition: all 150ms ease-out;
  border: 2px solid transparent;

  &:hover {
    transform: scale(1.1);
  }

  &.active {
    border-color: var(--td-text-primary);
    box-shadow: 0 0 0 3px rgba(0, 0, 0, 0.1);
  }
}

.form-tip-small {
  display: flex;
  align-items: flex-start;
  gap: 6px;
  font-size: 12px;
  color: var(--td-text-secondary);
  margin-top: 6px;
  line-height: 1.5;

  .el-icon {
    flex-shrink: 0;
    margin-top: 2px;
    color: var(--td-color-primary);
  }
}

// 定时日报配置卡片
.digest-card {
  border: 1px solid var(--td-border-color);
  border-radius: 12px;
  margin-bottom: 24px;
  background: var(--td-bg-page);

  .digest-card-header {
    padding: 16px 20px;
    border-bottom: 1px solid var(--td-border-color);

    .digest-card-title {
      display: flex;
      align-items: center;
      gap: 8px;
      font-size: 14px;
      font-weight: 600;
      color: var(--td-text-primary);
      margin-bottom: 4px;

      .builtin-tag {
        margin-left: 4px;
      }
    }

    .digest-card-desc {
      font-size: 12px;
      color: var(--td-text-placeholder);
      line-height: 1.6;
    }
  }

  .digest-card-body {
    padding: 16px 20px;
  }

  .digest-form {
    max-width: 600px;
  }

  .digest-form-tip {
    font-size: 12px;
    color: var(--td-text-placeholder);
    margin-left: 12px;
  }
}

// 通知渠道
.notification-builtin-card {
  display: flex;
  align-items: center;
  gap: 16px;
  padding: 16px 20px;
  background: var(--td-tag-success-bg);
  border: 1px solid #bbf7d0;
  border-radius: 12px;
  margin-bottom: 24px;

  .builtin-icon {
    width: 40px;
    height: 40px;
    border-radius: 10px;
    background: var(--td-color-success);
    color: var(--td-text-white);
    display: flex;
    align-items: center;
    justify-content: center;
    flex-shrink: 0;
  }

  .builtin-content {
    flex: 1;

    .builtin-title {
      font-size: 14px;
      font-weight: 600;
      color: #065f46;
      margin-bottom: 2px;
    }

    .builtin-desc {
      font-size: 12px;
      color: #047857;
      line-height: 1.5;
    }
  }

  .builtin-tag {
    flex-shrink: 0;
  }
}

.notification-section-label {
  display: flex;
  align-items: baseline;
  gap: 12px;
  margin-bottom: 16px;
  padding-bottom: 12px;
  border-bottom: 1px solid var(--td-border-color);

  .section-label-text {
    font-size: 14px;
    font-weight: 600;
    color: var(--td-text-regular);
  }

  .section-label-desc {
    font-size: 12px;
    color: var(--td-text-placeholder);
  }
}

// 通知渠道列表
.channels-list {
  display: flex;
  flex-direction: column;
  background: var(--td-bg-card);
  border: 1px solid var(--td-border-color);
  border-radius: 12px;
  overflow: hidden;
}

.channel-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 16px 20px;
  transition: background 0.15s;

  &:not(:last-child) {
    border-bottom: 1px solid var(--td-border-color);
  }

  &:hover {
    background: var(--td-bg-page);
  }

  &.disabled {
    .channel-row-left {
      opacity: 0.5;
    }
  }
}

.channel-row-left {
  display: flex;
  align-items: center;
  gap: 14px;
  flex: 1;
  min-width: 0;
}

.channel-icon {
  width: 40px;
  height: 40px;
  border-radius: 10px;
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
  color: var(--td-text-white);
  font-weight: 700;

  &.lark {
    background: #3370ff;
  }

  &.telegram {
    background: #2aabee;
  }
}

.channel-icon-text {
  font-size: 12px;
  font-weight: 700;
}

.channel-row-info {
  flex: 1;
  min-width: 0;
}

.channel-row-name {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 4px;

  > span {
    font-size: 14px;
    font-weight: 600;
    color: var(--td-text-primary);
  }

  .channel-type-tag {
    font-size: 11px;
  }
}

.channel-row-detail {
  .channel-config-text {
    font-size: 12px;
    color: var(--td-text-placeholder);
    font-family: 'SF Mono', 'Monaco', 'Menlo', monospace;
  }
}

.channel-row-events {
  display: flex;
  flex-wrap: wrap;
  gap: 4px;
  margin-top: 6px;

  .channel-event-tag {
    font-size: 11px;
  }
}

.event-types-wrapper {
  width: 100%;
}

.event-types-group {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 6px 12px;
  width: 100%;
}

.event-type-checkbox {
  width: 100%;
  margin-right: 0;
  padding: 8px 10px;
  border: 1px solid var(--td-border-color);
  border-radius: var(--td-radius-md);
  transition: var(--td-transition-color);

  &:hover {
    border-color: var(--td-color-primary);
    background: var(--td-tag-primary-bg);
  }

  :deep(.el-checkbox__label) {
    display: flex;
    flex-direction: column;
    align-items: flex-start;
    line-height: 1.4;
  }

  .event-type-label {
    font-size: 13px;
    color: var(--td-text-primary);
    font-weight: 500;
  }

  .event-type-desc {
    font-size: 11px;
    color: var(--td-text-placeholder);
    margin-top: 2px;
  }
}

.channel-row-right {
  display: flex;
  align-items: center;
  gap: 8px;
  flex-shrink: 0;
  margin-left: 16px;
}

.channel-status-indicator {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  font-size: 12px;
  padding: 2px 10px 2px 8px;
  border-radius: 12px;
  margin-right: 8px;

  .status-dot {
    width: 6px;
    height: 6px;
    border-radius: 50%;
    flex-shrink: 0;
  }

  &.active {
    color: var(--td-color-success);
    background: var(--td-tag-success-bg);

    .status-dot {
      background: var(--td-color-success);
      box-shadow: 0 0 6px rgba(16, 185, 129, 0.4);
    }
  }

  &.inactive {
    color: var(--td-text-secondary);
    background: var(--td-bg-section);

    .status-dot {
      background: var(--td-text-placeholder);
    }
  }
}

// 空状态
.channels-empty {
  display: flex;
  flex-direction: column;
  align-items: center;
  padding: 60px 20px;

  .empty-illustration {
    position: relative;
    margin-bottom: 24px;
  }

  .empty-icon-wrapper {
    width: 96px;
    height: 96px;
    border-radius: 12px;
    background: var(--td-tag-primary-bg);
    display: flex;
    align-items: center;
    justify-content: center;
    color: var(--td-color-primary);
  }

  .empty-decorations {
    position: absolute;
    inset: -12px;
    pointer-events: none;

    .deco {
      position: absolute;
      border-radius: 50%;
    }

    .deco-1 {
      width: 8px;
      height: 8px;
      background: #a5b4fc;
      top: -4px;
      right: 8px;
    }

    .deco-2 {
      width: 6px;
      height: 6px;
      background: #c4b5fd;
      bottom: 4px;
      left: -4px;
    }

    .deco-3 {
      width: 10px;
      height: 10px;
      background: #818cf8;
      opacity: 0.5;
      bottom: -6px;
      right: -2px;
    }
  }

  .empty-title {
    font-size: 16px;
    font-weight: 600;
    color: var(--td-text-regular);
    margin-bottom: 8px;
  }

  .empty-desc {
    font-size: 13px;
    color: var(--td-text-placeholder);
    margin-bottom: 24px;
    text-align: center;
    line-height: 1.6;
  }
}

// 通知渠道对话框
.channel-type-form-item {
  :deep(.el-form-item__label) {
    font-weight: 600 !important;
    font-size: 14px !important;
  }
}

.channel-type-cards {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 12px;
  width: 100%;

  &.disabled {
    pointer-events: none;
  }
}

.channel-type-card {
  position: relative;
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 16px;
  border: 2px solid var(--td-border-color);
  border-radius: 12px;
  cursor: pointer;
  transition: all 150ms ease-out;
  background: var(--td-bg-card);

  &:hover:not(.disabled) {
    border-color: var(--td-color-primary-light-hover);
    background: var(--td-bg-card-hover);
  }

  &.active {
    border-color: var(--td-color-primary);
    background: var(--td-color-primary-light);
    box-shadow: 0 0 0 3px rgba(59, 130, 246, 0.12);
  }

  &.disabled {
    opacity: 0.7;
    cursor: not-allowed;
  }

  .type-card-icon {
    width: 40px;
    height: 40px;
    border-radius: 10px;
    display: flex;
    align-items: center;
    justify-content: center;
    color: var(--td-text-white);
    font-size: 13px;
    font-weight: 700;
    flex-shrink: 0;

    &.lark {
      background: #3370ff;
    }

    &.telegram {
      background: #2aabee;
    }
  }

  .type-card-content {
    flex: 1;
    min-width: 0;

    .type-card-name {
      font-size: 14px;
      font-weight: 600;
      color: var(--td-text-primary);
      margin-bottom: 2px;
    }

    .type-card-desc {
      font-size: 11px;
      color: var(--td-text-placeholder);
      line-height: 1.4;
    }
  }

  .type-card-check {
    position: absolute;
    top: 8px;
    right: 8px;
    color: var(--td-color-primary);
    font-size: 18px;
  }
}

.dialog-section {
  margin-top: 20px;
  padding-top: 20px;
  border-top: 1px solid var(--td-border-color);
}

.dialog-section-title {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 14px;
  font-weight: 600;
  color: var(--td-text-regular);
  margin-bottom: 16px;

  .section-title-icon {
    display: inline-flex;
    align-items: center;
    justify-content: center;
    width: 24px;
    height: 24px;
    border-radius: 6px;
    color: var(--td-text-white);
    font-size: 10px;
    font-weight: 700;

    &.lark {
      background: #3370ff;
    }

    &.telegram {
      background: #2aabee;
    }
  }
}

.enable-switch-row {
  display: flex;
  align-items: center;
  gap: 12px;

  .enable-switch-label {
    font-size: 13px;
    color: var(--td-text-placeholder);
    transition: color 150ms ease-out;

    &.active {
      color: var(--td-color-success);
    }
  }
}

.dialog-footer {
  display: flex;
  justify-content: flex-end;
  gap: 12px;
}

// 工作流方案
.workflow-scheme-tip {
  display: flex;
  align-items: flex-start;
  gap: 8px;
  padding: 12px 16px;
  background: var(--td-color-primary-light);
  border: 1px solid var(--td-color-primary-light-hover);
  border-radius: 8px;
  margin-bottom: 20px;
  font-size: 13px;
  color: var(--td-color-primary);
  line-height: 1.5;

  .el-icon {
    flex-shrink: 0;
    margin-top: 2px;
    font-size: 16px;
  }
}

.schemes-list {
  display: flex;
  flex-direction: column;
  background: var(--td-bg-card);
  border: 1px solid var(--td-border-color);
  border-radius: 12px;
  overflow: hidden;
}

.scheme-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 16px 20px;
  transition: background 0.15s;

  &:not(:last-child) {
    border-bottom: 1px solid var(--td-border-color);
  }

  &:hover {
    background: var(--td-bg-page);
  }
}

.scheme-row-left {
  display: flex;
  align-items: center;
  gap: 14px;
  flex: 1;
  min-width: 0;
}

.scheme-type-icon {
  width: 40px;
  height: 40px;
  border-radius: 10px;
  display: flex;
  align-items: center;
  justify-content: center;
  color: var(--td-text-white);
  font-size: 18px;
  flex-shrink: 0;
}

.scheme-row-info {
  flex: 1;
  min-width: 0;
}

.scheme-row-name {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 4px;

  .scheme-type-name {
    font-size: 14px;
    font-weight: 600;
    color: var(--td-text-primary);
  }

  .scheme-arrow {
    color: var(--td-text-placeholder);
    font-size: 14px;
    flex-shrink: 0;
  }

  .scheme-workflow-name {
    font-size: 14px;
    font-weight: 500;
    color: var(--td-color-primary);
  }
}

.scheme-row-detail {
  font-size: 12px;
  color: var(--td-text-placeholder);
}

.scheme-row-right {
  display: flex;
  align-items: center;
  gap: 8px;
  flex-shrink: 0;
  margin-left: 16px;
}

.scheme-option {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 2px 0;
}

.scheme-option-icon {
  width: 28px;
  height: 28px;
  border-radius: 6px;
  display: flex;
  align-items: center;
  justify-content: center;
  color: var(--td-text-white);
  flex-shrink: 0;

  &.workflow {
    background: #8b5cf6;
  }
}

.scheme-option-info {
  display: flex;
  flex-direction: column;

  .scheme-option-name {
    font-size: 14px;
    font-weight: 500;
    color: var(--td-text-primary);
  }

  .scheme-option-desc {
    font-size: 12px;
    color: var(--td-text-placeholder);
    line-height: 1.3;
  }
}

// 权限配置对话框
.perm-config-content {
  min-height: 200px;
}

.perm-group {
  margin-bottom: 20px;

  &:last-child {
    margin-bottom: 0;
  }
}

.perm-group-header {
  padding: 8px 12px;
  background: var(--td-bg-section);
  border-radius: 6px;
  margin-bottom: 8px;

  :deep(.el-checkbox__label) {
    font-weight: 600;
    color: var(--td-text-primary);
    font-size: 14px;
  }
}

.perm-group-items {
  display: flex;
  flex-wrap: wrap;
  gap: 8px 24px;
  padding: 4px 12px 4px 28px;

  :deep(.el-checkbox__label) {
    font-size: 13px;
    color: var(--td-text-regular);
  }
}
</style>

<style lang="scss">
// 用户选择下拉框样式（不能用 scoped，因为是 popper）
.user-select-popper {
  .el-select-dropdown__item {
    height: auto !important;
    padding: 8px 12px !important;
    line-height: normal !important;
  }
}

// 通知渠道对话框样式（dialog teleport 到 body，需要非 scoped）
.channel-dialog {
  .el-dialog__header {
    padding: 20px 24px 16px;
    border-bottom: 1px solid var(--td-border-color);

    .el-dialog__title {
      font-size: 17px;
      font-weight: 600;
    }
  }

  .el-dialog__body {
    padding: 24px;
    max-height: 65vh;
    overflow-y: auto;
  }

  .el-dialog__footer {
    padding: 16px 24px;
    border-top: 1px solid var(--td-border-color);
  }

  .el-form-item__label {
    font-weight: 500;
    color: var(--td-text-regular);
    padding-bottom: 6px;
  }

  .el-input__wrapper {
    border-radius: 8px;
  }
}
</style>
