# TicketDesk

TicketDesk 是一个面向运维与技术团队的 **项目化工单与告警联动系统**，用于统一管理研发/运维工单，并将监控告警自动转化为可跟踪、可调度、可审计、可统计的工单，形成完整的问题处理闭环。

> **一切问题都是工单，一切告警都必须被跟进。**

---

## 功能特性

### 🎫 项目化工单系统（Jira-like）
- **多项目管理**：独立配置工单类型、工作流、自定义字段、SLA 策略
- **工单类型**：Epic / Task / Bug / Fault / Change / ServiceRequest
- **工单能力**：指派、关注人、评论、附件、标签、优先级（P0~P3）、自定义字段
- **工单关联**：父子工单、阻塞/被阻塞、重复/关联

### 🔔 告警自动建单与调度闭环
- **告警接入**：支持 Prometheus / Alertmanager / 夜莺（N9E）等告警源
- **告警指纹去重**：基于规则 ID + 资源标识 + 标签计算指纹，自动去重合并
- **自动建单**：告警触发 → 自动创建 Fault 工单，告警恢复 → 自动关闭
- **合并窗口**：同名告警在时间窗口内合并到同一工单，超时后创建新工单并标记旧工单为「已合并」
- **告警静默**：支持标签匹配器（`==` / `!=` / `=~` / `!~`），维护窗口内自动抑制建单
- **双向联动**：工单状态变更同步告警状态，告警恢复同步工单状态

### ⚙️ 流程引擎（审批 + 工作节点）
- **可视化设计器**：拖拽式工作流编辑器（Vue Flow），所见即所得编排流程
- **审批节点**：单人审批 / 会签（AND）/ 或签（OR），支持超时升级
- **工作节点**：指派负责人、子任务、完成条件校验
- **自定义流转条件**：边支持预设条件（通过/拒绝）和自定义条件（如「验收通过」「需要修改」等任意名称）
- **动态分支**：审批节点和工作节点均支持多条件分支，工单详情页自动渲染对应操作按钮
- **流程编排**：串行节点、条件分支、退回机制，工单状态随节点流转自动联动

### 📊 统计与报表
- **实时面板**：未关闭工单数、优先级分布、告警趋势、SLA 倒计时
- **效率指标**：MTTA（响应时间）、MTTR（修复时间）、SLA 命中率
- **数据导出**：CSV

### 📋 需求池
- **双层级**：全局需求池（跨项目）+ 项目级需求池
- **生命周期**：待评审 → 评审中 → 已接受 → 转化为工单
- **看板视图**：按状态 / 优先级 / 负责人 / 时间线分组

### 🔐 权限与通知
- **RBAC 权限模型**：系统级 + 项目级角色控制
- **通知渠道**：站内信（WebSocket 实时推送）、邮件、Webhook
- **全量审计**：字段修改、指派变化、节点流转、审批意见

---

## 技术栈

| 层级 | 技术 |
|------|------|
| **后端** | Go 1.25 + Gin |
| **前端** | Vue 3 + TypeScript + Element Plus |
| **数据库** | MySQL 8.0 |
| **缓存** | Redis 7 |
| **认证** | JWT（Access Token + Refresh Token） |
| **日志** | Zap（结构化日志） |
| **配置** | Viper（YAML + 环境变量覆盖） |
| **ORM** | GORM |
| **部署** | Docker Compose / Kubernetes (Helm) |
| **CI/CD** | GitHub Actions |
| **镜像仓库** | GitHub Container Registry (ghcr.io) |

---

## 项目结构

```
ticketdesk/
├── cmd/ticketdesk/          # 应用入口
├── internal/                # 内部业务代码
│   ├── model/               # 数据模型（全局共享）
│   ├── api/                 # API 层（路由、中间件、统一响应）
│   ├── core-project/        # 项目管理模块
│   ├── core-issue/          # 工单管理模块
│   ├── core-workflow/       # 工作流引擎模块
│   ├── core-user/           # 用户管理模块
│   ├── core-field/          # 自定义字段模块
│   ├── integration-alert/   # 告警集成模块
│   ├── requirement-pool/    # 需求池模块
│   ├── activity/            # 活动日志模块
│   ├── notification/        # 通知模块（邮件/Webhook）
│   ├── notification-inbox/  # 站内信模块（WebSocket）
│   ├── reporting/           # 报表统计模块
│   ├── scheduler/           # 定时任务模块
│   └── system-config/       # 系统配置模块
├── pkg/                     # 可复用公共库
│   ├── config/              # 配置管理
│   ├── logger/              # 日志
│   ├── database/            # 数据库
│   ├── redis/               # Redis
│   └── jwt/                 # JWT 认证
├── configs/                 # 配置文件
├── deploy/                  # 部署配置
│   ├── docker-compose.yaml      # 生产环境（全栈）
│   ├── docker-compose.dev.yaml  # 开发环境（热更新）
│   ├── .env.example             # 环境变量模板
│   ├── docker/                  # Dockerfile 集合
│   └── helm/                    # Helm Chart（Kubernetes 部署）
├── web/                     # 前端代码（Vue 3）
├── scripts/                 # 脚本
├── Makefile                 # 构建命令
└── CLAUDE.md                # AI 开发规范
```

