<template>
  <div class="datasource-page">
    <!-- 页头 -->
    <div class="page-header">
      <div class="header-info">
        <h2 class="header-title">告警数据源</h2>
        <p class="header-desc">管理告警数据源连接，支持 Prometheus 和夜莺 (Nightingale) 等监控系统</p>
      </div>
      <el-button type="primary" size="large" @click="openCreateDialog">
        <el-icon><Plus /></el-icon>
        添加数据源
      </el-button>
    </div>

    <!-- 数据源卡片列表 -->
    <div v-loading="loading" class="datasource-list">
      <!-- 空状态 -->
      <div v-if="!loading && datasources.length === 0" class="empty-state">
        <div class="empty-icon">
          <svg viewBox="0 0 120 120" width="120" height="120" fill="none">
            <rect x="20" y="30" width="80" height="60" rx="8" fill="#e0e7ff" stroke="#818cf8" stroke-width="2"/>
            <circle cx="60" cy="55" r="12" fill="#818cf8" opacity="0.3"/>
            <path d="M54 55l4 4 8-8" stroke="#4f46e5" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round"/>
            <rect x="35" y="70" width="50" height="4" rx="2" fill="#c7d2fe"/>
            <rect x="42" y="78" width="36" height="3" rx="1.5" fill="#e0e7ff"/>
            <path d="M60 20v6M44 24l3 5.2M76 24l-3 5.2" stroke="#818cf8" stroke-width="2" stroke-linecap="round"/>
          </svg>
        </div>
        <h3 class="empty-title">暂无数据源</h3>
        <p class="empty-desc">添加一个告警数据源，开始接收来自 Prometheus 或夜莺的告警事件</p>
        <el-button type="primary" @click="openCreateDialog">
          <el-icon><Plus /></el-icon>
          添加第一个数据源
        </el-button>
      </div>

      <!-- 卡片网格 -->
      <div v-else class="card-grid">
        <div v-for="ds in datasources" :key="ds.id" class="datasource-card" :class="{ 'card-disabled': ds.status === 0 }">
          <!-- 卡片顶部：类型色条 -->
          <div class="card-color-bar" :class="ds.type === 'prometheus' ? 'bar-prometheus' : 'bar-nightingale'" />

          <div class="card-body">
            <!-- 标题行 -->
            <div class="card-header">
              <div class="card-title-row">
                <span class="type-icon" v-html="getTypeLogo(ds.type)" />
                <div class="card-title-info">
                  <span class="card-name">{{ ds.name }}</span>
                  <div class="card-tags">
                    <el-tag size="small" :type="ds.type === 'prometheus' ? 'danger' : 'warning'" effect="light">
                      {{ getTypeLabel(ds.type) }}
                    </el-tag>
                    <el-tag size="small" :type="ds.push_mode ? 'info' : 'success'" effect="light">
                      {{ ds.push_mode ? 'Webhook 推送' : 'API 轮询' }}
                    </el-tag>
                  </div>
                </div>
              </div>
              <el-switch
                :model-value="ds.status === 1"
                inline-prompt
                active-text="ON"
                inactive-text="OFF"
                @change="(val: boolean) => handleToggleStatus(ds, val)"
              />
            </div>

            <!-- 描述 -->
            <div v-if="ds.description" class="card-desc">{{ ds.description }}</div>

            <!-- 详情网格 -->
            <div class="card-details">
              <div class="detail-item">
                <span class="detail-label">连接状态</span>
                <span v-if="ds.last_check_ok === true" class="status-badge status-ok">
                  <span class="status-indicator" />正常
                </span>
                <span v-else-if="ds.last_check_ok === false" class="status-badge status-fail">
                  <span class="status-indicator" />异常
                </span>
                <span v-else class="status-badge status-unknown">
                  <span class="status-indicator" />未检测
                </span>
              </div>

              <div v-if="ds.push_mode" class="detail-item detail-item-wide">
                <span class="detail-label">Webhook URL</span>
                <div class="webhook-row">
                  <code class="webhook-url">{{ getFullWebhookUrl(ds.webhook_url) }}</code>
                  <el-button link type="primary" size="small" @click="copyWebhookUrl(ds.webhook_url)">
                    复制
                  </el-button>
                </div>
              </div>

              <div v-if="!ds.push_mode" class="detail-item">
                <span class="detail-label">轮询间隔</span>
                <span class="detail-value">{{ ds.poll_interval }} 秒</span>
              </div>

              <div v-if="ds.last_check_at" class="detail-item">
                <span class="detail-label">最后检测</span>
                <span class="detail-value">{{ formatTime(ds.last_check_at) }}</span>
              </div>

              <div v-if="ds.last_check_ok === false && ds.last_check_msg" class="detail-item detail-item-wide">
                <span class="detail-label">错误信息</span>
                <span class="detail-value detail-error">{{ ds.last_check_msg }}</span>
              </div>
            </div>

            <!-- 操作栏 -->
            <div class="card-footer">
              <el-button size="small" @click="handleTestConnection(ds)">
                <el-icon><Connection /></el-icon>
                测试连接
              </el-button>
              <el-button size="small" @click="openEditDialog(ds)">
                <el-icon><Edit /></el-icon>
                编辑
              </el-button>
              <el-popconfirm
                title="确定要删除此数据源吗？关联的轮询器将被停止。"
                confirm-button-text="删除"
                cancel-button-text="取消"
                @confirm="handleDelete(ds.id)"
              >
                <template #reference>
                  <el-button size="small" type="danger" plain>
                    <el-icon><Delete /></el-icon>
                    删除
                  </el-button>
                </template>
              </el-popconfirm>
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- 创建/编辑对话框 -->
    <el-dialog
      v-model="dialogVisible"
      :title="isEdit ? '编辑数据源' : '添加数据源'"
      width="600px"
      :close-on-click-modal="false"
    >
      <el-form ref="formRef" :model="form" :rules="formRules" label-width="100px">
        <el-form-item label="名称" prop="name">
          <el-input v-model="form.name" placeholder="如 prod-nightingale、staging-prometheus" :disabled="isEdit" />
          <div class="form-hint">名称将作为告警来源标识，创建后不可修改</div>
        </el-form-item>

        <el-form-item label="类型" prop="type">
          <el-select v-model="form.type" placeholder="选择数据源类型" :disabled="isEdit" style="width: 100%">
            <el-option
              v-for="t in DatasourceTypes"
              :key="t.value"
              :label="t.label"
              :value="t.value"
            />
          </el-select>
        </el-form-item>

        <el-form-item label="描述">
          <el-input v-model="form.description" type="textarea" :rows="2" placeholder="可选描述" />
        </el-form-item>

        <el-form-item label="接入模式">
          <div class="mode-selector">
            <el-radio-group v-model="form.push_mode">
              <el-radio :value="true">Webhook 推送</el-radio>
              <el-radio :value="false">API 轮询</el-radio>
            </el-radio-group>
            <div class="form-hint">
              {{ form.push_mode ? '监控系统主动推送告警到 TicketDesk Webhook 地址' : '由 TicketDesk 定时拉取监控系统的活跃告警' }}
            </div>
          </div>
        </el-form-item>

        <!-- 夜莺配置 -->
        <template v-if="form.type === 'nightingale'">
          <el-form-item label="API 地址" prop="config.base_url">
            <el-input v-model="form.config.base_url" placeholder="http://nightingale:17000" />
          </el-form-item>
          <el-form-item label="Token">
            <el-input v-model="form.config.token" placeholder="夜莺 API Token（可选）" show-password />
          </el-form-item>
        </template>

        <!-- Prometheus 配置 -->
        <template v-if="form.type === 'prometheus'">
          <el-form-item label="API 地址">
            <el-input v-model="form.config.base_url" placeholder="http://prometheus:9090（推送模式可留空）" />
          </el-form-item>
        </template>

        <el-form-item v-if="!form.push_mode" label="轮询间隔">
          <el-input-number v-model="form.poll_interval" :min="5" :max="3600" :step="5" />
          <span style="margin-left: 8px; color: #909399">秒</span>
        </el-form-item>
      </el-form>

      <template #footer>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button :loading="testLoading" @click="handleTestBeforeSave">
          <el-icon><Connection /></el-icon>
          测试连接
        </el-button>
        <el-button type="primary" :loading="saveLoading" @click="handleSave">
          {{ isEdit ? '保存' : '创建' }}
        </el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { Plus, Edit, Delete, Connection } from '@element-plus/icons-vue'
