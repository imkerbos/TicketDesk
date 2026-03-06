// Package repository 提供告警数据访问层
package repository

import (
	"context"
	"encoding/json"
	"sort"
	"strings"
	"time"

	"gorm.io/gorm"

	"github.com/kerbos/ticketdesk/internal/integration-alert/dto"
	"github.com/kerbos/ticketdesk/internal/model"
)

// AlertRepository 告警仓库接口
type AlertRepository interface {
	Create(ctx context.Context, alert *model.Alert) error
	GetByID(ctx context.Context, id uint64) (*model.Alert, error)
	GetByFingerprint(ctx context.Context, fingerprint string) (*model.Alert, error)
	Update(ctx context.Context, alert *model.Alert) error
	Delete(ctx context.Context, id uint64) error
	List(ctx context.Context, req *dto.AlertListRequest) ([]*model.Alert, int64, error)
	Stats(ctx context.Context) (*dto.AlertStatsResponse, error)
	UpdateStatus(ctx context.Context, id uint64, status string) error
	Ack(ctx context.Context, id, userID uint64, ackAt time.Time) error
	Resolve(ctx context.Context, id, userID uint64, resolvedAt time.Time) error
	ListByIssueID(ctx context.Context, issueID uint64) ([]*model.Alert, error)
	UpdateStatusByIssueID(ctx context.Context, issueID uint64, status string) error
	GroupBy(ctx context.Context, groupBy string, req *dto.AlertGroupRequest) ([]dto.AlertGroupItem, error)
	ListBySourceAndStatus(ctx context.Context, source, status string) ([]*model.Alert, error)
	ListLabelKeys(ctx context.Context) ([]string, error)
}

// alertRepository 告警仓库实现
type alertRepository struct {
	db *gorm.DB
}

// NewAlertRepository 创建告警仓库
func NewAlertRepository(db *gorm.DB) AlertRepository {
	return &alertRepository{db: db}
}

// Create 创建告警
func (r *alertRepository) Create(ctx context.Context, alert *model.Alert) error {
	return r.db.WithContext(ctx).Create(alert).Error
}

// GetByID 根据 ID 获取告警
func (r *alertRepository) GetByID(ctx context.Context, id uint64) (*model.Alert, error) {
	var alert model.Alert
	err := r.db.WithContext(ctx).First(&alert, id).Error
	if err != nil {
		return nil, err
	}
	return &alert, nil
}

// GetByFingerprint 根据指纹获取告警
func (r *alertRepository) GetByFingerprint(ctx context.Context, fingerprint string) (*model.Alert, error) {
	var alert model.Alert
	err := r.db.WithContext(ctx).Where("fingerprint = ?", fingerprint).First(&alert).Error
	if err != nil {
		return nil, err
	}
	return &alert, nil
}

// Update 更新告警
func (r *alertRepository) Update(ctx context.Context, alert *model.Alert) error {
	return r.db.WithContext(ctx).Save(alert).Error
}

// Delete 硬删除告警
func (r *alertRepository) Delete(ctx context.Context, id uint64) error {
	return r.db.WithContext(ctx).Unscoped().Delete(&model.Alert{}, id).Error
}

