# 告警与工单联动功能完善总结

## 完成的功能

### 1. 告警合并逻辑 ✅

**功能描述**：支持在时间窗口内将相似告警合并到同一个工单，避免重复建单。

**实现细节**：
- 在 `AlertRule` 模型中添加了 `MergeWindow` 字段（单位：秒），默认值为 3600（1小时）
- 在自动建单逻辑中，检查是否存在符合条件的未关闭工单：
  - 同一项目
  - 同一工单类型
  - 同一优先级
  - 在合并时间窗口内创建
  - 状态不是 resolved 或 closed
- 如果找到可合并的工单，将新告警关联到现有工单，而不是创建新工单

**相关文件**：
- `internal/model/model.go` - 添加 MergeWindow 字段
- `internal/integration-alert/dto/alert_dto.go` - 更新 DTO
- `internal/integration-alert/service/alert_service.go` - 更新自动建单逻辑
- `internal/integration-alert/service/alert_service_helpers.go` - 添加 findMergeableIssue 方法

---

### 2. 告警静默功能 ✅

**功能描述**：支持按标签匹配临时屏蔽告警，避免在维护期间产生大量无效告警。

**实现细节**：
- 创建 `AlertSilence` 模型，包含以下字段：
  - `Name`：静默规则名称
  - `Description`：描述
  - `LabelMatchers`：标签匹配规则（JSON）
  - `StartsAt`：静默开始时间
  - `EndsAt`：静默结束时间
  - `CreatedBy`：创建人
  - `Comment`：静默原因
  - `Status`：状态（0-已取消, 1-生效中, 2-已过期）
- 在 Webhook 处理时，检查告警是否匹配任何生效中的静默规则
- 如果匹配，跳过告警处理（不创建告警记录，不自动建单）
- 提供完整的 CRUD API：
  - `POST /api/v1/alert-silences` - 创建静默
  - `GET /api/v1/alert-silences` - 列表查询
  - `GET /api/v1/alert-silences/:id` - 获取详情
  - `PUT /api/v1/alert-silences/:id` - 更新静默
  - `DELETE /api/v1/alert-silences/:id` - 删除静默
  - `POST /api/v1/alert-silences/:id/cancel` - 取消静默

**相关文件**：
- `internal/model/model.go` - 添加 AlertSilence 模型
- `internal/integration-alert/dto/alert_dto.go` - 添加静默相关 DTO
- `internal/integration-alert/repository/alert_repository.go` - 添加 AlertSilenceRepository
- `internal/integration-alert/service/alert_service_helpers.go` - 添加静默检查和管理方法
- `internal/integration-alert/handler/alert_handler.go` - 添加静默管理 Handler
- `internal/api/router/router.go` - 注册静默管理路由

---

### 3. 工单到告警的状态同步 ✅

**功能描述**：当工单状态变化时，自动同步更新关联的所有告警状态，实现双向联动。

**实现细节**：
- 创建 `AlertSyncService` 接口，避免循环依赖
- 在 `IssueService` 中添加可选的 `AlertSyncService` 依赖
- 在 `TransitionIssue` 方法中，工单状态更新后同步告警状态：
  - 工单状态为 `resolved` 或 `closed` → 告警状态更新为 `resolved`
  - 工单状态为 `reopened` → 告警状态更新为 `firing`
  - 其他状态不同步
- 在 `AlertRepository` 中添加 `UpdateStatusByIssueID` 方法，批量更新关联告警
- 在 `AlertService` 中实现 `SyncIssueStatus` 方法
- 在 Router 初始化时，将 `AlertService` 设置为 `IssueService` 的 `AlertSyncService`

**相关文件**：
- `internal/core-issue/service/alert_sync.go` - 定义 AlertSyncService 接口
- `internal/core-issue/service/issue_service.go` - 添加告警同步逻辑
- `internal/integration-alert/repository/alert_repository.go` - 添加批量更新方法
- `internal/integration-alert/service/alert_service_helpers.go` - 实现 SyncIssueStatus
- `internal/api/router/router.go` - 设置依赖关系

---

### 4. 告警分组查询功能 ✅

**功能描述**：支持按标签（如 cluster、namespace、service）对告警进行分组统计，方便查看告警分布。

**实现细节**：
- 在 `AlertRepository` 中添加 `GroupBy` 方法，按指定标签键分组统计
- 返回每个分组的：
  - 分组值（标签值）
  - 告警数量
  - 按严重程度统计（critical/warning/info）
  - 按状态统计（firing/resolved）
- 提供 API：`GET /api/v1/alerts/group?group_by=cluster&status=firing`
- 支持过滤条件：
  - `status`：按状态过滤
  - `severity`：按严重程度过滤

**相关文件**：
- `internal/integration-alert/dto/alert_dto.go` - 添加分组相关 DTO
- `internal/integration-alert/repository/alert_repository.go` - 实现 GroupBy 方法
- `internal/integration-alert/service/alert_service_helpers.go` - 实现 GroupAlerts 方法
- `internal/integration-alert/handler/alert_handler.go` - 添加 HandleGroupAlerts
- `internal/api/router/router.go` - 注册分组查询路由

---

## 数据库变更

### 新增字段
- `alert_rules.merge_window` - INT，默认值 3600（告警合并时间窗口）

### 新增表
- `alert_silences` - 告警静默表
  - `id` - BIGINT UNSIGNED，主键
  - `name` - VARCHAR(100)，静默规则名称
  - `description` - TEXT，描述
  - `label_matchers` - JSON，标签匹配规则
  - `starts_at` - DATETIME，静默开始时间
  - `ends_at` - DATETIME，静默结束时间
  - `created_by` - BIGINT UNSIGNED，创建人
  - `comment` - TEXT，静默原因
  - `status` - TINYINT，状态（0-已取消, 1-生效中, 2-已过期）
  - `created_at` - DATETIME，创建时间
  - `updated_at` - DATETIME，更新时间
  - `deleted_at` - DATETIME，软删除时间

