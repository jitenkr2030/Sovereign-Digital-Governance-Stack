/**
 * Database Models - Report Template
 * 
 * Defines the report template entity for report generation.
 */

import {
  Entity,
  PrimaryGeneratedColumn,
  Column,
  CreateDateColumn,
  UpdateDateColumn,
  Index,
} from 'typeorm';

export enum TemplateFormat {
  PDF = 'pdf',
  CSV = 'csv',
  JSON = 'json',
}

@Entity('report_templates')
@Index(['reportType'])
@Index(['isActive'])
export class ReportTemplate {
  @PrimaryGeneratedColumn('uuid')
  id: string;

  @Column('varchar')
  name: string;

  @Column('text', { nullable: true })
  description: string;

  @Column('varchar')
  reportType: string;

  @Column({
    type: 'enum',
    enum: TemplateFormat,
    default: TemplateFormat.PDF,
  })
  format: TemplateFormat;

  @Column('text')
  templateContent: string;

  @Column('text', { nullable: true })
  headerTemplate: string;

  @Column('text', { nullable: true })
  footerTemplate: string;

  @Column('jsonb', { nullable: true })
  templateVariables: string[];

  @Column('jsonb', { nullable: true })
  defaultParameters: Record<string, any>;

  @Column('jsonb', { nullable: true })
  styling: Record<string, any>;

  @Column('boolean', { default: true })
  isActive: boolean;

  @Column('integer', { default: 1 })
  version: number;

  @Column('uuid', { nullable: true })
  createdBy: string;

  @Column('uuid', { nullable: true })
  updatedBy: string;

  @CreateDateColumn()
  createdAt: Date;

  @UpdateDateColumn()
  updatedAt: Date;
}
