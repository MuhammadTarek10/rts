import { Module } from '@nestjs/common';
import { TokenService } from './token/token.service';
import { CookieService } from './cookie/cookie.service';
import { HashService } from './hash/hash.service';
import { JwtModule } from '@nestjs/jwt';
import { RandomService } from './random/random.service';

@Module({
  imports: [JwtModule.register({})],
  providers: [TokenService, CookieService, HashService, RandomService],
  exports: [TokenService, CookieService, HashService, RandomService],
})
export class UtilsModule {}
