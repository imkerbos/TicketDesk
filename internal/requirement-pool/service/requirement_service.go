package service

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/kerbos/ticketdesk/internal/model"
	"github.com/kerbos/ticketdesk/internal/requirement-pool/dto"
	"github.com/kerbos/ticketdesk/internal/requirement-pool/repository"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// RequirementService 需求业务逻辑接口
type RequirementService interface {
	Create(ctx context.Context, req *dto.CreateRequirementRequest, userID uint64) (*dto.RequirementResponse, error)
	GetByID(ctx context.Context, id uint64) (*dto.RequirementResponse, error)
	Update(ctx context.Context, id uint64, req *dto.UpdateRequirementRequest, userID uint64) error
	Delete(ctx context.Context, id uint64, userID uint64) error
	List(ctx context.Context, req *dto.RequirementListRequest) ([]*dto.RequirementResponse, int64, error)
	ConvertToIssue(ctx context.Context, id uint64, req *dto.ConvertToIssueRequest, userID uint64) (*model.Issue, error)
	AddComment(ctx context.Context, id uint64, req *dto.RequirementCommentRequest, userID uint64) (*dto.RequirementCommentResponse, error)
	GetKanban(ctx context.Context, req *dto.KanbanRequest) (*dto.KanbanResponse, error)
	GetReport(ctx context.Context, req *dto.ReportRequest) (*dto.ReportResponse, error)
}

// requirementService 需求业务逻辑实现
type requirementService struct {
	repo       repository.RequirementRepository
	poolRepo   repository.RequirementPoolRepository
	db         *gorm.DB
	logger     *zap.Logger
}

// NewRequirementService 创建需求业务逻辑实例
func NewRequirementService(
	repo repository.RequirementRepository,
	poolRepo repository.RequirementPoolRepository,
	db *gorm.DB,
	logger *zap.Logger,
) RequirementService {
	return &requirementService{
		repo:     repo,
		poolRepo: poolRepo,
		db:       db,
		logger:   logger,
	}
}

// Create 创建需求
func (s *requirementService) Create(ctx context.Context, req *dto.CreateRequirementRequest, userID uint64) (*dto.RequirementResponse, error) {
	// 验证需求池是否存在
	pool, err := s.poolRepo.GetByID(ctx, req.PoolID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("需求池不存在")
		}
		return nil, fmt.Errorf("获取需求池失败: %w", err)
	}

	// 验证需求池状态
	if pool.Status != model.RequirementPoolStatusActive {
		return nil, errors.New("需求池已归档，无法添加需求")
	}

	// 处理标签
	tags := ""
	if len(req.Tags) > 0 {
		tags = strings.Join(req.Tags, ",")
	}

	requirement := &model.Requirement{
		PoolID:          req.PoolID,
		Title:           req.Title,
		Description:     req.Description,
		Priority:        req.Priority,
		Status:          model.RequirementStatusPendingReview,
		Category:        req.Category,
		ReporterID:      req.ReporterID,
		AssigneeID:      req.AssigneeID,
		StartDate:       req.StartDate,
		EndDate:         req.EndDate,
		TargetProjectID: req.TargetProjectID,
		CreatedBy:       userID,
		Tags:            tags,
	}

	if err := s.repo.Create(ctx, requirement); err != nil {
		s.logger.Error("failed to create requirement",
			zap.Error(err),
			zap.String("title", req.Title),
			zap.Uint64("user_id", userID),
		)
		return nil, fmt.Errorf("创建需求失败: %w", err)
	}

	s.logger.Info("requirement created",
		zap.Uint64("requirement_id", requirement.ID),
		zap.String("title", requirement.Title),
		zap.Uint64("pool_id", requirement.PoolID),
		zap.Uint64("user_id", userID),
	)

	return s.GetByID(ctx, requirement.ID)
}

// GetByID 根据ID获取需求
func (s *requirementService) GetByID(ctx context.Context, id uint64) (*dto.RequirementResponse, error) {
	requirement, err := s.repo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("需求不存在")
		}
		s.logger.Error("failed to get requirement",
			zap.Error(err),
			zap.Uint64("requirement_id", id),
		)
		return nil, fmt.Errorf("获取需求失败: %w", err)
	}

	return s.toRequirementResponse(requirement), nil
}

