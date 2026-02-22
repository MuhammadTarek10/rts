<template>
  <div class="home">
    <section class="hero">
      <div class="hero-content">
        <h1>Building the Future with <span class="highlight">Vargo</span></h1>
        <p class="hero-subtitle">
          Experience the next generation of web development with our
          cutting-edge platform using Vue 3 and TypeScript.
        </p>
        <div class="hero-actions">
          <vargo-button @click="router.push('/about')" variant="primary"
            >Get Started</vargo-button
          >
          <vargo-button
            @click="router.push('/auth/sign-in')"
            variant="secondary"
            >Sign In</vargo-button
          >
        </div>
      </div>
      <div class="hero-visual">
        <div class="logo-container">
          <img src="/vite.svg" class="logo" alt="Vite logo" />
          <img src="../assets/vue.svg" class="logo vue" alt="Vue logo" />
        </div>
      </div>
    </section>

    <section class="features">
      <div class="feature-card">
        <h3>Lightning Fast</h3>
        <p>Powered by Vite for instant server start and lightning fast HMR.</p>
      </div>
      <div class="feature-card">
        <h3>Type Safe</h3>
        <p>Built with TypeScript for robust and maintainable code.</p>
      </div>
      <div class="feature-card">
        <h3>Component Driven</h3>
        <p>Modular architecture using Vue 3 Composition API.</p>
      </div>
    </section>
  </div>
</template>

<script setup lang="ts">
import { useRouter } from 'vue-router';
import VargoButton from '../components/atoms/VargoButton.vue';

const router = useRouter();
</script>

<style scoped>
.hero {
  display: flex;
  flex-direction: column;
  align-items: center;
  text-align: center;
  padding: 4rem 1rem;
  gap: 3rem;
  background: radial-gradient(
    circle at center,
    rgba(99, 102, 241, 0.1) 0%,
    transparent 70%
  );
}

.hero-content {
  max-width: 800px;
}

.highlight {
  color: var(--primary-color);
  background: linear-gradient(
    120deg,
    var(--primary-color),
    var(--secondary-color)
  );
  -webkit-background-clip: text;
  -webkit-text-fill-color: transparent;
}

.hero-subtitle {
  font-size: 1.25rem;
  color: var(--text-secondary);
  margin-bottom: 2rem;
  max-width: 600px;
  margin-left: auto;
  margin-right: auto;
}

.hero-actions {
  display: flex;
  justify-content: center;
  gap: 1rem;
}

.hero-visual {
  position: relative;
}

.logo-container {
  display: flex;
  justify-content: center;
  align-items: center;
  gap: 2rem;
}

.logo {
  height: 8rem;
  will-change: filter;
  transition: filter 300ms;
  filter: drop-shadow(0 0 2em rgba(0, 0, 0, 0.2));
}

.logo:hover {
  filter: drop-shadow(0 0 2em #646cffaa);
}

.logo.vue:hover {
  filter: drop-shadow(0 0 2em #42b883aa);
}

/* Features Section */
.features {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(300px, 1fr));
  gap: 2rem;
  padding: 4rem 1rem;
  margin: 0 auto;
  max-width: 1200px;
}

.feature-card {
  background-color: var(--bg-secondary);
  padding: 2rem;
  border-radius: var(--radius-lg);
  border: 1px solid var(--border-color);
  transition: transform var(--transition-fast);
}

.feature-card:hover {
  transform: translateY(-5px);
  border-color: var(--primary-color);
}

.feature-card h3 {
  color: var(--primary-color);
  margin-bottom: 1rem;
}

@media (min-width: 768px) {
  .hero {
    padding: 6rem 2rem;
  }

  .hero h1 {
    font-size: 3.5rem;
  }
}
</style>
