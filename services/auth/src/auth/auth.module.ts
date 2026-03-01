import { Module } from '@nestjs/common';
import { AuthService } from './auth.service';
import { AuthController } from './auth.controller';
import { UtilsModule } from 'src/core/utils/utils.module';
import { DatabaseModule } from 'src/core/database/database.module';
import { AuthRepository } from './auth.repository';
import { SessionRepository } from './session.repository';

@Module({
  imports: [DatabaseModule, UtilsModule],
  controllers: [AuthController],
  providers: [AuthService, AuthRepository, SessionRepository],
})
export class AuthModule {}
