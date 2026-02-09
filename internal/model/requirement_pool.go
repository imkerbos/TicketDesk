package model

import (
	"time"

	"gorm.io/gorm"
)

// RequirementPoolType 需求池类型
type RequirementPoolType string

const (
	RequirementPoolTypeGlobal  RequirementPoolType = "global"  // 全局需求池
	RequirementPoolTypeProject RequirementPoolType = "project" // 项目级需求池
)

// RequirementPoolStatus 需求池状态
type RequirementPoolStatus string

const (
	RequirementPoolStatusActive   RequirementPoolStatus = "active"   // 活跃
	RequirementPoolStatusArchived RequirementPoolStatus = "archived" // 已归档
)

// RequirementPool 需求池
type RequirementPool struct {
	ID          uint64                `gorm:"primaryKey;autoIncrement" json:"id"`
	Name        string                `gorm:"type:varchar(200);not null;index" json:"name"`
	Description string                `gorm:"type:text" json:"description"`
	Type        RequirementPoolType   `gorm:"type:varchar(20);not null;index" json:"type"`
	ProjectID   *uint64               `gorm:"index" json:"project_id,omitempty"` // 项目级需求池关联的项目ID
	OwnerID     uint64                `gorm:"not null;index" json:"owner_id"`    // 负责人ID
	Status      RequirementPoolStatus `gorm:"type:varchar(20);not null;index;default:'active'" json:"status"`
	CreatedAt   time.Time             `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt   time.Time             `gorm:"autoUpdateTime" json:"updated_at"`
	DeletedAt   gorm.DeletedAt        `gorm:"index" json:"deleted_at,omitempty"`

	// 关联
	Owner        *User          `gorm:"foreignKey:OwnerID" json:"owner,omitempty"`
	Project      *Project       `gorm:"foreignKey:ProjectID" json:"project,omitempty"`
	Requirements []*Requirement `gorm:"foreignKey:PoolID" json:"requirements,omitempty"`
}

// TableName 指定表名
func (RequirementPool) TableName() string {
	return "requirement_pools"
}