import { ElMessage, type FormInstance, type FormRules } from 'element-plus'
import {
  getDatasourceList,
  createDatasource,
  updateDatasource,
  deleteDatasource,
  testDatasource,
  testDatasourceById,
} from '@/api/alert'
import { DatasourceTypes } from '@/types/alert'
import type { AlertDatasource } from '@/types/alert'

const loading = ref(false)
const datasources = ref<AlertDatasource[]>([])
const dialogVisible = ref(false)
const isEdit = ref(false)
const editId = ref(0)
const saveLoading = ref(false)
const testLoading = ref(false)
const formRef = ref<FormInstance>()

const form = reactive({
  name: '',
  type: 'nightingale' as string,
  description: '',
  config: {} as Record<string, any>,
  push_mode: false,
  poll_interval: 30,
})

const formRules: FormRules = {
  name: [
    { required: true, message: '请输入数据源名称', trigger: 'blur' },
    { min: 1, max: 100, message: '名称长度 1-100 个字符', trigger: 'blur' },
    { pattern: /^[a-zA-Z0-9_-]+$/, message: '仅支持字母、数字、下划线和连字符', trigger: 'blur' },
  ],
  type: [{ required: true, message: '请选择数据源类型', trigger: 'change' }],
}

const fetchDatasources = async () => {
  loading.value = true
  try {
    const res = await getDatasourceList({ page: 1, page_size: 100 })
    datasources.value = res.data?.data?.items || []
  } catch {
    ElMessage.error('获取数据源列表失败')
  } finally {
    loading.value = false
  }
}

