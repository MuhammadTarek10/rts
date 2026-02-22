import {
  pgTable,
  index,
  uuid,
  pgEnum,
  text,
  timestamp,
} from 'drizzle-orm/pg-core';
import { user } from './user.schema';
import { relations } from 'drizzle-orm';
import { AUTH_STRATEGIES } from 'src/common/constants';

export const authStrategyEnum = pgEnum('auth_strategy', AUTH_STRATEGIES);

export const auth = pgTable(
  'auths',
  {
    id: uuid('id').primaryKey().defaultRandom(),
    user_id: uuid('user_id')
      .notNull()
      .references(() => user.id, { onDelete: 'cascade' }),
    strategy: authStrategyEnum('strategy').notNull(),
    password_hash: text('password_hash'),
    created_at: timestamp('created_at').notNull().defaultNow(),
    updated_at: timestamp('updated_at')
      .notNull()
      .defaultNow()
      .$onUpdate(() => new Date()),
  },
  (table) => [
    index('auth_strategy_user_id_idx').on(table.strategy, table.user_id),
  ],
);

export const authRelations = relations(auth, ({ one }) => ({
  user: one(user, {
    fields: [auth.user_id],
    references: [user.id],
  }),
}));
