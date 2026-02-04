// Package service 提供用户业务逻辑层
package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/kerbos/ticketdesk/internal/core-user/dto"
	"github.com/kerbos/ticketdesk/internal/core-user/repository"
	"github.com/kerbos/ticketdesk/internal/model"
	"github.com/kerbos/ticketdesk/pkg/jwt"
	"github.com/kerbos/ticketdesk/pkg/logger"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// 业务错误定义
var (
	ErrUserNotFound       = errors.New("用户不存在")
	ErrUserDisabled       = errors.New("用户已被禁用")
	ErrUsernameExists     = errors.New("用户名已存在")
	ErrEmailExists        = errors.New("邮箱已存在")
	ErrInvalidCredentials = errors.New("用户名或密码错误")
	ErrInvalidOldPassword = errors.New("原密码错误")
)

// UserService 用户服务接口
type UserService interface {
	Register(ctx context.Context, req *dto.RegisterRequest) (*dto.UserResponse, error)
	Login(ctx context.Context, req *dto.LoginRequest) (*dto.LoginResponse, error)
	RefreshToken(ctx context.Context, refreshToken string) (*dto.LoginResponse, error)
	GetCurrentUser(ctx context.Context, userID uint64) (*dto.UserResponse, error)
	GetUser(ctx context.Context, id uint64) (*dto.UserResponse, error)
	CreateUser(ctx context.Context, req *dto.CreateUserRequest) (*dto.UserResponse, error)
	UpdateUser(ctx context.Context, id uint64, req *dto.UpdateUserRequest) (*dto.UserResponse, error)
	UpdatePassword(ctx context.Context, userID uint64, req *dto.UpdatePasswordRequest) error
	ResetPassword(ctx context.Context, id uint64, req *dto.ResetPasswordRequest) error
	EnableUser(ctx context.Context, id uint64) error
	DisableUser(ctx context.Context, id uint64) error
	DeleteUser(ctx context.Context, id uint64) error
	ListUsers(ctx context.Context, req *dto.ListUsersRequest) ([]*dto.UserResponse, int64, error)
}

// userService 用户服务实现
type userService struct {
	userRepo     repository.UserRepository
	userRoleRepo repository.UserRoleRepository
	jwtManager   *jwt.Manager
}

// NewUserService 创建用户服务实例
func NewUserService(
	userRepo repository.UserRepository,
	userRoleRepo repository.UserRoleRepository,
	jwtManager *jwt.Manager,
) UserService {
	return &userService{
		userRepo:     userRepo,
		userRoleRepo: userRoleRepo,
		jwtManager:   jwtManager,
	}
}

// Register 用户注册
func (s *userService) Register(ctx context.Context, req *dto.RegisterRequest) (*dto.UserResponse, error) {
	// 检查用户名是否已存在
	exists, err := s.userRepo.ExistsByUsername(ctx, req.Username)
	if err != nil {
		logger.Error("failed to check username existence", zap.Error(err))
		return nil, fmt.Errorf("检查用户名失败: %w", err)
	}
	if exists {
		return nil, ErrUsernameExists
	}

	// 检查邮箱是否已存在
	exists, err = s.userRepo.ExistsByEmail(ctx, req.Email)
	if err != nil {
		logger.Error("failed to check email existence", zap.Error(err))
		return nil, fmt.Errorf("检查邮箱失败: %w", err)
	}
	if exists {
		return nil, ErrEmailExists
	}

	// 密码加密
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		logger.Error("failed to hash password", zap.Error(err))
		return nil, fmt.Errorf("密码加密失败: %w", err)
	}

	// 创建用户
	user := &model.User{
		Username:     req.Username,
		Email:        req.Email,
		PasswordHash: string(hashedPassword),
		DisplayName:  req.DisplayName,
		Status:       1, // 默认启用
	}

	if user.DisplayName == "" {
		user.DisplayName = user.Username
	}

	if err := s.userRepo.Create(ctx, user); err != nil {
		logger.Error("failed to create user", zap.Error(err))
		return nil, fmt.Errorf("创建用户失败: %w", err)
	}

	// 分配默认角色
	if err := s.userRoleRepo.AssignRole(ctx, user.ID, "user"); err != nil {
		logger.Warn("failed to assign default role", zap.Uint64("user_id", user.ID), zap.Error(err))
	}

	logger.Info("user registered successfully",
		zap.Uint64("user_id", user.ID),
		zap.String("username", user.Username),
	)

	return s.toUserResponse(ctx, user), nil
}

