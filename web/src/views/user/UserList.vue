<template>
  <div class="user-list-container">
    <!-- 页面头部 -->
    <div class="page-header">
      <div class="header-info">
        <div class="header-icon">
          <el-icon><UserFilled /></el-icon>
        </div>
        <div class="header-text">
          <h1 class="header-title">用户管理</h1>
          <p class="header-desc">管理系统用户账号、角色权限和访问控制</p>
        </div>
      </div>
      <el-button type="primary" @click="handleCreate">
        <el-icon><Plus /></el-icon>
        创建用户
      </el-button>
    </div>

    <!-- 统计卡片 -->
    <el-row :gutter="20" class="stats-row">
      <el-col :xs="12" :sm="6">
        <div class="stat-card total">
          <div class="stat-icon">
            <el-icon><User /></el-icon>
          </div>
          <div class="stat-content">
            <div class="stat-value">{{ stats.total }}</div>
            <div class="stat-label">总用户数</div>
          </div>
        </div>
      </el-col>
      <el-col :xs="12" :sm="6">
        <div class="stat-card active">
          <div class="stat-icon">
            <el-icon><CircleCheck /></el-icon>
          </div>
          <div class="stat-content">
            <div class="stat-value">{{ stats.active }}</div>
            <div class="stat-label">活跃用户</div>
          </div>
        </div>
      </el-col>
      <el-col :xs="12" :sm="6">
        <div class="stat-card admin">
          <div class="stat-icon">
            <el-icon><Avatar /></el-icon>
          </div>
          <div class="stat-content">
            <div class="stat-value">{{ stats.admin }}</div>
            <div class="stat-label">管理员</div>
          </div>
        </div>
      </el-col>
      <el-col :xs="12" :sm="6">
        <div class="stat-card disabled">
          <div class="stat-icon">
            <el-icon><CircleClose /></el-icon>
          </div>
          <div class="stat-content">
            <div class="stat-value">{{ stats.disabled }}</div>
            <div class="stat-label">已禁用</div>
          </div>
        </div>
      </el-col>
    </el-row>

    <!-- 搜索和筛选 -->
    <el-card shadow="never" class="filter-card">
      <div class="filter-content">
        <div class="filter-left">
          <el-input
            v-model="queryParams.keyword"
            placeholder="搜索用户名、显示名称或邮箱"
            clearable
            class="search-input"
            @clear="handleSearch"
            @keyup.enter="handleSearch"
          >
            <template #prefix>
              <el-icon><Search /></el-icon>
            </template>
          </el-input>
          <el-select
            v-model="queryParams.role"
            placeholder="角色"
            clearable
            class="filter-select"
            @change="handleSearch"
          >
            <el-option label="管理员" value="admin">
              <div class="option-content">
                <el-icon class="option-icon admin"><Avatar /></el-icon>
                <span>管理员</span>
              </div>
            </el-option>
            <el-option label="普通用户" value="user">
              <div class="option-content">
                <el-icon class="option-icon user"><User /></el-icon>
                <span>普通用户</span>
              </div>
            </el-option>
          </el-select>
          <el-select
            v-model="queryParams.status"
            placeholder="状态"
            clearable
            class="filter-select"
            @change="handleSearch"
          >
            <el-option label="启用" :value="1">
              <div class="option-content">
                <span class="status-dot active"></span>
                <span>启用</span>
              </div>
            </el-option>
            <el-option label="禁用" :value="0">
              <div class="option-content">
                <span class="status-dot disabled"></span>
                <span>禁用</span>
              </div>
            </el-option>
          </el-select>
        </div>
        <div class="filter-right">
          <el-button :icon="Refresh" @click="handleReset">重置</el-button>
        </div>
      </div>
    </el-card>

    <!-- 用户列表 -->
    <el-card shadow="never" class="table-card">
      <el-table
        v-loading="loading"
        :data="userList"
        style="width: 100%"
        :row-class-name="tableRowClassName"
      >
        <el-table-column prop="username" label="用户信息" min-width="240">
          <template #default="{ row }">
            <div class="user-info-cell">
              <div class="user-avatar" :class="row.role">
                {{ row.display_name?.charAt(0) || row.username.charAt(0) }}
              </div>
              <div class="user-details">
                <div class="user-name">
                  {{ row.display_name || row.username }}
                  <el-tag
                    v-if="row.role === 'admin'"
                    type="danger"
                    size="small"
                    effect="dark"
                    class="role-tag"
                  >
                    管理员
                  </el-tag>
                </div>
                <div class="user-meta">
                  <span class="username">@{{ row.username }}</span>
                  <span class="divider">·</span>
                  <span class="email">{{ row.email }}</span>
                </div>
              </div>
            </div>
          </template>
        </el-table-column>
        <el-table-column prop="status" label="状态" width="100" align="center">
          <template #default="{ row }">
            <div class="status-badge" :class="row.status === 1 ? 'active' : 'disabled'">
              <span class="status-dot"></span>
              <span class="status-text">{{ row.status === 1 ? '启用' : '禁用' }}</span>
            </div>
          </template>
        </el-table-column>
        <el-table-column prop="mfa_enabled" label="MFA" width="80" align="center">
          <template #default="{ row }">
            <el-tooltip :content="row.mfa_enabled ? 'MFA 已启用' : 'MFA 未启用'" placement="top">
              <el-icon :class="['mfa-icon', row.mfa_enabled ? 'enabled' : 'disabled']">
                <component :is="row.mfa_enabled ? CircleCheck : CircleClose" />
              </el-icon>
            </el-tooltip>
          </template>
        </el-table-column>
        <el-table-column prop="last_login_at" label="最后登录" width="180">
          <template #default="{ row }">
            <div class="time-cell" v-if="row.last_login_at">
              <el-icon><Clock /></el-icon>
              <span>{{ formatTime(row.last_login_at) }}</span>
            </div>
            <span v-else class="text-muted">从未登录</span>
          </template>
        </el-table-column>
        <el-table-column prop="created_at" label="创建时间" width="180">
          <template #default="{ row }">
            <div class="time-cell">
              <el-icon><Calendar /></el-icon>
              <span>{{ formatTime(row.created_at) }}</span>
            </div>
          </template>
        </el-table-column>
        <el-table-column label="操作" width="180" fixed="right" align="center">
          <template #default="{ row }">
            <div class="action-buttons">
              <el-tooltip content="编辑" placement="top">
                <el-button link type="primary" @click="handleEdit(row)">
                  <el-icon><Edit /></el-icon>
                </el-button>
              </el-tooltip>
              <el-tooltip content="重置密码" placement="top">
                <el-button link type="warning" @click="handleResetPassword(row)">
                  <el-icon><Key /></el-icon>
                </el-button>
              </el-tooltip>
              <el-tooltip :content="row.status === 1 ? '禁用' : '启用'" placement="top">
                <el-button
                  link
                  :type="row.status === 1 ? 'danger' : 'success'"
                  @click="handleToggleStatus(row)"
                >
                  <el-icon>
                    <component :is="row.status === 1 ? Lock : Unlock" />
                  </el-icon>
                </el-button>
              </el-tooltip>
            </div>
          </template>
        </el-table-column>
      </el-table>

      <div class="pagination-wrapper">
        <el-pagination
          v-model:current-page="queryParams.page"
          v-model:page-size="queryParams.page_size"
          :total="total"
          :page-sizes="[10, 20, 50]"
          layout="total, sizes, prev, pager, next, jumper"
          @size-change="loadUsers"
          @current-change="loadUsers"
        />
      </div>
    </el-card>

    <!-- 创建/编辑用户对话框 -->
    <el-dialog
      v-model="dialogVisible"
      :title="isEdit ? '编辑用户' : '创建用户'"
      width="520px"
      destroy-on-close
      class="user-dialog"
    >
      <el-form ref="formRef" :model="form" :rules="rules" label-position="top">
        <el-row :gutter="16">
          <el-col :span="12">
            <el-form-item label="用户名" prop="username">
              <el-input
                v-model="form.username"
                placeholder="请输入用户名"
                :disabled="isEdit"
              >
                <template #prefix>
                  <el-icon><User /></el-icon>
                </template>
              </el-input>
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="显示名称" prop="display_name">
              <el-input v-model="form.display_name" placeholder="请输入显示名称">
                <template #prefix>
                  <el-icon><Postcard /></el-icon>
                </template>
              </el-input>
            </el-form-item>
          </el-col>
        </el-row>
        <el-form-item label="邮箱" prop="email">
          <el-input v-model="form.email" placeholder="请输入邮箱">
            <template #prefix>
              <el-icon><Message /></el-icon>
            </template>
          </el-input>
        </el-form-item>
        <el-form-item v-if="!isEdit" label="密码" prop="password">
          <el-input v-model="form.password" type="password" placeholder="请输入密码" show-password>
            <template #prefix>
              <el-icon><Lock /></el-icon>
            </template>
          </el-input>
        </el-form-item>
        <el-form-item label="角色" prop="role">
          <el-radio-group v-model="form.role" class="role-radio-group">
            <el-radio value="user" border>
              <div class="role-option">
                <el-icon class="role-icon user"><User /></el-icon>
                <div class="role-text">
                  <span class="role-name">普通用户</span>
                  <span class="role-desc">可以使用基本功能</span>
                </div>
              </div>
            </el-radio>
            <el-radio value="admin" border>
              <div class="role-option">
                <el-icon class="role-icon admin"><Avatar /></el-icon>
                <div class="role-text">
                  <span class="role-name">管理员</span>
                  <span class="role-desc">拥有全部权限</span>
                </div>
              </div>
            </el-radio>
          </el-radio-group>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="submitLoading" @click="submitForm">
          <el-icon><Check /></el-icon>
          {{ isEdit ? '保存修改' : '创建用户' }}
        </el-button>
      </template>
    </el-dialog>

    <!-- 重置密码对话框 -->
    <el-dialog
      v-model="resetPasswordVisible"
      title="重置密码"
      width="400px"
      destroy-on-close
      class="password-dialog"
    >
      <div class="password-dialog-header">
        <div class="password-icon">
          <el-icon><Key /></el-icon>
        </div>
        <p class="password-tip">请为用户设置新的登录密码</p>
      </div>
      <el-form ref="resetFormRef" :model="resetForm" :rules="resetRules" label-position="top">
        <el-form-item label="新密码" prop="password">
          <el-input
            v-model="resetForm.password"
            type="password"
            placeholder="请输入新密码"
            show-password
          >
            <template #prefix>
              <el-icon><Lock /></el-icon>
            </template>
          </el-input>
        </el-form-item>
        <el-form-item label="确认密码" prop="confirmPassword">
          <el-input
            v-model="resetForm.confirmPassword"
            type="password"
            placeholder="请再次输入密码"
            show-password
          >
            <template #prefix>
              <el-icon><Lock /></el-icon>
            </template>
          </el-input>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="resetPasswordVisible = false">取消</el-button>
        <el-button type="primary" :loading="resetLoading" @click="submitResetPassword">
          <el-icon><Check /></el-icon>
          确认重置
        </el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, computed, onMounted } from 'vue'
