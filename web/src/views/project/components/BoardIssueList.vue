<template>
  <div class="board-issue-list">
    <!-- 列表头部 -->
    <div class="list-header">
      <span class="list-title">工单列表 <span class="list-count">({{ total }})</span></span>
      <el-button type="primary" size="small" :icon="Plus" @click="$emit('create')">创建</el-button>
    </div>

    <!-- 筛选区 -->
    <div class="list-filters">
      <el-input
        v-model="keyword"
        placeholder="搜索工单标题"
        clearable
        size="small"
        @clear="handleSearch"
        @keyup.enter="handleSearch"
      >
        <template #prefix>
          <el-icon><Search /></el-icon>
        </template>
      </el-input>
      <div class="filter-row">
        <el-select
          v-model="statusFilter"
          placeholder="状态"
          size="small"
          multiple
          collapse-tags
          clearable
          @change="handleSearch"
        >
          <el-option label="待处理" value="open" />
          <el-option label="进行中" value="in_progress" />
          <el-option label="待确认" value="pending_review" />
          <el-option label="重新打开" value="reopened" />
          <el-option label="已完成" value="resolved" />
          <el-option label="已终止" value="closed" />
          <el-option label="已合并" value="merged" />
        </el-select>
        <el-select
          v-model="priorityFilter"
          placeholder="优先级"
          size="small"
          clearable
          @change="handleSearch"
        >
          <el-option label="P0" value="P0" />
          <el-option label="P1" value="P1" />
          <el-option label="P2" value="P2" />
          <el-option label="P3" value="P3" />
        </el-select>
      </div>
    </div>

    <!-- 工单卡片列表 -->
    <div v-loading="loading" class="issue-cards">
      <div v-if="issueList.length === 0 && !loading" class="empty-state">
        <el-empty description="暂无工单" :image-size="60" />
      </div>
      <div
        v-for="item in issueList"
        :key="item.id"
        class="issue-card"
        :class="{ selected: item.issue_key === selectedKey }"
        @click="$emit('select', item.issue_key)"
      >
        <div class="card-top">
          <span class="issue-key">{{ item.issue_key }}</span>
          <el-tag :type="getPriorityType(item.priority)" size="small" effect="dark">{{ item.priority }}</el-tag>
        </div>
        <div class="card-title">{{ item.title }}</div>
        <div class="card-bottom">
          <div class="assignee-info">
            <div v-if="item.assignee" class="mini-avatar">{{ item.assignee.display_name?.charAt(0) || '?' }}</div>
            <div v-else class="mini-avatar unassigned">?</div>
            <span class="assignee-name">{{ item.assignee?.display_name || '未分配' }}</span>
          </div>
          <div class="status-badge" :class="item.status">
            <span class="status-dot"></span>
            <span>{{ getStatusText(item.status) }}</span>
          </div>
        </div>
      </div>
    </div>

    <!-- 简洁分页 -->
    <div v-if="totalPages > 1" class="list-pagination">
      <el-button size="small" :disabled="page <= 1" @click="changePage(page - 1)">
        <el-icon><ArrowLeft /></el-icon>
      </el-button>
      <span class="page-info">{{ page }} / {{ totalPages }}</span>
      <el-button size="small" :disabled="page >= totalPages" @click="changePage(page + 1)">
        <el-icon><ArrowRight /></el-icon>
      </el-button>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, watch } from 'vue'
import { Plus, Search, ArrowLeft, ArrowRight } from '@element-plus/icons-vue'
import { getIssueList } from '@/api/issue'
import type { Issue } from '@/types/issue'

const props = defineProps<{
  projectKey: string
  selectedKey: string
}>()

defineEmits<{
  select: [issueKey: string]
  create: []
}>()

const loading = ref(false)
const issueList = ref<Issue[]>([])
const total = ref(0)
const page = ref(1)
const pageSize = 30
const keyword = ref('')
const statusFilter = ref<string[]>([])
const priorityFilter = ref('')

// 默认排除终态工单
const excludedStatuses = ['resolved', 'closed', 'merged']

const totalPages = computed(() => Math.ceil(total.value / pageSize) || 1)

const loadIssues = async () => {
  loading.value = true
  try {
    const params: Record<string, any> = {
      project_key: props.projectKey,
      page: page.value,
      page_size: pageSize,
      keyword: keyword.value || undefined,
      priority: priorityFilter.value || undefined,
    }
    // 如果用户选择了状态，使用用户选择的
    if (statusFilter.value.length > 0) {
      params.status = statusFilter.value.join(',')
    }
    const { data } = await getIssueList(params)
    let items = data.data.items || []
    // 前端排除终态（仅在未手动选择状态时）
    if (statusFilter.value.length === 0) {
      items = items.filter((i: Issue) => !excludedStatuses.includes(i.status))
    }
    issueList.value = items
    total.value = data.data.total
  } catch (e) {
    console.error('Failed to load issues:', e)
  } finally {
    loading.value = false
  }
}

