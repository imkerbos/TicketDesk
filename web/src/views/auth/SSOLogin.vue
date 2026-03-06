<template>
  <div class="sso-login-page">
    <div class="login-container">
      <div v-if="error" class="login-error">
        <el-icon class="error-icon" :size="48"><CircleCloseFilled /></el-icon>
        <p class="error-text">{{ error }}</p>
        <el-button type="primary" size="large" class="back-button" @click="goToLogin">
          返回登录
        </el-button>
      </div>
      <div v-else class="login-loading">
        <el-icon class="loading-icon" :size="48"><Loading /></el-icon>
        <p class="loading-text">正在跳转到统一认证...</p>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { Loading, CircleCloseFilled } from '@element-plus/icons-vue'
import { getSSOAuthorizeURL } from '@/api/auth'

const router = useRouter()
const error = ref('')

const goToLogin = () => {
  router.push('/login')
}

onMounted(async () => {
  try {
    const res = await getSSOAuthorizeURL()
    const { authorize_url } = res.data.data
    window.location.href = authorize_url
  } catch {
    error.value = 'SSO 未启用或获取授权地址失败'
  }
})
</script>

<style scoped>
.sso-login-page {
  display: flex;
  align-items: center;
  justify-content: center;
  min-height: 100vh;
  background-color: var(--td-bg-page);
}

.login-container {
  text-align: center;
  padding: 48px;
}

.login-loading {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 16px;
}

.loading-icon {
  color: var(--td-color-primary);
  animation: spin 1s linear infinite;
}

@keyframes spin {
  from { transform: rotate(0deg); }
  to { transform: rotate(360deg); }
}

.loading-text {
  font-size: 16px;
  color: var(--td-text-secondary);
}

.login-error {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 16px;
}

.error-icon {
  color: var(--td-color-danger);
}

.error-text {
  font-size: 16px;
  color: var(--td-text-regular);
}

.back-button {
  margin-top: 8px;
  min-width: 160px;
  height: 44px;
  border-radius: 8px;
}
</style>