import { ElMessage, ElMessageBox, type FormInstance, type FormRules } from 'element-plus'
import {
  Search, Refresh, Plus, User, UserFilled, Avatar, Edit, Key, Lock, Unlock,
  CircleCheck, CircleClose, Clock, Calendar, Check, Postcard, Message
} from '@element-plus/icons-vue'
import { getUserList, createUser, updateUser, enableUser, disableUser, resetUserPassword } from '@/api/user'
import type { User as UserType, CreateUserRequest } from '@/types/user'
import dayjs from 'dayjs'

// 数据
const loading = ref(false)
const userList = ref<UserType[]>([])
const total = ref(0)

// 统计数据
const stats = computed(() => {
  const users = userList.value
  return {
    total: total.value,
    active: users.filter(u => u.status === 1).length,
    admin: users.filter(u => u.role === 'admin').length,
    disabled: users.filter(u => u.status === 0).length,
  }
})

// 查询参数
const queryParams = reactive({
  page: 1,
  page_size: 20,
  keyword: '',
  role: undefined as string | undefined,
  status: undefined as number | undefined,
})

// 创建/编辑对话框
const dialogVisible = ref(false)
const isEdit = ref(false)
const submitLoading = ref(false)
const formRef = ref<FormInstance>()
const editingUser = ref<UserType | null>(null)

