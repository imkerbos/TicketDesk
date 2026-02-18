// Package model 提供数据库迁移功能
package model

import (
	"time"

	"github.com/kerbos/ticketdesk/pkg/logger"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// AutoMigrate 自动迁移数据库表结构
func AutoMigrate(db *gorm.DB) error {
	logger.Info("starting database auto migration...")

	models := []interface{}{
		&User{},
		&Role{},
		&UserRole{},
		&Project{},
		&ProjectMember{},
		&ProjectRole{},
		&ProjectRoleMember{},
		&IssueType{},
		&Issue{},
		&IssueComment{},
		&IssueAttachment{},
		&IssueWatcher{},
		&IssueWorklog{},
		&Workflow{},
		&WorkflowNode{},
		&WorkflowEdge{},
		&WorkflowInstance{},
		&WorkflowHistory{},
		&ApprovalRecord{},
		&WorkflowScheme{},
		&Alert{},
		&AlertRule{},
		&AlertSilence{},
		&SystemConfig{},
		&Webhook{},
		&WebhookLog{},
		&ActivityLog{},
		&Notification{},
		&RequirementPool{},
		&Requirement{},
		&RequirementComment{},
		&RequirementAttachment{},
		&FieldDefinition{},
		&IssueTypeFieldScheme{},
		&IssueFieldValue{},
		&ProjectVersion{},
		&ProjectComponent{},
		&IssueLabel{},
		&AlertDatasource{},
		&ProjectNotificationChannel{},
		&ProjectRolePermission{},
	}

	for _, model := range models {
		if err := db.AutoMigrate(model); err != nil {
			logger.Error("failed to migrate model", zap.Error(err))
			return err
		}
	}

	logger.Info("database auto migration completed")

	// 创建复合索引（提升工单查询性能）
	createCompositeIndexes(db)

	// 迁移需求旧状态值到新状态值
	if err := migrateRequirementStatus(db); err != nil {
		logger.Error("failed to migrate requirement status", zap.Error(err))
		return err
	}

	return nil
}

// createCompositeIndexes 创建复合索引（提升工单查询性能）
func createCompositeIndexes(db *gorm.DB) {
	indexes := []struct {
		table string
		name  string
		cols  string
	}{
		{"issues", "idx_issues_project_status", "project_id, status"},
		{"issues", "idx_issues_assignee_status", "assignee_id, status"},
		{"issues", "idx_issues_reporter_created", "reporter_id, created_at DESC"},
		{"issues", "idx_issues_project_created", "project_id, created_at DESC"},
		{"issues", "idx_issues_status_created", "status, created_at DESC"},
	}

	for _, idx := range indexes {
		sql := "CREATE INDEX IF NOT EXISTS " + idx.name + " ON " + idx.table + " (" + idx.cols + ")"
		if err := db.Exec(sql).Error; err != nil {
			logger.Warn("failed to create composite index",
				zap.String("index", idx.name),
				zap.Error(err),
			)
		}
	}
}

// migrateRequirementStatus 迁移需求表中旧的状态值
func migrateRequirementStatus(db *gorm.DB) error {
	statusMapping := map[string]string{
		"pending":   "pending_review",
		"reviewing": "pending_review",
		"accepted":  "planning",
		"converted": "in_progress",
	}

	for oldStatus, newStatus := range statusMapping {
		result := db.Model(&Requirement{}).
			Where("status = ?", oldStatus).
			Update("status", newStatus)
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected > 0 {
			logger.Info("migrated requirement status",
				zap.String("from", oldStatus),
				zap.String("to", newStatus),
				zap.Int64("count", result.RowsAffected),
			)
		}
	}

	return nil
}

// SeedData 初始化种子数据
func SeedData(db *gorm.DB) error {
	logger.Info("seeding initial data...")

	// 初始化默认角色
	roles := []Role{
		{Name: "admin", DisplayName: "管理员", Description: "系统管理员，拥有所有权限"},
		{Name: "user", DisplayName: "普通用户", Description: "普通用户，基本操作权限"},
		{Name: "project_admin", DisplayName: "项目管理员", Description: "项目管理员，管理项目相关配置"},
	}

	for _, role := range roles {
		result := db.Where("name = ?", role.Name).FirstOrCreate(&role)
		if result.Error != nil {
			logger.Error("failed to seed role", zap.String("name", role.Name), zap.Error(result.Error))
			return result.Error
		}
	}

	// 初始化默认工单类型
	issueTypes := []IssueType{
		{Name: "Epic", DisplayName: "Epic", Description: "阶段性目标/大型需求", Icon: "lightning", Color: "#6554C0"},
		{Name: "Task", DisplayName: "任务", Description: "普通任务", Icon: "check-square", Color: "#4FADE6"},
		{Name: "Subtask", DisplayName: "子任务", Description: "子任务", Icon: "minus-square", Color: "#4FADE6"},
		{Name: "Bug", DisplayName: "缺陷", Description: "研发缺陷", Icon: "bug", Color: "#E5493A"},
		{Name: "Fault", DisplayName: "故障", Description: "生产故障/告警工单", Icon: "alert-triangle", Color: "#FF5630"},
		{Name: "Change", DisplayName: "变更", Description: "变更工单", Icon: "git-branch", Color: "#36B37E"},
		{Name: "ServiceRequest", DisplayName: "服务请求", Description: "服务申请", Icon: "help-circle", Color: "#00B8D9"},
	}

	for _, issueType := range issueTypes {
		result := db.Where("name = ? AND project_id IS NULL", issueType.Name).FirstOrCreate(&issueType)
		if result.Error != nil {
			logger.Error("failed to seed issue type", zap.String("name", issueType.Name), zap.Error(result.Error))
			return result.Error
		}
	}

	// 初始化默认管理员用户
	var adminUser User
	adminResult := db.Where("username = ?", "admin").First(&adminUser)
	if adminResult.Error != nil {
		if adminResult.Error == gorm.ErrRecordNotFound {
			// 创建默认管理员
			hashedPassword, err := bcrypt.GenerateFromPassword([]byte("admin123"), bcrypt.DefaultCost)
			if err != nil {
				logger.Error("failed to hash admin password", zap.Error(err))
				return err
			}

			adminUser = User{
				Username:     "admin",
				Email:        "admin@ticketdesk.local",
				PasswordHash: string(hashedPassword),
				DisplayName:  "系统管理员",
				Status:       1,
			}

			if err := db.Create(&adminUser).Error; err != nil {
				logger.Error("failed to create admin user", zap.Error(err))
				return err
			}

			// 分配管理员角色
			var adminRole Role
			if err := db.Where("name = ?", "admin").First(&adminRole).Error; err == nil {
				userRole := UserRole{
					UserID: adminUser.ID,
					RoleID: adminRole.ID,
				}
				if err := db.Create(&userRole).Error; err != nil {
					logger.Warn("failed to assign admin role", zap.Error(err))
				}
			}

			logger.Info("default admin user created",
				zap.String("username", "admin"),
				zap.String("password", "admin123"),
			)
		}
	}

	// 初始化默认系统配置
	defaultConfigs := []SystemConfig{
		{
			ConfigKey:   "general.site_url",
			ConfigValue: "http://localhost:5173",
			ConfigType:  "string",
			Category:    "general",
			Description: "站点域名（用于生成邮件中的链接）",
			IsSecret:    false,
		},
	}

	for _, config := range defaultConfigs {
		result := db.Where("config_key = ?", config.ConfigKey).FirstOrCreate(&config)
		if result.Error != nil {
			logger.Error("failed to seed system config", zap.String("key", config.ConfigKey), zap.Error(result.Error))
			return result.Error
		}
	}

	// 初始化系统字段
	if err := SeedSystemFields(db); err != nil {
		return err
	}

	// 初始化默认告警静默规则模板
	if err := seedDefaultAlertSilences(db, adminUser.ID); err != nil {
		return err
	}

	// 初始化默认工作流
	if err := SeedDefaultWorkflows(db); err != nil {
		return err
	}

	// 为已有项目初始化工作流方案（幂等）
	if err := seedExistingProjectWorkflowSchemes(db); err != nil {
		logger.Warn("failed to seed workflow schemes for existing projects", zap.Error(err))
	}

	// 为已有项目初始化角色权限（幂等）
	if err := seedExistingProjectRolePermissions(db); err != nil {
		logger.Warn("failed to seed role permissions for existing projects", zap.Error(err))
	}

	logger.Info("seed data completed")
	return nil
}

// InitProjectRoles 初始化项目的预置角色
func InitProjectRoles(db *gorm.DB, projectID uint64) error {
	logger.Info("initializing project roles", zap.Uint64("project_id", projectID))

	// 系统预置角色
	presetRoles := []ProjectRole{
		{
			ProjectID:   projectID,
			RoleKey:     "administrators",
			RoleName:    "管理员",
			Description: "项目管理员，拥有项目所有权限",
			IsSystem:    true,
			SortOrder:   1,
		},
		{
			ProjectID:   projectID,
			RoleKey:     "developers",
			RoleName:    "开发人员",
			Description: "项目开发人员",
			IsSystem:    true,
			SortOrder:   2,
		},
		{
			ProjectID:   projectID,
			RoleKey:     "testers",
			RoleName:    "测试人员",
			Description: "项目测试人员",
			IsSystem:    true,
			SortOrder:   3,
		},
		{
			ProjectID:   projectID,
			RoleKey:     "viewers",
			RoleName:    "只读用户",
			Description: "只能查看项目信息，不能修改",
			IsSystem:    true,
			SortOrder:   4,
		},
	}

	for _, role := range presetRoles {
		result := db.Where("project_id = ? AND role_key = ?", projectID, role.RoleKey).FirstOrCreate(&role)
		if result.Error != nil {
			logger.Error("failed to init project role",
				zap.Uint64("project_id", projectID),
				zap.String("role_key", role.RoleKey),
				zap.Error(result.Error))
			return result.Error
		}
	}

	logger.Info("project roles initialized", zap.Uint64("project_id", projectID))

	// 初始化角色默认权限
	if err := InitProjectRolePermissions(db, projectID); err != nil {
		logger.Warn("failed to init project role permissions", zap.Uint64("project_id", projectID), zap.Error(err))
	}

	return nil
}

// 系统角色默认权限定义
var defaultRolePermissions = map[string][]string{
	"administrators": {
		"project:view", "project:manage",
		"issue:view", "issue:create", "issue:edit", "issue:delete", "issue:assign",
		"member:view", "member:manage",
		"role:view", "role:manage",
		"workflow:view", "workflow:manage",
		"alert:view", "alert:manage",
	},
	"developers": {
		"project:view",
		"issue:view", "issue:create", "issue:edit", "issue:assign",
		"member:view",
		"role:view",
		"workflow:view",
		"alert:view",
	},
	"testers": {
		"project:view",
		"issue:view", "issue:create", "issue:edit",
		"member:view",
		"role:view",
		"workflow:view",
		"alert:view",
	},
	"viewers": {
		"project:view",
		"issue:view",
		"member:view",
		"role:view",
		"alert:view",
	},
}

// InitProjectRolePermissions 初始化项目角色的默认权限
func InitProjectRolePermissions(db *gorm.DB, projectID uint64) error {
	logger.Info("initializing project role permissions", zap.Uint64("project_id", projectID))

	for roleKey, permissions := range defaultRolePermissions {
		var role ProjectRole
		if err := db.Where("project_id = ? AND role_key = ?", projectID, roleKey).First(&role).Error; err != nil {
			continue // 角色不存在，跳过
		}

		for _, permKey := range permissions {
			perm := ProjectRolePermission{
				RoleID:        role.ID,
				PermissionKey: permKey,
			}
			db.Where("role_id = ? AND permission_key = ?", role.ID, permKey).FirstOrCreate(&perm)
		}
	}

	logger.Info("project role permissions initialized", zap.Uint64("project_id", projectID))
	return nil
}

// SeedSystemFields 初始化系统字段定义
func SeedSystemFields(db *gorm.DB) error {
	logger.Info("seeding system fields...")

	systemFields := []FieldDefinition{
		// Bug 特有字段
		{
			FieldKey:    "severity",
			FieldName:   "严重程度",
			FieldType:   FieldTypeSelect,
			Description: "缺陷的严重程度",
			IsSystem:    true,
			IsActive:    true,
			Options:     `[{"value":"blocker","label":"阻塞"},{"value":"critical","label":"严重"},{"value":"major","label":"主要"},{"value":"minor","label":"次要"},{"value":"trivial","label":"轻微"}]`,
			SortOrder:   1,
		},
		{
			FieldKey:    "environment",
			FieldName:   "环境",
			FieldType:   FieldTypeTextarea,
			Description: "问题发生的环境信息",
			IsSystem:    true,
			IsActive:    true,
			SortOrder:   2,
		},
		{
			FieldKey:    "steps_to_reproduce",
			FieldName:   "复现步骤",
			FieldType:   FieldTypeTextarea,
			Description: "问题复现的详细步骤",
			IsSystem:    true,
			IsActive:    true,
			SortOrder:   3,
		},
		// 版本相关字段
		{
			FieldKey:    "affects_version",
			FieldName:   "受影响版本",
			FieldType:   FieldTypeVersion,
			Description: "受此问题影响的版本",
			IsSystem:    true,
			IsActive:    true,
			SortOrder:   4,
		},
		{
			FieldKey:    "fix_version",
			FieldName:   "修复版本",
			FieldType:   FieldTypeVersion,
			Description: "修复此问题的目标版本",
			IsSystem:    true,
			IsActive:    true,
			SortOrder:   5,
		},
		// Epic 特有字段
		{
			FieldKey:    "epic_name",
			FieldName:   "Epic名称",
			FieldType:   FieldTypeText,
			Description: "Epic的简短名称（用于看板显示）",
			IsSystem:    true,
			IsActive:    true,
			SortOrder:   6,
		},
		{
			FieldKey:    "epic_color",
			FieldName:   "Epic颜色",
			FieldType:   FieldTypeSelect,
			Description: "Epic的标识颜色",
			IsSystem:    true,
			IsActive:    true,
			Options:     `[{"value":"#6554C0","label":"紫色"},{"value":"#4FADE6","label":"蓝色"},{"value":"#36B37E","label":"绿色"},{"value":"#FFAB00","label":"黄色"},{"value":"#FF5630","label":"红色"},{"value":"#00B8D9","label":"青色"}]`,
			SortOrder:   7,
		},
		// 日期字段
		{
			FieldKey:    "start_date",
			FieldName:   "开始日期",
			FieldType:   FieldTypeDate,
			Description: "计划开始日期",
			IsSystem:    true,
			IsActive:    true,
			SortOrder:   8,
		},
		{
			FieldKey:    "end_date",
			FieldName:   "结束日期",
			FieldType:   FieldTypeDate,
			Description: "计划结束日期",
			IsSystem:    true,
			IsActive:    true,
			SortOrder:   9,
		},
		// 时间估算字段
		{
			FieldKey:    "original_estimate",
			FieldName:   "预估时间",
			FieldType:   FieldTypeTimeEstimate,
			Description: "原始预估工作时间",
			IsSystem:    true,
			IsActive:    true,
			SortOrder:   10,
		},
		{
			FieldKey:    "remaining_estimate",
			FieldName:   "剩余时间",
			FieldType:   FieldTypeTimeEstimate,
			Description: "剩余预估工作时间",
			IsSystem:    true,
			IsActive:    true,
			SortOrder:   11,
		},
		// Epic链接字段
		{
			FieldKey:    "epic_link",
			FieldName:   "Epic链接",
			FieldType:   FieldTypeEpicLink,
			Description: "关联的Epic",
			IsSystem:    true,
			IsActive:    true,
			SortOrder:   12,
		},
		// 故事点
		{
			FieldKey:    "story_points",
			FieldName:   "故事点",
			FieldType:   FieldTypeNumber,
			Description: "任务复杂度估算",
			IsSystem:    true,
			IsActive:    true,
			SortOrder:   13,
		},
		// 通用字段
		{
			FieldKey:    "labels",
			FieldName:   "标签",
			FieldType:   FieldTypeLabel,
			Description: "工单标签",
			IsSystem:    true,
			IsActive:    true,
			SortOrder:   14,
		},
		{
			FieldKey:    "components",
			FieldName:   "组件",
			FieldType:   FieldTypeComponent,
			Description: "关联的项目组件",
			IsSystem:    true,
			IsActive:    true,
			SortOrder:   15,
		},
	}

	for _, field := range systemFields {
		// MySQL JSON 类型不接受空字符串，需要设置为有效 JSON
		if field.Options == "" {
			field.Options = "{}"
		}
		if field.Validation == "" {
			field.Validation = "{}"
		}
		result := db.Where("field_key = ? AND project_id IS NULL", field.FieldKey).FirstOrCreate(&field)
		if result.Error != nil {
			logger.Error("failed to seed system field", zap.String("field_key", field.FieldKey), zap.Error(result.Error))
			return result.Error
		}
	}

	logger.Info("system fields seeded")
	return nil
}

// InitProjectFieldSchemes 初始化项目的字段方案（为每个工单类型配置默认字段）
func InitProjectFieldSchemes(db *gorm.DB, projectID uint64) error {
	logger.Info("initializing project field schemes", zap.Uint64("project_id", projectID))

	// 获取系统字段
	var systemFields []FieldDefinition
	if err := db.Where("project_id IS NULL AND is_system = ?", true).Find(&systemFields).Error; err != nil {
		logger.Error("failed to get system fields", zap.Error(err))
		return err
	}

	// 创建字段key到ID的映射
	fieldMap := make(map[string]uint64)
	for _, field := range systemFields {
		fieldMap[field.FieldKey] = field.ID
	}

	// 获取全局工单类型
	var issueTypes []IssueType
	if err := db.Where("project_id IS NULL").Find(&issueTypes).Error; err != nil {
		logger.Error("failed to get issue types", zap.Error(err))
		return err
	}

	// 创建类型名到ID的映射
	typeMap := make(map[string]uint64)
	for _, it := range issueTypes {
		typeMap[it.Name] = it.ID
	}

	// 定义每种工单类型的字段配置
	// key: 工单类型名, value: 字段key列表及配置
	typeFieldConfig := map[string][]struct {
		FieldKey   string
		IsRequired bool
		SortOrder  int
	}{
		"Bug": {
			{FieldKey: "severity", IsRequired: true, SortOrder: 1},
			{FieldKey: "environment", IsRequired: false, SortOrder: 2},
			{FieldKey: "steps_to_reproduce", IsRequired: false, SortOrder: 3},
			{FieldKey: "affects_version", IsRequired: false, SortOrder: 4},
			{FieldKey: "fix_version", IsRequired: false, SortOrder: 5},
			{FieldKey: "epic_link", IsRequired: false, SortOrder: 6},
			{FieldKey: "labels", IsRequired: false, SortOrder: 7},
			{FieldKey: "components", IsRequired: false, SortOrder: 8},
		},
		"Epic": {
			{FieldKey: "epic_name", IsRequired: true, SortOrder: 1},
			{FieldKey: "epic_color", IsRequired: false, SortOrder: 2},
			{FieldKey: "start_date", IsRequired: false, SortOrder: 3},
			{FieldKey: "end_date", IsRequired: false, SortOrder: 4},
			{FieldKey: "story_points", IsRequired: false, SortOrder: 5},
			{FieldKey: "labels", IsRequired: false, SortOrder: 6},
			{FieldKey: "components", IsRequired: false, SortOrder: 7},
		},
		"Task": {
			{FieldKey: "original_estimate", IsRequired: false, SortOrder: 1},
			{FieldKey: "remaining_estimate", IsRequired: false, SortOrder: 2},
			{FieldKey: "start_date", IsRequired: false, SortOrder: 3},
			{FieldKey: "fix_version", IsRequired: false, SortOrder: 4},
			{FieldKey: "epic_link", IsRequired: false, SortOrder: 5},
			{FieldKey: "story_points", IsRequired: false, SortOrder: 6},
			{FieldKey: "labels", IsRequired: false, SortOrder: 7},
			{FieldKey: "components", IsRequired: false, SortOrder: 8},
		},
		"Fault": {
			{FieldKey: "severity", IsRequired: true, SortOrder: 1},
			{FieldKey: "environment", IsRequired: false, SortOrder: 2},
			{FieldKey: "affects_version", IsRequired: false, SortOrder: 3},
			{FieldKey: "labels", IsRequired: false, SortOrder: 4},
			{FieldKey: "components", IsRequired: false, SortOrder: 5},
		},
		"Change": {
			{FieldKey: "start_date", IsRequired: false, SortOrder: 1},
			{FieldKey: "end_date", IsRequired: false, SortOrder: 2},
			{FieldKey: "fix_version", IsRequired: false, SortOrder: 3},
			{FieldKey: "labels", IsRequired: false, SortOrder: 4},
			{FieldKey: "components", IsRequired: false, SortOrder: 5},
		},
		"ServiceRequest": {
			{FieldKey: "original_estimate", IsRequired: false, SortOrder: 1},
			{FieldKey: "labels", IsRequired: false, SortOrder: 2},
			{FieldKey: "components", IsRequired: false, SortOrder: 3},
		},
		"Subtask": {
			{FieldKey: "original_estimate", IsRequired: false, SortOrder: 1},
			{FieldKey: "remaining_estimate", IsRequired: false, SortOrder: 2},
			{FieldKey: "labels", IsRequired: false, SortOrder: 3},
		},
	}

	// 为每个工单类型创建字段方案
	for typeName, fields := range typeFieldConfig {
		typeID, ok := typeMap[typeName]
		if !ok {
			continue
		}

		for _, fieldConfig := range fields {
			fieldID, ok := fieldMap[fieldConfig.FieldKey]
			if !ok {
				continue
			}

			scheme := IssueTypeFieldScheme{
				ProjectID:       projectID,
				IssueTypeID:     typeID,
				FieldID:         fieldID,
				IsRequired:      fieldConfig.IsRequired,
				IsVisibleCreate: true,
				IsVisibleEdit:   true,
				IsVisibleDetail: true,
				SortOrder:       fieldConfig.SortOrder,
			}

			result := db.Where("project_id = ? AND issue_type_id = ? AND field_id = ?",
				projectID, typeID, fieldID).FirstOrCreate(&scheme)
			if result.Error != nil {
				logger.Error("failed to create field scheme",
					zap.Uint64("project_id", projectID),
					zap.String("type", typeName),
					zap.String("field", fieldConfig.FieldKey),
					zap.Error(result.Error))
				return result.Error
			}
		}
	}

	logger.Info("project field schemes initialized", zap.Uint64("project_id", projectID))
	return nil
}

// seedExistingProjectWorkflowSchemes 为已有项目初始化工作流方案
func seedExistingProjectWorkflowSchemes(db *gorm.DB) error {
	var projects []Project
	if err := db.Find(&projects).Error; err != nil {
		return err
	}

	for _, project := range projects {
		if err := InitProjectWorkflowSchemes(db, project.ID); err != nil {
			logger.Warn("failed to init workflow schemes for project",
				zap.Uint64("project_id", project.ID),
				zap.Error(err))
		}
	}
	return nil
}

// seedExistingProjectRolePermissions 为已有项目初始化角色权限
func seedExistingProjectRolePermissions(db *gorm.DB) error {
	var projects []Project
	if err := db.Find(&projects).Error; err != nil {
		return err
	}

	for _, project := range projects {
		if err := InitProjectRolePermissions(db, project.ID); err != nil {
			logger.Warn("failed to init role permissions for project",
				zap.Uint64("project_id", project.ID),
				zap.Error(err))
		}
	}
	return nil
}

// seedDefaultAlertSilences 初始化默认告警静默规则模板
func seedDefaultAlertSilences(db *gorm.DB, adminUserID uint64) error {
	logger.Info("seeding default alert silence templates...")

	// 使用固定的过去时间作为模板时间，状态为已取消(0)，用户使用时需修改时间并启用
	templateStart := time.Date(2025, 1, 1, 22, 0, 0, 0, time.Local)
	templateEnd := time.Date(2025, 1, 2, 6, 0, 0, 0, time.Local)
	templateStart2 := time.Date(2025, 1, 1, 0, 0, 0, 0, time.Local)
	templateEnd2 := time.Date(2025, 1, 2, 0, 0, 0, 0, time.Local)
	templateEnd7d := time.Date(2025, 1, 8, 0, 0, 0, 0, time.Local)

	silences := []AlertSilence{
		{
			Name:          "例行维护窗口 - 全局静默",
			Description:   "用于每周例行维护期间静默所有告警，避免维护操作触发大量无意义工单。使用时修改时间并启用。",
			LabelMatchers: `[{"key":"severity","value":"critical|warning|info","operator":"=~"}]`,
			StartsAt:      templateStart,
			EndsAt:        templateEnd,
			CreatedBy:     adminUserID,
			Comment:       "模板规则：例行维护期间静默所有告警，使用前请修改时间并启用",
			Status:        0,
		},
		{
			Name:          "例行维护窗口 - 指定主机",
			Description:   "用于对特定主机进行维护时静默该主机的所有告警。使用时修改 instance 为目标主机IP，修改时间并启用。",
			LabelMatchers: `[{"key":"instance","value":"10.0.0.1.*","operator":"=~"}]`,
			StartsAt:      templateStart,
			EndsAt:        templateEnd,
			CreatedBy:     adminUserID,
			Comment:       "模板规则：指定主机维护静默，使用前请修改主机IP、时间并启用",
			Status:        0,
		},
		{
			Name:          "例行维护窗口 - 指定业务组",
			Description:   "用于对特定业务组进行维护时静默该组的所有告警。使用时修改 group_name 为目标业务组，修改时间并启用。",
			LabelMatchers: `[{"key":"group_name","value":"prod-app","operator":"=="}]`,
			StartsAt:      templateStart,
			EndsAt:        templateEnd,
			CreatedBy:     adminUserID,
			Comment:       "模板规则：指定业务组维护静默，使用前请修改业务组名、时间并启用",
			Status:        0,
		},
		{
			Name:          "临时静默 - 指定告警名称",
			Description:   "用于临时屏蔽某个已知告警（如误报、正在处理中的告警）。使用时修改 alertname 为目标告警名称，修改时间并启用。",
			LabelMatchers: `[{"key":"alertname","value":"HostHighCpuUsage-level-1","operator":"=="}]`,
			StartsAt:      templateStart2,
			EndsAt:        templateEnd2,
			CreatedBy:     adminUserID,
			Comment:       "模板规则：临时屏蔽指定告警，使用前请修改告警名称、时间并启用",
			Status:        0,
		},
		{
			Name:          "临时静默 - Falco安全告警",
			Description:   "用于临时屏蔽 Falco 容器安全告警，例如在部署或调试期间容器行为可能触发大量安全告警。使用时修改时间并启用。",
			LabelMatchers: `[{"key":"alertname","value":".*Falco.*","operator":"=~"}]`,
			StartsAt:      templateStart2,
			EndsAt:        templateEnd2,
			CreatedBy:     adminUserID,
			Comment:       "模板规则：临时屏蔽Falco安全告警，使用前请修改时间并启用",
			Status:        0,
		},
		{
			Name:          "临时静默 - 低级别告警",
			Description:   "用于故障高峰期只关注严重告警，临时屏蔽 warning 和 info 级别告警，减少噪音。使用时修改时间并启用。",
			LabelMatchers: `[{"key":"severity","value":"warning|info","operator":"=~"}]`,
			StartsAt:      templateStart2,
			EndsAt:        templateEnd7d,
			CreatedBy:     adminUserID,
			Comment:       "模板规则：屏蔽低级别告警只保留critical，使用前请修改时间并启用",
			Status:        0,
		},
		{
			Name:          "临时静默 - 测试/UAT环境",
			Description:   "用于临时屏蔽测试和UAT环境的所有告警，例如在测试环境大规模变更期间。使用时修改时间并启用。",
			LabelMatchers: `[{"key":"group_name","value":".*uat.*|.*test.*","operator":"=~"}]`,
			StartsAt:      templateStart2,
			EndsAt:        templateEnd7d,
			CreatedBy:     adminUserID,
			Comment:       "模板规则：屏蔽测试/UAT环境告警，使用前请修改时间并启用",
			Status:        0,
		},
	}

	for _, silence := range silences {
		result := db.Where("name = ?", silence.Name).FirstOrCreate(&silence)
		if result.Error != nil {
			logger.Error("failed to seed alert silence", zap.String("name", silence.Name), zap.Error(result.Error))
			return result.Error
		}
	}

	logger.Info("default alert silence templates seeded")
	return nil
}
