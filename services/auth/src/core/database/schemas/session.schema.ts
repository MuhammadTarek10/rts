import { index, pgTable, uuid, timestamp, text } from 'drizzle-orm/pg-core';
import { auth } from './auth.schema';
import { relations } from 'drizzle-orm';

export const session = pgTable(
  'sessions',
  {
    id: uuid().primaryKey().defaultRandom(),
    auth_id: uuid()
      .notNull()
      .references(() => auth.id, { onDelete: 'cascade' }),
    refresh_token_hash: text('refresh_token_hash').notNull(),
    expires_at: timestamp('expires_at').notNull(),
    revoked_at: timestamp('revoked_at'),
    ip_address: text('ip_address'),
    user_agent: text('user_agent'),
    device_id: text('device_id'),
    created_at: timestamp('created_at').notNull().defaultNow(),
    deleted_at: timestamp('deleted_at'),
  },
  (table) => [
    index('sessions_auth_id_idx').on(table.auth_id),
    index('sessions_refresh_token_hash_idx').on(table.refresh_token_hash),
    index('sessions_expires_at_idx').on(table.expires_at),
    index('sessions_revoked_at_idx').on(table.revoked_at),
  ],
);

export const sessionsRelations = relations(session, ({ one }) => ({
  auth: one(auth, {
    fields: [session.auth_id],
    references: [auth.id],
  }),
}));
