<template>
  <div class="board-issue-list">
    <!-- 列表头部 -->
    <div class="list-header">
      <div class="header-left">
        <span class="list-title">工单</span>
        <span class="list-count">{{ total }}</span>
      </div>
      <el-button type="primary" size="small" class="create-btn" @click="$emit('create')">
        <el-icon><Plus /></el-icon>
        创建
      </el-button>
    </div>

    <!-- 筛选区 -->
    <div class="list-filters">
      <el-input
        v-model="keyword"
        placeholder="搜索标题或编号..."
        clearable
        size="small"
        class="search-input"
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
          collapse-tags-tooltip
          clearable
          class="filter-select"
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
          class="filter-select"
          @change="handleSearch"
        >
          <el-option label="P0 - 紧急" value="P0" />
          <el-option label="P1 - 高" value="P1" />
          <el-option label="P2 - 中" value="P2" />
          <el-option label="P3 - 低" value="P3" />
        </el-select>
        <el-popover
          :visible="moreFilterVisible"
          placement="bottom-end"
          :width="280"
          trigger="click"
          popper-class="board-more-filter-popover"
        >
          <template #reference>
            <el-button
              size="small"
              class="more-filter-btn"
              :class="{ active: extraFilterCount > 0 }"
              @click="moreFilterVisible = !moreFilterVisible"
            >
              <el-icon><Filter /></el-icon>
              <span v-if="extraFilterCount > 0" class="filter-badge">{{ extraFilterCount }}</span>
            </el-button>
          </template>
          <div class="more-filter-panel">
            <div class="filter-item">
              <label class="filter-label">工单类型</label>
              <el-select
                v-model="issueTypeFilter"
                placeholder="全部"
                size="small"
                clearable
                style="width: 100%"
                @change="handleSearch"
              >
                <el-option
                  v-for="t in issueTypes"
                  :key="t.id"
                  :label="t.display_name"
                  :value="t.id"
                />
              </el-select>
            </div>
            <div class="filter-item">
              <label class="filter-label">经办人</label>
              <el-select
                v-model="assigneeFilter"
                placeholder="全部"
                size="small"
                clearable
                filterable
                style="width: 100%"
                @change="handleSearch"
              >
                <el-option
                  v-for="m in members"
                  :key="m.user_id"
                  :label="m.user?.display_name || `用户${m.user_id}`"
                  :value="m.user_id"
                />
              </el-select>
            </div>
            <div class="filter-item">
              <label class="filter-label">分类</label>
              <el-select
                v-model="categoryFilter"
                placeholder="全部"
                size="small"
                clearable
                style="width: 100%"
                @change="handleSearch"
              >
                <el-option label="普通工单" value="normal" />
                <el-option label="告警工单" value="alert" />
              </el-select>
            </div>
            <div class="filter-actions">
              <el-button size="small" text @click="resetExtraFilters">重置</el-button>
            </div>
          </div>
        </el-popover>
      </div>
    </div>

    <!-- 工单卡片列表 -->
    <div v-loading="loading" class="issue-cards">
      <div v-if="issueList.length === 0 && !loading" class="empty-state">
        <el-empty description="暂无匹配的工单" :image-size="60" />
      </div>
      <div
        v-for="item in issueList"
        :key="item.id"
        class="issue-card"
        :class="{ selected: item.issue_key === selectedKey }"
        @click="$emit('select', item.issue_key)"
      >
        <div class="card-top">
          <div class="key-group">
            <span class="priority-bar" :class="item.priority"></span>
            <a class="issue-key-link" @click.prevent.stop="router.push(`/issues/${item.issue_key}`)">{{ item.issue_key }}</a>
          </div>
          <div class="status-badge" :class="item.status">
            <span class="status-dot"></span>
            <span>{{ getStatusText(item.status) }}</span>
          </div>
        </div>
        <span class="card-title">{{ item.title }}</span>
        <div class="card-bottom">
          <div class="assignee-info">
            <div v-if="item.assignee" class="mini-avatar">{{ item.assignee.display_name?.charAt(0) || '?' }}</div>
            <div v-else class="mini-avatar unassigned">?</div>
            <span class="assignee-name">{{ item.assignee?.display_name || '未分配' }}</span>
          </div>
          <span v-if="item.issue_type" class="issue-type">{{ item.issue_type.display_name }}</span>
        </div>
      </div>
    </div>

    <!-- 简洁分页 -->
    <div v-if="totalPages > 1" class="list-pagination">
      <el-button text size="small" :disabled="page <= 1" @click="changePage(page - 1)">
        <el-icon><ArrowLeft /></el-icon>
      </el-button>
      <span class="page-info">{{ page }} / {{ totalPages }}</span>
      <el-button text size="small" :disabled="page >= totalPages" @click="changePage(page + 1)">
        <el-icon><ArrowRight /></el-icon>
      </el-button>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, watch } from 'vue'