// Update 更新需求
func (s *requirementService) Update(ctx context.Context, id uint64, req *dto.UpdateRequirementRequest, userID uint64) error {
	requirement, err := s.repo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("需求不存在")
		}
		return fmt.Errorf("获取需求失败: %w", err)
	}

	// 更新字段
	if req.Title != nil {
		requirement.Title = *req.Title
	}
	if req.Description != nil {
		requirement.Description = *req.Description
	}
	if req.Priority != nil {
		requirement.Priority = *req.Priority
	}
	if req.Status != nil {
		// 验证状态转换
		if err := s.validateStatusTransition(requirement.Status, *req.Status); err != nil {
			return err
		}
		requirement.Status = *req.Status
	}
	if req.Category != nil {
		requirement.Category = *req.Category
	}
	if req.ReporterID != nil {
		requirement.ReporterID = req.ReporterID
	}
	if req.AssigneeID != nil {
		requirement.AssigneeID = req.AssigneeID
	}
	if req.StartDate != nil {
		requirement.StartDate = req.StartDate
	}
	if req.EndDate != nil {
		requirement.EndDate = req.EndDate
	}
	if req.Progress != nil {
		requirement.Progress = *req.Progress
	}
	if req.Result != nil {
		requirement.Result = *req.Result
	}
	if req.TargetProjectID != nil {
		requirement.TargetProjectID = req.TargetProjectID
	}
	if req.Tags != nil {
		requirement.Tags = strings.Join(req.Tags, ",")
	}

	if err := s.repo.Update(ctx, requirement); err != nil {
		s.logger.Error("failed to update requirement",
			zap.Error(err),
			zap.Uint64("requirement_id", id),
			zap.Uint64("user_id", userID),
		)
		return fmt.Errorf("更新需求失败: %w", err)
	}

	s.logger.Info("requirement updated",
		zap.Uint64("requirement_id", id),
		zap.Uint64("user_id", userID),
	)

	return nil
}

// Delete 删除需求
func (s *requirementService) Delete(ctx context.Context, id uint64, userID uint64) error {
	_, err := s.repo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("需求不存在")
		}
		return fmt.Errorf("获取需求失败: %w", err)
	}

	if err := s.repo.Delete(ctx, id); err != nil {
		s.logger.Error("failed to delete requirement",
			zap.Error(err),
			zap.Uint64("requirement_id", id),
			zap.Uint64("user_id", userID),
		)
		return fmt.Errorf("删除需求失败: %w", err)
	}

	s.logger.Info("requirement deleted",
		zap.Uint64("requirement_id", id),
		zap.Uint64("user_id", userID),
	)

	return nil
}

// List 获取需求列表
func (s *requirementService) List(ctx context.Context, req *dto.RequirementListRequest) ([]*dto.RequirementResponse, int64, error) {
	filter := &repository.RequirementFilter{
		PoolID:     req.PoolID,
		Status:     req.Status,
		Priority:   req.Priority,
		Category:   req.Category,
		AssigneeID: req.AssigneeID,
		CreatedBy:  req.CreatedBy,
		StartDate:  req.StartDate,
		EndDate:    req.EndDate,
		Keyword:    req.Keyword,
		Page:       req.Page,
		PageSize:   req.PageSize,
	}

	// 设置默认分页
	if filter.Page == 0 {
		filter.Page = 1
	}
	if filter.PageSize == 0 {
		filter.PageSize = 20
	}

	requirements, total, err := s.repo.List(ctx, filter)
	if err != nil {
		s.logger.Error("failed to list requirements",
			zap.Error(err),
		)
		return nil, 0, fmt.Errorf("获取需求列表失败: %w", err)
	}

	responses := make([]*dto.RequirementResponse, 0, len(requirements))
	for _, requirement := range requirements {
		responses = append(responses, s.toRequirementResponse(requirement))
	}

	return responses, total, nil
}

