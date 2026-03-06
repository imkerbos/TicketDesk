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
          <div class="feature-item stagger" style="--i: 0">
            <el-icon class="feature-icon"><Tickets /></el-icon>
            <span>项目化工单管理</span>
          </div>
          <div class="feature-item stagger" style="--i: 1">
            <el-icon class="feature-icon"><Bell /></el-icon>
            <span>告警自动建单</span>
          </div>
          <div class="feature-item stagger" style="--i: 2">
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
      <div :class="['login-container', { shake: shaking }]" @animationend="shaking = false">
        <div class="login-header fade-up" style="--i: 0">
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
          <el-form-item prop="username" label="用户名" class="fade-up" style="--i: 1">
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

          <el-form-item prop="password" label="密码" class="fade-up" style="--i: 2">
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

          <el-form-item class="remember-row fade-up" style="--i: 3">
            <div class="remember-forgot">
              <el-checkbox v-model="rememberMe">记住我</el-checkbox>
              <router-link to="/forgot-password" class="forgot-link">忘记密码？</router-link>
            </div>
          </el-form-item>

          <el-form-item class="fade-up" style="--i: 4">
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
        <div v-if="ssoConfig?.enabled" class="sso-section fade-up" style="--i: 5">
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
const shaking = ref(false)

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
    if (!valid) {
      shaking.value = true
      return
    }

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
      shaking.value = true
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
  background-color: var(--td-bg-page);
}

/* 左侧品牌区域 */
.brand-section {
  flex: 1;
  display: flex;
  flex-direction: column;
  justify-content: space-between;
  padding: 48px;
  background: var(--td-sidebar-bg);
  color: var(--td-text-white);
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
  background: var(--td-color-primary);
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
  color: var(--td-text-primary);
  margin: 0 0 8px;
}

.login-subtitle {
  font-size: 15px;
  color: var(--td-text-secondary);
  margin: 0;
}

.login-form {
  width: 100%;
}

.login-form :deep(.el-form-item__label) {
  font-size: 14px;
  font-weight: 500;
  color: var(--td-text-regular);
  padding-bottom: 8px;
}

.remember-forgot {
  display: flex;
  justify-content: space-between;
  align-items: center;
  width: 100%;
}

.forgot-link {
  font-size: 13px;
  color: var(--td-color-primary);
  text-decoration: none;
  transition: color 150ms ease-out;
  font-weight: 400;
}

.forgot-link:hover {
  color: var(--td-color-primary-hover);
}

.form-input :deep(.el-input__wrapper) {
  padding: 4px 12px;
  border-radius: 8px;
  box-shadow: 0 0 0 1px var(--td-border-color);
  transition: box-shadow 150ms ease-out;
}

.form-input :deep(.el-input__wrapper:hover) {
  box-shadow: 0 0 0 1px var(--td-border-color-dark);
}

.form-input :deep(.el-input__wrapper.is-focus) {
  box-shadow: 0 0 0 2px rgba(59, 130, 246, 0.2);
}

.input-icon {
  color: var(--td-text-placeholder);
  transition: color 150ms ease-out;
}

.form-input :deep(.el-input__wrapper.is-focus .input-icon) {
  color: var(--td-color-primary);
}

.remember-row {
  margin-bottom: 24px;
}

.remember-row :deep(.el-checkbox__label) {
  font-size: 14px;
  color: var(--td-text-regular);
}

.login-button {
  width: 100%;
  height: 48px;
  font-size: 16px;
  font-weight: 600;
  border-radius: 8px;
  background: var(--td-color-primary);
  border: none;
  transition: background-color 150ms ease-out, box-shadow 150ms ease-out;
}

.login-button:hover {
  background: var(--td-color-primary-hover);
  box-shadow: 0 2px 8px rgba(59, 130, 246, 0.25);
}

.login-button:active {
  background: var(--td-color-primary-active);
  box-shadow: none;
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
  border-top: 1px solid var(--td-border-color);
}

.sso-divider-text {
  padding: 0 16px;
  font-size: 13px;
  color: var(--td-text-placeholder);
}

.sso-button {
  width: 100%;
  height: 48px;
  font-size: 15px;
  font-weight: 500;
  border-radius: 8px;
  border: 1px solid var(--td-border-color-dark);
  background: var(--td-bg-card);
  color: var(--td-text-regular);
  transition: border-color 150ms ease-out, color 150ms ease-out, background-color 150ms ease-out;
}

.sso-button:hover {
  border-color: var(--td-color-primary);
  color: var(--td-color-primary);
  background: var(--td-tag-primary-bg);
}


/* 入场动画 */
@keyframes fadeUp {
  from {
    opacity: 0;
    transform: translateY(12px);
  }
  to {
    opacity: 1;
    transform: translateY(0);
  }
}

.fade-up {
  opacity: 0;
  animation: fadeUp 250ms ease-out forwards;
  animation-delay: calc(var(--i, 0) * 60ms + 80ms);
}

/* 左侧特性列表 stagger 入场 */
.stagger {
  opacity: 0;
  animation: fadeUp 250ms ease-out forwards;
  animation-delay: calc(var(--i, 0) * 60ms + 200ms);
}

/* 登录失败 shake */
@keyframes shake {
  0%, 100% { transform: translateX(0); }
  20% { transform: translateX(-6px); }
  40% { transform: translateX(5px); }
  60% { transform: translateX(-4px); }
  80% { transform: translateX(2px); }
}

.shake {
  animation: shake 250ms ease-out;
}

/* prefers-reduced-motion 降级 */
@media (prefers-reduced-motion: reduce) {
  .fade-up,
  .stagger {
    opacity: 1;
    animation: none;
  }
  .shake {
    animation: none;
  }
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
