import { Injectable, UnauthorizedException } from '@nestjs/common';
import { ConfigService } from '@nestjs/config';
import { PassportStrategy } from '@nestjs/passport';
import { ExtractJwt, Strategy } from 'passport-jwt';
import { ENVIRONMENT } from 'src/common/constants';
import { TokenPayload, UserWithSession } from 'src/core/utils/token/types';
import { AuthService } from '../auth.service';

const extractJwtFromCookie = (req: unknown): string | null => {
  const request = req as Request & { cookies?: Record<string, string> };
  if (request?.cookies?.access_token) {
    return request.cookies.access_token;
  }
  return null;
};

@Injectable()
export class JwtStrategy extends PassportStrategy(Strategy, 'jwt') {
  constructor(
    config: ConfigService,
    private readonly authService: AuthService,
  ) {
    super({
      jwtFromRequest: ExtractJwt.fromExtractors([
        extractJwtFromCookie,
        ExtractJwt.fromAuthHeaderAsBearerToken(),
      ]),
      secretOrKey: config.getOrThrow(ENVIRONMENT.JWT.ACCESS_SECRET),
    });
  }

  async validate(payload: TokenPayload): Promise<UserWithSession> {
    const user = await this.authService.findByUserId(payload.id);
    if (user) return { ...user, session_id: payload.session_id! };

    throw new UnauthorizedException(
      'You are not authorized to access this resource',
    );
  }
}
