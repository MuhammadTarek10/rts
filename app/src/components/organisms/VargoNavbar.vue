<template>
  <nav>
    <ul>
      <li>
        <router-link to="/">Home</router-link>
      </li>
      <li>
        <router-link to="/about">About</router-link>
      </li>
    </ul>
    <vargo-button @click="logout" variant="danger">Sign Out</vargo-button>
  </nav>
</template>

<script setup lang="ts">
import { useRouter } from 'vue-router';
import VargoButton from '../atoms/VargoButton.vue';
import { useUserStore } from '../../stores/user';

const router = useRouter();
const userStore = useUserStore();

const logout = () => {
  userStore.clearUser();
  router.push('/auth/sign-in');
};
</script>

<style scoped lang="less">
nav {
  display: flex;
  justify-content: space-between;
  position: sticky;
  top: 0;
  align-items: center;
  padding: 1rem 2rem;
  background-color: var(--bg-secondary);
  border-bottom: 1px solid var(--border-color);
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.1);

  ul {
    list-style: none;
    display: flex;
    gap: 2rem;
    padding: 0;
    margin: 0;
  }

  a {
    color: var(--text-secondary);
    text-decoration: none;
    font-weight: 500;
    transition: color 0.15s ease;
    padding: 0.5rem 0;
    position: relative;

    &:hover {
      color: var(--primary-color);
    }

    &.router-link-active {
      color: var(--primary-color);
      font-weight: 600;

      &::after {
        content: '';
        position: absolute;
        bottom: 0;
        left: 0;
        width: 100%;
        height: 2px;
        background-color: var(--primary-color);
        border-radius: 1px;
      }
    }
  }
}
</style>
