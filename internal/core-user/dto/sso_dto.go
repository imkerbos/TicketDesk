// Package dto 定义用户模块的数据传输对象
package dto

// SSOConfigResponse SSO 配置响应（前端展示用，不含敏感信息）
type SSOConfigResponse struct {
	Enabled      bool   `json:"enabled"`
	ProviderName string `json:"provider_name,omitempty"`
}

// SSOAuthorizeResponse SSO 授权 URL 响应
type SSOAuthorizeResponse struct {
	AuthorizeURL string `json:"authorize_url"`
}

// SSOCallbackRequest SSO 回调请求
type SSOCallbackRequest struct {
	Code  string `json:"code" binding:"required"`
	State string `json:"state" binding:"required"`
}

// SSOUserInfo 从 OIDC ID Token 提取的用户信息（内部使用）
type SSOUserInfo struct {
	Subject     string // OIDC sub claim
	Username    string
	Email       string
	DisplayName string
	AvatarURL   string
}
