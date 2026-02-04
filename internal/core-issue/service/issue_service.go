// Package service 提供工单业务逻辑层
package service

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/kerbos/ticketdesk/internal/core-issue/dto"
	"github.com/kerbos/ticketdesk/internal/core-issue/repository"
	projectRepo "github.com/kerbos/ticketdesk/internal/core-project/repository"
	userRepo "github.com/kerbos/ticketdesk/internal/core-user/repository"
	"github.com/kerbos/ticketdesk/internal/model"
	"github.com/kerbos/ticketdesk/pkg/logger"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// 业务错误定义
var (
	ErrIssueNotFound      = errors.New("工单不存在")
	ErrProjectNotFound    = errors.New("项目不存在")
	ErrIssueTypeNotFound  = errors.New("工单类型不存在")
	ErrUserNotFound       = errors.New("用户不存在")
	ErrCommentNotFound    = errors.New("评论不存在")
	ErrAlreadyWatching    = errors.New("已经关注该工单")
	ErrNotWatching        = errors.New("未关注该工单")
	ErrInvalidTransition  = errors.New("无效的状态流转")
)

// IssueService 工单服务接口
type IssueService interface {
	CreateIssue(ctx context.Context, req *dto.CreateIssueRequest, reporterID uint64) (*dto.IssueResponse, error)
	GetIssue(ctx context.Context, key string) (*dto.IssueResponse, error)
	UpdateIssue(ctx context.Context, key string, req *dto.UpdateIssueRequest) (*dto.IssueResponse, error)
	DeleteIssue(ctx context.Context, key string) error
	ListIssues(ctx context.Context, req *dto.ListIssuesRequest) ([]*dto.IssueResponse, int64, error)
	TransitionIssue(ctx context.Context, key string, req *dto.TransitionIssueRequest, userID uint64) (*dto.IssueResponse, error)
	AssignIssue(ctx context.Context, key string, assigneeID uint64) (*dto.IssueResponse, error)

	// Dashboard 专用
	ListMyTodoIssues(ctx context.Context, userID uint64, page, pageSize int) ([]*dto.IssueResponse, int64, error)
	ListMyCreatedIssues(ctx context.Context, userID uint64, page, pageSize int) ([]*dto.IssueResponse, int64, error)

	// 评论
	AddComment(ctx context.Context, issueKey string, req *dto.CreateCommentRequest, userID uint64) (*dto.CommentResponse, error)
	ListComments(ctx context.Context, issueKey string) ([]*dto.CommentResponse, error)
	DeleteComment(ctx context.Context, commentID uint64, userID uint64) error

	// 关注
	AddWatcher(ctx context.Context, issueKey string, userID uint64) error
	RemoveWatcher(ctx context.Context, issueKey string, userID uint64) error
	ListWatchers(ctx context.Context, issueKey string) ([]*dto.UserBrief, error)
}

// issueService 工单服务实现
type issueService struct {
	issueRepo      repository.IssueRepository
	commentRepo    repository.CommentRepository
	watcherRepo    repository.WatcherRepository
	projectRepo    projectRepo.ProjectRepository
	issueTypeRepo  projectRepo.IssueTypeRepository
	userRepo       userRepo.UserRepository
	alertSyncSvc   AlertSyncService   // 告警同步服务（可选）
	activityLogger ActivityLogger     // 活动日志记录器（可选）
	notifSender    NotificationSender // 通知发送服务（可选）
}

// ActivityLogger 活动日志记录器接口（避免循环依赖）
type ActivityLogger interface {
	LogActivity(ctx context.Context, userID uint64, userName, action, entityType string, entityID uint64, entityKey, details string) error
}

// NotificationSender 通知发送接口（避免循环依赖）
type NotificationSender interface {
	CreateNotification(ctx context.Context, req *NotificationRequest) error
}

// NotificationRequest 通知请求（本地定义，避免依赖 notification-inbox/dto）
type NotificationRequest struct {
	UserID     uint64
	Type       string
	Title      string
	Content    string
	EntityType string
	EntityID   uint64
	EntityKey  string
	ActorID    uint64
	ActorName  string
}

