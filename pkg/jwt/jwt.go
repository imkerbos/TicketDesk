// Package jwt 提供 JWT 认证功能
package jwt

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"

	"github.com/kerbos/ticketdesk/pkg/config"
)

var (
	// ErrTokenExpired Token 已过期
	ErrTokenExpired = errors.New("token has expired")
	// ErrTokenInvalid Token 无效
	ErrTokenInvalid = errors.New("token is invalid")
	// ErrTokenMalformed Token 格式错误
	ErrTokenMalformed = errors.New("token is malformed")
)

// Claims 自定义 JWT Claims
type Claims struct {
	UserID   uint64 `json:"user_id"`
	Username string `json:"username"`
	jwt.RegisteredClaims
}

// Manager JWT 管理器
type Manager struct {
	secret        []byte
	accessExpire  time.Duration
	refreshExpire time.Duration
	issuer        string
}

// NewManager 创建 JWT 管理器
func NewManager(cfg *config.JWTConfig) *Manager {
	return &Manager{
		secret:        []byte(cfg.Secret),
		accessExpire:  time.Duration(cfg.AccessExpire) * time.Second,
		refreshExpire: time.Duration(cfg.RefreshExpire) * time.Second,
		issuer:        cfg.Issuer,
	}
}

// GenerateAccessToken 生成 Access Token
func (m *Manager) GenerateAccessToken(userID uint64, username string) (string, error) {
	return m.generateToken(userID, username, m.accessExpire)
}

// GenerateRefreshToken 生成 Refresh Token
func (m *Manager) GenerateRefreshToken(userID uint64, username string) (string, error) {
	return m.generateToken(userID, username, m.refreshExpire)
}

// generateToken 生成 Token
func (m *Manager) generateToken(userID uint64, username string, expire time.Duration) (string, error) {
	now := time.Now()
	claims := Claims{
		UserID:   userID,
		Username: username,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    m.issuer,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(expire)),
			NotBefore: jwt.NewNumericDate(now),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(m.secret)
}

// ParseToken 解析 Token
func (m *Manager) ParseToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		return m.secret, nil
	})

	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, ErrTokenExpired
		}
		if errors.Is(err, jwt.ErrTokenMalformed) {
			return nil, ErrTokenMalformed
		}
		return nil, ErrTokenInvalid
	}

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}

	return nil, ErrTokenInvalid
}

// RefreshToken 刷新 Token
func (m *Manager) RefreshToken(tokenString string) (string, error) {
	claims, err := m.ParseToken(tokenString)
	if err != nil && !errors.Is(err, ErrTokenExpired) {
		return "", err
	}

	return m.GenerateAccessToken(claims.UserID, claims.Username)
}
