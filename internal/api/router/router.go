// Package router 提供路由配置
package router

import (
	"context"
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
	activityHandler "github.com/kerbos/ticketdesk/internal/activity/handler"
	activityRepo "github.com/kerbos/ticketdesk/internal/activity/repository"
	activityService "github.com/kerbos/ticketdesk/internal/activity/service"
	alertHandler "github.com/kerbos/ticketdesk/internal/integration-alert/handler"
	notifDto "github.com/kerbos/ticketdesk/internal/notification-inbox/dto"
	notifHandler "github.com/kerbos/ticketdesk/internal/notification-inbox/handler"
	notifRepo "github.com/kerbos/ticketdesk/internal/notification-inbox/repository"
	notifService "github.com/kerbos/ticketdesk/internal/notification-inbox/service"
	ws "github.com/kerbos/ticketdesk/internal/notification-inbox/websocket"
	emailService "github.com/kerbos/ticketdesk/internal/notification/email"
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
	reportHandler "github.com/kerbos/ticketdesk/internal/reporting/handler"
	reportRepo "github.com/kerbos/ticketdesk/internal/reporting/repository"
	reportService "github.com/kerbos/ticketdesk/internal/reporting/service"
	reqPoolHandler "github.com/kerbos/ticketdesk/internal/requirement-pool/handler"
	reqPoolRepo "github.com/kerbos/ticketdesk/internal/requirement-pool/repository"
	reqPoolService "github.com/kerbos/ticketdesk/internal/requirement-pool/service"
	fieldHandler "github.com/kerbos/ticketdesk/internal/core-field/handler"
	fieldRepo "github.com/kerbos/ticketdesk/internal/core-field/repository"
	fieldService "github.com/kerbos/ticketdesk/internal/core-field/service"
	configHandler "github.com/kerbos/ticketdesk/internal/system-config/handler"
	configRepo "github.com/kerbos/ticketdesk/internal/system-config/repository"
	configService "github.com/kerbos/ticketdesk/internal/system-config/service"
	"github.com/kerbos/ticketdesk/pkg/config"
	"github.com/kerbos/ticketdesk/pkg/jwt"
	"github.com/kerbos/ticketdesk/pkg/storage"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// Router 路由管理器
type Router struct {
	config                 *config.Config
	jwtManager             *jwt.Manager
	db                     *gorm.DB
	logger                 *zap.Logger
	userHandler            *userHandler.UserHandler
	projectHandler         *projectHandler.ProjectHandler
	issueHandler           *issueHandler.IssueHandler
	attachmentHandler      *issueHandler.AttachmentHandler
	workflowHandler        *workflowHandler.WorkflowHandler
	alertHandler           *alertHandler.AlertHandler
	configHandler          *configHandler.ConfigHandler
	reportHandler          *reportHandler.ReportHandler
	activityHandler        *activityHandler.ActivityHandler
	notificationHandler    *notifHandler.NotificationHandler
	wsHandler              *notifHandler.WebSocketHandler
	requirementPoolHandler *reqPoolHandler.RequirementPoolHandler
	requirementHandler     *reqPoolHandler.RequirementHandler
	fieldHandler           *fieldHandler.FieldHandler
	rbac                   *middleware.RBACMiddleware
	datasourceHandler      *alertHandler.DatasourceHandler
	datasourceService      alertService.DatasourceService
	notifChannelHandler    *projectHandler.NotificationChannelHandler
}

// NewRouter 创建路由管理器
func NewRouter(cfg *config.Config, jwtManager *jwt.Manager, db *gorm.DB) *Router {
	// ============ 初始化 System Config 模块（需要先初始化，因为邮件服务依赖它）============
	configRepository := configRepo.NewConfigRepository(db)
	webhookRepository := configRepo.NewWebhookRepository(db)
	webhookLogRepository := configRepo.NewWebhookLogRepository(db)
	configSvc := configService.NewConfigService(
		configRepository,
		webhookRepository,
		webhookLogRepository,
	)
	configHdl := configHandler.NewConfigHandler(configSvc)

	// ============ 初始化邮件服务 ============
	emailSvc := emailService.NewEmailService(configSvc)

	// ============ 初始化 User 模块 ============
	userRepository := userRepo.NewUserRepository(db)
	userRoleRepository := userRepo.NewUserRoleRepository(db)
	userSvc := userService.NewUserService(userRepository, userRoleRepository, jwtManager, emailSvc, configSvc)
	mfaSvc := userService.NewMFAService(userRepository)
	userHdl := userHandler.NewUserHandler(userSvc, mfaSvc)

	// ============ 初始化 Project 模块 ============
	projectRepository := projectRepo.NewProjectRepository(db)
	projectMemberRepository := projectRepo.NewProjectMemberRepository(db)
	issueTypeRepository := projectRepo.NewIssueTypeRepository(db)
	projectRoleRepository := projectRepo.NewProjectRoleRepository(db)
	projectRoleMemberRepository := projectRepo.NewProjectRoleMemberRepository(db)
	projectSvc := projectService.NewProjectService(
		projectRepository,
		projectMemberRepository,
		issueTypeRepository,
		projectRoleRepository,
		projectRoleMemberRepository,
		userRepository,
		db,
	)
	projectHdl := projectHandler.NewProjectHandler(projectSvc)

	// ============ 初始化 Issue 模块 ============
	issueRepository := issueRepo.NewIssueRepository(db)
	commentRepository := issueRepo.NewCommentRepository(db)
	watcherRepository := issueRepo.NewWatcherRepository(db)
	worklogRepository := issueRepo.NewWorklogRepository(db)
	issueSvc := issueService.NewIssueService(
		issueRepository,
		commentRepository,
		watcherRepository,
		worklogRepository,
		projectRepository,
		issueTypeRepository,
		userRepository,
	)
	issueHdl := issueHandler.NewIssueHandler(issueSvc)

	// ============ 初始化 Attachment 模块 ============
	attachmentRepository := issueRepo.NewAttachmentRepository(db)
	localStorage, err := storage.NewLocalStorage("./uploads")
	if err != nil {
		panic(fmt.Sprintf("failed to initialize local storage: %v", err))
	}
	attachmentSvc := issueService.NewAttachmentService(
		attachmentRepository,
		issueRepository,
		userRepository,
		localStorage,
	)
	attachmentHdl := issueHandler.NewAttachmentHandler(attachmentSvc)

	// ============ 初始化 Workflow 模块 ============
	workflowRepository := workflowRepo.NewWorkflowRepository(db)
	nodeRepository := workflowRepo.NewNodeRepository(db)
	edgeRepository := workflowRepo.NewEdgeRepository(db)
	workflowInstanceRepository := workflowRepo.NewWorkflowInstanceRepository(db)
	workflowHistoryRepository := workflowRepo.NewWorkflowHistoryRepository(db)
	approvalRecordRepository := workflowRepo.NewApprovalRecordRepository(db)
	workflowSchemeRepository := workflowRepo.NewWorkflowSchemeRepository(db)

	workflowSvc := workflowService.NewWorkflowService(
		workflowRepository,
		nodeRepository,
		edgeRepository,
		workflowSchemeRepository,
		issueTypeRepository,
	)

	workflowEngine := workflowService.NewWorkflowEngine(
		workflowInstanceRepository,
		workflowHistoryRepository,
		approvalRecordRepository,
		workflowRepository,
		nodeRepository,
		edgeRepository,
		workflowSchemeRepository,
		projectRoleRepository,
		userRepository,
		db,
	)

	// ============ 初始化 Activity 模块（提前初始化，因为 WorkflowHandler 依赖它）============
	activityRepository := activityRepo.NewActivityRepository(db)
	activitySvc := activityService.NewActivityService(activityRepository)
	activityHdl := activityHandler.NewActivityHandler(activitySvc)

	workflowHdl := workflowHandler.NewWorkflowHandler(workflowSvc, workflowEngine, issueRepository, projectRepository, activitySvc)

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

	// ============ 初始化 Report 模块 ============
	reportRepository := reportRepo.NewReportRepository(db)
	reportSvc := reportService.NewReportService(reportRepository, projectRepository)
	reportHdl := reportHandler.NewReportHandler(reportSvc)

	// ============ 设置活动日志记录器（避免循环依赖）============
	if issueServiceImpl, ok := issueSvc.(interface{ SetActivityLogger(issueService.ActivityLogger) }); ok {
		issueServiceImpl.SetActivityLogger(activitySvc)
	}

	// 设置附件服务的活动日志记录器
	if attachmentServiceImpl, ok := attachmentSvc.(interface{ SetActivityLogger(issueService.ActivityLogger) }); ok {
		attachmentServiceImpl.SetActivityLogger(activitySvc)
	}

	// ============ 初始化 Notification 模块 ============
	wsManager := ws.NewManager()
	go wsManager.Run()

	notificationRepository := notifRepo.NewNotificationRepository(db)
	notificationSvc := notifService.NewNotificationService(notificationRepository, wsManager)
	notificationHdl := notifHandler.NewNotificationHandler(notificationSvc)
	wsHdl := notifHandler.NewWebSocketHandler(wsManager, jwtManager)

	// ============ 设置通知服务（避免循环依赖）============
	notifAdapter := &notificationAdapter{svc: notificationSvc}
	if issueServiceImpl, ok := issueSvc.(interface {
		SetNotificationService(issueService.NotificationSender)
	}); ok {
		issueServiceImpl.SetNotificationService(notifAdapter)
	}

	// ============ 设置工作流引擎（避免循环依赖）============
	if issueServiceImpl, ok := issueSvc.(interface{ SetWorkflowEngine(issueService.WorkflowEngine) }); ok {
		issueServiceImpl.SetWorkflowEngine(workflowEngine)
	}

	// ============ 初始化需求池模块 ============
	// 初始化 logger（如果还没有的话）
	logger, _ := zap.NewProduction()

	reqPoolRepository := reqPoolRepo.NewRequirementPoolRepository(db)
	reqRepository := reqPoolRepo.NewRequirementRepository(db)

	reqPoolSvc := reqPoolService.NewRequirementPoolService(
		reqPoolRepository,
		reqRepository,
		logger,
	)
	reqSvc := reqPoolService.NewRequirementService(
		reqRepository,
		reqPoolRepository,
		db,
		logger,
	)

	reqPoolHdl := reqPoolHandler.NewRequirementPoolHandler(reqPoolSvc, logger)
	reqHdl := reqPoolHandler.NewRequirementHandler(reqSvc, logger)

	// ============ 初始化 Field 模块 ============
	fieldRepository := fieldRepo.NewFieldRepository(db)
	schemeRepository := fieldRepo.NewSchemeRepository(db)
	valueRepository := fieldRepo.NewValueRepository(db)
	versionRepository := fieldRepo.NewVersionRepository(db)
	componentRepository := fieldRepo.NewComponentRepository(db)
	labelRepository := fieldRepo.NewLabelRepository(db)
	fieldSvc := fieldService.NewFieldService(
		fieldRepository,
		schemeRepository,
		valueRepository,
		versionRepository,
		componentRepository,
		labelRepository,
		projectRepository,
		issueTypeRepository,
		userRepository,
		db,
	)
	fieldHdl := fieldHandler.NewFieldHandler(fieldSvc)

	// ============ 设置字段值保存服务（避免循环依赖）============
	fieldValueAdapter := &fieldValueSaverAdapter{svc: fieldSvc}
	if issueServiceImpl, ok := issueSvc.(interface {
		SetFieldValueSaver(issueService.FieldValueSaver)
	}); ok {
		issueServiceImpl.SetFieldValueSaver(fieldValueAdapter)
	}

	// 设置 Epic 链接获取服务
	if issueServiceImpl, ok := issueSvc.(interface {
		SetEpicLinkGetter(issueService.EpicLinkGetter)
	}); ok {
		issueServiceImpl.SetEpicLinkGetter(fieldSvc)
	}

	// ============ 初始化项目通知渠道模块 ============
	notifChannelRepo := projectRepo.NewNotificationChannelRepository(db)
	notifChannelSvc := projectService.NewNotificationChannelService(notifChannelRepo, projectRepository)
	notifChannelHdl := projectHandler.NewNotificationChannelHandler(notifChannelSvc, projectSvc)

	// ============ 设置项目通知服务（避免循环依赖）============
	if issueServiceImpl, ok := issueSvc.(interface {
		SetProjectNotifier(issueService.ProjectNotifier)
	}); ok {
		issueServiceImpl.SetProjectNotifier(notifChannelSvc)
	}

	// ============ 初始化 RBAC 中间件 ============
	rbac := middleware.NewRBACMiddleware(userRoleRepository)

	// ============ 初始化数据源服务 ============
	datasourceRepository := alertRepo.NewAlertDatasourceRepository(db)
	datasourceSvc := alertService.NewDatasourceService(
		datasourceRepository,
		alertSvc,
		alertRepository,
		&cfg.Nightingale,
	)
	datasourceHdl := alertHandler.NewDatasourceHandler(datasourceSvc, alertSvc)

	// 启动所有数据源 Poller
	if err := datasourceSvc.StartAllPollers(context.Background()); err != nil {
		logger.Error("failed to start datasource pollers", zap.Error(err))
	}

	return &Router{
		config:                 cfg,
		jwtManager:             jwtManager,
		db:                     db,
		logger:                 logger,
		userHandler:            userHdl,
		projectHandler:         projectHdl,
		issueHandler:           issueHdl,
		attachmentHandler:      attachmentHdl,
		workflowHandler:        workflowHdl,
		alertHandler:           alertHdl,
		configHandler:          configHdl,
		reportHandler:          reportHdl,
		activityHandler:        activityHdl,
		notificationHandler:    notificationHdl,
		wsHandler:              wsHdl,
		requirementPoolHandler: reqPoolHdl,
		requirementHandler:     reqHdl,
		fieldHandler:           fieldHdl,
		rbac:                   rbac,
		datasourceHandler:      datasourceHdl,
		datasourceService:      datasourceSvc,
		notifChannelHandler:    notifChannelHdl,
	}
}

// StopPollers 停止所有数据源轮询器（用于优雅关闭）
func (r *Router) StopPollers() {
	if r.datasourceService != nil {
		r.datasourceService.StopAllPollers()
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
		// WebSocket 连接（使用 query 参数认证，不走中间件）
		v1.GET("/ws", r.wsHandler.HandleWebSocket)

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
		// 忘记密码相关路由
		auth.POST("/forgot-password", r.userHandler.HandleForgotPassword)
		auth.GET("/verify-reset-token", r.userHandler.HandleVerifyResetToken)
		auth.POST("/reset-password", r.userHandler.HandleResetPasswordWithToken)
	}

	// 告警 Webhook（无需认证）
	alerts := rg.Group("/alerts")
	{
		alerts.POST("/webhook", r.alertHandler.HandleWebhook)
		alerts.POST("/nightingale", r.alertHandler.HandleNightingaleWebhook)
		alerts.POST("/datasource/:name/webhook", r.datasourceHandler.HandleDatasourceWebhook)
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

		// 报表相关
		r.registerReportRoutes(protected)

		// 活动日志相关
		r.registerActivityRoutes(protected)

		// 系统配置相关
		r.registerConfigRoutes(protected)

		// 通知相关
		r.registerNotificationRoutes(protected)

		// 需求池相关
		r.registerRequirementPoolRoutes(protected)

		// 字段配置相关
		r.registerFieldRoutes(protected)
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
		users.POST("", r.rbac.RequireAdmin(), r.userHandler.HandleCreateUser)
		users.GET("/:id", r.userHandler.HandleGetUser)
		users.PUT("/:id", r.rbac.RequireAdmin(), r.userHandler.HandleUpdateUser)
		users.POST("/:id/enable", r.rbac.RequireAdmin(), r.userHandler.HandleEnableUser)
		users.POST("/:id/disable", r.rbac.RequireAdmin(), r.userHandler.HandleDisableUser)
		users.POST("/:id/reset-password", r.rbac.RequireAdmin(), r.userHandler.HandleResetPassword)
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

		// 项目角色管理
		projects.GET("/:key/roles", r.projectHandler.HandleListRoles)
		projects.POST("/:key/roles", r.rbac.RequireProjectAdmin(), r.projectHandler.HandleCreateRole)
		projects.PUT("/:key/roles/:id", r.rbac.RequireProjectAdmin(), r.projectHandler.HandleUpdateRole)
		projects.DELETE("/:key/roles/:id", r.rbac.RequireProjectAdmin(), r.projectHandler.HandleDeleteRole)

		// 角色成员管理
		projects.GET("/:key/roles/:id/members", r.projectHandler.HandleListRoleMembers)
		projects.POST("/:key/roles/:id/members", r.rbac.RequireProjectAdmin(), r.projectHandler.HandleAddRoleMember)
		projects.DELETE("/:key/roles/:id/members/:user_id", r.rbac.RequireProjectAdmin(), r.projectHandler.HandleRemoveRoleMember)

		// 用户角色查询
		projects.GET("/:key/users/:user_id/roles", r.projectHandler.HandleGetUserRoles)

		// 通知渠道管理
		notifChannels := projects.Group("/:key/notification-channels")
		{
			notifChannels.GET("", r.notifChannelHandler.HandleListChannels)
			notifChannels.POST("", r.rbac.RequireProjectAdmin(), r.notifChannelHandler.HandleCreateChannel)
			notifChannels.GET("/:id", r.notifChannelHandler.HandleGetChannel)
			notifChannels.PUT("/:id", r.rbac.RequireProjectAdmin(), r.notifChannelHandler.HandleUpdateChannel)
			notifChannels.DELETE("/:id", r.rbac.RequireProjectAdmin(), r.notifChannelHandler.HandleDeleteChannel)
			notifChannels.POST("/:id/test", r.rbac.RequireProjectAdmin(), r.notifChannelHandler.HandleTestChannel)
		}

		// 工作流方案管理
		projects.GET("/:key/workflow-schemes", r.workflowHandler.HandleListSchemes)
		projects.POST("/:key/workflow-schemes", r.rbac.RequireProjectAdmin(), r.workflowHandler.HandleCreateScheme)
		projects.DELETE("/:key/workflow-schemes/:type_id", r.rbac.RequireProjectAdmin(), r.workflowHandler.HandleDeleteScheme)
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

		issues.POST("/:key/assign", r.issueHandler.HandleAssignIssue)

		// Epic 相关
		issues.GET("/:key/epic-issues", r.issueHandler.HandleListIssuesInEpic)

		// 子任务相关
		issues.GET("/:key/subtasks", r.issueHandler.HandleListSubtasks)

		// 评论管理
		issues.GET("/:key/comments", r.issueHandler.HandleListComments)
		issues.POST("/:key/comments", r.issueHandler.HandleAddComment)
		issues.DELETE("/:key/comments/:comment_id", r.issueHandler.HandleDeleteComment)

		// 关注人管理
		issues.GET("/:key/watchers", r.issueHandler.HandleListWatchers)
		issues.POST("/:key/watchers", r.issueHandler.HandleAddWatcher)
		issues.DELETE("/:key/watchers/:user_id", r.issueHandler.HandleRemoveWatcher)

		// 工作日志管理
		issues.GET("/:key/worklogs", r.issueHandler.HandleListWorklogs)
		issues.POST("/:key/worklogs", r.issueHandler.HandleAddWorklog)
		issues.PUT("/:key/worklogs/:worklog_id", r.issueHandler.HandleUpdateWorklog)
		issues.DELETE("/:key/worklogs/:worklog_id", r.issueHandler.HandleDeleteWorklog)

		// 附件管理
		issues.POST("/:key/attachments", r.attachmentHandler.HandleUploadAttachment)
		issues.GET("/:key/attachments", r.attachmentHandler.HandleListAttachments)
		issues.DELETE("/:key/attachments/:id", r.attachmentHandler.HandleDeleteAttachment)
		issues.GET("/:key/attachments/:id/download", r.attachmentHandler.HandleDownloadAttachment)

		// 工作流实例（通过工单 key 访问）
		issues.GET("/:key/workflow", r.workflowHandler.HandleGetInstanceByIssue)
		issues.POST("/:key/workflow/approve", r.workflowHandler.HandleApprove)
		issues.POST("/:key/workflow/reject", r.workflowHandler.HandleReject)
		issues.POST("/:key/workflow/complete", r.workflowHandler.HandleComplete)
		issues.GET("/:key/workflow/history", r.workflowHandler.HandleGetHistory)
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
		alerts.GET("/stats", r.alertHandler.HandleGetAlertStats)
		alerts.GET("/group", r.alertHandler.HandleGroupAlerts)
		alerts.GET("/:id", r.alertHandler.HandleGetAlert)

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

	// 告警数据源管理（需要管理员权限）
	alertDatasources := rg.Group("/alert-datasources")
	alertDatasources.Use(r.rbac.RequireAdmin())
	{
		alertDatasources.GET("", r.datasourceHandler.HandleListDatasources)
		alertDatasources.POST("", r.datasourceHandler.HandleCreateDatasource)
		alertDatasources.GET("/:id", r.datasourceHandler.HandleGetDatasource)
		alertDatasources.PUT("/:id", r.datasourceHandler.HandleUpdateDatasource)
		alertDatasources.DELETE("/:id", r.datasourceHandler.HandleDeleteDatasource)
		alertDatasources.POST("/test", r.datasourceHandler.HandleTestConnection)
		alertDatasources.POST("/:id/test", r.datasourceHandler.HandleTestConnectionByID)
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

// registerReportRoutes 注册报表路由
func (r *Router) registerReportRoutes(rg *gin.RouterGroup) {
	reports := rg.Group("/reports")
	{
		reports.GET("/dashboard", r.reportHandler.HandleGetDashboardStats)
		reports.GET("/issues", r.reportHandler.HandleGetIssueStats)
		reports.GET("/sla", r.reportHandler.HandleGetSLAReport)
		reports.GET("/alerts", r.reportHandler.HandleGetAlertStats)
		reports.GET("/user-performance", r.reportHandler.HandleGetUserPerformance)
	}
}

// registerActivityRoutes 注册活动日志路由
func (r *Router) registerActivityRoutes(rg *gin.RouterGroup) {
	activities := rg.Group("/activities")
	{
		activities.GET("", r.activityHandler.HandleListActivities)
		activities.GET("/recent", r.activityHandler.HandleGetRecentActivities)
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

// registerNotificationRoutes 注册通知路由
func (r *Router) registerNotificationRoutes(rg *gin.RouterGroup) {
	notifications := rg.Group("/notifications")
	{
		notifications.GET("", r.notificationHandler.HandleListNotifications)
		notifications.GET("/unread-count", r.notificationHandler.HandleGetUnreadCount)
		notifications.PUT("/:id/read", r.notificationHandler.HandleMarkAsRead)
		notifications.PUT("/read-all", r.notificationHandler.HandleMarkAllAsRead)
		notifications.DELETE("/:id", r.notificationHandler.HandleDeleteNotification)
	}
}

// ============ 通知适配器（桥接 issue_service.NotificationSender 和 notifService.NotificationService）============

// notificationAdapter 将 NotificationService 适配为 issue_service.NotificationSender
type notificationAdapter struct {
	svc notifService.NotificationService
}

// CreateNotification 适配创建通知调用
func (a *notificationAdapter) CreateNotification(ctx context.Context, req *issueService.NotificationRequest) error {
	return a.svc.CreateNotification(ctx, &notifDto.CreateNotificationRequest{
		UserID:     req.UserID,
		Type:       req.Type,
		Title:      req.Title,
		Content:    req.Content,
		EntityType: req.EntityType,
		EntityID:   req.EntityID,
		EntityKey:  req.EntityKey,
		ActorID:    req.ActorID,
		ActorName:  req.ActorName,
	})
}

// ============ 字段值保存适配器（桥接 issue_service.FieldValueSaver 和 fieldService.FieldService）============

// fieldValueSaverAdapter 将 FieldService 适配为 issue_service.FieldValueSaver
type fieldValueSaverAdapter struct {
	svc fieldService.FieldService
}

// SaveIssueFieldValues 适配保存字段值调用
func (a *fieldValueSaverAdapter) SaveIssueFieldValues(ctx context.Context, issueID uint64, values []issueService.CustomFieldValueInput) error {
	// 转换类型
	fieldValues := make([]fieldService.CustomFieldValueInput, len(values))
	for i, v := range values {
		fieldValues[i] = fieldService.CustomFieldValueInput{
			FieldID: v.FieldID,
			Value:   v.Value,
		}
	}
	return a.svc.SaveIssueFieldValues(ctx, issueID, fieldValues)
}

// ============ 兼容旧的 Setup 函数 ============

// Setup 设置路由（兼容旧接口）
func Setup(cfg *config.Config, jwtManager *jwt.Manager) *gin.Engine {
	panic("请使用 NewRouter(cfg, jwtManager, db).Setup() 方式初始化路由")
}

// registerFieldRoutes 注册字段配置路由
func (r *Router) registerFieldRoutes(rg *gin.RouterGroup) {
	// 字段定义（在项目下）
	projectFields := rg.Group("/projects/:key/fields")
	{
		projectFields.GET("", r.fieldHandler.ListFields)
		projectFields.POST("", r.rbac.RequireProjectAdmin(), r.fieldHandler.CreateField)
		projectFields.PUT("/:id", r.rbac.RequireProjectAdmin(), r.fieldHandler.UpdateField)
		projectFields.DELETE("/:id", r.rbac.RequireProjectAdmin(), r.fieldHandler.DeleteField)
	}

	// 字段方案（在项目的工单类型下）
	fieldScheme := rg.Group("/projects/:key/issue-types/:id/field-scheme")
	{
		fieldScheme.GET("", r.fieldHandler.GetFieldScheme)
		fieldScheme.PUT("", r.rbac.RequireProjectAdmin(), r.fieldHandler.UpdateFieldScheme)
	}

	// 版本管理
	versions := rg.Group("/projects/:key/versions")
	{
		versions.GET("", r.fieldHandler.ListVersions)
		versions.POST("", r.rbac.RequireProjectAdmin(), r.fieldHandler.CreateVersion)
		versions.PUT("/:id", r.rbac.RequireProjectAdmin(), r.fieldHandler.UpdateVersion)
		versions.DELETE("/:id", r.rbac.RequireProjectAdmin(), r.fieldHandler.DeleteVersion)
	}

	// 组件管理
	components := rg.Group("/projects/:key/components")
	{
		components.GET("", r.fieldHandler.ListComponents)
		components.POST("", r.rbac.RequireProjectAdmin(), r.fieldHandler.CreateComponent)
		components.PUT("/:id", r.rbac.RequireProjectAdmin(), r.fieldHandler.UpdateComponent)
		components.DELETE("/:id", r.rbac.RequireProjectAdmin(), r.fieldHandler.DeleteComponent)
	}

	// 标签管理
	labels := rg.Group("/projects/:key/labels")
	{
		labels.GET("", r.fieldHandler.ListLabels)
		labels.POST("", r.rbac.RequireProjectAdmin(), r.fieldHandler.CreateLabel)
		labels.PUT("/:id", r.rbac.RequireProjectAdmin(), r.fieldHandler.UpdateLabel)
		labels.DELETE("/:id", r.rbac.RequireProjectAdmin(), r.fieldHandler.DeleteLabel)
	}

	// 工单字段值
	rg.GET("/issue-field-values/:issue_id", r.fieldHandler.GetIssueFieldValues)
}

// registerRequirementPoolRoutes 注册需求池路由
func (r *Router) registerRequirementPoolRoutes(rg *gin.RouterGroup) {
	// 需求池管理
	pools := rg.Group("/requirement-pools")
	{
		pools.GET("", r.requirementPoolHandler.List)
		pools.POST("", r.requirementPoolHandler.Create)
		pools.GET("/:id", r.requirementPoolHandler.GetByID)
		pools.PUT("/:id", r.requirementPoolHandler.Update)
		pools.DELETE("/:id", r.requirementPoolHandler.Delete)
	}

	// 需求管理
	requirements := rg.Group("/requirements")
	{
		// 看板和报告（必须在 /:id 之前）
		requirements.GET("/kanban", r.requirementHandler.GetKanban)
		requirements.GET("/report", r.requirementHandler.GetReport)

		// 需求 CRUD
		requirements.GET("", r.requirementHandler.List)
		requirements.POST("", r.requirementHandler.Create)
		requirements.GET("/:id", r.requirementHandler.GetByID)
		requirements.PUT("/:id", r.requirementHandler.Update)
		requirements.DELETE("/:id", r.requirementHandler.Delete)

		// 需求操作
		requirements.POST("/:id/convert", r.requirementHandler.ConvertToIssue)
		requirements.POST("/:id/comments", r.requirementHandler.AddComment)
	}
}
