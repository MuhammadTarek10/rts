<template>
  <form
    class="sign-up-form"
    @submit="onSubmit"
  >
    <div class="name-row">
      <VargoFormField
        label="First Name"
        html-for="first_name"
        :error="firstNameError"
      >
        <VargoTextInput
          id="first_name"
          v-model="firstNameValue"
          type="text"
          placeholder="Billy"
          :invalid="!!firstNameError"
        />
      </VargoFormField>

      <VargoFormField
        label="Last Name"
        html-for="last_name"
        :error="lastNameError"
      >
        <VargoTextInput
          id="last_name"
          v-model="lastNameValue"
          type="text"
          placeholder="Butcher"
          :invalid="!!lastNameError"
        />
      </VargoFormField>
    </div>

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
        placeholder="Create a password"
        :invalid="!!passwordError"
      />
    </VargoFormField>

    <VargoFormField
      label="Confirm Password"
      html-for="confirm_password"
      :error="confirmPasswordError"
    >
      <VargoTextInput
        id="confirm_password"
        v-model="confirmPasswordValue"
        type="password"
        placeholder="Confirm your password"
        :invalid="!!confirmPasswordError"
      />
    </VargoFormField>

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
        Sign Up
      </VargoButton>
    </div>
  </form>
</template>

<script setup lang="ts">
import { useForm, useField } from 'vee-validate';
import { toTypedSchema } from '@vee-validate/zod';
import { signUpSchema } from '@/schemas/auth';
import VargoButton from '@/components/atoms/VargoButton.vue';
import VargoFormField from '@/components/molecules/VargoFormField.vue';
import VargoTextInput from '@/components/atoms/VargoTextInput.vue';

defineProps<{
  loading?: boolean;
  error?: string | null;
}>();

const emit = defineEmits<{
  (
    e: 'submit',
    payload: {
      email: string;
      password: string;
      first_name: string;
      last_name: string;
    },
  ): void;
}>();

const { handleSubmit } = useForm({
  validationSchema: toTypedSchema(signUpSchema),
});

const { value: firstNameValue, errorMessage: firstNameError } = useField<string>('first_name');
const { value: lastNameValue, errorMessage: lastNameError } = useField<string>('last_name');
const { value: emailValue, errorMessage: emailError } = useField<string>('email');
const { value: passwordValue, errorMessage: passwordError } = useField<string>('password');
const { value: confirmPasswordValue, errorMessage: confirmPasswordError } = useField<string>('confirm_password');

const onSubmit = handleSubmit((values) => {
  emit('submit', {
    email: values.email,
    password: values.password,
    first_name: values.first_name,
    last_name: values.last_name,
  });
});
</script>

<style scoped>
.sign-up-form {
  width: 100%;
}

.name-row {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 0.75rem;
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