每个业务模块遵循分层架构：
```
module/
├── dto/            # 数据传输对象
├── handler/        # HTTP 处理层
├── service/        # 业务逻辑层
└── repository/     # 数据访问层
```

---

## 快速开始

### 环境要求

- Go 1.25+
- Node.js 20+
- MySQL 8.0+
- Redis 6.0+
- Docker & Docker Compose

### 方式一：生产部署（推荐）

一键部署全栈服务（MySQL + Redis + Backend + Nginx），无需本地安装数据库。

```bash
# 1. 克隆代码
git clone <repo-url> && cd TicketDesk

# 2. 配置环境变量
cp deploy/.env.example deploy/.env
vim deploy/.env   # 修改密码和 JWT_SECRET

# 3. 一键启动
make prod-d

# 4. 访问
open http://localhost
```

**管理命令：**

```bash
make prod-d        # 后台启动
make prod-stop     # 停止
make prod-logs     # 查看日志
make prod-ps       # 查看容器状态
make prod-rebuild  # 重建镜像并启动
```

### 方式二：本地开发

需要本地安装 MySQL 和 Redis。

```bash
# 1. 安装依赖
make init

# 2. 配置
cp configs/config.example.yaml configs/config.yaml
vim configs/config.yaml   # 修改数据库连接信息

# 3. 启动后端（热更新）
make dev-backend

# 4. 启动前端（另一个终端）
make dev-frontend

# 5. 访问
open http://localhost:3100
```

### 方式三：Docker 开发模式

后端 + 前端容器化热更新，MySQL/Redis 使用宿主机。

```bash
make docker-dev
```

### 方式四：Kubernetes 部署（Helm）

使用 Helm Chart 部署到 Kubernetes 集群，支持 Gateway API 和 Istio 两种流量入口。

```bash
# 1. 配置 values.yaml
vim deploy/helm/values.yaml

# 2. 安装
make helm-install

# 3. 查看状态
kubectl get pods -n ticketdesk
```

**流量入口切换：**

```yaml
# Gateway API（默认，K8s 新标准）
ingress:
  className: gateway

# Istio VirtualService
ingress:
  className: istio
  istio:
    gateway: istio-system/default-gateway
```

**使用外部数据库：**

```yaml
mysql:
  enabled: false
  external:
    host: rds.xxx.com
    password: xxx

redis:
  enabled: false
  external:
    host: redis.xxx.com
    password: xxx
```

**管理命令：**

```bash
make helm-lint      # 校验 Chart
make helm-template  # 预览渲染结果
make helm-install   # 首次安装
make helm-upgrade   # 升级部署
make helm-uninstall # 卸载
```

---

## 环境变量说明

生产部署通过 `deploy/.env` 配置，主要变量：

| 变量 | 说明 | 默认值 |
|------|------|--------|
| `DB_ROOT_PASSWORD` | MySQL root 密码 | `ticketdesk_root_2026` |
| `DB_NAME` | 数据库名 | `ticketdesk` |
| `DB_USER` | 数据库用户 | `ticketdesk` |
| `DB_PASSWORD` | 数据库密码 | `ticketdesk_pwd_2026` |
| `REDIS_PASSWORD` | Redis 密码 | `ticketdesk_redis_2026` |
| `JWT_SECRET` | JWT 密钥（**必须修改**） | - |
| `WEB_PORT` | 前端对外端口 | `80` |
| `APP_PORT` | 后端内部端口 | `10010` |
| `LOG_LEVEL` | 日志级别 | `info` |

完整变量列表见 [`deploy/.env.example`](deploy/.env.example)。

