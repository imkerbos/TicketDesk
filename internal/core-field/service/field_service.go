// Package service 提供字段业务逻辑层
package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/kerbos/ticketdesk/internal/core-field/dto"
	"github.com/kerbos/ticketdesk/internal/core-field/repository"
	projectRepo "github.com/kerbos/ticketdesk/internal/core-project/repository"
	userRepo "github.com/kerbos/ticketdesk/internal/core-user/repository"
	"github.com/kerbos/ticketdesk/internal/model"
	"github.com/kerbos/ticketdesk/pkg/logger"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// 业务错误定义
var (
	ErrProjectNotFound      = errors.New("项目不存在")
	ErrFieldNotFound        = errors.New("字段不存在")
	ErrFieldKeyExists       = errors.New("字段Key已存在")
	ErrCannotDeleteSystem   = errors.New("不能删除系统字段")
	ErrCannotModifySystem   = errors.New("不能修改系统字段")
	ErrIssueTypeNotFound    = errors.New("工单类型不存在")
	ErrVersionNotFound      = errors.New("版本不存在")
	ErrVersionNameExists    = errors.New("版本名称已存在")
	ErrComponentNotFound    = errors.New("组件不存在")
	ErrComponentNameExists  = errors.New("组件名称已存在")
	ErrLabelNotFound        = errors.New("标签不存在")
	ErrLabelNameExists      = errors.New("标签名称已存在")
)

// FieldService 字段服务接口
type FieldService interface {
	// 字段定义
	CreateField(ctx context.Context, projectKey string, req *dto.CreateFieldRequest) (*dto.FieldDefinitionResponse, error)
	UpdateField(ctx context.Context, projectKey string, fieldID uint64, req *dto.UpdateFieldRequest) (*dto.FieldDefinitionResponse, error)
	DeleteField(ctx context.Context, projectKey string, fieldID uint64) error
	GetField(ctx context.Context, fieldID uint64) (*dto.FieldDefinitionResponse, error)
	ListFields(ctx context.Context, projectKey string) ([]*dto.FieldDefinitionResponse, error)

	// 字段方案
	GetFieldScheme(ctx context.Context, projectKey string, issueTypeID uint64) ([]*dto.FieldSchemeResponse, error)
	UpdateFieldScheme(ctx context.Context, projectKey string, issueTypeID uint64, req *dto.UpdateFieldSchemeRequest) ([]*dto.FieldSchemeResponse, error)

	// 字段值
	GetFieldValues(ctx context.Context, issueID uint64) ([]*dto.FieldValueResponse, error)
	SaveFieldValues(ctx context.Context, issueID uint64, values []*dto.CustomFieldValue) error
	SaveIssueFieldValues(ctx context.Context, issueID uint64, values []CustomFieldValueInput) error
	GetIssueIDsByEpicLink(ctx context.Context, epicID uint64) ([]uint64, error)

	// 版本管理
	CreateVersion(ctx context.Context, projectKey string, req *dto.CreateVersionRequest) (*dto.VersionResponse, error)
	UpdateVersion(ctx context.Context, projectKey string, versionID uint64, req *dto.UpdateVersionRequest) (*dto.VersionResponse, error)
	DeleteVersion(ctx context.Context, projectKey string, versionID uint64) error
	ListVersions(ctx context.Context, projectKey string) ([]*dto.VersionResponse, error)

	// 组件管理
	CreateComponent(ctx context.Context, projectKey string, req *dto.CreateComponentRequest) (*dto.ComponentResponse, error)
	UpdateComponent(ctx context.Context, projectKey string, componentID uint64, req *dto.UpdateComponentRequest) (*dto.ComponentResponse, error)
	DeleteComponent(ctx context.Context, projectKey string, componentID uint64) error
	ListComponents(ctx context.Context, projectKey string) ([]*dto.ComponentResponse, error)

	// 标签管理
	CreateLabel(ctx context.Context, projectKey string, req *dto.CreateLabelRequest) (*dto.LabelResponse, error)
	UpdateLabel(ctx context.Context, projectKey string, labelID uint64, req *dto.UpdateLabelRequest) (*dto.LabelResponse, error)
	DeleteLabel(ctx context.Context, projectKey string, labelID uint64) error
	ListLabels(ctx context.Context, projectKey string) ([]*dto.LabelResponse, error)
}

// fieldService 字段服务实现
type fieldService struct {
	fieldRepo     repository.FieldRepository
	schemeRepo    repository.SchemeRepository
	valueRepo     repository.ValueRepository
	versionRepo   repository.VersionRepository
	componentRepo repository.ComponentRepository
	labelRepo     repository.LabelRepository
	projectRepo   projectRepo.ProjectRepository
	issueTypeRepo projectRepo.IssueTypeRepository
	userRepo      userRepo.UserRepository
	db            *gorm.DB
}

