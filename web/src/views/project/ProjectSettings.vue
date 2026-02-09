<template>
  <div class="project-settings-container">
    <!-- 页面头部 -->
    <div class="page-header">
      <div class="header-left">
        <el-button @click="$router.push('/projects')" class="back-btn" circle>
          <el-icon><ArrowLeft /></el-icon>
        </el-button>
        <div class="header-info">
          <div class="header-icon">
            <el-icon><Setting /></el-icon>
          </div>
          <div class="header-text">
            <h1 class="header-title">项目设置</h1>
            <p class="header-desc">{{ projectKey }} - 配置项目信息和权限</p>
          </div>
        </div>
      </div>
    </div>

    <!-- 标签页 -->
    <el-card shadow="never" class="settings-card">
      <el-tabs v-model="activeTab" class="settings-tabs">
        <!-- 基本信息 -->
        <el-tab-pane name="basic">
          <template #label>
            <span class="tab-label">
              <el-icon><Document /></el-icon>
              基本信息
            </span>
          </template>
          <div v-loading="loading" class="tab-content">
            <el-form ref="basicFormRef" :model="basicForm" :rules="basicRules" label-width="120px" class="basic-form">
              <el-form-item label="项目标识">
                <el-input v-model="basicForm.project_key" disabled>
                  <template #prefix>
                    <el-icon><Key /></el-icon>
                  </template>
                </el-input>
                <div class="form-tip">
                  <el-icon><InfoFilled /></el-icon>
                  <span>项目标识创建后不可修改</span>
                </div>
              </el-form-item>
              <el-form-item label="项目名称" prop="name">
                <el-input v-model="basicForm.name" placeholder="请输入项目名称" maxlength="50" show-word-limit>
                  <template #prefix>
                    <el-icon><Folder /></el-icon>
                  </template>
                </el-input>
              </el-form-item>
              <el-form-item label="项目描述">
                <el-input
                  v-model="basicForm.description"
                  type="textarea"
                  :rows="4"
                  placeholder="请输入项目描述"
                  maxlength="500"
                  show-word-limit
                />
              </el-form-item>
              <el-form-item label="项目负责人">
                <el-select v-model="basicForm.lead_user_id" placeholder="请选择负责人" filterable style="width: 100%">
                  <template #prefix>
                    <el-icon><User /></el-icon>
                  </template>
                  <el-option v-for="u in users" :key="u.id" :label="u.display_name" :value="u.id">
                    <div class="user-option">
                      <div class="user-option-avatar">{{ u.display_name?.charAt(0) || '?' }}</div>
                      <div class="user-option-info">
                        <span class="user-option-name">{{ u.display_name }}</span>
                        <span class="user-option-username">@{{ u.username }}</span>
                      </div>
                    </div>
                  </el-option>
                </el-select>
              </el-form-item>
              <el-form-item label="项目状态">
                <div class="status-field">
                  <el-switch
                    v-model="basicForm.status"
                    :active-value="1"
                    :inactive-value="0"
                    active-text="启用"
                    inactive-text="禁用"
                    active-color="#10b981"
                    inactive-color="#9ca3af"
                  />
                  <div class="form-tip">
                    <el-icon><InfoFilled /></el-icon>
                    <span>禁用后项目将不可访问</span>
                  </div>
                </div>
              </el-form-item>
              <div class="form-actions">
                <el-button type="primary" :loading="saveLoading" @click="saveBasicInfo" size="large">
                  <el-icon><Check /></el-icon>
                  保存更改
                </el-button>
              </div>
            </el-form>
          </div>
        </el-tab-pane>

        <!-- 项目成员 -->
        <el-tab-pane name="members">
          <template #label>
            <span class="tab-label">
              <el-icon><User /></el-icon>
              项目成员
            </span>
          </template>
          <div v-loading="membersLoading" class="tab-content">
            <div class="section-header">
              <div class="section-title">
                <el-icon><UserFilled /></el-icon>
                <span>成员列表</span>
                <el-tag size="small" type="info" style="margin-left: 8px">{{ members.length }} 人</el-tag>
              </div>
              <el-button type="primary" @click="openMemberDialog">
                <el-icon><Plus /></el-icon>
                添加成员
              </el-button>
            </div>
            <el-table :data="members" stripe class="members-table">
              <el-table-column label="成员" min-width="200">
                <template #default="{ row }">
                  <div class="member-cell">
                    <div class="member-avatar" :class="row.role">
                      {{ row.user?.display_name?.charAt(0) || '?' }}
                    </div>
                    <div class="member-info">
                      <div class="member-name">{{ row.user?.display_name }}</div>
                      <div class="member-username">@{{ row.user?.username }}</div>
                    </div>
                  </div>
                </template>
              </el-table-column>
              <el-table-column label="角色" width="120">
                <template #default="{ row }">
                  <el-tag :type="getRoleType(row.role)" size="small">
                    {{ getRoleText(row.role) }}
                  </el-tag>
                </template>
              </el-table-column>
              <el-table-column label="操作" width="200" align="right">
                <template #default="{ row }">
                  <el-dropdown trigger="click" @command="(cmd: string) => handleEditMemberRole(row, cmd)">
                    <el-button size="small" type="primary" text>
                      更改角色
                      <el-icon class="el-icon--right"><ArrowLeft style="transform: rotate(-90deg)" /></el-icon>
                    </el-button>
                    <template #dropdown>
                      <el-dropdown-menu>
                        <el-dropdown-item command="owner" :disabled="row.role === 'owner'">
                          <el-icon><Star /></el-icon>
                          所有者
                        </el-dropdown-item>
                        <el-dropdown-item command="admin" :disabled="row.role === 'admin'">
                          <el-icon><Key /></el-icon>
                          管理员
                        </el-dropdown-item>
                        <el-dropdown-item command="member" :disabled="row.role === 'member'">
                          <el-icon><User /></el-icon>
                          成员
                        </el-dropdown-item>
                      </el-dropdown-menu>
                    </template>
                  </el-dropdown>
                  <el-button size="small" type="danger" text @click="handleRemoveMember(row)">
                    移除
                  </el-button>
                </template>
              </el-table-column>
            </el-table>
            <el-empty v-if="!membersLoading && members.length === 0" description="暂无成员" />
          </div>
        </el-tab-pane>

        <!-- 项目角色 -->
        <el-tab-pane name="roles">
          <template #label>
            <span class="tab-label">
              <el-icon><Avatar /></el-icon>
              项目角色
            </span>
          </template>
          <div v-loading="rolesLoading" class="tab-content">
            <div class="section-header">
              <div class="section-title">
                <el-icon><Avatar /></el-icon>
                <span>角色列表</span>
                <el-tag size="small" type="info" style="margin-left: 8px">{{ roles.length }} 个</el-tag>
              </div>
              <el-button type="primary" @click="openRoleDialog">
                <el-icon><Plus /></el-icon>
                创建角色
              </el-button>
            </div>
            <div class="roles-grid">
              <div v-for="role in roles" :key="role.id" class="role-card">
                <div class="role-icon" :class="getRoleIconClass(role.role_key)">
                  <el-icon><Avatar /></el-icon>
                </div>
                <div class="role-info">
                  <div class="role-header">
                    <span class="role-name">{{ role.role_name }}</span>
                    <el-tag v-if="role.is_system" size="small" type="info">系统</el-tag>
                  </div>
                  <div class="role-key">{{ role.role_key }}</div>
                  <div v-if="role.description" class="role-desc">{{ role.description }}</div>
                </div>
                <div class="role-actions">
                  <el-button size="small" @click="handleManageRoleMembers(role)">
                    <el-icon><User /></el-icon>
                    成员
                  </el-button>
                  <el-button v-if="!role.is_system" size="small" @click="handleEditRole(role)">
                    <el-icon><Edit /></el-icon>
                  </el-button>
                  <el-button v-if="!role.is_system" size="small" type="danger" @click="handleDeleteRole(role)">
                    <el-icon><Delete /></el-icon>
                  </el-button>
                </div>
              </div>
            </div>
            <el-empty v-if="!rolesLoading && roles.length === 0" description="暂无角色" />
          </div>
        </el-tab-pane>

        <!-- 工单类型 -->
        <el-tab-pane name="issue-types">
          <template #label>
            <span class="tab-label">
              <el-icon><Tickets /></el-icon>
              工单类型
            </span>
          </template>
          <div v-loading="issueTypesLoading" class="tab-content">
            <div class="section-header">
              <div class="section-title">
                <el-icon><Tickets /></el-icon>
                <span>类型列表</span>
                <el-tag size="small" type="info" style="margin-left: 8px">{{ issueTypes.length }} 个</el-tag>
              </div>
              <el-button type="primary" @click="openIssueTypeDialog">
                <el-icon><Plus /></el-icon>
                添加类型
              </el-button>
            </div>
            <div class="issue-types-grid">
              <div v-for="type in issueTypes" :key="type.id" class="issue-type-card">
                <div class="type-icon" :style="{ background: type.color || '#3b82f6' }">
                  <el-icon :size="24">
                    <component :is="getIssueTypeIcon(type.icon)" />
                  </el-icon>
                </div>
                <div class="type-info">
                  <div class="type-name">{{ type.display_name }}</div>
                  <div class="type-key">{{ type.name }}</div>
                  <div v-if="type.description" class="type-desc">{{ type.description }}</div>
                </div>
                <div class="type-actions">
                  <el-button size="small" @click="handleEditIssueTypeClick(type)">
                    <el-icon><Edit /></el-icon>
                  </el-button>
                  <el-button size="small" type="danger" @click="handleDeleteIssueType(type)">
                    <el-icon><Delete /></el-icon>
                  </el-button>
                </div>
              </div>
            </div>
            <el-empty v-if="!issueTypesLoading && issueTypes.length === 0" description="暂无工单类型" />
          </div>
        </el-tab-pane>

        <!-- 字段配置 -->
        <el-tab-pane name="fields">
          <template #label>
            <span class="tab-label">
              <el-icon><Grid /></el-icon>
              字段配置
            </span>
          </template>
          <div class="tab-content">
            <FieldConfigTab :project-key="projectKey" />
          </div>
        </el-tab-pane>
      </el-tabs>
    </el-card>

    <!-- 添加成员对话框 -->
    <el-dialog v-model="memberDialogVisible" title="添加成员" width="500px" destroy-on-close class="custom-dialog">
      <el-form ref="memberFormRef" :model="memberForm" :rules="memberRules" label-width="80px">
        <el-form-item label="用户" prop="user_id">
          <el-select
            v-model="memberForm.user_id"
            placeholder="请选择用户"
            filterable
            style="width: 100%"
            popper-class="user-select-popper"
          >
            <el-option v-for="u in availableUsers" :key="u.id" :label="u.display_name" :value="u.id">
              <div class="user-option">
                <div class="user-option-avatar">{{ u.display_name?.charAt(0) || '?' }}</div>
                <div class="user-option-info">
                  <span class="user-option-name">{{ u.display_name }}</span>
                  <span class="user-option-username">@{{ u.username }}</span>
                </div>
              </div>
            </el-option>
          </el-select>
        </el-form-item>
        <el-form-item label="角色" prop="role">
          <el-select v-model="memberForm.role" placeholder="请选择角色" style="width: 100%">
            <el-option label="所有者" value="owner">
              <div class="role-option">
                <el-icon><Star /></el-icon>
                <span>所有者</span>
              </div>
            </el-option>
            <el-option label="管理员" value="admin">
              <div class="role-option">
                <el-icon><Key /></el-icon>
                <span>管理员</span>
              </div>
            </el-option>
            <el-option label="成员" value="member">
              <div class="role-option">
                <el-icon><User /></el-icon>
                <span>成员</span>
              </div>
            </el-option>
          </el-select>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="memberDialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="addMemberLoading" @click="submitAddMember">
          <el-icon><Check /></el-icon>
          确定
        </el-button>
      </template>
    </el-dialog>

    <!-- 创建/编辑角色对话框 -->
    <el-dialog
      v-model="roleDialogVisible"
      :title="isEditingRole ? '编辑角色' : '创建角色'"
      width="500px"
      destroy-on-close
      class="custom-dialog"
    >
      <el-form ref="roleFormRef" :model="roleForm" :rules="roleRules" label-width="80px">
        <el-form-item label="角色标识" prop="role_key">
          <el-input v-model="roleForm.role_key" placeholder="如: developers" :disabled="isEditingRole" />
          <div class="form-tip">小写字母和下划线，创建后不可修改</div>
        </el-form-item>
        <el-form-item label="角色名称" prop="role_name">
          <el-input v-model="roleForm.role_name" placeholder="如: 开发人员" />
        </el-form-item>
        <el-form-item label="描述">
          <el-input v-model="roleForm.description" type="textarea" :rows="3" placeholder="角色描述" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="roleDialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="roleSubmitLoading" @click="submitRole">
          <el-icon><Check /></el-icon>
          确定
        </el-button>
      </template>
    </el-dialog>

    <!-- 创建/编辑工单类型对话框 -->
    <el-dialog
      v-model="issueTypeDialogVisible"
      :title="isEditingIssueType ? '编辑工单类型' : '创建工单类型'"
      width="560px"
      destroy-on-close
      class="custom-dialog"
    >
      <el-form ref="issueTypeFormRef" :model="issueTypeForm" :rules="issueTypeRules" label-width="80px">
        <el-form-item label="类型标识" prop="name">
          <el-input v-model="issueTypeForm.name" placeholder="如: epic, bug, task" :disabled="isEditingIssueType" />
          <div class="form-tip-small">
            <el-icon><InfoFilled /></el-icon>
            <span>内部标识，只能包含小写字母，创建后不可修改。例如: epic, bug, task</span>
          </div>
        </el-form-item>
        <el-form-item label="显示名称" prop="display_name">
          <el-input v-model="issueTypeForm.display_name" placeholder="如: Epic, 缺陷, 任务" />
          <div class="form-tip-small">用户看到的名称，可以使用任意字符</div>
        </el-form-item>
        <el-form-item label="图标">
          <div class="icon-selector">
            <div
              v-for="iconItem in availableIcons"
              :key="iconItem.name"
              class="icon-option"
              :class="{ active: issueTypeForm.icon === iconItem.name }"
              @click="issueTypeForm.icon = iconItem.name"
            >
              <el-icon :size="20">
                <component :is="iconItem.icon" />
              </el-icon>
              <span class="icon-label">{{ iconItem.label }}</span>
            </div>
          </div>
        </el-form-item>
        <el-form-item label="颜色">
          <div class="color-selector">
            <el-color-picker v-model="issueTypeForm.color" />
            <div class="color-presets">
              <div
                v-for="color in ['#3b82f6', '#10b981', '#f59e0b', '#ef4444', '#8b5cf6', '#ec4899', '#6b7280', '#06b6d4']"
                :key="color"
                class="color-preset"
                :style="{ background: color }"
                :class="{ active: issueTypeForm.color === color }"
                @click="issueTypeForm.color = color"
              />
            </div>
          </div>
        </el-form-item>
        <el-form-item label="描述">
          <el-input v-model="issueTypeForm.description" type="textarea" :rows="3" placeholder="类型描述" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="issueTypeDialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="issueTypeSubmitLoading" @click="submitIssueType">
          <el-icon><Check /></el-icon>
          确定
        </el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, computed, onMounted, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage, ElMessageBox, type FormInstance, type FormRules } from 'element-plus'