// ConvertToIssue 转化为工单
func (s *requirementService) ConvertToIssue(ctx context.Context, id uint64, req *dto.ConvertToIssueRequest, userID uint64) (*model.Issue, error) {
	requirement, err := s.repo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("需求不存在")
		}
		return nil, fmt.Errorf("获取需求失败: %w", err)
	}

	// 只有终态不允许转化
	if requirement.Status == model.RequirementStatusCompleted {
		return nil, errors.New("已完成的需求不能转化为工单")
	}
	if requirement.Status == model.RequirementStatusRejected {
		return nil, errors.New("已拒绝的需求不能转化为工单")
	}

	// 使用事务
	var issue *model.Issue
	err = s.db.Transaction(func(tx *gorm.DB) error {
		// 获取项目信息
		var project model.Project
		if err := tx.First(&project, req.ProjectID).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return errors.New("项目不存在")
			}
			return fmt.Errorf("查询项目失败: %w", err)
		}

		// 生成工单编号
		var maxNum int64
		if err := tx.Model(&model.Issue{}).
			Where("project_id = ?", req.ProjectID).
			Count(&maxNum).Error; err != nil {
			return fmt.Errorf("生成工单编号失败: %w", err)
		}
		nextNum := maxNum + 1
		issueKey := fmt.Sprintf("%s-%d", project.ProjectKey, nextNum)

		// 创建工单
		issue = &model.Issue{
			IssueKey:    issueKey,
			ProjectID:   req.ProjectID,
			IssueTypeID: req.IssueTypeID,
			Title:       requirement.Title,
			Description: requirement.Description,
			Priority:    string(requirement.Priority),
			Status:      "open",
			AssigneeID:  req.AssigneeID,
			ReporterID:  userID,
		}

		if err := tx.Create(issue).Error; err != nil {
			return fmt.Errorf("创建工单失败: %w", err)
		}

		// 记录转化关联，但不改变需求状态
		requirement.ConvertedIssueID = &issue.ID
		if err := tx.Save(requirement).Error; err != nil {
			return fmt.Errorf("更新需求关联失败: %w", err)
		}

		// 获取用户信息用于活动记录
		var user model.User
		if err := tx.First(&user, userID).Error; err == nil {
			// 创建需求活动记录
			requirementActivity := &model.ActivityLog{
				UserID:     userID,
				UserName:   user.Username,
				Action:     "converted",
				EntityType: "requirement",
				EntityID:   requirement.ID,
				EntityKey:  requirement.Title,
				Details:    fmt.Sprintf(`{"issue_id":%d,"issue_key":"%s","message":"需求已转化为工单 %s"}`, issue.ID, issue.IssueKey, issue.IssueKey),
			}
			tx.Create(requirementActivity)

			// 创建工单活动记录
			issueActivity := &model.ActivityLog{
				UserID:     userID,
				UserName:   user.Username,
				Action:     "created",
				EntityType: "issue",
				EntityID:   issue.ID,
				EntityKey:  issue.IssueKey,
				Details:    fmt.Sprintf(`{"requirement_id":%d,"message":"工单由需求「%s」转化创建"}`, requirement.ID, requirement.Title),
			}
			tx.Create(issueActivity)
		}

		return nil
	})

	if err != nil {
		s.logger.Error("failed to convert requirement to issue",
			zap.Error(err),
			zap.Uint64("requirement_id", id),
			zap.Uint64("user_id", userID),
		)
		return nil, err
	}

	s.logger.Info("requirement converted to issue",
		zap.Uint64("requirement_id", id),
		zap.Uint64("issue_id", issue.ID),
		zap.String("issue_key", issue.IssueKey),
		zap.Uint64("user_id", userID),
	)

	return issue, nil
}

// AddComment 添加评论
func (s *requirementService) AddComment(ctx context.Context, id uint64, req *dto.RequirementCommentRequest, userID uint64) (*dto.RequirementCommentResponse, error) {
	// 验证需求是否存在
	_, err := s.repo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("需求不存在")
		}
		return nil, fmt.Errorf("获取需求失败: %w", err)
	}

	comment := &model.RequirementComment{
		RequirementID: id,
		UserID:        userID,
		Content:       req.Content,
	}

	if err := s.db.WithContext(ctx).Create(comment).Error; err != nil {
		s.logger.Error("failed to create requirement comment",
			zap.Error(err),
			zap.Uint64("requirement_id", id),
			zap.Uint64("user_id", userID),
		)
		return nil, fmt.Errorf("添加评论失败: %w", err)
	}

	s.logger.Info("requirement comment created",
		zap.Uint64("comment_id", comment.ID),
		zap.Uint64("requirement_id", id),
		zap.Uint64("user_id", userID),
	)

	// 预加载用户信息
	if err := s.db.WithContext(ctx).Preload("User").First(comment, comment.ID).Error; err != nil {
		return nil, fmt.Errorf("获取评论失败: %w", err)
	}

	return s.toCommentResponse(comment), nil
}

