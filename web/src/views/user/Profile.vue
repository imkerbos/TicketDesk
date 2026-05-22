<template>
  <div class="profile-container">
    <el-row :gutter="24">
      <!-- 左侧：个人信息卡片 -->
      <el-col :xs="24" :lg="8">
        <el-card shadow="never" class="profile-card">
          <div class="profile-header">
            <div class="avatar-wrapper">
              <el-avatar :size="100" class="user-avatar">
                {{ profile?.display_name?.charAt(0) || profile?.username?.charAt(0) || '?' }}
              </el-avatar>
              <div class="avatar-overlay">
                <el-icon><Camera /></el-icon>
              </div>
            </div>
            <h2 class="username">{{ profile?.display_name || profile?.username }}</h2>
            <p class="user-role">
              <el-tag :type="profile?.roles?.includes('admin') ? 'danger' : 'info'" size="small">
                {{ profile?.roles?.includes('admin') ? '管理员' : '普通用户' }}
              </el-tag>
            </p>
          </div>

          <el-divider />

          <div class="profile-info">
            <div class="info-item">
              <el-icon><User /></el-icon>
              <div class="info-content">
                <span class="info-label">用户名</span>
                <span class="info-value">{{ profile?.username }}</span>
              </div>
            </div>
            <div class="info-item">
              <el-icon><Message /></el-icon>
              <div class="info-content">
                <span class="info-label">邮箱</span>
                <span class="info-value">{{ profile?.email || '未设置' }}</span>
              </div>
            </div>
            <div class="info-item">
              <el-icon><Calendar /></el-icon>
              <div class="info-content">
                <span class="info-label">注册时间</span>
                <span class="info-value">{{ formatDate(profile?.created_at) }}</span>
              </div>
            </div>
            <div class="info-item">
              <el-icon><Link /></el-icon>
              <div class="info-content">
                <span class="info-label">认证方式</span>
                <span class="info-value">
                  <el-tag :type="profile?.auth_source === 'sso' ? 'warning' : 'info'" size="small">
                    {{ profile?.auth_source === 'sso' ? `SSO (${profile?.sso_provider || 'SSO'})` : '本地账号' }}
                  </el-tag>
                </span>
              </div>
            </div>
          </div>
        </el-card>
      </el-col>

      <!-- 右侧：设置表单 -->
      <el-col :xs="24" :lg="16">
        <!-- 基本信息 -->
        <el-card shadow="never" class="settings-card">
          <template #header>
            <div class="card-header">
              <span class="card-title">
                <el-icon><EditPen /></el-icon>
                基本信息
              </span>
            </div>
          </template>

          <el-form
            ref="profileFormRef"
            :model="profileForm"
            :rules="profileRules"
            label-width="100px"
            class="settings-form"
          >
            <el-form-item label="显示名称" prop="display_name">
              <el-input
                v-model="profileForm.display_name"
                placeholder="请输入显示名称"
                maxlength="50"
                show-word-limit
              />
            </el-form-item>
            <el-form-item label="邮箱地址" prop="email">
              <el-input
                v-model="profileForm.email"
                placeholder="请输入邮箱地址"
                type="email"
              />
            </el-form-item>
            <el-form-item>
              <el-button type="primary" :loading="profileLoading" @click="submitProfile">
                保存修改
              </el-button>
            </el-form-item>
          </el-form>
        </el-card>

        <!-- 修改密码（仅本地用户） -->
        <el-card v-if="profile?.auth_source !== 'sso'" shadow="never" class="settings-card">
          <template #header>
            <div class="card-header">
              <span class="card-title">
                <el-icon><Lock /></el-icon>
                修改密码
              </span>
            </div>
          </template>

          <el-form
            ref="passwordFormRef"
            :model="passwordForm"
            :rules="passwordRules"
            label-width="100px"
            class="settings-form"
          >
            <el-form-item label="当前密码" prop="old_password">
              <el-input
                v-model="passwordForm.old_password"
                type="password"
                placeholder="请输入当前密码"
                show-password
              />
            </el-form-item>
            <el-form-item label="新密码" prop="new_password">
              <el-input
                v-model="passwordForm.new_password"
                type="password"
                placeholder="请输入新密码（至少6位）"
                show-password
              />
            </el-form-item>
            <el-form-item label="确认密码" prop="confirm_password">
              <el-input
                v-model="passwordForm.confirm_password"
                type="password"
                placeholder="请再次输入新密码"
                show-password
              />
            </el-form-item>
            <el-form-item>
              <el-button type="primary" :loading="passwordLoading" @click="submitPassword">
                修改密码
              </el-button>
            </el-form-item>
          </el-form>
        </el-card>

        <!-- 账户安全 -->
        <el-card shadow="never" class="settings-card">
          <template #header>
            <div class="card-header">
              <span class="card-title">
                <el-icon><Key /></el-icon>
                账户安全
              </span>
            </div>
          </template>

          <div class="security-info">
            <div class="security-item">
              <div class="security-left">
                <el-icon class="security-icon" :class="profile?.auth_source === 'sso' ? 'warning' : 'success'">
                  <component :is="profile?.auth_source === 'sso' ? Warning : CircleCheck" />
                </el-icon>
                <div class="security-content">
                  <span class="security-title">登录密码</span>
                  <span class="security-desc">
                    {{ profile?.auth_source === 'sso' ? '当前为 SSO 账号，密码由 SSO 提供方管理' : '已设置，建议定期更换密码' }}
                  </span>
                </div>
              </div>
              <el-button v-if="profile?.auth_source !== 'sso'" link type="primary" @click="scrollToPassword">修改</el-button>
            </div>

            <el-divider />

            <div class="security-item">
              <div class="security-left">
                <el-icon class="security-icon" :class="profile?.email ? 'success' : 'warning'">
                  <component :is="profile?.email ? CircleCheck : Warning" />
                </el-icon>
                <div class="security-content">
                  <span class="security-title">邮箱绑定</span>
                  <span class="security-desc">
                    {{ profile?.email ? `已绑定：${profile.email}` : '未绑定，建议绑定邮箱用于找回密码' }}
                  </span>
                </div>
              </div>
              <el-button link type="primary" @click="scrollToProfile">
                {{ profile?.email ? '修改' : '绑定' }}
              </el-button>
            </div>

            <el-divider />

            <div class="security-item">
              <div class="security-left">
                <el-icon class="security-icon" :class="mfaStatus?.enabled ? 'success' : 'warning'">
                  <component :is="mfaStatus?.enabled ? CircleCheck : Warning" />
                </el-icon>
                <div class="security-content">
                  <span class="security-title">双因素认证 (MFA)</span>
                  <span class="security-desc">
                    {{ mfaStatus?.enabled ? '已启用，登录时需要验证码' : '未启用，建议开启以增强账户安全' }}
                  </span>
                </div>
              </div>
              <el-button link type="primary" @click="mfaStatus?.enabled ? showDisableMFADialog() : startMFASetup()">
                {{ mfaStatus?.enabled ? '禁用' : '启用' }}
              </el-button>
            </div>
          </div>
        </el-card>
      </el-col>
    </el-row>

    <!-- MFA 设置对话框 -->
    <el-dialog v-model="mfaSetupDialogVisible" title="设置双因素认证" width="500px" :close-on-click-modal="false">
      <div class="mfa-setup">
        <div class="mfa-step">
          <h4>第 1 步：安装验证器应用</h4>
          <p>请在手机上安装 Google Authenticator、Microsoft Authenticator 或其他 TOTP 验证器应用。</p>
        </div>

        <div class="mfa-step">
          <h4>第 2 步：扫描二维码</h4>
          <p>使用验证器应用扫描以下二维码：</p>
          <div v-if="mfaSetupData" class="qr-code">
            <img :src="qrCodeUrl" alt="MFA QR Code" />
          </div>
          <p v-if="mfaSetupData" class="manual-key">
            或手动输入密钥：<code>{{ mfaSetupData.secret }}</code>
          </p>
        </div>

        <div class="mfa-step">
          <h4>第 3 步：输入验证码</h4>
          <p>输入验证器应用显示的 6 位数字验证码：</p>
          <el-input
            v-model="mfaVerifyCode"
            placeholder="000000"
            maxlength="6"
            class="verify-code-input"
            @keyup.enter="confirmEnableMFA"
          />
        </div>
      </div>

      <template #footer>
        <el-button @click="mfaSetupDialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="mfaEnabling" @click="confirmEnableMFA">
          启用 MFA
        </el-button>
      </template>
    </el-dialog>

    <!-- MFA 禁用对话框 -->
    <el-dialog v-model="mfaDisableDialogVisible" title="禁用双因素认证" width="400px">
      <p>请输入验证器应用中的 6 位验证码以确认禁用 MFA：</p>
      <el-input
        v-model="mfaDisableCode"
        placeholder="000000"
        maxlength="6"
        class="verify-code-input"
        @keyup.enter="confirmDisableMFA"
      />
      <template #footer>
        <el-button @click="mfaDisableDialogVisible = false">取消</el-button>
        <el-button type="danger" :loading="mfaDisabling" @click="confirmDisableMFA">
          禁用 MFA
        </el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { ElMessage, type FormInstance, type FormRules } from 'element-plus'