// Login 用户登录
func (s *userService) Login(ctx context.Context, req *dto.LoginRequest) (*dto.LoginResponse, error) {
	// 查找用户
	user, err := s.userRepo.GetByUsername(ctx, req.Username)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrInvalidCredentials
		}
		logger.Error("failed to get user by username", zap.Error(err))
		return nil, fmt.Errorf("查询用户失败: %w", err)
	}

	// 检查用户状态
	if user.Status == 0 {
		return nil, ErrUserDisabled
	}

	// 验证密码
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		return nil, ErrInvalidCredentials
	}

	// 更新最后登录时间
	now := time.Now()
	user.LastLoginAt = &now
	if err := s.userRepo.Update(ctx, user); err != nil {
		logger.Warn("failed to update last login time", zap.Uint64("user_id", user.ID), zap.Error(err))
		// 不影响登录流程，继续执行
	}

	// 生成 Token
	accessToken, err := s.jwtManager.GenerateAccessToken(user.ID, user.Username)
	if err != nil {
		logger.Error("failed to generate access token", zap.Error(err))
		return nil, fmt.Errorf("生成 Token 失败: %w", err)
	}

	refreshToken, err := s.jwtManager.GenerateRefreshToken(user.ID, user.Username)
	if err != nil {
		logger.Error("failed to generate refresh token", zap.Error(err))
		return nil, fmt.Errorf("生成 Refresh Token 失败: %w", err)
	}

	logger.Info("user logged in successfully",
		zap.Uint64("user_id", user.ID),
		zap.String("username", user.Username),
	)

	return &dto.LoginResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    7200, // 2小时
		User:         *s.toUserResponse(ctx, user),
	}, nil
}

// RefreshToken 刷新 Token
func (s *userService) RefreshToken(ctx context.Context, refreshToken string) (*dto.LoginResponse, error) {
	// 解析 Refresh Token
	claims, err := s.jwtManager.ParseToken(refreshToken)
	if err != nil {
		return nil, fmt.Errorf("无效的 Refresh Token: %w", err)
	}

	// 查找用户
	user, err := s.userRepo.GetByID(ctx, claims.UserID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, fmt.Errorf("查询用户失败: %w", err)
	}

	// 检查用户状态
	if user.Status == 0 {
		return nil, ErrUserDisabled
	}

	// 生成新的 Token
	newAccessToken, err := s.jwtManager.GenerateAccessToken(user.ID, user.Username)
	if err != nil {
		return nil, fmt.Errorf("生成 Token 失败: %w", err)
	}

	newRefreshToken, err := s.jwtManager.GenerateRefreshToken(user.ID, user.Username)
	if err != nil {
		return nil, fmt.Errorf("生成 Refresh Token 失败: %w", err)
	}

	return &dto.LoginResponse{
		AccessToken:  newAccessToken,
		RefreshToken: newRefreshToken,
		ExpiresIn:    7200,
		User:         *s.toUserResponse(ctx, user),
	}, nil
}

// GetCurrentUser 获取当前用户信息
func (s *userService) GetCurrentUser(ctx context.Context, userID uint64) (*dto.UserResponse, error) {
	return s.GetUser(ctx, userID)
}

// GetUser 获取用户详情
func (s *userService) GetUser(ctx context.Context, id uint64) (*dto.UserResponse, error) {
	user, err := s.userRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrUserNotFound
		}
		logger.Error("failed to get user", zap.Uint64("id", id), zap.Error(err))
		return nil, fmt.Errorf("查询用户失败: %w", err)
	}

	return s.toUserResponse(ctx, user), nil
}

