<template>
  <div class="forgot-password-page">
    <!-- 左侧品牌区域 -->
    <div class="brand-section">
      <div class="brand-content">
        <div class="brand-logo">
          <img v-if="brandStore.logoUrl" :src="brandStore.logoUrl" :alt="brandStore.systemName" class="logo-custom" />
          <span class="logo-text">{{ brandStore.systemName }}</span>
        </div>
        <h1 class="brand-title">忘记密码</h1>
        <p class="brand-description">
          请输入您的注册邮箱，我们将发送重置密码链接到您的邮箱。
        </p>
      </div>
      <div class="brand-footer">
        <span>{{ brandStore.copyrightText }}</span>
      </div>
    </div>

    <!-- 右侧表单区域 -->
    <div class="form-section">
      <div class="form-container">
        <div class="form-header">
          <h2 class="form-title">重置密码</h2>
          <p class="form-subtitle">输入您的邮箱地址</p>
        </div>

        <el-form
          ref="formRef"
          :model="form"
          :rules="rules"
          class="forgot-form"
          label-position="top"
          hide-required-asterisk
          @submit.prevent="handleSubmit"
        >
          <el-form-item prop="email" label="邮箱地址">
            <el-input
              v-model="form.email"
              placeholder="请输入注册邮箱"
              size="large"
              class="form-input"
            >
              <template #prefix>
                <el-icon class="input-icon"><Message /></el-icon>
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
              {{ loading ? '发送中...' : '发送重置链接' }}
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
</template>

<script setup lang="ts">
import { ref, reactive } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage, type FormInstance, type FormRules } from 'element-plus'
import { Message, ArrowLeft } from '@element-plus/icons-vue'
import { forgotPassword } from '@/api/auth'
import { useBrandStore } from '@/stores/brand'

const router = useRouter()
const brandStore = useBrandStore()
const formRef = ref<FormInstance>()
const loading = ref(false)

const form = reactive({
  email: '',
})

const rules: FormRules = {
  email: [
    { required: true, message: '请输入邮箱地址', trigger: 'blur' },
    { type: 'email', message: '请输入有效的邮箱地址', trigger: 'blur' },
  ],
}

const handleSubmit = async () => {
  if (!formRef.value) return

  await formRef.value.validate(async (valid) => {
    if (!valid) return

    loading.value = true
    try {
      await forgotPassword({ email: form.email })
      ElMessage.success('如果该邮箱已注册，您将收到重置密码的邮件')
      // 3秒后跳转到登录页
      setTimeout(() => {
        router.push('/login')
      }, 3000)
    } catch {
      // 错误已在 request 拦截器中处理
    } finally {
      loading.value = false
    }
  })
}
</script>

<style scoped>
.forgot-password-page {
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

.forgot-form {
  width: 100%;
}

.forgot-form :deep(.el-form-item__label) {
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
