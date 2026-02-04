# 项目角色与工作流完善计划

## 一、需求分析

### 1.1 当前系统现状

**已实现的功能：**
- ✅ 基础项目管理（Project、ProjectMember）
- ✅ 简单的项目成员角色（owner, admin, member）
- ✅ Issue Type 管理（全局和项目级别）
- ✅ 工作流基础框架（Workflow、WorkflowNode、WorkflowEdge）
- ✅ 系统级 RBAC（admin, user, project_admin）
- ✅ 工单管理（Issue、Comment、Watcher、Worklog）

**需要补充的功能：**
- ❌ 可配置的项目角色系统（Project Roles）
- ❌ 自定义字段系统（Custom Fields）
- ❌ 工作流与 Issue Type 的关联
- ❌ 工作流实例和状态管理
- ❌ 审批节点与项目角色的集成
- ❌ 基于项目角色的细粒度权限控制

### 1.2 参考 Jira 模型

**Jira 的核心概念：**
1. **Project Roles**：项目角色（如 Administrators, Developers, Users）
2. **Issue Types**：工单类型（Epic, Story, Task, Bug）
3. **Custom Fields**：自定义字段（可配置字段类型和选项）
4. **Workflows**：工作流（状态流转、条件、验证器、后处理）
5. **Screens**：屏幕配置（不同操作显示不同字段）
6. **Permission Schemes**：权限方案（基于角色的权限）
7. **Workflow Schemes**：工作流方案（Issue Type 与 Workflow 的映射）

---

## 二、技术方案设计

### 2.1 数据模型设计

#### 2.1.1 项目角色（Project Role）

```go
// ProjectRole 项目角色定义
type ProjectRole struct {
    BaseModel
    ProjectID   uint64 `gorm:"not null;index:idx_project_role_project" json:"project_id"`
    RoleKey     string `gorm:"size:50;not null" json:"role_key"`          // 角色标识，如 developers, testers
    RoleName    string `gorm:"size:100;not null" json:"role_name"`        // 角色名称
    Description string `gorm:"size:500" json:"description"`               // 角色描述
    IsDefault   bool   `gorm:"default:false" json:"is_default"`           // 是否默认角色

    // 关联
    Project     Project              `gorm:"foreignKey:ProjectID" json:"project,omitempty"`
    Members     []ProjectRoleMember  `gorm:"foreignKey:RoleID" json:"members,omitempty"`
}

// ProjectRoleMember 项目角色成员（多对多关系）
type ProjectRoleMember struct {
    BaseModel
    ProjectID uint64 `gorm:"not null;uniqueIndex:idx_project_role_member" json:"project_id"`
    RoleID    uint64 `gorm:"not null;uniqueIndex:idx_project_role_member" json:"role_id"`
    UserID    uint64 `gorm:"not null;uniqueIndex:idx_project_role_member" json:"user_id"`

    // 关联
    Project Project     `gorm:"foreignKey:ProjectID" json:"project,omitempty"`
    Role    ProjectRole `gorm:"foreignKey:RoleID" json:"role,omitempty"`
    User    User        `gorm:"foreignKey:UserID" json:"user,omitempty"`
}
```

#### 2.1.2 自定义字段（Custom Field）

