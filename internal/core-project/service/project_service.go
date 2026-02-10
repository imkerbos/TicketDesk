// Package service 提供项目业务逻辑层
package service

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/kerbos/ticketdesk/internal/core-project/dto"
	"github.com/kerbos/ticketdesk/internal/core-project/repository"
	userRepo "github.com/kerbos/ticketdesk/internal/core-user/repository"
	"github.com/kerbos/ticketdesk/internal/model"
	"github.com/kerbos/ticketdesk/pkg/logger"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// 业务错误定义
var (
	ErrProjectNotFound       = errors.New("项目不存在")
	ErrProjectKeyExists      = errors.New("项目 Key 已存在")
	ErrMemberNotFound        = errors.New("成员不存在")
	ErrMemberAlreadyExists   = errors.New("成员已存在")
	ErrCannotRemoveOwner     = errors.New("不能移除项目所有者")
	ErrIssueTypeNotFound     = errors.New("工单类型不存在")
	ErrNoPermission          = errors.New("没有操作权限")
	ErrRoleNotFound          = errors.New("角色不存在")
	ErrRoleKeyExists         = errors.New("角色 Key 已存在")
	ErrCannotDeleteSystemRole = errors.New("不能删除系统预置角色")
	ErrRoleMemberExists      = errors.New("用户已在该角色中")
	ErrRoleMemberNotFound    = errors.New("角色成员不存在")
)

// ProjectService 项目服务接口
type ProjectService interface {
	CreateProject(ctx context.Context, req *dto.CreateProjectRequest, creatorID uint64) (*dto.ProjectResponse, error)
	GetProject(ctx context.Context, key string) (*dto.ProjectResponse, error)
	GetProjectByID(ctx context.Context, id uint64) (*dto.ProjectResponse, error)
	UpdateProject(ctx context.Context, key string, req *dto.UpdateProjectRequest) (*dto.ProjectResponse, error)
	DeleteProject(ctx context.Context, key string) error
	ListProjects(ctx context.Context, req *dto.ListProjectsRequest, userID uint64, isAdmin bool) ([]*dto.ProjectResponse, int64, error)

	// 成员管理
	AddMember(ctx context.Context, projectKey string, req *dto.AddMemberRequest) (*dto.ProjectMemberResponse, error)
	UpdateMember(ctx context.Context, projectKey string, userID uint64, req *dto.UpdateMemberRequest) (*dto.ProjectMemberResponse, error)
	RemoveMember(ctx context.Context, projectKey string, userID uint64) error
	ListMembers(ctx context.Context, projectKey string) ([]*dto.ProjectMemberResponse, error)
	CheckPermission(ctx context.Context, projectKey string, userID uint64, requiredRoles ...string) (bool, error)

	// 工单类型管理
	CreateIssueType(ctx context.Context, projectKey string, req *dto.CreateIssueTypeRequest) (*dto.IssueTypeResponse, error)
	UpdateIssueType(ctx context.Context, id uint64, req *dto.UpdateIssueTypeRequest) (*dto.IssueTypeResponse, error)
	DeleteIssueType(ctx context.Context, id uint64) error
	ListIssueTypes(ctx context.Context, projectKey string) ([]*dto.IssueTypeResponse, error)

	// 项目角色管理
	CreateRole(ctx context.Context, projectKey string, req *dto.CreateProjectRoleRequest) (*dto.ProjectRoleResponse, error)
	UpdateRole(ctx context.Context, projectKey string, roleID uint64, req *dto.UpdateProjectRoleRequest) (*dto.ProjectRoleResponse, error)
	DeleteRole(ctx context.Context, projectKey string, roleID uint64) error
	ListRoles(ctx context.Context, projectKey string) ([]*dto.ProjectRoleResponse, error)
	AddRoleMember(ctx context.Context, projectKey string, roleID uint64, req *dto.AddRoleMemberRequest) (*dto.ProjectRoleMemberResponse, error)
	RemoveRoleMember(ctx context.Context, projectKey string, roleID, userID uint64) error
	ListRoleMembers(ctx context.Context, projectKey string, roleID uint64) ([]*dto.ProjectRoleMemberResponse, error)
	GetUserRoles(ctx context.Context, projectKey string, userID uint64) ([]*dto.ProjectRoleResponse, error)
}

