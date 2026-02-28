import { Injectable } from '@nestjs/common';
import { eq } from 'drizzle-orm';
import { BaseRepository } from 'src/core/database/base.repository';
import { auth, user, userProfile } from 'src/core/database/schemas';
import { SignUpDto } from './dtos/sign-up.dto';
import { AUTH_STRATEGIES, USER_ROLES, USER_STATUS } from 'src/common/constants';

@Injectable()
export class AuthRepository extends BaseRepository {
  async createUser(dto: SignUpDto, hash: string) {
    const data = await this.runInTransaction(async () => {
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

      const newProfile = await this.executeQuery(
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

      return { user: newUser[0], profile: newProfile[0] };
    });

    return data;
  }

  async findUserByEmail(email: string) {
    return await this.executeQuery(
      this.db.query.user.findFirst({
        where: eq(user.email, email),
      }),
    );
  }
}