// List 获取告警列表
func (r *alertRepository) List(ctx context.Context, req *dto.AlertListRequest) ([]*model.Alert, int64, error) {
	var alerts []*model.Alert
	var total int64

	query := r.db.WithContext(ctx).Model(&model.Alert{})

	// 过滤条件
	if req.Status != "" {
		query = query.Where("status = ?", req.Status)
	}
	if req.Severity != "" {
		query = query.Where("severity = ?", req.Severity)
	}
	if req.Source != "" {
		query = query.Where("source = ?", req.Source)
	}
	if req.AlertName != "" {
		query = query.Where("alert_name LIKE ?", "%"+req.AlertName+"%")
	}
	if req.IssueID > 0 {
		query = query.Where("issue_id = ?", req.IssueID)
	}

	// 标签筛选：支持 key==value（等于）、key!=value（不等于）、key=~value（LIKE 模糊）、key!~value（NOT LIKE）
	if req.LabelFilters != "" {
		for _, filter := range strings.Split(req.LabelFilters, ",") {
			filter = strings.TrimSpace(filter)
			if filter == "" {
				continue
			}
			// 按操作符优先级匹配：!~ =~ != ==
			if idx := strings.Index(filter, "!~"); idx > 0 {
				key := strings.TrimSpace(filter[:idx])
				val := strings.TrimSpace(filter[idx+2:])
				query = query.Where("(JSON_UNQUOTE(JSON_EXTRACT(labels, ?)) IS NULL OR JSON_UNQUOTE(JSON_EXTRACT(labels, ?)) NOT LIKE ?)",
					"$."+key, "$."+key, val)
			} else if idx := strings.Index(filter, "=~"); idx > 0 {
				key := strings.TrimSpace(filter[:idx])
				val := strings.TrimSpace(filter[idx+2:])
				query = query.Where("JSON_UNQUOTE(JSON_EXTRACT(labels, ?)) LIKE ?", "$."+key, val)
			} else if idx := strings.Index(filter, "!="); idx > 0 {
				key := strings.TrimSpace(filter[:idx])
				val := strings.TrimSpace(filter[idx+2:])
				query = query.Where("(JSON_UNQUOTE(JSON_EXTRACT(labels, ?)) IS NULL OR JSON_UNQUOTE(JSON_EXTRACT(labels, ?)) != ?)",
					"$."+key, "$."+key, val)
			} else if idx := strings.Index(filter, "=="); idx > 0 {
				key := strings.TrimSpace(filter[:idx])
				val := strings.TrimSpace(filter[idx+2:])
				query = query.Where("JSON_UNQUOTE(JSON_EXTRACT(labels, ?)) = ?", "$."+key, val)
			}
		}
	}

	// 统计总数
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 分页
	page := req.Page
	if page < 1 {
		page = 1
	}
	pageSize := req.PageSize
	if pageSize < 1 {
		pageSize = 20
	}
	offset := (page - 1) * pageSize

	// 查询数据
	err := query.Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&alerts).Error
	return alerts, total, err
}

// Stats 获取告警统计数据
func (r *alertRepository) Stats(ctx context.Context) (*dto.AlertStatsResponse, error) {
	var stats dto.AlertStatsResponse
	err := r.db.WithContext(ctx).Model(&model.Alert{}).
		Select(`
			COUNT(*) as total,
			SUM(CASE WHEN status = 'firing' THEN 1 ELSE 0 END) as firing,
			SUM(CASE WHEN status = 'resolved' THEN 1 ELSE 0 END) as resolved,
			SUM(CASE WHEN severity = 'critical' THEN 1 ELSE 0 END) as critical,
			SUM(CASE WHEN severity = 'warning' THEN 1 ELSE 0 END) as warning,
			SUM(CASE WHEN severity = 'info' THEN 1 ELSE 0 END) as info
		`).
		Scan(&stats).Error
	if err != nil {
		return nil, err
	}
	return &stats, nil
}

// UpdateStatus 更新告警状态
func (r *alertRepository) UpdateStatus(ctx context.Context, id uint64, status string) error {
	return r.db.WithContext(ctx).Model(&model.Alert{}).Where("id = ?", id).Update("status", status).Error
}

// Ack 确认告警
func (r *alertRepository) Ack(ctx context.Context, id, userID uint64, ackAt time.Time) error {
	return r.db.WithContext(ctx).Model(&model.Alert{}).Where("id = ?", id).Updates(map[string]interface{}{
		"ack_at": ackAt,
		"ack_by": userID,
	}).Error
}

// Resolve 解决告警
func (r *alertRepository) Resolve(ctx context.Context, id, userID uint64, resolvedAt time.Time) error {
	return r.db.WithContext(ctx).Model(&model.Alert{}).Where("id = ?", id).Updates(map[string]interface{}{
		"status":      "resolved",
		"resolved_at": resolvedAt,
		"resolved_by": userID,
	}).Error
}

// ListByIssueID 根据工单 ID 获取关联的告警列表
func (r *alertRepository) ListByIssueID(ctx context.Context, issueID uint64) ([]*model.Alert, error) {
	var alerts []*model.Alert
	err := r.db.WithContext(ctx).Where("issue_id = ?", issueID).Find(&alerts).Error
	return alerts, err
}

// UpdateStatusByIssueID 根据工单 ID 批量更新告警状态
func (r *alertRepository) UpdateStatusByIssueID(ctx context.Context, issueID uint64, status string) error {
	updates := map[string]interface{}{
		"status": status,
	}
	if status == "resolved" {
		updates["resolved_at"] = time.Now()
	}
	return r.db.WithContext(ctx).Model(&model.Alert{}).Where("issue_id = ?", issueID).Updates(updates).Error
}

// ListBySourceAndStatus 根据来源和状态查询告警列表
func (r *alertRepository) ListBySourceAndStatus(ctx context.Context, source, status string) ([]*model.Alert, error) {
	var alerts []*model.Alert
	err := r.db.WithContext(ctx).Where("source = ? AND status = ?", source, status).Find(&alerts).Error
	return alerts, err
}