import {
  Setting,
  Plus,
  User,
  UserFilled,
  Avatar,
  ArrowLeft,
  Document,
  InfoFilled,
  Key,
  Folder,
  Check,
  Star,
  Delete,
  Edit,
  Tickets,
  Lightning,
  CircleCheck,
  Warning,
  Flag,
  Promotion,
  Tools,
  Connection,
  Memo,
  Grid,
} from '@element-plus/icons-vue'
import FieldConfigTab from './components/FieldConfigTab.vue'
import {
  getProjectDetail,
  updateProject,
  getProjectMembers,
  addProjectMember,
  updateProjectMemberRole,
  removeProjectMember,
  getProjectRoles,
  createProjectRole,
  updateProjectRole,
  deleteProjectRole,
  getProjectIssueTypes,
  createProjectIssueType,
  updateProjectIssueType,
  deleteProjectIssueType,
} from '@/api/project'
import { getAllUsers } from '@/api/user'
import type {
  ProjectMember,
  ProjectRole,
  ProjectIssueType,
  UpdateProjectRequest,
  CreateProjectRoleRequest,
  CreateIssueTypeRequest,
} from '@/types/project'
import type { UserOption } from '@/types/user'

const route = useRoute()
const router = useRouter()
const projectKey = computed(() => route.params.key as string)

