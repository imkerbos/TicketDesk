# 需求池功能设计文档

## 1. 功能概述

需求池（Requirement Pool）是一个用于收集、评审和管理需求的系统，支持全局需求池和项目级需求池两种模式。

### 1.1 核心特性
- **双层级需求池**：支持全局需求池（跨项目）和项目级需求池
- **需求生命周期管理**：待评审 → 评审中 → 已接受 → 转化为工单
- **多维度看板**：按状态/优先级/负责人/时间线展示
- **统计报告**：支持时间范围统计和数据导出（CSV/Excel）
- **实时跟进**：每日/每周需求跟进提醒

## 2. 数据模型

### 2.1 需求池表（requirement_pools）
| 字段 | 类型 | 说明 |
|------|------|------|
| id | BIGINT | 主键 |
| name | VARCHAR(200) | 需求池名称 |
| description | TEXT | 描述 |
| type | VARCHAR(20) | 类型：global/project |
| project_id | BIGINT | 关联项目ID（项目级需求池） |
| owner_id | BIGINT | 负责人ID |
| status | VARCHAR(20) | 状态：active/archived |
| created_at | DATETIME | 创建时间 |
| updated_at | DATETIME | 更新时间 |
| deleted_at | DATETIME | 删除时间 |

### 2.2 需求表（requirements）
| 字段 | 类型 | 说明 |
|------|------|------|
| id | BIGINT | 主键 |
| pool_id | BIGINT | 所属需求池ID |
| title | VARCHAR(500) | 标题 |
| description | TEXT | 描述 |
| priority | VARCHAR(10) | 优先级：P0/P1/P2/P3 |
| status | VARCHAR(20) | 状态：pending/reviewing/accepted/rejected/converted |
| assignee_id | BIGINT | 负责人ID |
| estimated_hours | DECIMAL(10,2) | 预估工时 |
| target_date | DATE | 目标完成时间 |
| converted_issue_id | BIGINT | 转化后的工单ID |
| target_project_id | BIGINT | 目标项目ID |
| created_by | BIGINT | 创建人ID |
| tags | VARCHAR(500) | 标签（逗号分隔） |
| created_at | DATETIME | 创建时间 |
| updated_at | DATETIME | 更新时间 |
| deleted_at | DATETIME | 删除时间 |

### 2.3 需求评论表（requirement_comments）
| 字段 | 类型 | 说明 |
|------|------|------|
| id | BIGINT | 主键 |
| requirement_id | BIGINT | 需求ID |
| user_id | BIGINT | 用户ID |
| content | TEXT | 评论内容 |
| created_at | DATETIME | 创建时间 |
| updated_at | DATETIME | 更新时间 |
| deleted_at | DATETIME | 删除时间 |

### 2.4 需求附件表（requirement_attachments）
| 字段 | 类型 | 说明 |
|------|------|------|
| id | BIGINT | 主键 |
| requirement_id | BIGINT | 需求ID |
| file_name | VARCHAR(255) | 文件名 |
| file_path | VARCHAR(500) | 文件路径 |
| file_size | BIGINT | 文件大小 |
| file_type | VARCHAR(100) | 文件类型 |
| uploaded_by | BIGINT | 上传人ID |
| created_at | DATETIME | 创建时间 |
| deleted_at | DATETIME | 删除时间 |

## 3. API 设计

### 3.1 需求池 API

#### 3.1.1 创建需求池
```
POST /api/v1/requirement-pools
```

**请求体：**
```json
{
  "name": "产品需求池",
  "description": "产品团队的需求收集池",
  "type": "global",
  "owner_id": 1
}
```

#### 3.1.2 获取需求池列表
```
GET /api/v1/requirement-pools?type=global&status=active&page=1&page_size=20
```

#### 3.1.3 获取需求池详情
```
GET /api/v1/requirement-pools/:id
```

#### 3.1.4 更新需求池
```
PUT /api/v1/requirement-pools/:id
```

#### 3.1.5 删除需求池
```
DELETE /api/v1/requirement-pools/:id
```

### 3.2 需求 API

#### 3.2.1 创建需求
```
POST /api/v1/requirements
```

**请求体：**
```json
{
  "pool_id": 1,
  "title": "用户登录功能优化",
  "description": "优化用户登录流程，支持第三方登录",
  "priority": "P1",
  "assignee_id": 2,
  "estimated_hours": 16,
  "target_date": "2026-03-01",
  "target_project_id": 1,
  "tags": ["登录", "优化"]
}
```

