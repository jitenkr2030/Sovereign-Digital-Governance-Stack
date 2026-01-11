import { Module, forwardRef } from '@nestjs/common';
import { TypeOrmModule } from '@nestjs/typeorm';
import { RenewalsService } from './renewals.service';
import { License } from '../licenses/entities/license.entity';
import { KafkaModule } from '../kafka/kafka.module';

@Module({
  imports: [
    TypeOrmModule.forFeature([License]),
    forwardRef(() => KafkaModule),
  ],
  providers: [RenewalsService],
  exports: [RenewalsService],
})
export class RenewalsModule {}
