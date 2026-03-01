import { ApiProperty } from '@nestjs/swagger';
import { IsString, Matches } from 'class-validator';
import { PASSWORD_REGEX } from 'src/auth/dtos/constants';

export class ChangePasswordDto {
  @ApiProperty({ example: 'OldP@ssw0rd' })
  @IsString()
  current_password: string;

  @ApiProperty({ example: 'NewP@ssw0rd1' })
  @IsString()
  @Matches(PASSWORD_REGEX, {
    message:
      'Password must contain at least one lowercase letter, one uppercase letter, one number, and one special character',
  })
  new_password: string;
}
