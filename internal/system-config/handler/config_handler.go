// Package handler 提供系统配置 HTTP 处理层
package handler

import (
	"errors"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/kerbos/ticketdesk/internal/api/response"
	"github.com/kerbos/ticketdesk/internal/system-config/dto"
	"github.com/kerbos/ticketdesk/internal/system-config/service"
)

// ConfigHandler 系统配置处理器
type ConfigHandler struct {
	configService service.ConfigService
}

// NewConfigHandler 创建系统配置处理器
func NewConfigHandler(configService service.ConfigService) *ConfigHandler {
	return &ConfigHandler{configService: configService}
}

// ============ 配置管理 ============

// HandleGetConfig 获取单个配置
func (h *ConfigHandler) HandleGetConfig(c *gin.Context) {
	key := c.Param("key")
	if key == "" {
		response.BadRequest(c, "配置键不能为空")
		return
	}

	config, err := h.configService.GetConfig(c.Request.Context(), key)
	if err != nil {
		if errors.Is(err, service.ErrConfigNotFound) {
			response.NotFound(c, err.Error())
			return
		}
		response.InternalError(c, err.Error())
		return
	}

	response.Success(c, config)
}

// HandleGetConfigsByCategory 获取分类下的所有配置
func (h *ConfigHandler) HandleGetConfigsByCategory(c *gin.Context) {
	category := c.Query("category")
	if category == "" {
		response.BadRequest(c, "分类不能为空")
		return
	}

	configs, err := h.configService.GetConfigsByCategory(c.Request.Context(), category)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}

	response.Success(c, configs)
}

// HandleGetAllConfigs 获取所有配置
func (h *ConfigHandler) HandleGetAllConfigs(c *gin.Context) {
	configs, err := h.configService.GetAllConfigs(c.Request.Context())
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}

	response.Success(c, configs)
}

// HandleUpdateConfig 更新单个配置
func (h *ConfigHandler) HandleUpdateConfig(c *gin.Context) {
	key := c.Param("key")
	if key == "" {
		response.BadRequest(c, "配置键不能为空")
		return
	}

	var req dto.UpdateConfigRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数错误: "+err.Error())
		return
	}

	userID := c.GetUint64("user_id")
	if err := h.configService.UpdateConfig(c.Request.Context(), key, req.ConfigValue, userID); err != nil {
		response.InternalError(c, err.Error())
		return
	}

	response.Success(c, nil)
}

// HandleBatchUpdateConfigs 批量更新配置
func (h *ConfigHandler) HandleBatchUpdateConfigs(c *gin.Context) {
	var req dto.BatchUpdateConfigRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数错误: "+err.Error())
		return
	}

	userID := c.GetUint64("user_id")
	if err := h.configService.BatchUpdateConfigs(c.Request.Context(), req.Configs, userID); err != nil {
		response.InternalError(c, err.Error())
		return
	}

	response.Success(c, nil)
}

// ============ 邮件配置 ============

// HandleGetEmailConfig 获取邮件配置
func (h *ConfigHandler) HandleGetEmailConfig(c *gin.Context) {
	config, err := h.configService.GetEmailConfig(c.Request.Context())
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}

	response.Success(c, config)
}

// HandleUpdateEmailConfig 更新邮件配置
func (h *ConfigHandler) HandleUpdateEmailConfig(c *gin.Context) {
	var req dto.UpdateEmailConfigRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数错误: "+err.Error())
		return
	}

	userID := c.GetUint64("user_id")
	if err := h.configService.UpdateEmailConfig(c.Request.Context(), &req, userID); err != nil {
		response.InternalError(c, err.Error())
		return
	}

	response.Success(c, nil)
}

// ============ 安全配置 ============

// HandleGetSecurityConfig 获取安全配置
func (h *ConfigHandler) HandleGetSecurityConfig(c *gin.Context) {
	config, err := h.configService.GetSecurityConfig(c.Request.Context())
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}

	response.Success(c, config)
}

// HandleUpdateSecurityConfig 更新安全配置
func (h *ConfigHandler) HandleUpdateSecurityConfig(c *gin.Context) {
	var req dto.UpdateSecurityConfigRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数错误: "+err.Error())
		return
	}

	userID := c.GetUint64("user_id")
	if err := h.configService.UpdateSecurityConfig(c.Request.Context(), &req, userID); err != nil {
		response.InternalError(c, err.Error())
		return
	}

	response.Success(c, nil)
}

// ============ Webhook 管理 ============

// HandleCreateWebhook 创建 Webhook
func (h *ConfigHandler) HandleCreateWebhook(c *gin.Context) {
	var req dto.CreateWebhookRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数错误: "+err.Error())
		return
	}

	userID := c.GetUint64("user_id")
	webhook, err := h.configService.CreateWebhook(c.Request.Context(), &req, userID)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}

	response.Created(c, webhook)
}

// HandleGetWebhook 获取 Webhook
func (h *ConfigHandler) HandleGetWebhook(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的 Webhook ID")
		return
	}

	webhook, err := h.configService.GetWebhook(c.Request.Context(), id)
	if err != nil {
		if errors.Is(err, service.ErrWebhookNotFound) {
			response.NotFound(c, err.Error())
			return
		}
		response.InternalError(c, err.Error())
		return
	}

	response.Success(c, webhook)
}

// HandleUpdateWebhook 更新 Webhook
func (h *ConfigHandler) HandleUpdateWebhook(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的 Webhook ID")
		return
	}

	var req dto.UpdateWebhookRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数错误: "+err.Error())
		return
	}

	webhook, err := h.configService.UpdateWebhook(c.Request.Context(), id, &req)
	if err != nil {
		if errors.Is(err, service.ErrWebhookNotFound) {
			response.NotFound(c, err.Error())
			return
		}
		response.InternalError(c, err.Error())
		return
	}

	response.Success(c, webhook)
}

// HandleDeleteWebhook 删除 Webhook
func (h *ConfigHandler) HandleDeleteWebhook(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的 Webhook ID")
		return
	}

	if err := h.configService.DeleteWebhook(c.Request.Context(), id); err != nil {
		if errors.Is(err, service.ErrWebhookNotFound) {
			response.NotFound(c, err.Error())
			return
		}
		response.InternalError(c, err.Error())
		return
	}

	response.Success(c, nil)
}

// HandleListWebhooks 查询 Webhook 列表
func (h *ConfigHandler) HandleListWebhooks(c *gin.Context) {
	var req dto.ListWebhooksRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.BadRequest(c, "请求参数错误: "+err.Error())
		return
	}

	webhooks, total, err := h.configService.ListWebhooks(c.Request.Context(), &req)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}

	response.SuccessWithPage(c, webhooks, total, req.GetDefaultPage(), req.GetDefaultPageSize())
}

// ============ Webhook 日志 ============

// HandleListWebhookLogs 查询 Webhook 日志
func (h *ConfigHandler) HandleListWebhookLogs(c *gin.Context) {
	var req dto.ListWebhookLogsRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.BadRequest(c, "请求参数错误: "+err.Error())
		return
	}

	logs, total, err := h.configService.ListWebhookLogs(c.Request.Context(), &req)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}

	response.SuccessWithPage(c, logs, total, req.GetDefaultPage(), req.GetDefaultPageSize())
}
