<template>
  <div class="custom-node system-node" :class="{ selected: selected }">
    <Handle type="target" :position="Position.Top" />
    <div class="node-body">
      <div class="node-header">
        <div class="node-icon">
          <el-icon><Setting /></el-icon>
        </div>
        <div class="node-title">{{ data.label }}</div>
      </div>
      <div class="node-meta">
        <span class="node-type-badge">系统</span>
        <span v-if="data.config?.action" class="node-detail">{{ data.config.action }}</span>
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
import { Setting } from '@element-plus/icons-vue'
import type { NodeConfig } from '@/types/workflow'

const props = defineProps<{
  data: { label: string; config?: NodeConfig }
  selected?: boolean
}>()

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
.system-node {
  .node-body {
    background: #fff;
    border: 2px solid #8b5cf6;
    border-radius: 12px;
    min-width: 160px;
    overflow: hidden;
    box-shadow: 0 4px 12px rgba(139, 92, 246, 0.15);
    transition: box-shadow 0.2s, transform 0.2s;
  }

  &.selected .node-body {
    box-shadow: 0 0 0 3px rgba(139, 92, 246, 0.3), 0 4px 16px rgba(139, 92, 246, 0.25);
    transform: scale(1.02);
  }

  &:hover .node-body {
    box-shadow: 0 6px 16px rgba(139, 92, 246, 0.25);
  }

  .node-header {
    display: flex;
    align-items: center;
    gap: 8px;
    padding: 10px 14px;
    background: #8b5cf6;
    color: #fff;
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
    background: #ede9fe;
    color: #5b21b6;
    border-radius: 10px;
    font-weight: 500;
  }

  .node-detail {
    font-size: 11px;
    color: #5b21b6;
  }

  .node-status {
    padding: 4px 14px 8px;
    font-size: 11px;
    color: #6b7280;
  }
}
</style>
