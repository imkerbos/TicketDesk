# TicketDesk 前端项目

基于 Vue 3 + TypeScript + Element Plus 的现代化告警管理系统前端。

## 技术栈

- **框架**: Vue 3.4 + TypeScript
- **UI 组件库**: Element Plus 2.5
- **路由**: Vue Router 4
- **状态管理**: Pinia 2
- **HTTP 客户端**: Axios
- **构建工具**: Vite 5
- **图表库**: ECharts 5
- **日期处理**: Day.js

## 功能特性

### 1. 告警列表页面
- ✅ 实时告警统计卡片（总数、严重、警告、触发中）
- ✅ 多维度过滤（状态、严重程度、告警名称）
- ✅ 列表视图和分组视图切换
- ✅ 告警确认和解决操作
- ✅ 关联工单跳转
- ✅ 分页查询

### 2. 告警详情页面
- ✅ 完整的告警信息展示
- ✅ 标签和注解详细信息
- ✅ 关联工单信息
- ✅ 确认和解决操作

### 3. 告警静默管理
- ✅ 静默规则列表
- ✅ 创建/编辑静默规则
- ✅ 标签匹配器配置
- ✅ 时间范围设置
- ✅ 取消和删除静默

### 4. 告警规则配置
- ✅ 规则列表展示
- ✅ 创建/编辑规则
- ✅ 标签匹配器配置
- ✅ 合并窗口设置
- ✅ 自动解决开关
- ✅ 优先级和指派人配置

## 项目结构

```
web/
├── src/
│   ├── api/              # API 接口
│   │   └── alert.ts      # 告警相关 API
│   ├── components/       # 公共组件
│   ├── router/           # 路由配置
│   │   └── index.ts
│   ├── styles/           # 全局样式
│   │   └── index.scss
│   ├── types/            # TypeScript 类型定义
│   │   └── alert.ts      # 告警类型
│   ├── utils/            # 工具函数
│   │   └── request.ts    # Axios 封装
│   ├── views/            # 页面组件
│   │   └── alert/        # 告警相关页面
│   │       ├── AlertList.vue      # 告警列表
│   │       ├── AlertDetail.vue    # 告警详情
│   │       ├── AlertSilences.vue  # 告警静默
│   │       └── AlertRules.vue     # 告警规则
│   ├── App.vue           # 根组件
│   └── main.ts           # 入口文件
├── index.html            # HTML 模板
├── package.json          # 依赖配置
├── tsconfig.json         # TypeScript 配置
├── vite.config.ts        # Vite 配置
└── README.md             # 项目说明
```

## 快速开始

### 1. 安装依赖

```bash
cd web
npm install
```

### 2. 启动开发服务器

```bash
npm run dev
```

访问 http://localhost:3000

### 3. 构建生产版本

```bash
npm run build
```

构建产物将生成在 `dist/` 目录。

## 开发说明

### API 代理配置

开发环境下，API 请求会自动代理到后端服务（默认 http://localhost:10010）。

配置位于 `vite.config.ts`:

```typescript
server: {
  port: 3000,
  proxy: {
    '/api': {
      target: 'http://localhost:10010',
      changeOrigin: true,
    },
  },
}
```

### 自动导入

项目使用 `unplugin-auto-import` 和 `unplugin-vue-components` 实现自动导入：

- Vue 3 API（ref, reactive, computed 等）
- Vue Router API（useRouter, useRoute 等）
- Pinia API（defineStore 等）
- Element Plus 组件

无需手动 import，直接使用即可。

### 样式规范

- 使用 SCSS 预处理器
- 全局样式定义在 `src/styles/index.scss`
- 组件样式使用 `scoped` 避免污染
- 遵循 BEM 命名规范

### 代码规范

- 使用 TypeScript 严格模式
- 遵循 Vue 3 Composition API 风格
- 使用 `<script setup>` 语法糖
- 组件命名使用 PascalCase
- 文件命名使用 PascalCase

## 设计特点

### 1. 现代化 UI 设计
- 渐变色统计卡片
- 悬停动画效果
- 响应式布局
- 暗色模式支持

### 2. 专业的运维监控风格
- 清晰的严重程度标识（critical/warning/info）
- 直观的状态标签（firing/resolved）
- 丰富的数据可视化
- 快速操作按钮

### 3. 优秀的用户体验
- 快速过滤和搜索
- 列表/分组视图切换
- 一键确认和解决
- 实时数据更新

### 4. 完善的错误处理
- 统一的 API 错误拦截
- 友好的错误提示
- 自动登录过期处理

## 浏览器支持

- Chrome >= 90
- Firefox >= 88
- Safari >= 14
- Edge >= 90

## 后续优化建议

1. **性能优化**
   - 虚拟滚动（大数据量列表）
   - 组件懒加载
   - 图片懒加载

2. **功能增强**
   - 告警趋势图表
   - 实时推送（WebSocket）
   - 批量操作
   - 导出功能

3. **用户体验**
   - 快捷键支持
   - 自定义列显示
   - 保存过滤条件
   - 主题切换

4. **国际化**
   - 多语言支持
   - 时区处理

## License

MIT
