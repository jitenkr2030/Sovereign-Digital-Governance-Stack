import { Module } from '@nestjs/common';
import { ConfigModule, ConfigService } from '@nestjs/config';
import { TypeOrmModule } from '@nestjs/typeorm';
import { ScheduleModule } from '@nestjs/schedule';
import { ApplicationsModule } from './modules/applications/applications.module';
import { LicensesModule } from './modules/licenses/licenses.module';
import { DocumentsModule } from './modules/documents/documents.module';
import { WorkflowModule } from './modules/workflow/workflow.module';
import { RenewalsModule } from './modules/renewals/renewals.module';
import { DashboardModule } from './modules/dashboard/dashboard.module';
import { KafkaModule } from './modules/kafka/kafka.module';

@Module({
  imports: [
    // Configuration
    ConfigModule.forRoot({
      isGlobal: true,
      envFilePath: ['.env.local', '.env'],
    }),

    // Database
    TypeOrmModule.forRootAsync({
      imports: [ConfigModule],
      inject: [ConfigService],
      useFactory: (configService: ConfigService) => ({
        type: 'postgres',
        host: configService.get('DATABASE_HOST', 'localhost'),
        port: configService.get('DATABASE_PORT', 5432),
        username: configService.get('DATABASE_USER', 'csic_user'),
        password: configService.get('DATABASE_PASSWORD', 'password'),
        database: configService.get('DATABASE_NAME', 'csic_licensing'),
        entities: [__dirname + '/**/*.entity{.ts,.js}'],
        synchronize: configService.get('NODE_ENV') === 'development',
        logging: configService.get('NODE_ENV') === 'development',
        ssl: configService.get('DATABASE_SSL') === 'true' ? { rejectUnauthorized: false } : false,
      }),
    }),

    // Scheduled tasks
    ScheduleModule.forRoot(),

    // Feature modules
    KafkaModule,
    ApplicationsModule,
    LicensesModule,
    DocumentsModule,
    WorkflowModule,
    RenewalsModule,
    DashboardModule,
  ],
})
export class AppModule {}
