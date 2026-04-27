<template>
  <div class="reset-password-page">
    <!-- 左侧品牌区域 -->
    <div class="brand-section">
      <div class="brand-content">
        <div class="brand-logo">
          <img v-if="brandStore.logoUrl" :src="brandStore.logoUrl" :alt="brandStore.systemName" class="logo-custom" />
          <span class="logo-text">{{ brandStore.systemName }}</span>
        </div>
        <h1 class="brand-title">重置密码</h1>
        <p class="brand-description">
          请输入您的新密码，密码长度至少为 6 位。
        </p>
      </div>
      <div class="brand-footer">
        <span>{{ brandStore.copyrightText }}</span>
      </div>
    </div>

    <!-- 右侧表单区域 -->
    <div class="form-section">
      <div class="form-container">
        <div v-if="!tokenValid" class="error-state">
          <el-icon class="error-icon" :size="64"><CircleClose /></el-icon>
          <h2 class="error-title">链接无效或已过期</h2>
          <p class="error-message">重置密码链接已失效，请重新申请。</p>
          <el-button type="primary" size="large" @click="router.push('/forgot-password')">
            重新申请
          </el-button>
        </div>

        <div v-else>
          <div class="form-header">
            <h2 class="form-title">设置新密码</h2>
            <p class="form-subtitle">请输入您的新密码</p>
          </div>

          <el-form
            ref="formRef"
            :model="form"
            :rules="rules"
            class="reset-form"
            label-position="top"
            hide-required-asterisk
            @submit.prevent="handleSubmit"
          >
            <el-form-item prop="newPassword" label="新密码">
              <el-input
                v-model="form.newPassword"
                type="password"
                placeholder="请输入新密码（至少6位）"
                size="large"
                show-password
                class="form-input"
              >
                <template #prefix>
                  <el-icon class="input-icon"><Lock /></el-icon>
                </template>
              </el-input>
            </el-form-item>

            <el-form-item prop="confirmPassword" label="确认密码">
              <el-input
                v-model="form.confirmPassword"
                type="password"
                placeholder="请再次输入新密码"
                size="large"
                show-password
                class="form-input"
                @keyup.enter="handleSubmit"
              >
                <template #prefix>
                  <el-icon class="input-icon"><Lock /></el-icon>
                </template>
              </el-input>
            </el-form-item>

            <el-form-item>
              <el-button
                type="primary"
                size="large"
                class="submit-button"
                :loading="loading"
                @click="handleSubmit"
              >
                {{ loading ? '重置中...' : '重置密码' }}
              </el-button>
            </el-form-item>
          </el-form>

          <div class="form-footer">
            <router-link to="/login" class="back-link">
              <el-icon><ArrowLeft /></el-icon>
              <span>返回登录</span>
            </router-link>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { ElMessage, type FormInstance, type FormRules } from 'element-plus'
import { Lock, ArrowLeft, CircleClose } from '@element-plus/icons-vue'
import { verifyResetToken, resetPasswordWithToken } from '@/api/auth'
import { useBrandStore } from '@/stores/brand'

const router = useRouter()
const brandStore = useBrandStore()
const route = useRoute()
const formRef = ref<FormInstance>()
const loading = ref(false)
const tokenValid = ref(false)
const token = ref('')

const form = reactive({
  newPassword: '',
  confirmPassword: '',
})

const validateConfirmPassword = (_rule: any, value: any, callback: any) => {
  if (value === '') {
    callback(new Error('请再次输入密码'))
  } else if (value !== form.newPassword) {
    callback(new Error('两次输入的密码不一致'))
  } else {
    callback()
  }
}

const rules: FormRules = {
  newPassword: [
    { required: true, message: '请输入新密码', trigger: 'blur' },
    { min: 6, message: '密码长度至少为 6 位', trigger: 'blur' },
  ],
  confirmPassword: [
    { required: true, message: '请再次输入密码', trigger: 'blur' },
    { validator: validateConfirmPassword, trigger: 'blur' },
  ],
}