// NewIssueService 创建工单服务实例
func NewIssueService(
	issueRepo repository.IssueRepository,
	commentRepo repository.CommentRepository,
	watcherRepo repository.WatcherRepository,
	projectRepo projectRepo.ProjectRepository,
	issueTypeRepo projectRepo.IssueTypeRepository,
	userRepo userRepo.UserRepository,
) IssueService {
	return &issueService{
		issueRepo:      issueRepo,
		commentRepo:    commentRepo,
		watcherRepo:    watcherRepo,
		projectRepo:    projectRepo,
		issueTypeRepo:  issueTypeRepo,
		userRepo:       userRepo,
		alertSyncSvc:   nil, // 默认为 nil，可通过 SetAlertSyncService 设置
		activityLogger: nil, // 默认为 nil，可通过 SetActivityLogger 设置
	}
}

// SetAlertSyncService 设置告警同步服务（用于避免循环依赖）
func (s *issueService) SetAlertSyncService(alertSyncSvc AlertSyncService) {
	s.alertSyncSvc = alertSyncSvc
}

// SetActivityLogger 设置活动日志记录器（用于避免循环依赖）
func (s *issueService) SetActivityLogger(activityLogger ActivityLogger) {
	s.activityLogger = activityLogger
}

// SetNotificationService 设置通知发送服务（用于避免循环依赖）
func (s *issueService) SetNotificationService(notifSender NotificationSender) {
	s.notifSender = notifSender
}

// sendNotification 发送通知（内部辅助方法，异步不阻塞主流程）
func (s *issueService) sendNotification(actorID uint64, actorName string, req *NotificationRequest) {
	if s.notifSender == nil {
		return
	}
	req.ActorID = actorID
	req.ActorName = actorName
	go func() {
		if err := s.notifSender.CreateNotification(context.Background(), req); err != nil {
			logger.Warn("failed to send notification", zap.Error(err))
		}
	}()
}

// notifyWatchers 通知所有关注者（排除指定用户）
func (s *issueService) notifyWatchers(issue *model.Issue, excludeUserID uint64, actorID uint64, actorName, notifType, title, content string) {
	if s.notifSender == nil {
		return
	}
	watchers, err := s.watcherRepo.ListByIssue(context.Background(), issue.ID)
	if err != nil {
		logger.Warn("failed to list watchers for notification", zap.Error(err))
		return
	}
	for _, w := range watchers {
		if w.UserID == excludeUserID {
			continue
		}
		s.sendNotification(actorID, actorName, &NotificationRequest{
			UserID:     w.UserID,
			Type:       notifType,
			Title:      title,
			Content:    content,
			EntityType: "issue",
			EntityID:   issue.ID,
			EntityKey:  issue.IssueKey,
		})
	}
}

// logActivity 记录活动日志（内部辅助方法）
func (s *issueService) logActivity(ctx context.Context, userID uint64, userName, action, entityKey, details string, entityID uint64) {
	if s.activityLogger == nil {
		return
	}
	// 异步记录，不阻塞主流程
	go func() {
		if err := s.activityLogger.LogActivity(context.Background(), userID, userName, action, "issue", entityID, entityKey, details); err != nil {
			logger.Warn("failed to log activity", zap.Error(err))
		}
	}()
}

