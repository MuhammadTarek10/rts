import { createParamDecorator, ExecutionContext } from '@nestjs/common';
import { AppRequest } from './types';
import { RefreshTokenPayload } from '../utils/token/types';

export const RefreshToken = createParamDecorator(
  (data: keyof RefreshTokenPayload, ctx: ExecutionContext) => {
    const request = ctx.switchToHttp().getRequest<AppRequest>();
    const user = request.user as RefreshTokenPayload;
    return data ? user?.[data] : user;
  },
);
