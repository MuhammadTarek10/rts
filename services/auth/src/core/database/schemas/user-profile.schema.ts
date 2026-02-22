import { relations } from 'drizzle-orm';
import {
  date,
  pgTable,
  text,
  timestamp,
  uuid,
  uniqueIndex,
} from 'drizzle-orm/pg-core';
import { user } from './user.schema';

export const userProfile = pgTable(
  'user_profiles',
  {
    id: uuid('id').primaryKey().defaultRandom(),
    user_id: uuid('user_id')
      .notNull()
      .references(() => user.id, { onDelete: 'cascade' }),
    first_name: text('first_name'),
    last_name: text('last_name'),
    avatar_url: text('avatar_url'),
    phone_number: text('phone_number'),
    date_of_birth: date('date_of_birth'),
    country: text('country'),
    bio: text('bio'),
    created_at: timestamp('created_at').notNull().defaultNow(),
    updated_at: timestamp('updated_at')
      .notNull()
      .defaultNow()
      .$onUpdate(() => new Date()),
  },
  (table) => [uniqueIndex('user_profile_user_id_idx').on(table.user_id)],
);

export const userProfileRelations = relations(userProfile, ({ one }) => ({
  user: one(user, {
    fields: [userProfile.user_id],
    references: [user.id],
  }),
}));