---

## API 接口

### 告警静默管理
- `GET /api/v1/alert-silences` - 获取静默列表
- `POST /api/v1/alert-silences` - 创建静默
- `GET /api/v1/alert-silences/:id` - 获取静默详情
- `PUT /api/v1/alert-silences/:id` - 更新静默
- `DELETE /api/v1/alert-silences/:id` - 删除静默（需要 Admin 权限）
- `POST /api/v1/alert-silences/:id/cancel` - 取消静默

### 告警分组查询
- `GET /api/v1/alerts/group` - 按标签分组统计告警
  - 查询参数：
    - `group_by`：分组字段（必填）
    - `status`：按状态过滤（可选）
    - `severity`：按严重程度过滤（可选）

### 告警规则更新
- `POST /api/v1/alert-rules` - 创建规则时支持 `merge_window` 字段
- `PUT /api/v1/alert-rules/:id` - 更新规则时支持 `merge_window` 字段

---

## 使用示例

### 1. 创建告警规则（支持合并）

```bash
curl -X POST http://localhost:8080/api/v1/alert-rules \
  -H "Authorization: Bearer <token>" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "高 CPU 告警规则",
    "description": "CPU 使用率超过 80% 时自动建单",
    "project_id": 1,
    "issue_type_id": 4,
    "label_matchers": [
      {"key": "alertname", "operator": "==", "value": "HighCPU"},
      {"key": "severity", "operator": "==", "value": "critical"}
    ],
    "priority": "P1",
    "assignee_id": 2,
    "auto_resolve": true,
    "merge_window": 7200
  }'
```

### 2. 创建告警静默

```bash
curl -X POST http://localhost:8080/api/v1/alert-silences \
  -H "Authorization: Bearer <token>" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "生产环境维护静默",
    "description": "生产环境维护期间静默所有告警",
    "label_matchers": [
      {"key": "env", "operator": "==", "value": "production"}
    ],
    "starts_at": "2026-02-03T20:00:00+08:00",
    "ends_at": "2026-02-03T22:00:00+08:00",
    "comment": "生产环境数据库升级维护"
  }'
```

### 3. 按集群分组查询告警

```bash
curl -X GET "http://localhost:8080/api/v1/alerts/group?group_by=cluster&status=firing" \
  -H "Authorization: Bearer <token>"
```

响应示例：
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "group_by": "cluster",
    "items": [
      {
        "group_value": "prod-cluster-1",
        "count": 15,
        "severity": {
          "critical": 3,
          "warning": 10,
          "info": 2
        },
        "status": {
          "firing": 15
        }
      },
      {
        "group_value": "prod-cluster-2",
        "count": 8,
        "severity": {
          "critical": 1,
          "warning": 7
        },
        "status": {
          "firing": 8
        }
      }
    ],
    "total": 23
  }
}
```

---

## 架构优化

### 避免循环依赖
- 创建 `AlertSyncService` 接口，定义在 `core-issue` 模块
- `IssueService` 依赖 `AlertSyncService` 接口（而不是具体实现）
- `AlertService` 实现 `AlertSyncService` 接口
- 在 Router 初始化时，通过 `SetAlertSyncService` 方法注入依赖
- 这样避免了 `core-issue` 和 `integration-alert` 之间的循环依赖

### 代码组织
- 将辅助方法提取到 `alert_service_helpers.go`，保持主文件简洁
- 使用接口隔离，提高代码可测试性
- 遵循单一职责原则，每个方法职责清晰

---

## 测试建议

### 1. 告警合并测试
- 创建告警规则，设置 `merge_window` 为 60 秒
- 在 60 秒内发送多个相同的告警
- 验证只创建了一个工单，多个告警都关联到同一个工单

### 2. 告警静默测试
- 创建静默规则，匹配特定标签
- 发送匹配的告警
- 验证告警被静默，没有创建告警记录和工单
- 取消静默后，验证告警正常处理

### 3. 状态同步测试
- 创建一个关联了告警的工单
- 将工单状态变更为 `resolved`
- 验证关联的告警状态也变更为 `resolved`
- 将工单状态变更为 `reopened`
- 验证关联的告警状态变更为 `firing`

### 4. 分组查询测试
- 创建多个告警，包含不同的标签值
- 按不同标签键分组查询
- 验证分组统计结果正确

---

## 后续优化建议

1. **告警升级策略**：长时间未处理的告警自动升级优先级或重新指派
2. **告警通知**：集成邮件、Webhook、企业微信等通知渠道
3. **告警趋势分析**：统计告警趋势，识别频繁告警的服务
4. **告警抑制**：支持更复杂的抑制规则（如父子关系抑制）
5. **告警模板**：支持自定义工单标题和描述模板
6. **告警历史**：记录告警的完整生命周期历史
7. **性能优化**：对于大量告警的场景，考虑使用消息队列异步处理

---

## 总结

本次完善实现了告警与工单的深度联动，主要包括：
1. ✅ 告警合并逻辑 - 避免重复建单
2. ✅ 告警静默功能 - 维护期间屏蔽告警
3. ✅ 双向状态同步 - 工单和告警状态实时同步
4. ✅ 告警分组查询 - 快速了解告警分布

这些功能使得 TicketDesk 的告警管理更加智能和高效，符合生产环境的实际需求。
