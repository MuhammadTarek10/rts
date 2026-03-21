<template>
  <div class="profile-page">
    <div class="profile-header">
      <p class="profile-eyebrow">
        Settings
      </p>
      <h1>Profile</h1>
    </div>

    <section class="profile-section">
      <div class="section-title">
        <h2>Personal Information</h2>
        <p class="section-description">
          Update your personal details
        </p>
      </div>
      <ProfileForm
        :profile="userStore.user?.profile ?? null"
        :loading="isUpdatingProfile"
        @submit="handleUpdateProfile"
      />
    </section>

    <section class="profile-section">
      <div class="section-title">
        <h2>Change Password</h2>
        <p class="section-description">
          Update your account password
        </p>
      </div>
      <ChangePasswordForm
        :loading="isChangingPassword"
        @submit="handleChangePassword"
      />
    </section>

    <section class="profile-section danger-zone">
      <div class="section-title">
        <h2>Danger Zone</h2>
        <p class="section-description">
          Permanently delete your account and all associated data. This action
          cannot be undone.
        </p>
      </div>
      <VargoButton
        variant="danger"
        :loading="isDeleting"
        @click="handleDeleteAccount"
      >
        Delete Account
      </VargoButton>
    </section>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue';
import { useRouter } from 'vue-router';
import { useUserStore } from '@/stores/user';
import { authService } from '@/services/auth.service';
import { getErrorMessage } from '@/utils/error';
import ProfileForm from '@/components/organisms/ProfileForm.vue';
import ChangePasswordForm from '@/components/organisms/ChangePasswordForm.vue';
import VargoButton from '@/components/atoms/VargoButton.vue';
import type { UpdateProfilePayload, ChangePasswordPayload } from '@/types';
import { push } from 'notivue';

const router = useRouter();
const userStore = useUserStore();

const isUpdatingProfile = ref(false);
const isChangingPassword = ref(false);
const isDeleting = ref(false);

const handleUpdateProfile = async (payload: UpdateProfilePayload) => {
  isUpdatingProfile.value = true;
  try {
    const response = await authService.updateProfile(payload);
    userStore.setUser(response.data.data);
    push.success('Profile updated successfully');
  } catch (err) {
    push.error(getErrorMessage(err, 'Failed to update profile'));
  } finally {
    isUpdatingProfile.value = false;
  }
};

const handleChangePassword = async (payload: ChangePasswordPayload) => {
  isChangingPassword.value = true;
  try {
    await authService.changePassword(payload);
    push.success('Password changed. Please sign in again.');
    await userStore.signOut();
    router.push('/auth/sign-in');
  } catch (err) {
    push.error(getErrorMessage(err, 'Failed to change password'));
  } finally {
    isChangingPassword.value = false;
  }
};

const handleDeleteAccount = async () => {
  const confirmed = globalThis.confirm(
    'Are you sure you want to delete your account? This action cannot be undone.'
  );
  if (!confirmed) return;

  isDeleting.value = true;
  try {
    await authService.deleteAccount();
    userStore.clearUser();
    push.success('Your account has been deleted');
    router.push('/auth/sign-in');
  } catch (err) {
    push.error(getErrorMessage(err, 'Failed to delete account'));
  } finally {
    isDeleting.value = false;
  }
};
</script>

<style scoped lang="less">
.profile-page {
  max-width: 640px;
  margin: 0 auto;
  padding: 2rem 0;
}

.profile-header {
  margin-bottom: 2rem;
  animation: fadeUp 0.5s ease both;
}

.profile-eyebrow {
  font-size: 0.8rem;
  text-transform: uppercase;
  letter-spacing: 0.15em;
  color: var(--primary-color);
  margin-bottom: 0.5rem;
  font-weight: 600;
}

.profile-header h1 {
  font-size: 2.25rem;
}

.profile-section {
  background-color: var(--bg-secondary);
  border: 1px solid var(--border-color);
  border-radius: var(--radius-lg);
  padding: 1.5rem;
  margin-bottom: 1.25rem;
  animation: fadeUp 0.5s ease both;

  &:nth-child(2) {
    animation-delay: 0.05s;
  }
  &:nth-child(3) {
    animation-delay: 0.1s;
  }
  &:nth-child(4) {
    animation-delay: 0.15s;
  }
}

.section-title {
  margin-bottom: 1.25rem;

  h2 {
    font-size: 1.15rem;
    margin-bottom: 0.25rem;
  }

  .section-description {
    font-size: 0.88rem;
    color: var(--text-tertiary);
    margin-bottom: 0;
  }
}

.danger-zone {
  border-color: rgba(217, 72, 72, 0.15);
}

@keyframes fadeUp {
  from {
    opacity: 0;
    transform: translateY(16px);
  }
  to {
    opacity: 1;
    transform: translateY(0);
  }
}
</style>