// GetKanban 获取看板数据
func (s *requirementService) GetKanban(ctx context.Context, req *dto.KanbanRequest) (*dto.KanbanResponse, error) {
	filter := &repository.KanbanFilter{
		PoolID:    req.PoolID,
		GroupBy:   req.GroupBy,
		StartDate: req.StartDate,
		EndDate:   req.EndDate,
	}

	data, err := s.repo.GetKanbanData(ctx, filter)
	if err != nil {
		s.logger.Error("failed to get kanban data",
			zap.Error(err),
		)
		return nil, fmt.Errorf("获取看板数据失败: %w", err)
	}

	// 构建响应
	response := &dto.KanbanResponse{
		GroupBy: req.GroupBy,
		Columns: make([]*dto.KanbanColumn, 0),
		Total:   0,
	}

	// 根据分组类型生成列
	switch req.GroupBy {
	case "status":
		statuses := []model.RequirementStatus{
			model.RequirementStatusPendingReview,
			model.RequirementStatusPlanning,
			model.RequirementStatusInProgress,
			model.RequirementStatusCompleted,
			model.RequirementStatusOnHold,
			model.RequirementStatusRejected,
		}
		statusTitles := map[model.RequirementStatus]string{
			model.RequirementStatusPendingReview: "待评估",
			model.RequirementStatusPlanning:      "规划中",
			model.RequirementStatusInProgress:    "进行中",
			model.RequirementStatusCompleted:     "已完成",
			model.RequirementStatusOnHold:        "已搁置",
			model.RequirementStatusRejected:      "已拒绝",
		}
		for _, status := range statuses {
			key := string(status)
			requirements := data[key]
			column := &dto.KanbanColumn{
				Key:          key,
				Title:        statusTitles[status],
				Count:        int64(len(requirements)),
				Requirements: make([]*dto.RequirementResponse, 0),
			}
			for _, req := range requirements {
				column.Requirements = append(column.Requirements, s.toRequirementResponse(req))
			}
			response.Columns = append(response.Columns, column)
			response.Total += column.Count
		}

	case "priority":
		priorities := []model.RequirementPriority{
			model.RequirementPriorityP0,
			model.RequirementPriorityP1,
			model.RequirementPriorityP2,
			model.RequirementPriorityP3,
		}
		priorityTitles := map[model.RequirementPriority]string{
			model.RequirementPriorityP0: "P0 - 紧急",
			model.RequirementPriorityP1: "P1 - 高",
			model.RequirementPriorityP2: "P2 - 中",
			model.RequirementPriorityP3: "P3 - 低",
		}
		for _, priority := range priorities {
			key := string(priority)
			requirements := data[key]
			column := &dto.KanbanColumn{
				Key:          key,
				Title:        priorityTitles[priority],
				Count:        int64(len(requirements)),
				Requirements: make([]*dto.RequirementResponse, 0),
			}
			for _, req := range requirements {
				column.Requirements = append(column.Requirements, s.toRequirementResponse(req))
			}
			response.Columns = append(response.Columns, column)
			response.Total += column.Count
		}

	case "assignee", "timeline":
		// 动态生成列
		for key, requirements := range data {
			column := &dto.KanbanColumn{
				Key:          key,
				Title:        key,
				Count:        int64(len(requirements)),
				Requirements: make([]*dto.RequirementResponse, 0),
			}
			for _, req := range requirements {
				column.Requirements = append(column.Requirements, s.toRequirementResponse(req))
			}
			response.Columns = append(response.Columns, column)
			response.Total += column.Count
		}
	}

	return response, nil
}

// GetReport 获取报告数据
func (s *requirementService) GetReport(ctx context.Context, req *dto.ReportRequest) (*dto.ReportResponse, error) {
	filter := &repository.StatisticsFilter{
		PoolID:    req.PoolID,
		StartDate: req.StartDate,
		EndDate:   req.EndDate,
	}

	stats, err := s.repo.GetStatistics(ctx, filter)
	if err != nil {
		s.logger.Error("failed to get statistics",
			zap.Error(err),
		)
		return nil, fmt.Errorf("获取统计数据失败: %w", err)
	}

	response := &dto.ReportResponse{
		PoolID:         req.PoolID,
		StartDate:      req.StartDate,
		EndDate:        req.EndDate,
		TotalCreated:   stats.TotalCreated,
		TotalCompleted: stats.TotalCompleted,
		TotalRejected:  stats.TotalRejected,
		AvgProcessDays: stats.AvgProcessDays,
	}

	// 获取需求池名称
	if req.PoolID != nil {
		pool, err := s.poolRepo.GetByID(ctx, *req.PoolID)
		if err == nil {
			response.PoolName = pool.Name
		}
	}

	// 按状态统计
	statusCounts, err := s.repo.CountByStatus(ctx, 0)
	if req.PoolID != nil {
		statusCounts, err = s.repo.CountByStatus(ctx, *req.PoolID)
	}
	if err == nil {
		response.StatusSummary = make([]*dto.StatusSummary, 0)
		for status, count := range statusCounts {
			response.StatusSummary = append(response.StatusSummary, &dto.StatusSummary{
				Status: status,
				Count:  count,
			})
		}
	}

	// 按优先级统计
	priorityCounts, err := s.repo.CountByPriority(ctx, 0)
	if req.PoolID != nil {
		priorityCounts, err = s.repo.CountByPriority(ctx, *req.PoolID)
	}
	if err == nil {
		response.PrioritySummary = make([]*dto.PrioritySummary, 0)
		for priority, count := range priorityCounts {
			response.PrioritySummary = append(response.PrioritySummary, &dto.PrioritySummary{
				Priority: priority,
				Count:    count,
			})
		}
	}

	// TODO: 实现负责人统计和趋势数据
	response.AssigneeSummary = make([]*dto.AssigneeSummary, 0)
	response.TrendData = make([]*dto.TrendData, 0)

	return response, nil
}

