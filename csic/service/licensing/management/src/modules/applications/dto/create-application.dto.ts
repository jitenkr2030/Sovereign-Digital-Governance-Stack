import {
  IsString,
  IsUUID,
  IsEnum,
  IsOptional,
  IsObject,
  IsArray,
  ValidateNested,
  IsNumber,
  Min,
  Max,
} from 'class-validator';
import { Type } from 'class-transformer';
import { ApiProperty, ApiPropertyOptional } from '@nestjs/swagger';
import { ApplicationStatus, ApplicationPriority } from '../entities/application.entity';

export class SubmissionDataDto {
  @ApiProperty({ description: 'Registered office address' })
  @IsString()
  registeredAddress: string;

  @ApiProperty({ description: 'Principal place of business' })
  @IsString()
  principalPlaceOfBusiness: string;

  @ApiProperty({ description: 'Website URLs' })
  @IsArray()
  @IsString({ each: true })
  websites: string[];

  @ApiPropertyOptional({ description: 'Governing law jurisdiction' })
  @IsString()
  @IsOptional()
  governingLaw?: string;

  @ApiPropertyOptional({ description: 'Contact information' })
  @IsObject()
  @IsOptional()
  contactInfo?: {
    name: string;
    email: string;
    phone: string;
    position: string;
  };

  @ApiPropertyOptional({ description: 'Beneficial owners information' })
  @IsArray()
  @IsOptional()
  beneficialOwners?: Array<{
    name: string;
    nationality: string;
    ownershipPercentage: number;
    address: string;
  }>;
}

export class CreateApplicationDto {
  @ApiProperty({ description: 'Entity ID from the entity registry' })
  @IsUUID()
  entityId: string;

  @ApiProperty({ description: 'Entity legal name' })
  @IsString()
  entityName: string;

  @ApiProperty({ description: 'Entity type (corporation, partnership, etc.)' })
  @IsString()
  entityType: string;

  @ApiProperty({ description: 'License type ID' })
  @IsUUID()
  licenseTypeId: string;

  @ApiPropertyOptional({ description: 'Application priority level', enum: ApplicationPriority })
  @IsEnum(ApplicationPriority)
  @IsOptional()
  priority?: ApplicationPriority;

  @ApiPropertyOptional({ description: 'Submission data and business information' })
  @ValidateNested()
  @Type(() => SubmissionDataDto)
  @IsOptional()
  submissionData?: SubmissionDataDto;

  @ApiPropertyOptional({ description: 'Due date for application review' })
  @IsOptional()
  dueDate?: Date;
}

export class UpdateApplicationDto {
  @ApiPropertyOptional({ description: 'Application priority', enum: ApplicationPriority })
  @IsEnum(ApplicationPriority)
  @IsOptional()
  priority?: ApplicationPriority;

  @ApiPropertyOptional({ description: 'Submission data' })
  @ValidateNested()
  @Type(() => SubmissionDataDto)
  @IsOptional()
  submissionData?: SubmissionDataDto;

  @ApiPropertyOptional({ description: 'Required document types' })
  @IsArray()
  @IsString({ each: true })
  @IsOptional()
  requiredDocuments?: string[];
}

export class ReviewApplicationDto {
  @ApiProperty({ description: 'New status for the application', enum: ApplicationStatus })
  @IsEnum(ApplicationStatus)
  status: ApplicationStatus;

  @ApiPropertyOptional({ description: 'Review notes' })
  @IsString()
  @IsOptional()
  notes?: string;

  @ApiPropertyOptional({ description: 'Whether notes are internal only' })
  @IsOptional()
  isInternal?: boolean;

  @ApiPropertyOptional({ description: 'Decision reason (for approved/rejected)' })
  @IsString()
  @IsOptional()
  decisionReason?: string;
}

export class SubmitApplicationDto {
  @ApiPropertyOptional({ description: 'Acknowledgment of submission requirements' })
  @IsString()
  @IsOptional()
  acknowledgment?: string;
}

export class ApplicationFilterDto {
  @ApiPropertyOptional({ description: 'Filter by status', enum: ApplicationStatus })
  @IsEnum(ApplicationStatus)
  @IsOptional()
  status?: ApplicationStatus;

  @ApiPropertyOptional({ description: 'Filter by entity ID' })
  @IsUUID()
  @IsOptional()
  entityId?: string;

  @ApiPropertyOptional({ description: 'Filter by license type ID' })
  @IsUUID()
  @IsOptional()
  licenseTypeId?: string;

  @ApiPropertyOptional({ description: 'Filter by priority', enum: ApplicationPriority })
  @IsEnum(ApplicationPriority)
  @IsOptional()
  priority?: ApplicationPriority;

  @ApiPropertyOptional({ description: 'Filter by handler ID' })
  @IsUUID()
  @IsOptional()
  handlerId?: string;

  @ApiPropertyOptional({ description: 'Page number', default: 1 })
  @IsNumber()
  @Min(1)
  @IsOptional()
  @Type(() => Number)
  page?: number = 1;

  @ApiPropertyOptional({ description: 'Items per page', default: 20 })
  @IsNumber()
  @Min(1)
  @Max(100)
  @IsOptional()
  @Type(() => Number)
  limit?: number = 20;
}
