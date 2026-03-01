import { Request } from 'express';
import { RefreshTokenPayload, UserWithSession } from '../utils/token/types';

export interface AppRequest extends Request {
  user: UserWithSession | RefreshTokenPayload;
  headers: {
    authorization?: string;
    [key: string]: string | undefined;
  };
  cookies: Record<string, string>;
}