// projectService 项目服务实现
type projectService struct {
	projectRepo       repository.ProjectRepository
	memberRepo        repository.ProjectMemberRepository
	issueTypeRepo     repository.IssueTypeRepository
	roleRepo          repository.ProjectRoleRepository
	roleMemberRepo    repository.ProjectRoleMemberRepository
	userRepo          userRepo.UserRepository
	db                *gorm.DB
}

// NewProjectService 创建项目服务实例
func NewProjectService(
	projectRepo repository.ProjectRepository,
	memberRepo repository.ProjectMemberRepository,
	issueTypeRepo repository.IssueTypeRepository,
	roleRepo repository.ProjectRoleRepository,
	roleMemberRepo repository.ProjectRoleMemberRepository,
	userRepo userRepo.UserRepository,
	db *gorm.DB,
) ProjectService {
	return &projectService{
		projectRepo:    projectRepo,
		memberRepo:     memberRepo,
		issueTypeRepo:  issueTypeRepo,
		roleRepo:       roleRepo,
		roleMemberRepo: roleMemberRepo,
		userRepo:       userRepo,
		db:             db,
	}
}

// CreateProject 创建项目
func (s *projectService) CreateProject(ctx context.Context, req *dto.CreateProjectRequest, creatorID uint64) (*dto.ProjectResponse, error) {
	// 转换为大写
	projectKey := strings.ToUpper(req.ProjectKey)

	// 检查项目 Key 是否已存在
	exists, err := s.projectRepo.ExistsByKey(ctx, projectKey)
	if err != nil {
		logger.Error("failed to check project key existence", zap.Error(err))
		return nil, fmt.Errorf("检查���目 Key 失败: %w", err)
	}
	if exists {
		return nil, ErrProjectKeyExists
	}

	// 检查负责人是否存在
	_, err = s.userRepo.GetByID(ctx, req.LeadUserID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("负责人不存在")
		}
		return nil, fmt.Errorf("查询负责人失败: %w", err)
	}

	// 创建项目
	project := &model.Project{
		ProjectKey:  projectKey,
		Name:        req.Name,
		Description: req.Description,
		LeadUserID:  req.LeadUserID,
		Status:      1,
	}

	if err := s.projectRepo.Create(ctx, project); err != nil {
		logger.Error("failed to create project", zap.Error(err))
		return nil, fmt.Errorf("创建项目失败: %w", err)
	}

	// 将创建者添加为项目所有者
	member := &model.ProjectMember{
		ProjectID: project.ID,
		UserID:    creatorID,
		Role:      "owner",
	}
	if err := s.memberRepo.Create(ctx, member); err != nil {
		logger.Warn("failed to add creator as owner", zap.Error(err))
	}

	// 如果负责人不是创建者，也添加为成员
	if req.LeadUserID != creatorID {
		leadMember := &model.ProjectMember{
			ProjectID: project.ID,
			UserID:    req.LeadUserID,
			Role:      "admin",
		}
		if err := s.memberRepo.Create(ctx, leadMember); err != nil {
			logger.Warn("failed to add lead as admin", zap.Error(err))
		}
	}

	// 初始化项目预置角色
	if err := model.InitProjectRoles(s.db, project.ID); err != nil {
		logger.Warn("failed to init project roles", zap.Error(err))
	} else {
		// 将创建者添加到 administrators 角色
		adminRole, err := s.roleRepo.GetByProjectAndKey(ctx, project.ID, "administrators")
		if err == nil {
			adminMember := &model.ProjectRoleMember{
				ProjectID: project.ID,
				RoleID:    adminRole.ID,
				UserID:    creatorID,
			}
			if err := s.roleMemberRepo.Create(ctx, adminMember); err != nil {
				logger.Warn("failed to add creator to administrators role", zap.Error(err))
			}
		}

		// 如果负责人不是创建者，也添加到 administrators 角色
		if req.LeadUserID != creatorID && err == nil {
			leadAdminMember := &model.ProjectRoleMember{
				ProjectID: project.ID,
				RoleID:    adminRole.ID,
				UserID:    req.LeadUserID,
			}
			if err := s.roleMemberRepo.Create(ctx, leadAdminMember); err != nil {
				logger.Warn("failed to add lead to administrators role", zap.Error(err))
			}
		}
	}

	// 模版项目：初始化字段方案和工作流方案；空项目跳过
	if req.Template != "blank" {
		// 初始化项目字段方案
		if err := model.InitProjectFieldSchemes(s.db, project.ID); err != nil {
			logger.Warn("failed to init project field schemes", zap.Error(err))
		}

		// 初始化项目工作流方案
		if err := model.InitProjectWorkflowSchemes(s.db, project.ID); err != nil {
			logger.Warn("failed to init project workflow schemes", zap.Error(err))
		}
	}

	logger.Info("project created successfully",
		zap.Uint64("project_id", project.ID),
		zap.String("project_key", project.ProjectKey),
	)

	return s.toProjectResponse(ctx, project), nil
}

