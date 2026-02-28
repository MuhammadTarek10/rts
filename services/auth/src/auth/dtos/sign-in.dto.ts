import { ApiProperty } from '@nestjs/swagger';
import { IsEmail, Matches } from 'class-validator';
import { PASSWORD_REGEX } from './constants';

export class SignInDto {
  @ApiProperty({
    description: 'The email of the user',
    example: 'user@exmple.com',
  })
  @IsEmail()
  email: string;

  @ApiProperty({
    description: 'The password of the user',
    example: 'Text@123',
  })
  @Matches(PASSWORD_REGEX, {
    message:
      'Password must contain at least one lowercase letter, one uppercase letter, one number, and one special character',
  })
  password: string;
}