// CreateUser 创建用户（管理员）
func (s *userService) CreateUser(ctx context.Context, req *dto.CreateUserRequest) (*dto.UserResponse, error) {
	// 检查用户名是否已存在
	exists, err := s.userRepo.ExistsByUsername(ctx, req.Username)
	if err != nil {
		logger.Error("failed to check username existence", zap.Error(err))
		return nil, fmt.Errorf("检查用户名失败: %w", err)
	}
	if exists {
		return nil, ErrUsernameExists
	}

	// 检查邮箱是否已存在
	exists, err = s.userRepo.ExistsByEmail(ctx, req.Email)
	if err != nil {
		logger.Error("failed to check email existence", zap.Error(err))
		return nil, fmt.Errorf("检查邮箱失败: %w", err)
	}
	if exists {
		return nil, ErrEmailExists
	}

	// 密码加密
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		logger.Error("failed to hash password", zap.Error(err))
		return nil, fmt.Errorf("密码加密失败: %w", err)
	}

	// 设置默认状态
	status := int8(1) // 默认启用
	if req.Status != nil {
		status = *req.Status
	}

	// 创建用户
	user := &model.User{
		Username:     req.Username,
		Email:        req.Email,
		PasswordHash: string(hashedPassword),
		DisplayName:  req.DisplayName,
		Status:       status,
	}

	if user.DisplayName == "" {
		user.DisplayName = user.Username
	}

	if err := s.userRepo.Create(ctx, user); err != nil {
		logger.Error("failed to create user", zap.Error(err))
		return nil, fmt.Errorf("创建用户失败: %w", err)
	}

	// 分配角色
	if len(req.Roles) > 0 {
		for _, roleName := range req.Roles {
			if err := s.userRoleRepo.AssignRole(ctx, user.ID, roleName); err != nil {
				logger.Warn("failed to assign role",
					zap.Uint64("user_id", user.ID),
					zap.String("role", roleName),
					zap.Error(err),
				)
			}
		}
	} else {
		// 如果没有指定角色，分配默认角色
		if err := s.userRoleRepo.AssignRole(ctx, user.ID, "user"); err != nil {
			logger.Warn("failed to assign default role", zap.Uint64("user_id", user.ID), zap.Error(err))
		}
	}

	logger.Info("user created by admin",
		zap.Uint64("user_id", user.ID),
		zap.String("username", user.Username),
	)

	return s.toUserResponse(ctx, user), nil
}

// UpdateUser 更新用户信息
func (s *userService) UpdateUser(ctx context.Context, id uint64, req *dto.UpdateUserRequest) (*dto.UserResponse, error) {
	user, err := s.userRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, fmt.Errorf("查询用户失败: %w", err)
	}

	// 更新字段
	if req.DisplayName != nil {
		user.DisplayName = *req.DisplayName
	}
	if req.AvatarURL != nil {
		user.AvatarURL = *req.AvatarURL
	}
	if req.Email != nil && *req.Email != user.Email {
		// 检查邮箱是否已被使用
		exists, err := s.userRepo.ExistsByEmail(ctx, *req.Email)
		if err != nil {
			return nil, fmt.Errorf("检查邮箱失败: %w", err)
		}
		if exists {
			return nil, ErrEmailExists
		}
		user.Email = *req.Email
	}

	if err := s.userRepo.Update(ctx, user); err != nil {
		logger.Error("failed to update user", zap.Uint64("id", id), zap.Error(err))
		return nil, fmt.Errorf("更新用户失败: %w", err)
	}

	logger.Info("user updated successfully", zap.Uint64("user_id", id))

	return s.toUserResponse(ctx, user), nil
}

// UpdatePassword 修改密码
func (s *userService) UpdatePassword(ctx context.Context, userID uint64, req *dto.UpdatePasswordRequest) error {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrUserNotFound
		}
		return fmt.Errorf("查询用户失败: %w", err)
	}

	// 验证原密码
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.OldPassword)); err != nil {
		return ErrInvalidOldPassword
	}

	// 加密新密码
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("密码加密失败: %w", err)
	}

	user.PasswordHash = string(hashedPassword)
	if err := s.userRepo.Update(ctx, user); err != nil {
		return fmt.Errorf("更新密码失败: %w", err)
	}

	logger.Info("user password updated successfully", zap.Uint64("user_id", userID))

	return nil
}

