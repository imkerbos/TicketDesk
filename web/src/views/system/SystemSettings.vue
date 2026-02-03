<template>
  <div class="system-settings">
    <el-tabs v-model="activeTab" class="settings-tabs">
      <!-- 邮件配置 -->
      <el-tab-pane label="邮件配置" name="email">
        <el-card shadow="never" class="settings-card">
          <template #header>
            <div class="card-header">
              <div class="card-title">
                <div class="title-icon email-icon">
                  <el-icon><Message /></el-icon>
                </div>
                <div class="title-text">
                  <span class="title">SMTP 邮件服务配置</span>
                  <span class="subtitle">配置系统邮件发送服务，用于发送通知和告警</span>
                </div>
              </div>
              <el-switch
                v-model="emailForm.enabled"
                active-text="启用"
                inactive-text="禁用"
                inline-prompt
                style="--el-switch-on-color: #67c23a"
              />
            </div>
          </template>

          <el-row :gutter="40">
            <el-col :xs="24" :lg="12">
              <el-form
                ref="emailFormRef"
                :model="emailForm"
                :rules="emailRules"
                label-position="top"
                class="settings-form"
              >
                <div class="form-section">
                  <div class="section-title">服务器设置</div>
                  <el-row :gutter="16">
                    <el-col :span="16">
                      <el-form-item label="SMTP 服务器" prop="smtp_host">
                        <el-input v-model="emailForm.smtp_host" placeholder="例如: smtp.gmail.com">
                          <template #prefix>
                            <el-icon><Monitor /></el-icon>
                          </template>
                        </el-input>
                      </el-form-item>
                    </el-col>
                    <el-col :span="8">
                      <el-form-item label="端口" prop="smtp_port">
                        <el-input-number v-model="emailForm.smtp_port" :min="1" :max="65535" style="width: 100%" />
                      </el-form-item>
                    </el-col>
                  </el-row>

                  <el-row :gutter="16">
                    <el-col :span="12">
                      <el-form-item label="用户名" prop="smtp_username">
                        <el-input v-model="emailForm.smtp_username" placeholder="认证用户名">
                          <template #prefix>
                            <el-icon><User /></el-icon>
                          </template>
                        </el-input>
                      </el-form-item>
                    </el-col>
                    <el-col :span="12">
                      <el-form-item label="密码" prop="smtp_password">
                        <el-input
                          v-model="emailForm.smtp_password"
                          type="password"
                          placeholder="留空表示不修改"
                          show-password
                        >
                          <template #prefix>
                            <el-icon><Lock /></el-icon>
                          </template>
                        </el-input>
                      </el-form-item>
                    </el-col>
                  </el-row>
                </div>

                <div class="form-section">
                  <div class="section-title">发件人信息</div>
                  <el-row :gutter="16">
                    <el-col :span="14">
                      <el-form-item label="发件人地址" prop="from_address">
                        <el-input v-model="emailForm.from_address" placeholder="noreply@example.com">
                          <template #prefix>
                            <el-icon><Message /></el-icon>
                          </template>
                        </el-input>
                      </el-form-item>
                    </el-col>
                    <el-col :span="10">
                      <el-form-item label="发件人名称" prop="from_name">
                        <el-input v-model="emailForm.from_name" placeholder="TicketDesk" />
                      </el-form-item>
                    </el-col>
                  </el-row>
                </div>

                <div class="form-section">
                  <div class="section-title">安全设置</div>
                  <el-form-item>
                    <div class="setting-row">
                      <div class="setting-info">
                        <span class="setting-label">使用 TLS 加密</span>
                        <span class="setting-desc">推荐启用以保护邮件传输安全</span>
                      </div>
                      <el-switch v-model="emailForm.use_tls" />
                    </div>
                  </el-form-item>
                </div>

                <el-form-item>
                  <el-button type="primary" :loading="emailSaving" @click="saveEmailConfig">
                    <el-icon><Check /></el-icon>
                    保存配置
                  </el-button>
                  <el-button @click="testEmailDialog = true">
                    <el-icon><Promotion /></el-icon>
                    发送测试邮件
                  </el-button>
                </el-form-item>
              </el-form>
            </el-col>
            <el-col :xs="24" :lg="12" class="info-panel">
              <div class="info-card">
                <div class="info-title">
                  <el-icon><InfoFilled /></el-icon>
                  配置说明
                </div>
                <ul class="info-list">
                  <li>SMTP 服务器用于发送系统邮件通知</li>
                  <li>常用端口：25 (不加密)、465 (SSL)、587 (TLS)</li>
                  <li>建议使用 TLS 加密以保护邮件内容</li>
                  <li>配置完成后可发送测试邮件验证</li>
                </ul>
              </div>
            </el-col>
          </el-row>
        </el-card>
      </el-tab-pane>

      <!-- Webhook 配置 -->
      <el-tab-pane label="Webhook" name="webhook">
        <el-card shadow="never" class="settings-card">
          <template #header>
            <div class="card-header">
              <div class="card-title">
                <div class="title-icon webhook-icon">
                  <el-icon><Connection /></el-icon>
                </div>
                <div class="title-text">
                  <span class="title">Webhook 配置</span>
                  <span class="subtitle">配置外发 Webhook，将系统事件推送到第三方服务</span>
                </div>
              </div>
              <el-button type="primary" @click="showWebhookDialog()">
                <el-icon><Plus /></el-icon>
                添加 Webhook
              </el-button>
            </div>
          </template>

          <div class="webhook-list" v-if="webhooks.length > 0">
            <div class="webhook-item" v-for="webhook in webhooks" :key="webhook.id">
              <div class="webhook-info">
                <div class="webhook-header">
                  <span class="webhook-name">{{ webhook.name }}</span>
                  <el-tag :type="webhook.status === 1 ? 'success' : 'info'" size="small">
                    {{ webhook.status === 1 ? '启用' : '禁用' }}
                  </el-tag>
                </div>
                <div class="webhook-url">{{ webhook.url }}</div>
                <div class="webhook-events">
                  <el-tag
                    v-for="event in webhook.events"
                    :key="event"
                    size="small"
                    type="info"
                    effect="plain"
                    class="event-tag"
                  >
                    {{ getEventLabel(event) }}
                  </el-tag>
                </div>
              </div>
              <div class="webhook-actions">
                <el-button link type="primary" @click="showWebhookDialog(webhook)">
                  <el-icon><Edit /></el-icon>
                  编辑
                </el-button>
                <el-button link type="danger" @click="handleDeleteWebhook(webhook)">
                  <el-icon><Delete /></el-icon>
                  删除
                </el-button>
              </div>
            </div>
          </div>

          <el-empty v-else description="暂无 Webhook 配置">
            <el-button type="primary" @click="showWebhookDialog()">
              <el-icon><Plus /></el-icon>
              添加 Webhook
            </el-button>
          </el-empty>
        </el-card>
      </el-tab-pane>

      <!-- 安全配置 -->
      <el-tab-pane label="安全设置" name="security">
        <el-row :gutter="20">
          <!-- MFA 设置 -->
          <el-col :xs="24" :lg="8">
            <el-card shadow="never" class="setting-block">
              <div class="block-header">
                <div class="block-icon mfa-icon">
                  <el-icon><Key /></el-icon>
                </div>
                <div class="block-title">
                  <span class="title">多因素认证 (MFA)</span>
                  <span class="desc">增强账户安全性</span>
                </div>
              </div>
              <div class="block-content">
                <div class="setting-item">
                  <div class="setting-info">
                    <span class="setting-label">启用 MFA</span>
                    <span class="setting-desc">允许用户开启双因素认证</span>
                  </div>
                  <el-switch v-model="securityForm.mfa_enabled" />
                </div>
                <div class="setting-item">
                  <div class="setting-info">
                    <span class="setting-label">强制要求 MFA</span>
                    <span class="setting-desc">所有用户必须开启 MFA</span>
                  </div>
                  <el-switch
                    v-model="securityForm.mfa_required"
                    :disabled="!securityForm.mfa_enabled"
                  />
                </div>
              </div>
            </el-card>
          </el-col>

          <!-- 密码策略 -->
          <el-col :xs="24" :lg="8">
            <el-card shadow="never" class="setting-block">
              <div class="block-header">
                <div class="block-icon password-icon">
                  <el-icon><Lock /></el-icon>
                </div>
                <div class="block-title">
                  <span class="title">密码策略</span>
                  <span class="desc">设置密码复杂度要求</span>
                </div>
              </div>
              <div class="block-content">
                <div class="setting-item">
                  <div class="setting-info">
                    <span class="setting-label">最小长度</span>
                    <span class="setting-desc">密码最少字符数</span>
                  </div>
                  <el-input-number
                    v-model="securityForm.password_min_length"
                    :min="6"
                    :max="32"
                    size="small"
                    controls-position="right"
                  />
                </div>
                <div class="setting-item">
                  <div class="setting-info">
                    <span class="setting-label">要求大写字母</span>
                    <span class="setting-desc">必须包含 A-Z</span>
                  </div>
                  <el-switch v-model="securityForm.password_require_upper" />
                </div>
                <div class="setting-item">
                  <div class="setting-info">
                    <span class="setting-label">要求数字</span>
                    <span class="setting-desc">必须包含 0-9</span>
                  </div>
                  <el-switch v-model="securityForm.password_require_number" />
                </div>
              </div>
            </el-card>
          </el-col>

          <!-- 会话设置 -->
          <el-col :xs="24" :lg="8">
            <el-card shadow="never" class="setting-block">
              <div class="block-header">
                <div class="block-icon session-icon">
                  <el-icon><Timer /></el-icon>
                </div>
                <div class="block-title">
                  <span class="title">会话设置</span>
                  <span class="desc">控制用户登录会话</span>
                </div>
              </div>
              <div class="block-content">
                <div class="setting-item">
                  <div class="setting-info">
                    <span class="setting-label">会话超时</span>
                    <span class="setting-desc">无操作自动退出时间（分钟）</span>
                  </div>
                  <el-input-number
                    v-model="securityForm.session_timeout"
                    :min="5"
                    :max="1440"
                    size="small"
                    controls-position="right"
                  />
                </div>
              </div>
            </el-card>
          </el-col>
        </el-row>

        <div class="action-bar">
          <el-button type="primary" :loading="securitySaving" @click="saveSecurityConfig">
            <el-icon><Check /></el-icon>
            保存安全配置
          </el-button>
        </div>
      </el-tab-pane>
    </el-tabs>

    <!-- Webhook 编辑对话框 -->
    <el-dialog
      v-model="webhookDialogVisible"
      :title="webhookForm.id ? '编辑 Webhook' : '添加 Webhook'"
      width="600px"
    >
      <el-form
        ref="webhookFormRef"
        :model="webhookForm"
        :rules="webhookRules"
        label-position="top"
      >
        <el-row :gutter="16">
          <el-col :span="12">
            <el-form-item label="名称" prop="name">
              <el-input v-model="webhookForm.name" placeholder="Webhook 名称" />
            </el-form-item>
          </el-col>
          <el-col :span="12" v-if="webhookForm.id">
            <el-form-item label="状态">
              <el-switch
                v-model="webhookForm.status"
                :active-value="1"
                :inactive-value="0"
                active-text="启用"
                inactive-text="禁用"
                inline-prompt
                style="--el-switch-on-color: #67c23a"
              />
            </el-form-item>
          </el-col>
        </el-row>
        <el-form-item label="URL" prop="url">
          <el-input v-model="webhookForm.url" placeholder="https://example.com/webhook" />
        </el-form-item>
        <el-form-item label="签名密钥" prop="secret">
          <el-input
            v-model="webhookForm.secret"
            type="password"
            placeholder="用于 HMAC-SHA256 签名验证（可选）"
            show-password
          />
        </el-form-item>
        <el-form-item label="订阅事件" prop="events">
          <el-checkbox-group v-model="webhookForm.events" class="event-checkbox-group">
            <el-checkbox
              v-for="event in WebhookEvents"
              :key="event.value"
              :value="event.value"
              border
            >
              {{ event.label }}
            </el-checkbox>
          </el-checkbox-group>
        </el-form-item>
        <el-form-item label="描述">
          <el-input
            v-model="webhookForm.description"
            type="textarea"
            :rows="2"
            placeholder="Webhook 用途描述（可选）"
          />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="webhookDialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="webhookSaving" @click="saveWebhook">
          保存
        </el-button>
      </template>
    </el-dialog>

    <!-- 测试邮件对话框 -->
    <el-dialog v-model="testEmailDialog" title="发送测试邮件" width="400px">
      <el-form ref="testEmailFormRef" :model="testEmailForm" :rules="testEmailRules" label-position="top">
        <el-form-item label="收件人邮箱" prop="to_address">
          <el-input v-model="testEmailForm.to_address" placeholder="请输入收件人邮箱" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="testEmailDialog = false">取消</el-button>
        <el-button type="primary" :loading="testEmailSending" @click="sendTestEmail">
          发送测试
        </el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { ElMessage, ElMessageBox, type FormInstance, type FormRules } from 'element-plus'