// CreateIssue 创建工单
func (s *issueService) CreateIssue(ctx context.Context, req *dto.CreateIssueRequest, reporterID uint64) (*dto.IssueResponse, error) {
	// 获取项目
	project, err := s.projectRepo.GetByKey(ctx, strings.ToUpper(req.ProjectKey))
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrProjectNotFound
		}
		return nil, fmt.Errorf("查询项目失败: %w", err)
	}

	// 验证工单类型
	issueType, err := s.issueTypeRepo.GetByID(ctx, req.IssueTypeID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrIssueTypeNotFound
		}
		return nil, fmt.Errorf("查询工单类型失败: %w", err)
	}

	// 验证指派人
	if req.AssigneeID != nil {
		_, err := s.userRepo.GetByID(ctx, *req.AssigneeID)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil, fmt.Errorf("指派人不存在")
			}
			return nil, fmt.Errorf("查询指派人失败: %w", err)
		}
	}

	// 生成工单 Key
	nextNum, err := s.issueRepo.GetNextIssueNumber(ctx, project.ID)
	if err != nil {
		return nil, fmt.Errorf("生成工单编号失败: %w", err)
	}
	issueKey := repository.GenerateIssueKey(project.ProjectKey, nextNum)

	// 解析截止日期
	var dueDate *time.Time
	if req.DueDate != nil && *req.DueDate != "" {
		t, err := time.Parse("2006-01-02", *req.DueDate)
		if err != nil {
			return nil, fmt.Errorf("截止日期格式错误")
		}
		dueDate = &t
	}

	// 设置默认优先级
	priority := req.Priority
	if priority == "" {
		priority = "P2"
	}

	// 创建工单
	issue := &model.Issue{
		IssueKey:    issueKey,
		ProjectID:   project.ID,
		IssueTypeID: issueType.ID,
		Title:       req.Title,
		Description: req.Description,
		Priority:    priority,
		Status:      "open",
		ReporterID:  reporterID,
		AssigneeID:  req.AssigneeID,
		ParentID:    req.ParentID,
		DueDate:     dueDate,
	}

	if err := s.issueRepo.Create(ctx, issue); err != nil {
		logger.Error("failed to create issue", zap.Error(err))
		return nil, fmt.Errorf("创建工单失败: %w", err)
	}

	// 自动添加报告人为关注人
	watcher := &model.IssueWatcher{
		IssueID: issue.ID,
		UserID:  reporterID,
	}
	_ = s.watcherRepo.Create(ctx, watcher)

	// 记录活动日志
	reporter, _ := s.userRepo.GetByID(ctx, reporterID)
	reporterName := ""
	if reporter != nil {
		reporterName = reporter.DisplayName
		if reporterName == "" {
			reporterName = reporter.Username
		}
		s.logActivity(ctx, reporterID, reporter.Username, "创建了工单", issue.IssueKey, "", issue.ID)
	}

	// 通知：工单被指派
	if issue.AssigneeID != nil && *issue.AssigneeID != reporterID {
		s.sendNotification(reporterID, reporterName, &NotificationRequest{
			UserID:     *issue.AssigneeID,
			Type:       "issue_assigned",
			Title:      fmt.Sprintf("工单 %s 被指派给您", issue.IssueKey),
			Content:    issue.Title,
			EntityType: "issue",
			EntityID:   issue.ID,
			EntityKey:  issue.IssueKey,
		})
	}

	logger.Info("issue created successfully",
		zap.String("issue_key", issue.IssueKey),
		zap.Uint64("reporter_id", reporterID),
	)

	return s.toIssueResponse(ctx, issue, project.ProjectKey), nil
}

// GetIssue 获取工单详情
func (s *issueService) GetIssue(ctx context.Context, key string) (*dto.IssueResponse, error) {
	issue, err := s.issueRepo.GetByKey(ctx, strings.ToUpper(key))
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrIssueNotFound
		}
		return nil, fmt.Errorf("查询工单失败: %w", err)
	}

	project, _ := s.projectRepo.GetByID(ctx, issue.ProjectID)
	projectKey := ""
	if project != nil {
		projectKey = project.ProjectKey
	}

	resp := s.toIssueResponse(ctx, issue, projectKey)

	// 加载评论
	comments, _ := s.commentRepo.ListByIssue(ctx, issue.ID)
	resp.Comments = make([]*dto.CommentResponse, len(comments))
	for i, c := range comments {
		resp.Comments[i] = s.toCommentResponse(ctx, c)
	}

	// 加载关注人
	watchers, _ := s.watcherRepo.ListByIssue(ctx, issue.ID)
	resp.Watchers = make([]*dto.UserBrief, len(watchers))
	for i, w := range watchers {
		if user, err := s.userRepo.GetByID(ctx, w.UserID); err == nil {
			resp.Watchers[i] = &dto.UserBrief{
				ID:          user.ID,
				Username:    user.Username,
				DisplayName: user.DisplayName,
				AvatarURL:   user.AvatarURL,
			}
		}
	}

	return resp, nil
}

