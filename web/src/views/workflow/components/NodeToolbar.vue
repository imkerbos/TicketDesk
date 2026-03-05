<template>
  <div class="node-toolbar">
    <div class="toolbar-title">节点类型</div>
    <div class="toolbar-hint">拖拽到画布添加节点</div>
    <div class="toolbar-items">
      <div
        v-for="item in nodeTypes"
        :key="item.type"
        class="toolbar-item"
        :class="item.type"
        draggable="true"
        @dragstart="onDragStart($event, item.type)"
      >
        <div class="item-icon">
          <el-icon><component :is="item.icon" /></el-icon>
        </div>
        <div class="item-info">
          <div class="item-name">{{ item.label }}</div>
          <div class="item-desc">{{ item.desc }}</div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { VideoPlay, CircleCheck, Checked, Operation, Setting } from '@element-plus/icons-vue'
import type { NodeType } from '@/types/workflow'

const nodeTypes: { type: NodeType; label: string; desc: string; icon: any }[] = [
  { type: 'start', label: '开始', desc: '流程起点', icon: VideoPlay },
  { type: 'end', label: '结束', desc: '流程终点', icon: CircleCheck },
  { type: 'approval', label: '审批', desc: '审批节点', icon: Checked },
  { type: 'work', label: '工作', desc: '工作节点', icon: Operation },
  { type: 'system', label: '系统', desc: '自动执行', icon: Setting },
]

const onDragStart = (event: DragEvent, nodeType: NodeType) => {
  if (event.dataTransfer) {
    event.dataTransfer.setData('application/vueflow', nodeType)
    event.dataTransfer.effectAllowed = 'move'
  }
}
</script>

<style scoped lang="scss">
.node-toolbar {
  width: 200px;
  background: #fff;
  border-right: 1px solid #e5e7eb;
  padding: 16px;
  display: flex;
  flex-direction: column;
  gap: 12px;
  flex-shrink: 0;
  overflow-y: auto;
}

.toolbar-title {
  font-size: 14px;
  font-weight: 600;
  color: #1f2937;
}

.toolbar-hint {
  font-size: 11px;
  color: #9ca3af;
  margin-top: -8px;
}

.toolbar-items {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.toolbar-item {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 10px 12px;
  border-radius: 10px;
  cursor: grab;
  transition: all 0.2s;
  border: 1px solid #e5e7eb;
  background: #f9fafb;

  &:hover {
    transform: translateY(-1px);
    box-shadow: 0 4px 12px rgba(0, 0, 0, 0.08);
  }

  &:active {
    cursor: grabbing;
    transform: scale(0.98);
  }

  .item-icon {
    width: 36px;
    height: 36px;
    border-radius: 8px;
    display: flex;
    align-items: center;
    justify-content: center;
    font-size: 18px;
    color: #fff;
    flex-shrink: 0;
  }

  .item-info {
    flex: 1;
    min-width: 0;
  }

  .item-name {
    font-size: 13px;
    font-weight: 600;
    color: #1f2937;
  }

  .item-desc {
    font-size: 11px;
    color: #9ca3af;
  }

  &.start .item-icon {
    background: #10b981;
  }
  &.end .item-icon {
    background: #6b7280;
  }
  &.approval .item-icon {
    background: #f59e0b;
  }
  &.work .item-icon {
    background: #3b82f6;
  }
  &.system .item-icon {
    background: #8b5cf6;
  }

  &.start:hover { border-color: #10b981; background: #ecfdf5; }
  &.end:hover { border-color: #6b7280; background: #f3f4f6; }
  &.approval:hover { border-color: #f59e0b; background: #fffbeb; }
  &.work:hover { border-color: #3b82f6; background: #eff6ff; }
  &.system:hover { border-color: #8b5cf6; background: #f5f3ff; }
}
</style>
