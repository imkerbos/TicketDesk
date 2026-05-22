<template>
  <div class="custom-node approval-node" :class="{ selected: selected }">
    <Handle type="target" :position="Position.Top" />
    <div class="node-body">
      <div class="node-header">
        <div class="node-icon">
          <el-icon><Checked /></el-icon>
        </div>
        <div class="node-title">{{ data.label }}</div>
      </div>
      <div class="node-meta">
        <span class="node-type-badge">审批</span>
        <span v-if="approvalTypeText" class="node-detail">{{ approvalTypeText }}</span>
      </div>
      <div v-if="data.config?.target_status" class="node-status">
        → {{ statusText }}
      </div>
    </div>
    <Handle type="source" :position="Position.Bottom" />
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { Handle, Position } from '@vue-flow/core'
import { Checked } from '@element-plus/icons-vue'
import type { NodeConfig } from '@/types/workflow'

const props = defineProps<{
  data: { label: string; config?: NodeConfig }
  selected?: boolean
}>()

const approvalTypeText = computed(() => {
  const map: Record<string, string> = {
    single: '单人审批',
    countersign: '会签',
    or_sign: '或签',
  }
  return props.data.config?.approval_type ? map[props.data.config.approval_type] : ''
})

const statusText = computed(() => {
  const map: Record<string, string> = {
    open: '未开始',
    in_progress: '进行中',
    pending_review: '待确认',
    resolved: '已解决',
    closed: '已关闭',
  }
  return props.data.config?.target_status ? (map[props.data.config.target_status] || props.data.config.target_status) : ''
})
</script>

<style scoped lang="scss">
.approval-node {
  .node-body {
    background: var(--td-bg-card);
    border: 2px solid #f59e0b;
    border-radius: 12px;
    min-width: 160px;
    overflow: hidden;
    box-shadow: 0 4px 12px rgba(245, 158, 11, 0.15);
    transition: box-shadow 150ms ease-out, transform 150ms ease-out;
  }

  &.selected .node-body {
    box-shadow: 0 0 0 3px rgba(245, 158, 11, 0.3), 0 4px 16px rgba(245, 158, 11, 0.25);
    transform: scale(1.02);
  }

  &:hover .node-body {
    box-shadow: 0 6px 16px rgba(245, 158, 11, 0.25);
  }

  .node-header {
    display: flex;
    align-items: center;
    gap: 8px;
    padding: 10px 14px;
    background: var(--td-color-warning);
    color: var(--td-text-white);
  }

  .node-icon {
    font-size: 18px;
    line-height: 1;
    flex-shrink: 0;
  }

  .node-title {
    font-size: 13px;
    font-weight: 600;
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
  }

  .node-meta {
    display: flex;
    align-items: center;
    gap: 6px;
    padding: 8px 14px;
    flex-wrap: wrap;
  }

  .node-type-badge {
    font-size: 11px;
    padding: 2px 8px;
    background: var(--td-tag-warning-border);
    color: var(--td-tag-orange-text);
    border-radius: 10px;
    font-weight: 500;
  }

  .node-detail {
    font-size: 11px;
    color: var(--td-tag-orange-text);
  }

  .node-status {
    padding: 4px 14px 8px;
    font-size: 11px;
    color: var(--td-text-secondary);
  }
}
</style>
