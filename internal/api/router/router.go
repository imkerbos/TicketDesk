// Package router 提供路由配置
package router

import (
	"time"

	"github.com/gin-gonic/gin"
	alertHandler "github.com/kerbos/ticketdesk/internal/integration-alert/handler"
	alertRepo "github.com/kerbos/ticketdesk/internal/integration-alert/repository"
	alertService "github.com/kerbos/ticketdesk/internal/integration-alert/service"
	"github.com/kerbos/ticketdesk/internal/api/middleware"
	"github.com/kerbos/ticketdesk/internal/api/response"
	issueHandler "github.com/kerbos/ticketdesk/internal/core-issue/handler"
	issueRepo "github.com/kerbos/ticketdesk/internal/core-issue/repository"
	issueService "github.com/kerbos/ticketdesk/internal/core-issue/service"
	projectHandler "github.com/kerbos/ticketdesk/internal/core-project/handler"
	projectRepo "github.com/kerbos/ticketdesk/internal/core-project/repository"
	projectService "github.com/kerbos/ticketdesk/internal/core-project/service"
	userHandler "github.com/kerbos/ticketdesk/internal/core-user/handler"
	userRepo "github.com/kerbos/ticketdesk/internal/core-user/repository"
	userService "github.com/kerbos/ticketdesk/internal/core-user/service"
	workflowHandler "github.com/kerbos/ticketdesk/internal/core-workflow/handler"
	workflowRepo "github.com/kerbos/ticketdesk/internal/core-workflow/repository"
	workflowService "github.com/kerbos/ticketdesk/internal/core-workflow/service"
	configHandler "github.com/kerbos/ticketdesk/internal/system-config/handler"
	configRepo "github.com/kerbos/ticketdesk/internal/system-config/repository"
	configService "github.com/kerbos/ticketdesk/internal/system-config/service"
	"github.com/kerbos/ticketdesk/pkg/config"
	"github.com/kerbos/ticketdesk/pkg/jwt"
	"gorm.io/gorm"
)

// Router 路由管理器
type Router struct {
	config          *config.Config
	jwtManager      *jwt.Manager
	db              *gorm.DB
	userHandler     *userHandler.UserHandler
	projectHandler  *projectHandler.ProjectHandler
	issueHandler    *issueHandler.IssueHandler
	workflowHandler *workflowHandler.WorkflowHandler
	alertHandler    *alertHandler.AlertHandler
	configHandler   *configHandler.ConfigHandler
	rbac            *middleware.RBACMiddleware
}