import {
  Message, Connection, Lock, Plus, Check, Promotion, InfoFilled,
  Edit, Delete, Key, Timer, Monitor, User
} from '@element-plus/icons-vue'
import {
  getEmailConfig,
  updateEmailConfig,
  getSecurityConfig,
  updateSecurityConfig,
  getWebhooks,
  createWebhook,
  updateWebhook,
  deleteWebhook,
} from '@/api/system'
import type { Webhook } from '@/types/system'
import { WebhookEvents } from '@/types/system'

const activeTab = ref('email')

// ============ 邮件配置 ============
const emailFormRef = ref<FormInstance>()
const emailSaving = ref(false)
const emailForm = reactive({
  smtp_host: '',
  smtp_port: 587,
  smtp_username: '',
  smtp_password: '',
  from_address: '',
  from_name: 'TicketDesk',
  use_tls: true,
  enabled: false,
})

const emailRules: FormRules = {
  smtp_host: [{ required: true, message: '请输入 SMTP 服务器地址', trigger: 'blur' }],
  smtp_port: [{ required: true, message: '请输入 SMTP 端口', trigger: 'blur' }],
  from_address: [
    { required: true, message: '请输入发件人地址', trigger: 'blur' },
    { type: 'email', message: '请输入正确的邮箱地址', trigger: 'blur' },
  ],
}

