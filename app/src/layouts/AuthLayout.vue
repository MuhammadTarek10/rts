<template>
  <div class="auth-layout">
    <div class="auth-container">
      <div class="auth-card">
        <header class="auth-header">
          <h1>Vargo</h1>
          <p class="subtitle">Welcome back to the future</p>
        </header>
        <main class="auth-content">
          <slot />
        </main>
        <footer class="auth-footer">
          <slot name="footer" />
          <div class="default-footer-links" v-if="!$slots.footer">
            <router-link
              to="/auth/sign-in"
              v-if="$route.path.includes('sign-up')"
            >
              Already have an account? <strong>Sign In</strong>
            </router-link>
            <router-link
              to="/auth/sign-up"
              v-else-if="$route.path.includes('sign-in')"
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
  background-image:
    radial-gradient(at 0% 0%, rgba(99, 102, 241, 0.15) 0px, transparent 50%),
    radial-gradient(at 100% 100%, rgba(236, 72, 153, 0.15) 0px, transparent 50%);
  padding: 1rem;
}

.auth-container {
  width: 100%;
  max-width: 440px;
}

.auth-card {
  background-color: var(--bg-secondary);
  border-radius: var(--radius-lg);
  border: 1px solid var(--border-color);
  box-shadow: 0 25px 50px -12px rgba(0, 0, 0, 0.25);
  padding: 2.5rem;
  animation: slideUp 0.5s ease-out;
}

.auth-header {
  text-align: center;
  margin-bottom: 2rem;

  h1 {
    font-size: 2.5rem;
    color: var(--primary-color);
    margin-bottom: 0.5rem;
    letter-spacing: -1px;
    background: linear-gradient(
      to right,
      var(--primary-color),
      var(--secondary-color)
    );
    -webkit-background-clip: text;
    -webkit-text-fill-color: transparent;
  }

  .subtitle {
    font-size: 1rem;
    color: var(--text-secondary);
  }
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
    color: var(--primary-color);
    text-decoration: none;
    transition: color 0.15s ease;

    &:hover {
      text-decoration: underline;
      color: var(--primary-hover);
    }
  }
}

@keyframes slideUp {
  from {
    opacity: 0;
    transform: translateY(20px);
  }
  to {
    opacity: 1;
    transform: translateY(0);
  }
}
</style>
