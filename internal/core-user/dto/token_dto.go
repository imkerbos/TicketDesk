// Package dto 提供用户模块数据传输对象
package dto

import "time"

// CreateTokenRequest 创建 API token 请求
type CreateTokenRequest struct {
	Name      string `json:"name" binding:"required,max=100"`
	ExpiresIn int64  `json:"expires_in" binding:"omitempty,min=0"` // 秒数, 0 = 永久
}

// CreateTokenResponse 创建 API token 响应（含一次性明文）
type CreateTokenResponse struct {
	ID        uint64     `json:"id"`
	Name      string     `json:"name"`
	Token     string     `json:"token"` // 一次性明文，只在创建时返回
	Prefix    string     `json:"prefix"`
	ExpiresAt *time.Time `json:"expires_at"`
	CreatedAt time.Time  `json:"created_at"`
}

// TokenResponse API token 列表项响应
type TokenResponse struct {
	ID         uint64     `json:"id"`
	Name       string     `json:"name"`
	Prefix     string     `json:"prefix"`
	ExpiresAt  *time.Time `json:"expires_at"`
	LastUsedAt *time.Time `json:"last_used_at"`
	CreatedAt  time.Time  `json:"created_at"`
	IsExpired  bool       `json:"is_expired"`
}
