import { Injectable } from '@nestjs/common';
import { and, eq } from 'drizzle-orm';
import { BaseRepository } from 'src/core/database/base.repository';
import { auth, user, userProfile } from 'src/core/database/schemas';
import { SignUpDto } from './dtos/sign-up.dto';
import {
  AUTH_STRATEGIES,
  LOCKDOWN_DURATION,
  LOCKDOWN_THRESHOLD,
  USER_ROLES,
  USER_STATUS,
} from 'src/common/constants';

@Injectable()
export class AuthRepository extends BaseRepository {
  async createUser(dto: SignUpDto, hash: string) {
    const newUser = await this.executeQuery(
      this.db
        .insert(user)
        .values({
          email: dto.email,
          role: USER_ROLES.USER,
          status: USER_STATUS.PENDING,
        })
        .returning(),
    );

    await this.executeQuery(
      this.db
        .insert(userProfile)
        .values({
          user_id: newUser[0].id,
          first_name: dto.first_name,
          last_name: dto.last_name,
          phone_number: dto.phone_number,
          country: dto.country,
          date_of_birth: dto.date_of_birth,
        })
        .returning(),
    );

    await this.executeQuery(
      this.db.insert(auth).values({
        user_id: newUser[0].id,
        password_hash: hash,
        strategy: AUTH_STRATEGIES.LOCAL,
        is_primary: true,
      }),
    );

    return { user: newUser[0] };
  }

  async findUserByEmail(email: string) {
    return await this.executeQuery(
      this.db.query.user.findFirst({
        where: eq(user.email, email),
      }),
    );
  }

  async findByUserId(id: string) {
    return await this.executeQuery(
      this.db.query.user.findFirst({
        where: eq(user.id, id),
        with: {
          auths: true,
        },
      }),
    );
  }

  async findByEmailWithLocalAuth(email: string) {
    return await this.executeQuery(
      this.db.query.user.findFirst({
        where: eq(user.email, email),
        with: {
          auths: {
            where: eq(auth.strategy, AUTH_STRATEGIES.LOCAL),
          },
        },
      }),
    );
  }

  async resetFailedAttempts(user_id: string) {
    await this.executeQuery(
      this.db
        .update(auth)
        .set({
          failed_attempts: 0,
          locked_until: null,
        })
        .where(
          and(
            eq(auth.user_id, user_id),
            eq(auth.strategy, AUTH_STRATEGIES.LOCAL),
          ),
        ),
    );
  }

  async updateFailedAttempts(user_id: string) {
    // First, get the current failed_attempts value
    const [authRecord] = await this.executeQuery(
      this.db.query.auth.findMany({
        where: (fields, { eq, and }) =>
          and(
            eq(fields.user_id, user_id),
            eq(fields.strategy, AUTH_STRATEGIES.LOCAL),
          ),
        columns: { failed_attempts: true },
      }),
    );

    const failedAttempts = (authRecord?.failed_attempts ?? 0) + 1;
    let lockedUntil: Date | null = null;
    if (failedAttempts >= LOCKDOWN_THRESHOLD) {
      lockedUntil = new Date(Date.now() + LOCKDOWN_DURATION); // lock for 15 minutes
    }

    await this.executeQuery(
      this.db
        .update(auth)
        .set({
          failed_attempts: failedAttempts,
          ...(lockedUntil && { locked_until: lockedUntil }),
        })
        .where(
          and(
            eq(auth.user_id, user_id),
            eq(auth.strategy, AUTH_STRATEGIES.LOCAL),
          ),
        ),
    );
  }
}
