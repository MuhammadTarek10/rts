import { registerAs } from '@nestjs/config';

export default registerAs('broker', () => ({
  RABBITMQ_URI: process.env.RABBITMQ_URI,
}));
