<template>
  <div class="auth-page">
    <div class="page-title">
      <h2>Forgot Password</h2>
      <p class="instruction">
        Enter your email address and we'll send you a link to reset your
        password.
      </p>
    </div>

    <form @submit.prevent="handleSubmit">
      <vargo-form-field label="Email" htmlFor="email">
        <vargo-text-input
          id="email"
          type="email"
          v-model="email"
          required
          placeholder="Enter your email"
        />
      </vargo-form-field>

      <div class="form-actions">
        <vargo-button type="submit" variant="primary"
          >Send Reset Link</vargo-button
        >
      </div>
    </form>

    <div class="auth-links text-center mt-md">
      <router-link to="/auth/sign-in">Back to Sign In</router-link>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue';
import VargoButton from '../../components/atoms/VargoButton.vue';
import VargoFormField from '../../components/molecules/VargoFormField.vue';
import VargoTextInput from '../../components/atoms/VargoTextInput.vue';

const email = ref('');

const handleSubmit = () => {
  console.log('Reset password for:', email.value);
  // Add password reset logic here
};
</script>

<style scoped>
.page-title {
  text-align: center;
  margin-bottom: 2rem;
}

.page-title h2 {
  font-size: 1.5rem;
  margin-bottom: 0.5rem;
}

.instruction {
  color: var(--text-secondary);
  font-size: 0.95rem;
  line-height: 1.5;
}

.form-actions {
  display: flex;
  margin-top: 1.5rem;
}

.form-actions :deep(.atom-button) {
  width: 100%;
}

.auth-links {
  font-size: 0.9rem;
}

.auth-links a {
  color: var(--primary-color);
  text-decoration: none;
}
.auth-links a:hover {
  text-decoration: underline;
}
</style>
