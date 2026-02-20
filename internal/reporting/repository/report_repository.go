// Package repository 提供报表数据访问层
package repository

import (
	"context"
	"time"

	"gorm.io/gorm"
)

// ReportRepository 报表数据访问接口
type ReportRepository interface {
	// 工单统计
	CountIssuesByStatus(ctx context.Context, projectID *uint64) (map[string]int64, error)
	CountIssuesByPriority(ctx context.Context, projectID *uint64, startDate, endDate time.Time) (map[string]int64, error)
	CountIssuesByType(ctx context.Context, projectID *uint64, startDate, endDate time.Time) (map[string]int64, error)
	CountIssuesByAssignee(ctx context.Context, projectID *uint64, startDate, endDate time.Time) (map[string]int64, error)
	CountIssuesByEpic(ctx context.Context, projectID *uint64, startDate, endDate time.Time) (map[string]int64, error)
	CountIssuesCreatedByDate(ctx context.Context, projectID *uint64, startDate, endDate time.Time) ([]DateCount, error)
	CountIssuesResolvedByDate(ctx context.Context, projectID *uint64, startDate, endDate time.Time) ([]DateCount, error)
	CountIssuesClosedByDate(ctx context.Context, projectID *uint64, startDate, endDate time.Time) ([]DateCount, error)
	GetAverageResolveTime(ctx context.Context, projectID *uint64, startDate, endDate time.Time) (float64, error)

	// SLA 统计
	GetSLAStats(ctx context.Context, projectID *uint64, startDate, endDate time.Time) (*SLAStats, error)
	GetSLAStatsByPriority(ctx context.Context, projectID *uint64, startDate, endDate time.Time) ([]PrioritySLAStats, error)

	// 告警统计
	CountAlertsByStatus(ctx context.Context, projectID *uint64) (map[string]int64, error)
	CountAlertsBySeverity(ctx context.Context, projectID *uint64, startDate, endDate time.Time) (map[string]int64, error)
	CountAlertsByDate(ctx context.Context, projectID *uint64, startDate, endDate time.Time) ([]DateCount, error)
	GetTopAlerts(ctx context.Context, projectID *uint64, startDate, endDate time.Time, limit int) ([]AlertCount, error)
	GetAverageAckTime(ctx context.Context, projectID *uint64, startDate, endDate time.Time) (float64, error)

	// 项目统计
	CountProjects(ctx context.Context) (int64, error)
	CountProjectMembers(ctx context.Context) (int64, error)

	// 用户绩效
	GetUserPerformance(ctx context.Context, projectID *uint64, startDate, endDate time.Time) ([]UserPerformance, error)
}

// DateCount 日期计数
type DateCount struct {
	Date  string `json:"date"`
	Count int64  `json:"count"`
}

// AlertCount 告警计数
type AlertCount struct {
	AlertName string `json:"alert_name"`
	Count     int64  `json:"count"`
}

// SLAStats SLA 统计数据
type SLAStats struct {
	TotalIssues    int64
	ResolvedIssues int64
	TotalMTTA      float64 // 平均确认时间（分钟）
	TotalMTTR      float64 // 平均解决时间（分钟）
}

// PrioritySLAStats 按优先级的 SLA 统计
type PrioritySLAStats struct {
	Priority string
	Total    int64
	Resolved int64
	MTTA     float64
	MTTR     float64
}

// UserPerformance 用户绩效数据
type UserPerformance struct {
	UserID         uint64
	Username       string
	DisplayName    string
	Assigned       int64
	Resolved       int64
	AvgResolveTime float64
}

// reportRepository 报表数据访问实现
type reportRepository struct {
	db *gorm.DB
}

// NewReportRepository 创建报表数据访问实例
func NewReportRepository(db *gorm.DB) ReportRepository {
	return &reportRepository{db: db}
}

