<template>
  <div
    v-if="!userStore.isInitialized"
    class="app-loading"
  >
    <div class="app-loading-content">
      <h1 class="app-loading-logo">
        Vargo
      </h1>
      <div class="app-loading-bar">
        <div class="app-loading-bar-fill" />
      </div>
    </div>
  </div>
  <template v-else>
    <component :is="layout">
      <router-view />
    </component>
  </template>
  <vargo-toast-container />
</template>

<script setup lang="ts">
import { computed } from 'vue';
import { useRoute } from 'vue-router';
import { useUserStore } from '@/stores/user';
import MainLayout from '@/layouts/MainLayout.vue';
import AuthLayout from '@/layouts/AuthLayout.vue';
import VargoToastContainer from '@/components/atoms/VargoToastContainer.vue';

const route = useRoute();
const userStore = useUserStore();

const layouts = { MainLayout, AuthLayout };

const layout = computed(() => {
  const layoutName = (route.meta.layout as string) || 'MainLayout';
  return layouts[layoutName as keyof typeof layouts] || MainLayout;
});
</script>

<style scoped>
.app-loading {
  min-height: 100vh;
  display: flex;
  align-items: center;
  justify-content: center;
  background: var(--bg-color);
}

.app-loading-content {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 2rem;
}

.app-loading-logo {
  font-family: var(--font-display);
  font-size: 3rem;
  font-weight: 400;
  color: var(--primary-color);
  letter-spacing: -0.02em;
  animation: pulse 2s ease-in-out infinite;
}

.app-loading-bar {
  width: 120px;
  height: 2px;
  background: var(--bg-tertiary);
  border-radius: 1px;
  overflow: hidden;
}

.app-loading-bar-fill {
  width: 40%;
  height: 100%;
  border-radius: 1px;
  animation: shimmer 1.5s ease-in-out infinite;
  background: linear-gradient(
    90deg,
    transparent,
    var(--primary-color),
    transparent
  );
  background-size: 200% 100%;
}
</style>
