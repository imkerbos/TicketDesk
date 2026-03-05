<template>
  <div class="custom-node end-node" :class="{ selected: selected }">
    <Handle type="target" :position="Position.Top" />
    <div class="node-body">
      <div class="node-icon">
        <el-icon><CircleCheck /></el-icon>
      </div>
      <div class="node-label">{{ data.label }}</div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { Handle, Position } from '@vue-flow/core'
import { CircleCheck } from '@element-plus/icons-vue'

defineProps<{
  data: { label: string }
  selected?: boolean
}>()
</script>

<style scoped lang="scss">
.end-node {
  .node-body {
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: 6px;
    padding: 12px 20px;
    background: #6b7280;
    border-radius: 50px;
    color: #fff;
    min-width: 80px;
    box-shadow: 0 4px 12px rgba(107, 114, 128, 0.3);
    transition: box-shadow 0.2s, transform 0.2s;
  }

  &.selected .node-body {
    box-shadow: 0 0 0 3px rgba(107, 114, 128, 0.4), 0 4px 16px rgba(107, 114, 128, 0.4);
    transform: scale(1.02);
  }

  &:hover .node-body {
    box-shadow: 0 6px 16px rgba(107, 114, 128, 0.4);
  }

  .node-icon {
    font-size: 24px;
    line-height: 1;
  }

  .node-label {
    font-size: 12px;
    font-weight: 600;
    white-space: nowrap;
  }
}
</style>