// NewFieldService 创建字段服务实例
func NewFieldService(
	fieldRepo repository.FieldRepository,
	schemeRepo repository.SchemeRepository,
	valueRepo repository.ValueRepository,
	versionRepo repository.VersionRepository,
	componentRepo repository.ComponentRepository,
	labelRepo repository.LabelRepository,
	projectRepo projectRepo.ProjectRepository,
	issueTypeRepo projectRepo.IssueTypeRepository,
	userRepo userRepo.UserRepository,
	db *gorm.DB,
) FieldService {
	return &fieldService{
		fieldRepo:     fieldRepo,
		schemeRepo:    schemeRepo,
		valueRepo:     valueRepo,
		versionRepo:   versionRepo,
		componentRepo: componentRepo,
		labelRepo:     labelRepo,
		projectRepo:   projectRepo,
		issueTypeRepo: issueTypeRepo,
		userRepo:      userRepo,
		db:            db,
	}
}

// ============ 字段定义 ============

// CreateField 创建自定义字段
func (s *fieldService) CreateField(ctx context.Context, projectKey string, req *dto.CreateFieldRequest) (*dto.FieldDefinitionResponse, error) {
	project, err := s.projectRepo.GetByKey(ctx, strings.ToUpper(projectKey))
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrProjectNotFound
		}
		return nil, fmt.Errorf("查询项目失败: %w", err)
	}

	// 检查字段Key是否已存在
	_, err = s.fieldRepo.GetByKey(ctx, &project.ID, req.FieldKey)
	if err == nil {
		return nil, ErrFieldKeyExists
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, fmt.Errorf("检查字段Key失败: %w", err)
	}

	field := &model.FieldDefinition{
		ProjectID:    &project.ID,
		FieldKey:     req.FieldKey,
		FieldName:    req.FieldName,
		FieldType:    req.FieldType,
		Description:  req.Description,
		IsSystem:     false,
		IsActive:     true,
		Options:      req.Options,
		Validation:   req.Validation,
		DefaultValue: req.DefaultValue,
		SortOrder:    100,
	}

	if err := s.fieldRepo.Create(ctx, field); err != nil {
		logger.Error("failed to create field", zap.Error(err))
		return nil, fmt.Errorf("创建字段失败: %w", err)
	}

	logger.Info("field created",
		zap.String("project_key", projectKey),
		zap.String("field_key", req.FieldKey),
	)

	return s.toFieldResponse(field), nil
}

// UpdateField 更新字段定义
func (s *fieldService) UpdateField(ctx context.Context, projectKey string, fieldID uint64, req *dto.UpdateFieldRequest) (*dto.FieldDefinitionResponse, error) {
	project, err := s.projectRepo.GetByKey(ctx, strings.ToUpper(projectKey))
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrProjectNotFound
		}
		return nil, fmt.Errorf("查询项目失败: %w", err)
	}

	field, err := s.fieldRepo.GetByID(ctx, fieldID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrFieldNotFound
		}
		return nil, fmt.Errorf("查询字段失败: %w", err)
	}

	// 验证字段属于该项目（系统字段除外）
	if field.ProjectID != nil && *field.ProjectID != project.ID {
		return nil, ErrFieldNotFound
	}

	// 系统字段只能修改部分属性
	if field.IsSystem {
		if req.IsActive != nil {
			field.IsActive = *req.IsActive
		}
		if req.SortOrder != nil {
			field.SortOrder = *req.SortOrder
		}
	} else {
		if req.FieldName != nil {
			field.FieldName = *req.FieldName
		}
		if req.Description != nil {
			field.Description = *req.Description
		}
		if req.Options != nil {
			field.Options = *req.Options
		}
		if req.Validation != nil {
			field.Validation = *req.Validation
		}
		if req.DefaultValue != nil {
			field.DefaultValue = *req.DefaultValue
		}
		if req.IsActive != nil {
			field.IsActive = *req.IsActive
		}
		if req.SortOrder != nil {
			field.SortOrder = *req.SortOrder
		}
	}

	if err := s.fieldRepo.Update(ctx, field); err != nil {
		logger.Error("failed to update field", zap.Error(err))
		return nil, fmt.Errorf("更新字段失败: %w", err)
	}

	return s.toFieldResponse(field), nil
}

