// Package middleware 提供 HTTP 中间件
package middleware

import (
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/kerbos/ticketdesk/pkg/logger"
)

// LoggerMiddleware 请求日志中间件
func LoggerMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		query := c.Request.URL.RawQuery

		c.Next()

		latency := time.Since(start)
		statusCode := c.Writer.Status()

		fields := []zap.Field{
			zap.Int("status", statusCode),
			zap.String("method", c.Request.Method),
			zap.String("path", path),
			zap.String("query", query),
			zap.String("ip", c.ClientIP()),
			zap.String("user_agent", c.Request.UserAgent()),
			zap.Duration("latency", latency),
		}

		if userID, exists := c.Get("user_id"); exists {
			fields = append(fields, zap.Any("user_id", userID))
		}

		switch {
		case len(c.Errors) > 0:
			fields = append(fields, zap.String("errors", c.Errors.String()))
			logger.Error("request error", fields...)
		case statusCode >= 500:
			logger.Error("server error", fields...)
		case statusCode >= 400:
			logger.Warn("client error", fields...)
		default:
			logger.Info("request completed", fields...)
		}
	}
}
