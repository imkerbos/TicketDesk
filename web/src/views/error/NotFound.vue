<template>
  <div class="error-page">
    <div class="error-content">
      <div class="error-code">404</div>
      <div class="error-illustration">
        <svg viewBox="0 0 200 120" fill="none" xmlns="http://www.w3.org/2000/svg">
          <circle cx="100" cy="60" r="50" fill="#f0f5ff" stroke="#d6e4ff" stroke-width="2"/>
          <path d="M80 50 L90 60 L80 70" stroke="#4f6ef7" stroke-width="3" stroke-linecap="round" stroke-linejoin="round"/>
          <path d="M120 50 L110 60 L120 70" stroke="#4f6ef7" stroke-width="3" stroke-linecap="round" stroke-linejoin="round"/>
          <path d="M88 82 Q100 75 112 82" stroke="#4f6ef7" stroke-width="2.5" stroke-linecap="round" fill="none"/>
        </svg>
      </div>
      <h2 class="error-title">页面不存在</h2>
      <p class="error-desc">你访问的页面可能已被移除或地址有误</p>
      <div class="error-actions">
        <el-button type="primary" @click="router.push('/dashboard')">返回首页</el-button>
        <el-button @click="router.back()">返回上页</el-button>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { useRouter } from 'vue-router'
const router = useRouter()
</script>

<style scoped lang="scss">
.error-page {
  min-height: 100vh;
  display: flex;
  align-items: center;
  justify-content: center;
  background: #f5f7fa;
}

.error-content {
  text-align: center;
  padding: 40px;
}

.error-code {
  font-size: 120px;
  font-weight: 800;
  color: #3b82f6;
  line-height: 1;
  margin-bottom: 8px;
}

.error-illustration {
  width: 200px;
  margin: 0 auto 24px;
}

.error-title {
  font-size: 22px;
  font-weight: 600;
  color: #1f2937;
  margin: 0 0 8px;
}

.error-desc {
  font-size: 15px;
  color: #6b7280;
  margin: 0 0 32px;
}

.error-actions {
  display: flex;
  justify-content: center;
  gap: 12px;
}
</style>