// 从 URL 查询参数中获取 tab，如果没有则默认为 'basic'
const activeTab = ref((route.query.tab as string) || 'basic')
const loading = ref(false)
const saveLoading = ref(false)
const users = ref<UserOption[]>([])

// 监听 tab 变化，更新 URL
watch(activeTab, (newTab) => {
  router.replace({
    query: { ...route.query, tab: newTab },
  })
})

// 基本信息
const basicFormRef = ref<FormInstance>()
const basicForm = reactive<UpdateProjectRequest & { project_key?: string }>({
  project_key: '',
  name: '',
  description: '',
  lead_user_id: undefined,
  status: 1,
})
const basicRules: FormRules = {
  name: [{ required: true, message: '请输入项目名称', trigger: 'blur' }],
}

// 项目成员
const membersLoading = ref(false)
const members = ref<ProjectMember[]>([])
const memberDialogVisible = ref(false)
const addMemberLoading = ref(false)
const memberFormRef = ref<FormInstance>()
const memberForm = reactive({
  user_id: undefined as number | undefined,
  role: 'member',
})
const memberRules: FormRules = {
  user_id: [{ required: true, message: '请选择用户', trigger: 'change' }],
  role: [{ required: true, message: '请选择角色', trigger: 'change' }],
}

const availableUsers = computed(() => {
  const memberIds = members.value.map((m) => m.user_id)
  return users.value.filter((u) => !memberIds.includes(u.id))
})

