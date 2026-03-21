import { defineStore } from 'pinia';
import { ref, computed } from 'vue';
import { authService } from '@/services/auth.service';
import type { User, SignInPayload, SignUpPayload } from '@/types';

export const useUserStore = defineStore(
  'user',
  () => {
    const user = ref<User | null>(null);
    const isLoading = ref(false);
    const isInitialized = ref(false);

    const isAuthenticated = computed(() => user.value !== null);

    const displayName = computed(() => {
      if (!user.value) return '';
      const p = user.value.profile;
      if (p?.first_name || p?.last_name) {
        return [p.first_name, p.last_name].filter(Boolean).join(' ');
      }
      return user.value.email;
    });

    function setUser(data: User) {
      user.value = data;
    }

    function clearUser() {
      user.value = null;
    }

    async function fetchUser(): Promise<boolean> {
      try {
        isLoading.value = true;
        const response = await authService.getProfile();
        user.value = response.data.data;
        return true;
      } catch {
        clearUser();
        return false;
      } finally {
        isLoading.value = false;
        isInitialized.value = true;
      }
    }

    async function signIn(payload: SignInPayload): Promise<void> {
      await authService.signIn(payload);
      await fetchUser();
    }

    async function signUp(payload: SignUpPayload): Promise<void> {
      await authService.signUp(payload);
      await fetchUser();
    }

    async function signOut(): Promise<void> {
      try {
        await authService.signOut();
      } catch {
        // Even if the API call fails, clear local state
      } finally {
        clearUser();
      }
    }

    return {
      user,
      isLoading,
      isInitialized,
      isAuthenticated,
      displayName,
      setUser,
      clearUser,
      fetchUser,
      signIn,
      signUp,
      signOut,
    };
  },
  {
    persist: true,
  },
);
