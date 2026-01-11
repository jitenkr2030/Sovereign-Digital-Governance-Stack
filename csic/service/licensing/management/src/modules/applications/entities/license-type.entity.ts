import {
  Entity,
  PrimaryGeneratedColumn,
  Column,
  CreateDateColumn,
  UpdateDateColumn,
} from 'typeorm';

@Entity('license_types')
export class LicenseType {
  @PrimaryGeneratedColumn('uuid')
  id: string;

  @Column('varchar', { unique: true })
  code: string;

  @Column('varchar')
  name: string;

  @Column('text')
  description: string;

  @Column('integer')
  validityMonths: number;

  @Column('jsonb')
  requirements: string[];

  @Column('jsonb', { nullable: true })
  metadata: Record<string, any>;

  @Column('boolean', { default: true })
  isActive: boolean;

  @Column('integer', { default: 1 })
  version: number;

  @CreateDateColumn()
  createdAt: Date;

  @UpdateDateColumn()
  updatedAt: Date;
}
