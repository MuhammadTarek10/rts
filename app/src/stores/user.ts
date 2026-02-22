import { defineStore } from 'pinia';
import { ref } from 'vue';

export interface IUser {
  id?: string | null;
  name?: string | null;
  email?: string | null;
}

export const useUserStore = defineStore(
  'user',
  () => {
    const user = ref<IUser | null>(null);

    function setUser(userData: IUser) {
      user.value = userData;
    }

    async function fetchUser() {
      try {
        console.log('Fetching user from backend...');
      } catch (error) {
        clearUser();
        console.error('Failed to fetch user:', error);
      }
    }

    function clearUser() {
      user.value = null;
    }

    return { user, setUser, fetchUser, clearUser };
  },
  {
    persist: true,
  }
);
