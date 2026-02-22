import { registerAs } from '@nestjs/config';

export default registerAs('app', () => ({
  env: process.env.NODE_ENV || 'development',
  port: process.env.PORT || 8000,
  frontend_url: process.env.FRONTEND_URL || 'http://localhost:3000',
}));
