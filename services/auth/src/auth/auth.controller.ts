import { Controller, Post } from '@nestjs/common';
import { AuthService } from './auth.service';
import { ApiTags } from '@nestjs/swagger';

@ApiTags('Authentication')
@Controller()
export class AuthController {
  constructor(private readonly service: AuthService) {}

  @Post('sign-in')
  signIn() {
    console.log('Sign-in endpoint hit');
  }
}