// CountIssuesByStatus 按状态统计工单数量
func (r *reportRepository) CountIssuesByStatus(ctx context.Context, projectID *uint64) (map[string]int64, error) {
	type Result struct {
		Status string
		Count  int64
	}

	var results []Result
	query := r.db.WithContext(ctx).Table("issues").
		Select("status, COUNT(*) as count").
		Where("deleted_at IS NULL")

	if projectID != nil {
		query = query.Where("project_id = ?", *projectID)
	}

	err := query.Group("status").Find(&results).Error
	if err != nil {
		return nil, err
	}

	statusMap := make(map[string]int64)
	for _, res := range results {
		statusMap[res.Status] = res.Count
	}

	return statusMap, nil
}

// CountIssuesByPriority 按优先级统计工单数量
func (r *reportRepository) CountIssuesByPriority(ctx context.Context, projectID *uint64, startDate, endDate time.Time) (map[string]int64, error) {
	type Result struct {
		Priority string
		Count    int64
	}

	var results []Result
	query := r.db.WithContext(ctx).Table("issues").
		Select("priority, COUNT(*) as count").
		Where("deleted_at IS NULL").
		Where("created_at BETWEEN ? AND ?", startDate, endDate)

	if projectID != nil {
		query = query.Where("project_id = ?", *projectID)
	}

	err := query.Group("priority").Find(&results).Error
	if err != nil {
		return nil, err
	}

	priorityMap := make(map[string]int64)
	for _, res := range results {
		priorityMap[res.Priority] = res.Count
	}

	return priorityMap, nil
}

// CountIssuesByType 按类型统计工单数量
func (r *reportRepository) CountIssuesByType(ctx context.Context, projectID *uint64, startDate, endDate time.Time) (map[string]int64, error) {
	type Result struct {
		TypeName string
		Count    int64
	}

	var results []Result
	query := r.db.WithContext(ctx).Table("issues").
		Select("issue_types.display_name as type_name, COUNT(*) as count").
		Joins("LEFT JOIN issue_types ON issues.issue_type_id = issue_types.id").
		Where("issues.deleted_at IS NULL").
		Where("issues.created_at BETWEEN ? AND ?", startDate, endDate)

	if projectID != nil {
		query = query.Where("issues.project_id = ?", *projectID)
	}

	err := query.Group("issue_types.display_name").Find(&results).Error
	if err != nil {
		return nil, err
	}

	typeMap := make(map[string]int64)
	for _, res := range results {
		typeMap[res.TypeName] = res.Count
	}

	return typeMap, nil
}

// CountIssuesByAssignee 按指派人统计工单数量
func (r *reportRepository) CountIssuesByAssignee(ctx context.Context, projectID *uint64, startDate, endDate time.Time) (map[string]int64, error) {
	type Result struct {
		DisplayName string
		Count       int64
	}

	var results []Result
	query := r.db.WithContext(ctx).Table("issues").
		Select("COALESCE(users.display_name, users.username, '未指派') as display_name, COUNT(*) as count").
		Joins("LEFT JOIN users ON issues.assignee_id = users.id").
		Where("issues.deleted_at IS NULL").
		Where("issues.created_at BETWEEN ? AND ?", startDate, endDate)

	if projectID != nil {
		query = query.Where("issues.project_id = ?", *projectID)
	}

	err := query.Group("COALESCE(users.display_name, users.username, '未指派')").Find(&results).Error
	if err != nil {
		return nil, err
	}

	assigneeMap := make(map[string]int64)
	for _, res := range results {
		assigneeMap[res.DisplayName] = res.Count
	}

	return assigneeMap, nil
}