// GetProject 获取项目详情
func (s *projectService) GetProject(ctx context.Context, key string) (*dto.ProjectResponse, error) {
	project, err := s.projectRepo.GetByKey(ctx, strings.ToUpper(key))
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrProjectNotFound
		}
		logger.Error("failed to get project", zap.String("key", key), zap.Error(err))
		return nil, fmt.Errorf("查询项目失败: %w", err)
	}

	return s.toProjectResponse(ctx, project), nil
}

// GetProjectByID 根据 ID 获取项目
func (s *projectService) GetProjectByID(ctx context.Context, id uint64) (*dto.ProjectResponse, error) {
	project, err := s.projectRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrProjectNotFound
		}
		return nil, fmt.Errorf("查询项目失败: %w", err)
	}

	return s.toProjectResponse(ctx, project), nil
}

// UpdateProject 更新项目
func (s *projectService) UpdateProject(ctx context.Context, key string, req *dto.UpdateProjectRequest) (*dto.ProjectResponse, error) {
	project, err := s.projectRepo.GetByKey(ctx, strings.ToUpper(key))
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrProjectNotFound
		}
		return nil, fmt.Errorf("查询项目失败: %w", err)
	}

	// 更新字段
	if req.Name != nil {
		project.Name = *req.Name
	}
	if req.Description != nil {
		project.Description = *req.Description
	}
	if req.LeadUserID != nil {
		// 检查负责人是否存在
		_, err = s.userRepo.GetByID(ctx, *req.LeadUserID)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil, fmt.Errorf("负责人不存在")
			}
			return nil, fmt.Errorf("查询负责人失败: %w", err)
		}
		project.LeadUserID = *req.LeadUserID
	}
	if req.Status != nil {
		project.Status = *req.Status
	}

	if err := s.projectRepo.Update(ctx, project); err != nil {
		logger.Error("failed to update project", zap.String("key", key), zap.Error(err))
		return nil, fmt.Errorf("更新项目失败: %w", err)
	}

	logger.Info("project updated successfully", zap.String("project_key", key))

	return s.toProjectResponse(ctx, project), nil
}

// DeleteProject 删除项目
func (s *projectService) DeleteProject(ctx context.Context, key string) error {
	project, err := s.projectRepo.GetByKey(ctx, strings.ToUpper(key))
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrProjectNotFound
		}
		return fmt.Errorf("查询项目失败: %w", err)
	}

	if err := s.projectRepo.Delete(ctx, project.ID); err != nil {
		logger.Error("failed to delete project", zap.String("key", key), zap.Error(err))
		return fmt.Errorf("删除项目失败: %w", err)
	}

	logger.Info("project deleted successfully", zap.String("project_key", key))

	return nil
}

// ListProjects 分页查询项目列表
func (s *projectService) ListProjects(ctx context.Context, req *dto.ListProjectsRequest, userID uint64, isAdmin bool) ([]*dto.ProjectResponse, int64, error) {
	page := req.GetDefaultPage()
	pageSize := req.GetDefaultPageSize()
	offset := (page - 1) * pageSize

	projects, total, err := s.projectRepo.List(ctx, offset, pageSize, req.Keyword, req.Status)
	if err != nil {
		logger.Error("failed to list projects", zap.Error(err))
		return nil, 0, fmt.Errorf("查询项目列表失败: %w", err)
	}

	responses := make([]*dto.ProjectResponse, len(projects))
	for i, project := range projects {
		responses[i] = s.toProjectResponse(ctx, project)
	}

	return responses, total, nil
}