// NewRouter 创建路由管理器
func NewRouter(cfg *config.Config, jwtManager *jwt.Manager, db *gorm.DB) *Router {
	// ============ 初始化 User 模块 ============
	userRepository := userRepo.NewUserRepository(db)
	userRoleRepository := userRepo.NewUserRoleRepository(db)
	userSvc := userService.NewUserService(userRepository, userRoleRepository, jwtManager)
	mfaSvc := userService.NewMFAService(userRepository)
	userHdl := userHandler.NewUserHandler(userSvc, mfaSvc)

	// ============ 初始化 Project 模块 ============
	projectRepository := projectRepo.NewProjectRepository(db)
	projectMemberRepository := projectRepo.NewProjectMemberRepository(db)
	issueTypeRepository := projectRepo.NewIssueTypeRepository(db)
	projectSvc := projectService.NewProjectService(
		projectRepository,
		projectMemberRepository,
		issueTypeRepository,
		userRepository,
	)
	projectHdl := projectHandler.NewProjectHandler(projectSvc)

	// ============ 初始化 Issue 模块 ============
	issueRepository := issueRepo.NewIssueRepository(db)
	commentRepository := issueRepo.NewCommentRepository(db)
	watcherRepository := issueRepo.NewWatcherRepository(db)
	issueSvc := issueService.NewIssueService(
		issueRepository,
		commentRepository,
		watcherRepository,
		projectRepository,
		issueTypeRepository,
		userRepository,
	)
	issueHdl := issueHandler.NewIssueHandler(issueSvc)

	// ============ 初始化 Workflow 模块 ============
	workflowRepository := workflowRepo.NewWorkflowRepository(db)
	nodeRepository := workflowRepo.NewNodeRepository(db)
	edgeRepository := workflowRepo.NewEdgeRepository(db)
	workflowSvc := workflowService.NewWorkflowService(
		workflowRepository,
		nodeRepository,
		edgeRepository,
	)
	workflowHdl := workflowHandler.NewWorkflowHandler(workflowSvc)

	// ============ 初始化 Alert 模块 ============
	alertRepository := alertRepo.NewAlertRepository(db)
	alertRuleRepository := alertRepo.NewAlertRuleRepository(db)
	alertSilenceRepository := alertRepo.NewAlertSilenceRepository(db)
	alertSvc := alertService.NewAlertService(
		alertRepository,
		alertRuleRepository,
		alertSilenceRepository,
		issueRepository,
		projectRepository,
		issueTypeRepository,
		db,
	)
	alertHdl := alertHandler.NewAlertHandler(alertSvc)

	// ============ 设置告警同步服务（避免循环依赖）============
	// 将 issueSvc 转换为具体类型以调用 SetAlertSyncService
	if issueServiceImpl, ok := issueSvc.(interface{ SetAlertSyncService(issueService.AlertSyncService) }); ok {
		issueServiceImpl.SetAlertSyncService(alertSvc)
	}

	// ============ 初始化 System Config 模块 ============
	configRepository := configRepo.NewConfigRepository(db)
	webhookRepository := configRepo.NewWebhookRepository(db)
	webhookLogRepository := configRepo.NewWebhookLogRepository(db)
	configSvc := configService.NewConfigService(
		configRepository,
		webhookRepository,
		webhookLogRepository,
	)
	configHdl := configHandler.NewConfigHandler(configSvc)

	// ============ 初始化 RBAC 中间件 ============
	rbac := middleware.NewRBACMiddleware(userRoleRepository)

	return &Router{
		config:          cfg,
		jwtManager:      jwtManager,
		db:              db,
		userHandler:     userHdl,
		projectHandler:  projectHdl,
		issueHandler:    issueHdl,
		workflowHandler: workflowHdl,
		alertHandler:    alertHdl,
		configHandler:   configHdl,
		rbac:            rbac,
	}
}

// Setup 设置路由
func (r *Router) Setup() *gin.Engine {
	// 设置 Gin 模式
	if r.config.IsProduction() {
		gin.SetMode(gin.ReleaseMode)
	}

	engine := gin.New()

	// 全局中间件
	engine.Use(middleware.RecoveryMiddleware())
	engine.Use(middleware.LoggerMiddleware())
	engine.Use(middleware.CORSMiddleware())

	// 健康检查
	engine.GET("/health", healthCheck)

	// API v1 路由组
	v1 := engine.Group("/api/v1")
	{
		// 注册公开路由
		r.registerPublicRoutes(v1)

		// 注册需要认证的路由
		r.registerProtectedRoutes(v1)
	}

	return engine
}

// healthCheck 健康检查处理器
func healthCheck(c *gin.Context) {
	response.Success(c, gin.H{
		"status": "ok",
		"time":   time.Now().Format("2006-01-02 15:04:05-07:00"),
	})
}

// registerPublicRoutes 注册公开路由（无需认证）
func (r *Router) registerPublicRoutes(rg *gin.RouterGroup) {
	auth := rg.Group("/auth")
	{
		auth.POST("/login", r.userHandler.HandleLogin)
		auth.POST("/register", r.userHandler.HandleRegister)
		auth.POST("/refresh", r.userHandler.HandleRefreshToken)
		auth.POST("/mfa/verify", r.userHandler.HandleVerifyMFA)
	}

	// 告警 Webhook（无需认证）
	alerts := rg.Group("/alerts")
	{
		alerts.POST("/webhook", r.alertHandler.HandleWebhook)
	}
}

