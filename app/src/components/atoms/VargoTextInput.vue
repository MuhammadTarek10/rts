<template>
  <div class="input-wrapper">
    <input
      :id="id"
      v-model="value"
      :type="type"
      :placeholder="placeholder"
      :required="required"
      :disabled="disabled"
      class="vargo-input"
      :class="{ 'is-invalid': invalid }"
    >
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue';

interface Props {
  modelValue: string;
  type?: string;
  placeholder?: string;
  id?: string;
  required?: boolean;
  disabled?: boolean;
  invalid?: boolean;
}

const props = withDefaults(defineProps<Props>(), {
  type: 'text',
  required: false,
  disabled: false,
  invalid: false,
});

const emit = defineEmits<(e: 'update:modelValue', value: string) => void>();

const value = computed({
  get() {
    return props.modelValue ?? '';
  },
  set(value) {
    emit('update:modelValue', value);
  },
});
</script>

<style scoped>
.input-wrapper {
  position: relative;
  width: 100%;
}

.vargo-input {
  width: 100%;
  padding: 0.7rem 0.875rem;
  border: 1px solid var(--border-color);
  border-radius: var(--radius-md);
  font-size: 0.95rem;
  font-family: var(--font-sans);
  transition:
    border-color var(--transition-fast),
    box-shadow var(--transition-fast),
    background-color var(--transition-fast);
  box-sizing: border-box;
  background-color: var(--bg-color);
  color: var(--text-primary);
}

.vargo-input::placeholder {
  color: var(--text-tertiary);
}

.vargo-input:focus {
  outline: none;
  border-color: var(--primary-color);
  box-shadow: 0 0 0 3px rgba(212, 160, 83, 0.1);
}

.vargo-input.is-invalid {
  border-color: var(--danger-color);
}

.vargo-input.is-invalid:focus {
  box-shadow: 0 0 0 3px rgba(217, 72, 72, 0.1);
}

.vargo-input:disabled {
  background-color: var(--bg-tertiary);
  cursor: not-allowed;
  opacity: 0.5;
}
</style>