// CountIssuesByEpic 按 Epic 统计工单数量
func (r *reportRepository) CountIssuesByEpic(ctx context.Context, projectID *uint64, startDate, endDate time.Time) (map[string]int64, error) {
	type Result struct {
		EpicTitle string
		Count     int64
	}

	var results []Result
	query := r.db.WithContext(ctx).Table("issues").
		Select("COALESCE(epics.title, '无 Epic') as epic_title, COUNT(*) as count").
		Joins("LEFT JOIN issues AS epics ON issues.epic_id = epics.id").
		Where("issues.deleted_at IS NULL").
		Where("issues.created_at BETWEEN ? AND ?", startDate, endDate)

	if projectID != nil {
		query = query.Where("issues.project_id = ?", *projectID)
	}

	err := query.Group("COALESCE(epics.title, '无 Epic')").Find(&results).Error
	if err != nil {
		return nil, err
	}

	epicMap := make(map[string]int64)
	for _, res := range results {
		epicMap[res.EpicTitle] = res.Count
	}

	return epicMap, nil
}

// CountIssuesCreatedByDate 按日期统计创建的工单数量
func (r *reportRepository) CountIssuesCreatedByDate(ctx context.Context, projectID *uint64, startDate, endDate time.Time) ([]DateCount, error) {
	var results []DateCount
	query := r.db.WithContext(ctx).Table("issues").
		Select("DATE(created_at) as date, COUNT(*) as count").
		Where("deleted_at IS NULL").
		Where("created_at BETWEEN ? AND ?", startDate, endDate)

	if projectID != nil {
		query = query.Where("project_id = ?", *projectID)
	}

	err := query.Group("DATE(created_at)").Order("date").Find(&results).Error
	return results, err
}

// CountIssuesResolvedByDate 按日期统计解决的工单数量
func (r *reportRepository) CountIssuesResolvedByDate(ctx context.Context, projectID *uint64, startDate, endDate time.Time) ([]DateCount, error) {
	var results []DateCount
	query := r.db.WithContext(ctx).Table("issues").
		Select("DATE(resolved_at) as date, COUNT(*) as count").
		Where("deleted_at IS NULL").
		Where("resolved_at IS NOT NULL").
		Where("resolved_at BETWEEN ? AND ?", startDate, endDate)

	if projectID != nil {
		query = query.Where("project_id = ?", *projectID)
	}

	err := query.Group("DATE(resolved_at)").Order("date").Find(&results).Error
	return results, err
}

// CountIssuesClosedByDate 按日期统计关闭的工单数量
func (r *reportRepository) CountIssuesClosedByDate(ctx context.Context, projectID *uint64, startDate, endDate time.Time) ([]DateCount, error) {
	var results []DateCount
	query := r.db.WithContext(ctx).Table("issues").
		Select("DATE(closed_at) as date, COUNT(*) as count").
		Where("deleted_at IS NULL").
		Where("closed_at IS NOT NULL").
		Where("closed_at BETWEEN ? AND ?", startDate, endDate)

	if projectID != nil {
		query = query.Where("project_id = ?", *projectID)
	}

	err := query.Group("DATE(closed_at)").Order("date").Find(&results).Error
	return results, err
}

// GetAverageResolveTime 获取平均解决时间（小时）
func (r *reportRepository) GetAverageResolveTime(ctx context.Context, projectID *uint64, startDate, endDate time.Time) (float64, error) {
	var result struct {
		AvgTime float64
	}

	query := r.db.WithContext(ctx).Table("issues").
		Select("AVG(TIMESTAMPDIFF(HOUR, created_at, resolved_at)) as avg_time").
		Where("deleted_at IS NULL").
		Where("resolved_at IS NOT NULL").
		Where("created_at BETWEEN ? AND ?", startDate, endDate)

	if projectID != nil {
		query = query.Where("project_id = ?", *projectID)
	}

	err := query.Take(&result).Error
	if err != nil {
		return 0, err
	}

	return result.AvgTime, nil
}