// DeleteField 删除字段定义
func (s *fieldService) DeleteField(ctx context.Context, projectKey string, fieldID uint64) error {
	project, err := s.projectRepo.GetByKey(ctx, strings.ToUpper(projectKey))
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrProjectNotFound
		}
		return fmt.Errorf("查询项目失败: %w", err)
	}

	field, err := s.fieldRepo.GetByID(ctx, fieldID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrFieldNotFound
		}
		return fmt.Errorf("查询字段失败: %w", err)
	}

	// 不能删除系统字段
	if field.IsSystem {
		return ErrCannotDeleteSystem
	}

	// 验证字段属于该项目
	if field.ProjectID == nil || *field.ProjectID != project.ID {
		return ErrFieldNotFound
	}

	if err := s.fieldRepo.Delete(ctx, fieldID); err != nil {
		logger.Error("failed to delete field", zap.Error(err))
		return fmt.Errorf("删除字段失败: %w", err)
	}

	logger.Info("field deleted",
		zap.String("project_key", projectKey),
		zap.Uint64("field_id", fieldID),
	)

	return nil
}

// GetField 获取字段定义
func (s *fieldService) GetField(ctx context.Context, fieldID uint64) (*dto.FieldDefinitionResponse, error) {
	field, err := s.fieldRepo.GetByID(ctx, fieldID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrFieldNotFound
		}
		return nil, fmt.Errorf("查询字段失败: %w", err)
	}

	return s.toFieldResponse(field), nil
}

// ListFields 获取项目可用的所有字段
func (s *fieldService) ListFields(ctx context.Context, projectKey string) ([]*dto.FieldDefinitionResponse, error) {
	project, err := s.projectRepo.GetByKey(ctx, strings.ToUpper(projectKey))
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrProjectNotFound
		}
		return nil, fmt.Errorf("查询项目失败: %w", err)
	}

	fields, err := s.fieldRepo.ListAllAvailable(ctx, project.ID)
	if err != nil {
		return nil, fmt.Errorf("查询字段列表失败: %w", err)
	}

	responses := make([]*dto.FieldDefinitionResponse, len(fields))
	for i, field := range fields {
		responses[i] = s.toFieldResponse(field)
	}

	return responses, nil
}

// ============ 字段方案 ============

// GetFieldScheme 获取工单类型的字段方案
func (s *fieldService) GetFieldScheme(ctx context.Context, projectKey string, issueTypeID uint64) ([]*dto.FieldSchemeResponse, error) {
	project, err := s.projectRepo.GetByKey(ctx, strings.ToUpper(projectKey))
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrProjectNotFound
		}
		return nil, fmt.Errorf("查询项目失败: %w", err)
	}

	// 验证工单类型存在
	_, err = s.issueTypeRepo.GetByID(ctx, issueTypeID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrIssueTypeNotFound
		}
		return nil, fmt.Errorf("查询工单类型失败: %w", err)
	}

	schemes, err := s.schemeRepo.ListByProjectAndType(ctx, project.ID, issueTypeID)
	if err != nil {
		return nil, fmt.Errorf("查询字段方案失败: %w", err)
	}

	responses := make([]*dto.FieldSchemeResponse, len(schemes))
	for i, scheme := range schemes {
		responses[i] = s.toSchemeResponse(ctx, scheme)
	}

	return responses, nil
}

// UpdateFieldScheme 更新工单类型的字段方案
func (s *fieldService) UpdateFieldScheme(ctx context.Context, projectKey string, issueTypeID uint64, req *dto.UpdateFieldSchemeRequest) ([]*dto.FieldSchemeResponse, error) {
	project, err := s.projectRepo.GetByKey(ctx, strings.ToUpper(projectKey))
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrProjectNotFound
		}
		return nil, fmt.Errorf("查询项目失败: %w", err)
	}

	// 验证工单类型存在
	_, err = s.issueTypeRepo.GetByID(ctx, issueTypeID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrIssueTypeNotFound
		}
		return nil, fmt.Errorf("查询工单类型失败: %w", err)
	}

	// 使用事务更新
	err = s.db.Transaction(func(tx *gorm.DB) error {
		// 删除旧的方案
		if err := s.schemeRepo.DeleteByProjectAndType(ctx, project.ID, issueTypeID); err != nil {
			return fmt.Errorf("删除旧方案失败: %w", err)
		}

		// 创建新的方案
		schemes := make([]*model.IssueTypeFieldScheme, len(req.Items))
		for i, f := range req.Items {
			schemes[i] = &model.IssueTypeFieldScheme{
				ProjectID:       project.ID,
				IssueTypeID:     issueTypeID,
				FieldID:         f.FieldID,
				IsRequired:      f.IsRequired,
				IsVisibleCreate: f.IsVisibleCreate,
				IsVisibleEdit:   f.IsVisibleEdit,
				IsVisibleDetail: f.IsVisibleDetail,
				SortOrder:       f.SortOrder,
				DefaultValue:    f.DefaultValue,
			}
		}

		if err := s.schemeRepo.BatchCreate(ctx, schemes); err != nil {
			return fmt.Errorf("创建新方案失败: %w", err)
		}

		return nil
	})

	if err != nil {
		logger.Error("failed to update field scheme", zap.Error(err))
		return nil, err
	}

	logger.Info("field scheme updated",
		zap.String("project_key", projectKey),
		zap.Uint64("issue_type_id", issueTypeID),
	)

	return s.GetFieldScheme(ctx, projectKey, issueTypeID)
}