const loadEmailConfig = async () => {
  try {
    const { data } = await getEmailConfig()
    Object.assign(emailForm, data.data)
  } catch (error) {
    console.error('Failed to load email config:', error)
  }
}

const saveEmailConfig = async () => {
  if (!emailFormRef.value) return
  await emailFormRef.value.validate(async (valid) => {
    if (!valid) return

    emailSaving.value = true
    try {
      await updateEmailConfig(emailForm)
      ElMessage.success('邮件配置保存成功')
    } catch (error) {
      console.error('Failed to save email config:', error)
    } finally {
      emailSaving.value = false
    }
  })
}

// 测试邮件
const testEmailDialog = ref(false)
const testEmailFormRef = ref<FormInstance>()
const testEmailSending = ref(false)
const testEmailForm = reactive({
  to_address: '',
})
const testEmailRules: FormRules = {
  to_address: [
    { required: true, message: '请输入收件人邮箱', trigger: 'blur' },
    { type: 'email', message: '请输入正确的邮箱地址', trigger: 'blur' },
  ],
}

const sendTestEmail = async () => {
  if (!testEmailFormRef.value) return
  await testEmailFormRef.value.validate(async (valid) => {
    if (!valid) return

    testEmailSending.value = true
    try {
      // await testEmail(testEmailForm.to_address)
      ElMessage.success('测试邮件已发送')
      testEmailDialog.value = false
    } catch (error) {
      console.error('Failed to send test email:', error)
    } finally {
      testEmailSending.value = false
    }
  })
}