import {
  User,
  Message,
  Calendar,
  EditPen,
  Lock,
  Key,
  Camera,
  CircleCheck,
  Warning,
  Link,
} from '@element-plus/icons-vue'
import { getCurrentUser, updateCurrentUser, updatePassword, getMFAStatus, setupMFA, enableMFA, disableMFA } from '@/api/user'
import type { UserProfile, UpdatePasswordRequest } from '@/types/user'
import type { MFAStatusResponse, MFASetupResponse } from '@/api/user'
import dayjs from 'dayjs'

// 用户信息
const profile = ref<UserProfile & { created_at?: string } | null>(null)

// 基本信息表单
const profileFormRef = ref<FormInstance>()
const profileLoading = ref(false)
const profileForm = reactive({
  display_name: '',
  email: '',
})

const profileRules: FormRules = {
  display_name: [
    { required: true, message: '请输入显示名称', trigger: 'blur' },
    { max: 50, message: '最多50个字符', trigger: 'blur' },
  ],
  email: [
    { type: 'email', message: '请输入正确的邮箱地址', trigger: 'blur' },
  ],
}

// 密码表单
const passwordFormRef = ref<FormInstance>()
const passwordLoading = ref(false)
const passwordForm = reactive({
  old_password: '',
  new_password: '',
  confirm_password: '',
})

