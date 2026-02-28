import { ConflictException, Injectable } from '@nestjs/common';
import { HashService } from 'src/core/utils/hash/hash.service';
import { TokenService } from 'src/core/utils/token/token.service';
import { AuthRepository } from './auth.repository';
import { SignUpDto } from './dtos/sign-up.dto';
import { User } from 'src/core/database/schemas';

@Injectable()
export class AuthService {
  constructor(
    private readonly repo: AuthRepository,
    private readonly hashService: HashService,
    private readonly tokenService: TokenService,
  ) {}

  async findUserByEmail(email: string) {
    return await this.repo.findUserByEmail(email);
  }

  async signUp(dto: SignUpDto) {
    const user = await this.repo.findUserByEmail(dto.email);
    if (user?.deleted_at) this.restoreUser(user);
    if (user) throw new ConflictException('User already Exists');

    const hash = await this.hashService.hash(dto.password);
    const data = await this.repo.createUser(dto, hash);

    return { ...data.user, ...data.profile };
  }

  private restoreUser(user: User) {
    console.log({ user });
  }
}
