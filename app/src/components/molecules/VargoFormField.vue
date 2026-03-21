<template>
  <div class="form-field">
    <label
      v-if="label"
      :for="htmlFor"
      class="form-label"
    >{{ label }}</label>
    <slot />
    <span
      v-if="error"
      class="error-message"
    >{{ error }}</span>
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
  gap: 0.4rem;
  margin-bottom: 1.25rem;
  width: 100%;
}

.form-label {
  font-weight: 500;
  font-size: 0.85rem;
  color: var(--text-secondary);
  letter-spacing: 0.02em;
  text-transform: uppercase;
}

.error-message {
  font-size: 0.82rem;
  color: var(--danger-color);
  animation: errorReveal 0.25s ease;
}

@keyframes errorReveal {
  from {
    opacity: 0;
    transform: translateY(-4px);
  }
  to {
    opacity: 1;
    transform: translateY(0);
  }
}
</style>
