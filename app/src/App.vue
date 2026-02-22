<template>
  <component :is="layout">
    <router-view />
  </component>
</template>

<script setup lang="ts">
import { computed, onMounted } from 'vue';
import { useRoute } from 'vue-router';
import { useUserStore } from './stores/user';
import MainLayout from './layouts/MainLayout.vue';
import AuthLayout from './layouts/AuthLayout.vue';

const route = useRoute();
const userStore = useUserStore();

onMounted(() => {
  userStore.fetchUser();
});

// Map layout names to components.
// You can add more layouts here (e.g., AuthLayout: AuthLayoutComponent)
const layouts = {
  MainLayout,
  AuthLayout,
};

const layout = computed(() => {
  const layoutName = (route.meta.layout as string) || 'MainLayout';
  return layouts[layoutName as keyof typeof layouts] || MainLayout;
});
</script>