#### 3.2.2 获取需求列表
```
GET /api/v1/requirements?pool_id=1&status=pending&priority=P1&page=1&page_size=20
```

#### 3.2.3 获取需求详情
```
GET /api/v1/requirements/:id
```

#### 3.2.4 更新需求
```
PUT /api/v1/requirements/:id
```

#### 3.2.5 删除需求
```
DELETE /api/v1/requirements/:id
```

#### 3.2.6 转化为工单
```
POST /api/v1/requirements/:id/convert
```

**请求体：**
```json
{
  "project_id": 1,
  "issue_type_id": 1,
  "assignee_id": 2
}
```

#### 3.2.7 添加评论
```
POST /api/v1/requirements/:id/comments
```

**请求体：**
```json
{
  "content": "这个需求很重要，建议优先处理"
}
```

#### 3.2.8 上传附件
```
POST /api/v1/requirements/:id/attachments
```

### 3.3 看板 API

#### 3.3.1 获取看板数据
```
GET /api/v1/requirement-pools/:id/board?group_by=status
```

**group_by 参数：**
- `status`：按状态分组（待评审/评审中/已接受/已拒绝/已转化）
- `priority`：按优先级分组（P0/P1/P2/P3）
- `assignee`：按负责人分组
- `timeline`：按时间线分组（本周/本月/本季度）

**响应示例：**
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "group_by": "status",
    "columns": [
      {
        "key": "pending",
        "title": "待评审",
        "count": 10,
        "requirements": [...]
      },
      {
        "key": "reviewing",
        "title": "评审中",
        "count": 5,
        "requirements": [...]
      }
    ],
    "total": 15
  }
}
```

#### 3.3.2 更新需求状态（拖拽）
```
PUT /api/v1/requirements/:id/status
```

**请求体：**
```json
{
  "status": "reviewing"
}
```

### 3.4 报告 API

#### 3.4.1 获取统计数据
```
GET /api/v1/requirement-pools/:id/statistics?start_date=2026-01-01&end_date=2026-01-31&group_by=day
```

**响应示例：**
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "pool_id": 1,
    "pool_name": "产品需求池",
    "start_date": "2026-01-01",
    "end_date": "2026-01-31",
    "total_created": 50,
    "total_completed": 30,
    "total_converted": 25,
    "total_rejected": 5,
    "avg_process_days": 7.5,
    "status_summary": [
      {"status": "pending", "count": 10},
      {"status": "reviewing", "count": 5},
      {"status": "accepted", "count": 5},
      {"status": "converted", "count": 25},
      {"status": "rejected", "count": 5}
    ],
    "priority_summary": [
      {"priority": "P0", "count": 5},
      {"priority": "P1", "count": 15},
      {"priority": "P2", "count": 20},
      {"priority": "P3", "count": 10}
    ],
    "assignee_summary": [
      {
        "assignee_id": 1,
        "assignee_name": "张三",
        "total": 20,
        "completed": 15,
        "in_progress": 5
      }
    ],
    "trend_data": [
      {
        "date": "2026-01-01",
        "created": 2,
        "completed": 1,
        "converted": 1
      }
    ]
  }
}
```

#### 3.4.2 导出数据
```
GET /api/v1/requirement-pools/:id/export?format=csv&start_date=2026-01-01&end_date=2026-01-31
```

**format 参数：**
- `csv`：导出为 CSV 格式
- `xlsx`：导出为 Excel 格式

## 4. 前端页面设计

### 4.1 需求池列表页
- 展示所有需求池（全局 + 项目级）
- 支持筛选（类型、状态、负责人）
- 支持搜索
- 创建新需求池按钮

### 4.2 需求池详情页
- 需求池基本信息
- 需求列表（表格视图）
- 快速创建需求
- 切换到看板视图按钮

### 4.3 需求看板页
- 多维度看板切换（状态/优先级/负责人/时间线）
- 拖拽改变状态
- 快速编辑需求
- 筛选和搜索

### 4.4 需求详情页
- 需求基本信息
- 评论区
- 附件列表
- 转化为工单按钮
- 操作历史

### 4.5 需求报告页
- 统计图表（柱状图、饼图、折线图）
- 时间范围选择
- 导出功能（CSV/Excel）

## 5. 实现计划

### Phase 1：基础功能（1-2周）
- [ ] 数据模型创建（Model）
- [ ] DTO 定义
- [ ] Repository 层实现
- [ ] Service 层实现
- [ ] Handler 层实现
- [ ] 路由注册
- [ ] 数据库迁移

