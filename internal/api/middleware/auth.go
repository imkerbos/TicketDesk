// Package middleware 提供 HTTP 中间件
package middleware

import (
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/kerbos/ticketdesk/internal/api/response"
	"github.com/kerbos/ticketdesk/internal/core-user/service"
	"github.com/kerbos/ticketdesk/pkg/jwt"
)

// AuthMiddleware JWT 认证中间件，同时支持 PAT 鉴权
func AuthMiddleware(jwtManager *jwt.Manager, tokenSvc service.APITokenService) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			response.Unauthorized(c, "缺少认证信息")
			c.Abort()
			return
		}

		// Bearer token
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			response.Unauthorized(c, "认证格式错误")
			c.Abort()
			return
		}

		raw := parts[1]

		// API Token (PAT) 路径：前缀为 td_pat_ 时走 PAT 鉴权
		if service.IsAPIToken(raw) {
			user, _, err := tokenSvc.Authenticate(c.Request.Context(), raw)
			if err != nil {
				response.Unauthorized(c, "API token 无效或已过期")
				c.Abort()
				return
			}
			c.Set("user_id", user.ID)
			c.Set("username", user.Username)
			c.Set("is_pat", true)
			c.Next()
			return
		}

		// 否则走 JWT 鉴权（原有逻辑）
		claims, err := jwtManager.ParseToken(raw)
		if err != nil {
			switch err {
			case jwt.ErrTokenExpired:
				response.Unauthorized(c, "Token 已过期")
			case jwt.ErrTokenMalformed:
				response.Unauthorized(c, "Token 格式错误")
			default:
				response.Unauthorized(c, "Token 无效")
			}
			c.Abort()
			return
		}

		// 将用户信息存入上下文
		c.Set("user_id", claims.UserID)
		c.Set("username", claims.Username)

		c.Next()
	}
}

// CORSMiddleware 跨域中间件
func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Authorization")
		c.Header("Access-Control-Max-Age", "86400")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}

// RecoveryMiddleware 异常恢复中间件
func RecoveryMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				response.InternalError(c, "服务器内部错误")
				c.Abort()
			}
		}()
		c.Next()
	}
}
