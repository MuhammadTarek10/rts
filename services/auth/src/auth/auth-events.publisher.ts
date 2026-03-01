import { Inject, Injectable, Logger } from '@nestjs/common';
import { ClientProxy } from '@nestjs/microservices';
import { ENVIRONMENT } from 'src/common/constants';

export type UserRegisteredEvent = {
  id: string;
  email: string;
};

@Injectable()
export class AuthEventsPublisher {
  private readonly logger = new Logger(AuthEventsPublisher.name);

  constructor(
    @Inject(ENVIRONMENT.BROKER.KEY) private readonly brokerClient: ClientProxy,
  ) {}

  emitUserRegistered(payload: UserRegisteredEvent) {
    this.brokerClient
      .emit(ENVIRONMENT.BROKER.EVENTS.USER_REGISTERED, payload)
      .subscribe({
        error: (error: unknown) => {
          this.logger.warn(
            `Failed to publish UserRegistered event: ${String(error)}`,
          );
        },
      });
  }
}
