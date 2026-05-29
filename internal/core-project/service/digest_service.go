// Package service 提供项目级每日工单日报服务
package service

import (
	"context"
	"errors"
	"fmt"
	"sort"

	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/kerbos/ticketdesk/internal/core-project/repository"
	"github.com/kerbos/ticketdesk/internal/model"
	"github.com/kerbos/ticketdesk/pkg/logger"
)

// 日报作用域
const (
	DigestScopeAllOpen      = "all_open"      // 全部未完结
	DigestScopeAssignedOnly = "assigned_only" // 仅含已指派
)

// 终态 status 黑名单（无 workflow 时 fallback 判定）
var terminalIssueStatuses = []string{"done", "closed", "resolved", "cancelled"} //nolint:misspell // 状态值与前端约定保持一致

// DigestService 每日日报服务
type DigestService interface {
	// RunForProject 立即对指定项目跑一次日报；空 issues 不发送
	RunForProject(ctx context.Context, projectID uint64) error
}

// digestService 实现
type digestService struct {
	db          *gorm.DB
	projectRepo repository.ProjectRepository
	notifier    NotificationChannelService
}

// NewDigestService 创建日报服务实例
func NewDigestService(db *gorm.DB, projectRepo repository.ProjectRepository, notifier NotificationChannelService) DigestService {
	return &digestService{
		db:          db,
		projectRepo: projectRepo,
		notifier:    notifier,
	}
}

// digestIssueRow 内部查询结果
type digestIssueRow struct {
	IssueKey         string
	Title            string
	Priority         string
	Status           string
	IssueTypeName    string
	NodeName         string
	AssigneeID       *uint64
	AssigneeName     string
	AssigneeEmail    string
	AssigneeLarkID   string
	AssigneeTelegram string
}

// RunForProject 立即执行项目日报
func (s *digestService) RunForProject(ctx context.Context, projectID uint64) error {
	project, err := s.projectRepo.GetByID(ctx, projectID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return fmt.Errorf("项目不存在")
		}
		return fmt.Errorf("查询项目失败: %w", err)
	}

	rows, err := s.queryOpenIssues(ctx, project)
	if err != nil {
		return err
	}

	if len(rows) == 0 {
		logger.Info("daily digest: no open issues, skip",
			zap.Uint64("project_id", projectID),
			zap.String("project_key", project.ProjectKey),
		)
		return nil
	}

	data := buildDigestData(project, rows)
	if err := s.notifier.NotifyProject(ctx, projectID, "issue.daily_digest", data); err != nil {
		return fmt.Errorf("发送日报通知失败: %w", err)
	}

	logger.Info("daily digest sent",
		zap.Uint64("project_id", projectID),
		zap.String("project_key", project.ProjectKey),
		zap.Int("issue_count", len(rows)),
	)
	return nil
}

// queryOpenIssues 查询项目未完结工单（含 workflow 节点判断 + status fallback）
func (s *digestService) queryOpenIssues(ctx context.Context, project *model.Project) ([]digestIssueRow, error) {
	q := s.db.WithContext(ctx).Table("issues i").
		Select(`i.issue_key, i.title, i.priority, i.status,
			it.display_name AS issue_type_name,
			COALESCE(wn.name, '') AS node_name,
			i.assignee_id,
			COALESCE(u.display_name, u.username, '') AS assignee_name,
			COALESCE(u.email, '') AS assignee_email,
			COALESCE(u.lark_open_id, '') AS assignee_lark_id,
			COALESCE(u.telegram_user_id, '') AS assignee_telegram`).
		Joins("LEFT JOIN issue_types it ON i.issue_type_id = it.id").
		Joins("LEFT JOIN workflow_instances wi ON i.workflow_instance_id = wi.id").
		Joins("LEFT JOIN workflow_nodes wn ON wi.current_node_id = wn.id").
		Joins("LEFT JOIN users u ON i.assignee_id = u.id").
		Where("i.project_id = ?", project.ID).
		Where(`(
			(i.workflow_instance_id IS NOT NULL AND wn.node_type IS NOT NULL AND wn.node_type != 'end')
			OR
			(i.workflow_instance_id IS NULL AND i.status NOT IN ?)
		)`, terminalIssueStatuses).
		Order("i.assignee_id ASC, i.priority ASC, i.id ASC")

	if project.DailyDigestScope == DigestScopeAssignedOnly {
		q = q.Where("i.assignee_id IS NOT NULL")
	}

	var rows []digestIssueRow
	if err := q.Scan(&rows).Error; err != nil {
		return nil, fmt.Errorf("查询未完结工单失败: %w", err)
	}
	return rows, nil
}

// buildDigestData 构造发送给 sender 的 data map
// sender 端约定字段：
//   - project_name, project_key, total
//   - groups: []map{"label","assignee_name","mention":{display_name,lark_open_id,email,telegram_user_id},"items":[]map{"issue_key","title","issue_type","priority","status"}}
//   - mentions: 汇总所有有 ID 的指派人（用于消息尾部统一 @）
func buildDigestData(project *model.Project, rows []digestIssueRow) map[string]any {
	// 按 assignee 分组
	type group struct {
		AssigneeID   *uint64
		AssigneeName string
		Mention      map[string]any
		Items        []map[string]any
	}
	groupMap := make(map[uint64]*group)
	var unassigned *group
	for _, r := range rows {
		item := map[string]any{
			"issue_key":  r.IssueKey,
			"title":      r.Title,
			"issue_type": r.IssueTypeName,
			"priority":   r.Priority,
			"status":     r.Status,
			"node_name":  r.NodeName,
		}
		if r.AssigneeID == nil || *r.AssigneeID == 0 {
			if unassigned == nil {
				unassigned = &group{AssigneeName: "未指派"}
			}
			unassigned.Items = append(unassigned.Items, item)
			continue
		}
		g, ok := groupMap[*r.AssigneeID]
		if !ok {
			name := r.AssigneeName
			if name == "" {
				name = fmt.Sprintf("用户#%d", *r.AssigneeID)
			}
			g = &group{
				AssigneeID:   r.AssigneeID,
				AssigneeName: name,
			}
			if r.AssigneeLarkID != "" || r.AssigneeEmail != "" || r.AssigneeTelegram != "" {
				g.Mention = map[string]any{
					"display_name":     name,
					"lark_open_id":     r.AssigneeLarkID,
					"email":            r.AssigneeEmail,
					"telegram_user_id": r.AssigneeTelegram,
				}
			}
			groupMap[*r.AssigneeID] = g
		}
		g.Items = append(g.Items, item)
	}

	// 已指派分组按 name 排序，未指派排末尾
	groups := make([]*group, 0, len(groupMap)+1)
	for _, g := range groupMap {
		groups = append(groups, g)
	}
	sort.Slice(groups, func(i, j int) bool {
		return groups[i].AssigneeName < groups[j].AssigneeName
	})
	if unassigned != nil {
		groups = append(groups, unassigned)
	}

	// 序列化为 []map[string]any
	outGroups := make([]map[string]any, 0, len(groups))
	mentions := make([]map[string]any, 0)
	for _, g := range groups {
		gm := map[string]any{
			"assignee_name": g.AssigneeName,
			"items":         g.Items,
		}
		if g.Mention != nil {
			gm["mention"] = g.Mention
			mentions = append(mentions, g.Mention)
		}
		outGroups = append(outGroups, gm)
	}

	return map[string]any{
		"project_key":  project.ProjectKey,
		"project_name": project.Name,
		"total":        len(rows),
		"groups":       outGroups,
		"mentions":     mentions,
	}
}
