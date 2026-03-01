import { Injectable, UnauthorizedException } from '@nestjs/common';
import { ConfigService } from '@nestjs/config';
import { PassportStrategy } from '@nestjs/passport';
import { ExtractJwt, Strategy } from 'passport-jwt';
import { ENVIRONMENT } from 'src/common/constants';
import { AppRequest } from 'src/core/decorators/types';
import { RefreshTokenPayload } from 'src/core/utils/token/types';

const extractJwtFromCookie = (req: unknown): string | null => {
  const request = req as Request & { cookies?: Record<string, string> };
  if (request?.cookies?.refresh_token) {
    return request.cookies.refresh_token;
  }
  return null;
};

@Injectable()
export class RefreshTokenStrategy extends PassportStrategy(
  Strategy,
  'refresh',
) {
  constructor(config: ConfigService) {
    super({
      jwtFromRequest: ExtractJwt.fromExtractors([
        extractJwtFromCookie,
        ExtractJwt.fromAuthHeaderAsBearerToken(),
      ]),
      secretOrKey: config.getOrThrow(ENVIRONMENT.JWT.REFRESH_SECRET),
      passReqToCallback: true,
    });
  }

  validate(req: AppRequest, payload: RefreshTokenPayload): RefreshTokenPayload {
    if (!payload.session_id)
      throw new UnauthorizedException('Session ID missing from token');

    let refresh_token: string | undefined;

    refresh_token = req.cookies?.refresh_token;

    if (refresh_token) return { ...payload, refresh_token };

    refresh_token = req.headers?.authorization?.split(' ')[1];
    if (!refresh_token)
      throw new UnauthorizedException('Invalid refresh token');

    return {
      ...payload,
      refresh_token,
    };
  }
}
