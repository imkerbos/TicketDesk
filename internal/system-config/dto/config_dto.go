// Package dto 提供系统配置数据传输对象
package dto

import "time"

// ============ 系统配置 DTO ============

// ConfigResponse 配置响应
type ConfigResponse struct {
	ID          uint64    `json:"id"`
	ConfigKey   string    `json:"config_key"`
	ConfigValue string    `json:"config_value,omitempty"` // 敏感配置会隐藏值
	ConfigType  string    `json:"config_type"`
	Category    string    `json:"category"`
	Description string    `json:"description"`
	IsSecret    bool      `json:"is_secret"`
	UpdatedBy   *uint64   `json:"updated_by"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// UpdateConfigRequest 更新配置请求
type UpdateConfigRequest struct {
	ConfigValue string `json:"config_value"`
}

// BatchUpdateConfigRequest 批量更新配置请求
type BatchUpdateConfigRequest struct {
	Configs map[string]string `json:"configs" binding:"required"`
}

// ListConfigsRequest 查询配置列表请求
type ListConfigsRequest struct {
	Category string `form:"category"`
}

// ============ 邮件配置 DTO ============

// EmailConfig 邮件配置
type EmailConfig struct {
	SMTPHost     string `json:"smtp_host"`
	SMTPPort     int    `json:"smtp_port"`
	SMTPUsername string `json:"smtp_username"`
	SMTPPassword string `json:"smtp_password,omitempty"` // 返回时隐藏
	FromAddress  string `json:"from_address"`
	FromName     string `json:"from_name"`
	UseTLS       bool   `json:"use_tls"`
	Enabled      bool   `json:"enabled"`
}

// UpdateEmailConfigRequest 更新邮件配置请求
type UpdateEmailConfigRequest struct {
	SMTPHost     string `json:"smtp_host" binding:"required"`
	SMTPPort     int    `json:"smtp_port" binding:"required,min=1,max=65535"`
	SMTPUsername string `json:"smtp_username"`
	SMTPPassword string `json:"smtp_password"`
	FromAddress  string `json:"from_address" binding:"required,email"`
	FromName     string `json:"from_name"`
	UseTLS       bool   `json:"use_tls"`
	Enabled      bool   `json:"enabled"`
}

// TestEmailRequest 测试邮件发送请求
type TestEmailRequest struct {
	ToAddress string `json:"to_address" binding:"required,email"`
}

// ============ Webhook DTO ============

// WebhookResponse Webhook 响应
type WebhookResponse struct {
	ID          uint64            `json:"id"`
	Name        string            `json:"name"`
	URL         string            `json:"url"`
	Events      []string          `json:"events"`
	Headers     map[string]string `json:"headers,omitempty"`
	Status      int8              `json:"status"`
	Description string            `json:"description"`
	CreatedBy   uint64            `json:"created_by"`
	CreatedAt   time.Time         `json:"created_at"`
	UpdatedAt   time.Time         `json:"updated_at"`
}

// CreateWebhookRequest 创建 Webhook 请求
type CreateWebhookRequest struct {
	Name        string            `json:"name" binding:"required,max=100"`
	URL         string            `json:"url" binding:"required,url,max=500"`
	Secret      string            `json:"secret"`
	Events      []string          `json:"events" binding:"required,min=1"`
	Headers     map[string]string `json:"headers"`
	Description string            `json:"description"`
}

// UpdateWebhookRequest 更新 Webhook 请求
type UpdateWebhookRequest struct {
	Name        *string           `json:"name" binding:"omitempty,max=100"`
	URL         *string           `json:"url" binding:"omitempty,url,max=500"`
	Secret      *string           `json:"secret"`
	Events      []string          `json:"events"`
	Headers     map[string]string `json:"headers"`
	Status      *int8             `json:"status"`
	Description *string           `json:"description"`
}

// ListWebhooksRequest 查询 Webhook 列表请求
type ListWebhooksRequest struct {
	Page     int    `form:"page" binding:"omitempty,min=1"`
	PageSize int    `form:"page_size" binding:"omitempty,min=1,max=100"`
	Status   *int8  `form:"status"`
	Event    string `form:"event"`
}

// GetDefaultPage 获取默认页码
func (r *ListWebhooksRequest) GetDefaultPage() int {
	if r.Page <= 0 {
		return 1
	}
	return r.Page
}

// GetDefaultPageSize 获取默认每页数量
func (r *ListWebhooksRequest) GetDefaultPageSize() int {
	if r.PageSize <= 0 {
		return 20
	}
	return r.PageSize
}

// WebhookLogResponse Webhook 日志响应
type WebhookLogResponse struct {
	ID           uint64    `json:"id"`
	WebhookID    uint64    `json:"webhook_id"`
	Event        string    `json:"event"`
	Payload      string    `json:"payload"`
	ResponseCode int       `json:"response_code"`
	ResponseBody string    `json:"response_body"`
	Status       int8      `json:"status"`
	ErrorMessage string    `json:"error_message"`
	RetryCount   int       `json:"retry_count"`
	CreatedAt    time.Time `json:"created_at"`
}

// ListWebhookLogsRequest 查询 Webhook 日志请求
type ListWebhookLogsRequest struct {
	Page      int    `form:"page" binding:"omitempty,min=1"`
	PageSize  int    `form:"page_size" binding:"omitempty,min=1,max=100"`
	WebhookID uint64 `form:"webhook_id"`
	Event     string `form:"event"`
	Status    *int8  `form:"status"`
}

// GetDefaultPage 获取默认页码
func (r *ListWebhookLogsRequest) GetDefaultPage() int {
	if r.Page <= 0 {
		return 1
	}
	return r.Page
}

// GetDefaultPageSize 获取默认每页数量
func (r *ListWebhookLogsRequest) GetDefaultPageSize() int {
	if r.PageSize <= 0 {
		return 20
	}
	return r.PageSize
}

// ============ 安全配置 DTO ============

// SecurityConfig 安全配置
type SecurityConfig struct {
	MFAEnabled            bool `json:"mfa_enabled"`             // 是否启用 MFA
	MFARequired           bool `json:"mfa_required"`            // 是否强制要求 MFA
	PasswordMinLength     int  `json:"password_min_length"`     // 密码最小长度
	PasswordRequireUpper  bool `json:"password_require_upper"`  // 密码是否需要大写字母
	PasswordRequireNumber bool `json:"password_require_number"` // 密码是否需要数字
	SessionTimeout        int  `json:"session_timeout"`         // 会话超时时间（分钟）
}

// UpdateSecurityConfigRequest 更新安全配置请求
type UpdateSecurityConfigRequest struct {
	MFAEnabled            *bool `json:"mfa_enabled"`
	MFARequired           *bool `json:"mfa_required"`
	PasswordMinLength     *int  `json:"password_min_length" binding:"omitempty,min=6,max=32"`
	PasswordRequireUpper  *bool `json:"password_require_upper"`
	PasswordRequireNumber *bool `json:"password_require_number"`
	SessionTimeout        *int  `json:"session_timeout" binding:"omitempty,min=5,max=1440"`
}

// ============ 限流配置 DTO ============

// RateLimitConfig 限流配置
type RateLimitConfig struct {
	WebhookLimit int `json:"webhook_limit"` // Webhook 接口每 IP 每分钟最大请求数
	AuthLimit    int `json:"auth_limit"`    // 认证接口每 IP 每分钟最大请求数
	APILimit     int `json:"api_limit"`     // 全局 API 每 IP 每分钟最大请求数
}

// UpdateRateLimitConfigRequest 更新限流配置请求
type UpdateRateLimitConfigRequest struct {
	WebhookLimit *int `json:"webhook_limit" binding:"omitempty,min=10,max=10000"`
	AuthLimit    *int `json:"auth_limit" binding:"omitempty,min=5,max=1000"`
	APILimit     *int `json:"api_limit" binding:"omitempty,min=50,max=50000"`
}

// ============ 飞书配置 DTO ============

// LarkConfig 飞书配置
type LarkConfig struct {
	Enabled    bool   `json:"enabled"`          // 是否启用飞书通知
	WebhookURL string `json:"webhook_url"`      // 飞书机器人 Webhook URL
	Secret     string `json:"secret,omitempty"` // 签名密钥（返回时隐藏）
}

// UpdateLarkConfigRequest 更新飞书配置请求
type UpdateLarkConfigRequest struct {
	Enabled    bool   `json:"enabled"`
	WebhookURL string `json:"webhook_url" binding:"required_if=Enabled true"`
	Secret     string `json:"secret"` // 为空时不更新
}

// ============ Telegram 配置 DTO ============

// TelegramConfig Telegram 配置
type TelegramConfig struct {
	Enabled  bool   `json:"enabled"`             // 是否启用 Telegram 通知
	BotToken string `json:"bot_token,omitempty"` // Bot Token（返回时隐藏）
	ChatID   string `json:"chat_id"`             // Chat ID
}

// UpdateTelegramConfigRequest 更新 Telegram 配置请求
type UpdateTelegramConfigRequest struct {
	Enabled  bool   `json:"enabled"`
	BotToken string `json:"bot_token"` // 为空时不更新
	ChatID   string `json:"chat_id" binding:"required_if=Enabled true"`
}

// ============ SSO 配置 DTO ============

// SSOClaimMapping 单条 claim 映射
type SSOClaimMapping struct {
	LocalField string `json:"local_field"` // 本地字段名（username/email/display_name/avatar 或自定义扩展属性名）
	ClaimName  string `json:"claim_name"`  // OIDC claim 名称
}

// SSOConfig SSO 配置
type SSOConfig struct {
	Enabled        bool              `json:"enabled"`
	ProviderName   string            `json:"provider_name"`
	ClientID       string            `json:"client_id"`
	ClientSecret   string            `json:"client_secret,omitempty"` // 返回时隐藏
	IssuerURL      string            `json:"issuer_url"`
	RedirectURI    string            `json:"redirect_uri"`
	Scopes         string            `json:"scopes"`
	AutoCreateUser bool              `json:"auto_create_user"`
	DefaultRole    string            `json:"default_role"`
	ClaimMappings  []SSOClaimMapping `json:"claim_mappings"` // 动态 claim 映射列表
}

// UpdateSSOConfigRequest 更新 SSO 配置请求
type UpdateSSOConfigRequest struct {
	Enabled        bool              `json:"enabled"`
	ProviderName   string            `json:"provider_name"`
	ClientID       string            `json:"client_id"`
	ClientSecret   string            `json:"client_secret"` // 为空时不更新
	IssuerURL      string            `json:"issuer_url"`
	RedirectURI    string            `json:"redirect_uri"`
	Scopes         string            `json:"scopes"`
	AutoCreateUser bool              `json:"auto_create_user"`
	DefaultRole    string            `json:"default_role"`
	ClaimMappings  []SSOClaimMapping `json:"claim_mappings"`
}

// ============ 品牌配置 DTO ============

// BrandConfig 品牌配置
type BrandConfig struct {
	SystemName        string `json:"system_name"`
	SystemDescription string `json:"system_description"`
	CopyrightText     string `json:"copyright_text"`
	LogoURL           string `json:"logo_url"`
	FaviconURL        string `json:"favicon_url"`
	LoginTitle        string `json:"login_title"`
	LoginDescription  string `json:"login_description"`
}

// UpdateBrandConfigRequest 更新品牌配置请求
type UpdateBrandConfigRequest struct {
	SystemName        string `json:"system_name" binding:"required,max=50"`
	SystemDescription string `json:"system_description" binding:"max=200"`
	CopyrightText     string `json:"copyright_text" binding:"max=200"`
	LoginTitle        string `json:"login_title" binding:"max=100"`
	LoginDescription  string `json:"login_description" binding:"max=500"`
}
