import { createRouter, createWebHistory } from 'vue-router'
import { useUserStore } from '@/stores/user'

const router = createRouter({
  history: createWebHistory(import.meta.env.BASE_URL),
  routes: [
    {
      path: '/login',
      name: 'Login',
      component: () => import('@/views/auth/Login.vue'),
      meta: { title: '登录', public: true },
    },
    {
      path: '/forgot-password',
      name: 'ForgotPassword',
      component: () => import('@/views/auth/ForgotPassword.vue'),
      meta: { title: '忘记密码', public: true },
    },
    {
      path: '/reset-password',
      name: 'ResetPassword',
      component: () => import('@/views/auth/ResetPassword.vue'),
      meta: { title: '重置密码', public: true },
    },
    {
      path: '/auth/sso/callback',
      name: 'SSOCallback',
      component: () => import('@/views/auth/SSOCallback.vue'),
      meta: { title: 'SSO 登录', public: true },
    },
    {
      path: '/auth/sso/login',
      name: 'SSOLogin',
      component: () => import('@/views/auth/SSOLogin.vue'),
      meta: { title: 'SSO 登录', public: true },
    },
    {
      path: '/',
      redirect: '/dashboard',
    },
    // 首页/仪表盘
    {
      path: '/dashboard',
      name: 'Dashboard',
      component: () => import('@/views/dashboard/Dashboard.vue'),
      meta: { title: '首页' },
    },
    // 工单管理
    {
      path: '/issues',
      name: 'IssueList',
      component: () => import('@/views/issue/IssueList.vue'),
      meta: { title: '工单管理' },
    },
    {
      path: '/issues/:key',
      name: 'IssueDetail',
      component: () => import('@/views/issue/IssueDetail.vue'),
      meta: { title: '工单详情' },
    },
    // 项目管理
    {
      path: '/projects',
      name: 'ProjectList',
      component: () => import('@/views/project/ProjectList.vue'),
      meta: { title: '项目管理' },
    },
    {
      path: '/projects/:key/settings',
      name: 'ProjectSettings',
      component: () => import('@/views/project/ProjectSettings.vue'),
      meta: { title: '项目设置' },
    },
    {
      path: '/projects/:key/roles',
      name: 'ProjectRoles',
      component: () => import('@/views/project/ProjectRoles.vue'),
      meta: { title: '项目角色' },
    },
    // 工作流管理（需要项目管理员权限）
    {
      path: '/workflows',
      name: 'WorkflowList',
      component: () => import('@/views/workflow/WorkflowList.vue'),
      meta: { title: '工作流管理', requiresProjectAdmin: true },
    },
    {
      path: '/workflows/:id/designer',
      name: 'WorkflowDesigner',
      component: () => import('@/views/workflow/WorkflowDesigner.vue'),
      meta: { title: '工作流设计器', requiresProjectAdmin: true },
    },
    // 字段管理（需要项目管理员权限）
    {
      path: '/fields',
      name: 'FieldManagement',
      component: () => import('@/views/field/FieldManagement.vue'),
      meta: { title: '字段管理', requiresProjectAdmin: true },
    },
    // 告警中心
    {
      path: '/alerts',
      name: 'AlertList',
      component: () => import('@/views/alert/AlertList.vue'),
      meta: { title: '告警列表' },
    },
    {
      path: '/alerts/:id',
      name: 'AlertDetail',
      component: () => import('@/views/alert/AlertDetail.vue'),
      meta: { title: '告警详情' },
    },
    {
      path: '/alert-rules',
      name: 'AlertRules',
      component: () => import('@/views/alert/AlertRules.vue'),
      meta: { title: '告警规则', requiresProjectAdmin: true },
    },
    {
      path: '/alert-silences',
      name: 'AlertSilences',
      component: () => import('@/views/alert/AlertSilences.vue'),
      meta: { title: '告警静默', requiresProjectAdmin: true },
    },
    {
      path: '/alert-datasources',
      name: 'AlertDatasources',
      component: () => import('@/views/alert/AlertDatasources.vue'),
      meta: { title: '数据源管理', requiresAdmin: true },
    },
    // 通知中心
    {
      path: '/notifications',
      name: 'NotificationList',
      component: () => import('@/views/notification/NotificationList.vue'),
      meta: { title: '通知中心' },
    },
    // 需求池管理（需要项目管理员权限）
    {
      path: '/requirement-pools',
      name: 'RequirementPoolList',
      component: () => import('@/views/requirement/RequirementPoolList.vue'),
      meta: { title: '需求池管理', requiresProjectAdmin: true },
    },
    {
      path: '/requirements',
      name: 'RequirementList',
      component: () => import('@/views/requirement/RequirementList.vue'),
      meta: { title: '需求管理', requiresProjectAdmin: true },
    },
    {
      path: '/requirements/kanban',
      name: 'RequirementKanban',
      component: () => import('@/views/requirement/RequirementKanban.vue'),
      meta: { title: '需求看板', requiresProjectAdmin: true },
    },
    // 报表统计
    {
      path: '/reports',
      name: 'Reports',
      component: () => import('@/views/report/Reports.vue'),
      meta: { title: '报表统计' },
    },
    // 系统管理（需要管理员权限）
    {
      path: '/users',
      name: 'UserList',
      component: () => import('@/views/user/UserList.vue'),
      meta: { title: '用户管理', requiresAdmin: true },
    },
    // 个人设置
    {
      path: '/profile',
      name: 'Profile',
      component: () => import('@/views/user/Profile.vue'),
      meta: { title: '个人设置' },
    },
    // 系统设置（需要管理员权限）
    {
      path: '/settings',
      name: 'SystemSettings',
      component: () => import('@/views/system/SystemSettings.vue'),
      meta: { title: '系统设置', requiresAdmin: true },
    },
    // 错误页面
    {
      path: '/403',
      name: 'Forbidden',
      component: () => import('@/views/error/Forbidden.vue'),
      meta: { title: '无权访问', public: true },
    },
    {
      path: '/500',
      name: 'ServerError',
      component: () => import('@/views/error/ServerError.vue'),
      meta: { title: '服务器错误', public: true },
    },
    // 404 兜底（必须放在最后）
    {
      path: '/:pathMatch(.*)*',
      name: 'NotFound',
      component: () => import('@/views/error/NotFound.vue'),
      meta: { title: '页面不存在', public: true },
    },
  ],
})

router.beforeEach((to, _from, next) => {
  document.title = `${to.meta.title || 'TicketDesk'} - TicketDesk`

  const userStore = useUserStore()

  // 公开页面（如登录页）直接放行
  if (to.meta.public) {
    next()
    return
  }

  // 检查是否已登录
  if (!userStore.isLoggedIn) {
    next('/login')
    return
  }

  // 检查是否需要管理员权限
  if (to.meta.requiresAdmin && !userStore.isAdmin) {
    next('/403')
    return
  }

  // 检查是否需要项目管理员权限
  if (to.meta.requiresProjectAdmin && !userStore.isProjectAdmin) {
    next('/403')
    return
  }

  next()
})

export default router
