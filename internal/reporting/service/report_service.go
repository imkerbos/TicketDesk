// Package service 提供报表业务逻辑层
package service

import (
	"context"
	"time"

	"github.com/kerbos/ticketdesk/internal/reporting/dto"
	"github.com/kerbos/ticketdesk/internal/reporting/repository"
	projectRepo "github.com/kerbos/ticketdesk/internal/core-project/repository"
	"github.com/kerbos/ticketdesk/pkg/logger"
	"go.uber.org/zap"
)

// SLA 目标配置（分钟）
var slaTargets = map[string]int64{
	"P0": 60,   // 1 小时
	"P1": 240,  // 4 小时
	"P2": 1440, // 24 小时
	"P3": 4320, // 72 小时
}

// ReportService 报表服务接口
type ReportService interface {
	GetDashboardStats(ctx context.Context, req *dto.DashboardStatsRequest) (*dto.DashboardStatsResponse, error)
	GetIssueStats(ctx context.Context, req *dto.IssueStatsRequest) (*dto.IssueStatsResponse, error)
	GetSLAReport(ctx context.Context, req *dto.SLAReportRequest) (*dto.SLAReportResponse, error)
	GetAlertStats(ctx context.Context, req *dto.AlertStatsRequest) (*dto.AlertStatsResponse, error)
	GetUserPerformance(ctx context.Context, req *dto.IssueStatsRequest) ([]*dto.UserPerformanceResponse, error)
}

// reportService 报表服务实现
type reportService struct {
	reportRepo  repository.ReportRepository
	projectRepo projectRepo.ProjectRepository
}

// NewReportService 创建报表服务实例
func NewReportService(
	reportRepo repository.ReportRepository,
	projectRepo projectRepo.ProjectRepository,
) ReportService {
	return &reportService{
		reportRepo:  reportRepo,
		projectRepo: projectRepo,
	}
}

// GetDashboardStats 获取仪表盘统计数据
func (s *reportService) GetDashboardStats(ctx context.Context, req *dto.DashboardStatsRequest) (*dto.DashboardStatsResponse, error) {
	resp := &dto.DashboardStatsResponse{}

	var projectID *uint64
	if req.ProjectKey != "" {
		project, err := s.projectRepo.GetByKey(ctx, req.ProjectKey)
		if err != nil {
			return nil, err
		}
		projectID = &project.ID
	}

	// 工单统计
	issueStats, err := s.reportRepo.CountIssuesByStatus(ctx, projectID)
	if err != nil {
		logger.Error("failed to count issues by status", zap.Error(err))
		return nil, err
	}
	resp.IssueStats.TotalOpen = issueStats["open"]
	resp.IssueStats.TotalInProgress = issueStats["in_progress"]
	resp.IssueStats.TotalResolved = issueStats["resolved"]
	resp.IssueStats.TotalClosed = issueStats["closed"]

	// 今日和本周统计
	now := time.Now()
	todayStart := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	todayEnd := todayStart.Add(24 * time.Hour)
	weekStart := todayStart.AddDate(0, 0, -int(now.Weekday()))
	weekEnd := weekStart.AddDate(0, 0, 7)

	todayCreated, _ := s.reportRepo.CountIssuesCreatedByDate(ctx, projectID, todayStart, todayEnd)
	for _, item := range todayCreated {
		resp.IssueStats.TodayCreated += item.Count
	}

	weekCreated, _ := s.reportRepo.CountIssuesCreatedByDate(ctx, projectID, weekStart, weekEnd)
	for _, item := range weekCreated {
		resp.IssueStats.WeekCreated += item.Count
	}

	weekResolved, _ := s.reportRepo.CountIssuesResolvedByDate(ctx, projectID, weekStart, weekEnd)
	for _, item := range weekResolved {
		resp.IssueStats.WeekResolved += item.Count
	}

	// 告警统计
	alertStats, err := s.reportRepo.CountAlertsByStatus(ctx, projectID)
	if err != nil {
		logger.Warn("failed to count alerts by status", zap.Error(err))
	} else {
		resp.AlertStats.TotalFiring = alertStats["firing"]
		resp.AlertStats.TotalAcked = alertStats["acked"]
		resp.AlertStats.TotalResolved = alertStats["resolved"]
	}

	todayAlerts, _ := s.reportRepo.CountAlertsByDate(ctx, projectID, todayStart, todayEnd)
	for _, item := range todayAlerts {
		resp.AlertStats.TodayCreated += item.Count
	}

	weekAlerts, _ := s.reportRepo.CountAlertsByDate(ctx, projectID, weekStart, weekEnd)
	for _, item := range weekAlerts {
		resp.AlertStats.WeekCreated += item.Count
	}

	// 项目统计
	resp.ProjectStats.TotalProjects, _ = s.reportRepo.CountProjects(ctx)
	resp.ProjectStats.TotalMembers, _ = s.reportRepo.CountProjectMembers(ctx)

	return resp, nil
}

