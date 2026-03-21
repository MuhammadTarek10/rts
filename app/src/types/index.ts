export interface ApiResponse<T> {
  data: T;
  message: string;
  status: 'success' | 'error';
  error?: string;
}

export interface TokenResponse {
  access_token: string;
  refresh_token: string;
  expires_in: number;
}

export type UserStatus = 'pending' | 'active' | 'inactive' | 'suspended';
export type UserRole = 'user' | 'admin';

export interface UserProfile {
  first_name: string | null;
  last_name: string | null;
  avatar_url: string | null;
  phone_number: string | null;
  date_of_birth: string | null;
  country: string | null;
  bio: string | null;
}

export interface User {
  id: string;
  email: string;
  status: UserStatus;
  role: UserRole;
  profile: UserProfile | null;
  created_at: string;
  updated_at: string;
}

export interface SignInPayload {
  email: string;
  password: string;
}

export interface SignUpPayload {
  email: string;
  password: string;
  first_name?: string;
  last_name?: string;
  phone_number?: string;
  country?: string;
  date_of_birth?: string;
}

export interface UpdateProfilePayload {
  first_name?: string;
  last_name?: string;
  phone_number?: string;
  country?: string;
  date_of_birth?: string;
  bio?: string;
  avatar_url?: string;
}

export interface ChangePasswordPayload {
  current_password: string;
  new_password: string;
}

export interface SignInFormData {
  email: string;
  password: string;
}

export interface SignUpFormData {
  email: string;
  password: string;
  first_name: string;
  last_name: string;
}
