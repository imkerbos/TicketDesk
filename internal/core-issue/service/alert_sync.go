// Package service 提供告警同步接口
package service

import "context"

// AlertSyncService 告警同步服务接口（用于避免循环依赖）
type AlertSyncService interface {
	// SyncIssueStatus 同步工单状态到告警
	SyncIssueStatus(ctx context.Context, issueID uint64, issueStatus string) error
	// SyncMergedIssueStatus 同步状态到被合并的旧工单（当合并目标工单状态变更时级联更新）
	SyncMergedIssueStatus(ctx context.Context, issueID uint64, issueStatus string) error
}