```go
// CustomField 自定义字段定义
type CustomField struct {
    BaseModel
    ProjectID   *uint64 `gorm:"index:idx_custom_field_project" json:"project_id"` // null 表示全局字段
    FieldKey    string  `gorm:"size:50;not null" json:"field_key"`                // 字段标识
    FieldName   string  `gorm:"size:100;not null" json:"field_name"`              // 字段名称
    FieldType   string  `gorm:"size:20;not null" json:"field_type"`               // 字段类型
    Description string  `gorm:"size:500" json:"description"`                      // 字段描述
    IsRequired  bool    `gorm:"default:false" json:"is_required"`                 // 是否必填
    DefaultValue string `gorm:"size:500" json:"default_value"`                    // 默认值
    Options     string  `gorm:"type:json" json:"options"`                         // 选项配置（JSON）
    Validation  string  `gorm:"type:json" json:"validation"`                      // 验证规则（JSON）
    SortOrder   int     `gorm:"default:0" json:"sort_order"`                      // 排序

    // 关联
    Project Project `gorm:"foreignKey:ProjectID" json:"project,omitempty"`
}

// FieldType 枚举
const (
    FieldTypeText       = "text"        // 单行文本
    FieldTypeTextarea   = "textarea"    // 多行文本
    FieldTypeNumber     = "number"      // 数字
    FieldTypeDate       = "date"        // 日期
    FieldTypeDatetime   = "datetime"    // 日期时间
    FieldTypeSelect     = "select"      // 单选下拉
    FieldTypeMultiSelect = "multiselect" // 多选
    FieldTypeCheckbox   = "checkbox"    // 复选框
    FieldTypeRadio      = "radio"       // 单选按钮
    FieldTypeUser       = "user"        // 用户选择器
    FieldTypeURL        = "url"         // URL
    FieldTypeEmail      = "email"       // 邮箱
)

// IssueTypeField Issue Type 与自定义字段的关联
type IssueTypeField struct {
    BaseModel
    IssueTypeID   uint64 `gorm:"not null;uniqueIndex:idx_issue_type_field" json:"issue_type_id"`
    CustomFieldID uint64 `gorm:"not null;uniqueIndex:idx_issue_type_field" json:"custom_field_id"`
    IsRequired    bool   `gorm:"default:false" json:"is_required"`  // 是否必填（可覆盖字段默认设置）
    SortOrder     int    `gorm:"default:0" json:"sort_order"`       // 排序

    // 关联
    IssueType   IssueType   `gorm:"foreignKey:IssueTypeID" json:"issue_type,omitempty"`
    CustomField CustomField `gorm:"foreignKey:CustomFieldID" json:"custom_field,omitempty"`
}

// IssueFieldValue 工单自定义字段值
type IssueFieldValue struct {
    BaseModel
    IssueID       uint64 `gorm:"not null;uniqueIndex:idx_issue_field_value" json:"issue_id"`
    CustomFieldID uint64 `gorm:"not null;uniqueIndex:idx_issue_field_value" json:"custom_field_id"`
    FieldValue    string `gorm:"type:text" json:"field_value"` // 字段值（JSON 格式存储）

    // 关联
    Issue       Issue       `gorm:"foreignKey:IssueID" json:"issue,omitempty"`
    CustomField CustomField `gorm:"foreignKey:CustomFieldID" json:"custom_field,omitempty"`
}
```

#### 2.1.3 工作流方案（Workflow Scheme）

```go
// WorkflowScheme 工作流方案（Issue Type 与 Workflow 的映射）
type WorkflowScheme struct {
    BaseModel
    ProjectID   uint64 `gorm:"not null;index:idx_workflow_scheme_project" json:"project_id"`
    Name        string `gorm:"size:100;not null" json:"name"`
    Description string `gorm:"size:500" json:"description"`
    IsDefault   bool   `gorm:"default:false" json:"is_default"` // 是否默认方案

    // 关联
    Project  Project                  `gorm:"foreignKey:ProjectID" json:"project,omitempty"`
    Mappings []WorkflowSchemeMapping  `gorm:"foreignKey:SchemeID" json:"mappings,omitempty"`
}

// WorkflowSchemeMapping Issue Type 与 Workflow 的映射关系
type WorkflowSchemeMapping struct {
    BaseModel
    SchemeID    uint64 `gorm:"not null;uniqueIndex:idx_workflow_scheme_mapping" json:"scheme_id"`
    IssueTypeID uint64 `gorm:"not null;uniqueIndex:idx_workflow_scheme_mapping" json:"issue_type_id"`
    WorkflowID  uint64 `gorm:"not null" json:"workflow_id"`

    // 关联
    Scheme    WorkflowScheme `gorm:"foreignKey:SchemeID" json:"scheme,omitempty"`
    IssueType IssueType      `gorm:"foreignKey:IssueTypeID" json:"issue_type,omitempty"`
    Workflow  Workflow       `gorm:"foreignKey:WorkflowID" json:"workflow,omitempty"`
}
```

