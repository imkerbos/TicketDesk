package model

import "time"

// RequirementCategoryDef 需求分类定义（动态管理）
type RequirementCategoryDef struct {
	ID        uint64    `gorm:"primaryKey;autoIncrement" json:"id"`
	Name      string    `gorm:"type:varchar(30);uniqueIndex;not null" json:"name"`  // 唯一标识（如 feature）
	Label     string    `gorm:"type:varchar(50);not null" json:"label"`             // 显示名称（如 功能需求）
	Color     string    `gorm:"type:varchar(20);not null;default:'info'" json:"color"` // 标签颜色（primary/success/danger/warning/info）
	SortOrder int       `gorm:"default:0" json:"sort_order"`                        // 排序
	IsDefault bool      `gorm:"default:false" json:"is_default"`                    // 是否默认选中
	IsSystem  bool      `gorm:"default:false" json:"is_system"`                     // 系统预置不可删除
	CreatedAt time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt time.Time `gorm:"autoUpdateTime" json:"updated_at"`
}

// TableName 指定表名
func (RequirementCategoryDef) TableName() string {
	return "requirement_category_defs"
}
