<template>
  <div class="td-stat-tile" :class="[`td-stat-tile--${tone}`, { 'td-stat-tile--interactive': interactive, 'td-stat-tile--active': active }]">
    <div v-if="iconComponent || $slots.icon" class="td-stat-tile__icon">
      <slot name="icon">
        <component :is="iconComponent" />
      </slot>
    </div>
    <div class="td-stat-tile__body">
      <div class="td-stat-tile__label">{{ label }}</div>
      <div class="td-stat-tile__value-row">
        <div class="td-stat-tile__value">{{ value }}</div>
        <div
          v-if="delta !== undefined && delta !== null"
          class="td-stat-tile__delta"
          :class="{ 'is-up': delta > 0, 'is-down': delta < 0 }"
        >
          {{ delta > 0 ? '+' : '' }}{{ delta }}{{ deltaUnit }}
        </div>
      </div>
      <div v-if="hint" class="td-stat-tile__hint">{{ hint }}</div>
    </div>
  </div>
</template>

<script setup lang="ts">
import type { Component } from 'vue'

withDefaults(defineProps<{
  label: string
  value: number | string
  delta?: number | null
  deltaUnit?: string
  hint?: string
  iconComponent?: Component
  tone?: 'primary' | 'success' | 'warning' | 'danger' | 'info'
  interactive?: boolean
  active?: boolean
}>(), {
  tone: 'info',
  interactive: false,
  active: false,
  deltaUnit: '',
  delta: null,
  hint: '',
  iconComponent: undefined,
})
</script>

<style scoped lang="scss">
.td-stat-tile {
  display: flex;
  align-items: flex-start;
  gap: var(--td-space-3);
  padding: var(--td-space-4) var(--td-space-5);
  background: var(--td-bg-card);
  border-radius: var(--td-radius-lg);
  box-shadow: var(--td-elevation-1);
  transition: var(--td-transition-shadow), var(--td-transition-bg);

  &--interactive {
    cursor: pointer;

    &:hover {
      box-shadow: var(--td-elevation-3);
      background: var(--td-bg-card-hover);
    }
  }

  // 激活态 (作为筛选器被选中)
  &--active {
    box-shadow: var(--td-elevation-2), inset 3px 0 0 var(--td-color-primary);
    background: var(--td-tag-primary-bg);

    &.td-stat-tile--interactive:hover {
      box-shadow: var(--td-elevation-3), inset 3px 0 0 var(--td-color-primary);
      background: var(--td-tag-primary-bg);
    }
  }

  // 各色调激活态边条用对应颜色
  &--primary.td-stat-tile--active { box-shadow: var(--td-elevation-2), inset 3px 0 0 var(--td-color-primary); }
  &--success.td-stat-tile--active { box-shadow: var(--td-elevation-2), inset 3px 0 0 var(--td-color-success); background: var(--td-tag-success-bg); }
  &--warning.td-stat-tile--active { box-shadow: var(--td-elevation-2), inset 3px 0 0 var(--td-color-warning); background: var(--td-tag-warning-bg); }
  &--danger.td-stat-tile--active  { box-shadow: var(--td-elevation-2), inset 3px 0 0 var(--td-color-danger);  background: var(--td-tag-danger-bg); }
  &--info.td-stat-tile--active    { box-shadow: var(--td-elevation-2), inset 3px 0 0 var(--td-text-secondary); background: var(--td-tag-info-bg); }

  &__icon {
    width: 40px;
    height: 40px;
    border-radius: var(--td-radius-md);
    display: flex;
    align-items: center;
    justify-content: center;
    font-size: 20px;
    flex-shrink: 0;
    background: var(--td-tag-info-bg);
    color: var(--td-tag-info-text);
  }

  &__body {
    flex: 1;
    min-width: 0;
    display: flex;
    flex-direction: column;
    gap: 2px;
  }

  &__label {
    font-size: var(--td-font-sm);
    color: var(--td-text-secondary);
    font-weight: var(--td-weight-medium);
  }

  &__value-row {
    display: flex;
    align-items: baseline;
    gap: var(--td-space-2);
  }

  &__value {
    font-size: var(--td-font-2xl);
    font-weight: var(--td-weight-semibold);
    color: var(--td-text-primary);
    line-height: var(--td-leading-tight);
    letter-spacing: var(--td-tracking-tight);
  }

  &__delta {
    font-size: var(--td-font-sm);
    font-weight: var(--td-weight-medium);
    padding: 1px 6px;
    border-radius: var(--td-radius-xs);

    &.is-up {
      color: var(--td-color-success);
      background: var(--td-tag-success-bg);
    }

    &.is-down {
      color: var(--td-color-danger);
      background: var(--td-tag-danger-bg);
    }
  }

  &__hint {
    font-size: var(--td-font-xs);
    color: var(--td-text-placeholder);
    margin-top: var(--td-space-1);
  }

  &--primary &__icon { background: var(--td-tag-primary-bg);  color: var(--td-color-primary); }
  &--success &__icon { background: var(--td-tag-success-bg);  color: var(--td-color-success); }
  &--warning &__icon { background: var(--td-tag-warning-bg);  color: var(--td-color-warning); }
  &--danger &__icon  { background: var(--td-tag-danger-bg);   color: var(--td-color-danger); }
  &--info &__icon    { background: var(--td-tag-info-bg);     color: var(--td-tag-info-text); }
}
</style>
