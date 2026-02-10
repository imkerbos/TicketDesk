# 项目级通知渠道配置方案

## 设计思路

站内通知（WebSocket）= 全局默认开启，无需配置
外部通知（飞书/Telegram）= 项目级别配置，不同项目可推送到不同群

系统设置中的飞书/Telegram 配置作为"全局默认"，项目可以覆盖或独立配置。

## 数据模型

新建 `project_notification_channels` 表，每个项目可配置多个通知渠道：

```go
type ProjectNotificationChannel struct {
    BaseModel
    ProjectID   uint64 `gorm:"index;not null"`
    ChannelType string `gorm:"size:20;not null"`  // "lark", "telegram"
    Name        string `gorm:"size:100"`           // 渠道名称，如"运维飞书群"
    Config      string `gorm:"type:json;not null"` // JSON 配置
    Enabled     bool   `gorm:"default:true"`
    CreatedBy   uint64
}
```

Config JSON 结构：
- Lark: `{"webhook_url": "...", "secret": "..."}`
- Telegram: `{"bot_token": "...", "chat_id": "..."}`

## 通知触发流程

```
工单事件发生
  ├── 站内通知（全局，始终触发）→ WebSocket 推送给相关用户
  └── 外部通知（项目级）
       ├── 查询该项目的 notification_channels
       ├── 遍历已启用的渠道
       ├── Lark 渠道 → 用项目配置的 webhook_url 发送卡片
       └── Telegram 渠道 → 用项目配置的 bot_token + chat_id 发送
```

## 实现步骤

### Step 1: 数据模型 + Repository
- 在 `internal/model/model.go` 新增 `ProjectNotificationChannel` 模型
- 新建 `internal/core-project/repository/notification_channel_repository.go`
- CRUD + 按项目查询已启用渠道

### Step 2: DTO
- 在 `internal/core-project/dto/` 新增通知渠道 DTO
- CreateChannelRequest, UpdateChannelRequest, ChannelResponse

### Step 3: Service 层
- 在 `internal/core-project/service/` 新增通知渠道管理方法
- 或新建 `notification_channel_service.go`

### Step 4: Handler + 路由
- 新增 API:
  - GET    /api/v1/projects/:key/notification-channels
  - POST   /api/v1/projects/:key/notification-channels
  - PUT    /api/v1/projects/:key/notification-channels/:id
  - DELETE /api/v1/projects/:key/notification-channels/:id
  - POST   /api/v1/projects/:key/notification-channels/:id/test

### Step 5: 工单事件触发外部通知
- 修改 issue_service 的通知逻辑
- 工单事件发生时，查询项目的通知渠道配置
- 调用 lark/telegram service 发送（使用项目级配置而非全局配置）

### Step 6: 前端
- ProjectSettings.vue 新增「通知渠道」Tab
- 支持添加/编辑/删除/测试飞书和 Telegram 渠道

## 系统级 vs 项目级的关系

系统设置中的飞书/Telegram 配置保留，作用：
1. 告警通知（不属于任何项目）仍使用系统级配置
2. 未来可作为"默认模板"供项目快速引用

项目级配置独立存储，互不影响。

## 文件变更清单

| 操作 | 文件 |
|------|------|
| ✏️ | `internal/model/model.go` — 新增 ProjectNotificationChannel |
| 🆕 | `internal/core-project/repository/notification_channel_repository.go` |
| 🆕 | `internal/core-project/dto/notification_channel_dto.go` |
| 🆕 | `internal/core-project/service/notification_channel_service.go` |
| 🆕 | `internal/core-project/handler/notification_channel_handler.go` |
| ✏️ | `internal/api/router/router.go` — 注册新路由 |
| ✏️ | `internal/core-issue/service/issue_service.go` — 触发外部通知 |
| ✏️ | `web/src/views/project/ProjectSettings.vue` — 新增通知渠道 Tab |
| ✏️ | `web/src/api/project.ts` — 新增 API |
| ✏️ | `web/src/types/project.ts` — 新增类型 |