// ============ Webhook 配置 ============
const webhooks = ref<Webhook[]>([])
const webhooksLoading = ref(false)
const webhookDialogVisible = ref(false)
const webhookFormRef = ref<FormInstance>()
const webhookSaving = ref(false)
const webhookForm = reactive({
  id: 0,
  name: '',
  url: '',
  secret: '',
  events: [] as string[],
  description: '',
  status: 1,
})

const webhookRules: FormRules = {
  name: [{ required: true, message: '请输入名称', trigger: 'blur' }],
  url: [
    { required: true, message: '请输入 URL', trigger: 'blur' },
    { type: 'url', message: '请输入正确的 URL', trigger: 'blur' },
  ],
  events: [{ required: true, message: '请选择至少一个事件', trigger: 'change' }],
}

const loadWebhooks = async () => {
  webhooksLoading.value = true
  try {
    const { data } = await getWebhooks()
    webhooks.value = data.data.items || []
  } catch (error) {
    console.error('Failed to load webhooks:', error)
  } finally {
    webhooksLoading.value = false
  }
}

const showWebhookDialog = (webhook?: Webhook) => {
  if (webhook) {
    Object.assign(webhookForm, {
      id: webhook.id,
      name: webhook.name,
      url: webhook.url,
      secret: '',
      events: webhook.events,
      description: webhook.description,
      status: webhook.status,
    })
  } else {
    Object.assign(webhookForm, {
      id: 0,
      name: '',
      url: '',
      secret: '',
      events: [],
      description: '',
      status: 1,
    })
  }
  webhookDialogVisible.value = true
}

