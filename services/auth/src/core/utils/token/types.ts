import { User } from 'src/core/database/schemas';

export interface TokenPayload {
  id: string;
  email: string;
  session_id?: string;
}

export interface RefreshTokenPayload extends TokenPayload {
  session_id: string;
  refresh_token: string;
}

export interface TokenResponse {
  access_token: string;
  refresh_token: string;
  expires_in: number;
}

export interface UserWithSession extends User {
  session_id: string;
}
