// Package model 提供数据库迁移功能
package model

import (
	"github.com/kerbos/ticketdesk/pkg/logger"
	"go.uber.org/zap"
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
		&IssueType{},
		&Issue{},
		&IssueComment{},
		&IssueWatcher{},
		&Workflow{},
		&WorkflowNode{},
		&WorkflowEdge{},
		&Alert{},
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

	logger.Info("seed data completed")
	return nil
}
