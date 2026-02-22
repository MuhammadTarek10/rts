<template>
  <div class="form-field">
    <label v-if="label" :for="htmlFor" class="form-label">{{ label }}</label>
    <slot />
    <span v-if="error" class="error-message">{{ error }}</span>
  </div>
</template>

<script setup lang="ts">
defineProps<{
  label: string;
  htmlFor?: string;
  error?: string;
}>();
</script>

<style scoped>
.form-field {
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
  margin-bottom: 1.5rem;
  width: 100%;
}

.form-label {
  font-weight: 500;
  font-size: 0.9rem;
  color: var(--text-secondary);
  margin-left: 2px;
}

.error-message {
  font-size: 0.85rem;
  color: var(--danger-color);
  margin-top: 0.25rem;
  animation: fadeIn 0.3s ease;
}

@keyframes fadeIn {
  from {
    opacity: 0;
    transform: translateY(-5px);
  }
  to {
    opacity: 1;
    transform: translateY(0);
  }
}
</style>
