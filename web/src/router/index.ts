import { createRouter, createWebHistory } from 'vue-router'

const router = createRouter({
  history: createWebHistory(import.meta.env.BASE_URL),
  routes: [
    {
      path: '/login',
      name: 'Login',
      component: () => import('@/views/auth/Login.vue'),
      meta: { title: '登录' },
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
      meta: { title: '告警规则' },
    },
    {
      path: '/alert-silences',
      name: 'AlertSilences',
      component: () => import('@/views/alert/AlertSilences.vue'),
      meta: { title: '告警静默' },
    },
    // 系统管理
    {
      path: '/users',
      name: 'UserList',
      component: () => import('@/views/user/UserList.vue'),
      meta: { title: '用户管理' },
    },
    // 个人设置
    {
      path: '/profile',
      name: 'Profile',
      component: () => import('@/views/user/Profile.vue'),
      meta: { title: '个人设置' },
    },
    // 系统设置
    {
      path: '/settings',
      name: 'SystemSettings',
      component: () => import('@/views/system/SystemSettings.vue'),
      meta: { title: '系统设置' },
    },
  ],
})

router.beforeEach((to, _from, next) => {
  document.title = `${to.meta.title || 'TicketDesk'} - TicketDesk`
  next()
})

export default router