// ResetPassword 重置密码（管理员）
func (s *userService) ResetPassword(ctx context.Context, id uint64, req *dto.ResetPasswordRequest) error {
	user, err := s.userRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrUserNotFound
		}
		return fmt.Errorf("查询用户失败: %w", err)
	}

	// 加密新密码
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("密码加密失败: %w", err)
	}

	user.PasswordHash = string(hashedPassword)
	if err := s.userRepo.Update(ctx, user); err != nil {
		return fmt.Errorf("重置密码失败: %w", err)
	}

	logger.Info("user password reset by admin", zap.Uint64("user_id", id))

	return nil
}

// EnableUser 启用用户
func (s *userService) EnableUser(ctx context.Context, id uint64) error {
	user, err := s.userRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrUserNotFound
		}
		return fmt.Errorf("查询用户失败: %w", err)
	}

	if user.Status == 1 {
		return nil // 已经是启用状态
	}

	user.Status = 1
	if err := s.userRepo.Update(ctx, user); err != nil {
		logger.Error("failed to enable user", zap.Uint64("id", id), zap.Error(err))
		return fmt.Errorf("启用用户失败: %w", err)
	}

	logger.Info("user enabled successfully", zap.Uint64("user_id", id))

	return nil
}

// DisableUser 禁用用户
func (s *userService) DisableUser(ctx context.Context, id uint64) error {
	user, err := s.userRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrUserNotFound
		}
		return fmt.Errorf("查询用户失败: %w", err)
	}

	if user.Status == 0 {
		return nil // 已经是禁用状态
	}

	user.Status = 0
	if err := s.userRepo.Update(ctx, user); err != nil {
		logger.Error("failed to disable user", zap.Uint64("id", id), zap.Error(err))
		return fmt.Errorf("禁用用户失败: %w", err)
	}

	logger.Info("user disabled successfully", zap.Uint64("user_id", id))

	return nil
}

// DeleteUser 删除用户
func (s *userService) DeleteUser(ctx context.Context, id uint64) error {
	// 检查用户是否存在
	_, err := s.userRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrUserNotFound
		}
		return fmt.Errorf("查询用户失败: %w", err)
	}

	if err := s.userRepo.Delete(ctx, id); err != nil {
		logger.Error("failed to delete user", zap.Uint64("id", id), zap.Error(err))
		return fmt.Errorf("删除用户失败: %w", err)
	}

	logger.Info("user deleted successfully", zap.Uint64("user_id", id))

	return nil
}

// ListUsers 分页查询用户列表
func (s *userService) ListUsers(ctx context.Context, req *dto.ListUsersRequest) ([]*dto.UserResponse, int64, error) {
	page := req.GetDefaultPage()
	pageSize := req.GetDefaultPageSize()
	offset := (page - 1) * pageSize

	users, total, err := s.userRepo.List(ctx, offset, pageSize, req.Keyword, req.Status)
	if err != nil {
		logger.Error("failed to list users", zap.Error(err))
		return nil, 0, fmt.Errorf("查询用户列表失败: %w", err)
	}

	responses := make([]*dto.UserResponse, len(users))
	for i, user := range users {
		responses[i] = s.toUserResponse(ctx, user)
	}

	return responses, total, nil
}

// toUserResponse 将用户模型转换为响应 DTO
func (s *userService) toUserResponse(ctx context.Context, user *model.User) *dto.UserResponse {
	resp := &dto.UserResponse{
		ID:          user.ID,
		Username:    user.Username,
		Email:       user.Email,
		DisplayName: user.DisplayName,
		AvatarURL:   user.AvatarURL,
		Status:      user.Status,
		MFAEnabled:  user.MFAEnabled,
		LastLoginAt: user.LastLoginAt,
		CreatedAt:   user.CreatedAt,
		UpdatedAt:   user.UpdatedAt,
	}

	// 获取用户角色
	roles, err := s.userRoleRepo.GetUserRoleNames(ctx, user.ID)
	if err != nil {
		logger.Warn("failed to get user roles", zap.Uint64("user_id", user.ID), zap.Error(err))
	} else {
		resp.Roles = roles
	}

	return resp
}
