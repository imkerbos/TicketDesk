// Package handler 提供用户模块的 HTTP 处理器
package handler

import (
	"errors"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/kerbos/ticketdesk/internal/api/response"
	"github.com/kerbos/ticketdesk/internal/core-user/dto"
	"github.com/kerbos/ticketdesk/internal/core-user/service"
)

// UserHandler 用户处理器
type UserHandler struct {
	userService service.UserService
	mfaService  service.MFAService
}

// NewUserHandler 创建用户处理器实例
func NewUserHandler(userService service.UserService, mfaService service.MFAService) *UserHandler {
	return &UserHandler{
		userService: userService,
		mfaService:  mfaService,
	}
}

// HandleLogin 处理登录请求
// @Summary 用户登录
// @Description 使用用户名和密码登录
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body dto.LoginRequest true "登录请求"
// @Success 200 {object} dto.LoginResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Router /api/v1/auth/login [post]
func (h *UserHandler) HandleLogin(c *gin.Context) {
	var req dto.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数错误: "+err.Error())
		return
	}

	result, err := h.userService.Login(c.Request.Context(), &req)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrInvalidCredentials):
			response.Unauthorized(c, err.Error())
		case errors.Is(err, service.ErrUserDisabled):
			response.Forbidden(c, err.Error())
		default:
			response.InternalError(c, "登录失败")
		}
		return
	}

	response.Success(c, result)
}

// HandleRegister 处理注册请求
// @Summary 用户注册
// @Description 注册新用户
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body dto.RegisterRequest true "注册请求"
// @Success 201 {object} dto.UserResponse
// @Failure 400 {object} response.ErrorResponse
// @Router /api/v1/auth/register [post]
func (h *UserHandler) HandleRegister(c *gin.Context) {
	var req dto.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数错误: "+err.Error())
		return
	}

	result, err := h.userService.Register(c.Request.Context(), &req)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrUsernameExists):
			response.BadRequest(c, err.Error())
		case errors.Is(err, service.ErrEmailExists):
			response.BadRequest(c, err.Error())
		default:
			response.InternalError(c, "注册失败")
		}
		return
	}

	response.Created(c, result)
}

// HandleRefreshToken 处理刷新 Token 请求
// @Summary 刷新 Token
// @Description 使用 Refresh Token 获取新的 Access Token
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body dto.RefreshTokenRequest true "刷新 Token 请求"
// @Success 200 {object} dto.LoginResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Router /api/v1/auth/refresh [post]
func (h *UserHandler) HandleRefreshToken(c *gin.Context) {
	var req dto.RefreshTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数错误: "+err.Error())
		return
	}

	result, err := h.userService.RefreshToken(c.Request.Context(), req.RefreshToken)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrUserNotFound):
			response.Unauthorized(c, "用户不存在")
		case errors.Is(err, service.ErrUserDisabled):
			response.Forbidden(c, err.Error())
		default:
			response.Unauthorized(c, "Token 无效或已过期")
		}
		return
	}

	response.Success(c, result)
}

// HandleGetCurrentUser 获取当前用户信息
// @Summary 获取当前用户信息
// @Description 获取当前登录用户的详细信息
// @Tags User
// @Produce json
// @Success 200 {object} dto.UserResponse
// @Failure 401 {object} response.ErrorResponse
// @Router /api/v1/users/me [get]
// @Security BearerAuth
func (h *UserHandler) HandleGetCurrentUser(c *gin.Context) {
	userID := c.GetUint64("user_id")
	if userID == 0 {
		response.Unauthorized(c, "未获取到用户信息")
		return
	}

	result, err := h.userService.GetCurrentUser(c.Request.Context(), userID)
	if err != nil {
		if errors.Is(err, service.ErrUserNotFound) {
			response.NotFound(c, err.Error())
			return
		}
		response.InternalError(c, "获取用户信息失败")
		return
	}

	response.Success(c, result)
}