#### 2.1.4 工作流实例（Workflow Instance）

```go
// WorkflowInstance 工作流实例
type WorkflowInstance struct {
    BaseModel
    IssueID        uint64 `gorm:"not null;index:idx_workflow_instance_issue" json:"issue_id"`
    WorkflowID     uint64 `gorm:"not null" json:"workflow_id"`
    CurrentNodeID  uint64 `gorm:"not null" json:"current_node_id"`
    Status         string `gorm:"size:20;not null" json:"status"` // active, completed, cancelled
    StartedAt      *time.Time `json:"started_at"`
    CompletedAt    *time.Time `json:"completed_at"`

    // 关联
    Issue       Issue        `gorm:"foreignKey:IssueID" json:"issue,omitempty"`
    Workflow    Workflow     `gorm:"foreignKey:WorkflowID" json:"workflow,omitempty"`
    CurrentNode WorkflowNode `gorm:"foreignKey:CurrentNodeID" json:"current_node,omitempty"`
    History     []WorkflowHistory `gorm:"foreignKey:InstanceID" json:"history,omitempty"`
}

// WorkflowHistory 工作流流转历史
type WorkflowHistory struct {
    BaseModel
    InstanceID    uint64 `gorm:"not null;index:idx_workflow_history_instance" json:"instance_id"`
    FromNodeID    *uint64 `json:"from_node_id"`
    ToNodeID      uint64 `gorm:"not null" json:"to_node_id"`
    TransitionBy  uint64 `gorm:"not null" json:"transition_by"`
    TransitionAt  time.Time `gorm:"not null" json:"transition_at"`
    Comment       string `gorm:"type:text" json:"comment"`
    Action        string `gorm:"size:50" json:"action"` // forward, backward, approve, reject

    // 关联
    Instance     WorkflowInstance `gorm:"foreignKey:InstanceID" json:"instance,omitempty"`
    FromNode     *WorkflowNode    `gorm:"foreignKey:FromNodeID" json:"from_node,omitempty"`
    ToNode       WorkflowNode     `gorm:"foreignKey:ToNodeID" json:"to_node,omitempty"`
    TransitionUser User           `gorm:"foreignKey:TransitionBy" json:"transition_user,omitempty"`
}
```

#### 2.1.5 审批记录（Approval）

```go
// ApprovalRecord 审批记录
type ApprovalRecord struct {
    BaseModel
    InstanceID    uint64 `gorm:"not null;index:idx_approval_record_instance" json:"instance_id"`
    NodeID        uint64 `gorm:"not null" json:"node_id"`
    ApproverID    uint64 `gorm:"not null" json:"approver_id"`
    Status        string `gorm:"size:20;not null" json:"status"` // pending, approved, rejected
    Comment       string `gorm:"type:text" json:"comment"`
    ApprovedAt    *time.Time `json:"approved_at"`

    // 关联
    Instance  WorkflowInstance `gorm:"foreignKey:InstanceID" json:"instance,omitempty"`
    Node      WorkflowNode     `gorm:"foreignKey:NodeID" json:"node,omitempty"`
    Approver  User             `gorm:"foreignKey:ApproverID" json:"approver,omitempty"`
}
```

#### 2.1.6 权限方案（Permission Scheme）

