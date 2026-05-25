<template>
  <el-card shadow="never" class="settings-card">
    <template #header>
      <div class="card-header">
        <div class="card-title-group">
          <el-icon class="card-icon"><Key /></el-icon>
          <span>API 密钥</span>
          <el-tag size="small" type="info">{{ tokens.length }} / 20</el-tag>
        </div>
        <el-button type="primary" size="small" @click="openCreate">
          <el-icon><Plus /></el-icon>创建 Token
        </el-button>
      </div>
    </template>

    <div class="api-token-desc">
      用于外部系统通过 HTTP API 访问 TicketDesk (创建工单/查询等). Token 一旦撤销立即失效.
    </div>

    <div v-loading="loading">
      <el-table v-if="tokens.length > 0" :data="tokens" class="token-table">
        <el-table-column label="名称" min-width="150">
          <template #default="{ row }">
            <span class="token-name">{{ row.name }}</span>
          </template>
        </el-table-column>
        <el-table-column label="前缀" width="180">
          <template #default="{ row }">
            <code class="token-prefix">{{ row.prefix }}...</code>
          </template>
        </el-table-column>
        <el-table-column label="最后使用" width="160">
          <template #default="{ row }">
            <span v-if="row.last_used_at">{{ formatTime(row.last_used_at) }}</span>
            <span v-else class="never-used">从未使用</span>
          </template>
        </el-table-column>
        <el-table-column label="过期" width="160">
          <template #default="{ row }">
            <el-tag v-if="row.is_expired" type="danger" size="small">已过期</el-tag>
            <span v-else-if="row.expires_at">{{ formatTime(row.expires_at) }}</span>
            <el-tag v-else type="info" size="small">永久</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="操作" width="80" align="center">
          <template #default="{ row }">
            <el-button link type="danger" @click="confirmDelete(row)">撤销</el-button>
          </template>
        </el-table-column>
      </el-table>

      <TdEmptyState
        v-else-if="!loading"
        preset="first-time"
        title="还没有 API 密钥"
        description="创建后可用于外部系统调用 TicketDesk API"
      >
        <el-button type="primary" @click="openCreate">
          <el-icon><Plus /></el-icon>创建第一个 Token
        </el-button>
      </TdEmptyState>
    </div>

    <!-- 创建对话框 -->
    <el-dialog v-model="createDialogVisible" title="创建 API 密钥" width="480px" destroy-on-close>
      <el-form ref="formRef" :model="form" :rules="formRules" label-position="top">
        <el-form-item label="名称" prop="name">
          <el-input v-model="form.name" placeholder="例如: 监控系统对接" maxlength="100" show-word-limit />
        </el-form-item>
        <el-form-item label="过期时间">
          <el-radio-group v-model="form.expires_preset">
            <el-radio value="7d">7 天</el-radio>
            <el-radio value="30d">30 天</el-radio>
            <el-radio value="90d">90 天</el-radio>
            <el-radio value="365d">365 天</el-radio>
            <el-radio value="never">永久</el-radio>
          </el-radio-group>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="createDialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="creating" @click="submitCreate">创建</el-button>
      </template>
    </el-dialog>

    <!-- 一次性明文展示对话框 -->
    <el-dialog
      v-model="showTokenDialog"
      title="保存你的 API 密钥"
      width="600px"
      :close-on-click-modal="false"
      :show-close="false"
    >
      <el-alert type="warning" :closable="false" class="token-warning">
        这是 <strong>{{ newToken?.name }}</strong> 的唯一显示机会. 关闭后无法再查看明文, 请立即复制保存到安全位置.
      </el-alert>

      <div class="token-display">
        <code>{{ newToken?.token }}</code>
        <el-button type="primary" size="small" @click="copyToken">
          <el-icon><CopyDocument /></el-icon>复制
        </el-button>
      </div>

      <div class="usage-section">
        <div class="usage-label">使用示例 (curl)</div>
        <el-input
          :model-value="usageExample"
          type="textarea"
          :rows="3"
          readonly
          class="usage-snippet"
        />
      </div>

      <template #footer>
        <el-button type="primary" @click="confirmSaved">我已保存</el-button>
      </template>
    </el-dialog>
  </el-card>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { ElMessage, ElMessageBox, type FormInstance, type FormRules } from 'element-plus'
import { Key, Plus, CopyDocument } from '@element-plus/icons-vue'
import { listApiTokens, createApiToken, deleteApiToken } from '@/api/token'
import type { APIToken, CreateTokenResponse } from '@/types/token'
import dayjs from 'dayjs'

const tokens = ref<APIToken[]>([])
const loading = ref(false)

const createDialogVisible = ref(false)
const creating = ref(false)
const formRef = ref<FormInstance>()
const form = ref({
  name: '',
  expires_preset: '90d' as '7d' | '30d' | '90d' | '365d' | 'never',
})

const formRules: FormRules = {
  name: [
    { required: true, message: '请输入名称', trigger: 'blur' },
    { max: 100, message: '名称最长 100 字符', trigger: 'blur' },
  ],
}

const showTokenDialog = ref(false)
const newToken = ref<CreateTokenResponse | null>(null)