// GetIssueStats 获取工单统计数据
func (s *reportService) GetIssueStats(ctx context.Context, req *dto.IssueStatsRequest) (*dto.IssueStatsResponse, error) {
	resp := &dto.IssueStatsResponse{}

	// 解析日期范围
	dateRange := s.parseDateRange(req.StartDate, req.EndDate)

	var projectID *uint64
	if req.ProjectKey != "" {
		project, err := s.projectRepo.GetByKey(ctx, req.ProjectKey)
		if err != nil {
			return nil, err
		}
		projectID = &project.ID
	}

	// 获取状态分布
	statusMap, err := s.reportRepo.CountIssuesByStatus(ctx, projectID)
	if err != nil {
		return nil, err
	}
	resp.Summary.Open = statusMap["open"]
	resp.Summary.InProgress = statusMap["in_progress"]
	resp.Summary.Resolved = statusMap["resolved"]
	resp.Summary.Closed = statusMap["closed"]
	resp.Summary.Total = resp.Summary.Open + resp.Summary.InProgress + resp.Summary.Resolved + resp.Summary.Closed

	// 获取平均解决时间
	resp.Summary.AvgResolveTime, _ = s.reportRepo.GetAverageResolveTime(ctx, projectID, dateRange.StartDate, dateRange.EndDate)

	// 获取时间线数据
	createdByDate, _ := s.reportRepo.CountIssuesCreatedByDate(ctx, projectID, dateRange.StartDate, dateRange.EndDate)
	inProgressByDate, _ := s.reportRepo.CountIssuesInProgressByDate(ctx, projectID, dateRange.StartDate, dateRange.EndDate)
	resolvedByDate, _ := s.reportRepo.CountIssuesResolvedByDate(ctx, projectID, dateRange.StartDate, dateRange.EndDate)
	closedByDate, _ := s.reportRepo.CountIssuesClosedByDate(ctx, projectID, dateRange.StartDate, dateRange.EndDate)

	// 合并时间线数据
	resp.Timeline = s.mergeTimelineData(createdByDate, inProgressByDate, resolvedByDate, closedByDate)

	// 获取优先级分布
	priorityMap, _ := s.reportRepo.CountIssuesByPriority(ctx, projectID, dateRange.StartDate, dateRange.EndDate)
	resp.PriorityDistribution = s.mapToDistribution(priorityMap)

	// 获取类型分布
	typeMap, _ := s.reportRepo.CountIssuesByType(ctx, projectID, dateRange.StartDate, dateRange.EndDate)
	resp.TypeDistribution = s.mapToDistribution(typeMap)

	// 获取状态分布
	resp.StatusDistribution = s.mapToDistribution(statusMap)

	// 获取指派人分布
	assigneeMap, _ := s.reportRepo.CountIssuesByAssignee(ctx, projectID, dateRange.StartDate, dateRange.EndDate)
	resp.AssigneeDistribution = s.mapToDistribution(assigneeMap)

	// 获取 Epic 分布
	epicMap, _ := s.reportRepo.CountIssuesByEpic(ctx, projectID, dateRange.StartDate, dateRange.EndDate)
	resp.EpicDistribution = s.mapToDistribution(epicMap)

	return resp, nil
}

