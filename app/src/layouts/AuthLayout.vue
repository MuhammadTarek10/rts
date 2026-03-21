<template>
  <div class="auth-layout">
    <div class="auth-bg-accent" />
    <div class="auth-container">
      <div class="auth-card">
        <header class="auth-header">
          <h1>Vargo</h1>
          <div class="header-line" />
        </header>
        <main class="auth-content">
          <slot />
        </main>
        <footer class="auth-footer">
          <slot name="footer" />
          <div
            v-if="!$slots.footer"
            class="default-footer-links"
          >
            <router-link
              v-if="$route.path.includes('sign-up')"
              to="/auth/sign-in"
            >
              Already have an account? <strong>Sign In</strong>
            </router-link>
            <router-link
              v-else-if="$route.path.includes('sign-in')"
              to="/auth/sign-up"
            >
              Don't have an account? <strong>Sign Up</strong>
            </router-link>
          </div>
        </footer>
      </div>
    </div>
  </div>
</template>

<style scoped lang="less">
.auth-layout {
  min-height: 100vh;
  display: flex;
  align-items: center;
  justify-content: center;
  background: var(--bg-color);
  position: relative;
  overflow: hidden;
  padding: 1rem;
}

.auth-bg-accent {
  position: absolute;
  top: -40%;
  right: -20%;
  width: 600px;
  height: 600px;
  border-radius: 50%;
  background: radial-gradient(
    circle,
    rgba(212, 160, 83, 0.06) 0%,
    transparent 70%
  );
  pointer-events: none;
}

.auth-container {
  width: 100%;
  max-width: 440px;
  position: relative;
  z-index: 1;
}

.auth-card {
  background-color: var(--bg-secondary);
  border-radius: var(--radius-xl);
  border: 1px solid var(--border-color);
  padding: 2.5rem;
  animation: cardReveal 0.6s cubic-bezier(0.16, 1, 0.3, 1);
}

.auth-header {
  text-align: center;
  margin-bottom: 2rem;

  h1 {
    font-family: var(--font-display);
    font-size: 2.5rem;
    font-weight: 400;
    color: var(--primary-color);
    margin-bottom: 1rem;
    letter-spacing: -0.02em;
  }
}

.header-line {
  width: 40px;
  height: 1px;
  background: var(--primary-color);
  margin: 0 auto;
  opacity: 0.5;
}

.auth-content {
  margin-bottom: 1.5rem;
}

.auth-footer {
  text-align: center;
  font-size: 0.9rem;
  color: var(--text-tertiary);
  margin-top: 1rem;
  border-top: 1px solid var(--border-color);
  padding-top: 1.5rem;

  a {
    color: var(--text-secondary);
    text-decoration: none;
    transition: color 0.15s ease;

    strong {
      color: var(--primary-color);
    }

    &:hover {
      color: var(--primary-color);
    }
  }
}

@keyframes cardReveal {
  from {
    opacity: 0;
    transform: translateY(24px) scale(0.98);
  }
  to {
    opacity: 1;
    transform: translateY(0) scale(1);
  }
}
</style>
