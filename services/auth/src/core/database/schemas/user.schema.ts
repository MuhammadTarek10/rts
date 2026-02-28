import { relations } from 'drizzle-orm';
import { pgTable, text, timestamp, uuid } from 'drizzle-orm/pg-core';
import { auth } from './auth.schema';
import { userProfile } from './user-profile.schema';
import { userStatus, userRoles } from './enums';

export const user = pgTable('users', {
  id: uuid('id').primaryKey().defaultRandom(),
  email: text('email').notNull().unique(),
  status: userStatus().notNull().default('active'),
  role: userRoles().notNull().default('user'),
  email_verified_at: timestamp('email_verified_at'),
  last_login_at: timestamp('last_login_at'),
  password_changed_at: timestamp('password_changed_at'),
  created_at: timestamp('created_at').notNull().defaultNow(),
  updated_at: timestamp('updated_at')
    .notNull()
    .defaultNow()
    .$onUpdate(() => new Date()),
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

export type User = typeof user.$inferSelect;
export type NewUser = typeof user.$inferInsert;
