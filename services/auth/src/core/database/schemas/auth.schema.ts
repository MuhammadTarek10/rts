import {
  pgTable,
  index,
  uuid,
  text,
  timestamp,
  integer,
  boolean,
} from 'drizzle-orm/pg-core';
import { user } from './user.schema';
import { relations } from 'drizzle-orm';
import { authStrategy } from './enums';

export const auth = pgTable(
  'auths',
  {
    id: uuid('id').primaryKey().defaultRandom(),
    user_id: uuid('user_id')
      .notNull()
      .references(() => user.id, { onDelete: 'cascade' }),
    strategy: authStrategy('strategy').notNull(),
    provider_user_id: text('provider_user_id'),
    is_primary: boolean('is_primary').notNull().default(false),
    password_hash: text('password_hash'),
    failed_attempts: integer('failed_attempts').notNull().default(0),
    locked_until: timestamp('locked_until'),
    last_used_at: timestamp('last_used_at'),
    created_at: timestamp('created_at').notNull().defaultNow(),
    updated_at: timestamp('updated_at')
      .notNull()
      .defaultNow()
      .$onUpdate(() => new Date()),
    deleted_at: timestamp('deleted_at'),
  },
  (table) => [
    index('auth_strategy_user_id_idx').on(table.strategy, table.user_id),
    index('auth_strategy_provider_user_id_idx').on(
      table.strategy,
      table.provider_user_id,
    ),
    index('auth_locked_until_idx').on(table.locked_until),
  ],
);

export const authRelations = relations(auth, ({ one }) => ({
  user: one(user, {
    fields: [auth.user_id],
    references: [user.id],
  }),
}));
