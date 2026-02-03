// Package model 提供数据库迁移功能
package model

import (
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
		&IssueType{},
		&Issue{},
		&IssueComment{},
		&IssueWatcher{},
		&Workflow{},
		&WorkflowNode{},
		&WorkflowEdge{},
		&Alert{},
		&AlertRule{},
		&AlertSilence{},
		&SystemConfig{},
		&Webhook{},
		&WebhookLog{},
	}

	for _, model := range models {
		if err := db.AutoMigrate(model); err != nil {
			logger.Error("failed to migrate model", zap.Error(err))
			return err
		}
	}

	logger.Info("database auto migration completed")
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

	logger.Info("seed data completed")
	return nil
}
