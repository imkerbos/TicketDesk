<template>
  <header class="td-page-header">
    <div class="td-page-header__main">
      <div v-if="$slots.breadcrumb" class="td-page-header__breadcrumb">
        <slot name="breadcrumb"></slot>
      </div>
      <div class="td-page-header__title-row">
        <div class="td-page-header__title-group">
          <slot name="leading"></slot>
          <div class="td-page-header__title-text">
            <h1 class="td-page-header__title">
              <slot name="title"></slot>
            </h1>
            <p v-if="$slots.subtitle" class="td-page-header__subtitle">
              <slot name="subtitle"></slot>
            </p>
          </div>
        </div>
        <div v-if="$slots.actions" class="td-page-header__actions">
          <slot name="actions"></slot>
        </div>
      </div>
    </div>
  </header>
</template>

<script setup lang="ts">
// slot 驱动, 无 props
</script>

<style scoped lang="scss">
.td-page-header {
  background: var(--td-bg-card);
  border-radius: var(--td-radius-lg);
  box-shadow: var(--td-elevation-1);
  padding: var(--td-space-4) var(--td-space-6);
  margin-bottom: var(--td-space-5);

  &__breadcrumb {
    margin-bottom: var(--td-space-2);
    font-size: var(--td-font-sm);
    color: var(--td-text-placeholder);
  }

  &__title-row {
    display: flex;
    justify-content: space-between;
    align-items: center;
    gap: var(--td-space-4);
  }

  &__title-group {
    display: flex;
    align-items: center;
    gap: var(--td-space-3);
    min-width: 0;
  }

  &__title-text {
    display: flex;
    flex-direction: column;
    gap: 2px;
    min-width: 0;
  }

  &__title {
    font-size: var(--td-font-2xl);
    font-weight: var(--td-weight-semibold);
    color: var(--td-text-primary);
    line-height: var(--td-leading-tight);
    margin: 0;
    letter-spacing: var(--td-tracking-tight);
  }

  &__subtitle {
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

@media (max-width: 768px) {
  .td-page-header {
    padding: var(--td-space-3) var(--td-space-4);

    &__title {
      font-size: var(--td-font-xl);
    }
  }
}
</style>