// ListLabelKeys 获取所有告警中出现过的标签 key（去重）
func (r *alertRepository) ListLabelKeys(ctx context.Context) ([]string, error) {
	var labels []string
	if err := r.db.WithContext(ctx).Model(&model.Alert{}).Pluck("labels", &labels).Error; err != nil {
		return nil, err
	}
	keySet := make(map[string]struct{})
	for _, l := range labels {
		var m map[string]string
		if json.Unmarshal([]byte(l), &m) == nil {
			for k := range m {
				keySet[k] = struct{}{}
			}
		}
	}
	result := make([]string, 0, len(keySet))
	for k := range keySet {
		result = append(result, k)
	}
	sort.Strings(result)
	return result, nil
}

// GroupBy 按标签分组统计告警
func (r *alertRepository) GroupBy(ctx context.Context, groupBy string, req *dto.AlertGroupRequest) ([]dto.AlertGroupItem, error) {
	var alerts []*model.Alert
	query := r.db.WithContext(ctx).Model(&model.Alert{})

	// 过滤条件
	if req.Status != "" {
		query = query.Where("status = ?", req.Status)
	}
	if req.Severity != "" {
		query = query.Where("severity = ?", req.Severity)
	}

	if err := query.Find(&alerts).Error; err != nil {
		return nil, err
	}

	// 按标签分组
	groupMap := make(map[string]*dto.AlertGroupItem)
	for _, alert := range alerts {
		labels := JSONToLabels(alert.Labels)
		groupValue := labels[groupBy]
		if groupValue == "" {
			groupValue = "<unknown>"
		}

		item, exists := groupMap[groupValue]
		if !exists {
			item = &dto.AlertGroupItem{
				GroupValue: groupValue,
				Count:      0,
				Severity:   make(map[string]int64),
				Status:     make(map[string]int64),
			}
			groupMap[groupValue] = item
		}

		item.Count++
		item.Severity[alert.Severity]++
		item.Status[alert.Status]++
	}

	// 转换为切片
	result := make([]dto.AlertGroupItem, 0, len(groupMap))
	for _, item := range groupMap {
		result = append(result, *item)
	}

	return result, nil
}

// ============ AlertSilenceRepository ============

// AlertSilenceRepository 告警静默仓库接口
type AlertSilenceRepository interface {
	Create(ctx context.Context, silence *model.AlertSilence) error
	GetByID(ctx context.Context, id uint64) (*model.AlertSilence, error)
	Update(ctx context.Context, silence *model.AlertSilence) error
	Delete(ctx context.Context, id uint64) error
	List(ctx context.Context, req *dto.AlertSilenceListRequest) ([]*model.AlertSilence, int64, error)
	ListActive(ctx context.Context) ([]*model.AlertSilence, error)
	Cancel(ctx context.Context, id uint64) error
}

// alertSilenceRepository 告警静默仓库实现
type alertSilenceRepository struct {
	db *gorm.DB
}

// NewAlertSilenceRepository 创建告警静默仓库
func NewAlertSilenceRepository(db *gorm.DB) AlertSilenceRepository {
	return &alertSilenceRepository{db: db}
}

// Create 创建告警静默
func (r *alertSilenceRepository) Create(ctx context.Context, silence *model.AlertSilence) error {
	return r.db.WithContext(ctx).Create(silence).Error
}

// GetByID 根据 ID 获取告警静默
func (r *alertSilenceRepository) GetByID(ctx context.Context, id uint64) (*model.AlertSilence, error) {
	var silence model.AlertSilence
	err := r.db.WithContext(ctx).First(&silence, id).Error
	if err != nil {
		return nil, err
	}
	return &silence, nil
}

// Update 更新告警静默
func (r *alertSilenceRepository) Update(ctx context.Context, silence *model.AlertSilence) error {
	return r.db.WithContext(ctx).Save(silence).Error
}

// Delete 硬删除告警静默
func (r *alertSilenceRepository) Delete(ctx context.Context, id uint64) error {
	return r.db.WithContext(ctx).Unscoped().Delete(&model.AlertSilence{}, id).Error
}

// List 获取告警静默列表
func (r *alertSilenceRepository) List(ctx context.Context, req *dto.AlertSilenceListRequest) ([]*model.AlertSilence, int64, error) {
	var silences []*model.AlertSilence
	var total int64

	query := r.db.WithContext(ctx).Model(&model.AlertSilence{})

	// 过滤条件
	if req.Status != nil {
		query = query.Where("status = ?", *req.Status)
	}

	// 统计总数
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 分页
	page := req.Page
	if page < 1 {
		page = 1
	}
	pageSize := req.PageSize
	if pageSize < 1 {
		pageSize = 20
	}
	offset := (page - 1) * pageSize

	// 查询数据
	err := query.Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&silences).Error
	return silences, total, err
}

