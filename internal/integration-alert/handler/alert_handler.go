// Package handler 提供告警 HTTP 处理器
package handler

import (
	"encoding/json"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/kerbos/ticketdesk/internal/api/response"
	"github.com/kerbos/ticketdesk/internal/integration-alert/dto"
	"github.com/kerbos/ticketdesk/internal/integration-alert/service"
	"github.com/kerbos/ticketdesk/pkg/logger"
	"go.uber.org/zap"
)

// AlertHandler 告警处理器
type AlertHandler struct {
	alertService service.AlertService
}

// NewAlertHandler 创建告警处理器
func NewAlertHandler(alertService service.AlertService) *AlertHandler {
	return &AlertHandler{
		alertService: alertService,
	}
}

// HandleWebhook 处理告警 Webhook
func (h *AlertHandler) HandleWebhook(c *gin.Context) {
	var req dto.AlertWebhookRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}

	if err := h.alertService.HandleWebhook(c.Request.Context(), &req); err != nil {
		logger.Error("failed to handle webhook", zap.Error(err))
		response.InternalError(c, "处理 Webhook 失败")
		return
	}

	response.Success(c, gin.H{"message": "webhook processed"})
}

// HandleNightingaleWebhook 处理夜莺告警 Webhook
func (h *AlertHandler) HandleNightingaleWebhook(c *gin.Context) {
	body, err := c.GetRawData()
	if err != nil {
		response.BadRequest(c, "读取请求体失败: "+err.Error())
		return
	}

	// 先尝试解析为数组
	var events []dto.N9eAlertEvent
	if err := json.Unmarshal(body, &events); err != nil {
		// 失败则尝试解析为单个对象
		var single dto.N9eAlertEvent
		if err := json.Unmarshal(body, &single); err != nil {
			response.BadRequest(c, "参数错误: "+err.Error())
			return
		}
		events = []dto.N9eAlertEvent{single}
	}

	if err := h.alertService.HandleNightingaleWebhook(c.Request.Context(), events); err != nil {
		logger.Error("failed to handle nightingale webhook", zap.Error(err))
		response.InternalError(c, "处理夜莺 Webhook 失败")
		return
	}

	response.Success(c, gin.H{"message": "nightingale webhook processed"})
}

// HandleListAlerts 获取告警列表
func (h *AlertHandler) HandleListAlerts(c *gin.Context) {
	var req dto.AlertListRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}

	result, err := h.alertService.ListAlerts(c.Request.Context(), &req)
	if err != nil {
		logger.Error("failed to list alerts", zap.Error(err))
		response.InternalError(c, "获取告警列表失败")
		return
	}

	response.Success(c, result)
}

// HandleGetAlertStats 获取告警统计数据
func (h *AlertHandler) HandleGetAlertStats(c *gin.Context) {
	stats, err := h.alertService.GetAlertStats(c.Request.Context())
	if err != nil {
		logger.Error("failed to get alert stats", zap.Error(err))
		response.InternalError(c, "获取告警统计失败")
		return
	}

	response.Success(c, stats)
}

// HandleGetAlert 获取告警详情
func (h *AlertHandler) HandleGetAlert(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的告警 ID")
		return
	}

	alert, err := h.alertService.GetAlert(c.Request.Context(), id)
	if err != nil {
		logger.Error("failed to get alert", zap.Uint64("id", id), zap.Error(err))
		response.NotFound(c, "告警不存在")
		return
	}

	response.Success(c, alert)
}

// HandleAckAlert 确认告警
func (h *AlertHandler) HandleAckAlert(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的告警 ID")
		return
	}

	var req dto.AlertAckRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}

	// 获取当前用户 ID
	userID, exists := c.Get("user_id")
	if !exists {
		response.Unauthorized(c, "未认证")
		return
	}

	if err := h.alertService.AckAlert(c.Request.Context(), id, userID.(uint64), &req); err != nil {
		logger.Error("failed to ack alert", zap.Uint64("id", id), zap.Error(err))
		response.InternalError(c, "确认告警失败: "+err.Error())
		return
	}

	response.Success(c, gin.H{"message": "alert acknowledged"})
}

// HandleResolveAlert 解决告警
func (h *AlertHandler) HandleResolveAlert(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的告警 ID")
		return
	}

	var req dto.AlertResolveRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}

	// 获取当前用户 ID
	userID, exists := c.Get("user_id")
	if !exists {
		response.Unauthorized(c, "未认证")
		return
	}

	if err := h.alertService.ResolveAlert(c.Request.Context(), id, userID.(uint64), &req); err != nil {
		logger.Error("failed to resolve alert", zap.Uint64("id", id), zap.Error(err))
		response.InternalError(c, "解决告警失败: "+err.Error())
		return
	}

	response.Success(c, gin.H{"message": "alert resolved"})
}

// ============ 告警规则管理 ============

// HandleCreateAlertRule 创建告警规则
func (h *AlertHandler) HandleCreateAlertRule(c *gin.Context) {
	var req dto.CreateAlertRuleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}

	rule, err := h.alertService.CreateAlertRule(c.Request.Context(), &req)
	if err != nil {
		logger.Error("failed to create alert rule", zap.Error(err))
		response.InternalError(c, "创建告警规则失败")
		return
	}

	response.Created(c, rule)
}

// HandleGetAlertRule 获取告警规则详情
func (h *AlertHandler) HandleGetAlertRule(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的规则 ID")
		return
	}

	rule, err := h.alertService.GetAlertRule(c.Request.Context(), id)
	if err != nil {
		logger.Error("failed to get alert rule", zap.Uint64("id", id), zap.Error(err))
		response.NotFound(c, "告警规则不存在")
		return
	}

	response.Success(c, rule)
}

