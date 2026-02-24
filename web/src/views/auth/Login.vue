<template>
  <div class="login-page">
    <!-- 左侧品牌区域 -->
    <div class="brand-section">
      <div class="brand-content">
        <div class="brand-logo">
          <span class="logo-icon">T</span>
          <span class="logo-text">TicketDesk</span>
        </div>
        <h1 class="brand-title">工单与告警联动系统</h1>
        <p class="brand-description">
          一切问题都是工单，一切告警都必须被跟进。<br />
          为运维与技术团队打造的项目化工单管理平台。
        </p>
        <div class="feature-list">
          <div class="feature-item">
            <el-icon class="feature-icon"><Tickets /></el-icon>
            <span>项目化工单管理</span>
          </div>
          <div class="feature-item">
            <el-icon class="feature-icon"><Bell /></el-icon>
            <span>告警自动建单</span>
          </div>
          <div class="feature-item">
            <el-icon class="feature-icon"><Connection /></el-icon>
            <span>审批工作流引擎</span>
          </div>
        </div>
      </div>
      <div class="brand-footer">
        <span>&copy; 2026 TicketDesk. All rights reserved.</span>
      </div>
    </div>

    <!-- 右侧登录表单区域 -->
    <div class="login-section">
      <div class="login-container">
        <div class="login-header">
          <h2 class="login-title">欢迎回来</h2>
          <p class="login-subtitle">请登录您的账户</p>
        </div>

        <el-form
          ref="formRef"
          :model="form"
          :rules="rules"
          class="login-form"
          label-position="top"
          hide-required-asterisk
          @submit.prevent="handleLogin"
        >
          <el-form-item prop="username" label="用户名">
            <el-input
              v-model="form.username"
              placeholder="请输入用户名"
              size="large"
              class="form-input"
            >
              <template #prefix>
                <el-icon class="input-icon"><User /></el-icon>
              </template>
            </el-input>
          </el-form-item>

          <el-form-item prop="password" class="password-item">
            <template #label>
              <div class="password-label-row">
                <span>密码</span>
                <router-link to="/forgot-password" class="forgot-link">忘记密码？</router-link>
              </div>
            </template>
            <el-input
              v-model="form.password"
              type="password"
              placeholder="请输入密码"
              size="large"
              show-password
              class="form-input"
              @keyup.enter="handleLogin"
            >
              <template #prefix>
                <el-icon class="input-icon"><Lock /></el-icon>
              </template>
            </el-input>
          </el-form-item>

          <el-form-item class="remember-item">
            <el-checkbox v-model="rememberMe">记住我</el-checkbox>
          </el-form-item>

          <el-form-item>
            <el-button
              type="primary"
              size="large"
              class="login-button"
              :loading="loading"
              @click="handleLogin"
            >
              {{ loading ? '登录中...' : '登录' }}
            </el-button>
          </el-form-item>
        </el-form>

        <!-- SSO 登录区域 -->
        <div v-if="ssoConfig?.enabled" class="sso-section">
          <div class="sso-divider">
            <span class="sso-divider-text">或</span>
          </div>
          <el-button
            size="large"
            class="sso-button"
            :loading="ssoLoading"
            @click="handleSSOLogin"
          >
            {{ ssoConfig.provider_name || 'SSO 登录' }}
          </el-button>
        </div>

        <div class="login-footer">
          <span class="footer-text">还没有账户？</span>
          <a href="#" class="register-link" @click.prevent>联系管理员</a>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage, type FormInstance, type FormRules } from 'element-plus'
import { User, Lock, Tickets, Bell, Connection } from '@element-plus/icons-vue'
import { login, getSSOConfig, getSSOAuthorizeURL, type SSOConfigResponse } from '@/api/auth'
import { useUserStore } from '@/stores/user'

const router = useRouter()
const userStore = useUserStore()
const formRef = ref<FormInstance>()
const loading = ref(false)
const rememberMe = ref(false)
const ssoConfig = ref<SSOConfigResponse | null>(null)
const ssoLoading = ref(false)

const form = reactive({
  username: '',
  password: '',
})

const rules: FormRules = {
  username: [
    { required: true, message: '请输入用户名', trigger: 'blur' },
  ],
  password: [
    { required: true, message: '请输入密码', trigger: 'blur' },
    { min: 6, message: '密码长度至少为 6 位', trigger: 'blur' },
  ],
}

const handleLogin = async () => {
  if (!formRef.value) return

  await formRef.value.validate(async (valid) => {
    if (!valid) return

    loading.value = true
    try {
      const res = await login({
        username: form.username,
        password: form.password,
      })

      // 使用 store 保存用户信息
      const { access_token, refresh_token, user } = res.data.data
      userStore.login(access_token, refresh_token, user)

      ElMessage.success(`欢迎回来，${user.display_name || user.username}`)
      router.push('/')
    } catch {
      // 错误已在 request 拦截器中处理
    } finally {
      loading.value = false
    }
  })
}

// SSO 登录
const handleSSOLogin = async () => {
  ssoLoading.value = true
  try {
    const res = await getSSOAuthorizeURL()
    const { authorize_url } = res.data.data
    window.location.href = authorize_url
  } catch {
    ElMessage.error('获取 SSO 登录地址失败')
    ssoLoading.value = false
  }
}