// ============ 字段值 ============

// GetFieldValues 获取工单的自定义字段值
func (s *fieldService) GetFieldValues(ctx context.Context, issueID uint64) ([]*dto.FieldValueResponse, error) {
	values, err := s.valueRepo.ListByIssue(ctx, issueID)
	if err != nil {
		return nil, fmt.Errorf("查询字段值失败: %w", err)
	}

	responses := make([]*dto.FieldValueResponse, 0, len(values))
	for _, v := range values {
		field, err := s.fieldRepo.GetByID(ctx, v.FieldID)
		if err != nil {
			continue
		}

		resp := &dto.FieldValueResponse{
			FieldID:   v.FieldID,
			FieldKey:  field.FieldKey,
			FieldName: field.FieldName,
			FieldType: field.FieldType,
		}

		// 根据字段类型获取值
		switch field.FieldType {
		case model.FieldTypeNumber:
			if v.ValueNumber != nil {
				resp.Value = *v.ValueNumber
				resp.DisplayValue = fmt.Sprintf("%v", *v.ValueNumber)
			}
		case model.FieldTypeDate:
			if v.ValueDate != nil {
				resp.Value = v.ValueDate.Format("2006-01-02")
				resp.DisplayValue = v.ValueDate.Format("2006-01-02")
			}
		case model.FieldTypeMultiSelect, model.FieldTypeLabel, model.FieldTypeComponent:
			if v.ValueJSON != nil && *v.ValueJSON != "" {
				var arr []any
				if err := json.Unmarshal([]byte(*v.ValueJSON), &arr); err == nil {
					resp.Value = arr
					resp.DisplayValue = *v.ValueJSON
				}
			}
		case model.FieldTypeEpicLink:
			// Epic Link 特殊处理：查询关联的 Epic 信息
			if v.ValueNumber != nil {
				epicID := uint64(*v.ValueNumber)
				var epicIssue model.Issue
				if err := s.db.Where("id = ?", epicID).First(&epicIssue).Error; err == nil {
					resp.Value = map[string]any{
						"id":        epicIssue.ID,
						"issue_key": epicIssue.IssueKey,
						"title":     epicIssue.Title,
					}
					resp.DisplayValue = fmt.Sprintf("%s: %s", epicIssue.IssueKey, epicIssue.Title)
				} else {
					resp.Value = epicID
					resp.DisplayValue = fmt.Sprintf("%d", epicID)
				}
			}
		default:
			if v.ValueText != nil {
				resp.Value = *v.ValueText
				resp.DisplayValue = *v.ValueText
			}
		}

		responses = append(responses, resp)
	}

	return responses, nil
}

// SaveFieldValues 保存工单的自定义字段值
func (s *fieldService) SaveFieldValues(ctx context.Context, issueID uint64, values []*dto.CustomFieldValue) error {
	logger.Info("SaveFieldValues: starting",
		zap.Uint64("issue_id", issueID),
		zap.Int("values_count", len(values)),
	)
	fieldValues := make([]*model.IssueFieldValue, 0, len(values))

	for _, v := range values {
		logger.Info("SaveFieldValues: processing field",
			zap.Uint64("field_id", v.FieldID),
			zap.Any("value", v.Value),
		)
		field, err := s.fieldRepo.GetByID(ctx, v.FieldID)
		if err != nil {
			logger.Warn("SaveFieldValues: failed to get field", zap.Uint64("field_id", v.FieldID), zap.Error(err))
			continue
		}

		fv := &model.IssueFieldValue{
			IssueID: issueID,
			FieldID: v.FieldID,
		}

		// 根据字段类型设置值
		switch field.FieldType {
		case model.FieldTypeNumber:
			if num, ok := v.Value.(float64); ok {
				fv.ValueNumber = &num
			}
		case model.FieldTypeDate:
			if dateStr, ok := v.Value.(string); ok {
				if t, err := time.Parse("2006-01-02", dateStr); err == nil {
					fv.ValueDate = &t
				}
			}
		case model.FieldTypeMultiSelect, model.FieldTypeLabel, model.FieldTypeComponent:
			if arr, ok := v.Value.([]any); ok {
				if jsonBytes, err := json.Marshal(arr); err == nil {
					jsonStr := string(jsonBytes)
					fv.ValueJSON = &jsonStr
				}
			}
		case model.FieldTypeEpicLink:
			// Epic Link 保存为 number 类型（Issue ID）
			switch val := v.Value.(type) {
			case float64:
				fv.ValueNumber = &val
			case int:
				num := float64(val)
				fv.ValueNumber = &num
			case string:
				// 如果是字符串，尝试转换为数字
				if num, err := strconv.ParseFloat(val, 64); err == nil {
					fv.ValueNumber = &num
				}
			}
		default:
			if str, ok := v.Value.(string); ok {
				fv.ValueText = &str
			} else if v.Value != nil {
				str := fmt.Sprintf("%v", v.Value)
				fv.ValueText = &str
			}
		}

		fieldValues = append(fieldValues, fv)
	}

	if err := s.valueRepo.BatchUpsert(ctx, fieldValues); err != nil {
		logger.Error("failed to save field values", zap.Error(err))
		return fmt.Errorf("保存字段值失败: %w", err)
	}

	return nil
}

