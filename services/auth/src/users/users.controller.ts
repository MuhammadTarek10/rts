import {
  Body,
  Controller,
  Delete,
  Get,
  HttpCode,
  HttpStatus,
  Patch,
  UseGuards,
} from '@nestjs/common';
import { UsersService } from './users.service';
import {
  ApiBadRequestResponse,
  ApiBearerAuth,
  ApiCookieAuth,
  ApiExtraModels,
  ApiNoContentResponse,
  ApiOkResponse,
  ApiOperation,
  ApiTags,
  getSchemaPath,
} from '@nestjs/swagger';
import { JwtGuard } from 'src/auth/guards/jwt.guard';
import { CurrentUser } from 'src/core/decorators/current-user.decorator';
import { ResponseMessage } from 'src/core/decorators/response-message.decorator';
import { ResponseDto } from 'src/common/dtos/response.dto';
import { UpdateProfileDto } from './dtos/update-profile.dto';
import { ProfileResponseDto } from './dtos/profile-response.dto';
import { ChangePasswordDto } from './dtos/change-password.dto';
import { TOKEN } from 'src/core/utils/cookie/types';

@ApiTags('Identity')
@ApiExtraModels(ResponseDto, ProfileResponseDto)
@ApiBearerAuth()
@ApiCookieAuth(TOKEN.ACCESS_TOKEN)
@UseGuards(JwtGuard)
@Controller('users')
export class UsersController {
  constructor(private readonly service: UsersService) {}

  @ApiOperation({
    summary: 'Get current user profile',
    description:
      'Returns the authenticated user profile including profile details',
  })
  @ApiOkResponse({
    description: 'Profile retrieved successfully',
    schema: {
      allOf: [
        { $ref: getSchemaPath(ResponseDto) },
        {
          properties: {
            data: { $ref: getSchemaPath(ProfileResponseDto) },
          },
        },
      ],
    },
  })
  @ResponseMessage('Profile retrieved successfully')
  @Get('profile')
  async getProfile(@CurrentUser('id') userId: string) {
    return await this.service.getProfile(userId);
  }

  @ApiOperation({
    summary: 'Update current user profile',
    description: 'Partially updates the authenticated user profile',
  })
  @ApiOkResponse({
    description: 'Profile updated successfully',
    schema: {
      allOf: [
        { $ref: getSchemaPath(ResponseDto) },
        {
          properties: {
            data: { $ref: getSchemaPath(ProfileResponseDto) },
          },
        },
      ],
    },
  })
  @ApiBadRequestResponse({
    description: 'Bad request',
    schema: { allOf: [{ $ref: getSchemaPath(ResponseDto) }] },
  })
  @ResponseMessage('Profile updated successfully')
  @HttpCode(HttpStatus.OK)
  @Patch('profile')
  async updateProfile(
    @CurrentUser('id') userId: string,
    @Body() dto: UpdateProfileDto,
  ) {
    return await this.service.updateProfile(userId, dto);
  }

  @ApiOperation({
    summary: 'Delete current user account',
    description: 'Soft-deletes the authenticated user account',
  })
  @ApiNoContentResponse({ description: 'Account deleted successfully' })
  @ResponseMessage('Account deleted successfully')
  @HttpCode(HttpStatus.NO_CONTENT)
  @Delete('me')
  async deleteAccount(@CurrentUser('id') userId: string) {
    await this.service.deleteAccount(userId);
  }

  @ApiOperation({
    summary: 'Change current user password',
    description:
      'Verifies the current password, updates it, and invalidates all active sessions',
  })
  @ApiOkResponse({ description: 'Password changed successfully' })
  @ApiBadRequestResponse({
    description: 'Bad request',
    schema: { allOf: [{ $ref: getSchemaPath(ResponseDto) }] },
  })
  @ResponseMessage('Password changed successfully')
  @HttpCode(HttpStatus.OK)
  @Patch('change-password')
  async changePassword(
    @CurrentUser('id') userId: string,
    @Body() dto: ChangePasswordDto,
  ) {
    await this.service.changePassword(userId, dto);
  }
}
