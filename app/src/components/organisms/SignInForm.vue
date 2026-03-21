<template>
  <form
    class="sign-in-form"
    @submit="onSubmit"
  >
    <VargoFormField
      label="Email"
      html-for="email"
      :error="emailError"
    >
      <VargoTextInput
        id="email"
        v-model="emailValue"
        type="email"
        placeholder="billy.butcher@vaught.com"
        :invalid="!!emailError"
      />
    </VargoFormField>

    <VargoFormField
      label="Password"
      html-for="password"
      :error="passwordError"
    >
      <VargoTextInput
        id="password"
        v-model="passwordValue"
        type="password"
        placeholder="Enter your password"
        :invalid="!!passwordError"
      />
    </VargoFormField>

    <div class="form-options">
      <router-link
        to="/auth/forgot-password"
        class="forgot-link"
      >
        Forgot Password?
      </router-link>
    </div>

    <div
      v-if="error"
      class="form-error"
    >
      {{ error }}
    </div>

    <div class="form-actions">
      <VargoButton
        type="submit"
        variant="primary"
        :loading="loading"
      >
        Sign In
      </VargoButton>
    </div>
  </form>
</template>

<script setup lang="ts">
import { useForm, useField } from 'vee-validate';
import { toTypedSchema } from '@vee-validate/zod';
import { signInSchema } from '@/schemas/auth';
import VargoButton from '@/components/atoms/VargoButton.vue';
import VargoFormField from '@/components/molecules/VargoFormField.vue';
import VargoTextInput from '@/components/atoms/VargoTextInput.vue';

defineProps<{
  loading?: boolean;
  error?: string | null;
}>();

const emit = defineEmits<{
  (e: 'submit', payload: { email: string; password: string }): void;
}>();

const { handleSubmit } = useForm({
  validationSchema: toTypedSchema(signInSchema),
});

const { value: emailValue, errorMessage: emailError } = useField<string>('email');
const { value: passwordValue, errorMessage: passwordError } = useField<string>('password');

const onSubmit = handleSubmit((values) => {
  emit('submit', { email: values.email, password: values.password });
});
</script>

<style scoped>
.sign-in-form {
  width: 100%;
}

.form-options {
  display: flex;
  justify-content: flex-end;
  margin-top: -0.5rem;
  margin-bottom: 0.5rem;
}

.forgot-link {
  font-size: 0.82rem;
  color: var(--text-tertiary);
  text-decoration: none;
  transition: color var(--transition-fast);
}

.forgot-link:hover {
  color: var(--primary-color);
}

.form-error {
  padding: 0.7rem 0.875rem;
  border-radius: var(--radius-md);
  background-color: rgba(217, 72, 72, 0.08);
  border: 1px solid rgba(217, 72, 72, 0.2);
  color: var(--danger-color);
  font-size: 0.88rem;
  margin-bottom: 1rem;
}

.form-actions {
  display: flex;
  justify-content: center;
  margin-top: 0.75rem;
  width: 100%;
}

.form-actions :deep(.atom-button) {
  width: 100%;
}
</style>
