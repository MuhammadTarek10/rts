import { Module } from '@nestjs/common';
import { ConfigModule } from './core/config/config.module';
import { DatabaseModule } from './core/database/database.module';
import { UtilsModule } from './core/utils/utils.module';
import { AuthModule } from './auth/auth.module';
import { UsersModule } from './users/users.module';
import { BrokerModule } from './core/broker/broker.module';

@Module({
  imports: [
    ConfigModule,
    DatabaseModule,
    UtilsModule,
    BrokerModule,
    AuthModule,
    UsersModule,
  ],
})
export class AppModule {}
