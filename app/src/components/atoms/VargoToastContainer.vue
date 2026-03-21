<template>
  <div class="toast-container">
    <transition-group name="toast">
      <div
        v-for="toast in toasts"
        :key="toast.id"
        class="toast"
        :class="toast.type"
      >
        <span class="toast-message">{{ toast.message }}</span>
        <button
          class="toast-close"
          @click="removeToast(toast.id)"
        >
          &times;
        </button>
      </div>
    </transition-group>
  </div>
</template>

<script setup lang="ts">
import { useToast } from '@/composables/useToast';

const { toasts, removeToast } = useToast();
</script>

<style scoped lang="less">
.toast-container {
  position: fixed;
  top: 1.25rem;
  right: 1.25rem;
  z-index: 9999;
  display: flex;
  flex-direction: column;
  gap: 0.625rem;
  max-width: 380px;
}

.toast {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 0.75rem;
  padding: 0.8rem 1rem;
  border-radius: var(--radius-md);
  font-size: 0.88rem;
  font-family: var(--font-sans);
  backdrop-filter: blur(12px);
  border: 1px solid;
}

.toast.success {
  background-color: rgba(62, 186, 110, 0.12);
  border-color: rgba(62, 186, 110, 0.2);
  color: var(--success-color);
}

.toast.error {
  background-color: rgba(217, 72, 72, 0.12);
  border-color: rgba(217, 72, 72, 0.2);
  color: var(--danger-color);
}

.toast.info {
  background-color: rgba(90, 143, 212, 0.12);
  border-color: rgba(90, 143, 212, 0.2);
  color: var(--info-color);
}

.toast.warning {
  background-color: rgba(229, 168, 59, 0.12);
  border-color: rgba(229, 168, 59, 0.2);
  color: var(--warning-color);
}

.toast-message {
  flex: 1;
  line-height: 1.4;
}

.toast-close {
  background: none;
  border: none;
  color: inherit;
  font-size: 1.1rem;
  cursor: pointer;
  opacity: 0.5;
  padding: 0;
  line-height: 1;
  transition: opacity var(--transition-fast);
}

.toast-close:hover {
  opacity: 1;
}

.toast-enter-active {
  transition: all 0.35s cubic-bezier(0.16, 1, 0.3, 1);
}

.toast-leave-active {
  transition: all 0.25s ease;
}

.toast-enter-from {
  opacity: 0;
  transform: translateX(40px) scale(0.95);
}

.toast-leave-to {
  opacity: 0;
  transform: translateX(40px) scale(0.95);
}
</style>
