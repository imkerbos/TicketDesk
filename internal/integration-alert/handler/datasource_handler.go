// Package handler 提供告警数据源 HTTP 处理器
package handler

import (
	"encoding/json"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/kerbos/ticketdesk/internal/api/response"
	"github.com/kerbos/ticketdesk/internal/integration-alert/dto"
	"github.com/kerbos/ticketdesk/internal/integration-alert/service"
	"github.com/kerbos/ticketdesk/internal/model"
	"github.com/kerbos/ticketdesk/pkg/logger"
	"go.uber.org/zap"
)

// DatasourceHandler 数据源处理器
type DatasourceHandler struct {
	datasourceService service.DatasourceService
	alertService      service.AlertService
}

// NewDatasourceHandler 创建数据源处理器
func NewDatasourceHandler(datasourceService service.DatasourceService, alertService service.AlertService) *DatasourceHandler {
	return &DatasourceHandler{
		datasourceService: datasourceService,
		alertService:      alertService,
	}
}

// HandleListDatasources 获取数据源列表
func (h *DatasourceHandler) HandleListDatasources(c *gin.Context) {
	var req dto.DatasourceListRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}

	result, err := h.datasourceService.List(c.Request.Context(), &req)
	if err != nil {
		logger.Error("failed to list datasources", zap.Error(err))
		response.InternalError(c, "获取数据源列表失败")
		return
	}

	response.Success(c, result)
}

// HandleGetDatasource 获取数据源详情
func (h *DatasourceHandler) HandleGetDatasource(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的数据源 ID")
		return
	}

	ds, err := h.datasourceService.GetByID(c.Request.Context(), id)
	if err != nil {
		logger.Error("failed to get datasource", zap.Uint64("id", id), zap.Error(err))
		response.NotFound(c, "数据源不存在")
		return
	}

	response.Success(c, ds)
}

// HandleCreateDatasource 创建数据源
func (h *DatasourceHandler) HandleCreateDatasource(c *gin.Context) {
	var req dto.CreateDatasourceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}

	userID, exists := c.Get("user_id")
	if !exists {
		response.Unauthorized(c, "未认证")
		return
	}

	ds, err := h.datasourceService.Create(c.Request.Context(), &req, userID.(uint64))
	if err != nil {
		logger.Error("failed to create datasource", zap.Error(err))
		response.InternalError(c, "创建数据源失败: "+err.Error())
		return
	}

	response.Created(c, ds)
}

// HandleUpdateDatasource 更新数据源
func (h *DatasourceHandler) HandleUpdateDatasource(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的数据源 ID")
		return
	}

	var req dto.UpdateDatasourceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}

	ds, err := h.datasourceService.Update(c.Request.Context(), id, &req)
	if err != nil {
		logger.Error("failed to update datasource", zap.Uint64("id", id), zap.Error(err))
		response.InternalError(c, "更新数据源失败: "+err.Error())
		return
	}

	response.Success(c, ds)
}

// HandleDeleteDatasource 删除数据源
func (h *DatasourceHandler) HandleDeleteDatasource(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的数据源 ID")
		return
	}

	if err := h.datasourceService.Delete(c.Request.Context(), id); err != nil {
		logger.Error("failed to delete datasource", zap.Uint64("id", id), zap.Error(err))
		response.InternalError(c, "删除数据源失败")
		return
	}

	response.Success(c, gin.H{"message": "datasource deleted"})
}

// HandleTestConnection 测试数据源连接（创建前）
func (h *DatasourceHandler) HandleTestConnection(c *gin.Context) {
	var req dto.TestDatasourceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}

	result, err := h.datasourceService.TestConnection(c.Request.Context(), &req)
	if err != nil {
		logger.Error("failed to test connection", zap.Error(err))
		response.InternalError(c, "测试连接失败")
		return
	}

	response.Success(c, result)
}

// HandleTestConnectionByID 测试已保存数据源的连接
func (h *DatasourceHandler) HandleTestConnectionByID(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的数据源 ID")
		return
	}

	result, err := h.datasourceService.TestConnectionByID(c.Request.Context(), id)
	if err != nil {
		logger.Error("failed to test connection by id", zap.Uint64("id", id), zap.Error(err))
		response.InternalError(c, "测试连接失败")
		return
	}

	response.Success(c, result)
}

// HandleDatasourceWebhook 通用数据源 Webhook 入口
func (h *DatasourceHandler) HandleDatasourceWebhook(c *gin.Context) {
	name := c.Param("name")

	ds, err := h.datasourceService.GetByName(c.Request.Context(), name)
	if err != nil {
		logger.Error("datasource not found for webhook", zap.String("name", name), zap.Error(err))
		response.NotFound(c, "数据源不存在: "+name)
		return
	}

	if ds.Status == 0 {
		response.BadRequest(c, "数据源已禁用")
		return
	}

	switch ds.Type {
	case model.DatasourceTypePrometheus:
		var req dto.AlertWebhookRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			response.BadRequest(c, "参数错误: "+err.Error())
			return
		}
		if err := h.alertService.HandleWebhookWithSource(c.Request.Context(), &req, ds.Name); err != nil {
			logger.Error("failed to handle prometheus webhook", zap.String("source", ds.Name), zap.Error(err))
			response.InternalError(c, "处理 Webhook 失败")
			return
		}

	case model.DatasourceTypeNightingale:
		body, err := c.GetRawData()
		if err != nil {
			response.BadRequest(c, "读取请求体失败: "+err.Error())
			return
		}
		var events []dto.N9eAlertEvent
		if err := json.Unmarshal(body, &events); err != nil {
			var single dto.N9eAlertEvent
			if err := json.Unmarshal(body, &single); err != nil {
				response.BadRequest(c, "参数错误: "+err.Error())
				return
			}
			events = []dto.N9eAlertEvent{single}
		}
		if err := h.alertService.HandleNightingaleWebhookWithSource(c.Request.Context(), events, ds.Name); err != nil {
			logger.Error("failed to handle nightingale webhook", zap.String("source", ds.Name), zap.Error(err))
			response.InternalError(c, "处理夜莺 Webhook 失败")
			return
		}

	default:
		response.BadRequest(c, "不支持的数据源类型: "+ds.Type)
		return
	}

	response.Success(c, gin.H{"message": "webhook processed", "source": ds.Name})
}
