import {
  Body,
  Controller,
  HttpCode,
  HttpStatus,
  Post,
  Res,
  UseGuards,
} from '@nestjs/common';
import { AuthService } from './auth.service';
import {
  ApiBadRequestResponse,
  ApiBearerAuth,
  ApiBody,
  ApiCookieAuth,
  ApiCreatedResponse,
  ApiExtraModels,
  ApiOkResponse,
  ApiOperation,
  ApiTags,
  getSchemaPath,
} from '@nestjs/swagger';
import { SignUpDto } from './dtos/sign-up.dto';
import { LocalGuard } from './guards/local.guard';
import { CurrentUser } from 'src/core/decorators/current-user.decorator';
import { RefreshToken } from 'src/core/decorators/refresh-token.decorator';
import { User } from 'src/core/database/schemas';
import { RefreshGuard } from './guards/refresh-jwt.guard';
import { RefreshTokenPayload } from 'src/core/utils/token/types';
import { ResponseDto } from 'src/common/dtos/response.dto';
import { ResponseMessage } from 'src/core/decorators/response-message.decorator';
import { CookieService } from 'src/core/utils/cookie/cookie.service';
import { Response } from 'express';
import { SignInDto } from './dtos/sign-in.dto';
import { TOKEN } from 'src/core/utils/cookie/types';
import { JwtGuard } from './guards/jwt.guard';
import { TokenResponseDto } from './dtos/token-response.dto';

@ApiTags('Authentication')
@ApiExtraModels(ResponseDto, TokenResponseDto)
@Controller()
export class AuthController {
  constructor(
    private readonly service: AuthService,
    private readonly cookieService: CookieService,
  ) {}

  @ApiOperation({
    summary: 'Sign in a user',
    description:
      'Authenticates a user and returns tokens in both response body and HTTP-only cookies',
  })
  @ApiBody({ type: SignInDto })
  @ApiOkResponse({
    description:
      'User signed in successfully. Tokens are returned in response and set as HTTP-only cookies',
    schema: {
      allOf: [
        { $ref: getSchemaPath(ResponseDto) },
        {
          properties: {
            data: { $ref: getSchemaPath(TokenResponseDto) },
          },
        },
      ],
    },
  })
  @ApiBadRequestResponse({
    description: 'Bad request',
    schema: {
      allOf: [{ $ref: getSchemaPath(ResponseDto) }],
    },
  })
  @UseGuards(LocalGuard)
  @ResponseMessage('User signed in successfully')
  @HttpCode(HttpStatus.OK)
  @Post('sign-in')
  async signIn(
    @CurrentUser() user: User,
    @Res({ passthrough: true }) res: Response,
  ) {
    const tokens = await this.service.signIn(user);

    this.cookieService.setAuthCookies(
      res,
      tokens.access_token,
      tokens.refresh_token,
    );

    return tokens;
  }

  @ApiOperation({
    summary: 'Sign up a new user',
    description:
      'Creates a new user account and returns tokens in both response body and HTTP-only cookies',
  })
  @ApiBody({ type: SignUpDto })
  @ApiCreatedResponse({
    description:
      'User signed up successfully. Tokens are returned in response and set as HTTP-only cookies',
    schema: {
      allOf: [
        { $ref: getSchemaPath(ResponseDto) },
        {
          properties: {
            data: { $ref: getSchemaPath(TokenResponseDto) },
          },
        },
      ],
    },
  })
  @ApiBadRequestResponse({
    description: 'Bad request',
    schema: {
      allOf: [{ $ref: getSchemaPath(ResponseDto) }],
    },
  })
  @ResponseMessage('User signed up successfully')
  @Post('sign-up')
  async signUp(
    @Body() dto: SignUpDto,
    @Res({ passthrough: true }) res: Response,
  ) {
    const tokens = await this.service.signUp(dto);

    this.cookieService.setAuthCookies(
      res,
      tokens.access_token,
      tokens.refresh_token,
    );

    return tokens;
  }

  @ApiOperation({
    summary: "Refresh a user's token",
    description:
      'Use Bearer token or HTTP-only cookie (refresh_token) for authentication',
  })
  @ApiBearerAuth()
  @ApiCookieAuth(TOKEN.REFRESH_TOKEN)
  @ApiOkResponse({
    description: 'Token refreshed successfully',
    schema: {
      allOf: [
        { $ref: getSchemaPath(ResponseDto) },
        {
          properties: {
            data: { $ref: getSchemaPath(TokenResponseDto) },
          },
        },
      ],
    },
  })
  @ResponseMessage('Token refreshed successfully')
  @HttpCode(HttpStatus.OK)
  @UseGuards(RefreshGuard)
  @Post('refresh')
  async refreshTokens(
    @RefreshToken() user: RefreshTokenPayload,
    @Res({ passthrough: true }) res: Response,
  ) {
    const tokens = await this.service.refresh(user);

    this.cookieService.setAuthCookies(
      res,
      tokens.access_token,
      tokens.refresh_token,
    );

    return tokens;
  }

  @ApiOperation({
    summary: 'Sign out a user',
    description:
      'Use Bearer token or HTTP-only cookie (access_token) for authentication. Clears all authentication cookies',
  })
  @ApiOkResponse({ description: 'User signed out successfully' })
  @ApiBadRequestResponse({
    description: 'Bad request',
    schema: {
      allOf: [{ $ref: getSchemaPath(ResponseDto) }],
    },
  })
  @ResponseMessage('User signed out successfully')
  @ApiBearerAuth()
  @ApiCookieAuth('access_token')
  @UseGuards(JwtGuard)
  @HttpCode(HttpStatus.OK)
  @Post('logout')
  async logout(
    @CurrentUser('id') userId: string,
    @CurrentUser('session_id') sessionId: string,
    @Res({ passthrough: true }) res: Response,
  ) {
    await this.service.logout(userId, sessionId);
    this.cookieService.clearAuthCookies(res);
    return null;
  }
}