const validateConfirmPassword = (_rule: unknown, value: string, callback: (error?: Error) => void) => {
  if (value !== passwordForm.new_password) {
    callback(new Error('两次输入的密码不一致'))
  } else {
    callback()
  }
}

const passwordRules: FormRules = {
  old_password: [
    { required: true, message: '请输入当前密码', trigger: 'blur' },
  ],
  new_password: [
    { required: true, message: '请输入新密码', trigger: 'blur' },
    { min: 6, message: '密码至少6位', trigger: 'blur' },
  ],
  confirm_password: [
    { required: true, message: '请确认新密码', trigger: 'blur' },
    { validator: validateConfirmPassword, trigger: 'blur' },
  ],
}

// 加载用户信息
const loadProfile = async () => {
  try {
    const { data } = await getCurrentUser()
    profile.value = data.data
    profileForm.display_name = data.data.display_name || ''
    profileForm.email = data.data.email || ''
  } catch {
    // ignored
  }
}

// 保存基本信息
const submitProfile = async () => {
  if (!profileFormRef.value) return

  await profileFormRef.value.validate(async (valid) => {
    if (!valid) return

    profileLoading.value = true
    try {
      await updateCurrentUser({
        display_name: profileForm.display_name,
        email: profileForm.email,
      })
      ElMessage.success('保存成功')
      loadProfile()
    } catch {
      // ignored
    } finally {
      profileLoading.value = false
    }
  })
}

