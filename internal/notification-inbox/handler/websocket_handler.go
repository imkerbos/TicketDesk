// Package handler 提供 WebSocket 连接处理
package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"go.uber.org/zap"

	ws "github.com/kerbos/ticketdesk/internal/notification-inbox/websocket"
	"github.com/kerbos/ticketdesk/pkg/jwt"
	"github.com/kerbos/ticketdesk/pkg/logger"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true // 开发环境允许所有来源
	},
}

// WebSocketHandler WebSocket 处理器
type WebSocketHandler struct {
	wsManager  *ws.Manager
	jwtManager *jwt.Manager
}

// NewWebSocketHandler 创建 WebSocket 处理器
func NewWebSocketHandler(wsManager *ws.Manager, jwtManager *jwt.Manager) *WebSocketHandler {
	return &WebSocketHandler{
		wsManager:  wsManager,
		jwtManager: jwtManager,
	}
}

// HandleWebSocket 处理 WebSocket 连接
func (h *WebSocketHandler) HandleWebSocket(c *gin.Context) {
	// 从查询参数获取 token
	token := c.Query("token")
	if token == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "缺少认证信息"})
		return
	}

	// 验证 JWT token
	claims, err := h.jwtManager.ParseToken(token)
	if err != nil {
		logger.Warn("ws auth failed", zap.Error(err))
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Token 无效"})
		return
	}

	// 升级为 WebSocket 连接
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		logger.Error("ws upgrade failed", zap.Error(err))
		return
	}

	logger.Info("ws client connected",
		zap.Uint64("user_id", claims.UserID),
		zap.String("username", claims.Username),
	)

	// 注册客户端
	h.wsManager.RegisterClient(claims.UserID, conn)
}
