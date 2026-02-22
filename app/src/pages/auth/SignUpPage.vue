<template>
  <div class="auth-page">
    <div class="page-title">
      <h2>Create Account</h2>
      <p>Join us and start your journey</p>
    </div>

    <sign-up-form @submit="handleSignUp" />
  </div>
</template>

<script setup lang="ts">
import { useRouter } from 'vue-router';
import { useUserStore } from '../../stores/user';
import SignUpForm from '../../components/organisms/SignUpForm.vue';
import type { SignUpFormData } from '../../types';

const router = useRouter();
const userStore = useUserStore();

const handleSignUp = (payload: SignUpFormData) => {
  // Simulate registration
  userStore.setUser({ email: payload.email, id: '2', name: payload.name });
  router.push('/');
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
