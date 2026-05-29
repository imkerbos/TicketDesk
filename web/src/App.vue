<template>
  <div id="app">
    <!-- 版本更新提示：固定在浏览器顶部 -->
    <div v-if="hasNewVersion" class="update-banner">
      <el-icon><RefreshRight /></el-icon>
      <span v-if="dismissed">系统已更新，</span>
      <span v-else>检测到新版本，{{ countdown }}s 后自动刷新（</span>
      <a class="update-banner-action" @click.stop="reloadPage">立即刷新</a>
      <template v-if="!dismissed">
        <span>或</span>
        <a class="update-banner-action update-banner-dismiss" @click.stop="dismiss">先不刷新</a>
        <span>）</span>
      </template>
    </div>

    <!-- 登录页面：独立全屏显示 -->
    <router-view v-if="isAuthPage" />

    <!-- 主布局：带侧边栏 -->
    <el-container v-else class="main-layout" :class="{ 'has-update-banner': hasNewVersion }">
      <el-aside width="220px" class="sidebar">
        <div class="logo">
          <img v-if="brandStore.logoUrl" :src="brandStore.logoUrl" :alt="brandStore.systemName" class="logo-custom" />
          <span class="logo-text">{{ brandStore.systemName }}</span>
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
            <el-menu-item v-if="userStore.isProjectAdmin" index="/workflows">
              <span>工作流管理</span>
            </el-menu-item>
            <el-menu-item v-if="userStore.isProjectAdmin" index="/fields">
              <span>字段管理</span>
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

          <!-- 需求池（项目管理员及以上可见） -->
          <el-sub-menu v-if="userStore.isProjectAdmin" index="requirement-center">
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
            <el-menu-item index="/requirement-categories">
              <span>分类管理</span>
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
            <el-menu-item v-if="userStore.isProjectAdmin" index="/alert-rules">
              <span>告警规则</span>
            </el-menu-item>
            <el-menu-item v-if="userStore.isProjectAdmin" index="/alert-silences">
              <span>告警静默</span>
            </el-menu-item>
            <el-menu-item v-if="userStore.isAdmin" index="/alert-datasources">
              <span>数据源管理</span>
            </el-menu-item>
          </el-sub-menu>

          <!-- 报表统计 -->
          <el-menu-item index="/reports">
            <el-icon><DataAnalysis /></el-icon>
            <span>报表统计</span>
          </el-menu-item>

          <!-- API 文档 (Swagger UI) -->
          <el-menu-item index="/api-docs">
            <el-icon><Document /></el-icon>
            <span>API 文档</span>
          </el-menu-item>

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
        <div class="sidebar-footer">
          <span class="copyright-text">{{ brandStore.copyrightText }}</span>
        </div>
      </el-aside>
      <el-container class="main-container">
        <el-header class="header">
          <div class="header-content">
            <div v-if="$route.meta.title" class="page-title">{{ $route.meta.title }}</div>
            <div v-else class="page-title-spacer"></div>
            <div class="header-right">
              <el-tooltip
                :content="themeStore.mode === 'light' ? '切换到暗黑模式' : themeStore.mode === 'dark' ? '切换到跟随系统' : '切换到亮色模式'"
                placement="bottom"
              >
                <button class="theme-toggle" @click="themeStore.toggleMode()">
                  <el-icon v-if="themeStore.mode === 'light'" :size="18"><Sunny /></el-icon>
                  <el-icon v-else-if="themeStore.mode === 'dark'" :size="18"><Moon /></el-icon>
                  <el-icon v-else :size="18"><Monitor /></el-icon>
                </button>
              </el-tooltip>
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
import { House, Tickets, Folder, Bell, Message, Setting, ArrowDown, User, SwitchButton, Document, DataAnalysis, Sunny, Moon, Monitor, RefreshRight } from '@element-plus/icons-vue'
import { useUserStore } from '@/stores/user'
import { useNotificationStore } from '@/stores/notification'
import { useThemeStore } from '@/stores/theme'
import { useBrandStore } from '@/stores/brand'
import { useVersionCheck } from '@/composables/useVersionCheck'
import NotificationBell from '@/components/NotificationBell.vue'

