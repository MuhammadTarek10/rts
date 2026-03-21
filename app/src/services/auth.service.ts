import api from '@/lib/axios';
import type {
  ApiResponse,
  TokenResponse,
  User,
  SignInPayload,
  SignUpPayload,
  UpdateProfilePayload,
  ChangePasswordPayload,
} from '@/types';

export const authService = {
  signIn: (payload: SignInPayload) =>
    api.post<ApiResponse<TokenResponse>>('/auth/sign-in', payload),

  signUp: (payload: SignUpPayload) =>
    api.post<ApiResponse<TokenResponse>>('/auth/sign-up', payload),

  refresh: () => api.post<ApiResponse<TokenResponse>>('/auth/refresh'),

  signOut: () => api.post<ApiResponse<null>>('/auth/sign-out'),

  getProfile: () => api.get<ApiResponse<User>>('/auth/users/profile'),

  updateProfile: (payload: UpdateProfilePayload) =>
    api.patch<ApiResponse<User>>('/auth/users/profile', payload),

  changePassword: (payload: ChangePasswordPayload) =>
    api.patch<ApiResponse<null>>('/auth/users/change-password', payload),

  deleteAccount: () => api.delete<ApiResponse<null>>('/auth/users/me'),
};