// CustomFieldValueInput 用于避免循环依赖的输入类型
type CustomFieldValueInput struct {
	FieldID uint64
	Value   any
}

// SaveIssueFieldValues 保存工单字段值（满足 issue_service.FieldValueSaver 接口）
func (s *fieldService) SaveIssueFieldValues(ctx context.Context, issueID uint64, values []CustomFieldValueInput) error {
	// 转换为内部 DTO 类型
	dtoValues := make([]*dto.CustomFieldValue, len(values))
	for i, v := range values {
		dtoValues[i] = &dto.CustomFieldValue{
			FieldID: v.FieldID,
			Value:   v.Value,
		}
	}
	return s.SaveFieldValues(ctx, issueID, dtoValues)
}

// GetIssueIDsByEpicLink 获取链接到指定 Epic 的所有 Issue IDs
func (s *fieldService) GetIssueIDsByEpicLink(ctx context.Context, epicID uint64) ([]uint64, error) {
	// 查询 epic_link 字段的定义
	var epicLinkField model.FieldDefinition
	if err := s.db.Where("field_key = ?", "epic_link").First(&epicLinkField).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return []uint64{}, nil
		}
		return nil, fmt.Errorf("查询 epic_link 字段失败: %w", err)
	}

	// 查询所有 epic_link 字段值指向该 Epic 的 Issues
	var fieldValues []model.IssueFieldValue
	if err := s.db.Where("field_id = ? AND value_number = ?", epicLinkField.ID, float64(epicID)).
		Find(&fieldValues).Error; err != nil {
		logger.Error("failed to query epic link field values", zap.Error(err))
		return nil, fmt.Errorf("查询 Epic 关联工单失败: %w", err)
	}

	issueIDs := make([]uint64, len(fieldValues))
	for i, fv := range fieldValues {
		issueIDs[i] = fv.IssueID
	}

	return issueIDs, nil
}

// ============ 版本管理 ============

// CreateVersion 创建项目版本
func (s *fieldService) CreateVersion(ctx context.Context, projectKey string, req *dto.CreateVersionRequest) (*dto.VersionResponse, error) {
	project, err := s.projectRepo.GetByKey(ctx, strings.ToUpper(projectKey))
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrProjectNotFound
		}
		return nil, fmt.Errorf("查询项目失败: %w", err)
	}

	// 检查版本名称是否已存在
	_, err = s.versionRepo.GetByName(ctx, project.ID, req.Name)
	if err == nil {
		return nil, ErrVersionNameExists
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, fmt.Errorf("检查版本名称失败: %w", err)
	}

	version := &model.ProjectVersion{
		ProjectID:   project.ID,
		Name:        req.Name,
		Description: req.Description,
		Status:      model.VersionStatusUnreleased,
	}

	if req.ReleaseDate != nil {
		if t, err := time.Parse("2006-01-02", *req.ReleaseDate); err == nil {
			version.ReleaseDate = &t
		}
	}

	if err := s.versionRepo.Create(ctx, version); err != nil {
		logger.Error("failed to create version", zap.Error(err))
		return nil, fmt.Errorf("创建版本失败: %w", err)
	}

	logger.Info("version created",
		zap.String("project_key", projectKey),
		zap.String("version_name", req.Name),
	)

	return s.toVersionResponse(version), nil
}

// UpdateVersion 更新项目版本
func (s *fieldService) UpdateVersion(ctx context.Context, projectKey string, versionID uint64, req *dto.UpdateVersionRequest) (*dto.VersionResponse, error) {
	project, err := s.projectRepo.GetByKey(ctx, strings.ToUpper(projectKey))
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrProjectNotFound
		}
		return nil, fmt.Errorf("查询项目失败: %w", err)
	}

	version, err := s.versionRepo.GetByID(ctx, versionID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrVersionNotFound
		}
		return nil, fmt.Errorf("查询版本失败: %w", err)
	}

	// 验证版本属于该项目
	if version.ProjectID != project.ID {
		return nil, ErrVersionNotFound
	}

	if req.Name != nil {
		// 检查新名称是否已存在
		existing, err := s.versionRepo.GetByName(ctx, project.ID, *req.Name)
		if err == nil && existing.ID != versionID {
			return nil, ErrVersionNameExists
		}
		version.Name = *req.Name
	}
	if req.Description != nil {
		version.Description = *req.Description
	}
	if req.ReleaseDate != nil {
		if t, err := time.Parse("2006-01-02", *req.ReleaseDate); err == nil {
			version.ReleaseDate = &t
		}
	}
	if req.Status != nil {
		version.Status = *req.Status
	}

	if err := s.versionRepo.Update(ctx, version); err != nil {
		logger.Error("failed to update version", zap.Error(err))
		return nil, fmt.Errorf("更新版本失败: %w", err)
	}

	return s.toVersionResponse(version), nil
}

