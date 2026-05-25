<template>
  <div class="project-roles-container">
    <!-- 页面头部 -->
    <TdPageHeader>
      <template #leading>
        <div class="page-header-icon">
          <el-icon :size="20"><UserFilled /></el-icon>
        </div>
      </template>
      <template #title>项目角色管理</template>
      <template #subtitle>{{ projectKey }} - 管理项目角色和成员分配</template>
      <template #actions>
        <el-button @click="$router.push('/projects')">
          <el-icon><Back /></el-icon>
          返回项目
        </el-button>
        <el-button type="primary" @click="handleCreateRole">
          <el-icon><Plus /></el-icon>
          创建角色
        </el-button>
      </template>
    </TdPageHeader>

    <!-- 角色列表 -->
    <el-card shadow="never" class="roles-card">
      <div v-loading="loading" class="roles-list">
        <div v-for="role in roles" :key="role.id" class="role-item">
          <div class="role-header">
            <div class="role-info">
              <div class="role-icon" :class="getRoleIconClass(role.role_key)">
                <el-icon><Avatar /></el-icon>
              </div>
              <div class="role-details">
                <div class="role-title-row">
                  <h3 class="role-name">{{ role.role_name }}</h3>
                  <el-tag v-if="role.is_system" size="small" type="info">系统角色</el-tag>
                </div>
                <p class="role-key">{{ role.role_key }}</p>
                <p v-if="role.description" class="role-desc">{{ role.description }}</p>
              </div>
            </div>
            <div class="role-actions">
              <el-button size="small" @click="handleViewMembers(role)">
                <el-icon><User /></el-icon>
                成员管理
              </el-button>
              <el-button v-if="!role.is_system" size="small" @click="handleEditRole(role)">
                <el-icon><Edit /></el-icon>
                编辑
              </el-button>
              <el-button v-if="!role.is_system" size="small" type="danger" @click="handleDeleteRole(role)">
                <el-icon><Delete /></el-icon>
                删除
              </el-button>
            </div>
          </div>
        </div>
        <el-empty v-if="!loading && roles.length === 0" description="暂无角色" />
      </div>
    </el-card>

    <!-- 创建/编辑角色对话框 -->
    <el-dialog v-model="roleDialogVisible" :title="isEditMode ? '编辑角色' : '创建角色'" width="500px" destroy-on-close>
      <el-form ref="roleFormRef" :model="roleForm" :rules="roleRules" label-position="top">
        <el-form-item label="角色标识" prop="role_key">
          <el-input v-model="roleForm.role_key" placeholder="如: developers, testers" :disabled="isEditMode" />
        </el-form-item>
        <el-form-item label="角色名称" prop="role_name">
          <el-input v-model="roleForm.role_name" placeholder="如: 开发人员, 测试人员" />
        </el-form-item>
        <el-form-item label="描述">
          <el-input v-model="roleForm.description" type="textarea" :rows="3" placeholder="角色描述" />
        </el-form-item>
        <el-form-item label="排序">
          <el-input-number v-model="roleForm.sort_order" :min="0" :max="999" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="roleDialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="submitLoading" @click="submitRole">确定</el-button>
      </template>
    </el-dialog>

    <!-- 成员管理对话框 -->
    <el-dialog v-model="memberDialogVisible" :title="`${currentRole?.role_name} - 成员管理`" width="600px" destroy-on-close>
      <div class="member-dialog-content">
        <div class="member-add-section">
          <el-select v-model="selectedUserId" placeholder="选择用户" filterable style="width: 300px">
            <el-option v-for="u in availableUsers" :key="u.id" :label="u.display_name" :value="u.id">
              <span>{{ u.display_name }}</span>
              <span style="color: var(--td-text-placeholder); margin-left: 8px">{{ u.username }}</span>
            </el-option>
          </el-select>
          <el-button type="primary" :loading="addMemberLoading" :disabled="!selectedUserId" @click="handleAddMember">
            <el-icon><Plus /></el-icon>
            添加成员
          </el-button>
        </div>
        <el-divider />
        <div v-loading="membersLoading" class="member-list">
          <div v-for="member in roleMembers" :key="member.id" class="member-item">
            <div class="member-info">
              <div class="member-avatar">{{ member.user?.display_name?.charAt(0) || '?' }}</div>
              <div class="member-details">
                <span class="member-name">{{ member.user?.display_name }}</span>
                <span class="member-username">@{{ member.user?.username }}</span>
              </div>
            </div>
            <el-button size="small" type="danger" text @click="handleRemoveMember(member)">
              <el-icon><Delete /></el-icon>
              移除
            </el-button>
          </div>
          <el-empty v-if="!membersLoading && roleMembers.length === 0" description="暂无成员" />
        </div>
      </div>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, computed, onMounted } from 'vue'
