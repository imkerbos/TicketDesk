<template>
  <div
    class="td-card"
    :class="[
      `td-card--p-${padding}`,
      `td-card--e-${elevation}`,
      { 'td-card--hoverable': hoverable, 'td-card--interactive': interactive },
    ]"
    v-bind="$attrs"
  >
    <div v-if="$slots.header" class="td-card__header">
      <slot name="header"></slot>
    </div>
    <div class="td-card__body">
      <slot></slot>
    </div>
    <div v-if="$slots.footer" class="td-card__footer">
      <slot name="footer"></slot>
    </div>
  </div>
</template>

<script setup lang="ts">
withDefaults(defineProps<{
  padding?: 'none' | 'sm' | 'md' | 'lg'
  elevation?: 0 | 1 | 2 | 3
  hoverable?: boolean
  interactive?: boolean
}>(), {
  padding: 'md',
  elevation: 1,
  hoverable: false,
  interactive: false,
})
</script>

<style scoped lang="scss">
.td-card {
  background: var(--td-bg-card);
  border-radius: var(--td-radius-lg);
  transition: var(--td-transition-shadow), var(--td-transition-bg);
  position: relative;

  &--p-none { .td-card__body { padding: 0; } }
  &--p-sm   { .td-card__body { padding: var(--td-space-3); } }
  &--p-md   { .td-card__body { padding: var(--td-space-5); } }
  &--p-lg   { .td-card__body { padding: var(--td-space-6); } }

  &--e-0 { box-shadow: var(--td-elevation-0); }
  &--e-1 { box-shadow: var(--td-elevation-1); }
  &--e-2 { box-shadow: var(--td-elevation-2); }
  &--e-3 { box-shadow: var(--td-elevation-3); }

  &--hoverable:hover {
    box-shadow: var(--td-elevation-3);
  }

  &--interactive {
    cursor: pointer;

    &:hover {
      background: var(--td-bg-card-hover);
    }
  }

  &__header {
    padding: var(--td-space-4) var(--td-space-5);
    border-bottom: 1px solid var(--td-border-color);
    font-size: var(--td-font-md);
    font-weight: var(--td-weight-semibold);
    color: var(--td-text-primary);
  }

  &__footer {
    padding: var(--td-space-3) var(--td-space-5);
    border-top: 1px solid var(--td-border-color);
    background: var(--td-bg-section);
    border-bottom-left-radius: inherit;
    border-bottom-right-radius: inherit;
  }
}
</style>
