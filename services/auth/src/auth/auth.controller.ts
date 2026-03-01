import { Body, Controller, Post, UseGuards } from '@nestjs/common';
import { AuthService } from './auth.service';
import { ApiTags } from '@nestjs/swagger';
import { SignUpDto } from './dtos/sign-up.dto';
import { LocalGuard } from './guards/local.guard';
import { CurrentUser } from 'src/core/decorators/current-user.decorator';
import { RefreshToken } from 'src/core/decorators/refresh-token.decorator';
import { User } from 'src/core/database/schemas';
import { RefreshGuard } from './guards/refresh-jwt.guard';
import { RefreshTokenPayload } from 'src/core/utils/token/types';

@ApiTags('Authentication')
@Controller()
export class AuthController {
  constructor(private readonly service: AuthService) {}

  @UseGuards(LocalGuard)
  @Post('sign-in')
  async signIn(@CurrentUser() user: User) {
    return this.service.signIn(user);
  }

  @Post('sign-up')
  async signUp(@Body() dto: SignUpDto) {
    return await this.service.signUp(dto);
  }

  @UseGuards(RefreshGuard)
  @Post('refresh')
  async refreshTokens(@RefreshToken() user: RefreshTokenPayload) {
    return await this.service.refresh(user);
  }

  @UseGuards(RefreshGuard)
  @Post('logout')
  async logout(
    @CurrentUser('id') userId: string,
    @RefreshToken('session_id') sessionId: string,
  ) {
    return await this.service.logout(userId, sessionId);
  }
}
