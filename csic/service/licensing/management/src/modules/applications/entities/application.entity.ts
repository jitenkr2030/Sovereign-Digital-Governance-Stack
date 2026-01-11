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
import { LicenseType } from './license-type.entity';
import { License } from '../../licenses/entities/license.entity';
import { Document } from '../../documents/entities/document.entity';

export enum ApplicationStatus {
  DRAFT = 'DRAFT',
  SUBMITTED = 'SUBMITTED',
  UNDER_REVIEW = 'UNDER_REVIEW',
  PENDING_INFO = 'PENDING_INFO',
  APPROVED = 'APPROVED',
  REJECTED = 'REJECTED',
}

export enum ApplicationPriority {
  LOW = 'LOW',
  NORMAL = 'NORMAL',
  HIGH = 'HIGH',
  URGENT = 'URGENT',
}

@Entity('applications')
@Index(['status'])
@Index(['entityId'])
@Index(['currentHandlerId'])
export class Application {
  @PrimaryGeneratedColumn('uuid')
  id: string;

  @Column('uuid')
  entityId: string;

  @Column('varchar')
  entityName: string;

  @Column('varchar')
  entityType: string;

  @Column('uuid')
  licenseTypeId: string;

  @ManyToOne(() => LicenseType)
  @JoinColumn({ name: 'licenseTypeId' })
  licenseType: LicenseType;

  @Column({
    type: 'enum',
    enum: ApplicationStatus,
    default: ApplicationStatus.DRAFT,
  })
  status: ApplicationStatus;

  @Column({
    type: 'enum',
    enum: ApplicationPriority,
    default: ApplicationPriority.NORMAL,
  })
  priority: ApplicationPriority;

  @Column('jsonb', { nullable: true })
  submissionData: Record<string, any>;

  @Column('jsonb', { nullable: true })
  requiredDocuments: string[];

  @Column('jsonb', { nullable: true })
  reviewNotes: ReviewNote[];

  @Column('uuid', { nullable: true })
  currentHandlerId: string;

  @Column('varchar', { nullable: true })
  currentHandlerName: string;

  @Column('timestamp', { nullable: true })
  submittedAt: Date;

  @Column('timestamp', { nullable: true })
  reviewStartedAt: Date;

  @Column('timestamp', { nullable: true })
  decisionAt: Date;

  @Column('varchar', { nullable: true })
  decision: string;

  @Column('text', { nullable: true })
  decisionReason: string;

  @OneToMany(() => License, (license) => license.application)
  licenses: License[];

  @OneToMany(() => Document, (document) => document.application)
  documents: Document[];

  @CreateDateColumn()
  createdAt: Date;

  @UpdateDateColumn()
  updatedAt: Date;

  @Column('timestamp', { nullable: true })
  dueDate: Date;
}

export interface ReviewNote {
  id: string;
  handlerId: string;
  handlerName: string;
  content: string;
  createdAt: Date;
  isInternal: boolean;
}
