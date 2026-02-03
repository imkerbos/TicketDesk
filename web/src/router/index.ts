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
      redirect: '/alerts',
    },
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
  ],
})

router.beforeEach((to, from, next) => {
  document.title = `${to.meta.title || 'TicketDesk'} - TicketDesk`
  next()
})

export default router