const getTypeIcon = (type: string) => {
  const found = DatasourceTypes.find((t) => t.value === type)
  return found?.icon || '📡'
}

const getTypeLogo = (type: string) => {
  if (type === 'nightingale') {
    // 夜莺 logo — 紫色蜂鸟风格
    return `<svg viewBox="0 0 200 200" width="36" height="36" xmlns="http://www.w3.org/2000/svg">
      <defs>
        <linearGradient id="n9e-bg" x1="0%" y1="0%" x2="100%" y2="100%">
          <stop offset="0%" style="stop-color:#6366f1"/>
          <stop offset="100%" style="stop-color:#8b5cf6"/>
        </linearGradient>
      </defs>
      <circle cx="100" cy="100" r="96" fill="url(#n9e-bg)"/>
      <g transform="translate(100,100) scale(0.7) translate(-100,-100)">
        <path d="M140 50 C165 55 175 80 170 105 C165 130 145 155 115 170 L100 175 L95 165 C90 150 88 130 92 110 C95 95 102 80 115 68 C125 58 135 52 140 50Z" fill="#fff" opacity="0.95"/>
        <path d="M100 175 L85 170 C65 158 50 140 48 115 C46 95 55 78 70 65 L92 110 C88 130 90 150 95 165 Z" fill="#fff" opacity="0.75"/>
        <path d="M40 85 C45 80 55 78 70 65" fill="none" stroke="#fff" stroke-width="5" stroke-linecap="round" opacity="0.6"/>
        <circle cx="148" cy="72" r="5" fill="url(#n9e-bg)"/>
      </g>
    </svg>`
  }
  if (type === 'prometheus') {
    // Prometheus 官方 logo
    return `<svg viewBox="0 0 430 430" width="36" height="36" xmlns="http://www.w3.org/2000/svg">
      <path fill="#E75225" d="M215.926 7.068c115.684.024 210.638 93.784 210.493 207.844-.148 115.793-94.713 208.252-212.912 208.169C97.95 423 4.52 329.143 4.601 213.221 4.68 99.867 99.833 7.044 215.926 7.068zm-63.947 73.001c2.652 12.978.076 25.082-3.846 36.988-2.716 8.244-6.47 16.183-8.711 24.539-3.694 13.769-7.885 27.619-9.422 41.701-2.21 20.25 5.795 38.086 19.493 55.822L86.527 225.94c.11 1.978-.007 2.727.21 3.361 5.968 17.43 16.471 32.115 28.243 45.957 1.246 1.465 4.082 2.217 6.182 2.221 62.782.115 125.565.109 188.347.028 1.948-.003 4.546-.369 5.741-1.618 13.456-14.063 23.746-30.079 30.179-50.257l-66.658 12.976c4.397-8.567 9.417-16.1 12.302-24.377 9.869-28.315 5.779-55.69-8.387-81.509-11.368-20.72-21.854-41.349-16.183-66.32-12.005 11.786-16.615 26.79-19.541 42.253-2.882 15.23-4.58 30.684-6.811 46.136-.317-.467-.728-.811-.792-1.212-.258-1.621-.499-3.255-.587-4.893-1.355-25.31-6.328-49.696-16.823-72.987-6.178-13.71-12.99-27.727-6.622-44.081-4.31 2.259-8.205 4.505-10.997 7.711-8.333 9.569-11.779 21.062-12.666 33.645-.757 10.75-1.796 21.552-3.801 32.123-2.107 11.109-5.448 21.998-12.956 32.209-3.033-21.81-3.37-43.38-22.928-57.237zm161.877 216.523H116.942v34.007h196.914v-34.007zm-157.871 51.575c-.163 28.317 28.851 49.414 64.709 47.883 29.716-1.269 56.016-24.51 53.755-47.883H155.985z"/>
    </svg>`
  }
  return ''
}

