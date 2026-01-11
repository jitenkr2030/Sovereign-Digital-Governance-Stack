/**
 * Database Models - Report History
 * 
 * Tracks all generated reports with WORM storage information.
 */

import {
  Entity,
  PrimaryGeneratedColumn,
  Column,
  CreateDateColumn,
  UpdateDateColumn,
  ManyToOne,
  JoinColumn,
  Index,
} from 'typeorm';
import { ReportSchedule } from './reportSchedule.entity';

export enum ReportStatus {
  PENDING = 'PENDING',
  PROCESSING = 'PROCESSING',
  COMPLETED = 'COMPLETED',
  FAILED = 'FAILED',
  CANCELLED = 'CANCELLED',
}

export enum ReportFormat {
  PDF = 'pdf',
  CSV = 'csv',
  JSON = 'json',
}

export enum ReportTrigger {
  SCHEDULED = 'SCHEDULED',
  MANUAL = 'MANUAL',
  API = 'API',
  SYSTEM = 'SYSTEM',
}

@Entity('report_history')
@Index(['status'])
@Index(['entityId'])
@Index(['reportType'])
@Index(['generatedAt'])
@Index(['fileHash'], { unique: true })
export class ReportHistory {
  @PrimaryGeneratedColumn('uuid')
  id: string;

  @Column('uuid', { nullable: true })
  scheduleId: string;

  @ManyToOne(() => ReportSchedule, (schedule) => schedule.history, { nullable: true })
  @JoinColumn({ name: 'scheduleId' })
  schedule: ReportSchedule;

  @Column('uuid', { nullable: true })
  entityId: string;

  @Column('varchar')
  reportType: string;

  @Column('varchar')
  name: string;

  @Column({
    type: 'enum',
    enum: ReportFormat,
    default: ReportFormat.PDF,
  })
  format: ReportFormat;

  @Column({
    type: 'enum',
    enum: ReportStatus,
    default: ReportStatus.PENDING,
  })
  status: ReportStatus;

  @Column({
    type: 'enum',
    enum: ReportTrigger,
    default: ReportTrigger.MANUAL,
  })
  trigger: ReportTrigger;

  @Column('jsonb', { nullable: true })
  parameters: Record<string, any>;

  @Column('jsonb', { nullable: true })
  filterCriteria: Record<string, any>;

  // WORM Storage Information
  @Column('varchar', { nullable: true })
  s3Key: string;

  @Column('varchar', { nullable: true })
  s3Bucket: string;

  @Column('varchar', { nullable: true })
  storagePath: string;

  @Column('varchar', { nullable: true })
  fileHash: string;

  @Column('bigint', { nullable: true })
  fileSize: number;

  @Column('date', { nullable: true })
  reportPeriodStart: Date;

  @Column('date', { nullable: true })
  reportPeriodEnd: Date;

  // Download tracking for chain of custody
  @Column('integer', { default: 0 })
  downloadCount: number;

  @Column('timestamp', { nullable: true })
  lastDownloadedAt: Date;

  @Column('jsonb', { nullable: true })
  downloadLog: Array<{
    timestamp: Date;
    userId: string;
    ipAddress: string;
    userAgent: string;
  }>;

  // Access control
  @Column('jsonb', { nullable: true })
  allowedRoles: string[];

  @Column('boolean', { default: false })
  isConfidential: boolean;

  // WORM compliance fields
  @Column('boolean', { default: false })
  legalHold: boolean;

  @Column('date', { nullable: true })
  retentionUntil: Date;

  @Column('varchar', { nullable: true })
  retentionMode: string;

  // Error information
  @Column('text', { nullable: true })
  errorMessage: string;

  @Column('jsonb', { nullable: true })
  errorDetails: Record<string, any>;

  // Generation metadata
  @Column('integer', { default: 0 })
  generationTimeMs: number;

  @Column('uuid', { nullable: true })
  requestedBy: string;

  @Column('timestamp', { nullable: true })
  generatedAt: Date;

  @Column('timestamp', { nullable: true })
  completedAt: Date;

  @CreateDateColumn()
  createdAt: Date;

  @UpdateDateColumn()
  updatedAt: Date;
}