// 项目角色
const rolesLoading = ref(false)
const roles = ref<ProjectRole[]>([])
const roleDialogVisible = ref(false)
const isEditingRole = ref(false)
const editingRoleId = ref<number | null>(null)
const roleSubmitLoading = ref(false)
const roleFormRef = ref<FormInstance>()
const roleForm = reactive<CreateProjectRoleRequest>({
  role_key: '',
  role_name: '',
  description: '',
})
const roleRules: FormRules = {
  role_key: [
    { required: true, message: '请输入角色标识', trigger: 'blur' },
    { pattern: /^[a-z_]+$/, message: '只能包含小写字母和下划线', trigger: 'blur' },
  ],
  role_name: [{ required: true, message: '请输入角色名称', trigger: 'blur' }],
}

// 工单类型
const issueTypesLoading = ref(false)
const issueTypes = ref<ProjectIssueType[]>([])
const issueTypeDialogVisible = ref(false)
const isEditingIssueType = ref(false)
const editingIssueTypeId = ref<number | null>(null)
const issueTypeSubmitLoading = ref(false)
const issueTypeFormRef = ref<FormInstance>()
const issueTypeForm = reactive<CreateIssueTypeRequest>({
  name: '',
  display_name: '',
  description: '',
  icon: 'task',
  color: '#3b82f6',
})
const issueTypeRules: FormRules = {
  name: [
    { required: true, message: '请输入类型标识', trigger: 'blur' },
    { pattern: /^[a-z]+$/, message: '只能包含小写字母', trigger: 'blur' },
  ],
  display_name: [{ required: true, message: '请输入显示名称', trigger: 'blur' }],
}