// HandleGetUser 获取用户详情
// @Summary 获取用户详情
// @Description 根据用户 ID 获取用户详细信息
// @Tags User
// @Produce json
// @Param id path int true "用户 ID"
// @Success 200 {object} dto.UserResponse
// @Failure 404 {object} response.ErrorResponse
// @Router /api/v1/users/{id} [get]
// @Security BearerAuth
func (h *UserHandler) HandleGetUser(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的用户 ID")
		return
	}

	result, err := h.userService.GetUser(c.Request.Context(), id)
	if err != nil {
		if errors.Is(err, service.ErrUserNotFound) {
			response.NotFound(c, err.Error())
			return
		}
		response.InternalError(c, "获取用户信息失败")
		return
	}

	response.Success(c, result)
}

// HandleCreateUser 创建用户（管理员）
// @Summary 创建用户
// @Description 管理员创建新用户
// @Tags User
// @Accept json
// @Produce json
// @Param request body dto.CreateUserRequest true "创建用户请求"
// @Success 201 {object} dto.UserResponse
// @Failure 400 {object} response.ErrorResponse
// @Router /api/v1/users [post]
// @Security BearerAuth
func (h *UserHandler) HandleCreateUser(c *gin.Context) {
	var req dto.CreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数错误: "+err.Error())
		return
	}

	result, err := h.userService.CreateUser(c.Request.Context(), &req)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrUsernameExists):
			response.BadRequest(c, err.Error())
		case errors.Is(err, service.ErrEmailExists):
			response.BadRequest(c, err.Error())
		default:
			response.InternalError(c, "创建用户失败")
		}
		return
	}

	response.Created(c, result)
}

// HandleUpdateUser 更新用户信息
// @Summary 更新用户信息
// @Description 更新指定用户的信息
// @Tags User
// @Accept json
// @Produce json
// @Param id path int true "用户 ID"
// @Param request body dto.UpdateUserRequest true "更新用户请求"
// @Success 200 {object} dto.UserResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Router /api/v1/users/{id} [put]
// @Security BearerAuth
func (h *UserHandler) HandleUpdateUser(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的用户 ID")
		return
	}

	var req dto.UpdateUserRequest
	if bindErr := c.ShouldBindJSON(&req); bindErr != nil {
		response.BadRequest(c, "请求参数错误: "+bindErr.Error())
		return
	}

	result, err := h.userService.UpdateUser(c.Request.Context(), id, &req)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrUserNotFound):
			response.NotFound(c, err.Error())
		case errors.Is(err, service.ErrEmailExists):
			response.BadRequest(c, err.Error())
		default:
			response.InternalError(c, "更新用户失败")
		}
		return
	}

	response.Success(c, result)
}

// HandleUpdatePassword 修改密码
// @Summary 修改密码
// @Description 修改当前用户的密码
// @Tags User
// @Accept json
// @Produce json
// @Param request body dto.UpdatePasswordRequest true "修改密码请求"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.ErrorResponse
// @Router /api/v1/users/me/password [put]
// @Security BearerAuth
func (h *UserHandler) HandleUpdatePassword(c *gin.Context) {
	userID := c.GetUint64("user_id")
	if userID == 0 {
		response.Unauthorized(c, "未获取到用户信息")
		return
	}

	var req dto.UpdatePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数错误: "+err.Error())
		return
	}

	err := h.userService.UpdatePassword(c.Request.Context(), userID, &req)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrInvalidOldPassword):
			response.BadRequest(c, err.Error())
		case errors.Is(err, service.ErrUserNotFound):
			response.NotFound(c, err.Error())
		default:
			response.InternalError(c, "修改密码失败")
		}
		return
	}

	response.Success(c, gin.H{"message": "密码修改成功"})
}

// HandleResetPassword 重置用户密码（管理员）
// @Summary 重置用户密码
// @Description 管理员重置指定用户的密码
// @Tags User
// @Accept json
// @Produce json
// @Param id path int true "用户 ID"
// @Param request body dto.ResetPasswordRequest true "重置密码请求"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Router /api/v1/users/{id}/reset-password [post]
// @Security BearerAuth
func (h *UserHandler) HandleResetPassword(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的用户 ID")
		return
	}

	var req dto.ResetPasswordRequest
	if bindErr := c.ShouldBindJSON(&req); bindErr != nil {
		response.BadRequest(c, "请求参数错误: "+bindErr.Error())
		return
	}

	err = h.userService.ResetPassword(c.Request.Context(), id, &req)
	if err != nil {
		if errors.Is(err, service.ErrUserNotFound) {
			response.NotFound(c, err.Error())
			return
		}
		response.InternalError(c, "重置密码失败")
		return
	}

	response.Success(c, gin.H{"message": "密码重置成功"})
}

