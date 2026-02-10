# 可视化工作流编辑器 - 实现方案

## 需求概述

将当前基于列表/对话框的工作流管理界面，升级为类似 Jira 的**图形化、拖拽式工作流编辑器**，支持：
1. **可视化查看**：以流程图形式展示工作流的状态节点和流转关系
2. **拖拽式编辑**：拖动节点调整位置，拖拽连线定义流转路径
3. **节点配置**：点击节点弹出配置面板（审批人、指派人、目标状态等）
4. **实时预览**：编辑即所见

## 技术选型

### 核心库：Vue Flow（@vue-flow/core）
- **为什么选 Vue Flow**：
  - Vue 3 原生支持，与项目技术栈完美匹配
  - 基于 React Flow 的 Vue 移植版，社区活跃、文档完善
  - 内置拖拽、缩放、连线、minimap 等功能
  - 支持自定义节点和边的渲染
  - 轻量级，bundle size 小
  - MIT 开源协议

- **需要安装的包**：
  - `@vue-flow/core` - 核心库
  - `@vue-flow/background` - 背景网格
  - `@vue-flow/controls` - 缩放控制按钮
  - `@vue-flow/minimap` - 小地图导航

## 现有基础

### 后端（无需修改）
后端已经完整支持：
- ✅ `WorkflowNode` 模型有 `position_x`, `position_y` 字段
- ✅ `CreateNodeRequest` / `UpdateNodeRequest` 支持位置参数
- ✅ `WorkflowEdge` 模型有 `source_node_id`, `target_node_id`, `condition_expr`
- ✅ 完整的 CRUD API（节点、边、工作流）
- ✅ `NodeConfig` 支持 `target_status` 字段

### 前端
- ✅ `web/src/api/workflow.ts` - 完整的 API 调用
- ✅ `web/src/types/workflow.ts` - 完整的类型定义
- ⚠️ `web/src/views/workflow/WorkflowList.vue` - 当前是列表式 UI，需要重构

## 实现方案

### 页面结构设计

```
/workflows                → WorkflowList.vue（工作流列表页，保留）
/workflows/:id/designer  → WorkflowDesigner.vue（新增：可视化编辑器页面）
```

### 新增文件清单

```
web/src/views/workflow/
├── WorkflowList.vue              # 改造：列表页增加"设计"按钮，跳转到编辑器
├── WorkflowDesigner.vue          # 新增：可视化编辑器主页面
└── components/
    ├── nodes/
    │   ├── StartNode.vue         # 自定义开始节点
    │   ├── EndNode.vue           # 自定义结束节点
    │   ├── ApprovalNode.vue      # 自定义审批节点
    │   ├── WorkNode.vue          # 自定义工作节点
    │   └── SystemNode.vue        # 自定义系统节点
    ├── NodeConfigPanel.vue       # 右侧节点配置面板
    ├── EdgeConfigPanel.vue       # 边配置面板（条件表达式）
    └── NodeToolbar.vue           # 左侧节点工具栏（拖拽添加）
```

### 核心交互设计

#### 1. 整体布局（WorkflowDesigner.vue）
```
┌─────────────────────────────────────────────────────────┐
│  顶部工具栏：工作流名称 | 保存 | 返回列表 | 撤销/重做   │
├────────┬──────────────────────────────┬─────────────────┤
│ 节点   │                              │  配置面板       │
│ 工具栏 │     Vue Flow 画布区域         │  (选中节点时    │
│        │     (拖拽/缩放/连线)          │   显示配置)     │
│ ·开始  │                              │                 │
│ ·结束  │    ┌──────┐    ┌──────┐      │  节点名称: xxx  │
│ ·审批  │    │ 开始  │───→│ 审批  │     │  审批类型: 单人 │
│ ·工作  │    └──────┘    └──┬───┘      │  审批人: ...    │
│ ·系统  │                   │          │  目标状态: ...  │
│        │              ┌────▼───┐      │  超时: ...      │
│        │              │  结束   │      │                 │
│        │              └────────┘      │                 │
├────────┴──────────────────────────────┴─────────────────┤
│  底部：MiniMap | 缩放控制                                │
└─────────────────────────────────────────────────────────┘
```

