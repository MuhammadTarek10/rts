import {
  PostgresJsDatabase,
  PostgresJsQueryResultHKT,
} from 'drizzle-orm/postgres-js';
import * as schema from './schemas';
import { PgTransaction } from 'drizzle-orm/pg-core';
import { ExtractTablesWithRelations } from 'drizzle-orm';

export type DrizzleSchema = typeof schema;
export type PostgresDatabase = PostgresJsDatabase<DrizzleSchema>;
export type PostgresTransaction = PgTransaction<
  PostgresJsQueryResultHKT,
  DrizzleSchema,
  ExtractTablesWithRelations<DrizzleSchema>
>;