import { useRoute } from 'vue-router'
import { ElMessage, ElMessageBox, type FormInstance, type FormRules } from 'element-plus'
import { UserFilled, Plus, User, Edit, Delete, Avatar, Back } from '@element-plus/icons-vue'
import {
  getProjectRoles,
  createProjectRole,
  updateProjectRole,
  deleteProjectRole,
  getRoleMembers,
  addRoleMember,
  removeRoleMember,
} from '@/api/project'
import { getAllUsers } from '@/api/user'
import type { ProjectRole, ProjectRoleMember, CreateProjectRoleRequest } from '@/types/project'
import type { UserOption } from '@/types/user'

const route = useRoute()
const projectKey = computed(() => route.params.key as string)

const loading = ref(false)
const roles = ref<ProjectRole[]>([])
const users = ref<UserOption[]>([])

// 角色表单
const roleDialogVisible = ref(false)
const isEditMode = ref(false)
const editingRoleId = ref<number | null>(null)
const submitLoading = ref(false)
const roleFormRef = ref<FormInstance>()
const roleForm = reactive<CreateProjectRoleRequest>({
  role_key: '',
  role_name: '',
  description: '',
  sort_order: 0,
})
const roleRules: FormRules = {
  role_key: [
    { required: true, message: '请输入角色标识', trigger: 'blur' },
    { pattern: /^[a-z_]+$/, message: '只能包含小写字母和下划线', trigger: 'blur' },
  ],
  role_name: [{ required: true, message: '请输入角色名称', trigger: 'blur' }],
}

// 成员管理
const memberDialogVisible = ref(false)
const currentRole = ref<ProjectRole | null>(null)
const roleMembers = ref<ProjectRoleMember[]>([])
const membersLoading = ref(false)
const selectedUserId = ref<number | null>(null)
const addMemberLoading = ref(false)

const availableUsers = computed(() => {
  const memberIds = roleMembers.value.map((m) => m.user_id)
  return users.value.filter((u) => !memberIds.includes(u.id))
})

const loadRoles = async () => {
  loading.value = true
  try {
    const { data } = await getProjectRoles(projectKey.value)
    roles.value = data.data
  } catch {
    // ignored
  } finally {
    loading.value = false
  }
}

const loadUsers = async () => {
  try {
    const { data } = await getAllUsers()
    users.value = data.data
  } catch {
    // ignored
  }
}

const handleCreateRole = () => {
  isEditMode.value = false
  editingRoleId.value = null
  Object.assign(roleForm, { role_key: '', role_name: '', description: '', sort_order: 0 })
  roleDialogVisible.value = true
}

const handleEditRole = (role: ProjectRole) => {
  isEditMode.value = true
  editingRoleId.value = role.id
  Object.assign(roleForm, {
    role_key: role.role_key,
    role_name: role.role_name,
    description: role.description,
    sort_order: role.sort_order,
  })
  roleDialogVisible.value = true
}

const submitRole = async () => {
  if (!roleFormRef.value) return
  await roleFormRef.value.validate(async (valid) => {
    if (!valid) return
    submitLoading.value = true
    try {
      if (isEditMode.value && editingRoleId.value) {
        await updateProjectRole(projectKey.value, editingRoleId.value, {
          role_name: roleForm.role_name,
          description: roleForm.description,
          sort_order: roleForm.sort_order,
        })
        ElMessage.success('更新成功')
      } else {
        await createProjectRole(projectKey.value, roleForm)
        ElMessage.success('创建成功')
      }
      roleDialogVisible.value = false
      loadRoles()
    } catch {
      // ignored
    } finally {
      submitLoading.value = false
    }
  })
}

const handleDeleteRole = async (role: ProjectRole) => {
  try {
    await ElMessageBox.confirm(`确定要删除角色 "${role.role_name}" 吗？`, '删除确认', {
      type: 'warning',
    })
    await deleteProjectRole(projectKey.value, role.id)
    ElMessage.success('删除成功')
    loadRoles()
  } catch (error) {
    if (error !== 'cancel') {
      // ignored
    }
  }
}

const handleViewMembers = async (role: ProjectRole) => {
  currentRole.value = role
  selectedUserId.value = null
  memberDialogVisible.value = true
  await loadRoleMembers(role.id)
}