// ListActive 获取所有生效中的告警静默
func (r *alertSilenceRepository) ListActive(ctx context.Context) ([]*model.AlertSilence, error) {
	var silences []*model.AlertSilence
	now := time.Now()
	err := r.db.WithContext(ctx).
		Where("status = ?", 1).
		Where("starts_at <= ?", now).
		Where("ends_at > ?", now).
		Find(&silences).Error
	return silences, err
}

// Cancel 取消告警静默
func (r *alertSilenceRepository) Cancel(ctx context.Context, id uint64) error {
	return r.db.WithContext(ctx).Model(&model.AlertSilence{}).Where("id = ?", id).Update("status", 0).Error
}

// ============ AlertRuleRepository ============

// AlertRuleRepository 告警规则仓库接口
type AlertRuleRepository interface {
	Create(ctx context.Context, rule *model.AlertRule) error
	GetByID(ctx context.Context, id uint64) (*model.AlertRule, error)
	Update(ctx context.Context, rule *model.AlertRule) error
	Delete(ctx context.Context, id uint64) error
	List(ctx context.Context, req *dto.AlertRuleListRequest) ([]*model.AlertRule, int64, error)
	ListEnabled(ctx context.Context) ([]*model.AlertRule, error)
}

// alertRuleRepository 告警规则仓库实现
type alertRuleRepository struct {
	db *gorm.DB
}

// NewAlertRuleRepository 创建告警规则仓库
func NewAlertRuleRepository(db *gorm.DB) AlertRuleRepository {
	return &alertRuleRepository{db: db}
}

// Create 创建告警规则
func (r *alertRuleRepository) Create(ctx context.Context, rule *model.AlertRule) error {
	return r.db.WithContext(ctx).Create(rule).Error
}

// GetByID 根据 ID 获取告警规则
func (r *alertRuleRepository) GetByID(ctx context.Context, id uint64) (*model.AlertRule, error) {
	var rule model.AlertRule
	err := r.db.WithContext(ctx).First(&rule, id).Error
	if err != nil {
		return nil, err
	}
	return &rule, nil
}

// Update 更新告警规则
func (r *alertRuleRepository) Update(ctx context.Context, rule *model.AlertRule) error {
	return r.db.WithContext(ctx).Save(rule).Error
}

// Delete 硬删除告警规则
func (r *alertRuleRepository) Delete(ctx context.Context, id uint64) error {
	return r.db.WithContext(ctx).Unscoped().Delete(&model.AlertRule{}, id).Error
}

// List 获取告警规则列表
func (r *alertRuleRepository) List(ctx context.Context, req *dto.AlertRuleListRequest) ([]*model.AlertRule, int64, error) {
	var rules []*model.AlertRule
	var total int64

	query := r.db.WithContext(ctx).Model(&model.AlertRule{})

	// 过滤条件
	if req.ProjectID > 0 {
		query = query.Where("project_id = ?", req.ProjectID)
	}
	if req.Status != nil {
		query = query.Where("status = ?", *req.Status)
	}

	// 统计总数
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 分页
	page := req.Page
	if page < 1 {
		page = 1
	}
	pageSize := req.PageSize
	if pageSize < 1 {
		pageSize = 20
	}
	offset := (page - 1) * pageSize

	// 查询数据
	err := query.Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&rules).Error
	return rules, total, err
}

// ListEnabled 获取所有启用的告警规则
func (r *alertRuleRepository) ListEnabled(ctx context.Context) ([]*model.AlertRule, error) {
	var rules []*model.AlertRule
	err := r.db.WithContext(ctx).Where("status = ?", 1).Find(&rules).Error
	return rules, err
}

// ============ 辅助函数 ============

// LabelsToJSON 将 map 转换为 JSON 字符串
func LabelsToJSON(labels map[string]string) string {
	if labels == nil {
		return "{}"
	}
	data, err := json.Marshal(labels)
	if err != nil {
		return "{}"
	}
	return string(data)
}

// JSONToLabels 将 JSON 字符串转换为 map
func JSONToLabels(jsonStr string) map[string]string {
	var labels map[string]string
	if err := json.Unmarshal([]byte(jsonStr), &labels); err != nil {
		return make(map[string]string)
	}
	return labels
}