// validateStatusTransition 验证状态转换
func (s *requirementService) validateStatusTransition(from, to model.RequirementStatus) error {
	// 定义允许的状态转换
	allowedTransitions := map[model.RequirementStatus][]model.RequirementStatus{
		model.RequirementStatusPendingReview: {
			model.RequirementStatusPlanning,
			model.RequirementStatusInProgress,
			model.RequirementStatusRejected,
			model.RequirementStatusOnHold,
		},
		model.RequirementStatusPlanning: {
			model.RequirementStatusInProgress,
			model.RequirementStatusOnHold,
			model.RequirementStatusRejected,
			model.RequirementStatusPendingReview,
		},
		model.RequirementStatusInProgress: {
			model.RequirementStatusCompleted,
			model.RequirementStatusOnHold,
			model.RequirementStatusRejected,
		},
		model.RequirementStatusOnHold: {
			model.RequirementStatusPlanning,
			model.RequirementStatusInProgress,
			model.RequirementStatusPendingReview,
		},
		model.RequirementStatusCompleted: {},
		model.RequirementStatusRejected: {
			model.RequirementStatusPendingReview,
		},
	}

	allowed, exists := allowedTransitions[from]
	if !exists {
		return fmt.Errorf("无效的状态: %s", from)
	}

	for _, status := range allowed {
		if status == to {
			return nil
		}
	}

	return fmt.Errorf("不允许从 %s 转换到 %s", from, to)
}

// toRequirementResponse 转换为需求响应DTO
func (s *requirementService) toRequirementResponse(req *model.Requirement) *dto.RequirementResponse {
	resp := &dto.RequirementResponse{
		ID:               req.ID,
		PoolID:           req.PoolID,
		Title:            req.Title,
		Description:      req.Description,
		Priority:         req.Priority,
		Status:           req.Status,
		Category:         req.Category,
		ReporterID:       req.ReporterID,
		AssigneeID:       req.AssigneeID,
		StartDate:        req.StartDate,
		EndDate:          req.EndDate,
		Progress:         req.Progress,
		Result:           req.Result,
		ConvertedIssueID: req.ConvertedIssueID,
		TargetProjectID:  req.TargetProjectID,
		CreatedBy:        req.CreatedBy,
		CreatedAt:        req.CreatedAt,
		UpdatedAt:        req.UpdatedAt,
	}

	// 处理标签
	if req.Tags != "" {
		resp.Tags = strings.Split(req.Tags, ",")
	} else {
		resp.Tags = []string{}
	}

	// 关联数据
	if req.Pool != nil {
		resp.PoolName = req.Pool.Name
	}
	if req.Reporter != nil {
		resp.ReporterName = req.Reporter.Username
	}
	if req.Assignee != nil {
		resp.AssigneeName = req.Assignee.Username
	}
	if req.Creator != nil {
		resp.CreatorName = req.Creator.Username
	}
	if req.ConvertedIssue != nil {
		resp.ConvertedIssueKey = req.ConvertedIssue.IssueKey
		resp.ConvertedIssueStatus = req.ConvertedIssue.Status
	}
	if req.TargetProject != nil {
		resp.TargetProjectName = req.TargetProject.Name
	}

	// 评论数量
	if req.Comments != nil {
		resp.CommentCount = int64(len(req.Comments))
	}

	return resp
}

// toCommentResponse 转换为评论响应DTO
func (s *requirementService) toCommentResponse(comment *model.RequirementComment) *dto.RequirementCommentResponse {
	resp := &dto.RequirementCommentResponse{
		ID:            comment.ID,
		RequirementID: comment.RequirementID,
		UserID:        comment.UserID,
		Content:       comment.Content,
		CreatedAt:     comment.CreatedAt,
	}

	if comment.User != nil {
		resp.UserName = comment.User.Username
	}

	return resp
}
