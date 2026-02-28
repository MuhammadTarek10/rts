import { pgEnum } from 'drizzle-orm/pg-core';
import { AUTH_STRATEGIES, USER_ROLES, USER_STATUS } from 'src/common/constants';

export const authStrategy = pgEnum('auth_strategy', AUTH_STRATEGIES);
export const userStatus = pgEnum('user_status', USER_STATUS);
export const userRoles = pgEnum('user_roles', USER_ROLES);