// GetSLAStats 获取 SLA 统计数据
func (r *reportRepository) GetSLAStats(ctx context.Context, projectID *uint64, startDate, endDate time.Time) (*SLAStats, error) {
	var result struct {
		TotalIssues    int64
		ResolvedIssues int64
		AvgMTTR        float64
	}

	query := r.db.WithContext(ctx).Table("issues").
		Select(`
			COUNT(*) as total_issues,
			SUM(CASE WHEN resolved_at IS NOT NULL THEN 1 ELSE 0 END) as resolved_issues,
			AVG(CASE WHEN resolved_at IS NOT NULL THEN TIMESTAMPDIFF(MINUTE, created_at, resolved_at) ELSE NULL END) as avg_mttr
		`).
		Where("deleted_at IS NULL").
		Where("created_at BETWEEN ? AND ?", startDate, endDate)

	if projectID != nil {
		query = query.Where("project_id = ?", *projectID)
	}

	err := query.Take(&result).Error
	if err != nil {
		return nil, err
	}

	return &SLAStats{
		TotalIssues:    result.TotalIssues,
		ResolvedIssues: result.ResolvedIssues,
		TotalMTTR:      result.AvgMTTR,
	}, nil
}

// GetSLAStatsByPriority 按优先级获取 SLA 统计
func (r *reportRepository) GetSLAStatsByPriority(ctx context.Context, projectID *uint64, startDate, endDate time.Time) ([]PrioritySLAStats, error) {
	var results []struct {
		Priority string
		Total    int64
		Resolved int64
		AvgMTTR  float64
	}

	query := r.db.WithContext(ctx).Table("issues").
		Select(`
			priority,
			COUNT(*) as total,
			SUM(CASE WHEN resolved_at IS NOT NULL THEN 1 ELSE 0 END) as resolved,
			AVG(CASE WHEN resolved_at IS NOT NULL THEN TIMESTAMPDIFF(MINUTE, created_at, resolved_at) ELSE NULL END) as avg_mttr
		`).
		Where("deleted_at IS NULL").
		Where("created_at BETWEEN ? AND ?", startDate, endDate)

	if projectID != nil {
		query = query.Where("project_id = ?", *projectID)
	}

	err := query.Group("priority").Find(&results).Error
	if err != nil {
		return nil, err
	}

	stats := make([]PrioritySLAStats, len(results))
	for i, res := range results {
		stats[i] = PrioritySLAStats{
			Priority: res.Priority,
			Total:    res.Total,
			Resolved: res.Resolved,
			MTTR:     res.AvgMTTR,
		}
	}

	return stats, nil
}

// CountAlertsByStatus 按状态统计告警数量
func (r *reportRepository) CountAlertsByStatus(ctx context.Context, projectID *uint64) (map[string]int64, error) {
	type Result struct {
		Status string
		Count  int64
	}

	var results []Result
	query := r.db.WithContext(ctx).Table("alerts").
		Select("status, COUNT(*) as count").
		Where("deleted_at IS NULL")

	err := query.Group("status").Find(&results).Error
	if err != nil {
		return nil, err
	}

	statusMap := make(map[string]int64)
	for _, res := range results {
		statusMap[res.Status] = res.Count
	}

	return statusMap, nil
}

// CountAlertsBySeverity 按严重程度统计告警数量
func (r *reportRepository) CountAlertsBySeverity(ctx context.Context, projectID *uint64, startDate, endDate time.Time) (map[string]int64, error) {
	type Result struct {
		Severity string
		Count    int64
	}

	var results []Result
	query := r.db.WithContext(ctx).Table("alerts").
		Select("severity, COUNT(*) as count").
		Where("deleted_at IS NULL").
		Where("starts_at BETWEEN ? AND ?", startDate, endDate)

	err := query.Group("severity").Find(&results).Error
	if err != nil {
		return nil, err
	}

	severityMap := make(map[string]int64)
	for _, res := range results {
		severityMap[res.Severity] = res.Count
	}

	return severityMap, nil
}

