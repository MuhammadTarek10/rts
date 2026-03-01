import { z } from 'zod';

export const validationSchema = z.object({
  NODE_ENV: z
    .enum(['development', 'production', 'test', 'staging'])
    .default('development'),
  PORT: z.string().default('8001'),

  // * Frontend
  FRONTEND_URL: z.url().default('http://localhost:3000'),

  // * Database
  DATABASE_URL: z.string().min(1, 'DATABASE_URL is required'),

  // * JWT
  JWT_ACCESS_SECRET: z.string().min(1, 'JWT_ACCESS_SECRET is required'),
  JWT_ACCESS_EXPIRATION: z.string().default('86400'), // 1 day
  JWT_REFRESH_SECRET: z.string().min(1, 'JWT_REFRESH_SECRET is required'),
  JWT_REFRESH_EXPIRATION: z.string().default('604800'), // 7 days

  // * Swagger
  SWAGGER_USERNAME: z.string().min(1, 'SWAGGER_USERNAME is required'),
  SWAGGER_PASSWORD: z.string().min(1, 'SWAGGER_PASSWORD is required'),

  // * Broker
  RABBITMQ_URI: z.string().min(1, 'RABBITMQ_URI is required'),
});