const getTypeLabel = (type: string) => {
  const found = DatasourceTypes.find((t) => t.value === type)
  return found?.label || type
}

const getFullWebhookUrl = (path: string) => {
  return `${window.location.origin}${path}`
}

const copyWebhookUrl = async (path: string) => {
  try {
    await navigator.clipboard.writeText(getFullWebhookUrl(path))
    ElMessage.success('已复制到剪贴板')
  } catch {
    ElMessage.error('复制失败')
  }
}

const formatTime = (time: string) => {
  if (!time) return '-'
  return new Date(time).toLocaleString('zh-CN')
}

const resetForm = () => {
  form.name = ''
  form.type = 'nightingale'
  form.description = ''
  form.config = {}
  form.push_mode = false
  form.poll_interval = 30
}

const openCreateDialog = () => {
  resetForm()
  isEdit.value = false
  editId.value = 0
  dialogVisible.value = true
}

const openEditDialog = (ds: AlertDatasource) => {
  isEdit.value = true
  editId.value = ds.id
  form.name = ds.name
  form.type = ds.type
  form.description = ds.description
  form.config = { ...ds.config }
  form.push_mode = ds.push_mode
  form.poll_interval = ds.poll_interval
  dialogVisible.value = true
}

const handleSave = async () => {
  if (!formRef.value) return
  await formRef.value.validate()

  saveLoading.value = true
  try {
    if (isEdit.value) {
      await updateDatasource(editId.value, {
        description: form.description,
        config: form.config,
        push_mode: form.push_mode,
        poll_interval: form.poll_interval,
      })
      ElMessage.success('数据源已更新')
    } else {
      await createDatasource({
        name: form.name,
        type: form.type,
        description: form.description,
        config: form.config,
        push_mode: form.push_mode,
        poll_interval: form.poll_interval,
      })
      ElMessage.success('数据源已创建')
    }
    dialogVisible.value = false
    fetchDatasources()
  } catch (err: any) {
    ElMessage.error(err?.response?.data?.message || '保存失败')
  } finally {
    saveLoading.value = false
  }
}

const handleToggleStatus = async (ds: AlertDatasource, enabled: boolean) => {
  try {
    await updateDatasource(ds.id, { status: enabled ? 1 : 0 })
    ElMessage.success(enabled ? '数据源已启用' : '数据源已禁用')
    fetchDatasources()
  } catch {
    ElMessage.error('操作失败')
  }
}

const handleDelete = async (id: number) => {
  try {
    await deleteDatasource(id)
    ElMessage.success('数据源已删除')
    fetchDatasources()
  } catch {
    ElMessage.error('删除失败')
  }
}

const handleTestConnection = async (ds: AlertDatasource) => {
  try {
    const res = await testDatasourceById(ds.id)
    const result = res.data?.data
    if (result?.success) {
      ElMessage.success(`连接成功 (${result.latency_ms}ms)`)
    } else {
      ElMessage.warning(result?.message || '连接失败')
    }
    fetchDatasources()
  } catch {
    ElMessage.error('测试连接失败')
  }
}

const handleTestBeforeSave = async () => {
  testLoading.value = true
  try {
    const res = await testDatasource({
      type: form.type,
      config: form.config,
    })
    const result = res.data?.data
    if (result?.success) {
      ElMessage.success(`连接成功 (${result.latency_ms}ms)`)
    } else {
      ElMessage.warning(result?.message || '连接失败')
    }
  } catch {
    ElMessage.error('测试连接失败')
  } finally {
    testLoading.value = false
  }
}

onMounted(() => {
  fetchDatasources()
})
</script>

<style scoped>
.datasource-page {
  width: 100%;
}

/* ===== 页头 ===== */
.page-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 24px;
  padding: 28px 32px;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  border-radius: 16px;
  color: #fff;
}

.header-title {
  margin: 0 0 6px 0;
  font-size: 24px;
  font-weight: 700;
  letter-spacing: 0.5px;
}

.header-desc {
  margin: 0;
  font-size: 14px;
  opacity: 0.85;
}

.page-header :deep(.el-button) {
  flex-shrink: 0;
  border-color: rgba(255, 255, 255, 0.4);
  background: rgba(255, 255, 255, 0.15);
  color: #fff;
  font-weight: 500;
  backdrop-filter: blur(4px);
}

.page-header :deep(.el-button:hover) {
  background: rgba(255, 255, 255, 0.3);
  border-color: rgba(255, 255, 255, 0.6);
}

