import { Injectable, UnauthorizedException } from '@nestjs/common';
import { ConfigService } from '@nestjs/config';
import { JwtService } from '@nestjs/jwt';
import { RefreshTokenPayload, TokenPayload, TokenResponse } from './types';
import { ENVIRONMENT } from 'src/common/constants';

@Injectable()
export class TokenService {
  constructor(
    private readonly jwtService: JwtService,
    private readonly config: ConfigService,
  ) {}

  async generateAccessToken(payload: TokenPayload): Promise<string> {
    return this.jwtService.signAsync(
      {
        ...payload,
        jti: this.generateJti(),
      },
      {
        secret: this.config.getOrThrow(ENVIRONMENT.JWT.ACCESS_SECRET),
        expiresIn: this.getAccessTokenExpiresIn(),
      },
    );
  }

  async generateRefreshToken(payload: TokenPayload): Promise<string> {
    return this.jwtService.signAsync(
      {
        ...payload,
        jti: this.generateJti(),
      },
      {
        secret: this.config.getOrThrow(ENVIRONMENT.JWT.REFRESH_SECRET),
        expiresIn: this.getRefreshTokenExpiresIn(),
      },
    );
  }

  getAccessTokenExpiresIn(): number {
    return Number(
      this.config.getOrThrow<number>(ENVIRONMENT.JWT.ACCESS_EXPIRATION),
    );
  }

  getRefreshTokenExpiresIn(): number {
    return Number(
      this.config.getOrThrow<number>(ENVIRONMENT.JWT.REFRESH_EXPIRATION),
    );
  }

  async generateTokens(payload: TokenPayload): Promise<TokenResponse> {
    const [access_token, refresh_token] = await Promise.all([
      this.jwtService.signAsync(
        {
          ...payload,
          jti: this.generateJti(),
        },
        {
          secret: this.config.getOrThrow(ENVIRONMENT.JWT.ACCESS_SECRET),
          expiresIn: this.getAccessTokenExpiresIn(),
        },
      ),
      this.jwtService.signAsync(
        {
          ...payload,
          jti: this.generateJti(),
        },
        {
          secret: this.config.getOrThrow(ENVIRONMENT.JWT.REFRESH_SECRET),
          expiresIn: this.config.getOrThrow<number>(
            ENVIRONMENT.JWT.REFRESH_EXPIRATION,
          ),
        },
      ),
    ]);
    return {
      access_token,
      refresh_token,
      expires_in: this.getAccessTokenExpiresIn(),
    };
  }

  async verifyAccessToken(token: string): Promise<TokenPayload> {
    try {
      return await this.jwtService.verifyAsync(token, {
        secret: this.config.getOrThrow(ENVIRONMENT.JWT.ACCESS_SECRET),
      });
    } catch {
      throw new UnauthorizedException('Invalid access token');
    }
  }

  async verifyRefreshToken(token: string): Promise<RefreshTokenPayload> {
    try {
      return await this.jwtService.verifyAsync(token, {
        secret: this.config.getOrThrow(ENVIRONMENT.JWT.REFRESH_SECRET),
      });
    } catch {
      throw new UnauthorizedException('Invalid refresh token');
    }
  }

  generateJti(): string {
    return `${Date.now()}-${Math.random().toString(36).substring(2, 15)}`;
  }
}
