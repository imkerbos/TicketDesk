<template>
  <div class="error-page">
    <div class="error-content">
      <div class="error-code">500</div>
      <div class="error-illustration">
        <svg viewBox="0 0 200 120" fill="none" xmlns="http://www.w3.org/2000/svg">
          <circle cx="100" cy="60" r="50" fill="#fef2f2" stroke="#fecaca" stroke-width="2"/>
          <path d="M85 45 L115 75" stroke="#ef4444" stroke-width="3" stroke-linecap="round"/>
          <path d="M115 45 L85 75" stroke="#ef4444" stroke-width="3" stroke-linecap="round"/>
        </svg>
      </div>
      <h2 class="error-title">服务器异常</h2>
      <p class="error-desc">服务器开小差了，请稍后再试</p>
      <div class="error-actions">
        <el-button type="primary" @click="handleRefresh">刷新页面</el-button>
        <el-button @click="router.push('/dashboard')">返回首页</el-button>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { useRouter } from 'vue-router'
const router = useRouter()

const handleRefresh = () => {
  window.location.reload()
}
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
  background: linear-gradient(135deg, #ef4444 0%, #dc2626 100%);
  -webkit-background-clip: text;
  -webkit-text-fill-color: transparent;
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