```go
// PermissionScheme 权限方案
type PermissionScheme struct {
    BaseModel
    ProjectID   uint64 `gorm:"not null;index:idx_permission_scheme_project" json:"project_id"`
    Name        string `gorm:"size:100;not null" json:"name"`
    Description string `gorm:"size:500" json:"description"`
    IsDefault   bool   `gorm:"default:false" json:"is_default"`

    // 关联
    Project     Project                `gorm:"foreignKey:ProjectID" json:"project,omitempty"`
    Permissions []PermissionGrant      `gorm:"foreignKey:SchemeID" json:"permissions,omitempty"`
}

// PermissionGrant 权限授予
type PermissionGrant struct {
    BaseModel
    SchemeID       uint64 `gorm:"not null;index:idx_permission_grant_scheme" json:"scheme_id"`
    Permission     string `gorm:"size:50;not null" json:"permission"`      // 权限标识
    GranteeType    string `gorm:"size:20;not null" json:"grantee_type"`    // role, user, group
    GranteeID      uint64 `gorm:"not null" json:"grantee_id"`              // 角色ID/用户ID

    // 关联
    Scheme PermissionScheme `gorm:"foreignKey:SchemeID" json:"scheme,omitempty"`
}

// Permission 枚举
const (
    // 项目权限
    PermProjectView   = "project.view"
    PermProjectEdit   = "project.edit"
    PermProjectDelete = "project.delete"
    PermProjectAdmin  = "project.admin"

    // 工单权限
    PermIssueCreate  = "issue.create"
    PermIssueView    = "issue.view"
    PermIssueEdit    = "issue.edit"
    PermIssueDelete  = "issue.delete"
    PermIssueAssign  = "issue.assign"
    PermIssueComment = "issue.comment"
    PermIssueTransition = "issue.transition"

    // 工作流权限
    PermWorkflowView   = "workflow.view"
    PermWorkflowEdit   = "workflow.edit"
    PermWorkflowDelete = "workflow.delete"

    // 成员权限
    PermMemberView   = "member.view"
    PermMemberAdd    = "member.add"
    PermMemberRemove = "member.remove"
)
```

### 2.2 模块设计

#### 2.2.1 项目角色模块（core-project-role）

```
internal/core-project-role/
├── dto/
│   └── project_role_dto.go       # DTO 定义
├── repository/
│   └── project_role_repository.go # 数据访问层
├── service/
│   └── project_role_service.go    # 业务逻辑层
└── handler/
    └── project_role_handler.go    # HTTP 处理层
```

**核心功能：**
- 项目角色 CRUD
- 角色成员管理（添加/移除用户）
- 默认角色初始化（Administrators, Developers, Users）
- 角色权限查询

#### 2.2.2 自定义字段模块（core-custom-field）

```
internal/core-custom-field/
├── dto/
│   └── custom_field_dto.go
├── repository/
│   └── custom_field_repository.go
├── service/
│   └── custom_field_service.go
└── handler/
    └── custom_field_handler.go
```

**核心功能：**
- 自定义字段 CRUD
- 字段与 Issue Type 关联
- 字段值验证
- 字段值存储和查询

#### 2.2.3 工作流方案模块（core-workflow-scheme）

```
internal/core-workflow-scheme/
├── dto/
│   └── workflow_scheme_dto.go
├── repository/
│   └── workflow_scheme_repository.go
├── service/
│   └── workflow_scheme_service.go
└── handler/
    └── workflow_scheme_handler.go
```

**核心功能：**
- 工作流方案 CRUD
- Issue Type 与 Workflow 映射管理
- 默认方案配置

#### 2.2.4 工作流引擎增强（core-workflow）

**新增功能：**
- 工作流实例管理
- 流程流转逻辑
- 审批节点处理
- 审批记录管理
- 流转历史记录

#### 2.2.5 权限方案模块（core-permission）

```
internal/core-permission/
├── dto/
│   └── permission_dto.go
├── repository/
│   └── permission_repository.go
├── service/
│   └── permission_service.go
└── handler/
    └── permission_handler.go
```

**核心功能：**
- 权限方案 CRUD
- 权限授予管理
- 权限检查服务
- 默认权限方案初始化

---

## 三、实施计划

### Phase 1：项目角色系统（预计 3-4 天）

#### Task 1.1：数据模型实现
- [ ] 在 `internal/model/model.go` 中添加 `ProjectRole` 和 `ProjectRoleMember` 模型
- [ ] 更新数据库迁移逻辑
- [ ] 添加默认角色种子数据（Administrators, Developers, Users）