// 获取 SSO 配置
onMounted(async () => {
  try {
    const res = await getSSOConfig()
    ssoConfig.value = res.data.data
  } catch {
    // SSO 配置获取失败不影响正常登录
  }
})
</script>

<style scoped>
.login-page {
  display: flex;
  min-height: 100vh;
  background-color: #f8fafc;
}

/* 左侧品牌区域 */
.brand-section {
  flex: 1;
  display: flex;
  flex-direction: column;
  justify-content: space-between;
  padding: 48px;
  background: linear-gradient(135deg, #1e1e2d 0%, #2d2d44 100%);
  color: #fff;
}

.brand-content {
  max-width: 480px;
}

.brand-logo {
  display: flex;
  align-items: center;
  gap: 12px;
  margin-bottom: 48px;
}

.logo-icon {
  width: 48px;
  height: 48px;
  background: linear-gradient(135deg, #3b82f6 0%, #8b5cf6 100%);
  border-radius: 12px;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 24px;
  font-weight: 700;
}

.logo-text {
  font-size: 24px;
  font-weight: 700;
}

.brand-title {
  font-size: 36px;
  font-weight: 700;
  margin: 0 0 16px;
  line-height: 1.3;
}

.brand-description {
  font-size: 16px;
  line-height: 1.8;
  color: rgba(255, 255, 255, 0.7);
  margin: 0 0 48px;
}

.feature-list {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.feature-item {
  display: flex;
  align-items: center;
  gap: 12px;
  font-size: 15px;
  color: rgba(255, 255, 255, 0.85);
}

.feature-icon {
  width: 40px;
  height: 40px;
  background: rgba(255, 255, 255, 0.1);
  border-radius: 10px;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 18px;
}

.brand-footer {
  font-size: 13px;
  color: rgba(255, 255, 255, 0.5);
}

/* 右侧登录表单区域 */
.login-section {
  flex: 1;
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 48px;
}

.login-container {
  width: 100%;
  max-width: 400px;
}

.login-header {
  margin-bottom: 36px;
}

.login-title {
  font-size: 28px;
  font-weight: 700;
  color: #1f2937;
  margin: 0 0 8px;
}

.login-subtitle {
  font-size: 15px;
  color: #6b7280;
  margin: 0;
}

.login-form {
  width: 100%;
}

.login-form :deep(.el-form-item__label) {
  font-size: 14px;
  font-weight: 500;
  color: #374151;
  padding-bottom: 8px;
}

.password-item :deep(.el-form-item__label) {
  width: 100%;
}

.password-label-row {
  display: flex;
  justify-content: space-between;
  align-items: center;
  width: 100%;
}

.forgot-link {
  font-size: 13px;
  color: #3b82f6;
  text-decoration: none;
  transition: color 0.2s;
  font-weight: 400;
}

.forgot-link:hover {
  color: #2563eb;
}

.form-input :deep(.el-input__wrapper) {
  padding: 4px 12px;
  border-radius: 8px;
  box-shadow: 0 0 0 1px #e5e7eb;
  transition: all 0.2s;
}

.form-input :deep(.el-input__wrapper:hover) {
  box-shadow: 0 0 0 1px #d1d5db;
}

.form-input :deep(.el-input__wrapper.is-focus) {
  box-shadow: 0 0 0 2px rgba(59, 130, 246, 0.3);
}

.input-icon {
  color: #9ca3af;
}

.remember-item {
  margin-bottom: 24px;
}

.remember-item :deep(.el-checkbox__label) {
  font-size: 14px;
  color: #4b5563;
}

.login-button {
  width: 100%;
  height: 48px;
  font-size: 16px;
  font-weight: 600;
  border-radius: 8px;
  background: linear-gradient(135deg, #3b82f6 0%, #8b5cf6 100%);
  border: none;
  transition: all 0.3s;
}

.login-button:hover {
  transform: translateY(-1px);
  box-shadow: 0 8px 20px rgba(59, 130, 246, 0.3);
}

.login-button:active {
  transform: translateY(0);
}

/* SSO 登录区域 */
.sso-section {
  margin-top: 8px;
}

.sso-divider {
  display: flex;
  align-items: center;
  margin: 16px 0;
}

.sso-divider::before,
.sso-divider::after {
  content: '';
  flex: 1;
  border-top: 1px solid #e5e7eb;
}

.sso-divider-text {
  padding: 0 16px;
  font-size: 13px;
  color: #9ca3af;
}

.sso-button {
  width: 100%;
  height: 48px;
  font-size: 15px;
  font-weight: 500;
  border-radius: 8px;
  border: 1px solid #d1d5db;
  background: #fff;
  color: #374151;
  transition: all 0.2s;
}

.sso-button:hover {
  border-color: #3b82f6;
  color: #3b82f6;
  background: #f0f7ff;
}

.login-footer {
  margin-top: 24px;
  text-align: center;
}

.footer-text {
  font-size: 14px;
  color: #6b7280;
}

.register-link {
  font-size: 14px;
  color: #3b82f6;
  text-decoration: none;
  font-weight: 500;
  margin-left: 4px;
  transition: color 0.2s;
}

.register-link:hover {
  color: #2563eb;
}

/* 响应式设计 */
@media (max-width: 1024px) {
  .brand-section {
    display: none;
  }

  .login-section {
    flex: 1;
  }
}

@media (max-width: 480px) {
  .login-section {
    padding: 24px;
  }

  .login-title {
    font-size: 24px;
  }
}
</style>
