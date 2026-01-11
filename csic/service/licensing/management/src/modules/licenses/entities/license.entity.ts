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
import { Application } from '../../applications/entities/application.entity';
import { LicenseStatus } from './license-status.enum';

@Entity('licenses')
@Index(['status'])
@Index(['entityId'])
@Index(['expiryDate'])
@Index(['licenseNumber'], { unique: true })
export class License {
  @PrimaryGeneratedColumn('uuid')
  id: string;

  @Column('uuid')
  applicationId: string;

  @ManyToOne(() => Application, (application) => application.licenses)
  @JoinColumn({ name: 'applicationId' })
  application: Application;

  @Column('uuid')
  entityId: string;

  @Column('varchar')
  entityName: string;

  @Column('varchar')
  entityType: string;

  @Column('varchar')
  licenseNumber: string;

  @Column('uuid')
  licenseTypeId: string;

  @Column('varchar')
  licenseTypeName: string;

  @Column({
    type: 'enum',
    enum: LicenseStatus,
    default: LicenseStatus.ACTIVE,
  })
  status: LicenseStatus;

  @Column('date')
  issueDate: Date;

  @Column('date')
  expiryDate: Date;

  @Column('date', { nullable: true })
  renewalWindowStart: Date;

  @Column('date', { nullable: true })
  renewedAt: Date;

  @Column('date', { nullable: true })
  suspendedAt: Date;

  @Column('date', { nullable: true })
  revokedAt: Date;

  @Column('varchar', { nullable: true })
  suspendedReason: string;

  @Column('varchar', { nullable: true })
  revokedReason: string;

  @Column('varchar', { nullable: true })
  suspendedBy: string;

  @Column('varchar', { nullable: true })
  revokedBy: string;

  @Column('jsonb', { nullable: true })
  scope: string[];

  @Column('jsonb', { nullable: true })
  conditions: LicenseCondition[];

  @Column('jsonb', { nullable: true })
  metadata: Record<string, any>;

  @Column('varchar', { nullable: true })
  certificateUrl: string;

  @CreateDateColumn()
  createdAt: Date;

  @UpdateDateColumn()
  updatedAt: Date;
}

export interface LicenseCondition {
  id: string;
  title: string;
  description: string;
  isMet: boolean;
  dueDate?: Date;
}