const loadProjectDetail = async () => {
  loading.value = true
  try {
    const { data } = await getProjectDetail(projectKey.value)
    const project = data.data
    Object.assign(basicForm, {
      project_key: project.project_key,
      name: project.name,
      description: project.description,
      lead_user_id: project.lead_user_id,
      status: project.status,
    })
  } catch (error) {
    console.error('Failed to load project:', error)
  } finally {
    loading.value = false
  }
}

const loadUsers = async () => {
  try {
    const { data } = await getAllUsers()
    users.value = data.data
  } catch (error) {
    console.error('Failed to load users:', error)
  }
}

const loadMembers = async () => {
  membersLoading.value = true
  try {
    const { data } = await getProjectMembers(projectKey.value)
    members.value = data.data
  } catch (error) {
    console.error('Failed to load members:', error)
  } finally {
    membersLoading.value = false
  }
}

const loadRoles = async () => {
  rolesLoading.value = true
  try {
    const { data } = await getProjectRoles(projectKey.value)
    roles.value = data.data
  } catch (error) {
    console.error('Failed to load roles:', error)
  } finally {
    rolesLoading.value = false
  }
}

const loadIssueTypes = async () => {
  issueTypesLoading.value = true
  try {
    const { data } = await getProjectIssueTypes(projectKey.value)
    issueTypes.value = data.data
  } catch (error) {
    console.error('Failed to load issue types:', error)
  } finally {
    issueTypesLoading.value = false
  }
}

const saveBasicInfo = async () => {
  if (!basicFormRef.value) return
  await basicFormRef.value.validate(async (valid) => {
    if (!valid) return
    saveLoading.value = true
    try {
      await updateProject(projectKey.value, {
        name: basicForm.name,
        description: basicForm.description,
        lead_user_id: basicForm.lead_user_id,
        status: basicForm.status,
      })
      ElMessage.success('保存成功')
    } catch (error) {
      console.error('Failed to save:', error)
    } finally {
      saveLoading.value = false
    }
  })
}

const submitAddMember = async () => {
  if (!memberFormRef.value) return
  await memberFormRef.value.validate(async (valid) => {
    if (!valid) return
    addMemberLoading.value = true
    try {
      await addProjectMember(projectKey.value, {
        user_id: memberForm.user_id!,
        role: memberForm.role,
      })
      ElMessage.success('添加成功')
      memberDialogVisible.value = false
      loadMembers()
    } catch (error) {
      console.error('Failed to add member:', error)
    } finally {
      addMemberLoading.value = false
    }
  })
}

const openMemberDialog = () => {
  memberForm.user_id = undefined
  memberForm.role = 'member'
  memberDialogVisible.value = true
}

const handleEditMemberRole = async (member: ProjectMember, newRole: string) => {
  try {
    await updateProjectMemberRole(projectKey.value, member.user_id, newRole)
    ElMessage.success('角色更新成功')
    loadMembers()
  } catch (error) {
    console.error('Failed to update member role:', error)
    ElMessage.error('角色更新失败')
  }
}

const handleRemoveMember = async (member: ProjectMember) => {
  try {
    await ElMessageBox.confirm(`确定要移除成员 "${member.user?.display_name}" 吗？`, '移除确认', {
      type: 'warning',
    })
    await removeProjectMember(projectKey.value, member.user_id)
    ElMessage.success('移除成功')
    loadMembers()
  } catch (error) {
    if (error !== 'cancel') {
      console.error('Failed to remove member:', error)
    }
  }
}

const openRoleDialog = () => {
  isEditingRole.value = false
  editingRoleId.value = null
  Object.assign(roleForm, {
    role_key: '',
    role_name: '',
    description: '',
  })
  roleDialogVisible.value = true
}

const handleEditRole = (role: ProjectRole) => {
  isEditingRole.value = true
  editingRoleId.value = role.id
  Object.assign(roleForm, {
    role_key: role.role_key,
    role_name: role.role_name,
    description: role.description,
  })
  roleDialogVisible.value = true
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
      console.error('Failed to delete role:', error)
      ElMessage.error('删除失败')
    }
  }
}

const handleManageRoleMembers = (role: ProjectRole) => {
  ElMessage.info(`管理角色 "${role.role_name}" 的成员功能开发中`)
}

