import { Controller, Delete, Get, Patch, Post } from '@nestjs/common';
import { AuthService } from './auth.service';
import { ApiTags } from '@nestjs/swagger';

@ApiTags('Authentication')
@Controller()
export class AuthController {
  constructor(private readonly service: AuthService) {}

  @Post('sign-in')
  async signIn() {
    return await this.service.findUserByEmail('test@gmail.com');
  }

  @Post('sign-up')
  signUp() {
    console.log('Sign-up endpoint hit');
  }

  @Get('profile')
  getProfile() {
    console.log('Get profile endpoint hit');
  }

  @Patch('profile')
  updateProfile() {
    console.log('Update profile endpoint hit');
  }

  @Delete('profile')
  deleteProfile() {
    console.log('Delete profile endpoint hit');
  }
}