// DeleteVersion 删除项目版本
func (s *fieldService) DeleteVersion(ctx context.Context, projectKey string, versionID uint64) error {
	project, err := s.projectRepo.GetByKey(ctx, strings.ToUpper(projectKey))
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrProjectNotFound
		}
		return fmt.Errorf("查询项目失败: %w", err)
	}

	version, err := s.versionRepo.GetByID(ctx, versionID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrVersionNotFound
		}
		return fmt.Errorf("查询版本失败: %w", err)
	}

	// 验证版本属于该项目
	if version.ProjectID != project.ID {
		return ErrVersionNotFound
	}

	if err := s.versionRepo.Delete(ctx, versionID); err != nil {
		logger.Error("failed to delete version", zap.Error(err))
		return fmt.Errorf("删除版本失败: %w", err)
	}

	logger.Info("version deleted",
		zap.String("project_key", projectKey),
		zap.Uint64("version_id", versionID),
	)

	return nil
}

// ListVersions 获取项目版本列表
func (s *fieldService) ListVersions(ctx context.Context, projectKey string) ([]*dto.VersionResponse, error) {
	project, err := s.projectRepo.GetByKey(ctx, strings.ToUpper(projectKey))
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrProjectNotFound
		}
		return nil, fmt.Errorf("查询项目失败: %w", err)
	}

	versions, err := s.versionRepo.ListByProject(ctx, project.ID)
	if err != nil {
		return nil, fmt.Errorf("查询版本列表失败: %w", err)
	}

	responses := make([]*dto.VersionResponse, len(versions))
	for i, v := range versions {
		responses[i] = s.toVersionResponse(v)
	}

	return responses, nil
}

// ============ 组件管理 ============

// CreateComponent 创建项目组件
func (s *fieldService) CreateComponent(ctx context.Context, projectKey string, req *dto.CreateComponentRequest) (*dto.ComponentResponse, error) {
	project, err := s.projectRepo.GetByKey(ctx, strings.ToUpper(projectKey))
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrProjectNotFound
		}
		return nil, fmt.Errorf("查询项目失败: %w", err)
	}

	// 检查组件名称是否已存在
	_, err = s.componentRepo.GetByName(ctx, project.ID, req.Name)
	if err == nil {
		return nil, ErrComponentNameExists
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, fmt.Errorf("检查组件名称失败: %w", err)
	}

	component := &model.ProjectComponent{
		ProjectID:   project.ID,
		Name:        req.Name,
		Description: req.Description,
		LeadUserID:  req.LeadUserID,
	}

	if err := s.componentRepo.Create(ctx, component); err != nil {
		logger.Error("failed to create component", zap.Error(err))
		return nil, fmt.Errorf("创建组件失败: %w", err)
	}

	logger.Info("component created",
		zap.String("project_key", projectKey),
		zap.String("component_name", req.Name),
	)

	return s.toComponentResponse(ctx, component), nil
}

// UpdateComponent 更新项目组件
func (s *fieldService) UpdateComponent(ctx context.Context, projectKey string, componentID uint64, req *dto.UpdateComponentRequest) (*dto.ComponentResponse, error) {
	project, err := s.projectRepo.GetByKey(ctx, strings.ToUpper(projectKey))
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrProjectNotFound
		}
		return nil, fmt.Errorf("查询项目失败: %w", err)
	}

	component, err := s.componentRepo.GetByID(ctx, componentID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrComponentNotFound
		}
		return nil, fmt.Errorf("查询组件失败: %w", err)
	}

	// 验证组件属于该项目
	if component.ProjectID != project.ID {
		return nil, ErrComponentNotFound
	}

	if req.Name != nil {
		// 检查新名称是否已存在
		existing, err := s.componentRepo.GetByName(ctx, project.ID, *req.Name)
		if err == nil && existing.ID != componentID {
			return nil, ErrComponentNameExists
		}
		component.Name = *req.Name
	}
	if req.Description != nil {
		component.Description = *req.Description
	}
	if req.LeadUserID != nil {
		component.LeadUserID = req.LeadUserID
	}

	if err := s.componentRepo.Update(ctx, component); err != nil {
		logger.Error("failed to update component", zap.Error(err))
		return nil, fmt.Errorf("更新组件失败: %w", err)
	}

	return s.toComponentResponse(ctx, component), nil
}

