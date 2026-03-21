<template>
  <div class="home">
    <section class="hero">
      <div class="hero-content">
        <p class="hero-eyebrow">
          Dashboard
        </p>
        <h1>
          Welcome back,
          <span class="highlight">{{ userStore.displayName }}</span>
        </h1>
        <p class="hero-subtitle">
          Manage your retail operations from a single command center.
        </p>
      </div>
    </section>
  </div>
</template>

<script setup lang="ts">
import { useUserStore } from '@/stores/user';

const userStore = useUserStore();
</script>

<style scoped>
.home {
  padding-bottom: 2rem;
}

/* Hero */
.hero {
  padding: 3rem 0 4rem;
  animation: fadeUp 0.5s ease both;
}

.hero-eyebrow {
  font-size: 0.8rem;
  text-transform: uppercase;
  letter-spacing: 0.15em;
  color: var(--primary-color);
  margin-bottom: 0.75rem;
  font-weight: 600;
}

.hero h1 {
  font-size: 2.75rem;
  margin-bottom: 1rem;
}

.highlight {
  color: var(--primary-color);
}

.hero-subtitle {
  font-size: 1.15rem;
  color: var(--text-secondary);
  max-width: 540px;
  line-height: 1.6;
}

/* Section headers */
.section-header {
  display: flex;
  align-items: center;
  gap: 1rem;
  margin-bottom: 1.5rem;

  h2 {
    font-size: 1.35rem;
    margin-bottom: 0;
    white-space: nowrap;
  }
}

.section-line {
  flex: 1;
  height: 1px;
  background: var(--border-color);
}
</style>