// UpdateIssue 更新工单
func (s *issueService) UpdateIssue(ctx context.Context, key string, req *dto.UpdateIssueRequest) (*dto.IssueResponse, error) {
	issue, err := s.issueRepo.GetByKey(ctx, strings.ToUpper(key))
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrIssueNotFound
		}
		return nil, fmt.Errorf("查询工单失败: %w", err)
	}

	// 更新字段
	if req.Title != nil {
		issue.Title = *req.Title
	}
	if req.Description != nil {
		issue.Description = *req.Description
	}
	if req.Priority != nil {
		issue.Priority = *req.Priority
	}
	if req.AssigneeID != nil {
		// 验证指派人
		_, err := s.userRepo.GetByID(ctx, *req.AssigneeID)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil, fmt.Errorf("指派人不存在")
			}
			return nil, fmt.Errorf("查询指派人失败: %w", err)
		}
		issue.AssigneeID = req.AssigneeID
	}
	if req.DueDate != nil {
		if *req.DueDate == "" {
			issue.DueDate = nil
		} else {
			t, err := time.Parse("2006-01-02", *req.DueDate)
			if err != nil {
				return nil, fmt.Errorf("截止日期格式错误")
			}
			issue.DueDate = &t
		}
	}
	if req.IssueTypeID != nil {
		_, err := s.issueTypeRepo.GetByID(ctx, *req.IssueTypeID)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil, ErrIssueTypeNotFound
			}
			return nil, fmt.Errorf("查询工单类型失败: %w", err)
		}
		issue.IssueTypeID = *req.IssueTypeID
	}

	if err := s.issueRepo.Update(ctx, issue); err != nil {
		logger.Error("failed to update issue", zap.String("key", key), zap.Error(err))
		return nil, fmt.Errorf("更新工单失败: %w", err)
	}

	project, _ := s.projectRepo.GetByID(ctx, issue.ProjectID)
	projectKey := ""
	if project != nil {
		projectKey = project.ProjectKey
	}

	logger.Info("issue updated successfully", zap.String("issue_key", key))

	return s.toIssueResponse(ctx, issue, projectKey), nil
}

// DeleteIssue 删除工单
func (s *issueService) DeleteIssue(ctx context.Context, key string) error {
	issue, err := s.issueRepo.GetByKey(ctx, strings.ToUpper(key))
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrIssueNotFound
		}
		return fmt.Errorf("查询工单失败: %w", err)
	}

	if err := s.issueRepo.Delete(ctx, issue.ID); err != nil {
		logger.Error("failed to delete issue", zap.String("key", key), zap.Error(err))
		return fmt.Errorf("删除工单失败: %w", err)
	}

	logger.Info("issue deleted successfully", zap.String("issue_key", key))

	return nil
}

// ListIssues 分页查询工单列表
func (s *issueService) ListIssues(ctx context.Context, req *dto.ListIssuesRequest) ([]*dto.IssueResponse, int64, error) {
	page := req.GetDefaultPage()
	pageSize := req.GetDefaultPageSize()
	offset := (page - 1) * pageSize

	filter := &repository.IssueFilter{
		Status:      req.Status,
		Priority:    req.Priority,
		AssigneeID:  req.AssigneeID,
		ReporterID:  req.ReporterID,
		IssueTypeID: req.IssueTypeID,
		Keyword:     req.Keyword,
	}

	// 如果指定了项目 Key，获取项目 ID
	if req.ProjectKey != "" {
		project, err := s.projectRepo.GetByKey(ctx, strings.ToUpper(req.ProjectKey))
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil, 0, ErrProjectNotFound
			}
			return nil, 0, fmt.Errorf("查询项目失败: %w", err)
		}
		filter.ProjectID = &project.ID
	}

	issues, total, err := s.issueRepo.List(ctx, filter, offset, pageSize)
	if err != nil {
		logger.Error("failed to list issues", zap.Error(err))
		return nil, 0, fmt.Errorf("查询工单列表失败: %w", err)
	}

	// 缓存项目信息
	projectCache := make(map[uint64]*model.Project)

	responses := make([]*dto.IssueResponse, len(issues))
	for i, issue := range issues {
		projectKey := ""
		if project, ok := projectCache[issue.ProjectID]; ok {
			projectKey = project.ProjectKey
		} else if project, err := s.projectRepo.GetByID(ctx, issue.ProjectID); err == nil {
			projectCache[issue.ProjectID] = project
			projectKey = project.ProjectKey
		}
		responses[i] = s.toIssueResponse(ctx, issue, projectKey)
	}

	return responses, total, nil
}

// ListMyTodoIssues 获取我的待办工单（指派给我的待处理/进行中工单）
func (s *issueService) ListMyTodoIssues(ctx context.Context, userID uint64, page, pageSize int) ([]*dto.IssueResponse, int64, error) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}
	offset := (page - 1) * pageSize

	filter := &repository.IssueFilter{
		AssigneeID:   &userID,
		StatusNotIn:  []string{"closed", "resolved"},
	}

	issues, total, err := s.issueRepo.List(ctx, filter, offset, pageSize)
	if err != nil {
		logger.Error("failed to list my todo issues", zap.Error(err))
		return nil, 0, fmt.Errorf("查询我的待办工单失败: %w", err)
	}

	// 缓存项目信息
	projectCache := make(map[uint64]*model.Project)

	responses := make([]*dto.IssueResponse, len(issues))
	for i, issue := range issues {
		projectKey := ""
		if project, ok := projectCache[issue.ProjectID]; ok {
			projectKey = project.ProjectKey
		} else if project, err := s.projectRepo.GetByID(ctx, issue.ProjectID); err == nil {
			projectCache[issue.ProjectID] = project
			projectKey = project.ProjectKey
		}
		responses[i] = s.toIssueResponse(ctx, issue, projectKey)
	}

	return responses, total, nil
}

