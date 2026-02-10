# 重构方案：合并状态流转与工作流为一套系统

## 问题

工单有两套独立的流转系统：
1. **状态流转**：`handleTransition` → `POST /issues/:key/transition` → 直接改 `issue.status`
2. **工作流流转**：`approve/reject/complete` → workflow engine → `syncIssueStatus` 间接改 status

两套互不关联。改了状态工作流不动，走了工作流状态不同步到 UI 的状态流转按钮。

## 目标

**一套系统**：
- **有工作流的工单**：所有状态变更都通过工作流节点驱动，状态流转按钮消失，只显示工作流操作按钮
- **没有工作流的工单**：保留现有状态流转下拉菜单（行为不变）
- 工作流操作后，issue.status 实时同步，前端刷新后两边一致

## 现状分析

| 场景 | 当前行为 | 期望行为 |
|------|---------|---------|
| 有工作流 + 审批通过 | workflow 推进节点，syncIssueStatus 改 status，但无活动日志/通知 | 同左 + 活动日志 + 通知（已实现） |
| 有工作流 + 用户手动状态流转 | 直接改 status，workflow 不动 | **禁止手动流转**，只能通过工作流操作 |
| 无工作流 + 用户手动状态流转 | 直接改 status | 保持不变 |

## 实际需要改的东西

### 核心发现

前端 **已经做了互斥**：
```vue
<el-dropdown v-if="issue && !hasActiveWorkflow && getAvailableTransitions(issue.status).length > 0">
```
当 `hasActiveWorkflow = true` 时，状态流转下拉菜单已经隐藏了。

后端 workflow engine 的 `syncIssueStatus` **已经在改 issue.status**。

**真正的问题是**：`syncIssueStatus` 是一个原始 DB update，不走 issue_service 的 `TransitionIssue`，所以：
1. ~~不记录活动日志~~ → handler 层已补上
2. 不发通知
3. 不同步告警状态（alertSyncSvc）
4. 时间字段的处理逻辑可能不一致

### 需要修复的点

1. **后端 `syncIssueStatus`**：增强为调用完整的状态更新逻辑（通知 + 告警同步）
2. **后端 `TransitionIssue`**：增加校验，如果工单有活跃工作流，拒绝手动流转
3. **前端**：工作流操作成功后，确保 issue 状态实时刷新（已实现 `loadIssue()`）
4. **前端流程图**：已实现，保持不变

## 修改清单

### 文件 1: `internal/core-issue/service/issue_service.go`

**TransitionIssue 方法**（~L880）：在开头增加校验
```go
// 如果工单有活跃的工作流实例，禁止手动状态流转
if issue.WorkflowInstanceID != nil && *issue.WorkflowInstanceID > 0 {
    // 检查工作流实例是否活跃
    if h.workflowEngine != nil {
        instance, err := h.workflowEngine.GetInstanceStatus(ctx, *issue.WorkflowInstanceID)
        if err == nil && instance == "active" {
            return nil, fmt.Errorf("该工单已绑定活跃工作流，请通过工作流操作变更状态")
        }
    }
}
```

### 文件 2: `internal/core-workflow/service/workflow_engine.go`

**syncIssueStatus 方法**（~L706）：增强，增加告警同步
- 在更新 issue status 后，如果 issue 有 alert_id，同步告警状态
- 这需要注入 alertSyncSvc 或者通过回调实现

**更简单的方案**：在 syncIssueStatus 中，改为调用 issueRepo 的 Update 方法而不是原始 SQL，让 GORM hooks 生效。但这可能引入循环依赖。

**最终方案**：在 workflow_handler.go 的 Approve/Reject/Complete 成功后，调用 issue_service 的 `SyncStatusFromWorkflow` 新方法，该方法处理通知和告警同步。

### 文件 3: `internal/core-issue/service/issue_service.go`

新增方法 `SyncStatusFromWorkflow`：
```go
func (s *issueService) SyncStatusFromWorkflow(ctx context.Context, issueID uint64) error {
    // 重新加载 issue（因为 workflow engine 已经改了 status）
    issue, err := s.issueRepo.GetByID(ctx, issueID)
    // 同步告警状态
    // 发送通知
}
```

### 文件 4: `internal/core-workflow/handler/workflow_handler.go`

在 Approve/Reject/Complete 成功后，调用 `issueSvc.SyncStatusFromWorkflow()`

### 文件 5: 前端无需修改

前端已经：
- 有工作流时隐藏状态流转按钮 ✅
- 工作流操作后刷新 issue 数据 ✅
- 流程图已实现 ✅

## 总结

改动量很小：
1. `TransitionIssue` 加一个工作流活跃检查（防止 API 层面绕过）
2. `workflow_handler.go` 操作成功后调用 issue service 同步（通知 + 告警）
3. 前端不需要改
