import { APIS } from '@/lib/apis';
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
  signIn: async (payload: SignInPayload) => {
    const response = await api.post<ApiResponse<TokenResponse>>(
      APIS.AUTH.SIGNIN,
      payload
    );
    return response.data;
  },

  signUp: async (payload: SignUpPayload) => {
    const response = await api.post<ApiResponse<TokenResponse>>(
      APIS.AUTH.SIGNUP,
      payload
    );
    return response.data;
  },

  refresh: async () => {
    const response = await api.post<ApiResponse<TokenResponse>>(
      APIS.AUTH.REFRESH
    );
    return response.data;
  },

  signOut: async () => {
    const response = await api.post<ApiResponse<null>>(APIS.AUTH.SIGNOUT);
    return response.data;
  },

  getProfile: async () => {
    const response = await api.get<ApiResponse<User>>(APIS.AUTH.GET_PROFILE);
    return response.data;
  },

  updateProfile: async (payload: UpdateProfilePayload) => {
    const response = await api.patch<ApiResponse<User>>(
      APIS.AUTH.UPDATE_PROFILE,
      payload
    );
    return response.data;
  },

  changePassword: async (payload: ChangePasswordPayload) => {
    const response = await api.patch<ApiResponse<null>>(
      APIS.AUTH.CHANGE_PASSWORD,
      payload
    );
    return response.data;
  },

  deleteAccount: async () => {
    const response = await api.delete<ApiResponse<null>>(
      APIS.AUTH.DELETE_ACCOUNT
    );
    return response.data;
  },
};
