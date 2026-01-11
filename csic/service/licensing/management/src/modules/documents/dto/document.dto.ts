import { IsString, IsUUID, IsEnum, IsOptional, IsNotEmpty } from 'class-validator';
import { ApiProperty, ApiPropertyOptional } from '@nestjs/swagger';
import { DocumentType } from '../entities/document.entity';

export class UploadDocumentDto {
  @ApiProperty({ description: 'Application ID to upload document for' })
  @IsUUID()
  applicationId: string;

  @ApiProperty({ description: 'Type of document', enum: DocumentType })
  @IsEnum(DocumentType)
  documentType: DocumentType;
}

export class VerifyDocumentDto {
  @ApiPropertyOptional({ description: 'Verification notes' })
  @IsString()
  @IsOptional()
  notes?: string;
}

export class RejectDocumentDto {
  @ApiProperty({ description: 'Reason for rejection' })
  @IsString()
  @IsNotEmpty()
  reason: string;
}
