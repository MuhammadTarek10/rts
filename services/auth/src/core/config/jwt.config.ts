import { registerAs } from '@nestjs/config';

export default registerAs('jwt', () => ({
  access_secret: process.env.JWT_ACCESS_SECRET,
  access_expiration: process.env.JWT_ACCESS_EXPIRATION || '86400', // 1 day
  refresh_secret: process.env.JWT_REFRESH_SECRET,
  refresh_expiration: process.env.JWT_REFRESH_EXPIRATION || '604800', // 7 days
}));