const submitRole = async () => {
  if (!roleFormRef.value) return
  await roleFormRef.value.validate(async (valid) => {
    if (!valid) return
    roleSubmitLoading.value = true
    try {
      if (isEditingRole.value && editingRoleId.value) {
        await updateProjectRole(projectKey.value, editingRoleId.value, {
          role_name: roleForm.role_name,
          description: roleForm.description,
        })
        ElMessage.success('更新成功')
      } else {
        await createProjectRole(projectKey.value, roleForm)
        ElMessage.success('创建成功')
      }
      roleDialogVisible.value = false
      loadRoles()
    } catch (error) {
      console.error('Failed to submit role:', error)
      ElMessage.error(isEditingRole.value ? '更新失败' : '创建失败')
    } finally {
      roleSubmitLoading.value = false
    }
  })
}

const openIssueTypeDialog = () => {
  isEditingIssueType.value = false
  editingIssueTypeId.value = null
  Object.assign(issueTypeForm, {
    name: '',
    display_name: '',
    description: '',
    icon: 'task',
    color: '#3b82f6',
  })
  issueTypeDialogVisible.value = true
}

const handleEditIssueTypeClick = (issueType: ProjectIssueType) => {
  isEditingIssueType.value = true
  editingIssueTypeId.value = issueType.id
  Object.assign(issueTypeForm, {
    name: issueType.name,
    display_name: issueType.display_name,
    description: issueType.description,
    icon: issueType.icon,
    color: issueType.color,
  })
  issueTypeDialogVisible.value = true
}

const submitIssueType = async () => {
  if (!issueTypeFormRef.value) return
  await issueTypeFormRef.value.validate(async (valid) => {
    if (!valid) return
    issueTypeSubmitLoading.value = true
    try {
      if (isEditingIssueType.value && editingIssueTypeId.value) {
        await updateProjectIssueType(projectKey.value, editingIssueTypeId.value, {
          display_name: issueTypeForm.display_name,
          description: issueTypeForm.description,
          icon: issueTypeForm.icon,
          color: issueTypeForm.color,
        })
        ElMessage.success('更新成功')
      } else {
        await createProjectIssueType(projectKey.value, issueTypeForm)
        ElMessage.success('创建成功')
      }
      issueTypeDialogVisible.value = false
      loadIssueTypes()
    } catch (error) {
      console.error('Failed to submit issue type:', error)
      ElMessage.error(isEditingIssueType.value ? '更新失败' : '创建失败')
    } finally {
      issueTypeSubmitLoading.value = false
    }
  })
}

const handleDeleteIssueType = async (issueType: ProjectIssueType) => {
  try {
    await ElMessageBox.confirm(`确定要删除工单类型 "${issueType.display_name}" 吗？`, '删除确认', {
      type: 'warning',
    })
    await deleteProjectIssueType(projectKey.value, issueType.id)
    ElMessage.success('删除成功')
    loadIssueTypes()
  } catch (error) {
    if (error !== 'cancel') {
      console.error('Failed to delete issue type:', error)
    }
  }
}

const getRoleType = (role: string) => {
  const map: Record<string, any> = { owner: 'danger', admin: 'warning', member: 'info' }
  return map[role] || 'info'
}

const getRoleText = (role: string) => {
  const map: Record<string, string> = { owner: '所有者', admin: '管理员', member: '成员' }
  return map[role] || role
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

// 工单类型图标映射
const issueTypeIconMap: Record<string, any> = {
  epic: Lightning,
  story: Document,
  task: CircleCheck,
  bug: Warning,
  subtask: Memo,
  improvement: Tools,
  feature: Star,
  change: Connection,
  fault: Flag,
  alert: Promotion,
}

const getIssueTypeIcon = (iconName: string) => {
  return issueTypeIconMap[iconName] || Tickets
}

// 可选的图标列表
const availableIcons = [
  { name: 'epic', label: 'Epic', icon: Lightning },
  { name: 'story', label: 'Story', icon: Document },
  { name: 'task', label: 'Task', icon: CircleCheck },
  { name: 'bug', label: 'Bug', icon: Warning },
  { name: 'subtask', label: 'Subtask', icon: Memo },
  { name: 'improvement', label: 'Improvement', icon: Tools },
  { name: 'feature', label: 'Feature', icon: Star },
  { name: 'change', label: 'Change', icon: Connection },
  { name: 'fault', label: 'Fault', icon: Flag },
  { name: 'alert', label: 'Alert', icon: Promotion },
]

onMounted(() => {
  loadProjectDetail()
  loadUsers()
  loadMembers()
  loadRoles()
  loadIssueTypes()
})
</script>

<style scoped lang="scss">
.project-settings-container {
  width: 100%;
}

.page-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 24px;
  padding: 24px 32px;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  border-radius: 12px;
  color: #fff;

  .header-left {
    display: flex;
    align-items: center;
    gap: 16px;
  }

  .back-btn {
    background: rgba(255, 255, 255, 0.2);
    border: 1px solid rgba(255, 255, 255, 0.3);
    color: #fff;

    &:hover {
      background: rgba(255, 255, 255, 0.3);
    }
  }

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
}