// AddMember 添加项目成员
func (s *projectService) AddMember(ctx context.Context, projectKey string, req *dto.AddMemberRequest) (*dto.ProjectMemberResponse, error) {
	project, err := s.projectRepo.GetByKey(ctx, strings.ToUpper(projectKey))
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrProjectNotFound
		}
		return nil, fmt.Errorf("查询项目失败: %w", err)
	}

	// 检查用户是否存在
	user, err := s.userRepo.GetByID(ctx, req.UserID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("用户不存在")
		}
		return nil, fmt.Errorf("查询用户失败: %w", err)
	}

	// 检查是否已是成员
	isMember, err := s.memberRepo.IsMember(ctx, project.ID, req.UserID)
	if err != nil {
		return nil, fmt.Errorf("检查成员失败: %w", err)
	}
	if isMember {
		return nil, ErrMemberAlreadyExists
	}

	// 创建成员
	member := &model.ProjectMember{
		ProjectID: project.ID,
		UserID:    req.UserID,
		Role:      req.Role,
	}

	if err := s.memberRepo.Create(ctx, member); err != nil {
		logger.Error("failed to add member", zap.Error(err))
		return nil, fmt.Errorf("添加成员失败: %w", err)
	}

	logger.Info("member added successfully",
		zap.String("project_key", projectKey),
		zap.Uint64("user_id", req.UserID),
	)

	return s.toMemberResponse(member, user), nil
}

// UpdateMember 更新项目成员角色
func (s *projectService) UpdateMember(ctx context.Context, projectKey string, userID uint64, req *dto.UpdateMemberRequest) (*dto.ProjectMemberResponse, error) {
	project, err := s.projectRepo.GetByKey(ctx, strings.ToUpper(projectKey))
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrProjectNotFound
		}
		return nil, fmt.Errorf("查询项目失败: %w", err)
	}

	member, err := s.memberRepo.GetByProjectAndUser(ctx, project.ID, userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrMemberNotFound
		}
		return nil, fmt.Errorf("查询成员失败: %w", err)
	}

	member.Role = req.Role

	if err := s.memberRepo.Update(ctx, member); err != nil {
		logger.Error("failed to update member", zap.Error(err))
		return nil, fmt.Errorf("更新成员失败: %w", err)
	}

	user, _ := s.userRepo.GetByID(ctx, userID)

	return s.toMemberResponse(member, user), nil
}

// RemoveMember 移除项目成员
func (s *projectService) RemoveMember(ctx context.Context, projectKey string, userID uint64) error {
	project, err := s.projectRepo.GetByKey(ctx, strings.ToUpper(projectKey))
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrProjectNotFound
		}
		return fmt.Errorf("查询项目失败: %w", err)
	}

	member, err := s.memberRepo.GetByProjectAndUser(ctx, project.ID, userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrMemberNotFound
		}
		return fmt.Errorf("查询成员失败: %w", err)
	}

	// 不能移除所有者
	if member.Role == "owner" {
		return ErrCannotRemoveOwner
	}

	if err := s.memberRepo.Delete(ctx, member.ID); err != nil {
		logger.Error("failed to remove member", zap.Error(err))
		return fmt.Errorf("移除成员失败: %w", err)
	}

	logger.Info("member removed successfully",
		zap.String("project_key", projectKey),
		zap.Uint64("user_id", userID),
	)

	return nil
}

// ListMembers 获取项目成员列表
func (s *projectService) ListMembers(ctx context.Context, projectKey string) ([]*dto.ProjectMemberResponse, error) {
	project, err := s.projectRepo.GetByKey(ctx, strings.ToUpper(projectKey))
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrProjectNotFound
		}
		return nil, fmt.Errorf("查询项目失败: %w", err)
	}

	members, err := s.memberRepo.ListByProject(ctx, project.ID)
	if err != nil {
		return nil, fmt.Errorf("查询成员列表失败: %w", err)
	}

	responses := make([]*dto.ProjectMemberResponse, len(members))
	for i, member := range members {
		user, _ := s.userRepo.GetByID(ctx, member.UserID)
		responses[i] = s.toMemberResponse(member, user)
	}

	return responses, nil
}

