import {
  Injectable,
  NotFoundException,
  UnauthorizedException,
} from '@nestjs/common';
import { UserRepository } from './user.repository';
import { UpdateProfileDto } from './dtos/update-profile.dto';
import { ChangePasswordDto } from './dtos/change-password.dto';
import { HashService } from 'src/core/utils/hash/hash.service';

@Injectable()
export class UsersService {
  constructor(
    private readonly repo: UserRepository,
    private readonly hashService: HashService,
  ) {}

  async getProfile(id: string) {
    const profile = await this.repo.findProfileById(id);
    if (!profile || profile.deleted_at)
      throw new NotFoundException('User not found');
    return profile;
  }

  async updateProfile(id: string, dto: UpdateProfileDto) {
    const existing = await this.repo.findProfileById(id);
    if (!existing || existing.deleted_at)
      throw new NotFoundException('User not found');
    await this.repo.updateProfile(id, dto);
    return await this.repo.findProfileById(id);
  }

  async deleteAccount(id: string) {
    await this.repo.runInTransaction(async () => {
      const [deleted] = await this.repo.softDelete(id);
      if (!deleted) throw new NotFoundException('User not found');
      await this.repo.deleteAllSessions(id);
    });
  }

  async changePassword(id: string, dto: ChangePasswordDto) {
    const authRecord = await this.repo.findLocalAuth(id);
    if (!authRecord?.password_hash)
      throw new NotFoundException('No local authentication found');

    const isMatch = await this.hashService.verify(
      authRecord.password_hash,
      dto.current_password,
    );
    if (!isMatch)
      throw new UnauthorizedException('Current password is incorrect');

    const newHash = await this.hashService.hash(dto.new_password);

    await this.repo.runInTransaction(async () => {
      await this.repo.updatePasswordHash(id, newHash);
      await this.repo.deleteAllSessions(id);
    });
  }
}
