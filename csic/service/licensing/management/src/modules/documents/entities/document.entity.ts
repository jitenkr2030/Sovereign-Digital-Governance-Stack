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
import { Application } from '../../applications/entities/application.entity';

export enum DocumentType {
  PROOF_OF_INCORPORATION = 'proof_of_incorporation',
  AML_KYC_POLICY = 'aml_kyc_policy',
  BUSINESS_PLAN = 'business_plan',
  FINANCIAL_STATEMENTS = 'financial_statements',
  SECURITY_AUDIT_REPORT = 'security_audit_report',
  COLD_STORAGE_POLICY = 'cold_storage_policy',
  INSURANCE_PROOF = 'insurance_proof',
  CUSTODY_POLICY = 'custody_policy',
  ENERGY_SOURCE_DOCUMENTATION = 'energy_source_documentation',
  POOL_SOFTWARE_SECURITY_AUDIT = 'pool_software_security_audit',
  GEOLOCATION_DATA = 'geolocation_data',
  OPERATIONAL_RESERVE_PROOF = 'operational_reserve_proof',
  OTHER = 'other',
}

export enum DocumentStatus {
  UPLOADED = 'UPLOADED',
  PENDING_VERIFICATION = 'PENDING_VERIFICATION',
  VERIFIED = 'VERIFIED',
  REJECTED = 'REJECTED',
  QUARANTINED = 'QUARANTINED',
}

@Entity('documents')
@Index(['applicationId'])
@Index(['status'])
@Index(['documentType'])
export class Document {
  @PrimaryGeneratedColumn('uuid')
  id: string;

  @Column('uuid')
  applicationId: string;

  @ManyToOne(() => Application, (application) => application.documents)
  @JoinColumn({ name: 'applicationId' })
  application: Application;

  @Column('varchar')
  fileName: string;

  @Column('varchar')
  originalName: string;

  @Column('varchar')
  mimeType: string;

  @Column('bigint')
  fileSize: number;

  @Column('varchar')
  s3Key: string;

  @Column('varchar', { nullable: true })
  s3Bucket: string;

  @Column({
    type: 'enum',
    enum: DocumentType,
  })
  documentType: DocumentType;

  @Column({
    type: 'enum',
    enum: DocumentStatus,
    default: DocumentStatus.UPLOADED,
  })
  status: DocumentStatus;

  @Column('varchar')
  checksum: string;

  @Column('varchar', { nullable: true })
  checksumAlgorithm: string;

  @Column('text', { nullable: true })
  description: string;

  @Column('uuid', { nullable: true })
  uploadedBy: string;

  @Column('varchar', { nullable: true })
  uploadedByName: string;

  @Column('timestamp', { nullable: true })
  verifiedAt: Date;

  @Column('uuid', { nullable: true })
  verifiedBy: string;

  @Column('varchar', { nullable: true })
  verifiedByName: string;

  @Column('text', { nullable: true })
  verificationNotes: string;

  @Column('timestamp', { nullable: true })
  rejectedAt: Date;

  @Column('text', { nullable: true })
  rejectionReason: string;

  @Column('jsonb', { nullable: true })
  metadata: Record<string, any>;

  @Column('boolean', { default: false })
  isConfidential: boolean;

  @CreateDateColumn()
  createdAt: Date;

  @UpdateDateColumn()
  updatedAt: Date;

  @Column('timestamp', { nullable: true })
  quarantineUntil: Date;
}