---

## 架构概览

```
                    ┌─────────────┐
                    │   Browser   │
                    └──────┬──────┘
                           │ :80
                    ┌──────▼──────┐
                    │    Nginx    │  静态资源 + 反向代理
                    │  (Frontend) │
                    └──────┬──────┘
                           │ /api/*
                    ┌──────▼──────┐
                    │   Backend   │  Go + Gin
                    │   (API)     │  :10010
                    └───┬─────┬───┘
                        │     │
                 ┌──────▼┐   ┌▼──────┐
                 │ MySQL │   │ Redis │
                 │  8.0  │   │  7    │
                 └───────┘   └───────┘

  外部告警源 ──webhook──▶ Backend ──自动建单──▶ Issue
  (Prometheus/N9E)
```

**Kubernetes 架构：**

```
                    ┌─────────────┐
                    │   Browser   │
                    └──────┬──────┘
                           │
                    ┌──────▼──────┐
                    │  Gateway /  │  Gateway API 或
                    │  Istio VS   │  Istio VirtualService
                    └──┬──────┬───┘
                /api/* │      │ /*
             ┌─────────▼┐   ┌▼─────────┐
             │ Backend  │   │ Frontend │
             │ (Pod x2) │   │ (Pod x2) │
             └────┬──┬──┘   └──────────┘
                  │  │
           ┌─────▼┐ ┌▼─────┐
           │MySQL │ │Redis │  集群内 StatefulSet
           │(STS) │ │(STS) │  或外部托管服务
           └──────┘ └──────┘
```

---

## 常用命令

```bash
# ============ 开发 ============
make dev-backend     # 后端热更新
make dev-frontend    # 前端热更新
make build           # 构建后端二进制
make test            # 运行测试
make lint            # 代码静态检查
make fmt             # 格式化代码
make swagger         # 生成 API 文档

# ============ Docker 开发 ============
make docker-dev      # 容器化开发（热更新）
make docker-dev-stop # 停止开发容器

# ============ 生产部署（Docker Compose）============
make prod-d          # 后台启动
make prod-stop       # 停止
make prod-logs       # 查看日志
make prod-rebuild    # 重建并启动
make prod-ps         # 容器状态

# ============ 生产部署（Kubernetes）============
make helm-lint       # 校验 Chart
make helm-template   # 预览渲染结果
make helm-install    # 首次安装
make helm-upgrade    # 升级部署
make helm-uninstall  # 卸载
```

---

## 默认账号

系统启动后会自动创建管理员账号：

| 用户名 | 密码 | 角色 |
|--------|------|------|
| `admin` | `admin123` | 系统管理员 |

> ⚠️ 生产环境请立即修改默认密码。

---

## 性能优化

系统针对百万级工单场景做了专项优化，确保大数据量下查询性能稳定。

### 优化策略

- **复合索引**：为工单表的常见查询模式（项目+状态、指派人+状态、报告人等）创建包含排序列的复合索引，避免 filesort
- **封顶计数**：分页查询的 COUNT 使用 `LIMIT 10001` 子查询封顶，避免全表扫描。超过 10,000 条时前端显示 "10,000+"
- **关键字搜索优化**：仅搜索 `issue_key` 和 `title` 字段，不扫描 `description`（TEXT 类型）
- **索引干扰消除**：移除 issues 表低选择性的 `deleted_at` 单列索引，防止优化器误选

### 性能基准（100 万工单，MySQL 8.4）

| 查询场景 | 耗时 |
|----------|------|
| 列表查询（项目+状态+分页 20 条） | **0.3ms** |
| 我的待办（指派人+排除关闭） | **0.3ms** |
| 按报告人查询 | **0.25ms** |
| 工单详情（issue_key 精确查询） | **0.17ms** |
| 多条件组合（项目+状态+优先级+类型） | **1.2ms** |
| 分页计数（封顶 COUNT） | **6.7ms** |
| 关键字搜索（LIKE 模糊匹配） | **~95ms** |

### 推荐配置

| 档位 | 配置 | InnoDB Buffer Pool | 适用场景 |
|------|------|-------------------|---------|
| 入门 | 4C 8G SSD | 4G | 百万工单，30 QPS |
| **推荐** | **8C 16G NVMe** | **10G** | **百万工单，100 QPS，全场景 <7ms** |
| 豪华 | 16C 32G NVMe | 20G | 千万工单，500 QPS |

---

## License

MIT
