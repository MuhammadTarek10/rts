import { Module } from '@nestjs/common';
import { AuthService } from './auth.service';
import { AuthController } from './auth.controller';
import { UtilsModule } from 'src/core/utils/utils.module';
import { DatabaseModule } from 'src/core/database/database.module';
import { AuthRepository } from './auth.repository';
import { SessionRepository } from './session.repository';
import { LocalStrategy } from './strategies/local.strategy';
import { JwtStrategy } from './strategies/jwt.strategy';
import { RefreshTokenStrategy } from './strategies/refresh-jwt.strategy';
import { AuthEventsPublisher } from './auth-events.publisher';

@Module({
  imports: [DatabaseModule, UtilsModule],
  controllers: [AuthController],
  providers: [
    AuthService,
    AuthRepository,
    SessionRepository,
    LocalStrategy,
    JwtStrategy,
    RefreshTokenStrategy,
    AuthEventsPublisher,
  ],
})
export class AuthModule {}