// 修改密码
const submitPassword = async () => {
  if (!passwordFormRef.value) return

  await passwordFormRef.value.validate(async (valid) => {
    if (!valid) return

    passwordLoading.value = true
    try {
      const data: UpdatePasswordRequest = {
        old_password: passwordForm.old_password,
        new_password: passwordForm.new_password,
      }
      await updatePassword(data)
      ElMessage.success('密码修改成功')
      // 清空表单
      passwordForm.old_password = ''
      passwordForm.new_password = ''
      passwordForm.confirm_password = ''
      passwordFormRef.value?.resetFields()
    } catch {
      // ignored
    } finally {
      passwordLoading.value = false
    }
  })
}

// 滚动到密码区域
const scrollToPassword = () => {
  passwordFormRef.value?.$el?.scrollIntoView({ behavior: 'smooth', block: 'center' })
}

// 滚动到基本信息区域
const scrollToProfile = () => {
  profileFormRef.value?.$el?.scrollIntoView({ behavior: 'smooth', block: 'center' })
}

// 格式化日期
const formatDate = (date?: string) => {
  if (!date) return '-'
  return dayjs(date).format('YYYY-MM-DD')
}

// ============ MFA 相关 ============
const mfaStatus = ref<MFAStatusResponse | null>(null)
const mfaSetupDialogVisible = ref(false)
const mfaDisableDialogVisible = ref(false)
const mfaSetupData = ref<MFASetupResponse | null>(null)
const mfaVerifyCode = ref('')
const mfaDisableCode = ref('')
const mfaEnabling = ref(false)
const mfaDisabling = ref(false)
const qrCodeUrl = ref('')

// 加载 MFA 状态
const loadMFAStatus = async () => {
  try {
    const { data } = await getMFAStatus()
    mfaStatus.value = data.data
  } catch {
    // ignored
  }
}

// 开始 MFA 设置
const startMFASetup = async () => {
  try {
    const { data } = await setupMFA()
    mfaSetupData.value = data.data
    // 直接使用后端返回的 base64 编码的 QR 码图片
    qrCodeUrl.value = data.data.qr_code_data
    mfaVerifyCode.value = ''
    mfaSetupDialogVisible.value = true
  } catch {
    ElMessage.error('启动 MFA 设置失败')
  }
}

// 确认启用 MFA
const confirmEnableMFA = async () => {
  if (mfaVerifyCode.value.length !== 6) {
    ElMessage.warning('请输入 6 位验证码')
    return
  }

  mfaEnabling.value = true
  try {
    await enableMFA(mfaVerifyCode.value)
    ElMessage.success('MFA 已启用')
    mfaSetupDialogVisible.value = false
    loadMFAStatus()
  } catch {
    ElMessage.error('验证码错误，请重试')
  } finally {
    mfaEnabling.value = false
  }
}

// 显示禁用 MFA 对话框
const showDisableMFADialog = () => {
  mfaDisableCode.value = ''
  mfaDisableDialogVisible.value = true
}

// 确认禁用 MFA
const confirmDisableMFA = async () => {
  if (mfaDisableCode.value.length !== 6) {
    ElMessage.warning('请输入 6 位验证码')
    return
  }

  mfaDisabling.value = true
  try {
    await disableMFA(mfaDisableCode.value)
    ElMessage.success('MFA 已禁用')
    mfaDisableDialogVisible.value = false
    loadMFAStatus()
  } catch {
    ElMessage.error('验证码错误，请重试')
  } finally {
    mfaDisabling.value = false
  }
}

// 初始化
onMounted(() => {
  loadProfile()
  loadMFAStatus()
})
</script>

