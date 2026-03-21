import axios from 'axios';
import type { AxiosError, InternalAxiosRequestConfig } from 'axios';
import { APIS } from './apis';

const api = axios.create({
  baseURL: import.meta.env.VITE_API_BASE_URL || '/api',
  withCredentials: true,
  headers: { 'Content-Type': 'application/json' },
});

let isRefreshing = false;
let failedQueue: Array<{
  resolve: (value?: unknown) => void;
  reject: (reason?: unknown) => void;
}> = [];

const processQueue = (error: AxiosError | null) => {
  failedQueue.forEach(({ resolve, reject }) => {
    if (error) reject(error);
    else resolve();
  });
  failedQueue = [];
};

api.interceptors.response.use(
  (response) => response,
  async (error: AxiosError) => {
    const originalRequest = error.config as InternalAxiosRequestConfig & {
      _retry?: boolean;
    };

    if (
      error.response?.status !== 401 ||
      originalRequest._retry ||
      originalRequest.url?.includes(APIS.AUTH.REFRESH) ||
      originalRequest.url?.includes(APIS.AUTH.SIGNIN) ||
      originalRequest.url?.includes(APIS.AUTH.SIGNUP)
    ) {
      throw error;
    }

    if (isRefreshing) {
      return new Promise((resolve, reject) => {
        failedQueue.push({ resolve, reject });
      }).then(() => api(originalRequest));
    }

    originalRequest._retry = true;
    isRefreshing = true;

    try {
      await api.post(APIS.AUTH.REFRESH);
      processQueue(null);
      return api(originalRequest);
    } catch (refreshError) {
      processQueue(refreshError as AxiosError);
      const { useUserStore } = await import('@/stores/user');
      const store = useUserStore();
      store.clearUser();
      globalThis.location.href = '/auth/sign-in';
      throw refreshError;
    } finally {
      isRefreshing = false;
    }
  }
);

export default api;
