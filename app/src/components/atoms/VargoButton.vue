<template>
  <button
    class="atom-button"
    :class="[variant, { 'is-disabled': disabled }]"
    :disabled="disabled"
    @click="handleClick"
  >
    <slot>{{ text }}</slot>
  </button>
</template>

<script setup lang="ts">
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
});

const emit = defineEmits(['click']);

const handleClick = (event: MouseEvent) => {
  if (!props.disabled) {
    emit('click', event);
  }
};
</script>

<style scoped>
.atom-button {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  padding: 0.75rem 1.5rem;
  border-radius: var(--radius-md);
  border: 1px solid transparent;
  font-size: 1rem;
  font-weight: 600;
  font-family: inherit;
  cursor: pointer;
  transition: all var(--transition-fast);
  line-height: 1;
}

.atom-button:disabled,
.atom-button.is-disabled {
  opacity: 0.6;
  cursor: not-allowed;
  filter: grayscale(100%);
}

/* Primary Variant */
.atom-button.primary {
  background-color: var(--primary-color);
  color: white;
  box-shadow: 0 2px 4px rgba(99, 102, 241, 0.3);
}
.atom-button.primary:hover:not(:disabled) {
  background-color: var(--primary-hover);
  transform: translateY(-1px);
  box-shadow: 0 4px 6px rgba(99, 102, 241, 0.4);
}
.atom-button.primary:active:not(:disabled) {
  transform: translateY(0);
}

/* Secondary Variant */
.atom-button.secondary {
  background-color: transparent;
  color: var(--text-primary);
  border: 1px solid var(--border-color);
  background-color: var(--bg-secondary);
}
.atom-button.secondary:hover:not(:disabled) {
  border-color: var(--primary-color);
  color: var(--primary-color);
}

/* Danger Variant */
.atom-button.danger {
  background-color: var(--danger-color);
  color: white;
}
.atom-button.danger:hover:not(:disabled) {
  background-color: #dc2626; /* Darker red */
}

/* Ghost Variant */
.atom-button.ghost {
  background-color: transparent;
  color: var(--text-secondary);
}
.atom-button.ghost:hover:not(:disabled) {
  color: var(--text-primary);
  background-color: rgba(255, 255, 255, 0.05);
}
</style>
