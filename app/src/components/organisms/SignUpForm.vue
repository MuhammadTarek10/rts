<template>
  <form @submit.prevent="handleSubmit" class="sign-up-form">
    <vargo-form-field label="Full Name" htmlFor="name">
      <vargo-text-input
        id="name"
        type="text"
        placeholder="Billy Butcher"
        v-model="name"
        required
      />
    </vargo-form-field>

    <vargo-form-field label="Email" htmlFor="email">
      <vargo-text-input
        id="email"
        type="email"
        placeholder="billy.butcher@vaught.com"
        v-model="email"
        required
      />
    </vargo-form-field>

    <vargo-form-field label="Password" htmlFor="password">
      <vargo-text-input
        id="password"
        type="password"
        placeholder="Create a password"
        v-model="password"
        required
      />
    </vargo-form-field>

    <vargo-form-field label="Confirm Password" htmlFor="confirm-password">
      <vargo-text-input
        id="confirm-password"
        type="password"
        placeholder="Confirm your password"
        v-model="confirmPassword"
        required
      />
    </vargo-form-field>

    <div class="form-actions">
      <VargoButton type="submit" variant="primary">Sign Up</VargoButton>
    </div>
  </form>
</template>

<script setup lang="ts">
import { ref } from 'vue';
import VargoButton from '../atoms/VargoButton.vue';
import VargoFormField from '../molecules/VargoFormField.vue';
import VargoTextInput from '../atoms/VargoTextInput.vue';

const emit =
  defineEmits<
    (
      e: 'submit',
      payload: { name: string; email: string; password: string }
    ) => void
  >();

const name = ref('');
const email = ref('');
const password = ref('');
const confirmPassword = ref('');

const handleSubmit = () => {
  if (password.value !== confirmPassword.value) {
    alert("Passwords don't match"); // Or use a proper error handling
    return;
  }
  emit('submit', {
    name: name.value,
    email: email.value,
    password: password.value,
  });
};
</script>

<style scoped>
.sign-up-form {
  width: 100%;
}

.form-actions {
  display: flex;
  justify-content: center;
  margin-top: 1rem;
  width: 100%;
}

.form-actions :deep(.atom-button) {
  width: 100%;
}
</style>
