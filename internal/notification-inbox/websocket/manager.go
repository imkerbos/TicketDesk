// Package websocket 提供 WebSocket 连接管理
package websocket

import (
	"encoding/json"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/kerbos/ticketdesk/pkg/logger"
	"go.uber.org/zap"
)

const (
	// 写入超时
	writeWait = 10 * time.Second
	// Pong 响应超时
	pongWait = 60 * time.Second
	// Ping 发送间隔（必须小于 pongWait）
	pingPeriod = 30 * time.Second
	// 最大消息大小
	maxMessageSize = 512
)

// WSMessage WebSocket 消息
type WSMessage struct {
	Type string      `json:"type"` // notification, ping, pong
	Data interface{} `json:"data,omitempty"`
}

// WSClient WebSocket 客户端连接
type WSClient struct {
	userID  uint64
	conn    *websocket.Conn
	send    chan []byte
	manager *Manager
}

// Manager WebSocket 连接管理器
type Manager struct {
	clients    map[uint64]*WSClient // userID -> client
	register   chan *WSClient
	unregister chan *WSClient
	mu         sync.RWMutex
}

// NewManager 创建 WebSocket 管理器
func NewManager() *Manager {
	return &Manager{
		clients:    make(map[uint64]*WSClient),
		register:   make(chan *WSClient),
		unregister: make(chan *WSClient),
	}
}

// Run 启动管理器事件循环
func (m *Manager) Run() {
	for {
		select {
		case client := <-m.register:
			m.mu.Lock()
			// 关闭同一用户的旧连接
			if old, ok := m.clients[client.userID]; ok {
				close(old.send)
				old.conn.Close()
			}
			m.clients[client.userID] = client
			m.mu.Unlock()
			logger.Info("websocket client registered",
				zap.Uint64("user_id", client.userID),
			)

		case client := <-m.unregister:
			m.mu.Lock()
			if existing, ok := m.clients[client.userID]; ok && existing == client {
				delete(m.clients, client.userID)
				close(client.send)
			}
			m.mu.Unlock()
			logger.Info("websocket client unregistered",
				zap.Uint64("user_id", client.userID),
			)
		}
	}
}

// SendToUser 向指定用户推送消息
func (m *Manager) SendToUser(userID uint64, msg *WSMessage) {
	m.mu.RLock()
	client, ok := m.clients[userID]
	m.mu.RUnlock()

	if !ok {
		return
	}

	data, err := json.Marshal(msg)
	if err != nil {
		logger.Error("failed to marshal ws message", zap.Error(err))
		return
	}

	select {
	case client.send <- data:
	default:
		// 通道已满，放弃发送
		logger.Warn("ws send channel full, dropping message",
			zap.Uint64("user_id", userID),
		)
	}
}

// RegisterClient 注册新客户端
func (m *Manager) RegisterClient(userID uint64, conn *websocket.Conn) {
	client := &WSClient{
		userID:  userID,
		conn:    conn,
		send:    make(chan []byte, 256),
		manager: m,
	}

	m.register <- client

	// 启动读写协程
	go client.readPump()
	go client.writePump()
}

// readPump 读取客户端消息（处理心跳 pong）
func (c *WSClient) readPump() {
	defer func() {
		c.manager.unregister <- c
		c.conn.Close()
	}()

	c.conn.SetReadLimit(maxMessageSize)
	c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error {
		c.conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	for {
		_, _, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseNormalClosure) {
				logger.Warn("ws unexpected close",
					zap.Uint64("user_id", c.userID),
					zap.Error(err),
				)
			}
			break
		}
	}
}

// writePump 向客户端写入消息（处理心跳 ping）
func (c *WSClient) writePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				// 通道已关闭
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			if err := c.conn.WriteMessage(websocket.TextMessage, message); err != nil {
				return
			}

		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}
