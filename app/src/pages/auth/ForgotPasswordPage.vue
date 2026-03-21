<template>
  <div class="auth-page">
    <div class="page-title">
      <h2>Forgot Password</h2>
      <p class="instruction">
        Password reset is not yet available. Please contact support if you need
        to reset your password.
      </p>
    </div>

    <div class="auth-links">
      <router-link to="/auth/sign-in">
        Back to Sign In
      </router-link>
    </div>
  </div>
</template>

<style scoped>
.page-title {
  text-align: center;
  margin-bottom: 1.5rem;
}

.page-title h2 {
  font-size: 1.5rem;
  margin-bottom: 0.5rem;
}

.instruction {
  color: var(--text-tertiary);
  font-size: 0.92rem;
  line-height: 1.5;
}

.auth-links {
  text-align: center;
  margin-top: 0.5rem;
}

.auth-links a {
  color: var(--primary-color);
  font-size: 0.9rem;
  font-weight: 500;
  text-decoration: none;
}

.auth-links a:hover {
  color: var(--primary-hover);
}
</style>
