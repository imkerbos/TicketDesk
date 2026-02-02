// Package router 提供路由配置
package router

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/kerbos/ticketdesk/internal/api/middleware"
	"github.com/kerbos/ticketdesk/internal/api/response"
	"github.com/kerbos/ticketdesk/pkg/config"
	"github.com/kerbos/ticketdesk/pkg/jwt"
)

// Setup 设置路由
func Setup(cfg *config.Config, jwtManager *jwt.Manager) *gin.Engine {
	// 设置 Gin 模式
	if cfg.IsProduction() {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.New()

	// 全局中间件
	router.Use(middleware.RecoveryMiddleware())
	router.Use(middleware.LoggerMiddleware())
	router.Use(middleware.CORSMiddleware())

	// 健康检查
	router.GET("/health", healthCheck)

	// API v1 路由组
	v1 := router.Group("/api/v1")
	{
		// 注册公开路由
		registerPublicRoutes(v1)

		// 注册需要认证的路由
		registerProtectedRoutes(v1, jwtManager)
	}

	return router
}

// healthCheck 健康检查处理器
func healthCheck(c *gin.Context) {
	response.Success(c, gin.H{
		"status": "ok",
		"time":   time.Now().Format("2006-01-02 15:04:05-07:00"),
	})
}

// registerPublicRoutes 注册公开路由（无需认证）
func registerPublicRoutes(rg *gin.RouterGroup) {
	auth := rg.Group("/auth")
	{
		auth.POST("/login", handleLogin)
		auth.POST("/register", handleRegister)
		auth.POST("/refresh", handleRefreshToken)
	}
}

// registerProtectedRoutes 注册需要认证的路由
func registerProtectedRoutes(rg *gin.RouterGroup, jwtManager *jwt.Manager) {
	protected := rg.Group("")
	protected.Use(middleware.AuthMiddleware(jwtManager))
	{
		// 用户相关
		registerUserRoutes(protected)

		// 项目相关
		registerProjectRoutes(protected)

		// 工单相关
		registerIssueRoutes(protected)

		// 工作流相关
		registerWorkflowRoutes(protected)

		// 告警相关
		registerAlertRoutes(protected)
	}
}

// registerUserRoutes 注册用户路由
func registerUserRoutes(rg *gin.RouterGroup) {
	users := rg.Group("/users")
	{
		users.GET("/me", handleGetCurrentUser)
		users.GET("", handleListUsers)
		users.GET("/:id", handleGetUser)
		users.PUT("/:id", handleUpdateUser)
		users.DELETE("/:id", handleDeleteUser)
	}
}

// registerProjectRoutes 注册项目路由
func registerProjectRoutes(rg *gin.RouterGroup) {
	projects := rg.Group("/projects")
	{
		projects.GET("", handleListProjects)
		projects.POST("", handleCreateProject)
		projects.GET("/:key", handleGetProject)
		projects.PUT("/:key", handleUpdateProject)
		projects.DELETE("/:key", handleDeleteProject)
		projects.GET("/:key/members", handleListProjectMembers)
		projects.POST("/:key/members", handleAddProjectMember)
		projects.DELETE("/:key/members/:user_id", handleRemoveProjectMember)
	}
}

// registerIssueRoutes 注册工单路由
func registerIssueRoutes(rg *gin.RouterGroup) {
	issues := rg.Group("/issues")
	{
		issues.GET("", handleListIssues)
		issues.POST("", handleCreateIssue)
		issues.GET("/:key", handleGetIssue)
		issues.PUT("/:key", handleUpdateIssue)
		issues.DELETE("/:key", handleDeleteIssue)
		issues.POST("/:key/comments", handleAddComment)
		issues.GET("/:key/comments", handleListComments)
		issues.POST("/:key/watchers", handleAddWatcher)
		issues.DELETE("/:key/watchers/:user_id", handleRemoveWatcher)
		issues.POST("/:key/transition", handleTransitionIssue)
	}
}

// registerWorkflowRoutes 注册工作流路由
func registerWorkflowRoutes(rg *gin.RouterGroup) {
	workflows := rg.Group("/workflows")
	{
		workflows.GET("", handleListWorkflows)
		workflows.POST("", handleCreateWorkflow)
		workflows.GET("/:id", handleGetWorkflow)
		workflows.PUT("/:id", handleUpdateWorkflow)
		workflows.DELETE("/:id", handleDeleteWorkflow)
		workflows.GET("/:id/nodes", handleListWorkflowNodes)
		workflows.POST("/:id/nodes", handleCreateWorkflowNode)
	}
}

// registerAlertRoutes 注册告警路由
func registerAlertRoutes(rg *gin.RouterGroup) {
	alerts := rg.Group("/alerts")
	{
		alerts.GET("", handleListAlerts)
		alerts.POST("/webhook", handleAlertWebhook)
		alerts.GET("/:id", handleGetAlert)
		alerts.POST("/:id/ack", handleAckAlert)
		alerts.POST("/:id/resolve", handleResolveAlert)
	}
}