// ListMyCreatedIssues 获取我创建的工单
func (s *issueService) ListMyCreatedIssues(ctx context.Context, userID uint64, page, pageSize int) ([]*dto.IssueResponse, int64, error) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}
	offset := (page - 1) * pageSize

	filter := &repository.IssueFilter{
		ReporterID: &userID,
	}

	issues, total, err := s.issueRepo.List(ctx, filter, offset, pageSize)
	if err != nil {
		logger.Error("failed to list my created issues", zap.Error(err))
		return nil, 0, fmt.Errorf("查询我创建的工单失败: %w", err)
	}

	// 缓存项目信息
	projectCache := make(map[uint64]*model.Project)

	responses := make([]*dto.IssueResponse, len(issues))
	for i, issue := range issues {
		projectKey := ""
		if project, ok := projectCache[issue.ProjectID]; ok {
			projectKey = project.ProjectKey
		} else if project, err := s.projectRepo.GetByID(ctx, issue.ProjectID); err == nil {
			projectCache[issue.ProjectID] = project
			projectKey = project.ProjectKey
		}
		responses[i] = s.toIssueResponse(ctx, issue, projectKey)
	}

	return responses, total, nil
}

// TransitionIssue 工单状态流转
func (s *issueService) TransitionIssue(ctx context.Context, key string, req *dto.TransitionIssueRequest, userID uint64) (*dto.IssueResponse, error) {
	issue, err := s.issueRepo.GetByKey(ctx, strings.ToUpper(key))
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrIssueNotFound
		}
		return nil, fmt.Errorf("查询工单失败: %w", err)
	}

	// 验证状态流转
	if !isValidTransition(issue.Status, req.Status) {
		return nil, ErrInvalidTransition
	}

	now := time.Now()
	issue.Status = req.Status

	// 更新相关时间字段
	switch req.Status {
	case "resolved":
		issue.ResolvedAt = &now
	case "closed":
		issue.ClosedAt = &now
	case "reopened":
		issue.Status = "open"
		issue.ResolvedAt = nil
		issue.ClosedAt = nil
	}

	if err := s.issueRepo.Update(ctx, issue); err != nil {
		return nil, fmt.Errorf("更新工单状态失败: %w", err)
	}

	// 同步告警状态
	if s.alertSyncSvc != nil {
		if err := s.alertSyncSvc.SyncIssueStatus(ctx, issue.ID, issue.Status); err != nil {
			logger.Error("failed to sync alert status",
				zap.Uint64("issue_id", issue.ID),
				zap.String("issue_status", issue.Status),
				zap.Error(err),
			)
			// 不中断工单状态更新流程
		}
	}

	// 添加状态变更评论
	if req.Comment != "" {
		comment := &model.IssueComment{
			IssueID: issue.ID,
			UserID:  userID,
			Content: fmt.Sprintf("[状态变更为 %s] %s", req.Status, req.Comment),
		}
		_ = s.commentRepo.Create(ctx, comment)
	}

	// 记录活动日志
	user, _ := s.userRepo.GetByID(ctx, userID)
	userName := ""
	if user != nil {
		userName = user.DisplayName
		if userName == "" {
			userName = user.Username
		}
		actionText := s.getStatusActionText(req.Status)
		s.logActivity(ctx, userID, user.Username, actionText, issue.IssueKey, "", issue.ID)
	}

	// 通知：关注者和创建者
	statusText := s.getStatusActionText(req.Status)
	notifTitle := fmt.Sprintf("工单 %s %s", issue.IssueKey, statusText)
	s.notifyWatchers(issue, userID, userID, userName, "issue_status_changed", notifTitle, issue.Title)

	// 通知创建者（如果不是操作人且不在关注者中）
	if issue.ReporterID != userID {
		s.sendNotification(userID, userName, &NotificationRequest{
			UserID:     issue.ReporterID,
			Type:       "issue_status_changed",
			Title:      fmt.Sprintf("您创建的工单 %s %s", issue.IssueKey, statusText),
			Content:    issue.Title,
			EntityType: "issue",
			EntityID:   issue.ID,
			EntityKey:  issue.IssueKey,
		})
	}

	project, _ := s.projectRepo.GetByID(ctx, issue.ProjectID)
	projectKey := ""
	if project != nil {
		projectKey = project.ProjectKey
	}

	logger.Info("issue transitioned successfully",
		zap.String("issue_key", key),
		zap.String("new_status", req.Status),
	)

	return s.toIssueResponse(ctx, issue, projectKey), nil
}

