import { ApiProperty, ApiPropertyOptional } from '@nestjs/swagger';

class UserProfileDto {
  @ApiProperty({ example: 'uuid' })
  id: string;

  @ApiPropertyOptional({ example: 'John' })
  first_name: string | null;

  @ApiPropertyOptional({ example: 'Doe' })
  last_name: string | null;

  @ApiPropertyOptional({ example: 'https://example.com/avatar.png' })
  avatar_url: string | null;

  @ApiPropertyOptional({ example: '+1234567890' })
  phone_number: string | null;

  @ApiPropertyOptional({ example: '1990-01-01' })
  date_of_birth: string | null;

  @ApiPropertyOptional({ example: 'US' })
  country: string | null;

  @ApiPropertyOptional({
    example: 'Software engineer and open source enthusiast.',
  })
  bio: string | null;
}

export class ProfileResponseDto {
  @ApiProperty({ example: 'uuid' })
  id: string;

  @ApiProperty({ example: 'john.doe@example.com' })
  email: string;

  @ApiProperty({
    example: 'active',
    enum: ['pending', 'active', 'inactive', 'suspended'],
  })
  status: string;

  @ApiProperty({ example: 'user', enum: ['user', 'admin'] })
  role: string;

  @ApiPropertyOptional({ type: () => UserProfileDto })
  profile: UserProfileDto | null;

  @ApiProperty({ example: '2024-01-01T00:00:00.000Z' })
  created_at: Date;

  @ApiProperty({ example: '2024-01-01T00:00:00.000Z' })
  updated_at: Date;
}
