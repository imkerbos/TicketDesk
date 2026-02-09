<template>
  <div id="app">
    <!-- 登录页面：独立全屏显示 -->
    <router-view v-if="isAuthPage" />

    <!-- 主布局：带侧边栏 -->
    <el-container v-else class="main-layout">
      <el-aside width="220px" class="sidebar">
        <div class="logo">
          <span class="logo-icon">T</span>
          <span class="logo-text">TicketDesk</span>
        </div>
        <el-menu
          :default-active="activeMenu"
          :default-openeds="['project-center', 'alert-center', 'system']"
          background-color="#1e1e2d"
          text-color="#9899ac"
          active-text-color="#fff"
          class="sidebar-menu"
          @select="handleMenuSelect"
        >
          <!-- 首页 -->
          <el-menu-item index="/dashboard">
            <el-icon><House /></el-icon>
            <span>首页</span>
          </el-menu-item>

          <!-- 工单管理 -->
          <el-menu-item index="/issues">
            <el-icon><Tickets /></el-icon>
            <span>工单管理</span>
          </el-menu-item>

          <!-- 项目管理 -->
          <el-sub-menu index="project-center">
            <template #title>
              <el-icon><Folder /></el-icon>
              <span>项目管理</span>
            </template>
            <el-menu-item index="/projects">
              <span>项目列表</span>
            </el-menu-item>
            <el-menu-item index="/workflows">
              <span>工作流管理</span>
            </el-menu-item>
          </el-sub-menu>

          <!-- 通知中心 -->
          <el-menu-item index="/notifications">
            <el-icon><Message /></el-icon>
            <span>通知中心</span>
            <el-badge
              v-if="notificationStore.unreadCount > 0"
              :value="notificationStore.unreadCount"
              :max="99"
              class="menu-badge"
            />
          </el-menu-item>

          <!-- 需求池 -->
          <el-sub-menu index="requirement-center">
            <template #title>
              <el-icon><Document /></el-icon>
              <span>需求池</span>
            </template>
            <el-menu-item index="/requirement-pools">
              <span>需求池管理</span>
            </el-menu-item>
            <el-menu-item index="/requirements">
              <span>需求管理</span>
            </el-menu-item>
            <el-menu-item index="/requirements/kanban">
              <span>需求看板</span>
            </el-menu-item>
          </el-sub-menu>

          <!-- 告警中心 -->
          <el-sub-menu index="alert-center">
            <template #title>
              <el-icon><Bell /></el-icon>
              <span>告警中心</span>
            </template>
            <el-menu-item index="/alerts">
              <span>告警列表</span>
            </el-menu-item>
            <el-menu-item index="/alert-rules">
              <span>告警规则</span>
            </el-menu-item>
            <el-menu-item index="/alert-silences">
              <span>告警静默</span>
            </el-menu-item>
            <el-menu-item v-if="userStore.isAdmin" index="/alert-datasources">
              <span>数据源管理</span>
            </el-menu-item>
          </el-sub-menu>

          <!-- 系统管理（仅管理员可见） -->
          <el-sub-menu v-if="userStore.isAdmin" index="system">
            <template #title>
              <el-icon><Setting /></el-icon>
              <span>系统管理</span>
            </template>
            <el-menu-item index="/users">
              <span>用户管理</span>
            </el-menu-item>
            <el-menu-item index="/settings">
              <span>系统设置</span>
            </el-menu-item>
          </el-sub-menu>
        </el-menu>
      </el-aside>
      <el-container class="main-container">
        <el-header class="header">
          <div class="header-content">
            <div class="page-title">{{ $route.meta.title }}</div>
            <div class="header-right">
              <NotificationBell />
              <el-dropdown trigger="click">
                <div class="user-info">
                  <el-avatar :size="32" class="user-avatar">{{ userStore.displayName?.charAt(0) || 'U' }}</el-avatar>
                  <span class="user-name">{{ userStore.displayName || '用户' }}</span>
                  <el-icon class="arrow-icon"><ArrowDown /></el-icon>
                </div>
                <template #dropdown>
                  <el-dropdown-menu>
                    <el-dropdown-item @click="router.push('/profile')">
                      <el-icon><User /></el-icon>
                      个人设置
                    </el-dropdown-item>
                    <el-dropdown-item divided @click="handleLogout">
                      <el-icon><SwitchButton /></el-icon>
                      退出登录
                    </el-dropdown-item>
                  </el-dropdown-menu>
                </template>
              </el-dropdown>
            </div>
          </div>
        </el-header>
        <el-main class="main-content">
          <router-view />
        </el-main>
      </el-container>
    </el-container>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, onUnmounted, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { House, Tickets, Folder, Bell, Message, Setting, ArrowDown, User, SwitchButton, Document } from '@element-plus/icons-vue'
import { useUserStore } from '@/stores/user'
import { useNotificationStore } from '@/stores/notification'
import NotificationBell from '@/components/NotificationBell.vue'

const route = useRoute()
const router = useRouter()
const userStore = useUserStore()
const notificationStore = useNotificationStore()

// 初始化用户状态
onMounted(() => {
  userStore.initFromStorage()
  // 登录后连接 WebSocket
  if (userStore.isLoggedIn) {
    notificationStore.connectWebSocket()
  }
})

// 监听登录状态变化
watch(() => userStore.isLoggedIn, (loggedIn) => {
  if (loggedIn) {
    notificationStore.connectWebSocket()
  } else {
    notificationStore.disconnectWebSocket()
  }
})

