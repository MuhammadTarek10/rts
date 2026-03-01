import {
  ConflictException,
  Injectable,
  UnauthorizedException,
} from '@nestjs/common';
import { HashService } from 'src/core/utils/hash/hash.service';
import { TokenService } from 'src/core/utils/token/token.service';
import { AuthRepository } from './auth.repository';
import { SignUpDto } from './dtos/sign-up.dto';
import { User } from 'src/core/database/schemas';
import { AUTH_STRATEGIES } from 'src/common/constants';
import { SessionRepository } from './session.repository';
import { RefreshTokenPayload } from 'src/core/utils/token/types';
import { RandomService } from 'src/core/utils/random/random.service';

@Injectable()
export class AuthService {
  constructor(
    private readonly repo: AuthRepository,
    private readonly sessionRepo: SessionRepository,
    private readonly hashService: HashService,
    private readonly tokenService: TokenService,
    private readonly randomService: RandomService,
  ) {}

  async validateUser(email: string, password: string) {
    const user = await this.repo.findByEmailWithLocalAuth(email);
    if (!user || user.deleted_at) return null;

    if (
      !user.auths.length ||
      !user.auths[0].password_hash ||
      user.auths[0].strategy !== AUTH_STRATEGIES.LOCAL
    )
      return null;

    const isMatch = await this.hashService.verify(
      user.auths[0].password_hash,
      password,
    );
    if (!isMatch) {
      // increment failed attempts and lock account if necessary
      await this.repo.updateFailedAttempts(user.id);
      return null;
    }

    await this.repo.resetFailedAttempts(user.id);
    return user;
  }

  async signUp(dto: SignUpDto) {
    return await this.repo.runInTransaction(async () => {
      const user = await this.repo.findUserByEmail(dto.email);
      if (user?.deleted_at) this.restoreUser(user);
      if (user) throw new ConflictException('User already Exists');

      const hash = await this.hashService.hash(dto.password);
      const data = await this.repo.createUser(dto, hash);
      const sessionId = this.randomService.generateRandomUUID();
      const tokens = await this.tokenService.generateTokens({
        ...data.user,
        session_id: sessionId,
      });

      await this.createSession(sessionId, data.user.id, tokens.refresh_token);

      return tokens;
    });
  }

  async signIn(user: User) {
    const sessionId = this.randomService.generateRandomUUID();
    const tokens = await this.tokenService.generateTokens({
      ...user,
      session_id: sessionId,
    });
    await this.createSession(sessionId, user.id, tokens.refresh_token);
    return tokens;
  }

  async refresh(payload: RefreshTokenPayload) {
    return await this.repo.runInTransaction(async () => {
      const user = await this.repo.findByUserId(payload.id);
      if (!user) throw new UnauthorizedException('User not found');

      const auth = await this.repo.findByUserId(user.id);
      if (!auth) throw new UnauthorizedException();

      const session = await this.sessionRepo.findById(payload.session_id);
      if (!session) throw new UnauthorizedException('Invalid session');

      const isMatch = await this.hashService.verify(
        session.refresh_token_hash,
        payload.refresh_token,
      );
      if (!isMatch) throw new UnauthorizedException('Invalid refresh token');

      const tokens = await this.tokenService.generateTokens({
        ...user,
        session_id: session.id,
      });
      await this.updateSession(session.id, tokens.refresh_token);

      return tokens;
    });
  }

  async logout(userId: string, sessionId?: string) {
    if (sessionId) {
      await this.sessionRepo.deleteById(sessionId);
    } else {
      const auth = await this.repo.findByUserId(userId);
      if (auth) {
        await this.sessionRepo.deleteByAuthId(auth.id);
      }
    }
  }

  async findByUserId(id: string) {
    return await this.repo.findByUserId(id);
  }

  private async createSession(
    id: string,
    userId: string,
    refreshToken: string,
  ) {
    const hash = await this.hashService.hash(refreshToken);
    const expiresAt = new Date(
      Date.now() + this.tokenService.getRefreshTokenExpiresIn() * 1000,
    );
    await this.sessionRepo.create(id, userId, hash, expiresAt);
  }

  private async updateSession(id: string, refreshToken: string) {
    const hash = await this.hashService.hash(refreshToken);
    const expiresAt = new Date(
      Date.now() + this.tokenService.getRefreshTokenExpiresIn() * 1000,
    );
    await this.sessionRepo.update(id, hash, expiresAt);
  }

  private restoreUser(user: User) {
    console.log({ user });
  }
}
