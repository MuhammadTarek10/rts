import { NestFactory } from '@nestjs/core';
import { AppModule } from './app.module';
import { ConfigService } from '@nestjs/config';
import { DocumentBuilder, SwaggerModule } from '@nestjs/swagger';
import { ENVIRONMENT } from './common/constants';
import basicAuth from 'express-basic-auth';
import cookieParser from 'cookie-parser';
import { ValidationPipe, VersioningType } from '@nestjs/common';

async function bootstrap() {
  const app = await NestFactory.create(AppModule);

  app.setGlobalPrefix('api');
  app.use(cookieParser());
  app.useGlobalPipes(
    new ValidationPipe({
      whitelist: true,
      transform: true,
    }),
  );

  app.enableVersioning({
    type: VersioningType.URI,
    prefix: 'v',
    defaultVersion: '1',
  });
  const config = app.get(ConfigService);
  const frontendUrl = config.get<string>(ENVIRONMENT.FRONTEND_URL);
  const allowedOrigins = [frontendUrl].filter(Boolean);
  app.enableCors({
    origin: allowedOrigins.length > 0 ? allowedOrigins : '*',
    methods: ['GET', 'POST', 'PUT', 'PATCH', 'DELETE', 'OPTIONS'],
    credentials: true,
  });

  const docConfig = new DocumentBuilder()
    .setTitle('Auth Service API')
    .setDescription('API documentation for the Auth Service')
    .setVersion('1.0')
    .addBearerAuth({
      type: 'http',
      scheme: 'bearer',
      bearerFormat: 'JWT',
      description:
        'JWT Bearer token for access. Alternatively, use HTTP-only cookies.',
    })
    .addCookieAuth('access_token', {
      type: 'apiKey',
      in: 'cookie',
      name: 'access_token',
      description: 'HTTP-only cookie containing the access token',
    })
    .addCookieAuth('refresh_token', {
      type: 'apiKey',
      in: 'cookie',
      name: 'refresh_token',
      description: 'HTTP-only cookie containing the refresh token',
    })
    .build();

  const document = SwaggerModule.createDocument(app, docConfig);

  app.use(
    ['/docs', '/docs-json'],
    basicAuth({
      challenge: true,
      users: {
        [config.getOrThrow<string>(ENVIRONMENT.SWAGGER.USERNAME) || 'admin']:
          config.getOrThrow<string>(ENVIRONMENT.SWAGGER.PASSWORD) || 'admin',
      },
    }),
  );
  SwaggerModule.setup('docs', app, document);

  const port = config.getOrThrow<number>(ENVIRONMENT.PORT);

  await app.listen(port);
}

void bootstrap();
