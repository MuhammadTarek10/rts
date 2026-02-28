import { Inject, Injectable } from '@nestjs/common';
import { ENVIRONMENT } from 'src/common/constants';
import { PostgresDatabase, PostgresTransaction } from './types';

@Injectable()
export class BaseRepository {
  constructor(
    @Inject(ENVIRONMENT.DATABASE.KEY)
    protected readonly db: PostgresDatabase,
  ) {}

  async runInTransaction<T>(
    fn: (tx: PostgresTransaction) => Promise<T>,
  ): Promise<T> {
    try {
      return await this.db.transaction(fn);
    } catch (error) {
      this.handleError(error, 'Transaction failed');
      throw error; // Re-throw the error after logging
    }
  }

  handleError(error: any, message: string) {
    console.error(message, error);
  }

  async executeQuery<T>(query: Promise<T>): Promise<T> {
    try {
      return await query;
    } catch (error) {
      this.handleError(error, 'Database query failed');
      throw error; // Re-throw the error after logging
    }
  }
}
