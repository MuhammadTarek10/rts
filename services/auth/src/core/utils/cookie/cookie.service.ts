import { Injectable } from '@nestjs/common';
import { ConfigService } from '@nestjs/config';
import { Response } from 'express';
import { ENVIRONMENT } from 'src/common/constants';
import { TOKEN } from './types';

@Injectable()
export class CookieService {
  private readonly isProduction: boolean;

  constructor(private readonly config: ConfigService) {
    this.isProduction = this.config.get(ENVIRONMENT.NODE_ENV) === 'production';
  }

  setAuthCookies(
    res: Response,
    accessToken: string,
    refreshToken: string,
    accessTokenExpiresIn: number,
    refreshTokenExpiresIn: number,
  ): void {
    const cookieOptions = this.getCookieOptions();

    // Set access token cookie
    res.cookie(TOKEN.ACCESS_TOKEN, accessToken, {
      ...cookieOptions,
      maxAge: accessTokenExpiresIn * 1000,
    });

    // Set refresh token cookie
    res.cookie(TOKEN.REFRESH_TOKEN, refreshToken, {
      ...cookieOptions,
      maxAge: refreshTokenExpiresIn * 1000,
    });
  }

  clearAuthCookies(res: Response): void {
    const cookieOptions = this.getCookieOptions();

    res.cookie(TOKEN.ACCESS_TOKEN, '', {
      ...cookieOptions,
      maxAge: 0,
    });

    res.cookie(TOKEN.REFRESH_TOKEN, '', {
      ...cookieOptions,
      maxAge: 0,
    });
  }

  /**
   * Get cookie options based on environment
   * @returns Cookie options object
   */
  private getCookieOptions() {
    return {
      httpOnly: true, // Prevents JavaScript access to cookies
      sameSite: 'lax' as const, // CSRF protection while allowing some cross-site requests
      secure: this.isProduction, // Only send cookie over HTTPS in production
      path: '/', // Cookie available for all routes
    };
  }
}
