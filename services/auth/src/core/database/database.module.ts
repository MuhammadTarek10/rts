import { Module } from '@nestjs/common';
import * as schema from './schemas';
import { drizzle } from 'drizzle-orm/postgres-js';
import { ENVIRONMENT } from 'src/common/constants';
import { ConfigService } from '@nestjs/config';
import postgres from 'postgres';

@Module({
  providers: [
    {
      provide: ENVIRONMENT.DATABASE.KEY,
      useFactory: (config: ConfigService) => {
        const client = postgres(
          config.getOrThrow<string>(ENVIRONMENT.DATABASE.URL),
          {
            prepare: false,
          },
        );
        return drizzle(client, { schema });
      },
      inject: [ConfigService],
    },
  ],
  exports: [ENVIRONMENT.DATABASE.KEY],
})
export class DatabaseModule {}
