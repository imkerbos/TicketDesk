<template>
  <section class="td-section">
    <header v-if="title || $slots.actions" class="td-section__header">
      <div class="td-section__title-group">
        <h2 v-if="title" class="td-section__title">{{ title }}</h2>
        <p v-if="description" class="td-section__description">{{ description }}</p>
      </div>
      <div v-if="$slots.actions" class="td-section__actions">
        <slot name="actions"></slot>
      </div>
    </header>
    <div class="td-section__body">
      <slot></slot>
    </div>
  </section>
</template>

<script setup lang="ts">
defineProps<{
  title?: string
  description?: string
}>()
</script>

<style scoped lang="scss">
.td-section {
  display: flex;
  flex-direction: column;
  gap: var(--td-space-3);

  &__header {
    display: flex;
    justify-content: space-between;
    align-items: flex-start;
    gap: var(--td-space-4);
  }

  &__title-group {
    display: flex;
    flex-direction: column;
    gap: 2px;
    min-width: 0;
  }

  &__title {
    font-size: var(--td-font-lg);
    font-weight: var(--td-weight-semibold);
    color: var(--td-text-primary);
    margin: 0;
    line-height: var(--td-leading-tight);
  }

  &__description {
    font-size: var(--td-font-sm);
    color: var(--td-text-secondary);
    margin: 0;
    line-height: var(--td-leading-normal);
  }

  &__actions {
    display: flex;
    align-items: center;
    gap: var(--td-space-2);
    flex-shrink: 0;
  }
}
</style>