// HandleEnableUser 启用用户
// @Summary 启用用户
// @Description 启用指定用户
// @Tags User
// @Produce json
// @Param id path int true "用户 ID"
// @Success 200 {object} response.Response
// @Failure 404 {object} response.ErrorResponse
// @Router /api/v1/users/{id}/enable [post]
// @Security BearerAuth
func (h *UserHandler) HandleEnableUser(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的用户 ID")
		return
	}

	err = h.userService.EnableUser(c.Request.Context(), id)
	if err != nil {
		if errors.Is(err, service.ErrUserNotFound) {
			response.NotFound(c, err.Error())
			return
		}
		response.InternalError(c, "启用用户失败")
		return
	}

	response.Success(c, gin.H{"message": "用户已启用"})
}

// HandleDisableUser 禁用用户
// @Summary 禁用用户
// @Description 禁用指定用户
// @Tags User
// @Produce json
// @Param id path int true "用户 ID"
// @Success 200 {object} response.Response
// @Failure 404 {object} response.ErrorResponse
// @Router /api/v1/users/{id}/disable [post]
// @Security BearerAuth
func (h *UserHandler) HandleDisableUser(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的用户 ID")
		return
	}

	err = h.userService.DisableUser(c.Request.Context(), id)
	if err != nil {
		if errors.Is(err, service.ErrUserNotFound) {
			response.NotFound(c, err.Error())
			return
		}
		response.InternalError(c, "禁用用户失败")
		return
	}

	response.Success(c, gin.H{"message": "用户已禁用"})
}

// HandleDeleteUser 删除用户
// @Summary 删除用户
// @Description 删除指定用户（软删除）
// @Tags User
// @Produce json
// @Param id path int true "用户 ID"
// @Success 200 {object} response.Response
// @Failure 404 {object} response.ErrorResponse
// @Router /api/v1/users/{id} [delete]
// @Security BearerAuth
func (h *UserHandler) HandleDeleteUser(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的用户 ID")
		return
	}

	err = h.userService.DeleteUser(c.Request.Context(), id)
	if err != nil {
		if errors.Is(err, service.ErrUserNotFound) {
			response.NotFound(c, err.Error())
			return
		}
		response.InternalError(c, "删除用户失败")
		return
	}

	response.Success(c, gin.H{"message": "用户删除成功"})
}

// HandleListUsers 获取用户列表
// @Summary 获取用户列表
// @Description 分页获取用户列表
// @Tags User
// @Produce json
// @Param page query int false "页码" default(1)
// @Param page_size query int false "每页数量" default(20)
// @Param keyword query string false "搜索关键字"
// @Param status query int false "用户状态 (0-禁用, 1-启用)"
// @Success 200 {object} response.PageData
// @Router /api/v1/users [get]
// @Security BearerAuth
func (h *UserHandler) HandleListUsers(c *gin.Context) {
	var req dto.ListUsersRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.BadRequest(c, "请求参数错误: "+err.Error())
		return
	}

	users, total, err := h.userService.ListUsers(c.Request.Context(), &req)
	if err != nil {
		response.InternalError(c, "获取用户列表失败")
		return
	}

	response.SuccessWithPage(c, users, total, req.GetDefaultPage(), req.GetDefaultPageSize())
}

// HandleListAllUsers 获取所有用户（用于选择器）
// @Summary 获取所有用户
// @Description 获取所有启用的用户列表（不分页，用于选择器）
// @Tags User
// @Produce json
// @Success 200 {array} dto.UserBrief
// @Router /api/v1/users/all [get]
// @Security BearerAuth
func (h *UserHandler) HandleListAllUsers(c *gin.Context) {
	// 获取所有启用的用户
	statusEnabled := int8(1)
	req := &dto.ListUsersRequest{
		Page:     1,
		PageSize: 1000,           // 获取足够多的用户
		Status:   &statusEnabled, // 只获取启用的用户
	}

	users, _, err := h.userService.ListUsers(c.Request.Context(), req)
	if err != nil {
		response.InternalError(c, "获取用户列表失败")
		return
	}

	// 转换为简要信息
	briefs := make([]map[string]interface{}, len(users))
	for i, u := range users {
		briefs[i] = map[string]interface{}{
			"id":           u.ID,
			"username":     u.Username,
			"display_name": u.DisplayName,
		}
	}

	response.Success(c, briefs)
}

