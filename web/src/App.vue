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
          :default-active="$route.path"
          router
          background-color="#1e1e2d"
          text-color="#9899ac"
          active-text-color="#fff"
          class="sidebar-menu"
        >
          <el-menu-item index="/alerts">
            <el-icon><Bell /></el-icon>
            <span>告警列表</span>
          </el-menu-item>
          <el-menu-item index="/alert-rules">
            <el-icon><Setting /></el-icon>
            <span>告警规则</span>
          </el-menu-item>
          <el-menu-item index="/alert-silences">
            <el-icon><MuteNotification /></el-icon>
            <span>告警静默</span>
          </el-menu-item>
        </el-menu>
      </el-aside>
      <el-container class="main-container">
        <el-header class="header">
          <div class="header-content">
            <div class="page-title">{{ $route.meta.title }}</div>
            <div class="header-right">
              <el-dropdown trigger="click">
                <div class="user-info">
                  <el-avatar :size="32" class="user-avatar">管</el-avatar>
                  <span class="user-name">管理员</span>
                  <el-icon class="arrow-icon"><ArrowDown /></el-icon>
                </div>
                <template #dropdown>
                  <el-dropdown-menu>
                    <el-dropdown-item>个人设置</el-dropdown-item>
                    <el-dropdown-item divided @click="handleLogout">退出登录</el-dropdown-item>
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
import { computed } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { Bell, Setting, MuteNotification, ArrowDown } from '@element-plus/icons-vue'

const route = useRoute()
const router = useRouter()

// 判断是否为认证页面（登录/注册等）
const isAuthPage = computed(() => {
  return route.path === '/login' || route.path === '/register'
})

const handleLogout = () => {
  localStorage.removeItem('token')
  router.push('/login')
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
}

.sidebar-menu .el-menu-item:hover {
  background-color: rgba(255, 255, 255, 0.05) !important;
}

.sidebar-menu .el-menu-item.is-active {
  background: linear-gradient(90deg, rgba(59, 130, 246, 0.2) 0%, rgba(139, 92, 246, 0.2) 100%) !important;
  color: #fff;
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
