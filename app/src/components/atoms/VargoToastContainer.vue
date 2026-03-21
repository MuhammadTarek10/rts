<template>
  <Notivue v-slot="item">
    <Notification
      :item="item"
      :theme="theme"
      :hide-close="true"
    />
  </Notivue>
</template>

<script setup lang="ts">
import { Notivue, Notification } from 'notivue';
import type { NotivueTheme } from 'notivue';

const theme: NotivueTheme = {
  // Layout
  '--nv-width': '340px',
  '--nv-spacing': '0.75rem',
  '--nv-radius': '10px',
  '--nv-border-width': '1px',
  '--nv-icon-size': '1.1rem',
  '--nv-tip-width': '3px',
  '--nv-shadow': '0 4px 20px rgba(0, 0, 0, 0.35)',
  '--nv-message-size': '0.875rem',

  // Global
  '--nv-global-bg': '#1a1a1f',
  '--nv-global-fg': '#e8e8ec',
  '--nv-global-border': 'rgba(255, 255, 255, 0.06)',
  '--nv-global-accent': '#d4a053',

  // Per-type accents
  '--nv-success-accent': '#3eba6e',
  '--nv-error-accent': '#d94848',
  '--nv-warning-accent': '#e5a83b',
  '--nv-info-accent': '#5a8fd4',
  '--nv-promise-accent': '#d4a053',
};
</script>