// HandleUpdateAlertRule 更新告警规则
func (h *AlertHandler) HandleUpdateAlertRule(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的规则 ID")
		return
	}

	var req dto.UpdateAlertRuleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}

	rule, err := h.alertService.UpdateAlertRule(c.Request.Context(), id, &req)
	if err != nil {
		logger.Error("failed to update alert rule", zap.Uint64("id", id), zap.Error(err))
		response.InternalError(c, "更新告警规则失败")
		return
	}

	response.Success(c, rule)
}

// HandleDeleteAlertRule 删除告警规则
func (h *AlertHandler) HandleDeleteAlertRule(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的规则 ID")
		return
	}

	if err := h.alertService.DeleteAlertRule(c.Request.Context(), id); err != nil {
		logger.Error("failed to delete alert rule", zap.Uint64("id", id), zap.Error(err))
		response.InternalError(c, "删除告警规则失败")
		return
	}

	response.Success(c, gin.H{"message": "alert rule deleted"})
}

// HandleListAlertRules 获取告警规则列表
func (h *AlertHandler) HandleListAlertRules(c *gin.Context) {
	var req dto.AlertRuleListRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}

	result, err := h.alertService.ListAlertRules(c.Request.Context(), &req)
	if err != nil {
		logger.Error("failed to list alert rules", zap.Error(err))
		response.InternalError(c, "获取告警规则列表失败")
		return
	}

	response.Success(c, result)
}

// HandleGetAlertLabelKeys 获取所有告警标签 key 列表
func (h *AlertHandler) HandleGetAlertLabelKeys(c *gin.Context) {
	keys, err := h.alertService.GetAlertLabelKeys(c.Request.Context())
	if err != nil {
		logger.Error("failed to get alert label keys", zap.Error(err))
		response.InternalError(c, "获取标签列表失败")
		return
	}

	response.Success(c, keys)
}

// HandleGroupAlerts 按标签分组统计告警
func (h *AlertHandler) HandleGroupAlerts(c *gin.Context) {
	var req dto.AlertGroupRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}

	result, err := h.alertService.GroupAlerts(c.Request.Context(), &req)
	if err != nil {
		logger.Error("failed to group alerts", zap.Error(err))
		response.InternalError(c, "分组统计告警失败")
		return
	}

	response.Success(c, result)
}

// ============ 告警静默管理 ============

// HandleCreateAlertSilence 创建告警静默
func (h *AlertHandler) HandleCreateAlertSilence(c *gin.Context) {
	var req dto.CreateAlertSilenceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}

	// 获取当前用户 ID
	userID, exists := c.Get("user_id")
	if !exists {
		response.Unauthorized(c, "未认证")
		return
	}

	silence, err := h.alertService.CreateAlertSilence(c.Request.Context(), &req, userID.(uint64))
	if err != nil {
		logger.Error("failed to create alert silence", zap.Error(err))
		response.InternalError(c, "创建告警静默失败")
		return
	}

	response.Created(c, silence)
}

// HandleGetAlertSilence 获取告警静默详情
func (h *AlertHandler) HandleGetAlertSilence(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的静默 ID")
		return
	}

	silence, err := h.alertService.GetAlertSilence(c.Request.Context(), id)
	if err != nil {
		logger.Error("failed to get alert silence", zap.Uint64("id", id), zap.Error(err))
		response.NotFound(c, "告警静默不存在")
		return
	}

	response.Success(c, silence)
}

// HandleUpdateAlertSilence 更新告警静默
func (h *AlertHandler) HandleUpdateAlertSilence(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的静默 ID")
		return
	}

	var req dto.UpdateAlertSilenceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}

	silence, err := h.alertService.UpdateAlertSilence(c.Request.Context(), id, &req)
	if err != nil {
		logger.Error("failed to update alert silence", zap.Uint64("id", id), zap.Error(err))
		response.InternalError(c, "更新告警静默失败")
		return
	}

	response.Success(c, silence)
}

// HandleDeleteAlertSilence 删除告警静默
func (h *AlertHandler) HandleDeleteAlertSilence(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的静默 ID")
		return
	}

	if err := h.alertService.DeleteAlertSilence(c.Request.Context(), id); err != nil {
		logger.Error("failed to delete alert silence", zap.Uint64("id", id), zap.Error(err))
		response.InternalError(c, "删除告警静默失败")
		return
	}

	response.Success(c, gin.H{"message": "alert silence deleted"})
}

// HandleCancelAlertSilence 取消告警静默
func (h *AlertHandler) HandleCancelAlertSilence(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的静默 ID")
		return
	}

	if err := h.alertService.CancelAlertSilence(c.Request.Context(), id); err != nil {
		logger.Error("failed to cancel alert silence", zap.Uint64("id", id), zap.Error(err))
		response.InternalError(c, "取消告警静默失败")
		return
	}

	response.Success(c, gin.H{"message": "alert silence cancelled"})
}

// HandleListAlertSilences 获取告警静默列表
func (h *AlertHandler) HandleListAlertSilences(c *gin.Context) {
	var req dto.AlertSilenceListRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}

	result, err := h.alertService.ListAlertSilences(c.Request.Context(), &req)
	if err != nil {
		logger.Error("failed to list alert silences", zap.Error(err))
		response.InternalError(c, "获取告警静默列表失败")
		return
	}

	response.Success(c, result)
}
