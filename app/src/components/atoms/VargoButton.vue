<template>
  <button
    class="atom-button"
    :class="[variant, { 'is-disabled': isDisabled, 'is-loading': loading }]"
    :disabled="isDisabled"
    @click="handleClick"
  >
    <span
      v-if="loading"
      class="spinner"
    />
    <slot>{{ text }}</slot>
  </button>
</template>

<script setup lang="ts">
import { computed } from 'vue';

const props = defineProps({
  text: {
    type: String,
    default: 'Button',
  },
  variant: {
    type: String,
    default: 'primary',
    validator: (value: string) =>
      ['primary', 'secondary', 'danger', 'ghost'].includes(value),
  },
  disabled: {
    type: Boolean,
    default: false,
  },
  loading: {
    type: Boolean,
    default: false,
  },
});

const emit = defineEmits(['click']);

const isDisabled = computed(() => props.disabled || props.loading);

const handleClick = (event: MouseEvent) => {
  if (!isDisabled.value) {
    emit('click', event);
  }
};
</script>

<style scoped>
.atom-button {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: 0.5rem;
  padding: 0.7rem 1.5rem;
  border-radius: var(--radius-md);
  border: 1px solid transparent;
  font-size: 0.95rem;
  font-weight: 600;
  font-family: var(--font-sans);
  cursor: pointer;
  transition:
    background-color var(--transition-fast),
    border-color var(--transition-fast),
    color var(--transition-fast),
    box-shadow var(--transition-fast),
    transform var(--transition-fast);
  line-height: 1;
  letter-spacing: 0.01em;
}

.atom-button:disabled,
.atom-button.is-disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.atom-button.is-loading {
  opacity: 0.7;
}

/* Spinner */
.spinner {
  width: 14px;
  height: 14px;
  border: 2px solid rgba(255, 255, 255, 0.2);
  border-top-color: currentColor;
  border-radius: 50%;
  animation: spin 0.6s linear infinite;
}

@keyframes spin {
  to {
    transform: rotate(360deg);
  }
}

/* Primary — Amber */
.atom-button.primary {
  background-color: var(--primary-color);
  color: #0a0a0a;
  font-weight: 700;
}
.atom-button.primary:hover:not(:disabled) {
  background-color: var(--primary-hover);
  box-shadow: 0 0 20px rgba(212, 160, 83, 0.25);
}
.atom-button.primary:active:not(:disabled) {
  transform: scale(0.98);
}

/* Secondary */
.atom-button.secondary {
  background-color: transparent;
  color: var(--text-primary);
  border: 1px solid var(--border-color);
}
.atom-button.secondary:hover:not(:disabled) {
  border-color: var(--primary-color);
  color: var(--primary-color);
  background-color: var(--primary-muted);
}

/* Danger */
.atom-button.danger {
  background-color: rgba(217, 72, 72, 0.12);
  color: var(--danger-color);
  border: 1px solid rgba(217, 72, 72, 0.25);
}
.atom-button.danger:hover:not(:disabled) {
  background-color: rgba(217, 72, 72, 0.2);
  border-color: rgba(217, 72, 72, 0.4);
}

.atom-button.danger .spinner {
  border-color: rgba(217, 72, 72, 0.2);
  border-top-color: var(--danger-color);
}

/* Ghost */
.atom-button.ghost {
  background-color: transparent;
  color: var(--text-secondary);
}
.atom-button.ghost:hover:not(:disabled) {
  color: var(--text-primary);
  background-color: rgba(255, 255, 255, 0.04);
}
</style>
