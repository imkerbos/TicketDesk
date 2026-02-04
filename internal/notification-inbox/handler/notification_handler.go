// Package handler 提供站内通知 HTTP 处理层
package handler

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/kerbos/ticketdesk/internal/api/response"
	"github.com/kerbos/ticketdesk/internal/notification-inbox/dto"
	"github.com/kerbos/ticketdesk/internal/notification-inbox/service"
)

// NotificationHandler 通知 HTTP 处理器
type NotificationHandler struct {
	svc service.NotificationService
}

// NewNotificationHandler 创建通知处理器
func NewNotificationHandler(svc service.NotificationService) *NotificationHandler {
	return &NotificationHandler{svc: svc}
}

// HandleListNotifications 获取通知列表
func (h *NotificationHandler) HandleListNotifications(c *gin.Context) {
	userID, _ := c.Get("user_id")

	var req dto.ListNotificationsRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}

	notifications, total, err := h.svc.ListNotifications(c.Request.Context(), userID.(uint64), &req)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}

	response.SuccessWithPage(c, notifications, total, req.GetDefaultPage(), req.GetDefaultPageSize())
}

// HandleGetUnreadCount 获取未读通知数量
func (h *NotificationHandler) HandleGetUnreadCount(c *gin.Context) {
	userID, _ := c.Get("user_id")

	count, err := h.svc.GetUnreadCount(c.Request.Context(), userID.(uint64))
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}

	response.Success(c, dto.UnreadCountResponse{Count: count})
}

// HandleMarkAsRead 标记通知为已读
func (h *NotificationHandler) HandleMarkAsRead(c *gin.Context) {
	userID, _ := c.Get("user_id")

	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的通知ID")
		return
	}

	if err := h.svc.MarkAsRead(c.Request.Context(), id, userID.(uint64)); err != nil {
		response.InternalError(c, err.Error())
		return
	}

	response.Success(c, nil)
}

// HandleMarkAllAsRead 全部标记为已读
func (h *NotificationHandler) HandleMarkAllAsRead(c *gin.Context) {
	userID, _ := c.Get("user_id")

	if err := h.svc.MarkAllAsRead(c.Request.Context(), userID.(uint64)); err != nil {
		response.InternalError(c, err.Error())
		return
	}

	response.Success(c, nil)
}

// HandleDeleteNotification 删除通知
func (h *NotificationHandler) HandleDeleteNotification(c *gin.Context) {
	userID, _ := c.Get("user_id")

	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的通知ID")
		return
	}

	if err := h.svc.DeleteNotification(c.Request.Context(), id, userID.(uint64)); err != nil {
		response.InternalError(c, err.Error())
		return
	}

	response.Success(c, nil)
}
