// Package handler 提供系统配置 HTTP 处理层
package handler

import (
	"errors"
	"fmt"
	"path/filepath"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/kerbos/ticketdesk/internal/api/response"
	"github.com/kerbos/ticketdesk/internal/notification/lark"
	"github.com/kerbos/ticketdesk/internal/notification/telegram"
	"github.com/kerbos/ticketdesk/internal/system-config/dto"
	"github.com/kerbos/ticketdesk/internal/system-config/service"
	"github.com/kerbos/ticketdesk/pkg/storage"
)

// ConfigHandler 系统配置处理器
type ConfigHandler struct {
	configService   service.ConfigService
	larkService     lark.LarkService
	telegramService telegram.TelegramService
	localStorage    *storage.LocalStorage
}

// NewConfigHandler 创建系统配置处理器
func NewConfigHandler(configService service.ConfigService) *ConfigHandler {
	return &ConfigHandler{configService: configService}
}

// SetLarkService 设置飞书服务（用于测试发送）
func (h *ConfigHandler) SetLarkService(larkService lark.LarkService) {
	h.larkService = larkService
}

// SetTelegramService 设置 Telegram 服务（用于测试发送）
func (h *ConfigHandler) SetTelegramService(telegramService telegram.TelegramService) {
	h.telegramService = telegramService
}

// SetLocalStorage 设置本地存储（用于品牌资源上传）
func (h *ConfigHandler) SetLocalStorage(ls *storage.LocalStorage) {
	h.localStorage = ls
}

// publicConfigWhitelist 允许普通用户读取的配置 key 白名单
var publicConfigWhitelist = map[string]bool{
	"worklog.work_types": true,
}

// HandleGetPublicConfig 获取公共配置（普通用户可访问，受白名单限制）
func (h *ConfigHandler) HandleGetPublicConfig(c *gin.Context) {
	key := c.Param("key")
	if key == "" {
		response.BadRequest(c, "配置键不能为空")
		return
	}

	if !publicConfigWhitelist[key] {
		response.Forbidden(c, "无权访问该配置项")
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

// ============ 限流配置 ============

// HandleGetRateLimitConfig 获取限流配置
func (h *ConfigHandler) HandleGetRateLimitConfig(c *gin.Context) {
	config, err := h.configService.GetRateLimitConfig(c.Request.Context())
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}

	response.Success(c, config)
}

// HandleUpdateRateLimitConfig 更新限流配置
func (h *ConfigHandler) HandleUpdateRateLimitConfig(c *gin.Context) {
	var req dto.UpdateRateLimitConfigRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数错误: "+err.Error())
		return
	}

	userID := c.GetUint64("user_id")
	if err := h.configService.UpdateRateLimitConfig(c.Request.Context(), &req, userID); err != nil {
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
	if bindErr := c.ShouldBindJSON(&req); bindErr != nil {
		response.BadRequest(c, "请求参数错误: "+bindErr.Error())
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

// ============ 飞书配置 ============

// HandleGetLarkConfig 获取飞书配置
func (h *ConfigHandler) HandleGetLarkConfig(c *gin.Context) {
	config, err := h.configService.GetLarkConfig(c.Request.Context())
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}

	response.Success(c, config)
}

// HandleUpdateLarkConfig 更新飞书配置
func (h *ConfigHandler) HandleUpdateLarkConfig(c *gin.Context) {
	var req dto.UpdateLarkConfigRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数错误: "+err.Error())
		return
	}

	userID := c.GetUint64("user_id")
	if err := h.configService.UpdateLarkConfig(c.Request.Context(), &req, userID); err != nil {
		response.InternalError(c, err.Error())
		return
	}

	response.Success(c, nil)
}

// HandleTestLark 测试飞书通知
func (h *ConfigHandler) HandleTestLark(c *gin.Context) {
	if h.larkService == nil {
		response.InternalError(c, "飞书服务未初始化")
		return
	}

	if err := h.larkService.SendTestMessage(c.Request.Context()); err != nil {
		response.InternalError(c, "飞书测试消息发送失败: "+err.Error())
		return
	}

	response.Success(c, nil)
}

// ============ Telegram 配置 ============

// HandleGetTelegramConfig 获取 Telegram 配置
func (h *ConfigHandler) HandleGetTelegramConfig(c *gin.Context) {
	config, err := h.configService.GetTelegramConfig(c.Request.Context())
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}

	response.Success(c, config)
}

