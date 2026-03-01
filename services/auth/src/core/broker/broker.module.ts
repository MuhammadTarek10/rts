import { Global, Module } from '@nestjs/common';
import { ConfigModule, ConfigService } from '@nestjs/config';
import { ClientsModule, Transport } from '@nestjs/microservices';
import { ENVIRONMENT } from 'src/common/constants';

@Global()
@Module({
  imports: [
    ClientsModule.registerAsync([
      {
        name: ENVIRONMENT.BROKER.KEY,
        imports: [ConfigModule],
        inject: [ConfigService],
        useFactory: (config: ConfigService) => ({
          transport: Transport.RMQ,
          options: {
            urls: [config.getOrThrow<string>(ENVIRONMENT.BROKER.URI)],
            queue: ENVIRONMENT.BROKER.QUEUES.AUTH_EVENTS,
            queueOptions: {
              durable: true,
            },
          },
        }),
      },
    ]),
  ],
  exports: [ClientsModule],
})
export class BrokerModule {}