// CheckPermission 检查用户在项目中的权限
func (s *projectService) CheckPermission(ctx context.Context, projectKey string, userID uint64, requiredRoles ...string) (bool, error) {
	project, err := s.projectRepo.GetByKey(ctx, strings.ToUpper(projectKey))
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return false, ErrProjectNotFound
		}
		return false, err
	}

	role, err := s.memberRepo.GetUserRole(ctx, project.ID, userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return false, nil
		}
		return false, err
	}

	for _, r := range requiredRoles {
		if role == r {
			return true, nil
		}
	}

	return false, nil
}

// CreateIssueType 创建工单类型
func (s *projectService) CreateIssueType(ctx context.Context, projectKey string, req *dto.CreateIssueTypeRequest) (*dto.IssueTypeResponse, error) {
	project, err := s.projectRepo.GetByKey(ctx, strings.ToUpper(projectKey))
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrProjectNotFound
		}
		return nil, fmt.Errorf("查询项目失败: %w", err)
	}

	issueType := &model.IssueType{
		ProjectID:   &project.ID,
		Name:        req.Name,
		DisplayName: req.DisplayName,
		Description: req.Description,
		Icon:        req.Icon,
		Color:       req.Color,
	}

	if err := s.issueTypeRepo.Create(ctx, issueType); err != nil {
		logger.Error("failed to create issue type", zap.Error(err))
		return nil, fmt.Errorf("创建工单类型失败: %w", err)
	}

	return s.toIssueTypeResponse(issueType), nil
}

// UpdateIssueType 更新工单类型
func (s *projectService) UpdateIssueType(ctx context.Context, id uint64, req *dto.UpdateIssueTypeRequest) (*dto.IssueTypeResponse, error) {
	issueType, err := s.issueTypeRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrIssueTypeNotFound
		}
		return nil, fmt.Errorf("查询工单类型失败: %w", err)
	}

	if req.DisplayName != nil {
		issueType.DisplayName = *req.DisplayName
	}
	if req.Description != nil {
		issueType.Description = *req.Description
	}
	if req.Icon != nil {
		issueType.Icon = *req.Icon
	}
	if req.Color != nil {
		issueType.Color = *req.Color
	}

	if err := s.issueTypeRepo.Update(ctx, issueType); err != nil {
		return nil, fmt.Errorf("更新工单类型失败: %w", err)
	}

	return s.toIssueTypeResponse(issueType), nil
}

// DeleteIssueType 删除工单类型
func (s *projectService) DeleteIssueType(ctx context.Context, id uint64) error {
	issueType, err := s.issueTypeRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrIssueTypeNotFound
		}
		return fmt.Errorf("查询工单类型失败: %w", err)
	}

	// 不能删除全局工单类型
	if issueType.ProjectID == nil {
		return fmt.Errorf("不能删除全局工单类型")
	}

	if err := s.issueTypeRepo.Delete(ctx, id); err != nil {
		return fmt.Errorf("删除工单类型失败: %w", err)
	}

	return nil
}

// ListIssueTypes 获取项目可用的工单类型
func (s *projectService) ListIssueTypes(ctx context.Context, projectKey string) ([]*dto.IssueTypeResponse, error) {
	project, err := s.projectRepo.GetByKey(ctx, strings.ToUpper(projectKey))
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrProjectNotFound
		}
		return nil, fmt.Errorf("查询项目失败: %w", err)
	}

	issueTypes, err := s.issueTypeRepo.ListAvailableForProject(ctx, project.ID)
	if err != nil {
		return nil, fmt.Errorf("查询工单类型失败: %w", err)
	}

	responses := make([]*dto.IssueTypeResponse, len(issueTypes))
	for i, it := range issueTypes {
		responses[i] = s.toIssueTypeResponse(it)
	}

	return responses, nil
}

// toProjectResponse 将项目模型转换为响应 DTO
func (s *projectService) toProjectResponse(ctx context.Context, project *model.Project) *dto.ProjectResponse {
	resp := &dto.ProjectResponse{
		ID:          project.ID,
		ProjectKey:  project.ProjectKey,
		Name:        project.Name,
		Description: project.Description,
		LeadUserID:  project.LeadUserID,
		Status:      project.Status,
		CreatedAt:   project.CreatedAt,
		UpdatedAt:   project.UpdatedAt,
	}

	// 获取负责人信息
	if leadUser, err := s.userRepo.GetByID(ctx, project.LeadUserID); err == nil {
		resp.LeadUser = &dto.UserBrief{
			ID:          leadUser.ID,
			Username:    leadUser.Username,
			DisplayName: leadUser.DisplayName,
			AvatarURL:   leadUser.AvatarURL,
		}
	}

	// 获取成员数量
	if count, err := s.memberRepo.CountByProject(ctx, project.ID); err == nil {
		resp.MemberCount = int(count)
	}

	return resp
}