// GetSLAReport 获取 SLA 报表
func (s *reportService) GetSLAReport(ctx context.Context, req *dto.SLAReportRequest) (*dto.SLAReportResponse, error) {
	resp := &dto.SLAReportResponse{}

	// 解析日期
	startDate, _ := time.Parse("2006-01-02", req.StartDate)
	endDate, _ := time.Parse("2006-01-02", req.EndDate)
	endDate = endDate.Add(24*time.Hour - time.Second) // 包含结束日期当天

	var projectID *uint64
	if req.ProjectKey != "" {
		project, err := s.projectRepo.GetByKey(ctx, req.ProjectKey)
		if err != nil {
			return nil, err
		}
		projectID = &project.ID
	}

	// 获取 SLA 统计
	slaStats, err := s.reportRepo.GetSLAStats(ctx, projectID, startDate, endDate)
	if err != nil {
		return nil, err
	}

	resp.Summary.TotalIssues = slaStats.TotalIssues
	resp.Summary.ResolvedIssues = slaStats.ResolvedIssues
	resp.Summary.MTTR = slaStats.TotalMTTR

	// 按优先级统计
	priorityStats, _ := s.reportRepo.GetSLAStatsByPriority(ctx, projectID, startDate, endDate)
	resp.ByPriority = make([]dto.SLAPriorityStats, len(priorityStats))
	for i, ps := range priorityStats {
		slaTarget := slaTargets[ps.Priority]
		slaMet := int64(0)
		if ps.MTTR > 0 && ps.MTTR <= float64(slaTarget) {
			slaMet = ps.Resolved
		}
		slaRate := float64(0)
		if ps.Resolved > 0 {
			slaRate = float64(slaMet) / float64(ps.Resolved) * 100
		}

		resp.ByPriority[i] = dto.SLAPriorityStats{
			Priority:  ps.Priority,
			Total:     ps.Total,
			Resolved:  ps.Resolved,
			MTTR:      ps.MTTR,
			SLATarget: slaTarget,
			SLAMet:    slaMet,
			SLARate:   slaRate,
		}

		resp.Summary.SLAMet += slaMet
	}

	if resp.Summary.ResolvedIssues > 0 {
		resp.Summary.SLARate = float64(resp.Summary.SLAMet) / float64(resp.Summary.ResolvedIssues) * 100
	}
	resp.Summary.SLAViolated = resp.Summary.ResolvedIssues - resp.Summary.SLAMet

	return resp, nil
}

// GetAlertStats 获取告警统计
func (s *reportService) GetAlertStats(ctx context.Context, req *dto.AlertStatsRequest) (*dto.AlertStatsResponse, error) {
	resp := &dto.AlertStatsResponse{}

	dateRange := s.parseDateRange(req.StartDate, req.EndDate)

	var projectID *uint64
	if req.ProjectKey != "" {
		project, err := s.projectRepo.GetByKey(ctx, req.ProjectKey)
		if err != nil {
			return nil, err
		}
		projectID = &project.ID
	}

	// 获取状态分布
	statusMap, err := s.reportRepo.CountAlertsByStatus(ctx, projectID)
	if err != nil {
		return nil, err
	}
	resp.Summary.Firing = statusMap["firing"]
	resp.Summary.Acked = statusMap["acked"]
	resp.Summary.Resolved = statusMap["resolved"]
	resp.Summary.Total = resp.Summary.Firing + resp.Summary.Acked + resp.Summary.Resolved

	// 获取平均确认时间
	resp.Summary.AvgAckTime, _ = s.reportRepo.GetAverageAckTime(ctx, projectID, dateRange.StartDate, dateRange.EndDate)

	// 获取时间线数据
	alertsByDate, _ := s.reportRepo.CountAlertsByDate(ctx, projectID, dateRange.StartDate, dateRange.EndDate)
	resp.Timeline = make([]dto.AlertTimelineItem, len(alertsByDate))
	for i, item := range alertsByDate {
		resp.Timeline[i] = dto.AlertTimelineItem{
			Date:   item.Date,
			Firing: item.Count,
		}
	}

	// 获取严重程度分布
	severityMap, _ := s.reportRepo.CountAlertsBySeverity(ctx, projectID, dateRange.StartDate, dateRange.EndDate)
	resp.SeverityDistribution = s.mapToDistribution(severityMap)

	// 获取 Top 告警
	topAlerts, _ := s.reportRepo.GetTopAlerts(ctx, projectID, dateRange.StartDate, dateRange.EndDate, 10)
	resp.TopAlerts = make([]dto.TopAlertItem, len(topAlerts))
	for i, item := range topAlerts {
		resp.TopAlerts[i] = dto.TopAlertItem{
			AlertName: item.AlertName,
			Count:     item.Count,
		}
	}

	return resp, nil
}

