// Package handler 提供用户模块 HTTP 处理层
package handler

import (
	"errors"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/kerbos/ticketdesk/internal/api/response"
	"github.com/kerbos/ticketdesk/internal/core-user/dto"
	"github.com/kerbos/ticketdesk/internal/core-user/service"
)

// APITokenHandler API token HTTP 处理器
type APITokenHandler struct {
	tokenSvc service.APITokenService
}

// NewAPITokenHandler 创建 API token 处理器实例
func NewAPITokenHandler(tokenSvc service.APITokenService) *APITokenHandler {
	return &APITokenHandler{tokenSvc: tokenSvc}
}

// HandleCreate 创建 API token
// 注意：不允许用 PAT 创建新 PAT，避免权限提升
func (h *APITokenHandler) HandleCreate(c *gin.Context) {
	// 拒绝用 PAT 自己创建 PAT
	if isPAT, _ := c.Get("is_pat"); isPAT == true {
		response.Forbidden(c, "不能用 API token 创建新 token，请用账号登录")
		return
	}

	var req dto.CreateTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数错误: "+err.Error())
		return
	}

	userID := c.GetUint64("user_id")
	result, err := h.tokenSvc.Create(c.Request.Context(), userID, &req)
	if err != nil {
		if errors.Is(err, service.ErrTokenLimitExceeded) {
			response.BadRequest(c, err.Error())
			return
		}
		response.InternalError(c, "创建 token 失败")
		return
	}
	response.Created(c, result)
}

// HandleList 列出当前用户所有 token
func (h *APITokenHandler) HandleList(c *gin.Context) {
	userID := c.GetUint64("user_id")
	tokens, err := h.tokenSvc.List(c.Request.Context(), userID)
	if err != nil {
		response.InternalError(c, "查询 token 列表失败")
		return
	}
	response.Success(c, tokens)
}

// HandleDelete 撤销指定 token
func (h *APITokenHandler) HandleDelete(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的 token ID")
		return
	}
	userID := c.GetUint64("user_id")
	if err := h.tokenSvc.Delete(c.Request.Context(), userID, id); err != nil {
		if errors.Is(err, service.ErrTokenNotFound) {
			response.NotFound(c, err.Error())
			return
		}
		response.InternalError(c, "撤销 token 失败")
		return
	}
	response.Success(c, nil)
}
