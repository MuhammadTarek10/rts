<template>
  <div class="about-page">
    <div class="about-header">
      <p class="about-eyebrow">
        About
      </p>
      <h1>Vargo</h1>
      <p class="about-subtitle">
        A modern microservices platform for retail operations.
      </p>
    </div>

    <div class="about-content">
      <p>
        Vargo is built on a polyglot microservices architecture — combining
        NestJS, ASP.NET Core, and Go to deliver a scalable retail operations
        platform. Each service is purpose-built with the right tool for the job.
      </p>
      <p>
        The system manages authentication, product catalogs, inventory tracking,
        and more through a unified API gateway, with event-driven communication
        via RabbitMQ.
      </p>
    </div>

    <div class="about-divider" />

    <div class="tech-grid">
      <div
        v-for="(tech, i) in technologies"
        :key="tech.name"
        class="tech-card"
        :style="{ animationDelay: `${0.2 + i * 0.06}s` }"
      >
        <span class="tech-label">{{ tech.label }}</span>
        <span class="tech-name">{{ tech.name }}</span>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
const technologies = [
  { label: 'Frontend', name: 'Vue 3' },
  { label: 'Auth', name: 'NestJS' },
  { label: 'Catalog', name: '.NET 10' },
  { label: 'Inventory', name: 'Go' },
  { label: 'Database', name: 'PostgreSQL' },
  { label: 'Documents', name: 'MongoDB' },
  { label: 'Cache', name: 'Redis' },
  { label: 'Messaging', name: 'RabbitMQ' },
];
</script>

<style scoped>
.about-page {
  max-width: 720px;
  margin: 0 auto;
  padding: 2rem 0;
}

.about-header {
  margin-bottom: 2.5rem;
  animation: fadeUp 0.5s ease both;
}

.about-eyebrow {
  font-size: 0.8rem;
  text-transform: uppercase;
  letter-spacing: 0.15em;
  color: var(--primary-color);
  margin-bottom: 0.5rem;
  font-weight: 600;
}

.about-header h1 {
  font-size: 3rem;
  margin-bottom: 0.75rem;
}

.about-subtitle {
  font-size: 1.2rem;
  color: var(--text-secondary);
  line-height: 1.5;
}

.about-content {
  animation: fadeUp 0.5s ease 0.1s both;

  p {
    font-size: 1.05rem;
    line-height: 1.8;
    margin-bottom: 1.25rem;
  }
}

.about-divider {
  width: 40px;
  height: 1px;
  background: var(--primary-color);
  opacity: 0.4;
  margin: 2.5rem 0;
}

.tech-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(150px, 1fr));
  gap: 0.75rem;
}

.tech-card {
  background-color: var(--bg-secondary);
  border: 1px solid var(--border-color);
  border-radius: var(--radius-md);
  padding: 1rem;
  display: flex;
  flex-direction: column;
  gap: 0.25rem;
  transition:
    border-color var(--transition-normal),
    background-color var(--transition-normal);
  animation: fadeUp 0.5s ease both;

  &:hover {
    border-color: rgba(212, 160, 83, 0.2);
  }
}

.tech-label {
  font-size: 0.72rem;
  text-transform: uppercase;
  letter-spacing: 0.1em;
  color: var(--text-tertiary);
  font-weight: 600;
}

.tech-name {
  font-family: var(--font-display);
  font-size: 1.2rem;
  color: var(--text-primary);
}

@keyframes fadeUp {
  from {
    opacity: 0;
    transform: translateY(16px);
  }
  to {
    opacity: 1;
    transform: translateY(0);
  }
}
</style>