// toMemberResponse 将成员模型转换为响应 DTO
func (s *projectService) toMemberResponse(member *model.ProjectMember, user *model.User) *dto.ProjectMemberResponse {
	resp := &dto.ProjectMemberResponse{
		ID:        member.ID,
		ProjectID: member.ProjectID,
		UserID:    member.UserID,
		Role:      member.Role,
		CreatedAt: member.CreatedAt,
	}

	if user != nil {
		resp.User = &dto.UserBrief{
			ID:          user.ID,
			Username:    user.Username,
			DisplayName: user.DisplayName,
			AvatarURL:   user.AvatarURL,
		}
	}

	return resp
}

// toIssueTypeResponse 将工单类型模型转换为响应 DTO
func (s *projectService) toIssueTypeResponse(issueType *model.IssueType) *dto.IssueTypeResponse {
	return &dto.IssueTypeResponse{
		ID:          issueType.ID,
		ProjectID:   issueType.ProjectID,
		Name:        issueType.Name,
		DisplayName: issueType.DisplayName,
		Description: issueType.Description,
		Icon:        issueType.Icon,
		Color:       issueType.Color,
		CreatedAt:   issueType.CreatedAt,
		UpdatedAt:   issueType.UpdatedAt,
	}
}

// ============ 项目角色管理 ============

// CreateRole 创建项目角色
func (s *projectService) CreateRole(ctx context.Context, projectKey string, req *dto.CreateProjectRoleRequest) (*dto.ProjectRoleResponse, error) {
	project, err := s.projectRepo.GetByKey(ctx, strings.ToUpper(projectKey))
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrProjectNotFound
		}
		return nil, fmt.Errorf("查询项目失败: %w", err)
	}

	// 检查角色 Key 是否已存在
	_, err = s.roleRepo.GetByProjectAndKey(ctx, project.ID, req.RoleKey)
	if err == nil {
		return nil, ErrRoleKeyExists
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, fmt.Errorf("检查角色失败: %w", err)
	}

	role := &model.ProjectRole{
		ProjectID:   project.ID,
		RoleKey:     req.RoleKey,
		RoleName:    req.RoleName,
		Description: req.Description,
		IsSystem:    false,
		SortOrder:   100, // 自定义角色排在后面
	}

	if err := s.roleRepo.Create(ctx, role); err != nil {
		logger.Error("failed to create project role", zap.Error(err))
		return nil, fmt.Errorf("创建角色失败: %w", err)
	}

	logger.Info("project role created",
		zap.String("project_key", projectKey),
		zap.String("role_key", req.RoleKey),
	)

	return s.toRoleResponse(ctx, role), nil
}

// UpdateRole 更新项目角色
func (s *projectService) UpdateRole(ctx context.Context, projectKey string, roleID uint64, req *dto.UpdateProjectRoleRequest) (*dto.ProjectRoleResponse, error) {
	project, err := s.projectRepo.GetByKey(ctx, strings.ToUpper(projectKey))
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrProjectNotFound
		}
		return nil, fmt.Errorf("查询项目失败: %w", err)
	}

	role, err := s.roleRepo.GetByID(ctx, roleID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrRoleNotFound
		}
		return nil, fmt.Errorf("查询角色失败: %w", err)
	}

	// 验证角色属于该项目
	if role.ProjectID != project.ID {
		return nil, ErrRoleNotFound
	}

	if req.RoleName != nil {
		role.RoleName = *req.RoleName
	}
	if req.Description != nil {
		role.Description = *req.Description
	}

	if err := s.roleRepo.Update(ctx, role); err != nil {
		logger.Error("failed to update project role", zap.Error(err))
		return nil, fmt.Errorf("更新角色失败: %w", err)
	}

	return s.toRoleResponse(ctx, role), nil
}