// registerProtectedRoutes 注册需要认证的路由
func (r *Router) registerProtectedRoutes(rg *gin.RouterGroup) {
	protected := rg.Group("")
	protected.Use(middleware.AuthMiddleware(r.jwtManager))
	protected.Use(r.rbac.LoadUserRoles())
	{
		// 用户相关
		r.registerUserRoutes(protected)

		// 项目相关
		r.registerProjectRoutes(protected)

		// 工单相关
		r.registerIssueRoutes(protected)

		// 工作流相关
		r.registerWorkflowRoutes(protected)

		// 告警相关
		r.registerAlertRoutes(protected)

		// 系统配置相关
		r.registerConfigRoutes(protected)
	}
}

// registerUserRoutes 注册用户路由
func (r *Router) registerUserRoutes(rg *gin.RouterGroup) {
	users := rg.Group("/users")
	{
		// 当前用户相关
		users.GET("/me", r.userHandler.HandleGetCurrentUser)
		users.PUT("/me/password", r.userHandler.HandleUpdatePassword)

		// MFA 相关
		users.GET("/me/mfa", r.userHandler.HandleGetMFAStatus)
		users.POST("/me/mfa/setup", r.userHandler.HandleSetupMFA)
		users.POST("/me/mfa/enable", r.userHandler.HandleEnableMFA)
		users.POST("/me/mfa/disable", r.userHandler.HandleDisableMFA)

		// 获取所有用户（用于选择器，必须在 /:id 之前）
		users.GET("/all", r.userHandler.HandleListAllUsers)

		// 用户管理（需要管理员权限）
		users.GET("", r.userHandler.HandleListUsers)
		users.GET("/:id", r.userHandler.HandleGetUser)
		users.PUT("/:id", r.rbac.RequireAdmin(), r.userHandler.HandleUpdateUser)
		users.DELETE("/:id", r.rbac.RequireAdmin(), r.userHandler.HandleDeleteUser)
	}
}

// registerProjectRoutes 注册项目路由
func (r *Router) registerProjectRoutes(rg *gin.RouterGroup) {
	projects := rg.Group("/projects")
	{
		// 获取所有项目（用于选择器，必须在 /:key 之前）
		projects.GET("/all", r.projectHandler.HandleListAllProjects)

		// 项目 CRUD
		projects.GET("", r.projectHandler.HandleListProjects)
		projects.POST("", r.rbac.RequireProjectAdmin(), r.projectHandler.HandleCreateProject)
		projects.GET("/:key", r.projectHandler.HandleGetProject)
		projects.PUT("/:key", r.rbac.RequireProjectAdmin(), r.projectHandler.HandleUpdateProject)
		projects.DELETE("/:key", r.rbac.RequireAdmin(), r.projectHandler.HandleDeleteProject)

		// 项目成员管理
		projects.GET("/:key/members", r.projectHandler.HandleListMembers)
		projects.POST("/:key/members", r.rbac.RequireProjectAdmin(), r.projectHandler.HandleAddMember)
		projects.PUT("/:key/members/:user_id", r.rbac.RequireProjectAdmin(), r.projectHandler.HandleUpdateMember)
		projects.DELETE("/:key/members/:user_id", r.rbac.RequireProjectAdmin(), r.projectHandler.HandleRemoveMember)

		// 工单类型管理
		projects.GET("/:key/issue-types", r.projectHandler.HandleListIssueTypes)
		projects.POST("/:key/issue-types", r.rbac.RequireProjectAdmin(), r.projectHandler.HandleCreateIssueType)
		projects.PUT("/:key/issue-types/:id", r.rbac.RequireProjectAdmin(), r.projectHandler.HandleUpdateIssueType)
		projects.DELETE("/:key/issue-types/:id", r.rbac.RequireProjectAdmin(), r.projectHandler.HandleDeleteIssueType)
	}
}

