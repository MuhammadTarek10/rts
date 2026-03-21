import { ref } from 'vue';

export interface Toast {
  id: number;
  message: string;
  type: 'success' | 'error' | 'info' | 'warning';
}

const toasts = ref<Toast[]>([]);
let nextId = 0;

export function useToast() {
  function addToast(
    message: string,
    type: Toast['type'] = 'info',
    duration = 4000,
  ) {
    const id = nextId++;
    toasts.value.push({ id, message, type });
    setTimeout(() => removeToast(id), duration);
  }

  function removeToast(id: number) {
    toasts.value = toasts.value.filter((t) => t.id !== id);
  }

  return {
    toasts,
    removeToast,
    success: (msg: string) => addToast(msg, 'success'),
    error: (msg: string) => addToast(msg, 'error'),
    info: (msg: string) => addToast(msg, 'info'),
    warning: (msg: string) => addToast(msg, 'warning'),
  };
}