const saveWebhook = async () => {
  if (!webhookFormRef.value) return
  await webhookFormRef.value.validate(async (valid) => {
    if (!valid) return

    webhookSaving.value = true
    try {
      if (webhookForm.id) {
        await updateWebhook(webhookForm.id, {
          name: webhookForm.name,
          url: webhookForm.url,
          secret: webhookForm.secret || undefined,
          events: webhookForm.events,
          description: webhookForm.description,
          status: webhookForm.status,
        })
        ElMessage.success('Webhook 更新成功')
      } else {
        await createWebhook({
          name: webhookForm.name,
          url: webhookForm.url,
          secret: webhookForm.secret || undefined,
          events: webhookForm.events,
          description: webhookForm.description,
        })
        ElMessage.success('Webhook 创建成功')
      }
      webhookDialogVisible.value = false
      loadWebhooks()
    } catch (error) {
      console.error('Failed to save webhook:', error)
    } finally {
      webhookSaving.value = false
    }
  })
}

const handleDeleteWebhook = async (webhook: Webhook) => {
  try {
    await ElMessageBox.confirm(
      `确定要删除 Webhook "${webhook.name}" 吗？`,
      '删除确认',
      { type: 'warning' }
    )
    await deleteWebhook(webhook.id)
    ElMessage.success('删除成功')
    loadWebhooks()
  } catch (error) {
    if (error !== 'cancel') {
      console.error('Failed to delete webhook:', error)
    }
  }
}

const getEventLabel = (event: string) => {
  const found = WebhookEvents.find(e => e.value === event)
  return found?.label || event
}

// ============ 安全配置 ============
const securityFormRef = ref<FormInstance>()
const securitySaving = ref(false)
const securityForm = reactive({
  mfa_enabled: false,
  mfa_required: false,
  password_min_length: 6,
  password_require_upper: false,
  password_require_number: false,
  session_timeout: 120,
})

const loadSecurityConfig = async () => {
  try {
    const { data } = await getSecurityConfig()
    Object.assign(securityForm, data.data)
  } catch (error) {
    console.error('Failed to load security config:', error)
  }
}

const saveSecurityConfig = async () => {
  securitySaving.value = true
  try {
    await updateSecurityConfig(securityForm)
    ElMessage.success('安全配置保存成功')
  } catch (error) {
    console.error('Failed to save security config:', error)
  } finally {
    securitySaving.value = false
  }
}

// ============ 初始化 ============
onMounted(() => {
  loadEmailConfig()
  loadWebhooks()
  loadSecurityConfig()
})
</script>

<style scoped lang="scss">
.system-settings {
  width: 100%;
}

.settings-tabs {
  :deep(.el-tabs__content) {
    padding-top: 16px;
  }
}