import { useRouter } from 'vue-router'
import { Plus, Search, ArrowLeft, ArrowRight, Filter } from '@element-plus/icons-vue'
import { getIssueList } from '@/api/issue'
import { getProjectIssueTypes, getProjectMembers } from '@/api/project'
import type { Issue } from '@/types/issue'
import type { ProjectIssueType, ProjectMember } from '@/types/project'

const router = useRouter()

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

// 更多筛选
const moreFilterVisible = ref(false)
const issueTypeFilter = ref<number | undefined>(undefined)
const assigneeFilter = ref<number | undefined>(undefined)
const categoryFilter = ref('')
const issueTypes = ref<ProjectIssueType[]>([])
const members = ref<ProjectMember[]>([])

// 额外筛选生效数量
const extraFilterCount = computed(() => {
  let count = 0
  if (issueTypeFilter.value) count++
  if (assigneeFilter.value) count++
  if (categoryFilter.value) count++
  return count
})

const resetExtraFilters = () => {
  issueTypeFilter.value = undefined
  assigneeFilter.value = undefined
  categoryFilter.value = ''
  moreFilterVisible.value = false
  handleSearch()
}

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
      issue_type_id: issueTypeFilter.value || undefined,
      assignee_id: assigneeFilter.value || undefined,
      category: categoryFilter.value || undefined,
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

// 加载筛选选项数据
const loadFilterOptions = async () => {
  try {
    const [typesRes, membersRes] = await Promise.all([
      getProjectIssueTypes(props.projectKey),
      getProjectMembers(props.projectKey),
    ])
    issueTypes.value = typesRes.data.data || []
    members.value = membersRes.data.data || []
  } catch (e) {
    console.error('Failed to load filter options:', e)
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
  loadFilterOptions()
})

onMounted(() => {
  loadIssues()
  loadFilterOptions()
})
</script>

<style scoped lang="scss">
.board-issue-list {
  display: flex;
  flex-direction: column;
  height: 100%;
  background: var(--td-bg-card);
}

// ---- 头部 ----
.list-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 14px 16px;
  flex-shrink: 0;

  .header-left {
    display: flex;
    align-items: center;
    gap: 8px;
  }

  .list-title {
    font-size: 14px;
    font-weight: 600;
    color: var(--td-text-primary);
  }

  .list-count {
    display: inline-flex;
    align-items: center;
    justify-content: center;
    min-width: 22px;
    height: 20px;
    padding: 0 6px;
    border-radius: 10px;
    background: var(--td-bg-section);
    font-size: 11px;
    font-weight: 600;
    color: var(--td-text-secondary);
  }

  .create-btn {
    border-radius: 6px;
  }
}

// ---- 筛选区 ----
.list-filters {
  padding: 0 12px 12px;
  display: flex;
  flex-direction: column;
  gap: 8px;
  flex-shrink: 0;

  .search-input {
    :deep(.el-input__wrapper) {
      border-radius: 8px;
      box-shadow: 0 0 0 1px var(--td-border-color);

      &:focus-within {
        box-shadow: 0 0 0 2px rgba(59, 130, 246, 0.2);
      }
    }
  }

  .filter-row {
    display: flex;
    gap: 6px;

    .filter-select {
      flex: 1;
    }

    .more-filter-btn {
      flex-shrink: 0;
      padding: 4px 8px;
      position: relative;

      &.active {
        color: var(--td-color-primary);
        border-color: var(--td-color-primary);
        background: var(--td-tag-primary-bg);
      }

      .filter-badge {
        position: absolute;
        top: -4px;
        right: -4px;
        min-width: 16px;
        height: 16px;
        padding: 0 4px;
        border-radius: 8px;
        background: var(--td-color-primary);
        color: #fff;
        font-size: 10px;
        font-weight: 600;
        display: flex;
        align-items: center;
        justify-content: center;
        line-height: 1;
      }
    }
  }
}

// ---- 工单卡片列表 ----
.issue-cards {
  flex: 1;
  overflow-y: auto;
  padding: 4px 8px;

  // 自定义滚动条
  &::-webkit-scrollbar {
    width: 4px;
  }

  &::-webkit-scrollbar-track {
    background: transparent;
  }

  &::-webkit-scrollbar-thumb {
    background: var(--td-text-disabled);
    border-radius: 4px;

    &:hover {
      background: var(--td-text-placeholder);
    }
  }
}

