/**
 * Database Models - Report Schedule
 * 
 * Defines the report schedule entity for automated report generation.
 */

import {
  Entity,
  PrimaryGeneratedColumn,
  Column,
  CreateDateColumn,
  UpdateDateColumn,
  ManyToOne,
  OneToMany,
  JoinColumn,
  Index,
} from 'typeorm';
import { ReportTemplate } from './reportTemplate.entity';
import { ReportHistory } from './reportHistory.entity';

export enum ScheduleStatus {
  ACTIVE = 'ACTIVE',
  PAUSED = 'PAUSED',
  DISABLED = 'DISABLED',
}

export enum ScheduleTrigger {
  CRON = 'CRON',
  MANUAL = 'MANUAL',
  EVENT = 'EVENT',
}

@Entity('report_schedules')
@Index(['status'])
@Index(['entityId'])
@Index(['reportType'])
export class ReportSchedule {
  @PrimaryGeneratedColumn('uuid')
  id: string;

  @Column('uuid', { nullable: true })
  entityId: string;

  @Column('varchar')
  name: string;

  @Column('text', { nullable: true })
  description: string;

  @Column('varchar')
  reportType: string;

  @Column('uuid')
  templateId: string;

  @ManyToOne(() => ReportTemplate)
  @JoinColumn({ name: 'templateId' })
  template: ReportTemplate;

  @Column({
    type: 'enum',
    enum: ScheduleTrigger,
    default: ScheduleTrigger.CRON,
  })
  trigger: ScheduleTrigger;

  @Column('varchar', { nullable: true })
  cronExpression: string;

  @Column('jsonb', { nullable: true })
  parameters: Record<string, any>;

  @Column('jsonb', { nullable: true })
  recipients: string[];

  @Column('jsonb', { nullable: true })
  filterCriteria: Record<string, any>;

  @Column({
    type: 'enum',
    enum: ScheduleStatus,
    default: ScheduleStatus.ACTIVE,
  })
  status: ScheduleStatus;

  @Column('integer', { default: 0 })
  lastRunAt: Date;

  @Column('integer', { default: 0 })
  nextRunAt: Date;

  @Column('integer', { default: 0 })
  runCount: number;

  @Column('boolean', { default: false })
  includeAttachments: boolean;

  @Column('varchar', { default: 'pdf' })
  defaultFormat: string;

  @Column('boolean', { default: true })
  generateOnTrigger: boolean;

  @Column('uuid', { nullable: true })
  createdBy: string;

  @OneToMany(() => ReportHistory, (history) => history.schedule)
  history: ReportHistory[];

  @CreateDateColumn()
  createdAt: Date;

  @UpdateDateColumn()
  updatedAt: Date;
}
