<template>
  <div class="auth-page">
    <div class="page-title">
      <h2>Sign In</h2>
      <p>Enter your credentials to continue</p>
    </div>

    <SignInForm
      :loading="isLoading"
      :error="errorMessage"
      @submit="handleSignIn"
    />
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue';
import { useRouter } from 'vue-router';
import { useUserStore } from '@/stores/user';
import { getErrorMessage } from '@/utils/error';
import SignInForm from '@/components/organisms/SignInForm.vue';
import type { SignInFormData } from '@/types';
import { push } from 'notivue';

const router = useRouter();
const userStore = useUserStore();

const isLoading = ref(false);
const errorMessage = ref<string | null>(null);

function isInternalRedirect(path: string): boolean {
  return path.startsWith('/') && !path.startsWith('//');
}

const handleSignIn = async (payload: SignInFormData) => {
  isLoading.value = true;
  errorMessage.value = null;

  try {
    await userStore.signIn(payload);
    push.success('Signed in successfully');
    const redirect = router.currentRoute.value.query.redirect as string;
    router.push(redirect && isInternalRedirect(redirect) ? redirect : '/');
  } catch (err) {
    errorMessage.value = getErrorMessage(err, 'Invalid email or password');
    push.error(errorMessage.value);
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