### Phase 2：看板功能（1周）
- [ ] 看板数据查询逻辑
- [ ] 多维度分组实现
- [ ] 拖拽状态更新
- [ ] 前端看板组件

### Phase 3：报告功能（1周）
- [ ] 统计数据计算
- [ ] 趋势数据生成
- [ ] CSV 导出
- [ ] Excel 导出
- [ ] 前端报告页面

### Phase 4：高级功能（1周）
- [ ] 需求评论
- [ ] 需求附件上传
- [ ] 需求转化为工单
- [ ] 定时提醒（每日/每周）
- [ ] 通知集成

## 6. 技术要点

### 6.1 需求转化为工单
需求转化为工单时，需要：
1. 创建新工单，复制需求的标题、描述等信息
2. 更新需求状态为 `converted`
3. 记录 `converted_issue_id`
4. 保持需求与工单的关联关系

### 6.2 看板分组逻辑
- **按状态分组**：固定的 5 个状态列
- **按优先级分组**：固定的 4 个优先级列
- **按负责人分组**：动态生成，根据实际负责人
- **按时间线分组**：本周/本月/本季度，根据 `target_date` 分组

### 6.3 统计报告计算
- **平均处理天数**：从创建到完成的平均天数
- **趋势数据**：按天/周/月分组统计
- **负责人统计**：每个负责人的需求数量和完成情况

### 6.4 数据导出
- **CSV 导出**：使用 `encoding/csv` 包
- **Excel 导出**：使用 `github.com/xuri/excelize/v2` 包

## 7. 数据库索引优化

```sql
-- 需求池表索引
CREATE INDEX idx_requirement_pools_type ON requirement_pools(type);
CREATE INDEX idx_requirement_pools_status ON requirement_pools(status);
CREATE INDEX idx_requirement_pools_owner_id ON requirement_pools(owner_id);
CREATE INDEX idx_requirement_pools_project_id ON requirement_pools(project_id);

-- 需求表索引
CREATE INDEX idx_requirements_pool_id ON requirements(pool_id);
CREATE INDEX idx_requirements_status ON requirements(status);
CREATE INDEX idx_requirements_priority ON requirements(priority);
CREATE INDEX idx_requirements_assignee_id ON requirements(assignee_id);
CREATE INDEX idx_requirements_created_by ON requirements(created_by);
CREATE INDEX idx_requirements_target_date ON requirements(target_date);
CREATE INDEX idx_requirements_created_at ON requirements(created_at);

-- 复合索引（用于常见查询）
CREATE INDEX idx_requirements_pool_status ON requirements(pool_id, status);
CREATE INDEX idx_requirements_pool_priority ON requirements(pool_id, priority);
CREATE INDEX idx_requirements_assignee_status ON requirements(assignee_id, status);
```

## 8. 权限控制

### 8.1 需求池权限
- **全局需求池**：只有管理员和需求池负责人可以编辑
- **项目级需求池**：项目成员可以查看，项目管理员可以编辑

### 8.2 需求权限
- **创建需求**：所有登录用户
- **编辑需求**：需求创建人、需求负责人、需求池负责人
- **删除需求**：需求创建人、需求池负责人
- **转化为工单**：需求池负责人、项目管理员

## 9. 通知规则

### 9.1 需求创建通知
- 通知需求池负责人
- 通知需求负责人（如果指定）

### 9.2 需求状态变更通知
- 通知需求创建人
- 通知需求负责人

### 9.3 需求转化为工单通知
- 通知需求创建人
- 通知工单负责人

### 9.4 定时提醒
- **每日提醒**：提醒负责人今日到期的需求
- **每周提醒**：提醒负责人本周到期的需求
- **逾期提醒**：提醒负责人已逾期的需求

## 10. 前端技术栈

- **框架**：Vue 3 + TypeScript
- **UI 组件库**：Element Plus
- **状态管理**：Pinia
- **路由**：Vue Router
- **HTTP 客户端**：Axios
- **看板组件**：Vue Draggable
- **图表库**：ECharts

## 11. 后续优化方向

1. **需求模板**：支持创建需求模板，快速创建需求
2. **需求批量操作**：批量修改状态、优先级、负责人
3. **需求关联**：需求之间的依赖关系
4. **需求投票**：团队成员对需求进行投票
5. **需求评分**：对需求的价值、紧急程度、复杂度进行评分
6. **需求归档**：自动归档已完成的需求
7. **需求搜索**：全文搜索需求内容
8. **需求导入**：从 Excel/CSV 批量导入需求