#### Task 1.2：Repository 层实现
- [ ] 创建 `internal/core-project-role/repository/project_role_repository.go`
- [ ] 实现角色 CRUD 方法
- [ ] 实现角色成员管理方法
- [ ] 实现角色查询方法（按项目、按用户）

#### Task 1.3：Service 层实现
- [ ] 创建 `internal/core-project-role/service/project_role_service.go`
- [ ] 实现业务逻辑和验证
- [ ] 实现默认角色初始化逻辑
- [ ] 添加错误处理

#### Task 1.4：Handler 层实现
- [ ] 创建 `internal/core-project-role/handler/project_role_handler.go`
- [ ] 实现 RESTful API 端点
- [ ] 添加 Swagger 文档注释
- [ ] 添加权限检查中间件

#### Task 1.5：DTO 定义
- [ ] 创建 `internal/core-project-role/dto/project_role_dto.go`
- [ ] 定义请求和响应 DTO
- [ ] 添加参数验证规则

#### Task 1.6：路由注册
- [ ] 在 `internal/api/router/router.go` 中注册项目角色路由
- [ ] 配置权限中间件

#### Task 1.7：测试
- [ ] 编写单元测试
- [ ] 编写集成测试
- [ ] API 测试

### Phase 2：自定义字段系统（预计 4-5 天）

#### Task 2.1：数据模型实现
- [ ] 添加 `CustomField`、`IssueTypeField`、`IssueFieldValue` 模型
- [ ] 更新数据库迁移
- [ ] 设计字段类型枚举和验证规则

#### Task 2.2：Repository 层实现
- [ ] 创建 `custom_field_repository.go`
- [ ] 实现字段 CRUD 方法
- [ ] 实现字段与 Issue Type 关联方法
- [ ] 实现字段值存储和查询方法

#### Task 2.3：Service 层实现
- [ ] 创建 `custom_field_service.go`
- [ ] 实现字段验证逻辑（根据字段类型）
- [ ] 实现字段值转换和存储逻辑
- [ ] 实现字段配置管理

#### Task 2.4：Handler 层实现
- [ ] 创建 `custom_field_handler.go`
- [ ] 实现字段管理 API
- [ ] 实现字段与 Issue Type 关联 API
- [ ] 添加 Swagger 文档

#### Task 2.5：集成到工单模块
- [ ] 修改 `Issue` 创建/更新逻辑，支持自定义字段
- [ ] 修改 `Issue` 查询逻辑，返回自定义字段值
- [ ] 添加字段值验证

#### Task 2.6：测试
- [ ] 单元测试（各种字段类型的验证）
- [ ] 集成测试
- [ ] API 测试

### Phase 3：工作流方案系统（预计 3-4 天）

#### Task 3.1：数据模型实现
- [ ] 添加 `WorkflowScheme` 和 `WorkflowSchemeMapping` 模型
- [ ] 更新数据库迁移
- [ ] 添加默认工作流方案

#### Task 3.2：Repository 层实现
- [ ] 创建 `workflow_scheme_repository.go`
- [ ] 实现方案 CRUD 方法
- [ ] 实现映射管理方法
- [ ] 实现根据 Issue Type 查询 Workflow 的方法

#### Task 3.3：Service 层实现
- [ ] 创建 `workflow_scheme_service.go`
- [ ] 实现方案管理逻辑
- [ ] 实现映射管理逻辑
- [ ] 实现默认方案初始化

#### Task 3.4：Handler 层实现
- [ ] 创建 `workflow_scheme_handler.go`
- [ ] 实现 API 端点
- [ ] 添加 Swagger 文档

#### Task 3.5：集成到工单模块
- [ ] 修改工单创建逻辑，根据 Issue Type 自动关联 Workflow
- [ ] 创建工作流实例

#### Task 3.6：测试
- [ ] 单元测试
- [ ] 集成测试
- [ ] API 测试

### Phase 4：工作流实例与审批系统（预计 5-6 天）

#### Task 4.1：数据模型实现
- [ ] 添加 `WorkflowInstance`、`WorkflowHistory`、`ApprovalRecord` 模型
- [ ] 更新数据库迁移