// CountAlertsByDate 按日期统计告警数量
func (r *reportRepository) CountAlertsByDate(ctx context.Context, projectID *uint64, startDate, endDate time.Time) ([]DateCount, error) {
	var results []DateCount
	query := r.db.WithContext(ctx).Table("alerts").
		Select("DATE(starts_at) as date, COUNT(*) as count").
		Where("deleted_at IS NULL").
		Where("starts_at BETWEEN ? AND ?", startDate, endDate)

	err := query.Group("DATE(starts_at)").Order("date").Find(&results).Error
	return results, err
}

// GetTopAlerts 获取告警排名
func (r *reportRepository) GetTopAlerts(ctx context.Context, projectID *uint64, startDate, endDate time.Time, limit int) ([]AlertCount, error) {
	var results []AlertCount
	query := r.db.WithContext(ctx).Table("alerts").
		Select("alert_name, COUNT(*) as count").
		Where("deleted_at IS NULL").
		Where("starts_at BETWEEN ? AND ?", startDate, endDate)

	err := query.Group("alert_name").Order("count DESC").Limit(limit).Find(&results).Error
	return results, err
}

// GetAverageAckTime 获取平均确认时间（分钟）
func (r *reportRepository) GetAverageAckTime(ctx context.Context, projectID *uint64, startDate, endDate time.Time) (float64, error) {
	var result struct {
		AvgTime float64
	}

	query := r.db.WithContext(ctx).Table("alerts").
		Select("AVG(TIMESTAMPDIFF(MINUTE, starts_at, acked_at)) as avg_time").
		Where("deleted_at IS NULL").
		Where("acked_at IS NOT NULL").
		Where("starts_at BETWEEN ? AND ?", startDate, endDate)

	err := query.Take(&result).Error
	if err != nil {
		return 0, err
	}

	return result.AvgTime, nil
}

// CountProjects 统计项目总数
func (r *reportRepository) CountProjects(ctx context.Context) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Table("projects").
		Where("deleted_at IS NULL").
		Count(&count).Error
	return count, err
}

// CountProjectMembers 统计项目成员总数（去重）
func (r *reportRepository) CountProjectMembers(ctx context.Context) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Table("project_members").
		Select("COUNT(DISTINCT user_id)").
		Take(&count).Error
	return count, err
}

// GetUserPerformance 获取用户绩效
func (r *reportRepository) GetUserPerformance(ctx context.Context, projectID *uint64, startDate, endDate time.Time) ([]UserPerformance, error) {
	var results []struct {
		UserID         uint64
		Username       string
		DisplayName    string
		Assigned       int64
		Resolved       int64
		AvgResolveTime float64
	}

	query := r.db.WithContext(ctx).Table("issues").
		Select(`
			users.id as user_id,
			users.username,
			users.display_name,
			COUNT(*) as assigned,
			SUM(CASE WHEN issues.resolved_at IS NOT NULL THEN 1 ELSE 0 END) as resolved,
			AVG(CASE WHEN issues.resolved_at IS NOT NULL THEN TIMESTAMPDIFF(HOUR, issues.created_at, issues.resolved_at) ELSE NULL END) as avg_resolve_time
		`).
		Joins("JOIN users ON issues.assignee_id = users.id").
		Where("issues.deleted_at IS NULL").
		Where("issues.created_at BETWEEN ? AND ?", startDate, endDate)

	if projectID != nil {
		query = query.Where("issues.project_id = ?", *projectID)
	}

	err := query.Group("users.id, users.username, users.display_name").Find(&results).Error
	if err != nil {
		return nil, err
	}

	performance := make([]UserPerformance, len(results))
	for i, res := range results {
		performance[i] = UserPerformance{
			UserID:         res.UserID,
			Username:       res.Username,
			DisplayName:    res.DisplayName,
			Assigned:       res.Assigned,
			Resolved:       res.Resolved,
			AvgResolveTime: res.AvgResolveTime,
		}
	}

	return performance, nil
}
