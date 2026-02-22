<template>
  <div class="auth-page">
    <div class="page-title">
      <h2>Sign In</h2>
      <p>Please enter your credentials to continue</p>
    </div>

    <sign-in-form @submit="handleSignIn" />

    <template v-if="true">
      <!-- Logic for footer links to be teleported or rendered in the layout slot -->
    </template>
  </div>

  <!-- In Vue 3, we can use Teleport if we want to render into the layout's footer slot from here, 
       but for simplicity we'll just put links below the form inside the main slot 
       since AuthLayout renders <slot /> then <slot name="footer" /> -->
</template>

<script setup lang="ts">
import { useRouter } from 'vue-router';
import { useUserStore } from '../../stores/user';
import SignInForm from '../../components/organisms/SignInForm.vue';
import type { SignInFormData } from '../../types';

const router = useRouter();
const userStore = useUserStore();

const handleSignIn = (payload: SignInFormData) => {
  userStore.setUser({ email: payload.email, id: '1', name: 'User' });
  const redirect = (router.currentRoute.value.query.redirect as string) || '/';
  router.push(redirect);
};
</script>

<style scoped>
.page-title {
  text-align: center;
  margin-bottom: 2rem;
}

.page-title h2 {
  font-size: 1.5rem;
  margin-bottom: 0.5rem;
}

.page-title p {
  font-size: 0.9rem;
  color: var(--text-secondary);
}
</style>
