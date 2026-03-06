<template>
  <div class="error-page">
    <div class="error-content">
      <div class="error-code">403</div>
      <div class="error-illustration">
        <svg viewBox="0 0 200 120" fill="none" xmlns="http://www.w3.org/2000/svg">
          <circle cx="100" cy="60" r="50" fill="#fff7ed" stroke="#fed7aa" stroke-width="2" />
          <rect x="88" y="40" width="24" height="30" rx="4" stroke="#f59e0b" stroke-width="2.5" fill="none" />
          <path d="M92 40 V34 Q100 22 108 34 V40" stroke="#f59e0b" stroke-width="2.5" stroke-linecap="round" fill="none" />
          <circle cx="100" cy="53" r="3" fill="#f59e0b" />
          <line x1="100" y1="56" x2="100" y2="62" stroke="#f59e0b" stroke-width="2" stroke-linecap="round" />
        </svg>
      </div>
      <h2 class="error-title">没有访问权限</h2>
      <p class="error-desc">你没有权限访问此页面，请联系管理员</p>
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
  background: var(--td-bg-page);
}

.error-content {
  text-align: center;
  padding: 40px;
}

.error-code {
  font-size: 120px;
  font-weight: 800;
  color: var(--td-color-warning);
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
  color: var(--td-text-primary);
  margin: 0 0 8px;
}

.error-desc {
  font-size: 15px;
  color: var(--td-text-secondary);
  margin: 0 0 32px;
}

.error-actions {
  display: flex;
  justify-content: center;
  gap: 12px;
}
</style>
