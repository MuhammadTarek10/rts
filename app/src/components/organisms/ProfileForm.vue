<template>
  <form
    class="profile-form"
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
          placeholder="First name"
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
          placeholder="Last name"
          :invalid="!!lastNameError"
        />
      </VargoFormField>
    </div>

    <VargoFormField
      label="Phone Number"
      html-for="phone_number"
      :error="phoneError"
    >
      <VargoTextInput
        id="phone_number"
        v-model="phoneValue"
        type="tel"
        placeholder="Phone number"
        :invalid="!!phoneError"
      />
    </VargoFormField>

    <VargoFormField
      label="Country"
      html-for="country"
      :error="countryError"
    >
      <VargoTextInput
        id="country"
        v-model="countryValue"
        type="text"
        placeholder="Country"
        :invalid="!!countryError"
      />
    </VargoFormField>

    <VargoFormField
      label="Date of Birth"
      html-for="date_of_birth"
      :error="dobError"
    >
      <VargoTextInput
        id="date_of_birth"
        v-model="dobValue"
        type="date"
        :invalid="!!dobError"
      />
    </VargoFormField>

    <VargoFormField
      label="Bio"
      html-for="bio"
      :error="bioError"
    >
      <VargoTextInput
        id="bio"
        v-model="bioValue"
        type="text"
        placeholder="Tell us about yourself"
        :invalid="!!bioError"
      />
    </VargoFormField>

    <div class="form-actions">
      <VargoButton
        type="submit"
        variant="primary"
        :loading="loading"
        :disabled="!meta.dirty"
      >
        Save Changes
      </VargoButton>
    </div>
  </form>
</template>

<script setup lang="ts">
import { useForm, useField } from 'vee-validate';
import { toTypedSchema } from '@vee-validate/zod';
import { profileSchema } from '@/schemas/auth';
import VargoButton from '@/components/atoms/VargoButton.vue';
import VargoFormField from '@/components/molecules/VargoFormField.vue';
import VargoTextInput from '@/components/atoms/VargoTextInput.vue';
import type { UserProfile } from '@/types';

const props = defineProps<{
  profile: UserProfile | null;
  loading?: boolean;
}>();

const emit = defineEmits<{
  (
    e: 'submit',
    payload: {
      first_name?: string;
      last_name?: string;
      phone_number?: string;
      country?: string;
      date_of_birth?: string;
      bio?: string;
    },
  ): void;
}>();

const { handleSubmit, meta } = useForm({
  validationSchema: toTypedSchema(profileSchema),
  initialValues: {
    first_name: props.profile?.first_name ?? '',
    last_name: props.profile?.last_name ?? '',
    phone_number: props.profile?.phone_number ?? '',
    country: props.profile?.country ?? '',
    date_of_birth: props.profile?.date_of_birth ?? '',
    bio: props.profile?.bio ?? '',
  },
});

const { value: firstNameValue, errorMessage: firstNameError } = useField<string>('first_name');
const { value: lastNameValue, errorMessage: lastNameError } = useField<string>('last_name');
const { value: phoneValue, errorMessage: phoneError } = useField<string>('phone_number');
const { value: countryValue, errorMessage: countryError } = useField<string>('country');
const { value: dobValue, errorMessage: dobError } = useField<string>('date_of_birth');
const { value: bioValue, errorMessage: bioError } = useField<string>('bio');

const onSubmit = handleSubmit((values) => {
  const payload: Record<string, string | undefined> = {};
  for (const [key, val] of Object.entries(values)) {
    if (val && val.trim()) {
      payload[key] = val.trim();
    }
  }
  emit('submit', payload);
});
</script>

<style scoped>
.profile-form {
  width: 100%;
}

.name-row {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 0.75rem;
}

.form-actions {
  display: flex;
  justify-content: flex-end;
  margin-top: 0.75rem;
}
</style>
