<template>
  <div class="default-layout">
    <vargo-navbar />

    <main class="main-content">
      <div class="container">
        <!-- The page content will be injected here -->
        <slot />
      </div>
    </main>

    <footer class="main-footer">
      <section class="container">
        <p>© 2026 Vargo App. All rights reserved.</p>
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
  padding: 2rem 0;
}

.main-footer {
  background-color: var(--bg-secondary);
  border-top: 1px solid var(--border-color);
  padding: 2rem 0;
  margin-top: auto;

  p {
    margin: 0;
    color: var(--text-tertiary);
  }

  .container {
    display: flex;
    justify-content: space-between;
    align-items: center;
    flex-wrap: wrap;
    gap: 1rem;
  }

  .footer-links {
    display: flex;
    gap: 1.5rem;
    list-style: none;

    a {
      color: var(--text-secondary);
      font-size: 0.9rem;

      &:hover {
        color: var(--primary-color);
      }
    }
  }
}
</style>