// DeleteComponent 删除项目组件
func (s *fieldService) DeleteComponent(ctx context.Context, projectKey string, componentID uint64) error {
	project, err := s.projectRepo.GetByKey(ctx, strings.ToUpper(projectKey))
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrProjectNotFound
		}
		return fmt.Errorf("查询项目失败: %w", err)
	}

	component, err := s.componentRepo.GetByID(ctx, componentID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrComponentNotFound
		}
		return fmt.Errorf("查询组件失败: %w", err)
	}

	// 验证组件属于该项目
	if component.ProjectID != project.ID {
		return ErrComponentNotFound
	}

	if err := s.componentRepo.Delete(ctx, componentID); err != nil {
		logger.Error("failed to delete component", zap.Error(err))
		return fmt.Errorf("删除组件失败: %w", err)
	}

	logger.Info("component deleted",
		zap.String("project_key", projectKey),
		zap.Uint64("component_id", componentID),
	)

	return nil
}

// ListComponents 获取项目组件列表
func (s *fieldService) ListComponents(ctx context.Context, projectKey string) ([]*dto.ComponentResponse, error) {
	project, err := s.projectRepo.GetByKey(ctx, strings.ToUpper(projectKey))
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrProjectNotFound
		}
		return nil, fmt.Errorf("查询项目失败: %w", err)
	}

	components, err := s.componentRepo.ListByProject(ctx, project.ID)
	if err != nil {
		return nil, fmt.Errorf("查询组件列表失败: %w", err)
	}

	responses := make([]*dto.ComponentResponse, len(components))
	for i, c := range components {
		responses[i] = s.toComponentResponse(ctx, c)
	}

	return responses, nil
}

// ============ 标签管理 ============

// CreateLabel 创建标签
func (s *fieldService) CreateLabel(ctx context.Context, projectKey string, req *dto.CreateLabelRequest) (*dto.LabelResponse, error) {
	project, err := s.projectRepo.GetByKey(ctx, strings.ToUpper(projectKey))
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrProjectNotFound
		}
		return nil, fmt.Errorf("查询项目失败: %w", err)
	}

	// 检查标签名称是否已存在
	_, err = s.labelRepo.GetByName(ctx, project.ID, req.Name)
	if err == nil {
		return nil, ErrLabelNameExists
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, fmt.Errorf("检查标签名称失败: %w", err)
	}

	label := &model.IssueLabel{
		ProjectID:   project.ID,
		Name:        req.Name,
		Color:       req.Color,
		Description: req.Description,
	}

	if err := s.labelRepo.Create(ctx, label); err != nil {
		logger.Error("failed to create label", zap.Error(err))
		return nil, fmt.Errorf("创建标签失败: %w", err)
	}

	logger.Info("label created",
		zap.String("project_key", projectKey),
		zap.String("label_name", req.Name),
	)

	return s.toLabelResponse(label), nil
}

// UpdateLabel 更新标签
func (s *fieldService) UpdateLabel(ctx context.Context, projectKey string, labelID uint64, req *dto.UpdateLabelRequest) (*dto.LabelResponse, error) {
	project, err := s.projectRepo.GetByKey(ctx, strings.ToUpper(projectKey))
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrProjectNotFound
		}
		return nil, fmt.Errorf("查询项目失败: %w", err)
	}

	label, err := s.labelRepo.GetByID(ctx, labelID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrLabelNotFound
		}
		return nil, fmt.Errorf("查询标签失败: %w", err)
	}

	// 验证标签属于该项目
	if label.ProjectID != project.ID {
		return nil, ErrLabelNotFound
	}

	if req.Name != nil {
		// 检查新名称是否已存在
		existing, err := s.labelRepo.GetByName(ctx, project.ID, *req.Name)
		if err == nil && existing.ID != labelID {
			return nil, ErrLabelNameExists
		}
		label.Name = *req.Name
	}
	if req.Color != nil {
		label.Color = *req.Color
	}
	if req.Description != nil {
		label.Description = *req.Description
	}

	if err := s.labelRepo.Update(ctx, label); err != nil {
		logger.Error("failed to update label", zap.Error(err))
		return nil, fmt.Errorf("更新标签失败: %w", err)
	}

	return s.toLabelResponse(label), nil
}

