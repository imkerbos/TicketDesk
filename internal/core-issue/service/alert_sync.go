// Package service 提供告警同步接口
package service

import "context"

// AlertSyncService 告警同步服务接口（用于避免循环依赖）
type AlertSyncService interface {
	// SyncIssueStatus 同步工单状态到告警
	SyncIssueStatus(ctx context.Context, issueID uint64, issueStatus string) error
}