#### 2. 节点样式设计
每种节点类型有独特的视觉样式：
- **开始节点**：圆形，绿色渐变，播放图标
- **结束节点**：圆形，灰色渐变，完成图标
- **审批节点**：圆角矩形，橙色/黄色渐变，审批图标，显示审批类型标签
- **工作节点**：圆角矩形，蓝色渐变，工作图标，显示指派信息
- **系统节点**：圆角矩形，紫色渐变，齿轮图标，显示动作名称

#### 3. 交互流程
- **添加节点**：从左侧工具栏拖拽节点到画布
- **连接节点**：从节点的连接点（handle）拖拽到另一个节点
- **配置节点**：点击节点，右侧面板显示配置表单
- **配置边**：点击边，弹出条件表达式配置
- **删除**：选中节点/边后按 Delete 键，或右键菜单删除
- **移动节点**：直接拖拽节点到新位置
- **保存**：点击保存按钮，批量同步所有变更到后端

#### 4. 数据同步策略
采用**本地编辑 + 手动保存**模式：
- 加载时：从后端获取 nodes 和 edges，转换为 Vue Flow 格式
- 编辑时：所有操作在本地进行（增删改节点/边/位置）
- 保存时：对比本地与服务端数据，计算差异，批量调用 API：
  - 新增的节点 → `createNode()`
  - 修改的节点 → `updateNode()`（位置、配置、名称）
  - 删除的节点 → `deleteNode()`
  - 新增的边 → `createEdge()`
  - 删除的边 → `deleteEdge()`

### 类型扩展

在 `web/src/types/workflow.ts` 中添加：
```typescript
// 节点配置扩展 target_status
export interface NodeConfig {
  // ... 现有字段
  target_status?: string  // 进入该节点时工单目标状态
}
```

### 路由变更

在 `web/src/router/index.ts` 中添加：
```typescript
{
  path: '/workflows/:id/designer',
  name: 'WorkflowDesigner',
  component: () => import('@/views/workflow/WorkflowDesigner.vue'),
  meta: { title: '工作流设计器' }
}
```

### WorkflowList.vue 改造
- 保留现有列表页面
- 每个工作流卡片增加"设计工作流"按钮（主要操作）
- 点击后跳转到 `/workflows/:id/designer`
- 原有的"查看节点"按钮可保留作为备选，或替换为"设计"

## 实现步骤

### Step 1: 安装依赖
```bash
cd web && npm install @vue-flow/core @vue-flow/background @vue-flow/controls @vue-flow/minimap
```

### Step 2: 创建自定义节点组件（5个）
每个节点组件接收 Vue Flow 的 node props，渲染对应样式

### Step 3: 创建 NodeToolbar（左侧拖拽面板）
可拖拽的节点类型列表，支持 drag & drop 到画布

### Step 4: 创建 NodeConfigPanel（右侧配置面板）
根据选中节点类型动态显示配置表单（复用现有 WorkflowList.vue 中的表单逻辑）

### Step 5: 创建 WorkflowDesigner.vue（主页面）
整合 Vue Flow 画布 + 工具栏 + 配置面板 + 保存逻辑

### Step 6: 更新路由和 WorkflowList.vue
添加路由，列表页增加"设计"入口

### Step 7: 更新类型定义
NodeConfig 添加 target_status 字段

## 不需要修改的部分

- ❌ 后端代码 - 现有 API 完全满足需求
- ❌ 数据库模型 - position_x/position_y 已存在
- ❌ API 调用层 - `web/src/api/workflow.ts` 已完整

## 风险与注意事项

1. **Vue Flow 版本兼容性**：确保与 Vue 3.4 兼容
2. **大量节点性能**：Vue Flow 内置虚拟化，一般工作流节点数不多，无性能问题
3. **并发编辑**：当前不考虑多人同时编辑同一工作流的冲突问题
4. **浏览器兼容性**：Vue Flow 需要现代浏览器（Chrome/Firefox/Edge）