// GetUserPerformance 获取用户绩效
func (s *reportService) GetUserPerformance(ctx context.Context, req *dto.IssueStatsRequest) ([]*dto.UserPerformanceResponse, error) {
	dateRange := s.parseDateRange(req.StartDate, req.EndDate)

	var projectID *uint64
	if req.ProjectKey != "" {
		project, err := s.projectRepo.GetByKey(ctx, req.ProjectKey)
		if err != nil {
			return nil, err
		}
		projectID = &project.ID
	}

	performance, err := s.reportRepo.GetUserPerformance(ctx, projectID, dateRange.StartDate, dateRange.EndDate)
	if err != nil {
		return nil, err
	}

	resp := make([]*dto.UserPerformanceResponse, len(performance))
	for i, p := range performance {
		resp[i] = &dto.UserPerformanceResponse{
			UserID:         p.UserID,
			Username:       p.Username,
			DisplayName:    p.DisplayName,
			Assigned:       p.Assigned,
			Resolved:       p.Resolved,
			AvgResolveTime: p.AvgResolveTime,
		}
	}

	return resp, nil
}

// parseDateRange 解析日期范围
func (s *reportService) parseDateRange(startDateStr, endDateStr string) dto.DateRange {
	now := time.Now()
	endDate := now

	// 默认最近 30 天
	startDate := now.AddDate(0, 0, -30)

	if startDateStr != "" {
		if parsed, err := time.Parse("2006-01-02", startDateStr); err == nil {
			startDate = parsed
		}
	}

	if endDateStr != "" {
		if parsed, err := time.Parse("2006-01-02", endDateStr); err == nil {
			endDate = parsed.Add(24*time.Hour - time.Second)
		}
	}

	return dto.DateRange{
		StartDate: startDate,
		EndDate:   endDate,
	}
}

// mergeTimelineData 合并时间线数据
func (s *reportService) mergeTimelineData(created, inProgress, resolved, closed []repository.DateCount) []dto.TimelineItem {
	dateMap := make(map[string]*dto.TimelineItem)

	for _, item := range created {
		if _, exists := dateMap[item.Date]; !exists {
			dateMap[item.Date] = &dto.TimelineItem{Date: item.Date}
		}
		dateMap[item.Date].Created = item.Count
	}

	for _, item := range inProgress {
		if _, exists := dateMap[item.Date]; !exists {
			dateMap[item.Date] = &dto.TimelineItem{Date: item.Date}
		}
		dateMap[item.Date].InProgress = item.Count
	}

	for _, item := range resolved {
		if _, exists := dateMap[item.Date]; !exists {
			dateMap[item.Date] = &dto.TimelineItem{Date: item.Date}
		}
		dateMap[item.Date].Resolved = item.Count
	}

	for _, item := range closed {
		if _, exists := dateMap[item.Date]; !exists {
			dateMap[item.Date] = &dto.TimelineItem{Date: item.Date}
		}
		dateMap[item.Date].Closed = item.Count
	}

	result := make([]dto.TimelineItem, 0, len(dateMap))
	for _, item := range dateMap {
		result = append(result, *item)
	}

	return result
}

// mapToDistribution 将 map 转换为分布数据
func (s *reportService) mapToDistribution(data map[string]int64) []dto.DistributionItem {
	var total int64
	for _, count := range data {
		total += count
	}

	result := make([]dto.DistributionItem, 0, len(data))
	for name, count := range data {
		ratio := float64(0)
		if total > 0 {
			ratio = float64(count) / float64(total) * 100
		}
		result = append(result, dto.DistributionItem{
			Name:  name,
			Value: count,
			Ratio: ratio,
		})
	}

	return result
}
