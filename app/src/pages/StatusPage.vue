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

    <section class="services">
      <div class="section-header">
        <h2>Services</h2>
        <div class="section-line" />
      </div>
      <div class="service-grid">
        <div
          v-for="(service, i) in services"
          :key="service.name"
          class="service-card"
          :style="{ animationDelay: `${i * 0.08}s` }"
        >
          <div class="service-card-header">
            <h3>{{ service.name }}</h3>
            <span
              class="service-status"
              :class="service.status"
            >{{
              service.statusLabel
            }}</span>
          </div>
          <p>{{ service.description }}</p>
        </div>
      </div>
    </section>

    <section class="quick-stats">
      <div class="section-header">
        <h2>Overview</h2>
        <div class="section-line" />
      </div>
      <div class="stats-grid">
        <div
          v-for="(stat, i) in stats"
          :key="stat.label"
          class="stat-card"
          :style="{ animationDelay: `${0.3 + i * 0.06}s` }"
        >
          <span class="stat-value">{{ stat.value }}</span>
          <span class="stat-label">{{ stat.label }}</span>
        </div>
      </div>
    </section>
  </div>
</template>

<script setup lang="ts">
import { useUserStore } from '@/stores/user';

const userStore = useUserStore();

const services = [
  {
    name: 'Authentication',
    description: 'User management, sessions, and access control.',
    status: 'active',
    statusLabel: 'Active',
  },
  {
    name: 'Catalog',
    description: 'Product listings, categories, and media assets.',
    status: 'active',
    statusLabel: 'Active',
  },
  {
    name: 'Inventory',
    description: 'Stock levels, warehouses, and reservations.',
    status: 'active',
    statusLabel: 'Active',
  },
  {
    name: 'Orders',
    description: 'Order processing, fulfillment, and tracking.',
    status: 'planned',
    statusLabel: 'Planned',
  },
  {
    name: 'Payments',
    description: 'Payment processing and transaction management.',
    status: 'planned',
    statusLabel: 'Planned',
  },
  {
    name: 'Notifications',
    description: 'Email, SMS, and push notification delivery.',
    status: 'planned',
    statusLabel: 'Planned',
  },
];

const stats = [
  { value: '3', label: 'Active Services' },
  { value: '6', label: 'Total Planned' },
  { value: '99.9%', label: 'Uptime' },
  { value: '< 50ms', label: 'Avg Latency' },
];
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

/* Services */
.services {
  margin-bottom: 3rem;
}

.service-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(320px, 1fr));
  gap: 1rem;
}

.service-card {
  background-color: var(--bg-secondary);
  border: 1px solid var(--border-color);
  border-radius: var(--radius-lg);
  padding: 1.25rem 1.5rem;
  transition:
    border-color var(--transition-normal),
    background-color var(--transition-normal);
  animation: fadeUp 0.5s ease both;

  &:hover {
    border-color: rgba(212, 160, 83, 0.2);
    background-color: var(--bg-tertiary);
  }

  p {
    font-size: 0.9rem;
    color: var(--text-tertiary);
    margin: 0;
    line-height: 1.5;
  }
}

.service-card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 0.5rem;

  h3 {
    font-size: 1.1rem;
    margin-bottom: 0;
  }
}

.service-status {
  font-size: 0.72rem;
  font-weight: 600;
  text-transform: uppercase;
  letter-spacing: 0.08em;
  padding: 0.2rem 0.5rem;
  border-radius: var(--radius-sm);
  font-family: var(--font-sans);
}

.service-status.active {
  color: var(--success-color);
  background: rgba(62, 186, 110, 0.1);
}

.service-status.planned {
  color: var(--text-tertiary);
  background: rgba(92, 92, 96, 0.15);
}

/* Stats */
.stats-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(200px, 1fr));
  gap: 1rem;
}

.stat-card {
  background-color: var(--bg-secondary);
  border: 1px solid var(--border-color);
  border-radius: var(--radius-lg);
  padding: 1.5rem;
  display: flex;
  flex-direction: column;
  gap: 0.25rem;
  animation: fadeUp 0.5s ease both;
}

.stat-value {
  font-family: var(--font-display);
  font-size: 2rem;
  color: var(--primary-color);
  line-height: 1;
}

.stat-label {
  font-size: 0.82rem;
  color: var(--text-tertiary);
  text-transform: uppercase;
  letter-spacing: 0.08em;
  font-weight: 500;
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

@media (max-width: 768px) {
  .hero h1 {
    font-size: 2rem;
  }

  .service-grid {
    grid-template-columns: 1fr;
  }

  .stats-grid {
    grid-template-columns: repeat(2, 1fr);
  }
}
</style>