// DeleteRole 删除项目角色
func (s *projectService) DeleteRole(ctx context.Context, projectKey string, roleID uint64) error {
	project, err := s.projectRepo.GetByKey(ctx, strings.ToUpper(projectKey))
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrProjectNotFound
		}
		return fmt.Errorf("查询项目失败: %w", err)
	}

	role, err := s.roleRepo.GetByID(ctx, roleID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrRoleNotFound
		}
		return fmt.Errorf("查询角色失败: %w", err)
	}

	// 验证角色属于该项目
	if role.ProjectID != project.ID {
		return ErrRoleNotFound
	}

	// 不能删除系统预置角色
	if role.IsSystem {
		return ErrCannotDeleteSystemRole
	}

	if err := s.roleRepo.Delete(ctx, roleID); err != nil {
		logger.Error("failed to delete project role", zap.Error(err))
		return fmt.Errorf("删除角色失败: %w", err)
	}

	logger.Info("project role deleted",
		zap.String("project_key", projectKey),
		zap.Uint64("role_id", roleID),
	)

	return nil
}

// ListRoles 获取项目角色列表
func (s *projectService) ListRoles(ctx context.Context, projectKey string) ([]*dto.ProjectRoleResponse, error) {
	project, err := s.projectRepo.GetByKey(ctx, strings.ToUpper(projectKey))
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrProjectNotFound
		}
		return nil, fmt.Errorf("查询项目失败: %w", err)
	}

	roles, err := s.roleRepo.ListByProject(ctx, project.ID)
	if err != nil {
		return nil, fmt.Errorf("查询角色列表失败: %w", err)
	}

	responses := make([]*dto.ProjectRoleResponse, len(roles))
	for i, role := range roles {
		responses[i] = s.toRoleResponse(ctx, role)
	}

	return responses, nil
}

// AddRoleMember 添加角色成员
func (s *projectService) AddRoleMember(ctx context.Context, projectKey string, roleID uint64, req *dto.AddRoleMemberRequest) (*dto.ProjectRoleMemberResponse, error) {
	project, err := s.projectRepo.GetByKey(ctx, strings.ToUpper(projectKey))
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrProjectNotFound
		}
		return nil, fmt.Errorf("查询项目失败: %w", err)
	}

	role, err := s.roleRepo.GetByID(ctx, roleID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrRoleNotFound
		}
		return nil, fmt.Errorf("查询角色失败: %w", err)
	}

	// 验证角色属于该项目
	if role.ProjectID != project.ID {
		return nil, ErrRoleNotFound
	}

	// 检查用户是否存在
	user, err := s.userRepo.GetByID(ctx, req.UserID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("用户不存在")
		}
		return nil, fmt.Errorf("查询用户失败: %w", err)
	}

	// 检查是否已是角色成员
	exists, err := s.roleMemberRepo.Exists(ctx, roleID, req.UserID)
	if err != nil {
		return nil, fmt.Errorf("检查角色成员失败: %w", err)
	}
	if exists {
		return nil, ErrRoleMemberExists
	}

	member := &model.ProjectRoleMember{
		ProjectID: project.ID,
		RoleID:    roleID,
		UserID:    req.UserID,
	}

	if err := s.roleMemberRepo.Create(ctx, member); err != nil {
		logger.Error("failed to add role member", zap.Error(err))
		return nil, fmt.Errorf("添加角色成员失败: %w", err)
	}

	logger.Info("role member added",
		zap.String("project_key", projectKey),
		zap.Uint64("role_id", roleID),
		zap.Uint64("user_id", req.UserID),
	)

	return s.toRoleMemberResponse(member, user), nil
}