const form = reactive<CreateUserRequest>({
  username: '',
  email: '',
  password: '',
  display_name: '',
  role: 'user',
})

const rules: FormRules = {
  username: [
    { required: true, message: '请输入用户名', trigger: 'blur' },
    { min: 3, max: 20, message: '长度为 3-20 个字符', trigger: 'blur' },
  ],
  email: [
    { required: true, message: '请输入邮箱', trigger: 'blur' },
    { type: 'email', message: '请输入有效的邮箱地址', trigger: 'blur' },
  ],
  password: [
    { required: true, message: '请输入密码', trigger: 'blur' },
    { min: 6, message: '密码至少 6 个字符', trigger: 'blur' },
  ],
  display_name: [
    { required: true, message: '请输入显示名称', trigger: 'blur' },
  ],
  role: [
    { required: true, message: '请选择角色', trigger: 'change' },
  ],
}

// 重置密码对话框
const resetPasswordVisible = ref(false)
const resetLoading = ref(false)
const resetFormRef = ref<FormInstance>()
const resetUserId = ref<number>(0)

const resetForm = reactive({
  password: '',
  confirmPassword: '',
})

const resetRules: FormRules = {
  password: [
    { required: true, message: '请输入新密码', trigger: 'blur' },
    { min: 6, message: '密码至少 6 个字符', trigger: 'blur' },
  ],
  confirmPassword: [
    { required: true, message: '请确认密码', trigger: 'blur' },
    {
      validator: (_rule, value, callback) => {
        if (value !== resetForm.password) {
          callback(new Error('两次输入的密码不一致'))
        } else {
          callback()
        }
      },
      trigger: 'blur',
    },
  ],
}

