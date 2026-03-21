<template>
  <div class="default-layout">
    <vargo-navbar />

    <main class="main-content">
      <div class="container">
        <slot />
      </div>
    </main>

    <footer class="main-footer">
      <section class="container">
        <span class="footer-brand">Vargo</span>
        <p class="footer-copy">
          &copy; 2026
        </p>
        <ul class="footer-links">
          <li><a href="#">Privacy</a></li>
          <li><a href="#">Terms</a></li>
          <li><a href="#">Contact</a></li>
        </ul>
      </section>
    </footer>
  </div>
</template>

<script setup lang="ts">
import VargoNavbar from '../components/organisms/VargoNavbar.vue';
</script>

<style scoped lang="less">
.default-layout {
  display: flex;
  flex-direction: column;
  min-height: 100vh;
  background-color: var(--bg-color);
}

.main-content {
  flex: 1;
  padding: 2.5rem 0;
}

.main-footer {
  border-top: 1px solid var(--border-color);
  padding: 1.5rem 0;
  margin-top: auto;

  .container {
    display: flex;
    justify-content: space-between;
    align-items: center;
    flex-wrap: wrap;
    gap: 1rem;
  }

  .footer-brand {
    font-family: var(--font-display);
    font-size: 1.1rem;
    color: var(--text-tertiary);
    letter-spacing: -0.02em;
  }

  .footer-copy {
    margin: 0;
    color: var(--text-tertiary);
    font-size: 0.85rem;
  }

  .footer-links {
    display: flex;
    gap: 1.5rem;
    list-style: none;
    padding: 0;
    margin: 0;

    a {
      color: var(--text-tertiary);
      font-size: 0.85rem;
      font-weight: 400;
      text-decoration: none;
      transition: color var(--transition-fast);

      &:hover {
        color: var(--primary-color);
      }
    }
  }
}
</style>
