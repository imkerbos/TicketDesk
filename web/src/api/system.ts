import request from '@/utils/request'
import type {
  SystemConfig,
  EmailConfig,
  SecurityConfig,
  RateLimitConfig,
  LarkConfig,
  UpdateLarkConfigRequest,
  TelegramConfig,
  UpdateTelegramConfigRequest,
  SSOAdminConfig,
  UpdateSSOConfigRequest,
  Webhook,
  CreateWebhookRequest,
  UpdateWebhookRequest,
  WebhookLog,
} from '@/types/system'

// ============ 系统配置 ============

// 获取所有配置
export function getAllConfigs() {
  return request.get<{ data: SystemConfig[] }>('/system/configs')
}

// 获取分类配置
export function getConfigsByCategory(category: string) {
  return request.get<{ data: SystemConfig[] }>('/system/configs/category', {
    params: { category },
  })
}

// 获取单个配置
export function getConfig(key: string) {
  return request.get<{ data: SystemConfig }>(`/system/configs/${key}`)
}

// 获取公共配置（普通用户可访问）
export function getPublicConfig(key: string) {
  return request.get<{ data: SystemConfig }>(`/system/configs/public/${key}`)
}

// 获取配置值（只返回值）
export function getConfigValue(key: string) {
  return request.get<{ data: string }>(`/system/configs/${key}`)
}

// 更新单个配置
export function updateConfig(key: string, value: string) {
  return request.put(`/system/configs/${key}`, { config_value: value })
}

// 批量更新配置
export function batchUpdateConfigs(configs: Record<string, string>) {
  return request.put('/system/configs', { configs })
}

// ============ 邮件配置 ============

// 获取邮件配置
export function getEmailConfig() {
  return request.get<{ data: EmailConfig }>('/system/email')
}

// 更新邮件配置
export function updateEmailConfig(config: Partial<EmailConfig> & { smtp_host: string; smtp_port: number; from_address: string }) {
  return request.put('/system/email', config)
}

// 测试邮件发送
export function testEmail(toAddress: string) {
  return request.post('/system/email/test', { to_address: toAddress })
}

// ============ 安全配置 ============

// 获取安全配置
export function getSecurityConfig() {
  return request.get<{ data: SecurityConfig }>('/system/security')
}

// 更新安全配置
export function updateSecurityConfig(config: Partial<SecurityConfig>) {
  return request.put('/system/security', config)
}

// ============ 限流配置 ============

// 获取限流配置
export function getRateLimitConfig() {
  return request.get<{ data: RateLimitConfig }>('/system/ratelimit')
}

// 更新限流配置
export function updateRateLimitConfig(config: Partial<RateLimitConfig>) {
  return request.put('/system/ratelimit', config)
}

// ============ Webhook 管理 ============

// 获取 Webhook 列表
export function getWebhooks(params?: { page?: number; page_size?: number; status?: number; event?: string }) {
  return request.get<{
    data: {
      items: Webhook[]
      total: number
      page: number
      page_size: number
    }
  }>('/system/webhooks', { params })
}

// 获取单个 Webhook
export function getWebhook(id: number) {
  return request.get<{ data: Webhook }>(`/system/webhooks/${id}`)
}

// 创建 Webhook
export function createWebhook(data: CreateWebhookRequest) {
  return request.post<{ data: Webhook }>('/system/webhooks', data)
}

// 更新 Webhook
export function updateWebhook(id: number, data: UpdateWebhookRequest) {
  return request.put<{ data: Webhook }>(`/system/webhooks/${id}`, data)
}

// 删除 Webhook
export function deleteWebhook(id: number) {
  return request.delete(`/system/webhooks/${id}`)
}

// ============ Webhook 日志 ============

// 获取 Webhook 日志列表
export function getWebhookLogs(params?: {
  page?: number
  page_size?: number
  webhook_id?: number
  event?: string
  status?: number
}) {
  return request.get<{
    data: {
      items: WebhookLog[]
      total: number
      page: number
      page_size: number
    }
  }>('/system/webhook-logs', { params })
}

// ============ 飞书配置 ============

// 获取飞书配置
export function getLarkConfig() {
  return request.get<{ data: LarkConfig }>('/system/lark')
}

// 更新飞书配置
export function updateLarkConfig(config: UpdateLarkConfigRequest) {
  return request.put('/system/lark', config)
}

// 测试飞书通知
export function testLark() {
  return request.post('/system/lark/test')
}

// ============ Telegram 配置 ============

// 获取 Telegram 配置
export function getTelegramConfig() {
  return request.get<{ data: TelegramConfig }>('/system/telegram')
}

// 更新 Telegram 配置
export function updateTelegramConfig(config: UpdateTelegramConfigRequest) {
  return request.put('/system/telegram', config)
}

// 测试 Telegram 通知
export function testTelegram() {
  return request.post('/system/telegram/test')
}

// ============ SSO 配置 ============

// 获取 SSO 配置
export function getSSOAdminConfig() {
  return request.get<{ data: SSOAdminConfig }>('/system/sso')
}

// 更新 SSO 配置
export function updateSSOConfig(config: UpdateSSOConfigRequest) {
  return request.put('/system/sso', config)
}
