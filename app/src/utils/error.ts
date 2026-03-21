import axios from 'axios';

interface ErrorResponse {
  message?: string | string[];
  error?: string;
  statusCode?: number;
}

export function getErrorMessage(
  err: unknown,
  fallback = 'An unexpected error occurred'
): string {
  if (axios.isAxiosError(err)) {
    if (!err.response) {
      return 'Unable to connect. Please check your internet connection.';
    }

    const data = err.response.data as ErrorResponse | undefined;

    if (data?.message) {
      if (Array.isArray(data.message)) {
        return data.message[0] || fallback;
      }
      return data.message;
    }
  }

  return fallback;
}