// AssignIssue 指派工单
func (s *issueService) AssignIssue(ctx context.Context, key string, assigneeID uint64) (*dto.IssueResponse, error) {
	issue, err := s.issueRepo.GetByKey(ctx, strings.ToUpper(key))
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrIssueNotFound
		}
		return nil, fmt.Errorf("查询工单失败: %w", err)
	}

	// 验证指派人
	_, err = s.userRepo.GetByID(ctx, assigneeID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, fmt.Errorf("查询用户失败: %w", err)
	}

	oldAssigneeID := issue.AssigneeID
	issue.AssigneeID = &assigneeID
	if err := s.issueRepo.Update(ctx, issue); err != nil {
		return nil, fmt.Errorf("指派工单失败: %w", err)
	}

	// 通知：新指派人（如果指派人有变化）
	if oldAssigneeID == nil || *oldAssigneeID != assigneeID {
		// 从上下文获取操作人信息
		actorID, _ := ctx.Value("user_id").(uint64)
		actorName := ""
		if actor, err := s.userRepo.GetByID(ctx, actorID); err == nil {
			actorName = actor.DisplayName
			if actorName == "" {
				actorName = actor.Username
			}
		}
		s.sendNotification(actorID, actorName, &NotificationRequest{
			UserID:     assigneeID,
			Type:       "issue_assigned",
			Title:      fmt.Sprintf("工单 %s 被指派给您", issue.IssueKey),
			Content:    issue.Title,
			EntityType: "issue",
			EntityID:   issue.ID,
			EntityKey:  issue.IssueKey,
		})

		// 通知创建者
		if issue.ReporterID != actorID && issue.ReporterID != assigneeID {
			s.sendNotification(actorID, actorName, &NotificationRequest{
				UserID:     issue.ReporterID,
				Type:       "issue_updated",
				Title:      fmt.Sprintf("您创建的工单 %s 被重新指派", issue.IssueKey),
				Content:    issue.Title,
				EntityType: "issue",
				EntityID:   issue.ID,
				EntityKey:  issue.IssueKey,
			})
		}
	}

	project, _ := s.projectRepo.GetByID(ctx, issue.ProjectID)
	projectKey := ""
	if project != nil {
		projectKey = project.ProjectKey
	}

	return s.toIssueResponse(ctx, issue, projectKey), nil
}

// AddComment 添加评论
func (s *issueService) AddComment(ctx context.Context, issueKey string, req *dto.CreateCommentRequest, userID uint64) (*dto.CommentResponse, error) {
	issue, err := s.issueRepo.GetByKey(ctx, strings.ToUpper(issueKey))
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrIssueNotFound
		}
		return nil, fmt.Errorf("查询工单失败: %w", err)
	}

	comment := &model.IssueComment{
		IssueID: issue.ID,
		UserID:  userID,
		Content: req.Content,
	}

	if err := s.commentRepo.Create(ctx, comment); err != nil {
		return nil, fmt.Errorf("添加评论失败: %w", err)
	}

	// 记录活动日志
	user, _ := s.userRepo.GetByID(ctx, userID)
	userName := ""
	if user != nil {
		userName = user.DisplayName
		if userName == "" {
			userName = user.Username
		}
		s.logActivity(ctx, userID, user.Username, "评论了工单", issue.IssueKey, "", issue.ID)
	}

	// 通知关注者
	notifTitle := fmt.Sprintf("工单 %s 有新评论", issue.IssueKey)
	s.notifyWatchers(issue, userID, userID, userName, "issue_commented", notifTitle, req.Content)

	// 通知创建者（如果不是评论人）
	if issue.ReporterID != userID {
		s.sendNotification(userID, userName, &NotificationRequest{
			UserID:     issue.ReporterID,
			Type:       "issue_commented",
			Title:      fmt.Sprintf("您创建的工单 %s 有新评论", issue.IssueKey),
			Content:    req.Content,
			EntityType: "issue",
			EntityID:   issue.ID,
			EntityKey:  issue.IssueKey,
		})
	}

	// 解析 @提及
	mentions := extractMentions(req.Content)
	notifiedUsers := make(map[uint64]bool)
	for _, username := range mentions {
		mentionedUser, err := s.userRepo.GetByUsername(ctx, username)
		if err != nil || mentionedUser == nil {
			continue
		}
		// 不重复通知自己和已通知的用户
		if mentionedUser.ID == userID || notifiedUsers[mentionedUser.ID] {
			continue
		}
		notifiedUsers[mentionedUser.ID] = true
		s.sendNotification(userID, userName, &NotificationRequest{
			UserID:     mentionedUser.ID,
			Type:       "mention",
			Title:      fmt.Sprintf("%s 在工单 %s 中提及了您", userName, issue.IssueKey),
			Content:    req.Content,
			EntityType: "issue",
			EntityID:   issue.ID,
			EntityKey:  issue.IssueKey,
		})
	}

	return s.toCommentResponse(ctx, comment), nil
}

