import { defineConfig } from 'drizzle-kit';

export default defineConfig({
  out: './src/core/database/migrations',
  schema: './src/**/schemas/*.ts',
  dialect: 'postgresql',
  migrations: {
    table: 'migrations',
    schema: 'public',
  },
  dbCredentials: {
    url: process.env.DATABASE_URL!,
  },
  verbose: true,
  strict: true,
});
