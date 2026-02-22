<template>
  <form @submit.prevent="handleSubmit" class="sign-in-form">
    <VargoFormField label="Email" htmlFor="email">
      <VargoTextInput
        id="email"
        type="email"
        placeholder="billy.butcher@vaught.com"
        v-model="email"
        required
      />
    </VargoFormField>

    <VargoFormField label="Password" htmlFor="password">
      <VargoTextInput
        id="password"
        type="password"
        placeholder="Enter your password"
        v-model="password"
        required
      />
    </VargoFormField>

    <div class="form-actions">
      <VargoButton type="submit" variant="primary">Sign In</VargoButton>
    </div>
  </form>
</template>

<script setup lang="ts">
import { ref } from 'vue';
import VargoButton from '../atoms/VargoButton.vue';
import VargoFormField from '../molecules/VargoFormField.vue';
import VargoTextInput from '../atoms/VargoTextInput.vue';

const emit = defineEmits<{
  (e: 'submit', payload: { email: string; password: string }): void;
}>();

const email = ref('');
const password = ref('');

const handleSubmit = () => {
  emit('submit', { email: email.value, password: password.value });
};
</script>

<style scoped>
.sign-in-form {
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
