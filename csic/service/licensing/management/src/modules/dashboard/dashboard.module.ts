import { Module } from '@nestjs/common';
import { TypeOrmModule } from '@nestjs/typeorm';
import { DashboardController } from './dashboard.controller';
import { DashboardService } from './dashboard.service';
import { Application } from '../applications/entities/application.entity';
import { License } from '../licenses/entities/license.entity';
import { LicenseType } from '../applications/entities/license-type.entity';

@Module({
  imports: [TypeOrmModule.forFeature([Application, License, LicenseType])],
  controllers: [DashboardController],
  providers: [DashboardService],
  exports: [DashboardService],
})
export class DashboardModule {}