const route = useRoute()
const router = useRouter()
const userStore = useUserStore()
const notificationStore = useNotificationStore()
const themeStore = useThemeStore()
const brandStore = useBrandStore()
const { hasNewVersion, countdown, dismissed, reload: reloadPage, dismiss } = useVersionCheck()

// 登录后连接 WebSocket
onMounted(() => {
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
  // 项目子页面（概览、看板、设置、角色）高亮项目列表
  if (path.startsWith('/projects/')) {
    return '/projects'
  }
  // 工作流设计器页面高亮工作流管理
  if (path.startsWith('/workflows/') && path.includes('/designer')) {
    return '/workflows'
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

<style scoped lang="scss">
.main-layout {
  min-height: 100vh;
}

.sidebar {
  position: fixed;
  left: 0;
  top: 0;
  bottom: 0;
  background: var(--td-sidebar-bg);
  border-right: 1px solid var(--td-sidebar-border);
  z-index: var(--td-z-sticky);
  overflow-y: auto;
  display: flex;
  flex-direction: column;
  box-shadow: var(--td-elevation-1);
}

// 自定义滚动条 (sidebar)
.sidebar::-webkit-scrollbar { width: 4px; }
.sidebar::-webkit-scrollbar-thumb { background: rgba(255, 255, 255, 0.08); border-radius: 4px; }
.sidebar::-webkit-scrollbar-thumb:hover { background: rgba(255, 255, 255, 0.16); }

.logo {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: var(--td-space-2);
  padding: var(--td-space-5) var(--td-space-4);
  border-bottom: 1px solid var(--td-sidebar-border);
  flex-shrink: 0;
}

.logo-custom {
  width: 32px;
  height: 32px;
  border-radius: var(--td-radius-md);
  object-fit: contain;
  box-shadow: var(--td-elevation-1);
}

.logo-text {
  font-size: var(--td-font-lg);
  font-weight: var(--td-weight-semibold);
  color: var(--td-text-white);
  letter-spacing: var(--td-tracking-tight);
}

.sidebar-menu {
  border-right: none;
  padding: var(--td-space-3) var(--td-space-2);
  flex: 1;
  overflow-y: auto;
}

.sidebar-menu .el-menu-item {
  margin: 2px 0;
  border-radius: var(--td-radius-md);
  height: 40px;
  line-height: 40px;
  font-size: var(--td-font-md);
  font-weight: var(--td-weight-medium);
  position: relative;
  transition: var(--td-transition-bg), var(--td-transition-color);
}

.sidebar-menu .el-menu-item .menu-badge {
  position: absolute;
  right: var(--td-space-4);
  top: 50%;
  transform: translateY(-50%);
  display: flex;
  align-items: center;
}

.sidebar-menu .el-menu-item .menu-badge :deep(.el-badge__content) {
  position: static;
  transform: none;
  background: var(--td-color-danger);
  border: 2px solid rgba(255, 255, 255, 0.15);
  font-weight: var(--td-weight-semibold);
  font-size: var(--td-font-xs);
  height: 18px;
  line-height: 14px;
  padding: 0 6px;
  min-width: 18px;
  display: inline-flex;
  align-items: center;
  justify-content: center;
}

.sidebar-menu .el-menu-item:hover {
  background-color: var(--td-sidebar-hover) !important;
  color: var(--td-text-white) !important;
}

.sidebar-menu .el-menu-item.is-active {
  background: var(--td-sidebar-active-bg) !important;
  color: var(--td-sidebar-text-active) !important;
  box-shadow: inset 3px 0 0 var(--td-color-primary);
  font-weight: var(--td-weight-semibold);
}

.sidebar-menu :deep(.el-sub-menu__title) {
  margin: 2px 0;
  border-radius: var(--td-radius-md);
  height: 40px;
  line-height: 40px;
  font-size: var(--td-font-md);
  font-weight: var(--td-weight-medium);
  transition: var(--td-transition-bg), var(--td-transition-color);
}

.sidebar-menu :deep(.el-sub-menu__title:hover) {
  background-color: var(--td-sidebar-hover) !important;
  color: var(--td-text-white) !important;
}

.sidebar-menu :deep(.el-sub-menu .el-menu-item) {
  height: 36px;
  line-height: 36px;
  padding-left: 48px !important;
  min-width: auto;
  font-size: var(--td-font-base);
  font-weight: var(--td-weight-regular);
}

.sidebar-footer {
  padding: var(--td-space-3) var(--td-space-4);
  border-top: 1px solid var(--td-sidebar-border);
  margin-top: auto;
  flex-shrink: 0;
}

.copyright-text {
  font-size: var(--td-font-xs);
  color: rgba(255, 255, 255, 0.35);
  line-height: var(--td-leading-normal);
  overflow-wrap: break-word;
}

.main-container {
  margin-left: 220px;
  min-height: 100vh;
  background-color: var(--td-bg-page);
}

.header {
  background-color: var(--td-header-bg);
  box-shadow: var(--td-elevation-1);
  padding: 0 var(--td-space-6);
  height: 60px;
  position: sticky;
  top: 0;
  z-index: var(--td-z-sticky);
  border-bottom: 1px solid var(--td-border-color);
}

.header-content {
  display: flex;
  justify-content: space-between;
  align-items: center;
  height: 100%;
}

.page-title {
  font-size: var(--td-font-xl);
  font-weight: var(--td-weight-semibold);
  color: var(--td-text-primary);
  letter-spacing: var(--td-tracking-tight);
}

.page-title-spacer {
  // 占位, 让 header-right 保持靠右
}

.header-right {
  display: flex;
  align-items: center;
  gap: var(--td-space-2);
}

.theme-toggle {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 36px;
  height: 36px;
  border: none;
  border-radius: var(--td-radius-md);
  background: transparent;
  color: var(--td-text-secondary);
  cursor: pointer;
  transition: var(--td-transition-bg), var(--td-transition-color);
}

.theme-toggle:hover {
  background-color: var(--td-bg-section);
  color: var(--td-color-primary);
}

.user-info {
  display: flex;
  align-items: center;
  gap: var(--td-space-2);
  cursor: pointer;
  padding: var(--td-space-1) var(--td-space-3);
  border-radius: var(--td-radius-md);
  transition: var(--td-transition-bg);
  height: 36px;
}

.user-info:hover {
  background-color: var(--td-bg-section);
}

.user-avatar {
  background: var(--td-color-primary);
  font-size: var(--td-font-md);
  font-weight: var(--td-weight-semibold);
  box-shadow: var(--td-elevation-1);
}

.user-name {
  font-size: var(--td-font-md);
  color: var(--td-text-regular);
  font-weight: var(--td-weight-medium);
  max-width: 120px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.arrow-icon {
  font-size: var(--td-font-sm);
  color: var(--td-text-placeholder);
  transition: var(--td-transition-color);
}

.user-info:hover .arrow-icon {
  color: var(--td-text-primary);
}

.update-banner {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  height: 40px;
  z-index: var(--td-z-toast);
  display: flex;
  align-items: center;
  justify-content: center;
  gap: var(--td-space-1);
  background-color: var(--td-color-primary);
  color: #fff;
  font-size: var(--td-font-md);
  font-weight: var(--td-weight-medium);
  transition: var(--td-transition-bg);
  box-shadow: var(--td-elevation-2);
}

.update-banner-action {
  color: #fff;
  text-decoration: underline;
  cursor: pointer;
  font-weight: var(--td-weight-bold);
  margin: 0 4px;
}

.update-banner-action:hover {
  opacity: 0.85;
}

.update-banner-dismiss {
  opacity: 0.85;
  font-weight: var(--td-weight-medium);
}

.has-update-banner .sidebar {
  top: 40px;
}

.has-update-banner .main-container {
  padding-top: 40px;
}

.main-content {
  padding: var(--td-space-6);
  min-height: calc(100vh - 60px);
}

@media (prefers-reduced-motion: reduce) {
  .sidebar-menu .el-menu-item,
  .sidebar-menu :deep(.el-sub-menu__title),
  .theme-toggle,
  .user-info,
  .arrow-icon,
  .update-banner {
    transition: none !important;
  }
}

@media (max-width: 768px) {
  .main-container {
    margin-left: 0;
  }

  .sidebar {
    transform: translateX(-100%);
    transition: transform 200ms var(--td-ease-in-out);
  }

  .header {
    padding: 0 var(--td-space-3);
  }

  .user-name {
    display: none;
  }
}
</style>
