# TicketDesk

TicketDesk 是一个面向运维与技术团队的 **项目化工单与告警联动系统**，用于统一管理研发 / 运维工单，并将监控告警自动转化为可跟踪、可调度、可审计、可统计的工单，形成完整的问题处理闭环。

TicketDesk 的核心理念是：

> **一切问题都是工单，一切告警都必须被跟进。**

---

## 1. 核心能力概览

TicketDesk 聚焦三个核心方向：

1️⃣ **项目化工单系统（Jira-like）**  
2️⃣ **告警自动建单与调度闭环**  
3️⃣ **审批 + 工作节点驱动的流程引擎**

并辅以 SLA、统计面板、审计与报告能力，支撑日常运维与故障管理。

---

## 2. 工单系统设计（Jira-like）

### 2.1 Project（项目）
- 工单的一级归属单位
- 不同项目可独立配置：
  - 工单类型（Issue Type）
  - 工作流（Workflow）
  - 自定义字段（Custom Fields）
  - SLA 策略
  - 成员与权限

示例：
- OPS-INFRA
- PAY-SERVICE
- GAME-BACKEND

---

### 2.2 Issue Type（工单类型）

内置推荐类型（可扩展）：

| 类型 | 说明 |
|---|---|
| Epic | 阶段性目标 / 大型需求 |
| Task | 普通任务 |
| Bug | 研发缺陷 |
| Fault | 生产故障 / 告警工单 |
| Change | 变更工单 |
| ServiceRequest | 服务申请 |

不同 Issue Type 可绑定不同工作流与字段模板。

---

### 2.3 Workflow（流程引擎：审批节点 + 工作节点）

TicketDesk 的 Workflow 不只是状态流转，而是 **由流程节点驱动的可编排流程引擎**。

工作流由 **Node（节点）+ Edge（流转条件）** 构成。

---

#### 2.3.1 节点类型

##### ✅ Approval Node（审批节点）
用于授权、合规、高风险操作前置检查。

支持能力：
- 审批人来源：
  - 指定用户
  - 角色
  - 用户组 / 值班组
- 审批策略：
  - 单人审批
  - 会签（AND）
  - 或签（OR）
  - 按比例通过（如 2/3）
- 超时处理：
  - 超时自动升级（Escalation）
- 审批结果：
  - 通过 / 拒绝 / 退回补充
- 全量审计：
  - 审批人
  - 时间
  - 意见
  - 附件

---

##### ✅ Work Node（工作节点）
用于实际执行、处理问题的阶段。

支持能力：
- 指派策略：
  - 指定负责人
  - 角色
  - 组 / 值班组
  - 标签路由（service → team）
- 子任务（Subtask / Checklist）
- 完成条件校验
- 处理产出沉淀：
  - 处理结论
  - 证据附件
  - 变更单号 / 发布链接 / 回滚方案

---

##### （可扩展）System Node（自动节点）
用于自动化动作（后续阶段）：
- 自动静默告警
- 自动回写告警系统
- 自动触发 CI/CD / 回滚
- 自动创建协作群 / 通知

---

#### 2.3.2 流程能力（MVP 要求）
- 串行节点：A → B → C
- 条件分支：按字段（环境 / 优先级 / 类型）分支
- 退回机制：审批拒绝 → 指定工作节点
- 状态由节点驱动自动变化

示例（高风险变更）：
```
Submit
→ Work(方案准备)
→ Approval(负责人审批)
→ Approval(安全审批)
→ Work(执行)
→ Work(验证)
→ Close
```

---

## 3. 工单基础能力

- 指派（Assignee）
- 关注人（Watcher）
- 评论（Comment）
- 附件（Attachment）
- 标签（Labels）
- 优先级（P0 ~ P3）
- 自定义字段（Custom Fields）
- 工单关联关系：
  - 父子工单
  - Epic ↔ Task
  - 阻塞 / 被阻塞
  - 重复 / 关联

---

## 4. 告警 → 工单联动设计（核心能力）

### 4.1 告警接入
TicketDesk 通过 API 对接外部监控 / 告警系统，周期性拉取 **活跃告警（Active Alerts）**。

支持来源：
- Prometheus / Alertmanager
- 云厂商监控
- APM / 日志告警系统

---

### 4.2 告警去重与合并（Alert Fingerprint）

通过 **告警指纹（Fingerprint）** 控制建单行为。

指纹建议字段：
- 告警规则 ID / 名称
- 资源标识（实例 / Pod / 服务）
- 环境（prod / staging）
- 集群 / 区域
- 核心标签（job / namespace / service）

处理策略：
- 无对应未关闭工单 → 创建新的 `Fault` 工单
- 已存在 → 追加告警事件到工单时间线

---

### 4.3 自动建单规则
- 告警严重级别 → 工单优先级
- 告警标签 → 项目映射
- 告警来源 → 工单组件 / 类型

---

### 4.4 告警生命周期联动
- 告警触发 → 建单 / 合并
- 告警恢复（Resolved）：
  - 追加事件
  - 自动建议进入“待关闭”节点
- 工单 Ack / 处理中：
  - 可回写告警系统（ack / comment / silence）
- 工单关闭：
  - 解除静默
  - 补充结论（可选）

---

## 5. 调度与升级（Escalation）

- 自动指派负责人 / 值班组
- 升级策略示例：
  - P0：5 分钟未 Ack → 升级值班负责人
  - 30 分钟未恢复 → 升级到组负责人
- 支持通知渠道：
  - IM
  - Email
  - Webhook

---

## 6. SLA 与效率指标

内置 SLA 计算：
- MTTA（响应时间）
- MTTR（修复时间）
- SLA 命中率
- 超时统计

支持：
- 按项目 / 类型 / 优先级配置
- SLA 暂停（等待用户 / 外部依赖）

---

## 7. 统计面板与报告

### 7.1 实时面板
- 未关闭工单数
- P0 / P1 分布
- 告警工单趋势
- SLA 倒计时

### 7.2 报告指标
- MTTA / MTTR
- 告警噪声比（重复 vs 有效）
- Top 根因 / 组件
- 项目 / 人员负载

支持导出：
- CSV
- PDF

---

## 8. 权限与审计

- RBAC 权限模型
- 项目级权限控制
- 全量审计日志：
  - 字段修改
  - 指派变化
  - 节点流转
  - 审批意见
  - 告警回写

---

## 9. 系统架构（建议）

模块划分：
- core-project
- core-issue
- core-workflow
- integration-alert
- scheduler
- notification
- reporting
- api

---

## 10. 技术栈建议

- Backend：Go
- API：REST (OpenAPI)
- DB：PostgreSQL / MySQL
- Cache / Queue：Redis（可选）
- Auth：JWT / OIDC（可对接统一 SSO）
- Deploy：Docker / Kubernetes

---

## 11. Roadmap

### Phase 1（MVP）
- 项目 & 工单基础
- 审批节点 + 工作节点（串行 + 退回）
- 告警自动建 Fault 工单
- SLA 基础能力
- 基础面板

### Phase 2
- 权限模型完善
- 值班表 / Oncall 集成
- 升级策略
- 高级搜索与视图

### Phase 3
- 并行审批
- 自动化节点
- Postmortem / 知识库
- CMDB / CI-CD 联动

---

## 12. Design Principles

- 工单是事实来源（Single Source of Truth）
- 告警不是问题，工单才是
- 自动化优先，人工兜底
- 可审计、可追责、可复盘

---

## 13. License
TBD