const handleSearch = () => {
  page.value = 1
  loadIssues()
}

const changePage = (p: number) => {
  page.value = p
  loadIssues()
}

type TagType = 'primary' | 'success' | 'warning' | 'info' | 'danger'
const getPriorityType = (priority: string): TagType => {
  const map: Record<string, TagType> = { P0: 'danger', P1: 'warning', P2: 'info', P3: 'success' }
  return map[priority] || 'info'
}

const getStatusText = (status: string) => {
  const map: Record<string, string> = {
    open: '待处理', in_progress: '进行中', pending_review: '待确认',
    resolved: '已完成', closed: '已终止', reopened: '重新打开', merged: '已合并',
  }
  return map[status] || status
}

watch(() => props.projectKey, () => {
  page.value = 1
  loadIssues()
})

onMounted(() => {
  loadIssues()
})
</script>

<style scoped lang="scss">
.board-issue-list {
  display: flex;
  flex-direction: column;
  height: 100%;
  background: #fff;
}

.list-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 16px 16px 12px;
  border-bottom: 1px solid #e5e7eb;

  .list-title {
    font-size: 15px;
    font-weight: 600;
    color: #1f2937;
  }

  .list-count {
    font-weight: 400;
    color: #9ca3af;
    font-size: 13px;
  }
}

.list-filters {
  padding: 12px 16px;
  border-bottom: 1px solid #f0f0f0;
  display: flex;
  flex-direction: column;
  gap: 8px;

  .filter-row {
    display: flex;
    gap: 8px;

    .el-select {
      flex: 1;
    }
  }
}

.issue-cards {
  flex: 1;
  overflow-y: auto;
  padding: 8px;
}

.issue-card {
  padding: 12px;
  margin-bottom: 4px;
  border-radius: 8px;
  border: 1px solid transparent;
  cursor: pointer;
  transition: background-color 150ms ease-out, border-color 150ms ease-out;

  &:hover {
    background: #f9fafb;
  }

  &.selected {
    background: #eff6ff;
    border-left: 3px solid #3b82f6;
    padding-left: 9px;
  }

  .card-top {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-bottom: 6px;
  }

  .issue-key {
    font-size: 12px;
    font-weight: 600;
    color: #6b7280;
  }

  .card-title {
    font-size: 14px;
    color: #1f2937;
    line-height: 1.5;
    margin-bottom: 8px;
    display: -webkit-box;
    -webkit-line-clamp: 2;
    -webkit-box-orient: vertical;
    overflow: hidden;
  }

  .card-bottom {
    display: flex;
    justify-content: space-between;
    align-items: center;
  }

  .assignee-info {
    display: flex;
    align-items: center;
    gap: 6px;
  }

  .mini-avatar {
    width: 22px;
    height: 22px;
    border-radius: 6px;
    background: #3b82f6;
    color: #fff;
    display: flex;
    align-items: center;
    justify-content: center;
    font-size: 10px;
    font-weight: 600;
    flex-shrink: 0;

    &.unassigned {
      background: #d1d5db;
    }
  }

  .assignee-name {
    font-size: 12px;
    color: #6b7280;
  }
}

.status-badge {
  display: inline-flex;
  align-items: center;
  gap: 4px;
  font-size: 11px;
  padding: 2px 8px;
  border-radius: 10px;
  font-weight: 500;

  .status-dot {
    width: 6px;
    height: 6px;
    border-radius: 50%;
  }

  &.open { background: #fef3c7; color: #92400e; .status-dot { background: #f59e0b; } }
  &.in_progress { background: #dbeafe; color: #1e40af; .status-dot { background: #3b82f6; } }
  &.pending_review { background: #e0e7ff; color: #3730a3; .status-dot { background: #6366f1; } }
  &.resolved { background: #d1fae5; color: #065f46; .status-dot { background: #10b981; } }
  &.closed { background: #f3f4f6; color: #374151; .status-dot { background: #6b7280; } }
  &.reopened { background: #fee2e2; color: #991b1b; .status-dot { background: #ef4444; } }
  &.merged { background: #ede9fe; color: #5b21b6; .status-dot { background: #8b5cf6; } }
}

.empty-state {
  padding: 40px 0;
}

.list-pagination {
  display: flex;
  justify-content: center;
  align-items: center;
  gap: 12px;
  padding: 12px;
  border-top: 1px solid #e5e7eb;

  .page-info {
    font-size: 13px;
    color: #6b7280;
  }
}

@media (prefers-reduced-motion: reduce) {
  .issue-card {
    transition: none;
  }
}
</style>