// 表格行样式
const tableRowClassName = ({ row }: { row: UserType }) => {
  if (row.status === 0) {
    return 'disabled-row'
  }
  return ''
}

// 加载用户列表
const loadUsers = async () => {
  loading.value = true
  try {
    const { data } = await getUserList(queryParams)
    userList.value = data.data.items
    total.value = data.data.total
  } catch (error) {
    console.error('Failed to load users:', error)
  } finally {
    loading.value = false
  }
}

// 搜索
const handleSearch = () => {
  queryParams.page = 1
  loadUsers()
}

// 重置
const handleReset = () => {
  queryParams.page = 1
  queryParams.keyword = ''
  queryParams.role = undefined
  queryParams.status = undefined
  loadUsers()
}

// 创建用户
const handleCreate = () => {
  isEdit.value = false
  editingUser.value = null
  Object.assign(form, {
    username: '',
    email: '',
    password: '',
    display_name: '',
    role: 'user',
  })
  dialogVisible.value = true
}

// 编辑用户
const handleEdit = (user: UserType) => {
  isEdit.value = true
  editingUser.value = user
  Object.assign(form, {
    username: user.username,
    email: user.email,
    password: '',
    display_name: user.display_name,
    role: user.role,
  })
  dialogVisible.value = true
}

// 提交表单
const submitForm = async () => {
  if (!formRef.value) return

  await formRef.value.validate(async (valid) => {
    if (!valid) return

    submitLoading.value = true
    try {
      if (isEdit.value && editingUser.value) {
        await updateUser(editingUser.value.id, {
          email: form.email,
          display_name: form.display_name,
          role: form.role,
        })
        ElMessage.success('更新成功')
      } else {
        await createUser(form)
        ElMessage.success('创建成功')
      }
      dialogVisible.value = false
      loadUsers()
    } catch (error) {
      console.error('Failed to submit:', error)
    } finally {
      submitLoading.value = false
    }
  })
}

// 重置密码
const handleResetPassword = (user: UserType) => {
  resetUserId.value = user.id
  resetForm.password = ''
  resetForm.confirmPassword = ''
  resetPasswordVisible.value = true
}

// 提交重置密码
const submitResetPassword = async () => {
  if (!resetFormRef.value) return

  await resetFormRef.value.validate(async (valid) => {
    if (!valid) return

    resetLoading.value = true
    try {
      await resetUserPassword(resetUserId.value, resetForm.password)
      ElMessage.success('密码重置成功')
      resetPasswordVisible.value = false
    } catch (error) {
      console.error('Failed to reset password:', error)
    } finally {
      resetLoading.value = false
    }
  })
}

// 切换状态
const handleToggleStatus = async (user: UserType) => {
  const action = user.status === 1 ? '禁用' : '启用'
  try {
    await ElMessageBox.confirm(
      `确定要${action}用户 "${user.display_name}" 吗？`,
      '提示',
      { type: 'warning' }
    )

    if (user.status === 1) {
      await disableUser(user.id)
    } else {
      await enableUser(user.id)
    }
    ElMessage.success(`${action}成功`)
    loadUsers()
  } catch (error) {
    if (error !== 'cancel') {
      console.error('Failed to toggle status:', error)
    }
  }
}

// 工具函数
const formatTime = (time: string) => {
  return dayjs(time).format('YYYY-MM-DD HH:mm')
}

// 初始化
onMounted(() => {
  loadUsers()
})
</script>

<style scoped lang="scss">
.user-list-container {
  width: 100%;
}

