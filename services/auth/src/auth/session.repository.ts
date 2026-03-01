import { Injectable } from '@nestjs/common';
import { eq } from 'drizzle-orm';
import { BaseRepository } from 'src/core/database/base.repository';
import { auth, session } from 'src/core/database/schemas';

@Injectable()
export class SessionRepository extends BaseRepository {
  async create(id: string, userId: string, hash: string, expires_at: Date) {
    const data = await this.findAuthByUserId(userId);

    const [sessionData] = await this.executeQuery(
      this.db
        .insert(session)
        .values({
          id,
          auth_id: data!.id,
          refresh_token_hash: hash,
          expires_at,
        })
        .returning(),
    );
    return sessionData;
  }

  async findById(id: string) {
    return await this.executeQuery(
      this.db.query.session.findFirst({
        where: eq(session.id, id),
      }),
    );
  }

  async findByUserId(userId: string) {
    const data = await this.findAuthByUserId(userId);
    if (!data) return null;

    return await this.executeQuery(
      this.db.query.session.findFirst({
        where: eq(session.auth_id, data.id),
      }),
    );
  }

  async findByAuthIdAndHash(auth_id: string, hash: string) {
    return await this.executeQuery(
      this.db.query.session.findFirst({
        where: (fields, { eq, and }) =>
          and(eq(fields.auth_id, auth_id), eq(fields.refresh_token_hash, hash)),
      }),
    );
  }

  async deleteById(id: string) {
    return await this.executeQuery(
      this.db.delete(session).where(eq(session.id, id)),
    );
  }

  async deleteByAuthId(auth_id: string) {
    return await this.executeQuery(
      this.db.delete(session).where(eq(session.auth_id, auth_id)),
    );
  }

  async update(id: string, hash: string, expires_at: Date) {
    return await this.executeQuery(
      this.db
        .update(session)
        .set({ refresh_token_hash: hash, expires_at })
        .where(eq(session.id, id)),
    );
  }

  async updateWithAuthId(auth_id: string, hash: string, expires_at: Date) {
    return await this.executeQuery(
      this.db
        .update(session)
        .set({ refresh_token_hash: hash, expires_at })
        .where(eq(session.auth_id, auth_id)),
    );
  }

  async findAuthByUserId(userId: string) {
    return await this.executeQuery(
      this.db.query.auth.findFirst({
        where: eq(auth.user_id, userId),
      }),
    );
  }
}