// RemoveRoleMember 移除角色成员
func (s *projectService) RemoveRoleMember(ctx context.Context, projectKey string, roleID, userID uint64) error {
	project, err := s.projectRepo.GetByKey(ctx, strings.ToUpper(projectKey))
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrProjectNotFound
		}
		return fmt.Errorf("查询项目失败: %w", err)
	}

	role, err := s.roleRepo.GetByID(ctx, roleID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrRoleNotFound
		}
		return fmt.Errorf("查询角色失败: %w", err)
	}

	// 验证角色属于该项目
	if role.ProjectID != project.ID {
		return ErrRoleNotFound
	}

	// 检查成员是否存在
	_, err = s.roleMemberRepo.GetByRoleAndUser(ctx, roleID, userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrRoleMemberNotFound
		}
		return fmt.Errorf("查询角色成员失败: %w", err)
	}

	if err := s.roleMemberRepo.DeleteByRoleAndUser(ctx, roleID, userID); err != nil {
		logger.Error("failed to remove role member", zap.Error(err))
		return fmt.Errorf("移除角色成员失败: %w", err)
	}

	logger.Info("role member removed",
		zap.String("project_key", projectKey),
		zap.Uint64("role_id", roleID),
		zap.Uint64("user_id", userID),
	)

	return nil
}

// ListRoleMembers 获取角色成员列表
func (s *projectService) ListRoleMembers(ctx context.Context, projectKey string, roleID uint64) ([]*dto.ProjectRoleMemberResponse, error) {
	project, err := s.projectRepo.GetByKey(ctx, strings.ToUpper(projectKey))
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrProjectNotFound
		}
		return nil, fmt.Errorf("查询项目失败: %w", err)
	}

	role, err := s.roleRepo.GetByID(ctx, roleID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrRoleNotFound
		}
		return nil, fmt.Errorf("查询角色失败: %w", err)
	}

	// 验证角色属于该项目
	if role.ProjectID != project.ID {
		return nil, ErrRoleNotFound
	}

	members, err := s.roleMemberRepo.ListByRole(ctx, roleID)
	if err != nil {
		return nil, fmt.Errorf("查询角色成员列表失败: %w", err)
	}

	responses := make([]*dto.ProjectRoleMemberResponse, len(members))
	for i, member := range members {
		user, _ := s.userRepo.GetByID(ctx, member.UserID)
		responses[i] = s.toRoleMemberResponse(member, user)
	}

	return responses, nil
}

// GetUserRoles 获取用户在项目中的角色
func (s *projectService) GetUserRoles(ctx context.Context, projectKey string, userID uint64) ([]*dto.ProjectRoleResponse, error) {
	project, err := s.projectRepo.GetByKey(ctx, strings.ToUpper(projectKey))
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrProjectNotFound
		}
		return nil, fmt.Errorf("查询项目失败: %w", err)
	}

	members, err := s.roleMemberRepo.ListByProjectAndUser(ctx, project.ID, userID)
	if err != nil {
		return nil, fmt.Errorf("查询用户角色失败: %w", err)
	}

	responses := make([]*dto.ProjectRoleResponse, 0, len(members))
	for _, member := range members {
		role, err := s.roleRepo.GetByID(ctx, member.RoleID)
		if err != nil {
			continue
		}
		responses = append(responses, s.toRoleResponse(ctx, role))
	}

	return responses, nil
}

// toRoleResponse 将角色模型转换为响应 DTO
func (s *projectService) toRoleResponse(ctx context.Context, role *model.ProjectRole) *dto.ProjectRoleResponse {
	resp := &dto.ProjectRoleResponse{
		ID:          role.ID,
		ProjectID:   role.ProjectID,
		RoleKey:     role.RoleKey,
		RoleName:    role.RoleName,
		Description: role.Description,
		IsSystem:    role.IsSystem,
		SortOrder:   role.SortOrder,
		CreatedAt:   role.CreatedAt,
		UpdatedAt:   role.UpdatedAt,
	}

	// 获取成员数量
	if count, err := s.roleRepo.CountMembersByRole(ctx, role.ID); err == nil {
		resp.MemberCount = int(count)
	}

	return resp
}

// toRoleMemberResponse 将角色成员模型转换为响应 DTO
func (s *projectService) toRoleMemberResponse(member *model.ProjectRoleMember, user *model.User) *dto.ProjectRoleMemberResponse {
	resp := &dto.ProjectRoleMemberResponse{
		ID:        member.ID,
		ProjectID: member.ProjectID,
		RoleID:    member.RoleID,
		UserID:    member.UserID,
		CreatedAt: member.CreatedAt,
	}

	if user != nil {
		resp.User = &dto.UserBrief{
			ID:          user.ID,
			Username:    user.Username,
			DisplayName: user.DisplayName,
			AvatarURL:   user.AvatarURL,
		}
	}

	return resp
}