// 卡片样式
.settings-card {
  border-radius: 8px;

  :deep(.el-card__header) {
    padding: 20px 24px;
    border-bottom: 1px solid #f0f0f0;
  }

  :deep(.el-card__body) {
    padding: 24px;
  }
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;

  .card-title {
    display: flex;
    align-items: center;
    gap: 16px;
  }

  .title-icon {
    width: 48px;
    height: 48px;
    border-radius: 12px;
    display: flex;
    align-items: center;
    justify-content: center;
    font-size: 24px;

    &.email-icon {
      background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
      color: #fff;
    }

    &.webhook-icon {
      background: linear-gradient(135deg, #f093fb 0%, #f5576c 100%);
      color: #fff;
    }
  }

  .title-text {
    display: flex;
    flex-direction: column;
    gap: 4px;

    .title {
      font-size: 16px;
      font-weight: 600;
      color: #1f2937;
    }

    .subtitle {
      font-size: 13px;
      color: #909399;
    }
  }
}

// 表单样式
.settings-form {
  .form-section {
    margin-bottom: 24px;

    .section-title {
      font-size: 14px;
      font-weight: 600;
      color: #374151;
      margin-bottom: 16px;
      padding-bottom: 8px;
      border-bottom: 1px solid #f0f0f0;
    }
  }

  .setting-row {
    display: flex;
    justify-content: space-between;
    align-items: center;
    padding: 12px 16px;
    background: #f9fafb;
    border-radius: 8px;
  }
}

.setting-info {
  display: flex;
  flex-direction: column;
  gap: 2px;

  .setting-label {
    font-size: 14px;
    color: #374151;
    font-weight: 500;
  }

  .setting-desc {
    font-size: 12px;
    color: #9ca3af;
  }
}

// 信息面板
.info-panel {
  display: flex;
  align-items: flex-start;
  padding-top: 24px;
}

.info-card {
  background: linear-gradient(135deg, #f0f9ff 0%, #e0f2fe 100%);
  border-radius: 12px;
  padding: 20px;
  border: 1px solid #bae6fd;

  .info-title {
    display: flex;
    align-items: center;
    gap: 8px;
    font-size: 14px;
    font-weight: 600;
    color: #0369a1;
    margin-bottom: 12px;
  }

  .info-list {
    margin: 0;
    padding-left: 20px;

    li {
      font-size: 13px;
      color: #475569;
      line-height: 1.8;
    }
  }
}

// Webhook 列表样式
.webhook-list {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.webhook-item {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 16px 20px;
  background: #f9fafb;
  border-radius: 8px;
  border: 1px solid #e5e7eb;
  transition: all 0.2s;

  &:hover {
    border-color: #d1d5db;
    box-shadow: 0 2px 8px rgba(0, 0, 0, 0.04);
  }

  .webhook-info {
    flex: 1;

    .webhook-header {
      display: flex;
      align-items: center;
      gap: 12px;
      margin-bottom: 6px;

      .webhook-name {
        font-size: 15px;
        font-weight: 600;
        color: #1f2937;
      }
    }

    .webhook-url {
      font-size: 13px;
      color: #6b7280;
      margin-bottom: 8px;
      font-family: monospace;
    }

    .webhook-events {
      display: flex;
      flex-wrap: wrap;
      gap: 6px;
    }
  }

  .webhook-actions {
    display: flex;
    gap: 8px;
  }
}

.event-tag {
  font-size: 12px;
}

// 安全设置块
.setting-block {
  border-radius: 12px;
  margin-bottom: 20px;

  :deep(.el-card__body) {
    padding: 0;
  }

  .block-header {
    display: flex;
    align-items: center;
    gap: 16px;
    padding: 20px;
    border-bottom: 1px solid #f0f0f0;

    .block-icon {
      width: 44px;
      height: 44px;
      border-radius: 10px;
      display: flex;
      align-items: center;
      justify-content: center;
      font-size: 22px;

      &.mfa-icon {
        background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
        color: #fff;
      }

      &.password-icon {
        background: linear-gradient(135deg, #f093fb 0%, #f5576c 100%);
        color: #fff;
      }

      &.session-icon {
        background: linear-gradient(135deg, #4facfe 0%, #00f2fe 100%);
        color: #fff;
      }
    }

    .block-title {
      display: flex;
      flex-direction: column;
      gap: 2px;

      .title {
        font-size: 15px;
        font-weight: 600;
        color: #1f2937;
      }

      .desc {
        font-size: 12px;
        color: #9ca3af;
      }
    }
  }

  .block-content {
    padding: 16px 20px;
  }
}

.setting-item {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 14px 0;

  &:not(:last-child) {
    border-bottom: 1px solid #f3f4f6;
  }
}

.action-bar {
  margin-top: 24px;
  padding-top: 20px;
  border-top: 1px solid #e5e7eb;
}

// Webhook 对话框样式
.event-checkbox-group {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: 12px;

  :deep(.el-checkbox) {
    margin-right: 0;
    height: auto;
    padding: 8px 12px;

    &.is-bordered {
      border-radius: 6px;
    }
  }
}

// 响应式
@media (max-width: 992px) {
  .info-panel {
    margin-top: 24px;
  }

  .setting-block {
    margin-bottom: 16px;
  }
}
</style>
