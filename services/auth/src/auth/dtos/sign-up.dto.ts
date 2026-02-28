import { ApiProperty, ApiPropertyOptional } from '@nestjs/swagger';
import {
  IsEmail,
  IsOptional,
  IsString,
  Length,
  Matches,
  MinLength,
} from 'class-validator';
import { PASSWORD_REGEX } from './constants';

export class SignUpDto {
  @ApiProperty({
    description: 'The email of the user',
    example: 'user@example.com',
  })
  @IsEmail()
  email: string;

  @ApiProperty({
    description: 'The password of the user',
    example: 'P@ssw0rd',
  })
  @IsString()
  @Matches(PASSWORD_REGEX, {
    message:
      'Password must contain at least one lowercase letter, one uppercase letter, one number, and one special character',
  })
  password: string;

  @ApiPropertyOptional({
    description: 'The first name of the user',
    example: 'John',
  })
  @IsOptional()
  @Length(1, 50, { message: 'First name must be between 1 and 50 characters' })
  first_name: string;

  @ApiPropertyOptional({
    description: 'The last name of the user',
    example: 'Doe',
  })
  @IsOptional()
  @Length(1, 50, { message: 'Last name must be between 1 and 50 characters' })
  last_name: string;

  @ApiPropertyOptional({
    description: 'The phone number of the user',
    example: '+1234567890',
  })
  @IsOptional()
  @MinLength(7, { message: 'Phone number must be at least 7 characters' })
  phone_number: string;

  @ApiPropertyOptional({
    description: 'The country of the user',
    example: 'USA',
  })
  @IsOptional()
  @Length(2, 50, { message: 'Country must be between 2 and 50 characters' })
  country: string;

  @ApiPropertyOptional({
    description: 'date of birth of the user',
    example: '1990-01-01',
  })
  @IsOptional()
  date_of_birth: string;
}