// DeleteLabel 删除标签
func (s *fieldService) DeleteLabel(ctx context.Context, projectKey string, labelID uint64) error {
	project, err := s.projectRepo.GetByKey(ctx, strings.ToUpper(projectKey))
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrProjectNotFound
		}
		return fmt.Errorf("查询项目失败: %w", err)
	}

	label, err := s.labelRepo.GetByID(ctx, labelID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrLabelNotFound
		}
		return fmt.Errorf("查询标签失败: %w", err)
	}

	// 验证标签属于该项目
	if label.ProjectID != project.ID {
		return ErrLabelNotFound
	}

	if err := s.labelRepo.Delete(ctx, labelID); err != nil {
		logger.Error("failed to delete label", zap.Error(err))
		return fmt.Errorf("删除标签失败: %w", err)
	}

	logger.Info("label deleted",
		zap.String("project_key", projectKey),
		zap.Uint64("label_id", labelID),
	)

	return nil
}

// ListLabels 获取项目标签列表
func (s *fieldService) ListLabels(ctx context.Context, projectKey string) ([]*dto.LabelResponse, error) {
	project, err := s.projectRepo.GetByKey(ctx, strings.ToUpper(projectKey))
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrProjectNotFound
		}
		return nil, fmt.Errorf("查询项目失败: %w", err)
	}

	labels, err := s.labelRepo.ListByProject(ctx, project.ID)
	if err != nil {
		return nil, fmt.Errorf("查询标签列表失败: %w", err)
	}

	responses := make([]*dto.LabelResponse, len(labels))
	for i, l := range labels {
		responses[i] = s.toLabelResponse(l)
	}

	return responses, nil
}

// ============ 转换方法 ============

func (s *fieldService) toFieldResponse(field *model.FieldDefinition) *dto.FieldDefinitionResponse {
	return &dto.FieldDefinitionResponse{
		ID:           field.ID,
		ProjectID:    field.ProjectID,
		FieldKey:     field.FieldKey,
		FieldName:    field.FieldName,
		FieldType:    field.FieldType,
		Description:  field.Description,
		IsSystem:     field.IsSystem,
		IsActive:     field.IsActive,
		Options:      field.Options,
		Validation:   field.Validation,
		DefaultValue: field.DefaultValue,
		SortOrder:    field.SortOrder,
		CreatedAt:    field.CreatedAt,
		UpdatedAt:    field.UpdatedAt,
	}
}

func (s *fieldService) toSchemeResponse(ctx context.Context, scheme *model.IssueTypeFieldScheme) *dto.FieldSchemeResponse {
	resp := &dto.FieldSchemeResponse{
		ID:              scheme.ID,
		ProjectID:       scheme.ProjectID,
		IssueTypeID:     scheme.IssueTypeID,
		FieldID:         scheme.FieldID,
		IsRequired:      scheme.IsRequired,
		IsVisibleCreate: scheme.IsVisibleCreate,
		IsVisibleEdit:   scheme.IsVisibleEdit,
		IsVisibleDetail: scheme.IsVisibleDetail,
		SortOrder:       scheme.SortOrder,
		DefaultValue:    scheme.DefaultValue,
		CreatedAt:       scheme.CreatedAt,
		UpdatedAt:       scheme.UpdatedAt,
	}

	// 获取字段定义
	if field, err := s.fieldRepo.GetByID(ctx, scheme.FieldID); err == nil {
		resp.Field = s.toFieldResponse(field)
	}

	return resp
}

func (s *fieldService) toVersionResponse(version *model.ProjectVersion) *dto.VersionResponse {
	return &dto.VersionResponse{
		ID:          version.ID,
		ProjectID:   version.ProjectID,
		Name:        version.Name,
		Description: version.Description,
		ReleaseDate: version.ReleaseDate,
		Status:      version.Status,
		SortOrder:   version.SortOrder,
		CreatedAt:   version.CreatedAt,
		UpdatedAt:   version.UpdatedAt,
	}
}

func (s *fieldService) toComponentResponse(ctx context.Context, component *model.ProjectComponent) *dto.ComponentResponse {
	resp := &dto.ComponentResponse{
		ID:          component.ID,
		ProjectID:   component.ProjectID,
		Name:        component.Name,
		Description: component.Description,
		LeadUserID:  component.LeadUserID,
		CreatedAt:   component.CreatedAt,
		UpdatedAt:   component.UpdatedAt,
	}

	// 获取负责人信息
	if component.LeadUserID != nil {
		if user, err := s.userRepo.GetByID(ctx, *component.LeadUserID); err == nil {
			resp.LeadUser = &dto.UserBrief{
				ID:          user.ID,
				Username:    user.Username,
				DisplayName: user.DisplayName,
				AvatarURL:   user.AvatarURL,
			}
		}
	}

	return resp
}

func (s *fieldService) toLabelResponse(label *model.IssueLabel) *dto.LabelResponse {
	return &dto.LabelResponse{
		ID:          label.ID,
		ProjectID:   label.ProjectID,
		Name:        label.Name,
		Color:       label.Color,
		Description: label.Description,
		CreatedAt:   label.CreatedAt,
		UpdatedAt:   label.UpdatedAt,
	}
}