// ListComments 获取评论列表
func (s *issueService) ListComments(ctx context.Context, issueKey string) ([]*dto.CommentResponse, error) {
	issue, err := s.issueRepo.GetByKey(ctx, strings.ToUpper(issueKey))
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrIssueNotFound
		}
		return nil, fmt.Errorf("查询工单失败: %w", err)
	}

	comments, err := s.commentRepo.ListByIssue(ctx, issue.ID)
	if err != nil {
		return nil, fmt.Errorf("查询评论失败: %w", err)
	}

	responses := make([]*dto.CommentResponse, len(comments))
	for i, c := range comments {
		responses[i] = s.toCommentResponse(ctx, c)
	}

	return responses, nil
}

// DeleteComment 删除评论
func (s *issueService) DeleteComment(ctx context.Context, commentID uint64, userID uint64) error {
	comment, err := s.commentRepo.GetByID(ctx, commentID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrCommentNotFound
		}
		return fmt.Errorf("查询评论失败: %w", err)
	}

	// 只有评论作者可以删除
	if comment.UserID != userID {
		return fmt.Errorf("只能删除自己的评论")
	}

	return s.commentRepo.Delete(ctx, commentID)
}

// AddWatcher 添加关注人
func (s *issueService) AddWatcher(ctx context.Context, issueKey string, userID uint64) error {
	issue, err := s.issueRepo.GetByKey(ctx, strings.ToUpper(issueKey))
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrIssueNotFound
		}
		return fmt.Errorf("查询工单失败: %w", err)
	}

	// 检查是否已关注
	watching, err := s.watcherRepo.IsWatching(ctx, issue.ID, userID)
	if err != nil {
		return fmt.Errorf("检查关注状态失败: %w", err)
	}
	if watching {
		return ErrAlreadyWatching
	}

	watcher := &model.IssueWatcher{
		IssueID: issue.ID,
		UserID:  userID,
	}

	return s.watcherRepo.Create(ctx, watcher)
}

// RemoveWatcher 移除关注人
func (s *issueService) RemoveWatcher(ctx context.Context, issueKey string, userID uint64) error {
	issue, err := s.issueRepo.GetByKey(ctx, strings.ToUpper(issueKey))
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrIssueNotFound
		}
		return fmt.Errorf("查询工单失败: %w", err)
	}

	return s.watcherRepo.DeleteByIssueAndUser(ctx, issue.ID, userID)
}

// ListWatchers 获取关注人列表
func (s *issueService) ListWatchers(ctx context.Context, issueKey string) ([]*dto.UserBrief, error) {
	issue, err := s.issueRepo.GetByKey(ctx, strings.ToUpper(issueKey))
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrIssueNotFound
		}
		return nil, fmt.Errorf("查询工单失败: %w", err)
	}

	watchers, err := s.watcherRepo.ListByIssue(ctx, issue.ID)
	if err != nil {
		return nil, fmt.Errorf("查询关注人失败: %w", err)
	}

	responses := make([]*dto.UserBrief, 0, len(watchers))
	for _, w := range watchers {
		if user, err := s.userRepo.GetByID(ctx, w.UserID); err == nil {
			responses = append(responses, &dto.UserBrief{
				ID:          user.ID,
				Username:    user.Username,
				DisplayName: user.DisplayName,
				AvatarURL:   user.AvatarURL,
			})
		}
	}

	return responses, nil
}

