import { relations } from 'drizzle-orm';
import { integer, pgEnum, pgTable, timestamp, uuid } from 'drizzle-orm/pg-core';
import { USER_STATUS } from 'src/common/constants';
import { auth } from './auth.schema';
import { userProfile } from './user-profile.schema';

const userStatus = pgEnum('user_status', USER_STATUS);

export const user = pgTable('users', {
  id: uuid('id').primaryKey().defaultRandom(),
  status: userStatus().notNull().default('active'),
  email_verified_at: timestamp('email_verified_at'),
  last_login_at: timestamp('last_login_at'),
  password_changed_at: timestamp('password_changed_at'),
  refresh_token_version: integer('refresh_token_version').notNull().default(0),
  deleted_at: timestamp('deleted_at'),
});

export const userRelations = relations(user, ({ many, one }) => ({
  auths: many(auth, {
    relationName: 'user',
  }),
  profile: one(userProfile, {
    fields: [user.id],
    references: [userProfile.user_id],
  }),
}));
