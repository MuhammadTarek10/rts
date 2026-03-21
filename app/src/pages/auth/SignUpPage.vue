<template>
  <div class="auth-page">
    <div class="page-title">
      <h2>Create Account</h2>
      <p>Join us and start your journey</p>
    </div>

    <SignUpForm
      :loading="isLoading"
      :error="errorMessage"
      @submit="handleSignUp"
    />
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue';
import { useRouter } from 'vue-router';
import { useUserStore } from '@/stores/user';
import { useToast } from '@/composables/useToast';
import { getErrorMessage } from '@/utils/error';
import SignUpForm from '@/components/organisms/SignUpForm.vue';
import type { SignUpFormData } from '@/types';

const router = useRouter();
const userStore = useUserStore();
const toast = useToast();

const isLoading = ref(false);
const errorMessage = ref<string | null>(null);

const handleSignUp = async (payload: SignUpFormData) => {
  isLoading.value = true;
  errorMessage.value = null;

  try {
    await userStore.signUp(payload);
    toast.success('Account created successfully');
    router.push('/');
  } catch (err) {
    errorMessage.value = getErrorMessage(err);
    toast.error(errorMessage.value);
  } finally {
    isLoading.value = false;
  }
};
</script>

<style scoped>
.page-title {
  text-align: center;
  margin-bottom: 1.75rem;
}

.page-title h2 {
  font-size: 1.5rem;
  margin-bottom: 0.35rem;
}

.page-title p {
  font-size: 0.9rem;
  color: var(--text-tertiary);
}
</style>