.settings-card {
  border-radius: 12px;

  :deep(.el-card__body) {
    padding: 0;
  }
}

.settings-tabs {
  :deep(.el-tabs__header) {
    padding: 0 24px;
    margin: 0;
    background: #f9fafb;
  }

  :deep(.el-tabs__nav-wrap) {
    &::after {
      height: 1px;
      background-color: #e5e7eb;
    }
  }

  :deep(.el-tabs__item) {
    padding: 0 20px;
    height: 48px;
    line-height: 48px;
    font-size: 14px;

    &.is-active {
      color: #667eea;
      font-weight: 600;
    }
  }

  :deep(.el-tabs__active-bar) {
    background-color: #667eea;
    height: 3px;
  }

  :deep(.el-tabs__content) {
    padding: 0;
  }
}

.tab-label {
  display: flex;
  align-items: center;
  gap: 6px;
}

.tab-content {
  padding: 24px;
  min-height: 400px;
  background: #f9fafb;
}

.basic-form {
  background: #fff;
  padding: 32px;
  border-radius: 12px;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.05);

  :deep(.el-form-item__label) {
    font-weight: 500;
    color: #374151;
  }

  :deep(.el-input__wrapper) {
    border-radius: 8px;
  }

  :deep(.el-textarea__inner) {
    border-radius: 8px;
  }

  .status-field {
    display: flex;
    flex-direction: column;
    gap: 12px;
  }

  .form-tip {
    display: inline-flex;
    align-items: center;
    gap: 6px;
    margin-top: 8px;
    padding: 8px 12px;
    background: #f0f9ff;
    border-radius: 6px;
    font-size: 12px;
    color: #0369a1;
    line-height: 1.4;

    .el-icon {
      flex-shrink: 0;
      font-size: 14px;
    }
  }

  .form-actions {
    display: flex;
    justify-content: flex-end;
    gap: 12px;
    margin-top: 24px;
    padding-top: 24px;
    border-top: 1px solid #e5e7eb;
  }
}

.section-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 20px;

  h3 {
    font-size: 16px;
    font-weight: 600;
    color: #1f2937;
    margin: 0;
  }
}

// 成员表格
.members-table {
  background: #fff;
  border-radius: 12px;
  overflow: hidden;

  :deep(.el-table__header) {
    th {
      background: #f9fafb !important;
      font-weight: 600;
      color: #374151;
    }
  }

  :deep(.el-table__row) {
    td {
      padding: 16px 0;
    }
  }
}

.member-cell {
  display: flex;
  align-items: center;
  gap: 12px;
}