/* ===== 空状态 ===== */
.empty-state {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 80px 20px;
  background: #fff;
  border: 2px dashed #e5e7eb;
  border-radius: 16px;
}

.empty-icon {
  margin-bottom: 20px;
  opacity: 0.8;
}

.empty-title {
  margin: 0 0 8px 0;
  font-size: 18px;
  font-weight: 600;
  color: #374151;
}

.empty-desc {
  margin: 0 0 24px 0;
  font-size: 14px;
  color: #9ca3af;
}

/* ===== 卡片列表 ===== */
.card-grid {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.datasource-card {
  background: #fff;
  border: 1px solid #e5e7eb;
  border-radius: 12px;
  overflow: hidden;
  transition: box-shadow 0.25s, transform 0.25s;
}

.datasource-card:hover {
  box-shadow: 0 8px 24px rgba(0, 0, 0, 0.08);
  transform: translateY(-2px);
}

.datasource-card.card-disabled {
  opacity: 0.6;
}

.card-color-bar {
  height: 4px;
}

.bar-prometheus {
  background: linear-gradient(90deg, #e74c3c 0%, #f39c12 100%);
}

.bar-nightingale {
  background: linear-gradient(90deg, #f59e0b 0%, #10b981 100%);
}

.card-body {
  padding: 20px 24px;
}

/* ===== 卡片头部 ===== */
.card-header {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  margin-bottom: 12px;
}

.card-title-row {
  display: flex;
  align-items: flex-start;
  gap: 12px;
  min-width: 0;
}

.type-icon {
  width: 36px;
  height: 36px;
  flex-shrink: 0;
  display: flex;
  align-items: center;
  justify-content: center;
}

.card-title-info {
  display: flex;
  flex-direction: column;
  gap: 6px;
  min-width: 0;
}

.card-name {
  font-size: 17px;
  font-weight: 600;
  color: #1f2937;
  word-break: break-all;
}

.card-tags {
  display: flex;
  gap: 6px;
  flex-wrap: wrap;
}

.card-desc {
  color: #6b7280;
  font-size: 13px;
  margin-bottom: 16px;
  line-height: 1.5;
}

/* ===== 详情区域 ===== */
.card-details {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(160px, 1fr));
  gap: 12px 32px;
  padding: 16px 0;
  border-top: 1px solid #f3f4f6;
  border-bottom: 1px solid #f3f4f6;
  margin-bottom: 16px;
}

.detail-item {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.detail-item-wide {
  grid-column: 1 / -1;
}

.detail-label {
  font-size: 12px;
  color: #9ca3af;
  font-weight: 500;
  text-transform: uppercase;
  letter-spacing: 0.5px;
}

.detail-value {
  font-size: 13px;
  color: #374151;
}

.detail-error {
  color: #ef4444;
  font-size: 12px;
  line-height: 1.4;
}

/* ===== 状态徽章 ===== */
.status-badge {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  font-size: 13px;
  font-weight: 500;
}

.status-indicator {
  width: 8px;
  height: 8px;
  border-radius: 50%;
  display: inline-block;
}

.status-ok .status-indicator {
  background: #10b981;
  box-shadow: 0 0 6px rgba(16, 185, 129, 0.4);
}

.status-ok {
  color: #059669;
}

.status-fail .status-indicator {
  background: #ef4444;
  box-shadow: 0 0 6px rgba(239, 68, 68, 0.4);
}

.status-fail {
  color: #dc2626;
}

.status-unknown .status-indicator {
  background: #d1d5db;
}

.status-unknown {
  color: #9ca3af;
}

/* ===== Webhook URL ===== */
.webhook-row {
  display: flex;
  align-items: center;
  gap: 8px;
}

.webhook-url {
  background: #f9fafb;
  border: 1px solid #e5e7eb;
  padding: 4px 10px;
  border-radius: 6px;
  font-size: 12px;
  color: #374151;
  font-family: 'SF Mono', 'Fira Code', monospace;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  max-width: 100%;
}

/* ===== 操作栏 ===== */
.card-footer {
  display: flex;
  gap: 8px;
  flex-wrap: wrap;
}

/* ===== 表单 ===== */
.mode-selector {
  display: flex;
  flex-direction: column;
  gap: 4px;
  width: 100%;
}

.form-hint {
  font-size: 12px;
  color: #9ca3af;
  margin-top: 4px;
  line-height: 1.4;
}

/* ===== 响应式 ===== */
@media (max-width: 768px) {
  .card-details {
    grid-template-columns: 1fr 1fr;
  }
}
</style>