// 页面头部
.page-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 24px;
  padding: 24px;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  border-radius: 12px;
  color: #fff;

  .header-info {
    display: flex;
    align-items: center;
    gap: 16px;
  }

  .header-icon {
    width: 56px;
    height: 56px;
    background: rgba(255, 255, 255, 0.2);
    border-radius: 14px;
    display: flex;
    align-items: center;
    justify-content: center;
    font-size: 28px;
  }

  .header-text {
    .header-title {
      font-size: 22px;
      font-weight: 600;
      margin: 0 0 4px 0;
    }

    .header-desc {
      font-size: 14px;
      margin: 0;
      opacity: 0.9;
    }
  }

  :deep(.el-button) {
    background: rgba(255, 255, 255, 0.2);
    border: 1px solid rgba(255, 255, 255, 0.3);
    color: #fff;

    &:hover {
      background: rgba(255, 255, 255, 0.3);
    }
  }
}

// 统计卡片
.stats-row {
  margin-bottom: 20px;
}

.stat-card {
  display: flex;
  align-items: center;
  gap: 16px;
  padding: 20px;
  background: #fff;
  border-radius: 12px;
  box-shadow: 0 2px 12px rgba(0, 0, 0, 0.04);
  transition: all 0.3s;

  &:hover {
    transform: translateY(-2px);
    box-shadow: 0 4px 16px rgba(0, 0, 0, 0.08);
  }

  .stat-icon {
    width: 48px;
    height: 48px;
    border-radius: 12px;
    display: flex;
    align-items: center;
    justify-content: center;
    font-size: 24px;
  }

  .stat-content {
    .stat-value {
      font-size: 28px;
      font-weight: 700;
      line-height: 1.2;
    }

    .stat-label {
      font-size: 13px;
      color: #909399;
      margin-top: 2px;
    }
  }

  &.total {
    .stat-icon {
      background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
      color: #fff;
    }
    .stat-value {
      color: #667eea;
    }
  }

  &.active {
    .stat-icon {
      background: linear-gradient(135deg, #11998e 0%, #38ef7d 100%);
      color: #fff;
    }
    .stat-value {
      color: #11998e;
    }
  }

  &.admin {
    .stat-icon {
      background: linear-gradient(135deg, #f093fb 0%, #f5576c 100%);
      color: #fff;
    }
    .stat-value {
      color: #f5576c;
    }
  }

  &.disabled {
    .stat-icon {
      background: linear-gradient(135deg, #bdc3c7 0%, #95a5a6 100%);
      color: #fff;
    }
    .stat-value {
      color: #95a5a6;
    }
  }
}

// 筛选卡片
.filter-card {
  margin-bottom: 20px;
  border-radius: 12px;

  :deep(.el-card__body) {
    padding: 16px 20px;
  }

  .filter-content {
    display: flex;
    justify-content: space-between;
    align-items: center;
    flex-wrap: wrap;
    gap: 12px;
  }

  .filter-left {
    display: flex;
    align-items: center;
    gap: 12px;
    flex-wrap: wrap;
  }

  .search-input {
    width: 280px;
  }

  .filter-select {
    width: 130px;
  }
}

.option-content {
  display: flex;
  align-items: center;
  gap: 8px;

  .option-icon {
    font-size: 16px;

    &.admin {
      color: #f5576c;
    }

    &.user {
      color: #667eea;
    }
  }
}

.status-dot {
  width: 8px;
  height: 8px;
  border-radius: 50%;

  &.active {
    background: #67c23a;
  }

  &.disabled {
    background: #909399;
  }
}

// 表格卡片
.table-card {
  border-radius: 12px;

  :deep(.el-card__body) {
    padding: 0;
  }

  :deep(.el-table) {
    border-radius: 12px;

    th.el-table__cell {
      background: #f8fafc;
      font-weight: 600;
      color: #374151;
    }

    .disabled-row {
      background: #fafafa;

      td {
        color: #9ca3af;
      }
    }
  }
}

// 用户信息单元格
.user-info-cell {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 8px 0;

  .user-avatar {
    width: 40px;
    height: 40px;
    border-radius: 10px;
    display: flex;
    align-items: center;
    justify-content: center;
    font-size: 16px;
    font-weight: 600;
    color: #fff;
    flex-shrink: 0;

    &.admin {
      background: linear-gradient(135deg, #f093fb 0%, #f5576c 100%);
    }

    &.user {
      background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
    }
  }

  .user-details {
    min-width: 0;

    .user-name {
      font-size: 14px;
      font-weight: 600;
      color: #1f2937;
      display: flex;
      align-items: center;
      gap: 8px;

      .role-tag {
        font-size: 10px;
        padding: 0 6px;
        height: 18px;
        line-height: 18px;
      }
    }

    .user-meta {
      font-size: 12px;
      color: #9ca3af;
      margin-top: 2px;
      display: flex;
      align-items: center;
      gap: 6px;

      .divider {
        color: #d1d5db;
      }

      .email {
        max-width: 180px;
        overflow: hidden;
        text-overflow: ellipsis;
        white-space: nowrap;
      }
    }
  }
}

// 状态徽章
.status-badge {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  padding: 4px 10px;
  border-radius: 20px;
  font-size: 12px;

  .status-dot {
    width: 6px;
    height: 6px;
    border-radius: 50%;
  }

  &.active {
    background: #ecfdf5;
    color: #059669;

    .status-dot {
      background: #10b981;
    }
  }

  &.disabled {
    background: #f3f4f6;
    color: #6b7280;

    .status-dot {
      background: #9ca3af;
    }
  }
}

// MFA 图标
.mfa-icon {
  font-size: 18px;

  &.enabled {
    color: #10b981;
  }

  &.disabled {
    color: #d1d5db;
  }
}

// 时间单元格
.time-cell {
  display: flex;
  align-items: center;
  gap: 6px;
  font-size: 13px;
  color: #6b7280;

  .el-icon {
    font-size: 14px;
    color: #9ca3af;
  }
}

.text-muted {
  color: #d1d5db;
  font-size: 13px;
}

// 操作按钮
.action-buttons {
  display: flex;
  justify-content: center;
  gap: 8px;

  .el-button {
    font-size: 16px;
    padding: 4px;

    &:hover {
      background: #f3f4f6;
      border-radius: 6px;
    }
  }
}

// 分页
.pagination-wrapper {
  padding: 20px;
  display: flex;
  justify-content: flex-end;
  border-top: 1px solid #f0f0f0;
}

// 用户对话框
.user-dialog {
  :deep(.el-dialog__body) {
    padding: 20px 24px;
  }
}

.role-radio-group {
  display: flex;
  gap: 16px;
  width: 100%;

  :deep(.el-radio) {
    flex: 1;
    height: auto;
    padding: 16px;
    margin-right: 0;

    &.is-bordered {
      border-radius: 10px;
    }

    &.is-checked {
      border-color: #667eea;
      background: #f8f7ff;
    }

    .el-radio__input {
      display: none;
    }

    .el-radio__label {
      padding-left: 0;
      width: 100%;
    }
  }
}

.role-option {
  display: flex;
  align-items: center;
  gap: 12px;

  .role-icon {
    width: 40px;
    height: 40px;
    border-radius: 10px;
    display: flex;
    align-items: center;
    justify-content: center;
    font-size: 20px;

    &.user {
      background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
      color: #fff;
    }

    &.admin {
      background: linear-gradient(135deg, #f093fb 0%, #f5576c 100%);
      color: #fff;
    }
  }

  .role-text {
    display: flex;
    flex-direction: column;
    gap: 2px;

    .role-name {
      font-size: 14px;
      font-weight: 600;
      color: #1f2937;
    }

    .role-desc {
      font-size: 12px;
      color: #9ca3af;
    }
  }
}

// 重置密码对话框
.password-dialog {
  .password-dialog-header {
    text-align: center;
    margin-bottom: 24px;

    .password-icon {
      width: 64px;
      height: 64px;
      margin: 0 auto 16px;
      background: linear-gradient(135deg, #fbbf24 0%, #f59e0b 100%);
      border-radius: 16px;
      display: flex;
      align-items: center;
      justify-content: center;
      font-size: 28px;
      color: #fff;
    }

    .password-tip {
      font-size: 14px;
      color: #6b7280;
      margin: 0;
    }
  }
}

// 响应式
@media (max-width: 768px) {
  .page-header {
    flex-direction: column;
    align-items: flex-start;
    gap: 16px;
  }

  .filter-card {
    .filter-content {
      flex-direction: column;
      align-items: stretch;
    }

    .filter-left {
      flex-direction: column;
    }

    .search-input,
    .filter-select {
      width: 100%;
    }
  }

  .stat-card {
    padding: 16px;

    .stat-content {
      .stat-value {
        font-size: 22px;
      }
    }
  }

  .role-radio-group {
    flex-direction: column;
  }
}
</style>