// HandleUpdateTelegramConfig 更新 Telegram 配置
func (h *ConfigHandler) HandleUpdateTelegramConfig(c *gin.Context) {
	var req dto.UpdateTelegramConfigRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数错误: "+err.Error())
		return
	}

	userID := c.GetUint64("user_id")
	if err := h.configService.UpdateTelegramConfig(c.Request.Context(), &req, userID); err != nil {
		response.InternalError(c, err.Error())
		return
	}

	response.Success(c, nil)
}

// HandleTestTelegram 测试 Telegram 通知
func (h *ConfigHandler) HandleTestTelegram(c *gin.Context) {
	if h.telegramService == nil {
		response.InternalError(c, "Telegram 服务未初始化")
		return
	}

	if err := h.telegramService.SendTestMessage(c.Request.Context()); err != nil {
		response.InternalError(c, "Telegram 测试消息发送失败: "+err.Error())
		return
	}

	response.Success(c, nil)
}

// ============ SSO 配置 ============

// HandleGetSSOConfig 获取 SSO 配置
func (h *ConfigHandler) HandleGetSSOConfig(c *gin.Context) {
	config, err := h.configService.GetSSOConfig(c.Request.Context())
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}

	response.Success(c, config)
}

// HandleUpdateSSOConfig 更新 SSO 配置
func (h *ConfigHandler) HandleUpdateSSOConfig(c *gin.Context) {
	var req dto.UpdateSSOConfigRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数错误: "+err.Error())
		return
	}

	userID := c.GetUint64("user_id")
	if err := h.configService.UpdateSSOConfig(c.Request.Context(), &req, userID); err != nil {
		response.InternalError(c, err.Error())
		return
	}

	response.Success(c, nil)
}

// ============ 品牌配置 ============

// HandleGetBrandConfig 获取品牌配置（公开接口）
func (h *ConfigHandler) HandleGetBrandConfig(c *gin.Context) {
	config, err := h.configService.GetBrandConfig(c.Request.Context())
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}

	response.Success(c, config)
}

// HandleUpdateBrandConfig 更新品牌配置
func (h *ConfigHandler) HandleUpdateBrandConfig(c *gin.Context) {
	var req dto.UpdateBrandConfigRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数错误: "+err.Error())
		return
	}

	userID := c.GetUint64("user_id")
	if err := h.configService.UpdateBrandConfig(c.Request.Context(), &req, userID); err != nil {
		response.InternalError(c, err.Error())
		return
	}

	response.Success(c, nil)
}

// brandAssetAllowedExts 品牌资源允许的文件扩展名
var brandAssetAllowedExts = map[string]bool{
	".svg":  true,
	".png":  true,
	".ico":  true,
	".jpg":  true,
	".jpeg": true,
	".webp": true,
}

// HandleUploadBrandAsset 上传品牌资源（Logo/Favicon）
func (h *ConfigHandler) HandleUploadBrandAsset(c *gin.Context) {
	if h.localStorage == nil {
		response.InternalError(c, "文件存储服务未初始化")
		return
	}

	assetType := c.PostForm("type")
	if assetType != "logo" && assetType != "favicon" {
		response.BadRequest(c, "类型参数错误，仅支持 logo 或 favicon")
		return
	}

	file, err := c.FormFile("file")
	if err != nil {
		response.BadRequest(c, "请选择文件上传")
		return
	}

	// 检查文件大小（2MB）
	if file.Size > 2*1024*1024 {
		response.BadRequest(c, "文件大小不能超过 2MB")
		return
	}

	// 检查文件扩展名
	ext := filepath.Ext(file.Filename)
	if !brandAssetAllowedExts[ext] {
		response.BadRequest(c, "不支持的文件格式，仅支持 SVG、PNG、ICO、JPG、WEBP")
		return
	}

	// 生成唯一文件名
	filename := fmt.Sprintf("brand/%s_%d%s", assetType, time.Now().UnixMilli(), ext)

	src, err := file.Open()
	if err != nil {
		response.InternalError(c, "打开文件失败")
		return
	}
	defer src.Close()

	relPath, err := h.localStorage.SaveTo(src, filename)
	if err != nil {
		response.InternalError(c, "保存文件失败: "+err.Error())
		return
	}

	// 构建访问 URL
	assetURL := "/api/v1/brand/assets/" + relPath

	// 更新对应的配置键
	userID := c.GetUint64("user_id")
	configKey := service.KeyBrandLogoURL
	if assetType == "favicon" {
		configKey = service.KeyBrandFaviconURL
	}
	if err := h.configService.UpdateConfig(c.Request.Context(), configKey, assetURL, userID); err != nil {
		response.InternalError(c, "更新配置失败: "+err.Error())
		return
	}

	response.Success(c, gin.H{"url": assetURL})
}