const usageExample = computed(() => {
  if (!newToken.value) return ''
  return `curl -X POST ${window.location.origin}/api/v1/issues \\\n  -H "Authorization: Bearer ${newToken.value.token}" \\\n  -H "Content-Type: application/json" \\\n  -d '{"project_key":"YOUR_PROJ","issue_type_id":1,"title":"..."}'`
})

// 将过期时间预设转换为秒数
const presetToSeconds = (preset: string): number => {
  const day = 24 * 60 * 60
  switch (preset) {
    case '7d': return 7 * day
    case '30d': return 30 * day
    case '90d': return 90 * day
    case '365d': return 365 * day
    case 'never':
    default: return 0
  }
}

const formatTime = (s: string | null) => (s ? dayjs(s).format('YYYY-MM-DD HH:mm') : '')

const loadTokens = async () => {
  loading.value = true
  try {
    const { data } = await listApiTokens()
    tokens.value = data.data || []
  } catch {
    // 静默处理，避免影响页面渲染
  } finally {
    loading.value = false
  }
}

const openCreate = () => {
  if (tokens.value.length >= 20) {
    ElMessage.warning('已达 20 个 token 上限, 请先撤销不用的')
    return
  }
  form.value = { name: '', expires_preset: '90d' }
  createDialogVisible.value = true
}

const submitCreate = async () => {
  if (!formRef.value) return
  await formRef.value.validate(async (valid) => {
    if (!valid) return
    creating.value = true
    try {
      const { data } = await createApiToken({
        name: form.value.name,
        expires_in: presetToSeconds(form.value.expires_preset),
      })
      newToken.value = data.data
      createDialogVisible.value = false
      showTokenDialog.value = true
      await loadTokens()
    } catch {
      ElMessage.error('创建失败')
    } finally {
      creating.value = false
    }
  })
}

const copyToken = async () => {
  if (!newToken.value) return
  try {
    await navigator.clipboard.writeText(newToken.value.token)
    ElMessage.success('已复制到剪贴板')
  } catch {
    ElMessage.error('复制失败, 请手动复制')
  }
}

const confirmSaved = () => {
  showTokenDialog.value = false
  newToken.value = null
}

const confirmDelete = async (token: APIToken) => {
  try {
    await ElMessageBox.confirm(
      `确定要撤销 "${token.name}" 吗? 使用此 token 的外部系统将立即无法访问.`,
      '撤销 Token',
      { type: 'warning', confirmButtonText: '撤销', cancelButtonText: '取消', confirmButtonClass: 'el-button--danger' },
    )
  } catch {
    return
  }
  try {
    await deleteApiToken(token.id)
    ElMessage.success('已撤销')
    await loadTokens()
  } catch {
    ElMessage.error('撤销失败')
  }
}

onMounted(() => {
  loadTokens()
})
</script>

<style scoped lang="scss">
.settings-card {
  border-radius: var(--td-radius-lg);
  border: none;
  box-shadow: var(--td-elevation-1);
  margin-bottom: var(--td-space-5);

  :deep(.el-card__header) {
    padding: var(--td-space-4) var(--td-space-5);
    border-bottom: 1px solid var(--td-border-color);
  }
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.card-title-group {
  display: flex;
  align-items: center;
  gap: var(--td-space-2);
  font-size: var(--td-font-md);
  font-weight: var(--td-weight-semibold);
  color: var(--td-text-primary);

  .card-icon {
    color: var(--td-color-primary);
    font-size: 18px;
  }
}

.api-token-desc {
  font-size: var(--td-font-sm);
  color: var(--td-text-secondary);
  margin-bottom: var(--td-space-4);
  line-height: var(--td-leading-normal);
}

.token-table {
  .token-name {
    font-weight: var(--td-weight-medium);
    color: var(--td-text-primary);
  }

  .token-prefix {
    font-family: var(--el-font-family-mono, monospace);
    font-size: var(--td-font-sm);
    color: var(--td-text-secondary);
    padding: 2px var(--td-space-2);
    background: var(--td-bg-section);
    border-radius: var(--td-radius-xs);
  }

  .never-used {
    color: var(--td-text-placeholder);
    font-style: italic;
  }
}

// 一次性明文展示
.token-warning {
  margin-bottom: var(--td-space-4);
}

.token-display {
  display: flex;
  align-items: center;
  gap: var(--td-space-2);
  padding: var(--td-space-3) var(--td-space-4);
  background: var(--td-bg-section);
  border: 1px solid var(--td-border-color);
  border-radius: var(--td-radius-md);
  margin-bottom: var(--td-space-4);

  code {
    flex: 1;
    font-family: var(--el-font-family-mono, monospace);
    font-size: var(--td-font-md);
    color: var(--td-text-primary);
    word-break: break-all;
    user-select: all;
  }
}

.usage-section {
  .usage-label {
    font-size: var(--td-font-sm);
    color: var(--td-text-secondary);
    margin-bottom: var(--td-space-2);
    font-weight: var(--td-weight-medium);
  }
}

.usage-snippet {
  :deep(.el-textarea__inner) {
    font-family: var(--el-font-family-mono, monospace);
    font-size: var(--td-font-sm);
  }
}
</style>
