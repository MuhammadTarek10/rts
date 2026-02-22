<template>
  <div class="input-wrapper">
    <input
      v-model="value"
      :type="type"
      :placeholder="placeholder"
      :id="id"
      :required="required"
      :disabled="disabled"
      class="vargo-input"
      :class="{ 'is-invalid': invalid }"
    />
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
    return props.modelValue;
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
  padding: 0.75rem 1rem;
  border: 1px solid var(--border-color);
  border-radius: var(--radius-md);
  font-size: 1rem;
  font-family: var(--font-sans);
  transition: all var(--transition-fast);
  box-sizing: border-box;
  background-color: var(--bg-secondary);
  color: var(--text-primary);
}

.vargo-input::placeholder {
  color: var(--text-tertiary);
}

.vargo-input:focus {
  outline: none;
  border-color: var(--primary-color);
  background-color: var(--bg-tertiary);
  box-shadow: 0 0 0 3px rgba(99, 102, 241, 0.2);
}

.vargo-input.is-invalid {
  border-color: var(--danger-color);
}

.vargo-input.is-invalid:focus {
  box-shadow: 0 0 0 3px rgba(239, 68, 68, 0.2);
}

.vargo-input:disabled {
  background-color: rgba(255, 255, 255, 0.05);
  cursor: not-allowed;
  opacity: 0.7;
}
</style>
