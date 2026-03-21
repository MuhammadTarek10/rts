import axios from 'axios';
import type { ApiResponse } from '@/types';

export function getErrorMessage(
  err: unknown,
  fallback = 'An unexpected error occurred',
): string {
  if (axios.isAxiosError(err)) {
    if (!err.response) {
      return 'Unable to connect. Please check your internet connection.';
    }

    const data = err.response.data as ApiResponse<null> | undefined;
    if (data?.error) return data.error;
    if (data?.message && data.message !== 'Bad Request') return data.message;

    switch (err.response.status) {
      case 401:
        return 'Invalid email or password';
      case 409:
        return 'An account with this email already exists';
      case 400:
        return 'Please check your input and try again';
      default:
        return fallback;
    }
  }

  return fallback;
}
