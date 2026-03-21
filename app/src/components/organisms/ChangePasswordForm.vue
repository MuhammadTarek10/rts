<template>
  <form
    class="change-password-form"
    @submit="onSubmit"
  >
    <VargoFormField
      label="Current Password"
      html-for="current_password"
      :error="currentError"
    >
      <VargoTextInput
        id="current_password"
        v-model="currentValue"
        type="password"
        placeholder="Enter current password"
        :invalid="!!currentError"
      />
    </VargoFormField>

    <VargoFormField
      label="New Password"
      html-for="new_password"
      :error="newError"
    >
      <VargoTextInput
        id="new_password"
        v-model="newValue"
        type="password"
        placeholder="Enter new password"
        :invalid="!!newError"
      />
    </VargoFormField>

    <VargoFormField
      label="Confirm New Password"
      html-for="confirm_password"
      :error="confirmError"
    >
      <VargoTextInput
        id="confirm_password"
        v-model="confirmValue"
        type="password"
        placeholder="Confirm new password"
        :invalid="!!confirmError"
      />
    </VargoFormField>

    <div class="form-actions">
      <VargoButton
        type="submit"
        variant="primary"
        :loading="loading"
      >
        Change Password
      </VargoButton>
    </div>
  </form>
</template>

<script setup lang="ts">
import { useForm, useField } from 'vee-validate';
import { toTypedSchema } from '@vee-validate/zod';
import { changePasswordSchema } from '@/schemas/auth';
import VargoButton from '@/components/atoms/VargoButton.vue';
import VargoFormField from '@/components/molecules/VargoFormField.vue';
import VargoTextInput from '@/components/atoms/VargoTextInput.vue';

defineProps<{
  loading?: boolean;
}>();

const emit = defineEmits<{
  (
    e: 'submit',
    payload: { current_password: string; new_password: string },
  ): void;
}>();

const { handleSubmit } = useForm({
  validationSchema: toTypedSchema(changePasswordSchema),
});

const { value: currentValue, errorMessage: currentError } = useField<string>('current_password');
const { value: newValue, errorMessage: newError } = useField<string>('new_password');
const { value: confirmValue, errorMessage: confirmError } = useField<string>('confirm_password');

const onSubmit = handleSubmit((values) => {
  emit('submit', {
    current_password: values.current_password,
    new_password: values.new_password,
  });
});
</script>

<style scoped>
.change-password-form {
  width: 100%;
}

.form-actions {
  display: flex;
  justify-content: flex-end;
  margin-top: 0.75rem;
}
</style>
