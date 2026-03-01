import { Injectable } from '@nestjs/common';
import { and, eq } from 'drizzle-orm';
import { BaseRepository } from 'src/core/database/base.repository';
import { auth, session, user, userProfile } from 'src/core/database/schemas';
import { UpdateProfileDto } from './dtos/update-profile.dto';
import { AUTH_STRATEGIES } from 'src/common/constants';

@Injectable()
export class UserRepository extends BaseRepository {
  async findProfileById(id: string) {
    return await this.executeQuery(
      this.db.query.user.findFirst({
        where: eq(user.id, id),
        with: { profile: true },
      }),
    );
  }

  async updateProfile(userId: string, dto: UpdateProfileDto) {
    return await this.executeQuery(
      this.db
        .update(userProfile)
        .set(dto)
        .where(eq(userProfile.user_id, userId))
        .returning(),
    );
  }

  async softDelete(id: string) {
    return await this.executeQuery(
      this.db
        .update(user)
        .set({ deleted_at: new Date() })
        .where(eq(user.id, id))
        .returning(),
    );
  }

  async findLocalAuth(userId: string) {
    return await this.executeQuery(
      this.db.query.auth.findFirst({
        where: and(
          eq(auth.user_id, userId),
          eq(auth.strategy, AUTH_STRATEGIES.LOCAL),
        ),
      }),
    );
  }

  async updatePasswordHash(userId: string, hash: string) {
    return await this.executeQuery(
      this.db
        .update(auth)
        .set({ password_hash: hash })
        .where(
          and(
            eq(auth.user_id, userId),
            eq(auth.strategy, AUTH_STRATEGIES.LOCAL),
          ),
        )
        .returning(),
    );
  }

  async deleteAllSessions(userId: string) {
    const authRecord = await this.findLocalAuth(userId);
    if (!authRecord) return;
    return await this.executeQuery(
      this.db.delete(session).where(eq(session.auth_id, authRecord.id)),
    );
  }
}
