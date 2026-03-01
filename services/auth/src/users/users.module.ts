import { Module } from '@nestjs/common';
import { UsersService } from './users.service';
import { UsersController } from './users.controller';
import { UserRepository } from './user.repository';
import { DatabaseModule } from 'src/core/database/database.module';
import { UtilsModule } from 'src/core/utils/utils.module';

@Module({
  imports: [DatabaseModule, UtilsModule],
  controllers: [UsersController],
  providers: [UsersService, UserRepository],
})
export class UsersModule {}