// ============ MFA 相关处理器 ============

// HandleGetMFAStatus 获取 MFA 状态
// @Summary 获取 MFA 状态
// @Description 获取当前用户的 MFA 状态
// @Tags MFA
// @Produce json
// @Success 200 {object} dto.MFAStatusResponse
// @Router /api/v1/users/me/mfa [get]
// @Security BearerAuth
func (h *UserHandler) HandleGetMFAStatus(c *gin.Context) {
	userID := c.GetUint64("user_id")
	if userID == 0 {
		response.Unauthorized(c, "未获取到用户信息")
		return
	}

	status, err := h.mfaService.GetMFAStatus(c.Request.Context(), userID)
	if err != nil {
		if errors.Is(err, service.ErrUserNotFound) {
			response.NotFound(c, err.Error())
			return
		}
		response.InternalError(c, "获取 MFA 状态失败")
		return
	}

	response.Success(c, status)
}

// HandleSetupMFA 开始 MFA 设置
// @Summary 开始 MFA 设置
// @Description 生成 MFA 密钥和二维码 URL
// @Tags MFA
// @Produce json
// @Success 200 {object} dto.MFASetupResponse
// @Failure 400 {object} response.ErrorResponse
// @Router /api/v1/users/me/mfa/setup [post]
// @Security BearerAuth
func (h *UserHandler) HandleSetupMFA(c *gin.Context) {
	userID := c.GetUint64("user_id")
	if userID == 0 {
		response.Unauthorized(c, "未获取到用户信息")
		return
	}

	result, err := h.mfaService.SetupMFA(c.Request.Context(), userID)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrMFAAlreadyEnabled):
			response.BadRequest(c, err.Error())
		case errors.Is(err, service.ErrUserNotFound):
			response.NotFound(c, err.Error())
		default:
			response.InternalError(c, "设置 MFA 失败")
		}
		return
	}

	response.Success(c, result)
}

// HandleEnableMFA 验证并启用 MFA
// @Summary 验证并启用 MFA
// @Description 验证 TOTP 码并启用 MFA
// @Tags MFA
// @Accept json
// @Produce json
// @Param request body dto.MFAVerifyRequest true "MFA 验证请求"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.ErrorResponse
// @Router /api/v1/users/me/mfa/enable [post]
// @Security BearerAuth
func (h *UserHandler) HandleEnableMFA(c *gin.Context) {
	userID := c.GetUint64("user_id")
	if userID == 0 {
		response.Unauthorized(c, "未获取到用户信息")
		return
	}

	var req dto.MFAVerifyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数错误: "+err.Error())
		return
	}

	err := h.mfaService.VerifyAndEnableMFA(c.Request.Context(), userID, req.Code)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrMFAAlreadyEnabled):
			response.BadRequest(c, err.Error())
		case errors.Is(err, service.ErrMFASetupNotStarted):
			response.BadRequest(c, err.Error())
		case errors.Is(err, service.ErrMFAInvalidCode):
			response.BadRequest(c, err.Error())
		case errors.Is(err, service.ErrUserNotFound):
			response.NotFound(c, err.Error())
		default:
			response.InternalError(c, "启用 MFA 失败")
		}
		return
	}

	response.Success(c, gin.H{"message": "MFA 已启用"})
}

// HandleDisableMFA 禁用 MFA
// @Summary 禁用 MFA
// @Description 验证 TOTP 码并禁用 MFA
// @Tags MFA
// @Accept json
// @Produce json
// @Param request body dto.MFAVerifyRequest true "MFA 验证请求"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.ErrorResponse
// @Router /api/v1/users/me/mfa/disable [post]
// @Security BearerAuth
func (h *UserHandler) HandleDisableMFA(c *gin.Context) {
	userID := c.GetUint64("user_id")
	if userID == 0 {
		response.Unauthorized(c, "未获取到用户信息")
		return
	}

	var req dto.MFAVerifyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数错误: "+err.Error())
		return
	}

	err := h.mfaService.DisableMFA(c.Request.Context(), userID, req.Code)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrMFANotEnabled):
			response.BadRequest(c, err.Error())
		case errors.Is(err, service.ErrMFAInvalidCode):
			response.BadRequest(c, err.Error())
		case errors.Is(err, service.ErrUserNotFound):
			response.NotFound(c, err.Error())
		default:
			response.InternalError(c, "禁用 MFA 失败")
		}
		return
	}

	response.Success(c, gin.H{"message": "MFA 已禁用"})
}