#### Task 4.2：工作流引擎增强
- [ ] 在 `core-workflow/service/` 中创建 `workflow_instance_service.go`
- [ ] 实现工作流实例创建逻辑
- [ ] 实现流程流转逻辑（forward, backward）
- [ ] 实现审批节点处理逻辑（单人审批、会签、或签）
- [ ] 实现工作节点处理逻辑（自动分配）

#### Task 4.3：审批服务实现
- [ ] 创建 `approval_service.go`
- [ ] 实现审批提交逻辑
- [ ] 实现审批状态检查
- [ ] 实现审批通知

#### Task 4.4：与项目角色集成
- [ ] 修改 WorkflowNode 配置，支持项目角色作为审批人
- [ ] 实现根据项目角色查找审批人逻辑
- [ ] 实现审批权限检查

#### Task 4.5：Handler 层实现
- [ ] 创建 `workflow_instance_handler.go`
- [ ] 实现流转 API（提交审批、通过、拒绝、退回）
- [ ] 实现审批历史查询 API
- [ ] 添加 Swagger 文档

#### Task 4.6：集成到工单模块
- [ ] 修改工单状态流转逻辑，使用工作流引擎
- [ ] 修改工单详情 API，返回工作流状态
- [ ] 添加工单审批操作 API

#### Task 4.7：测试
- [ ] 单元测试（各种审批类型）
- [ ] 集成测试（完整流程）
- [ ] API 测试

### Phase 5：权限方案系统（预计 4-5 天）

#### Task 5.1：数据模型实现
- [ ] 添加 `PermissionScheme` 和 `PermissionGrant` 模型
- [ ] 更新数据库迁移
- [ ] 定义权限枚举

#### Task 5.2：Repository 层实现
- [ ] 创建 `permission_repository.go`
- [ ] 实现方案 CRUD 方法
- [ ] 实现权限授予管理方法
- [ ] 实现权限查询方法

#### Task 5.3：Service 层实现
- [ ] 创建 `permission_service.go`
- [ ] 实现权限检查逻辑
- [ ] 实现默认权限方案初始化
- [ ] 实现权限继承逻辑

#### Task 5.4：权限中间件增强
- [ ] 修改 `internal/api/middleware/rbac.go`
- [ ] 添加项目级权限检查
- [ ] 添加工单级权限检查
- [ ] 集成项目角色权限

#### Task 5.5：Handler 层实现
- [ ] 创建 `permission_handler.go`
- [ ] 实现权限方案管理 API
- [ ] 实现权限授予管理 API
- [ ] 添加 Swagger 文档

#### Task 5.6：集成到各模块
- [ ] 在项目管理模块中添加权限检查
- [ ] 在工单管理模块中添加权限检查
- [ ] 在工作流模块中添加权限检查

#### Task 5.7：测试
- [ ] 单元测试
- [ ] 权限测试（各种场景）
- [ ] API 测试

### Phase 6：前端开发（预计 6-8 天）

#### Task 6.1：项目角色管理页面
- [ ] 项目角色列表页面
- [ ] 项目角色创建/编辑页面
- [ ] 角色成员管理页面

#### Task 6.2：自定义字段管理页面
- [ ] 自定义字段列表页面
- [ ] 自定义字段创建/编辑页面
- [ ] 字段与 Issue Type 关联配置页面

#### Task 6.3：工作流方案管理页面
- [ ] 工作流方案列表页面
- [ ] 工作流方案创建/编辑页面
- [ ] Issue Type 与 Workflow 映射配置页面

#### Task 6.4：工单页面增强
- [ ] 工单创建页面支持自定义字段
- [ ] 工单详情页面显示自定义字段
- [ ] 工单详情页面显示工作流状态
- [ ] 工单审批操作界面

#### Task 6.5：权限方案管理页面
- [ ] 权限方案列表页面
- [ ] 权限方案创建/编辑页面
- [ ] 权限授予配置页面

#### Task 6.6：测试
- [ ] 功能测试
- [ ] UI/UX 测试
- [ ] 浏览器兼容性测试