// registerIssueRoutes 注册工单路由
func (r *Router) registerIssueRoutes(rg *gin.RouterGroup) {
	issues := rg.Group("/issues")
	{
		// Dashboard 专用（必须放在 /:key 路由之前，避免被匹配为 key）
		issues.GET("/my-todo", r.issueHandler.HandleListMyTodoIssues)
		issues.GET("/my-created", r.issueHandler.HandleListMyCreatedIssues)

		// 工单 CRUD
		issues.GET("", r.issueHandler.HandleListIssues)
		issues.POST("", r.issueHandler.HandleCreateIssue)
		issues.GET("/:key", r.issueHandler.HandleGetIssue)
		issues.PUT("/:key", r.issueHandler.HandleUpdateIssue)
		issues.DELETE("/:key", r.issueHandler.HandleDeleteIssue)

		// 工单操作
		issues.POST("/:key/transition", r.issueHandler.HandleTransitionIssue)
		issues.POST("/:key/assign", r.issueHandler.HandleAssignIssue)

		// 评论管理
		issues.GET("/:key/comments", r.issueHandler.HandleListComments)
		issues.POST("/:key/comments", r.issueHandler.HandleAddComment)
		issues.DELETE("/:key/comments/:comment_id", r.issueHandler.HandleDeleteComment)

		// 关注人管理
		issues.GET("/:key/watchers", r.issueHandler.HandleListWatchers)
		issues.POST("/:key/watchers", r.issueHandler.HandleAddWatcher)
		issues.DELETE("/:key/watchers/:user_id", r.issueHandler.HandleRemoveWatcher)
	}
}

// registerWorkflowRoutes 注册工作流路由
func (r *Router) registerWorkflowRoutes(rg *gin.RouterGroup) {
	workflows := rg.Group("/workflows")
	{
		// 工作流 CRUD
		workflows.GET("", r.workflowHandler.HandleListWorkflows)
		workflows.POST("", r.rbac.RequireProjectAdmin(), r.workflowHandler.HandleCreateWorkflow)
		workflows.GET("/:id", r.workflowHandler.HandleGetWorkflow)
		workflows.PUT("/:id", r.rbac.RequireProjectAdmin(), r.workflowHandler.HandleUpdateWorkflow)
		workflows.DELETE("/:id", r.rbac.RequireAdmin(), r.workflowHandler.HandleDeleteWorkflow)

		// 节点管理
		workflows.GET("/:id/nodes", r.workflowHandler.HandleListNodes)
		workflows.POST("/:id/nodes", r.rbac.RequireProjectAdmin(), r.workflowHandler.HandleCreateNode)
		workflows.GET("/:id/nodes/:node_id", r.workflowHandler.HandleGetNode)
		workflows.PUT("/:id/nodes/:node_id", r.rbac.RequireProjectAdmin(), r.workflowHandler.HandleUpdateNode)
		workflows.DELETE("/:id/nodes/:node_id", r.rbac.RequireProjectAdmin(), r.workflowHandler.HandleDeleteNode)

		// 边管理
		workflows.GET("/:id/edges", r.workflowHandler.HandleListEdges)
		workflows.POST("/:id/edges", r.rbac.RequireProjectAdmin(), r.workflowHandler.HandleCreateEdge)
		workflows.GET("/:id/edges/:edge_id", r.workflowHandler.HandleGetEdge)
		workflows.PUT("/:id/edges/:edge_id", r.rbac.RequireProjectAdmin(), r.workflowHandler.HandleUpdateEdge)
		workflows.DELETE("/:id/edges/:edge_id", r.rbac.RequireProjectAdmin(), r.workflowHandler.HandleDeleteEdge)
	}
}