onUnmounted(() => {
  notificationStore.disconnectWebSocket()
})

// 判断是否为认证页面（登录/注册等）
const isAuthPage = computed(() => {
  return route.path === '/login' || route.path === '/register' || route.path === '/forgot-password' || route.path === '/reset-password'
})

// 计算当前激活的菜单
const activeMenu = computed(() => {
  const path = route.path
  // 工单详情页面高亮工单列表
  if (path.startsWith('/issues/')) {
    return '/issues'
  }
  // 告警详情页面高亮告警列表
  if (path.startsWith('/alerts/') && path !== '/alerts') {
    return '/alerts'
  }
  // 项目角色页面高亮项目列表
  if (path.startsWith('/projects/') && path.includes('/roles')) {
    return '/projects'
  }
  return path
})

const handleLogout = () => {
  notificationStore.disconnectWebSocket()
  userStore.logout()
  router.push('/login')
}

// 处理菜单点击，支持 Ctrl/Cmd+Click 打开新标签页
const handleMenuSelect = (index: string) => {
  // 检测是否按住 Ctrl (Windows) 或 Cmd (Mac)
  const lastClickEvent = window.event as MouseEvent | undefined
  if (lastClickEvent && (lastClickEvent.ctrlKey || lastClickEvent.metaKey)) {
    // 在新标签页打开
    window.open(index, '_blank')
  } else {
    // 在当前标签页跳转
    router.push(index)
  }
}
</script>

<style scoped>
.main-layout {
  min-height: 100vh;
}

.sidebar {
  position: fixed;
  left: 0;
  top: 0;
  bottom: 0;
  background: linear-gradient(180deg, #1e1e2d 0%, #1a1a27 100%);
  border-right: 1px solid rgba(255, 255, 255, 0.05);
  z-index: 100;
  overflow-y: auto;
}

.logo {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 10px;
  padding: 24px 20px;
  border-bottom: 1px solid rgba(255, 255, 255, 0.05);
}

.logo-icon {
  width: 36px;
  height: 36px;
  background: linear-gradient(135deg, #3b82f6 0%, #8b5cf6 100%);
  border-radius: 8px;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 20px;
  font-weight: 700;
  color: #fff;
}

.logo-text {
  font-size: 18px;
  font-weight: 600;
  color: #fff;
}

.sidebar-menu {
  border-right: none;
  padding: 12px 8px;
}

.sidebar-menu .el-menu-item {
  margin: 4px 0;
  border-radius: 8px;
  height: 44px;
  line-height: 44px;
  position: relative;
}

.sidebar-menu .el-menu-item .menu-badge {
  position: absolute;
  right: 16px;
  top: 50%;
  transform: translateY(-50%);
  display: flex;
  align-items: center;
}

.sidebar-menu .el-menu-item .menu-badge :deep(.el-badge__content) {
  position: static;
  transform: none;
  background: linear-gradient(135deg, #ef4444 0%, #dc2626 100%);
  border: 2px solid rgba(255, 255, 255, 0.2);
  font-weight: 600;
  font-size: 11px;
  height: 18px;
  line-height: 14px;
  padding: 0 6px;
  min-width: 18px;
  display: inline-flex;
  align-items: center;
  justify-content: center;
}

.sidebar-menu .el-menu-item:hover {
  background-color: rgba(255, 255, 255, 0.05) !important;
}

.sidebar-menu .el-menu-item.is-active {
  background: linear-gradient(90deg, rgba(59, 130, 246, 0.2) 0%, rgba(139, 92, 246, 0.2) 100%) !important;
  color: #fff;
}

.sidebar-menu :deep(.el-sub-menu__title) {
  margin: 4px 0;
  border-radius: 8px;
  height: 44px;
  line-height: 44px;
}

.sidebar-menu :deep(.el-sub-menu__title:hover) {
  background-color: rgba(255, 255, 255, 0.05) !important;
}

.sidebar-menu :deep(.el-sub-menu .el-menu-item) {
  height: 40px;
  line-height: 40px;
  padding-left: 52px !important;
  min-width: auto;
}

.main-container {
  margin-left: 220px;
  min-height: 100vh;
  background-color: #f5f7fa;
}

.header {
  background-color: #fff;
  box-shadow: 0 1px 4px rgba(0, 21, 41, 0.06);
  padding: 0 24px;
  height: 64px;
}

.header-content {
  display: flex;
  justify-content: space-between;
  align-items: center;
  height: 100%;
}

.page-title {
  font-size: 18px;
  font-weight: 600;
  color: #1f2937;
}

.header-right {
  display: flex;
  align-items: center;
  gap: 16px;
}

.user-info {
  display: flex;
  align-items: center;
  gap: 8px;
  cursor: pointer;
  padding: 8px 12px;
  border-radius: 8px;
  transition: background-color 0.2s;
}

.user-info:hover {
  background-color: #f5f7fa;
}

.user-avatar {
  background: linear-gradient(135deg, #3b82f6 0%, #8b5cf6 100%);
  font-size: 14px;
}

.user-name {
  font-size: 14px;
  color: #374151;
  font-weight: 500;
}

.arrow-icon {
  font-size: 12px;
  color: #9ca3af;
}

.main-content {
  padding: 24px;
  min-height: calc(100vh - 64px);
}
</style>
