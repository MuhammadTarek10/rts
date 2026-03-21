<template>
  <nav class="navbar">
    <div class="navbar-inner">
      <div class="navbar-left">
        <router-link
          to="/"
          class="navbar-brand"
        >
          Vargo
        </router-link>
        <div class="navbar-divider" />
        <ul class="navbar-links">
          <li>
            <router-link to="/">
              Home
            </router-link>
          </li>
          <li>
            <router-link to="/about">
              About
            </router-link>
          </li>
        </ul>
      </div>
      <div class="navbar-right">
        <span class="user-name">{{ userStore.displayName }}</span>
        <router-link
          to="/profile"
          class="profile-link"
        >
          Profile
        </router-link>
        <button
          class="sign-out-btn"
          :disabled="isLoggingOut"
          @click="logout"
        >
          <span
            v-if="isLoggingOut"
            class="btn-spinner"
          />
          Sign Out
        </button>
      </div>
    </div>
  </nav>
</template>

<script setup lang="ts">
import { ref } from 'vue';
import { useRouter } from 'vue-router';
import { useUserStore } from '@/stores/user';

const router = useRouter();
const userStore = useUserStore();
const isLoggingOut = ref(false);

const logout = async () => {
  isLoggingOut.value = true;
  await userStore.signOut();
  isLoggingOut.value = false;
  router.push('/auth/sign-in');
};
</script>

<style scoped lang="less">
.navbar {
  position: sticky;
  top: 0;
  z-index: 100;
  background-color: rgba(10, 10, 12, 0.85);
  backdrop-filter: blur(12px);
  border-bottom: 1px solid var(--border-color);
  animation: slideDown 0.4s ease;
}

.navbar-inner {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 0 2rem;
  height: 56px;
  max-width: 1200px;
  margin: 0 auto;
}

.navbar-left {
  display: flex;
  align-items: center;
  gap: 1rem;
}

.navbar-brand {
  font-family: var(--font-display);
  font-size: 1.35rem;
  font-weight: 400;
  color: var(--primary-color);
  text-decoration: none;
  letter-spacing: -0.02em;

  &:hover {
    color: var(--primary-hover);
  }
}

.navbar-divider {
  width: 1px;
  height: 20px;
  background: var(--border-color);
}

.navbar-links {
  list-style: none;
  display: flex;
  gap: 0.25rem;
  padding: 0;
  margin: 0;

  a {
    color: var(--text-tertiary);
    text-decoration: none;
    font-weight: 500;
    font-size: 0.9rem;
    padding: 0.4rem 0.75rem;
    border-radius: var(--radius-sm);
    transition:
      color var(--transition-fast),
      background-color var(--transition-fast);

    &:hover {
      color: var(--text-primary);
      background-color: rgba(255, 255, 255, 0.04);
    }

    &.router-link-active {
      color: var(--primary-color);
    }
  }
}

.navbar-right {
  display: flex;
  align-items: center;
  gap: 1rem;
}

.user-name {
  color: var(--text-tertiary);
  font-size: 0.85rem;
  font-weight: 500;
}

.profile-link {
  color: var(--text-secondary);
  font-size: 0.85rem;
  font-weight: 500;
  text-decoration: none;
  transition: color var(--transition-fast);

  &:hover {
    color: var(--primary-color);
  }
}

.sign-out-btn {
  display: inline-flex;
  align-items: center;
  gap: 0.375rem;
  background: transparent;
  border: 1px solid var(--border-color);
  color: var(--text-secondary);
  font-family: var(--font-sans);
  font-size: 0.82rem;
  font-weight: 500;
  padding: 0.35rem 0.75rem;
  border-radius: var(--radius-sm);
  cursor: pointer;
  transition:
    border-color var(--transition-fast),
    color var(--transition-fast);

  &:hover:not(:disabled) {
    border-color: var(--danger-color);
    color: var(--danger-color);
  }

  &:disabled {
    opacity: 0.5;
    cursor: not-allowed;
  }
}

.btn-spinner {
  width: 10px;
  height: 10px;
  border: 1.5px solid rgba(255, 255, 255, 0.15);
  border-top-color: currentColor;
  border-radius: 50%;
  animation: spin 0.6s linear infinite;
}

@keyframes slideDown {
  from {
    opacity: 0;
    transform: translateY(-8px);
  }
  to {
    opacity: 1;
    transform: translateY(0);
  }
}

@keyframes spin {
  to {
    transform: rotate(360deg);
  }
}
</style>
