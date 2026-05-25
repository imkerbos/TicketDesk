<template>
  <div class="td-empty-state" :class="`td-empty-state--${tone}`">
    <div class="td-empty-state__icon">
      <component :is="iconComponent" />
    </div>
    <h3 class="td-empty-state__title">{{ title }}</h3>
    <p v-if="description" class="td-empty-state__description">{{ description }}</p>
    <div v-if="$slots.default" class="td-empty-state__actions">
      <slot></slot>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { Box, Search, WarningFilled, DataBoard, Folder } from '@element-plus/icons-vue'

type Preset = 'no-data' | 'no-result' | 'error' | 'no-permission' | 'first-time'

const props = withDefaults(defineProps<{
  preset?: Preset
  title: string
  description?: string
  tone?: 'neutral' | 'primary' | 'warning' | 'danger'
}>(), {
  preset: 'no-data',
  tone: 'neutral',
  description: '',
})

const iconComponent = computed(() => {
  switch (props.preset) {
    case 'no-result': return Search
    case 'error': return WarningFilled
    case 'no-permission': return Folder
    case 'first-time': return DataBoard
    case 'no-data':
    default: return Box
  }
})
</script>

<style scoped lang="scss">
.td-empty-state {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: var(--td-space-12) var(--td-space-6);
  gap: var(--td-space-3);
  text-align: center;
  color: var(--td-text-secondary);

  &__icon {
    width: 72px;
    height: 72px;
    border-radius: var(--td-radius-2xl);
    display: flex;
    align-items: center;
    justify-content: center;
    margin-bottom: var(--td-space-2);
    font-size: 32px;
    color: var(--td-text-placeholder);
    background: var(--td-bg-section);
    transition: var(--td-transition-bg), var(--td-transition-color);
  }

  &__title {
    font-size: var(--td-font-md);
    font-weight: var(--td-weight-medium);
    color: var(--td-text-regular);
    margin: 0;
  }

  &__description {
    font-size: var(--td-font-sm);
    color: var(--td-text-placeholder);
    margin: 0;
    max-width: 380px;
    line-height: var(--td-leading-normal);
  }

  &__actions {
    margin-top: var(--td-space-2);
    display: flex;
    gap: var(--td-space-2);
  }

  &--primary &__icon { background: var(--td-tag-primary-bg); color: var(--td-color-primary); }
  &--warning &__icon { background: var(--td-tag-warning-bg); color: var(--td-color-warning); }
  &--danger &__icon  { background: var(--td-tag-danger-bg);  color: var(--td-color-danger); }
}
</style>
