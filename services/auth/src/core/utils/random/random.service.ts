import { Injectable } from '@nestjs/common';

@Injectable()
export class RandomService {
  generateRandomUUID(): string {
    return crypto.randomUUID();
  }
}
