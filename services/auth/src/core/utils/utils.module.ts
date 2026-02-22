import { Module } from '@nestjs/common';
import { TokenService } from './token/token.service';
import { CookieService } from './cookie/cookie.service';
import { HashService } from './hash/hash.service';
import { JwtModule } from '@nestjs/jwt';

@Module({
  imports: [JwtModule.register({})],
  providers: [TokenService, CookieService, HashService],
  exports: [TokenService, CookieService, HashService],
})
export class UtilsModule {}