// isValidTransition 验证状态流转是否有效
func isValidTransition(from, to string) bool {
	validTransitions := map[string][]string{
		"open":        {"in_progress", "resolved", "closed"},
		"in_progress": {"open", "resolved", "closed"},
		"resolved":    {"closed", "reopened"},
		"closed":      {"reopened"},
	}

	allowed, ok := validTransitions[from]
	if !ok {
		return false
	}

	for _, s := range allowed {
		if s == to {
			return true
		}
	}
	return false
}

// getStatusActionText 获取状态对应的动作文本
func (s *issueService) getStatusActionText(status string) string {
	statusActions := map[string]string{
		"open":        "打开了工单",
		"in_progress": "开始处理工单",
		"resolved":    "解决了工单",
		"closed":      "关闭了工单",
		"reopened":    "重新打开了工单",
	}
	if action, ok := statusActions[status]; ok {
		return action
	}
	return "更新了工单状态"
}

// mentionRegex 匹配 @username 格式的提及
var mentionRegex = regexp.MustCompile(`@(\w+)`)

// extractMentions 从文本中提取 @提及的用户名
func extractMentions(content string) []string {
	matches := mentionRegex.FindAllStringSubmatch(content, -1)
	seen := make(map[string]bool)
	var usernames []string
	for _, match := range matches {
		if len(match) >= 2 && !seen[match[1]] {
			seen[match[1]] = true
			usernames = append(usernames, match[1])
		}
	}
	return usernames
}

// toIssueResponse 将工单模型转换为响应 DTO
func (s *issueService) toIssueResponse(ctx context.Context, issue *model.Issue, projectKey string) *dto.IssueResponse {
	resp := &dto.IssueResponse{
		ID:          issue.ID,
		IssueKey:    issue.IssueKey,
		ProjectID:   issue.ProjectID,
		ProjectKey:  projectKey,
		IssueTypeID: issue.IssueTypeID,
		Title:       issue.Title,
		Description: issue.Description,
		Priority:    issue.Priority,
		Status:      issue.Status,
		ReporterID:  issue.ReporterID,
		AssigneeID:  issue.AssigneeID,
		ParentID:    issue.ParentID,
		DueDate:     issue.DueDate,
		ResolvedAt:  issue.ResolvedAt,
		ClosedAt:    issue.ClosedAt,
		CreatedAt:   issue.CreatedAt,
		UpdatedAt:   issue.UpdatedAt,
	}

	// 加载工单类型
	if issueType, err := s.issueTypeRepo.GetByID(ctx, issue.IssueTypeID); err == nil {
		resp.IssueType = &dto.IssueTypeBrief{
			ID:          issueType.ID,
			Name:        issueType.Name,
			DisplayName: issueType.DisplayName,
			Icon:        issueType.Icon,
			Color:       issueType.Color,
		}
	}

	// 加载报告人
	if reporter, err := s.userRepo.GetByID(ctx, issue.ReporterID); err == nil {
		resp.Reporter = &dto.UserBrief{
			ID:          reporter.ID,
			Username:    reporter.Username,
			DisplayName: reporter.DisplayName,
			AvatarURL:   reporter.AvatarURL,
		}
	}

	// 加载指派人
	if issue.AssigneeID != nil {
		if assignee, err := s.userRepo.GetByID(ctx, *issue.AssigneeID); err == nil {
			resp.Assignee = &dto.UserBrief{
				ID:          assignee.ID,
				Username:    assignee.Username,
				DisplayName: assignee.DisplayName,
				AvatarURL:   assignee.AvatarURL,
			}
		}
	}

	// 加载父工单 Key
	if issue.ParentID != nil {
		if parent, err := s.issueRepo.GetByID(ctx, *issue.ParentID); err == nil {
			resp.ParentKey = parent.IssueKey
		}
	}

	return resp
}

// toCommentResponse 将评论模型转换为响应 DTO
func (s *issueService) toCommentResponse(ctx context.Context, comment *model.IssueComment) *dto.CommentResponse {
	resp := &dto.CommentResponse{
		ID:        comment.ID,
		IssueID:   comment.IssueID,
		UserID:    comment.UserID,
		Content:   comment.Content,
		CreatedAt: comment.CreatedAt,
		UpdatedAt: comment.UpdatedAt,
	}

	if user, err := s.userRepo.GetByID(ctx, comment.UserID); err == nil {
		resp.User = &dto.UserBrief{
			ID:          user.ID,
			Username:    user.Username,
			DisplayName: user.DisplayName,
			AvatarURL:   user.AvatarURL,
		}
	}

	return resp
}