### Phase 7：文档与部署（预计 2-3 天）

#### Task 7.1：文档编写
- [ ] API 文档更新（Swagger）
- [ ] 数据库设计文档更新
- [ ] 用户使用手册
- [ ] 管理员配置手册

#### Task 7.2：部署准备
- [ ] 数据库迁移脚本
- [ ] 配置文件更新
- [ ] Docker 镜像构建
- [ ] Kubernetes 部署文件更新

#### Task 7.3：测试与验收
- [ ] 完整功能测试
- [ ] 性能测试
- [ ] 安全测试
- [ ] 用户验收测试

---

## 四、技术难点与风险

### 4.1 技术难点

1. **自定义字段的动态验证**
   - 不同字段类型有不同的验证规则
   - 需要设计灵活的验证框架
   - 解决方案：使用策略模式，为每种字段类型实现验证器

2. **工作流引擎的复杂性**
   - 审批节点的多种类型（单人、会签、或签）
   - 流程流转的条件判断
   - 解决方案：使用状态机模式，清晰定义状态和转换

3. **权限系统的性能**
   - 权限检查频繁，可能影响性能
   - 解决方案：使用 Redis 缓存权限数据，定期刷新

4. **数据迁移**
   - 现有工单需要关联工作流实例
   - 解决方案：编写数据迁移脚本，批量创建工作流实例

### 4.2 风险与应对

| 风险 | 影响 | 概率 | 应对措施 |
|------|------|------|----------|
| 工作流引擎复杂度超预期 | 高 | 中 | 分阶段实现，先实现基础流转，再实现复杂审批 |
| 自定义字段性能问题 | 中 | 低 | 使用 JSON 字段存储，添加索引优化查询 |
| 权限系统设计不合理 | 高 | 低 | 参考 Jira 权限模型，充分评审设计 |
| 前端开发工作量大 | 中 | 中 | 复用现有组件，使用 UI 库加速开发 |
| 数据迁移失败 | 高 | 低 | 充分测试迁移脚本，做好数据备份 |

---

## 五、验收标准

### 5.1 功能验收

- [ ] 项目管理员可以创建和管理项目角色
- [ ] 项目管理员可以为角色添加/移除成员
- [ ] 项目管理员可以创建和管理自定义字段
- [ ] 项目管理员可以配置 Issue Type 的自定义字段
- [ ] 项目管理员可以创建和管理工作流方案
- [ ] 项目管理员可以配置 Issue Type 与 Workflow 的映射
- [ ] 用户创建工单时，自动关联对应的工作流
- [ ] 用户可以提交审批、通过、拒绝、退回工单
- [ ] 审批节点支持单人审批、会签、或签
- [ ] 工作节点支持自动分配给指定角色
- [ ] 项目管理员可以创建和管理权限方案
- [ ] 项目管理员可以为角色授予权限
- [ ] 系统根据权限方案进行权限检查
- [ ] 所有操作有审计日志

### 5.2 性能验收

- [ ] API 响应时间 P95 < 500ms
- [ ] 权限检查响应时间 < 50ms
- [ ] 工作流流转响应时间 < 1s
- [ ] 支持 1000+ 并发用户

### 5.3 安全验收

- [ ] 所有 API 需要认证
- [ ] 敏感操作需要权限检查
- [ ] 审计日志完整
- [ ] 无 SQL 注入、XSS 等漏洞

---

## 六、总结

本计划涵盖了项目角色、自定义字段、工作流方案、工作流实例、审批系统和权限方案的完整实现。预计总开发时间为 **27-35 天**（约 5-7 周）。

**关键里程碑：**
- Week 1：项目角色系统完成
- Week 2：自定义字段系统完成
- Week 3：工作流方案系统完成
- Week 4-5：工作流实例与审批系统完成
- Week 6：权限方案系统完成
- Week 7-8：前端开发与测试

**建议：**
1. 优先实现核心功能，后续迭代优化
2. 每个 Phase 完成后进行 Code Review
3. 及时编写文档和测试用例
4. 定期与用户沟通，确保需求理解正确