.member-avatar {
  width: 40px;
  height: 40px;
  border-radius: 10px;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  color: #fff;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 16px;
  font-weight: 600;
  flex-shrink: 0;

  &.owner {
    background: linear-gradient(135deg, #f59e0b 0%, #d97706 100%);
  }
  &.admin {
    background: linear-gradient(135deg, #3b82f6 0%, #2563eb 100%);
  }
  &.member {
    background: linear-gradient(135deg, #10b981 0%, #059669 100%);
  }
}

.member-info {
  .member-name {
    font-size: 14px;
    font-weight: 600;
    color: #1f2937;
    margin-bottom: 2px;
  }

  .member-username {
    font-size: 12px;
    color: #9ca3af;
  }
}

// 用户单元格
.user-cell {
  display: flex;
  align-items: center;
  gap: 12px;
}

.user-avatar {
  width: 36px;
  height: 36px;
  border-radius: 8px;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  color: #fff;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 14px;
  font-weight: 600;
}

.user-info {
  .user-name {
    font-size: 14px;
    font-weight: 500;
    color: #1f2937;
  }

  .user-username {
    font-size: 12px;
    color: #9ca3af;
  }
}

// 角色网格
.roles-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(320px, 1fr));
  gap: 16px;
}

.role-card {
  display: flex;
  gap: 16px;
  padding: 20px;
  background: #fff;
  border: 1px solid #e5e7eb;
  border-radius: 12px;
  transition: all 0.2s;

  &:hover {
    box-shadow: 0 4px 12px rgba(0, 0, 0, 0.08);
    border-color: #d1d5db;
  }
}

.role-icon {
  width: 48px;
  height: 48px;
  border-radius: 12px;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 24px;
  color: #fff;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  flex-shrink: 0;

  &.admin {
    background: linear-gradient(135deg, #f59e0b 0%, #d97706 100%);
  }
  &.dev {
    background: linear-gradient(135deg, #3b82f6 0%, #2563eb 100%);
  }
  &.test {
    background: linear-gradient(135deg, #10b981 0%, #059669 100%);
  }
  &.view {
    background: linear-gradient(135deg, #6b7280 0%, #4b5563 100%);
  }
}

.role-info {
  flex: 1;
  min-width: 0;

  .role-header {
    display: flex;
    align-items: center;
    gap: 8px;
    margin-bottom: 4px;
  }

  .role-name {
    font-size: 15px;
    font-weight: 600;
    color: #1f2937;
  }

  .role-key {
    font-size: 12px;
    color: #9ca3af;
    font-family: monospace;
    margin-bottom: 4px;
  }

  .role-desc {
    font-size: 13px;
    color: #6b7280;
    line-height: 1.5;
  }
}

.role-actions {
  display: flex;
  gap: 8px;
  flex-shrink: 0;
  align-items: flex-start;
}

// 工单类型网格
.issue-types-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(300px, 1fr));
  gap: 16px;
}

.issue-type-card {
  display: flex;
  align-items: center;
  gap: 16px;
  padding: 20px;
  background: #fff;
  border: 1px solid #e5e7eb;
  border-radius: 12px;
  transition: all 0.2s;

  &:hover {
    box-shadow: 0 4px 12px rgba(0, 0, 0, 0.08);
    border-color: #d1d5db;
  }
}

// 工单类型图标
.type-icon {
  width: 48px;
  height: 48px;
  border-radius: 12px;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 24px;
  color: #fff;
  flex-shrink: 0;
}

.type-info {
  flex: 1;
  min-width: 0;

  .type-name {
    font-size: 15px;
    font-weight: 600;
    color: #1f2937;
    margin-bottom: 4px;
  }

  .type-key {
    font-size: 12px;
    color: #9ca3af;
    font-family: monospace;
    margin-bottom: 4px;
  }

  .type-desc {
    font-size: 13px;
    color: #6b7280;
    line-height: 1.5;
  }
}

.type-actions {
  display: flex;
  gap: 8px;
  flex-shrink: 0;
}

// 对话框样式
.custom-dialog {
  :deep(.el-dialog__header) {
    padding: 20px 24px;
    border-bottom: 1px solid #e5e7eb;
  }

  :deep(.el-dialog__body) {
    padding: 24px;
  }
}

.user-option {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 4px 0;
}

.user-option-avatar {
  width: 36px;
  height: 36px;
  min-width: 36px;
  border-radius: 8px;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  color: #fff;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 14px;
  font-weight: 600;
  flex-shrink: 0;
}

.user-option-info {
  display: flex;
  flex-direction: column;
  flex: 1;
  min-width: 0;

  .user-option-name {
    font-size: 14px;
    font-weight: 500;
    color: #1f2937;
    line-height: 1.4;
  }

  .user-option-username {
    font-size: 12px;
    color: #9ca3af;
    line-height: 1.4;
  }
}

.role-option {
  display: flex;
  align-items: center;
  gap: 8px;
}

// 区块标题
.section-title {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 16px;
  font-weight: 600;
  color: #1f2937;
}

// 图标选择器
.icon-selector {
  display: grid;
  grid-template-columns: repeat(5, 1fr);
  gap: 8px;
}

.icon-option {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 6px;
  padding: 12px 8px;
  border: 2px solid #e5e7eb;
  border-radius: 8px;
  cursor: pointer;
  transition: all 0.2s;

  &:hover {
    border-color: #667eea;
    background: #f5f7ff;
  }

  &.active {
    border-color: #667eea;
    background: #f5f7ff;
    box-shadow: 0 0 0 3px rgba(102, 126, 234, 0.1);
  }

  .icon-label {
    font-size: 11px;
    color: #6b7280;
    text-align: center;
  }
}

// 颜色选择器
.color-selector {
  display: flex;
  align-items: center;
  gap: 16px;
}

.color-presets {
  display: flex;
  gap: 8px;
}

.color-preset {
  width: 32px;
  height: 32px;
  border-radius: 6px;
  cursor: pointer;
  transition: all 0.2s;
  border: 2px solid transparent;

  &:hover {
    transform: scale(1.1);
  }

  &.active {
    border-color: #1f2937;
    box-shadow: 0 0 0 3px rgba(0, 0, 0, 0.1);
  }
}

.form-tip-small {
  display: flex;
  align-items: flex-start;
  gap: 6px;
  font-size: 12px;
  color: #6b7280;
  margin-top: 6px;
  line-height: 1.5;

  .el-icon {
    flex-shrink: 0;
    margin-top: 2px;
    color: #3b82f6;
  }
}
</style>

<style lang="scss">
// 用户选择下拉框样式（不能用 scoped，因为是 popper）
.user-select-popper {
  .el-select-dropdown__item {
    height: auto !important;
    padding: 8px 12px !important;
    line-height: normal !important;
  }
}
</style>
