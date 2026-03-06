package dto

// CreateCategoryRequest 创建需求分类请求
type CreateCategoryRequest struct {
	Name  string `json:"name" binding:"required,min=1,max=30"`
	Label string `json:"label" binding:"required,min=1,max=50"`
	Color string `json:"color" binding:"required,oneof=primary success danger warning info"`
}

// UpdateCategoryRequest 更新需求分类请求
type UpdateCategoryRequest struct {
	Label     *string `json:"label" binding:"omitempty,min=1,max=50"`
	Color     *string `json:"color" binding:"omitempty,oneof=primary success danger warning info"`
	SortOrder *int    `json:"sort_order" binding:"omitempty,min=0"`
	IsDefault *bool   `json:"is_default"`
}

// CategoryResponse 需求分类响应
type CategoryResponse struct {
	ID        uint64 `json:"id"`
	Name      string `json:"name"`
	Label     string `json:"label"`
	Color     string `json:"color"`
	SortOrder int    `json:"sort_order"`
	IsDefault bool   `json:"is_default"`
	IsSystem  bool   `json:"is_system"`
}