// HandleVerifyMFA 验证 MFA 码（登录流程）
// @Summary 验证 MFA 码
// @Description 登录流程中验证 MFA 码
// @Tags MFA
// @Accept json
// @Produce json
// @Param request body dto.MFALoginRequest true "MFA 登录请求"
// @Success 200 {object} dto.LoginResponse
// @Failure 400 {object} response.ErrorResponse
// @Router /api/v1/auth/mfa/verify [post]
func (h *UserHandler) HandleVerifyMFA(c *gin.Context) {
	var req dto.MFALoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数错误: "+err.Error())
		return
	}

	err := h.mfaService.VerifyMFA(c.Request.Context(), req.UserID, req.Code)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrMFANotEnabled):
			response.BadRequest(c, err.Error())
		case errors.Is(err, service.ErrMFAInvalidCode):
			response.BadRequest(c, "验证码错误")
		case errors.Is(err, service.ErrUserNotFound):
			response.NotFound(c, err.Error())
		default:
			response.InternalError(c, "验证 MFA 失败")
		}
		return
	}

	// 验证成功后，生成完整的登录 Token
	// 这里需要调用 userService 来完成登录
	// 暂时返回成功消息
	response.Success(c, gin.H{"message": "MFA 验证成功"})
}

// HandleForgotPassword 处理忘记密码请求
// @Summary 忘记密码
// @Description 请求重置密码，系统将发送重置链接到用户邮箱
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body dto.ForgotPasswordRequest true "忘记密码请求"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.ErrorResponse
// @Router /api/v1/auth/forgot-password [post]
func (h *UserHandler) HandleForgotPassword(c *gin.Context) {
	var req dto.ForgotPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数错误: "+err.Error())
		return
	}

	err := h.userService.ForgotPassword(c.Request.Context(), &req)
	if err != nil {
		response.InternalError(c, "处理请求失败")
		return
	}

	response.Success(c, gin.H{"message": "如果该邮箱已注册，您将收到重置密码的邮件"})
}

// HandleVerifyResetToken 验证重置密码令牌
// @Summary 验证重置密码令牌
// @Description 验证重置密码令牌是否有效
// @Tags Auth
// @Produce json
// @Param token query string true "重置密码令牌"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.ErrorResponse
// @Router /api/v1/auth/verify-reset-token [get]
func (h *UserHandler) HandleVerifyResetToken(c *gin.Context) {
	var req dto.VerifyResetTokenRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.BadRequest(c, "请求参数错误: "+err.Error())
		return
	}

	err := h.userService.VerifyResetToken(c.Request.Context(), req.Token)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrInvalidResetToken):
			response.BadRequest(c, "重置密码令牌无效")
		case errors.Is(err, service.ErrResetTokenExpired):
			response.BadRequest(c, "重置密码令牌已过期")
		default:
			response.InternalError(c, "验证令牌失败")
		}
		return
	}

	response.Success(c, gin.H{"message": "令牌有效"})
}

// HandleResetPasswordWithToken 使用令牌重置密码
// @Summary 使用令牌重置密码
// @Description 使用重置密码令牌设置新密码
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body dto.ResetPasswordWithTokenRequest true "重置密码请求"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.ErrorResponse
// @Router /api/v1/auth/reset-password [post]
func (h *UserHandler) HandleResetPasswordWithToken(c *gin.Context) {
	var req dto.ResetPasswordWithTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数错误: "+err.Error())
		return
	}

	err := h.userService.ResetPasswordWithToken(c.Request.Context(), &req)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrInvalidResetToken):
			response.BadRequest(c, "重置密码令牌无效")
		case errors.Is(err, service.ErrResetTokenExpired):
			response.BadRequest(c, "重置密码令牌已过期")
		default:
			response.InternalError(c, "重置密码失败")
		}
		return
	}

	response.Success(c, gin.H{"message": "密码重置成功"})
}
