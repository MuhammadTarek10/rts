<template>
  <div class="about-page">
    <div class="card">
      <header class="page-header">
        <h1>About Vargo</h1>
        <p class="subtitle">Empowering developers to build better.</p>
      </header>

      <div class="content">
        <p>
          Vargo is a demonstration application showcasing a modern Vue 3
          architecture with TypeScript, Pinia, and a custom design system
          without relying on heavy utility-first frameworks.
        </p>
        <p>
          Our mission is to provide clean, maintainable, and scalable
          foundations for enterprise-grade applications.
        </p>
      </div>

      <div class="stats">
        <div class="stat-item">
          <span class="value">100%</span>
          <span class="label">Custom Styles</span>
        </div>
        <div class="stat-item">
          <span class="value">Vue 3</span>
          <span class="label">Framework</span>
        </div>
        <div class="stat-item">
          <span class="value">TS</span>
          <span class="label">Type Safety</span>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.about-page {
  max-width: 800px;
  margin: 0 auto;
}

.page-header {
  margin-bottom: 2rem;
  border-bottom: 1px solid var(--border-color);
  padding-bottom: 1rem;
}

.subtitle {
  font-size: 1.25rem;
  color: var(--text-secondary);
}

.content p {
  font-size: 1.1rem;
  margin-bottom: 1.5rem;
  line-height: 1.8;
}

.stats {
  display: flex;
  justify-content: space-around;
  margin-top: 3rem;
  padding-top: 2rem;
  border-top: 1px solid var(--border-color);
}

.stat-item {
  display: flex;
  flex-direction: column;
  align-items: center;
}

.stat-item .value {
  font-size: 2.5rem;
  font-weight: 800;
  color: var(--accent-color);
  line-height: 1;
}

.stat-item .label {
  margin-top: 0.5rem;
  font-size: 0.9rem;
  text-transform: uppercase;
  letter-spacing: 1px;
  color: var(--text-tertiary);
}
</style>