const loadRoleMembers = async (roleId: number) => {
  membersLoading.value = true
  try {
    const { data } = await getRoleMembers(projectKey.value, roleId)
    roleMembers.value = data.data
  } catch {
    // ignored
  } finally {
    membersLoading.value = false
  }
}

const handleAddMember = async () => {
  if (!selectedUserId.value || !currentRole.value) return
  addMemberLoading.value = true
  try {
    await addRoleMember(projectKey.value, currentRole.value.id, { user_id: selectedUserId.value })
    ElMessage.success('添加成功')
    selectedUserId.value = null
    await loadRoleMembers(currentRole.value.id)
  } catch {
    // ignored
  } finally {
    addMemberLoading.value = false
  }
}

const handleRemoveMember = async (member: ProjectRoleMember) => {
  if (!currentRole.value) return
  try {
    await ElMessageBox.confirm(`确定要移除成员 "${member.user?.display_name}" 吗？`, '移除确认', {
      type: 'warning',
    })
    await removeRoleMember(projectKey.value, currentRole.value.id, member.user_id)
    ElMessage.success('移除成功')
    await loadRoleMembers(currentRole.value.id)
  } catch (error) {
    if (error !== 'cancel') {
      // ignored
    }
  }
}

const getRoleIconClass = (roleKey: string) => {
  const classMap: Record<string, string> = {
    administrators: 'admin',
    developers: 'dev',
    testers: 'test',
    viewers: 'view',
  }
  return classMap[roleKey] || 'default'
}

onMounted(() => {
  loadRoles()
  loadUsers()
})
</script>

<style scoped lang="scss">
.project-roles-container {
  width: 100%;
}

// 页面头部 icon (TdPageHeader leading slot)
.page-header-icon {
  width: 40px;
  height: 40px;
  background: var(--td-tag-primary-bg);
  border-radius: var(--td-radius-md);
  display: flex;
  align-items: center;
  justify-content: center;
  color: var(--td-color-primary);
  flex-shrink: 0;
}

.roles-card {
  border-radius: 12px;

  :deep(.el-card__body) {
    padding: 24px;
  }
}

.roles-list {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.role-item {
  background: var(--td-bg-page);
  border-radius: 12px;
  padding: 20px;
  transition: box-shadow 150ms ease-out;

  &:hover {
    box-shadow: 0 4px 12px rgba(0, 0, 0, 0.08);
  }
}

.role-header {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
}

.role-info {
  display: flex;
  gap: 16px;
}

.role-icon {
  width: 48px;
  height: 48px;
  border-radius: 12px;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 24px;
  color: var(--td-text-white);
  background: var(--td-color-primary);

  &.admin {
    background: var(--td-color-warning);
  }
  &.dev {
    background: var(--td-color-primary);
  }
  &.test {
    background: var(--td-color-success);
  }
  &.view {
    background: var(--td-text-secondary);
  }
}

.role-details {
  .role-title-row {
    display: flex;
    align-items: center;
    gap: 8px;
    margin-bottom: 4px;
  }

  .role-name {
    font-size: 16px;
    font-weight: 600;
    color: var(--td-text-primary);
    margin: 0;
  }

  .role-key {
    font-size: 13px;
    color: var(--td-text-secondary);
    margin: 0 0 4px 0;
    font-family: monospace;
  }

  .role-desc {
    font-size: 13px;
    color: var(--td-text-placeholder);
    margin: 0;
  }
}

.role-actions {
  display: flex;
  gap: 8px;
}

// 成员管理对话框
.member-dialog-content {
  .member-add-section {
    display: flex;
    gap: 12px;
    align-items: center;
  }

  .member-list {
    min-height: 200px;
    max-height: 400px;
    overflow-y: auto;
  }

  .member-item {
    display: flex;
    justify-content: space-between;
    align-items: center;
    padding: 12px;
    border-radius: 8px;
    margin-bottom: 8px;
    background: var(--td-bg-page);

    &:hover {
      background: var(--td-bg-section);
    }
  }

  .member-info {
    display: flex;
    align-items: center;
    gap: 12px;
  }

  .member-avatar {
    width: 36px;
    height: 36px;
    border-radius: 8px;
    background: var(--td-color-primary);
    color: var(--td-text-white);
    display: flex;
    align-items: center;
    justify-content: center;
    font-size: 14px;
    font-weight: 600;
  }

  .member-details {
    display: flex;
    flex-direction: column;
  }

  .member-name {
    font-size: 14px;
    font-weight: 500;
    color: var(--td-text-primary);
  }

  .member-username {
    font-size: 12px;
    color: var(--td-text-placeholder);
  }
}
</style>