.issue-card {
  padding: 10px 12px;
  margin-bottom: 2px;
  border-radius: 8px;
  cursor: pointer;
  border: 1px solid transparent;
  transition: background-color 150ms ease-out, border-color 150ms ease-out;
  position: relative;

  &:hover {
    background: var(--td-bg-page);
    border-color: var(--td-border-color);
  }

  &.selected {
    background: var(--td-tag-primary-bg);
    border-color: var(--td-tag-primary-border);
    box-shadow: inset 3px 0 0 #3b82f6;
  }

  .card-top {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-bottom: 4px;
  }

  .key-group {
    display: flex;
    align-items: center;
    gap: 6px;
  }

  .priority-bar {
    width: 3px;
    height: 14px;
    border-radius: 2px;
    flex-shrink: 0;

    &.P0 { background: var(--td-color-danger); }
    &.P1 { background: var(--td-color-warning); }
    &.P2 { background: var(--td-color-primary); }
    &.P3 { background: var(--td-color-success); }
  }

  .issue-key-link {
    font-size: 12px;
    font-weight: 600;
    color: var(--td-text-secondary);
    letter-spacing: 0.02em;
    text-decoration: none;
    transition: color 150ms ease-out;

    &:hover {
      color: var(--td-color-primary);
      text-decoration: underline;
    }
  }

  .card-title {
    display: -webkit-box;
    -webkit-line-clamp: 2;
    -webkit-box-orient: vertical;
    overflow: hidden;
    word-break: break-all;
    font-size: 13px;
    color: var(--td-text-primary);
    line-height: 1.5;
    margin-bottom: 8px;
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
    width: 20px;
    height: 20px;
    border-radius: 50%;
    background: var(--td-color-primary);
    color: var(--td-text-white);
    display: flex;
    align-items: center;
    justify-content: center;
    font-size: 10px;
    font-weight: 600;
    flex-shrink: 0;

    &.unassigned {
      background: var(--td-text-disabled);
      color: var(--td-text-white);
    }
  }

  .assignee-name {
    font-size: 12px;
    color: var(--td-text-placeholder);
    max-width: 80px;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }

  .issue-type {
    font-size: 11px;
    color: var(--td-text-placeholder);
    background: var(--td-bg-section);
    padding: 1px 6px;
    border-radius: 4px;
  }
}

// ---- 状态徽章 ----
.status-badge {
  display: inline-flex;
  align-items: center;
  gap: 4px;
  font-size: 11px;
  padding: 1px 7px;
  border-radius: 10px;
  font-weight: 500;
  white-space: nowrap;

  .status-dot {
    width: 5px;
    height: 5px;
    border-radius: 50%;
  }

  &.open { background: var(--td-tag-warning-border); color: var(--td-tag-orange-text); .status-dot { background: var(--td-color-warning); } }
  &.in_progress { background: var(--td-tag-primary-border); color: var(--td-tag-primary-text); .status-dot { background: var(--td-color-primary); } }
  &.pending_review { background: var(--td-tag-indigo-bg); color: var(--td-tag-indigo-text); .status-dot { background: var(--td-tag-indigo-text); } }
  &.resolved { background: var(--td-tag-success-border); color: var(--td-tag-success-text); .status-dot { background: var(--td-color-success); } }
  &.closed { background: var(--td-bg-section); color: var(--td-text-regular); .status-dot { background: var(--td-text-secondary); } }
  &.reopened { background: var(--td-tag-danger-border); color: var(--td-color-danger); .status-dot { background: var(--td-color-danger); } }
  &.merged { background: var(--td-tag-purple-bg); color: var(--td-tag-purple-text); .status-dot { background: var(--td-tag-purple-text); } }
}

// ---- 空状态 & 分页 ----
.empty-state {
  padding: 48px 0;
}

.list-pagination {
  display: flex;
  justify-content: center;
  align-items: center;
  gap: 8px;
  padding: 8px 12px;
  border-top: 1px solid var(--td-divider-color);
  flex-shrink: 0;

  .page-info {
    font-size: 12px;
    color: var(--td-text-placeholder);
    min-width: 48px;
    text-align: center;
  }
}

// ---- 更多筛选面板（scoped 内部分） ----
.more-filter-panel {
  .filter-item {
    margin-bottom: 12px;

    &:last-of-type {
      margin-bottom: 8px;
    }
  }

  .filter-label {
    display: block;
    font-size: 12px;
    color: var(--td-text-secondary);
    margin-bottom: 4px;
  }

  .filter-actions {
    display: flex;
    justify-content: flex-end;
    padding-top: 4px;
    border-top: 1px solid var(--td-divider-color);
  }
}

@media (prefers-reduced-motion: reduce) {
  .issue-card {
    transition: none;
  }
}
</style>