<style scoped lang="scss">
.profile-container {
  max-width: 1200px;
  margin: 0 auto;
}

.profile-card {
  text-align: center;

  .profile-header {
    padding: 20px 0;

    .avatar-wrapper {
      position: relative;
      display: inline-block;
      cursor: pointer;

      .user-avatar {
        background: var(--td-color-primary);
        font-size: 36px;
        font-weight: 600;
      }

      .avatar-overlay {
        position: absolute;
        bottom: 0;
        right: 0;
        width: 32px;
        height: 32px;
        background: var(--td-bg-card);
        border-radius: 50%;
        display: flex;
        align-items: center;
        justify-content: center;
        box-shadow: 0 2px 8px rgba(0, 0, 0, 0.15);
        color: var(--td-text-secondary);
        transition: all 150ms ease-out;

        &:hover {
          background: var(--td-color-primary);
          color: var(--td-text-white);
        }
      }
    }

    .username {
      margin: 16px 0 8px;
      font-size: 20px;
      font-weight: 600;
      color: var(--td-text-primary);
    }

    .user-role {
      margin: 0;
    }
  }

  .profile-info {
    text-align: left;

    .info-item {
      display: flex;
      align-items: center;
      gap: 12px;
      padding: 12px 0;

      &:not(:last-child) {
        border-bottom: 1px solid var(--td-divider-color);
      }

      > .el-icon {
        font-size: 20px;
        color: var(--td-color-info);
      }

      .info-content {
        display: flex;
        flex-direction: column;
        gap: 4px;

        .info-label {
          font-size: 12px;
          color: var(--td-color-info);
        }

        .info-value {
          font-size: 14px;
          color: var(--td-text-primary);
        }
      }
    }
  }
}

.settings-card {
  margin-bottom: 20px;

  .card-header {
    display: flex;
    align-items: center;

    .card-title {
      display: flex;
      align-items: center;
      gap: 8px;
      font-size: 16px;
      font-weight: 600;
    }
  }

  .settings-form {
    max-width: 500px;
  }
}

.security-info {
  .security-item {
    display: flex;
    justify-content: space-between;
    align-items: center;
    padding: 8px 0;

    .security-left {
      display: flex;
      align-items: center;
      gap: 12px;

      .security-icon {
        font-size: 24px;

        &.success {
          color: var(--td-color-success);
        }

        &.warning {
          color: var(--td-color-warning);
        }
      }

      .security-content {
        display: flex;
        flex-direction: column;
        gap: 4px;

        .security-title {
          font-size: 14px;
          font-weight: 500;
          color: var(--td-text-primary);
        }

        .security-desc {
          font-size: 12px;
          color: var(--td-color-info);
        }
      }
    }
  }
}

@media (max-width: 992px) {
  .profile-card {
    margin-bottom: 20px;
  }
}

// MFA 设置样式
.mfa-setup {
  .mfa-step {
    margin-bottom: 24px;

    h4 {
      margin: 0 0 8px;
      font-size: 14px;
      font-weight: 600;
      color: var(--td-text-primary);
    }

    p {
      margin: 0;
      font-size: 14px;
      color: var(--td-text-secondary);
    }
  }

  .qr-code {
    display: flex;
    justify-content: center;
    margin: 16px 0;

    img {
      width: 200px;
      height: 200px;
      border: 1px solid var(--td-border-color);
      border-radius: 8px;
    }
  }

  .manual-key {
    text-align: center;
    font-size: 12px;
    color: var(--td-color-info);

    code {
      display: inline-block;
      padding: 4px 8px;
      margin-top: 4px;
      background: var(--td-bg-page);
      border-radius: 4px;
      font-family: monospace;
      font-size: 12px;
      word-break: break-all;
    }
  }
}

.verify-code-input {
  :deep(.el-input__inner) {
    text-align: center;
    font-size: 24px;
    letter-spacing: 8px;
    font-family: monospace;
  }
}
</style>
