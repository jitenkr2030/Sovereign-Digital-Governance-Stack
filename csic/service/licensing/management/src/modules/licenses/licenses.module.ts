import { Module, forwardRef } from '@nestjs/common';
import { TypeOrmModule } from '@nestjs/typeorm';
import { LicensesController } from './licenses.controller';
import { LicensesService } from './licenses.service';
import { License } from './entities/license.entity';
import { Application } from '../applications/entities/application.entity';
import { KafkaModule } from '../kafka/kafka.module';
import { WorkflowModule } from '../workflow/workflow.module';

@Module({
  imports: [
    TypeOrmModule.forFeature([License, Application]),
    forwardRef(() => KafkaModule),
    forwardRef(() => WorkflowModule),
  ],
  controllers: [LicensesController],
  providers: [LicensesService],
  exports: [LicensesService],
})
export class LicensesModule {}