// registerAlertRoutes 注册告警路由
func (r *Router) registerAlertRoutes(rg *gin.RouterGroup) {
	alerts := rg.Group("/alerts")
	{
		// 告警查询
		alerts.GET("", r.alertHandler.HandleListAlerts)
		alerts.GET("/:id", r.alertHandler.HandleGetAlert)
		alerts.GET("/group", r.alertHandler.HandleGroupAlerts)

		// 告警操作
		alerts.POST("/:id/ack", r.alertHandler.HandleAckAlert)
		alerts.POST("/:id/resolve", r.alertHandler.HandleResolveAlert)

		// Webhook（无需认证，但需要在公开路由中注册）
		// 这里暂时放在受保护路由中，实际使用时可能需要移到公开路由
	}

	// 告警规则管理
	alertRules := rg.Group("/alert-rules")
	{
		alertRules.GET("", r.alertHandler.HandleListAlertRules)
		alertRules.POST("", r.rbac.RequireProjectAdmin(), r.alertHandler.HandleCreateAlertRule)
		alertRules.GET("/:id", r.alertHandler.HandleGetAlertRule)
		alertRules.PUT("/:id", r.rbac.RequireProjectAdmin(), r.alertHandler.HandleUpdateAlertRule)
		alertRules.DELETE("/:id", r.rbac.RequireAdmin(), r.alertHandler.HandleDeleteAlertRule)
	}

	// 告警静默管理
	alertSilences := rg.Group("/alert-silences")
	{
		alertSilences.GET("", r.alertHandler.HandleListAlertSilences)
		alertSilences.POST("", r.alertHandler.HandleCreateAlertSilence)
		alertSilences.GET("/:id", r.alertHandler.HandleGetAlertSilence)
		alertSilences.PUT("/:id", r.alertHandler.HandleUpdateAlertSilence)
		alertSilences.DELETE("/:id", r.rbac.RequireAdmin(), r.alertHandler.HandleDeleteAlertSilence)
		alertSilences.POST("/:id/cancel", r.alertHandler.HandleCancelAlertSilence)
	}
}

// registerConfigRoutes 注册系统配置路由
func (r *Router) registerConfigRoutes(rg *gin.RouterGroup) {
	// 系统配置（需要管理员权限）
	configs := rg.Group("/system/configs")
	configs.Use(r.rbac.RequireAdmin())
	{
		configs.GET("", r.configHandler.HandleGetAllConfigs)
		configs.GET("/category", r.configHandler.HandleGetConfigsByCategory)
		configs.GET("/:key", r.configHandler.HandleGetConfig)
		configs.PUT("/:key", r.configHandler.HandleUpdateConfig)
		configs.PUT("", r.configHandler.HandleBatchUpdateConfigs)
	}

	// 邮件配置（需要管理员权限）
	email := rg.Group("/system/email")
	email.Use(r.rbac.RequireAdmin())
	{
		email.GET("", r.configHandler.HandleGetEmailConfig)
		email.PUT("", r.configHandler.HandleUpdateEmailConfig)
	}

	// 安全配置（需要管理员权限）
	security := rg.Group("/system/security")
	security.Use(r.rbac.RequireAdmin())
	{
		security.GET("", r.configHandler.HandleGetSecurityConfig)
		security.PUT("", r.configHandler.HandleUpdateSecurityConfig)
	}

	// Webhook 管理（需要管理员权限）
	webhooks := rg.Group("/system/webhooks")
	webhooks.Use(r.rbac.RequireAdmin())
	{
		webhooks.GET("", r.configHandler.HandleListWebhooks)
		webhooks.POST("", r.configHandler.HandleCreateWebhook)
		webhooks.GET("/:id", r.configHandler.HandleGetWebhook)
		webhooks.PUT("/:id", r.configHandler.HandleUpdateWebhook)
		webhooks.DELETE("/:id", r.configHandler.HandleDeleteWebhook)
	}

	// Webhook 日志（需要管理员权限）
	webhookLogs := rg.Group("/system/webhook-logs")
	webhookLogs.Use(r.rbac.RequireAdmin())
	{
		webhookLogs.GET("", r.configHandler.HandleListWebhookLogs)
	}
}

// ============ 兼容旧的 Setup 函数 ============

// Setup 设置路由（兼容旧接口）
func Setup(cfg *config.Config, jwtManager *jwt.Manager) *gin.Engine {
	panic("请使用 NewRouter(cfg, jwtManager, db).Setup() 方式初始化路由")
}
