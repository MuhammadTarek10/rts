import { createParamDecorator, ExecutionContext } from '@nestjs/common';
import { UserWithSession } from '../utils/token/types';
import { AppRequest } from './types';

export const CurrentUser = createParamDecorator(
  (data: keyof UserWithSession, ctx: ExecutionContext) => {
    const request = ctx.switchToHttp().getRequest<AppRequest>();
    const user = request.user as UserWithSession;
    return data ? user?.[data] : user;
  },
);