// 验证令牌
const checkToken = async () => {
  token.value = route.query.token as string
  if (!token.value) {
    ElMessage.error('缺少重置密码令牌')
    tokenValid.value = false
    return
  }

  try {
    await verifyResetToken(token.value)
    tokenValid.value = true
  } catch {
    tokenValid.value = false
  }
}

const handleSubmit = async () => {
  if (!formRef.value) return

  await formRef.value.validate(async (valid) => {
    if (!valid) return

    loading.value = true
    try {
      await resetPasswordWithToken({
        token: token.value,
        new_password: form.newPassword,
      })
      ElMessage.success('密码重置成功，请使用新密码登录')
      // 2秒后跳转到登录页
      setTimeout(() => {
        router.push('/login')
      }, 2000)
    } catch {
      // 错误已在 request 拦截器中处理
    } finally {
      loading.value = false
    }
  })
}

onMounted(() => {
  checkToken()
})
</script>

<style scoped>
.reset-password-page {
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

.logo-custom {
  width: 48px;
  height: 48px;
  border-radius: 12px;
  object-fit: contain;
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
  margin: 0;
}

.brand-footer {
  font-size: 13px;
  color: rgba(255, 255, 255, 0.5);
}

/* 右侧表单区域 */
.form-section {
  flex: 1;
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 48px;
}

.form-container {
  width: 100%;
  max-width: 400px;
}

/* 错误状态 */
.error-state {
  text-align: center;
  padding: 40px 20px;
}

.error-icon {
  color: var(--td-color-danger);
  margin-bottom: 24px;
}

.error-title {
  font-size: 24px;
  font-weight: 700;
  color: var(--td-text-primary);
  margin: 0 0 12px;
}

.error-message {
  font-size: 15px;
  color: var(--td-text-secondary);
  margin: 0 0 32px;
}

/* 表单 */
.form-header {
  margin-bottom: 36px;
}

.form-title {
  font-size: 28px;
  font-weight: 700;
  color: var(--td-text-primary);
  margin: 0 0 8px;
}

.form-subtitle {
  font-size: 15px;
  color: var(--td-text-secondary);
  margin: 0;
}

.reset-form {
  width: 100%;
}

.reset-form :deep(.el-form-item__label) {
  font-size: 14px;
  font-weight: 500;
  color: var(--td-text-regular);
  padding-bottom: 8px;
}

.form-input :deep(.el-input__wrapper) {
  padding: 4px 12px;
  border-radius: 8px;
  box-shadow: 0 0 0 1px var(--td-border-color);
  transition: all 0.2s;
}

.form-input :deep(.el-input__wrapper:hover) {
  box-shadow: 0 0 0 1px var(--td-border-color-dark);
}

.form-input :deep(.el-input__wrapper.is-focus) {
  box-shadow: 0 0 0 2px rgba(59, 130, 246, 0.3);
}

.input-icon {
  color: var(--td-text-placeholder);
}

.submit-button {
  width: 100%;
  height: 48px;
  font-size: 16px;
  font-weight: 600;
  border-radius: 8px;
  background: var(--td-color-primary);
  border: none;
  transition: all 0.3s;
}

.submit-button:hover {
  transform: translateY(-1px);
  box-shadow: 0 8px 20px rgba(59, 130, 246, 0.3);
}

.submit-button:active {
  transform: translateY(0);
}

.form-footer {
  margin-top: 24px;
  text-align: center;
}

.back-link {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  font-size: 14px;
  color: var(--td-color-primary);
  text-decoration: none;
  font-weight: 500;
  transition: color 0.2s;
}

.back-link:hover {
  color: var(--td-color-primary-hover);
}

/* 响应式设计 */
@media (max-width: 1024px) {
  .brand-section {
    display: none;
  }

  .form-section {
    flex: 1;
  }
}

@media (max-width: 480px) {
  .form-section {
    padding: 24px;
  }

  .form-title {
    font-size: 24px;
  }
}
</style>
